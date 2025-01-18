package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
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

func (c *statusCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return subcommands.ExitSuccess
}
