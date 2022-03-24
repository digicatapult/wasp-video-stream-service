package util

import (
	"log"

	"go.uber.org/zap"
)

// ConfigureLogging will setup the logging based on environment variables
func ConfigureLogging() {
	cfg := zap.NewDevelopmentConfig()
	if GetEnv(EnvKey, "") == "production" {
		cfg = zap.NewProductionConfig()

		lvl, err := zap.ParseAtomicLevel(GetEnv(LogLevelKey, "debug"))
		if err != nil {
			panic("invalid log level")
		}

		log.Printf("setting level: %s", lvl.String())

		cfg.Level = lvl
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("error initializing the logger")
	}

	defer func() {
		syncErr := logger.Sync()
		if err != nil {
			log.Printf("error whilst syncing zap: %s\n", syncErr)
		}
	}()

	zap.ReplaceGlobals(logger)
}
