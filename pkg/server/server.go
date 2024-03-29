package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nightfury1204/movie-listing-app/routes"

	"github.com/nightfury1204/movie-listing-app/pkg/logger"
	"github.com/nightfury1204/movie-listing-app/pkg/omdb"
	"github.com/pkg/errors"
)

// Run will start the server
func Run(cfg *Config) error {
	if cfg == nil {
		return errors.New("provided config is nil")
	}

	log := logger.GetLogger()

	// initialize omdb client
	omdb.Initialize(cfg.OMDBAPIUrl, cfg.OMDBAPIToken)

	handler := routes.NewMacaron()
	log.Info("Registering routes")
	routes.RegisterRoutes(handler)

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", cfg.Port),
		Handler: handler,
	}

	// notify the interrupt signal
	signal.Notify(cfg.StopCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	log.Info("Server Started", "listening port", cfg.Port)

	<-cfg.StopCh
	log.Info("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		panic(fmt.Sprintf("Server Shutdown Failed:%v", err))
	}
	log.Info("Server Exited Properly")
	return nil
}
