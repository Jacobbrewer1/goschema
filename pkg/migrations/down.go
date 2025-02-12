package migrations

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jacobbrewer1/goschema/pkg/logging"
)

func (v *versioning) MigrateDown() error {
	if err := v.createTableIfNotExists(); err != nil {
		return fmt.Errorf("error checking or creating migration tables: %w", err)
	}

	// Get all files in the migration location.
	files, err := getFiles(v.migrationLocation)
	if err != nil {
		return fmt.Errorf("error getting files: %w", err)
	}

	// Filter the files
	files = filterFiles(files, ".sql")
	files = filterFiles(files, down+".sql")

	// Order the files by datetime at the prefix.
	orderedFiles, err := orderFiles(files)
	if err != nil {
		return fmt.Errorf("error ordering files: %w", err)
	}

	// Reverse the order of the files.
	for i, j := 0, len(orderedFiles)-1; i < j; i, j = i+1, j-1 {
		orderedFiles[i], orderedFiles[j] = orderedFiles[j], orderedFiles[i]
	}

	// Get the current version.
	currentVersion, err := v.getCurrentVersion()
	if err != nil && !errors.Is(err, ErrNoCurrentVersion) {
		return fmt.Errorf("error getting current version: %w", err)
	} else if errors.Is(err, ErrNoCurrentVersion) {
		currentVersion = ""
	}

	// Migrate down.
	count := 0
	for _, f := range orderedFiles {
		slog.Debug("Migrating down", slog.String(logging.KeyFile, f.Name()))

		// Get the datetime prefix.
		prefix, err := getDatetimePrefix(f.Name())
		if err != nil {
			return fmt.Errorf("error getting datetime prefix: %w", err)
		}

		// Is the version below the current version?
		parsed, err := time.Parse(FilePrefix, prefix)
		if err != nil {
			return fmt.Errorf("error parsing datetime prefix: %w", err)
		}

		// There should be a current version. If there is not, then we should not migrate down.
		if currentVersion == "" {
			continue
		}

		currentParsed, err := time.Parse(FilePrefix, currentVersion)
		if err != nil {
			return fmt.Errorf("error parsing current version: %w", err)
		}

		if parsed.After(currentParsed) {
			continue
		}

		if v.steps > 0 && count == v.steps {
			break
		}
		count++

		// Migrate down.
		if err := v.migrate(f, down); err != nil {
			return fmt.Errorf("error migrating down: %w", err)
		}
	}

	return nil
}
