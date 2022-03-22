package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"github.com/gorilla/mux"

	"github.com/digicatapult/wasp-ingest-rtmp/services"
)

func main() {
	var rtmpURL string

	flag.StringVar(&rtmpURL, "rtmp", "default", "The url of the rtmp stream to ingest")
	flag.Parse()

	cfg := zap.NewDevelopmentConfig()
	if os.Getenv("ENV") == "production" {
		cfg = zap.NewProductionConfig()

		lvl, err := zap.ParseAtomicLevel(os.Getenv("LOG_LEVEL"))
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
		err := logger.Sync()
		if err != nil {
			log.Printf("error whilst syncing zap: %s\n", err)
		}
	}()

	zap.ReplaceGlobals(logger)

	addr := fmt.Sprintf(":%d", 9999)
	srv := &http.Server{Addr: addr}

	go services.ChunksToWebsocket(rtmpURL)

	router := mux.NewRouter()

	// for profiler
	router.HandleFunc(`/ws`, services.HandleWs)
	srv.Handler = router

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.S().With("error", err).Fatal("http.ListenAndServe")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case <-stop:
		zap.S().Info("closing down")
	}
}
