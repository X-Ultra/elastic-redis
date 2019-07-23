package cluster

import (
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	DataPath string
	LogsPath string

	// DevMode is used to enable a development server mode.
	DevMode bool

	// Bootstrap mode is used to bring up the first Consul server.
	// It is required so that it can elect a leader without any
	// other nodes being present
	Bootstrap bool
}

func (c *Config) Init() {
	if c.DevMode {
		return
	}

	absDataPath, err := filepath.Abs(c.DataPath)
	absLogsPath, err2 := filepath.Abs(c.LogsPath)
	if err != nil {
		log.Fatalf("failed to determine absolute data path: %s", err.Error())
	}
	if err := os.MkdirAll(absDataPath, 0755); err != nil {
		log.Fatalf("failed to determine absolute data path: %s", err.Error())
	}

	if err2 != nil {
		log.Fatalf("failed to determine absolute logs path: %s", err2.Error())
	}
	if err := os.MkdirAll(absLogsPath, 0755); err != nil {
		log.Fatalf("failed to determine absolute logs path: %s", err2.Error())
	}
}
