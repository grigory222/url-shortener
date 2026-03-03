package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/grigory/url-shortener/internal/app"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config file")
	flag.Parse()

	if err := app.Run(*configPath); err != nil {
		slog.Error("startup error", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
