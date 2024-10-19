package migrations

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	versionTable = "goschema_migration_version"
	historyTable = "goschema_migration_history"

	up            = "up"
	down          = "down"
	migratingUp   = "migrating_up"
	migratingDown = "migrating_down"
	stateError    = "error"

	FilePrefix = "20060102150405"
)

var (
	// ErrNoCurrentVersion is the error when there is no current version.
	ErrNoCurrentVersion = errors.New("no current version")
)

type Versioning interface {
	MigrateUp() error
	MigrateDown() error
}

type versioning struct {
	db *sqlx.DB

	// migrationLocation is the location of the migrations.
	migrationLocation string
}

func NewVersioning(db *sqlx.DB, migrationLocation string) Versioning {
	return &versioning{
		db:                db,
		migrationLocation: migrationLocation,
	}
}

func (v *versioning) createOrAppendVersion(version string) error {
	if err := v.createTableIfNotExists(); err != nil {
		return fmt.Errorf("error checking or creating migration tables: %w", err)
	}

	// See if the version exists in the database.
	exists := false
	err := v.db.Get(&exists, "SELECT EXISTS (SELECT 1 FROM "+versionTable+" WHERE version = ?)", version)
	if err != nil {
		return fmt.Errorf("error checking if version exists: %w", err)
	}

	if exists {
		// Is the version already the current version? Or is it below the current version?
		// If the version is below the current version, we don't need to do anything.
		current := false
		err = v.db.Get(&current, "SELECT isCurrent FROM "+versionTable+" WHERE version = ?", version)
		if err != nil {
			return fmt.Errorf("error getting current version: %w", err)
		}

		if current {
			return nil
		}
	}

	// The version does not exist or is not the current version.
	// We need to update the current version to false for all versions.
	_, err = v.db.Exec("UPDATE " + versionTable + " SET isCurrent = false WHERE isCurrent = true")
	if err != nil {
		return fmt.Errorf("error updating current version: %w", err)
	}

	// Insert the new version.
	_, err = v.db.Exec("INSERT INTO "+versionTable+" (version, current) VALUES (?, true)", version)
	if err != nil {
		return fmt.Errorf("error inserting version: %w", err)
	}

	// Insert the version into the history table.
	_, err = v.db.Exec("INSERT INTO "+historyTable+" (version, action, created_at) VALUES (?, ?, ?)", version, up, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("error inserting version into history: %w", err)
	}

	return nil
}

func (v *versioning) createTableIfNotExists() error {
	schema, err := v.getSchema()
	if err != nil {
		return fmt.Errorf("error getting schema: %w", err)
	}

	exists, err := v.doesVersionTableExist(schema)
	if err != nil {
		return fmt.Errorf("error checking if migration_version table exists: %w", err)
	}

	if !exists {
		if err = v.createVersionTable(schema); err != nil {
			return fmt.Errorf("error creating migration_version table: %w", err)
		}
	}

	exists, err = v.doesHistoryTableExist(schema)
	if err != nil {
		return fmt.Errorf("error checking if migration_history table exists: %w", err)
	}

	if !exists {
		if err = v.createHistoryTable(schema); err != nil {
			return fmt.Errorf("error creating migration_history table: %w", err)
		}
	}

	return nil
}

func (v *versioning) getSchema() (string, error) {
	var schema string
	err := v.db.Get(&schema, "SELECT DATABASE()")
	if err != nil {
		return "", fmt.Errorf("error getting schema: %w", err)
	}

	return schema, nil
}

func (v *versioning) doesVersionTableExist(schema string) (bool, error) {
	sqlStmt := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = ?
			AND table_name = ?
		);
`

	exists := false
	err := v.db.Get(&exists, sqlStmt, schema, versionTable)
	if err != nil {
		return false, fmt.Errorf("error checking if migration_version table exists: %w", err)
	}

	return exists, nil
}

func (v *versioning) createVersionTable(schema string) error {
	sqlStmt := fmt.Sprintf(`
		CREATE TABLE %s.%s (
			version VARCHAR(255) NOT NULL PRIMARY KEY,
		    isCurrent BOOLEAN NOT NULL DEFAULT false
		);
`, schema, versionTable)

	_, err := v.db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("error creating migration_version table: %w", err)
	}

	return nil
}

func (v *versioning) doesHistoryTableExist(schema string) (bool, error) {
	sqlStmt := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = ?
			AND table_name = ?
		);
`

	exists := false
	err := v.db.Get(&exists, sqlStmt, schema, historyTable)
	if err != nil {
		return false, fmt.Errorf("error checking if migration_history table exists: %w", err)
	}

	return exists, nil
}

func (v *versioning) createHistoryTable(schema string) error {
	sqlStmt := fmt.Sprintf(`
		CREATE TABLE %s.%s (
			id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			version VARCHAR(255) NOT NULL,
		    action enum('up', 'down', 'migrating_up', 'migrating_down', 'error') NOT NULL,
			created_at TIMESTAMP
		);
`, schema, historyTable)

	_, err := v.db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("error creating migration_history table: %w", err)
	}

	return nil
}

func (v *versioning) getCurrentVersion() (string, error) {
	var version string
	err := v.db.Get(&version, "SELECT version FROM "+versionTable+" WHERE isCurrent = true")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("error getting current version: %w", err)
	}

	return version, nil
}

func (v *versioning) mustSetCurrentVersion(version string) {
	if err := v.setCurrentVersion(version); err != nil {
		panic(err)
	}
}

func (v *versioning) setCurrentVersion(version string) error {
	_, err := v.db.Exec("UPDATE " + versionTable + " SET isCurrent = false WHERE isCurrent = true")
	if err != nil {
		return fmt.Errorf("error updating current version: %w", err)
	}

	_, err = v.db.Exec("INSERT INTO "+versionTable+" (version, isCurrent) VALUES (?, true) ON DUPLICATE KEY UPDATE isCurrent = true", version)
	if err != nil {
		return fmt.Errorf("error setting current version: %w", err)
	}

	return nil
}

func (v *versioning) mustCreateHistory(version string, action string) {
	if err := v.createHistory(version, action); err != nil {
		panic(err)
	}
}

func (v *versioning) createHistory(version string, action string) error {
	_, err := v.db.Exec("INSERT INTO "+historyTable+" (version, action, created_at) VALUES (?, ?, ?)", version, action, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("error creating history: %w", err)
	}

	return nil
}

func (v *versioning) getPreviousVersion() (string, error) {
	sqlStmt := `
SELECT version
FROM goschema_migration_version
WHERE version < (SELECT version FROM goschema_migration_version WHERE isCurrent = true)
ORDER BY version DESC
LIMIT 1;
`

	var version string
	err := v.db.Get(&version, sqlStmt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("error getting previous version: %w", err)
	}

	return version, nil
}

func (v *versioning) mustSetNoCurrentVersion() {
	if err := v.setNoCurrentVersion(); err != nil {
		panic(err)
	}
}

func (v *versioning) setNoCurrentVersion() error {
	_, err := v.db.Exec("UPDATE " + versionTable + " SET isCurrent = false WHERE isCurrent = true")
	if err != nil {
		return fmt.Errorf("error setting no current version: %w", err)
	}

	return nil
}
