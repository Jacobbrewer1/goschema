package goschema

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/subcommands"
)

func Run() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(new(versionCmd), "")
	subcommands.Register(new(generateCmd), "")
	subcommands.Register(new(createCmd), "")
	subcommands.Register(new(migrateCmd), "")

	flag.Parse()

	// Listen for ctrl+c and kill signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		got := <-sig
		slog.Info("Received signal, shutting down", slog.String("signal", got.String()))
		cancel()
	}()

	os.Exit(int(subcommands.Execute(ctx)))
}
