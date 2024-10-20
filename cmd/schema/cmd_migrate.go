package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/subcommands"
	"github.com/jacobbrewer1/goschema/pkg/migrations"
)

type migrateCmd struct {
	// up is the flag to migrate up.
	up bool

	// down is the flag to migrate down.
	down bool

	// migrationLocation is where the migrations are located.
	migrationLocation string

	// steps is the number of steps to migrate.
	steps int
}

func (m *migrateCmd) Name() string {
	return "migrate"
}

func (m *migrateCmd) Synopsis() string {
	return "Migrate the database"
}

func (m *migrateCmd) Usage() string {
	return `migrate:
  Migrate the database.
`
}

func (m *migrateCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&m.up, "up", false, "Migrate up.")
	f.BoolVar(&m.down, "down", false, "Migrate down.")
	f.StringVar(&m.migrationLocation, "loc", "./migrations", "The location of the migrations.")
	f.IntVar(&m.steps, "steps", 0, "The number of steps to migrate (0 means all).")
}

func (m *migrateCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if m.up && m.down {
		slog.Error("Cannot migrate up and down at the same time")
		return subcommands.ExitUsageError
	} else if !m.up && !m.down {
		slog.Error("Must specify up or down")
		return subcommands.ExitUsageError
	}

	if e := os.Getenv(migrations.DbEnvVar); e == "" {
		slog.Error(fmt.Sprintf("Environment variable %s is not set", migrations.DbEnvVar))
		return subcommands.ExitFailure
	}

	absPath, err := filepath.Abs(m.migrationLocation)
	if err != nil {
		slog.Error("Error getting absolute path", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	db, err := migrations.ConnectDB()
	if err != nil {
		slog.Error("Error connecting to the database", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	switch {
	case m.up:
		if err := migrations.NewVersioning(db, absPath, m.steps).MigrateUp(); err != nil {
			slog.Error("Error migrating up", slog.String("error", err.Error()))
			return subcommands.ExitFailure
		}
	case m.down:
		if err := migrations.NewVersioning(db, absPath, m.steps).MigrateDown(); err != nil {
			slog.Error("Error migrating down", slog.String("error", err.Error()))
			return subcommands.ExitFailure
		}
	}

	slog.Info("Migration complete")

	return subcommands.ExitSuccess
}
