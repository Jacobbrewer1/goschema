package main

import (
	"context"
	"flag"
	"log/slog"
	"path/filepath"

	"github.com/google/subcommands"
	generation2 "github.com/jacobbrewer1/goschema/pkg/generation"
)

type generateCmd struct {
	// templatesLocation is the location of the templates to use.
	templatesLocation string

	// outputLocation is the location to write the generated files to.
	outputLocation string

	// sqlLocation is the location of the SQL files to use.
	sqlLocation string

	// fileExtensionPrefix is the prefix to add to the generated file extension.
	fileExtensionPrefix string
}

func (g *generateCmd) Name() string {
	return "generate"
}

func (g *generateCmd) Synopsis() string {
	return "Generate GO types from a MySQL schema"
}

func (g *generateCmd) Usage() string {
	return `generate:
  Generate GO types from a MySQL schema.
`
}

func (g *generateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&g.templatesLocation, "templates", "./templates/*.tmpl", "The location of the templates to use.")
	f.StringVar(&g.outputLocation, "out", ".", "The location to write the generated files to.")
	f.StringVar(&g.sqlLocation, "sql", "./pkg/models/*.sql", "The location of the SQL files to use.")
	f.StringVar(&g.fileExtensionPrefix, "extension", "", "The prefix to add to the generated file extension.")
}

func (g *generateCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var err error
	g.outputLocation, err = filepath.Abs(g.outputLocation)
	if err != nil {
		slog.Error("Error getting absolute path", slog.String("outputLocation", g.outputLocation), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	// Load the SQL file locations as abs.
	g.sqlLocation, err = filepath.Abs(g.sqlLocation)
	if err != nil {
		slog.Error("Error getting absolute path", slog.String("sqlLocation", g.sqlLocation), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	tables, err := generation2.LoadSQL(g.sqlLocation)
	if err != nil {
		slog.Error("Error loading SQL", slog.String("templatesLocation", g.templatesLocation), slog.String("outputLocation", g.outputLocation), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	} else if len(tables) == 0 {
		slog.Info("No tables found", slog.String("sqlLocation", g.sqlLocation))
		return subcommands.ExitFailure
	}

	err = generation2.RenderTemplates(tables, g.templatesLocation, g.outputLocation, g.fileExtensionPrefix)
	if err != nil {
		slog.Error("Error rendering templates", slog.String("templatesLocation", g.templatesLocation), slog.String("outputLocation", g.outputLocation), slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
