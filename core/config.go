package core

import (
	"encoding/json"
	"log/slog"
	"os"
)

type ConfigTokenPaths map[string][]string
type ConfigTokens map[string]ConfigTokenPaths

type Config struct {
	Bind   string
	Tokens ConfigTokens
}

func NewConfig(file string) Config {
	c := new(Config)

	fh, err := os.Open(file)
	if err != nil {
		slog.Error("Failed to open configuration file", "error", err)
		os.Exit(1)
	}
	defer fh.Close()

	if err := json.NewDecoder(fh).Decode(&c); err != nil {
		slog.Error("Failed to parse configuration file", "error", err)
		os.Exit(1)
	}

	return *c
}
