package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Jacobbrewer1/goschema/pkg/migrations"
	"github.com/google/subcommands"
)

type createCmd struct {
	// name is the name of the migration to create.
	name string

	// OutputLocation is the location to write the generated files to.
	outputLocation string
}

func (c *createCmd) Name() string {
	return "create"
}

func (c *createCmd) Synopsis() string {
	return "Create a new migration"
}

func (c *createCmd) Usage() string {
	return `create:
  Create a new migration.
`
}

func (c *createCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.name, "name", "", "The name of the migration to create.")
	f.StringVar(&c.outputLocation, "out", ".", "The location to write the generated files to.")
}

func (c *createCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if c.name == "" {
		slog.Error("Name is required")
		return subcommands.ExitUsageError
	}

	if c.outputLocation == "" {
		slog.Error("Output location is required")
		return subcommands.ExitUsageError
	}

	// File name is timestamp_name.up.sql and timestamp_name.down.sql
	// The timestamp is the current time in the format YYYYMMDDHHMMSS
	// The name is the name of the migration with spaces as underscores

	now := time.Now().UTC()
	name := fmt.Sprintf("%s_%s", now.Format(migrations.FilePrefix), strings.TrimSpace(c.name))
	name = strings.ReplaceAll(name, " ", "_")

	upName := fmt.Sprintf("%s.up.sql", name)
	downName := fmt.Sprintf("%s.down.sql", name)

	upPath := fmt.Sprintf("%s/%s", c.outputLocation, upName)
	downPath := fmt.Sprintf("%s/%s", c.outputLocation, downName)

	upAbs, err := filepath.Abs(upPath)
	if err != nil {
		slog.Error("Error getting absolute path", slog.String("path", upPath), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	downAbs, err := filepath.Abs(downPath)
	if err != nil {
		slog.Error("Error getting absolute path", slog.String("path", downPath), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	if err := createFile(upAbs); err != nil {
		slog.Error("Error creating file", slog.String("path", upAbs), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	slog.Info("Up migration created", slog.String("path", upAbs))

	if err := createFile(downAbs); err != nil {
		slog.Error("Error creating file", slog.String("path", downAbs), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	slog.Info("Down migration created", slog.String("path", downAbs))

	return subcommands.ExitSuccess
}

func createFile(name string) error {
	// Create the path if it does not exist.
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating path: %w", err)
	}

	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	return f.Close()
}
