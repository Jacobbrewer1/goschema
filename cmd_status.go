package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strconv"

	"github.com/google/subcommands"
	"github.com/jacobbrewer1/goschema/pkg/logging"
	"github.com/jacobbrewer1/goschema/pkg/migrations"
	"github.com/pterm/pterm"
)

type statusCmd struct{}

func (c *statusCmd) Name() string {
	return "status"
}

func (c *statusCmd) Synopsis() string {
	return "Print the status of the database migrations."
}

func (c *statusCmd) Usage() string {
	return `status:
  Print the status of the database migrations.
`
}

func (c *statusCmd) SetFlags(f *flag.FlagSet) {}

func (c *statusCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	if e := os.Getenv(migrations.DbEnvVar); e == "" {
		slog.Error("Database environment variable not set",
			slog.String(logging.KeyVariable, migrations.DbEnvVar))
		return subcommands.ExitFailure
	}

	db, err := migrations.ConnectDB()
	if err != nil {
		slog.Error("Error connecting to the database",
			slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	versions, err := migrations.NewVersioning(db, "", 0).GetStatus()
	if err != nil {
		slog.Error("Error getting the status",
			slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	tableDataStr := make([][]string, 0)
	tableDataStr = append(tableDataStr, []string{"Version", "Current", "Created At"})
	for _, v := range versions {
		tableDataStr = append(tableDataStr, []string{v.Version, strconv.FormatBool(v.IsCurrent), v.CreatedAt.String()})
	}

	var tableData pterm.TableData = tableDataStr

	if err := pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render(); err != nil {
		slog.Error("Error rendering table",
			slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
