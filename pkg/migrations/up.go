package migrations

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

var (
	// ErrLocationIsNotDirectory is the error when the location is not a directory.
	ErrLocationIsNotDirectory = errors.New("location is not a directory")
)

func (v *versioning) MigrateUp() error {
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
	files = filterFiles(files, up+".sql")

	// Order the files by datetime at the prefix.
	orderedFiles, err := orderFiles(files)
	if err != nil {
		return fmt.Errorf("error ordering files: %w", err)
	}

	// Get the current version.
	currentVersion, err := v.getCurrentVersion()
	if err != nil && !errors.Is(err, ErrNoCurrentVersion) {
		return fmt.Errorf("error getting current version: %w", err)
	} else if errors.Is(err, ErrNoCurrentVersion) {
		currentVersion = ""
	}

	// Migrate up.
	for _, f := range orderedFiles {
		slog.Debug(fmt.Sprintf("Migrating up: %s", f.Name()))

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

		if currentVersion != "" {
			currentParsed, err := time.Parse(FilePrefix, currentVersion)
			if err != nil {
				return fmt.Errorf("error parsing current version: %w", err)
			}

			if parsed.Before(currentParsed) {
				continue
			}
		}

		// Migrate up.
		if err := v.migrate(f, up); err != nil {
			return fmt.Errorf("error migrating up: %w", err)
		}
	}

	return nil
}

func (v *versioning) migrate(f os.DirEntry, direction string) error {
	// Get the datetime prefix.
	prefix, err := getDatetimePrefix(f.Name())
	if err != nil {
		return fmt.Errorf("error getting datetime prefix: %w", err)
	}

	switch direction {
	case up:
		v.mustCreateHistory(prefix, migratingUp)
	case down:
		v.mustCreateHistory(prefix, migratingDown)
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}

	// Open the file.
	file, err := os.Open(f.Name())
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("error closing file", slog.String("error", err.Error()))
		}
	}()

	// Read the file.
	b, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Begin a transaction.
	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			slog.Error("error rolling back transaction", slog.String("error", err.Error()))
		}
	}()

	// Execute the file.
	if _, err := tx.Exec(string(b)); err != nil {
		v.mustCreateHistory(prefix, stateError)
		return fmt.Errorf("error executing file: %w", err)
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	switch direction {
	case up:
		v.mustSetCurrentVersion(prefix)
		v.mustCreateHistory(prefix, up)
	case down:
		// Set the current version to the previous version.
		prev, err := v.getPreviousVersion()
		if err != nil {
			return fmt.Errorf("error getting previous version: %w", err)
		} else if prev == "" {
			v.mustSetNoCurrentVersion()
		} else {
			v.mustSetCurrentVersion(prev)
		}

		v.mustCreateHistory(prefix, down)
	}

	return nil
}

func orderFiles(files []os.DirEntry) ([]os.DirEntry, error) {
	ordered := make([]os.DirEntry, 0, len(files))
	for _, f := range files {
		// Get the datetime prefix.
		prefix, err := getDatetimePrefix(f.Name())
		if err != nil {
			return nil, fmt.Errorf("error getting datetime prefix: %w", err)
		}

		if len(ordered) == 0 {
			ordered = append(ordered, f)
			continue
		}

		for i, o := range ordered {
			// Get the datetime prefix.
			op, err := getDatetimePrefix(o.Name())
			if err != nil {
				return nil, fmt.Errorf("error getting datetime prefix: %w", err)
			}

			// Parse the datetime prefix.
			parsed, err := time.Parse(FilePrefix, prefix)
			if err != nil {
				return nil, fmt.Errorf("error parsing datetime prefix: %w", err)
			}

			// Parse the datetime prefix.
			oparsed, err := time.Parse(FilePrefix, op)
			if err != nil {
				return nil, fmt.Errorf("error parsing datetime prefix: %w", err)
			}

			if parsed.Before(oparsed) {
				ordered = append(ordered[:i], append([]os.DirEntry{f}, ordered[i:]...)...)
				break
			} else if i == len(ordered)-1 {
				ordered = append(ordered, f)
				break
			}
		}
	}

	return ordered, nil
}

func getDatetimePrefix(name string) (string, error) {
	// Get the datetime prefix.
	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		return "", fmt.Errorf("error getting datetime prefix: %s", name)
	}

	return parts[0], nil
}

func getFiles(location string) ([]os.DirEntry, error) {
	// Is the location a directory?
	if !isDirectory(location) {
		return nil, ErrLocationIsNotDirectory
	}

	// Get all files in the directory.
	f, err := os.ReadDir(location)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	return f, nil
}

func isDirectory(location string) bool {
	s, err := os.Stat(location)
	if err != nil {
		slog.Error(fmt.Sprintf("error stating location: %s", location), slog.String("error", err.Error()))
		return false
	}

	return s.IsDir()
}

func filterFiles(files []os.DirEntry, ext string) []os.DirEntry {
	filtered := make([]os.DirEntry, 0, len(files))
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), strings.ToLower(ext)) {
			filtered = append(filtered, f)
		}
	}

	return filtered
}
