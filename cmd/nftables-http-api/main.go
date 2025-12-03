package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/tacerus/nftables-http-api/core"
	"github.com/tacerus/nftables-http-api/server"
)

func main() {
	var (
		configArg   string
		logLevelArg string
	)

	flag.StringVar(&configArg, "config", "./config.json", "Configuration file")
	flag.StringVar(&logLevelArg, "loglevel", "info", "Logging level")

	flag.Parse()

	slog.SetDefault(newSlog(newLogLevel(logLevelArg)))

	slog.Info("Booting server ...")

	app := server.NewApp(core.NewConfig(configArg))
	if app == nil {
		os.Exit(1)
	}

	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt)

	srv := app.Start()
	defer srv.Shutdown(context.Background())

main:
	for {
		select {
		case <-cs:
			slog.Debug("Received interrupt")
			break main
		}
	}

	slog.Info("Shutting down ...")
}
