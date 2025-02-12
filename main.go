package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/subcommands"
	"github.com/jacobbrewer1/goschema/pkg/logging"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(new(versionCmd), "")
	subcommands.Register(new(generateCmd), "")
	subcommands.Register(new(createCmd), "")
	subcommands.Register(new(migrateCmd), "")
	subcommands.Register(new(statusCmd), "")

	flag.Parse()

	// Listen for ctrl+c and kill signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		got := <-sig
		slog.Info("Received signal, shutting down", slog.String(logging.KeySignal, got.String()))
		cancel()
	}()

	os.Exit(int(subcommands.Execute(ctx))) // nolint: gocritic
}
