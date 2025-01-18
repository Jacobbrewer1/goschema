package migrations

import (
	"fmt"

	"github.com/jacobbrewer1/goschema/pkg/models"
)

func (v *versioning) GetStatus() ([]*models.GoschemaMigrationVersion, error) {
	sqlStmt := `
		SELECT version
		FROM goschema_migration_version
		ORDER BY created_at DESC;
	`

	ids := make([]string, 0)
	if err := v.db.Select(&ids, sqlStmt); err != nil {
		return nil, fmt.Errorf("get status ids: %w", err)
	}

	versions := make([]*models.GoschemaMigrationVersion, len(ids))
	for i, id := range ids {
		ver, err := models.GoschemaMigrationVersionByVersion(v.db, id)
		if err != nil {
			return nil, fmt.Errorf("get status version by version: %w", err)
		}

		versions[i] = ver
	}

	return versions, nil
}
