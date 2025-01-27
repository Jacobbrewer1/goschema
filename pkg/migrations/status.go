package migrations

import (
	"fmt"
	"sort"

	"github.com/jacobbrewer1/goschema/pkg/models"
)

func (v *versioning) GetStatus() ([]*models.GoschemaMigrationVersion, error) {
	versions, err := models.GetAllGoschemaMigrationVersion(v.db)
	if err != nil {
		return nil, fmt.Errorf("error getting goschema migration versions: %w", err)
	}

	// Sort the versions by created_at.
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreatedAt.Before(versions[j].CreatedAt)
	})

	return versions, nil
}
