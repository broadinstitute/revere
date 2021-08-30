package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/broadinstitute/revere/internal/api"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/pubsub"
	"github.com/broadinstitute/revere/internal/pubsub/pubsubapi"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Continuously translate input events to output communications",
	Long: `Continuously read input event sources and notify output 
communication channels as described in the configuration file.

Input event sources:
	- Google Cloud Monitoring via Google Cloud Pub/Sub`,
	Run: Serve,
}

func Serve(*cobra.Command, []string) {
	config, err := configuration.AssembleConfig(viper.GetViper())
	cobra.CheckErr(err)

	pubsubClient, err := pubsubapi.Client(config)
	cobra.CheckErr(err)
	pubsubCtx, cancelPubsub := context.WithCancel(context.Background())

	apiServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Api.Port),
		Handler: api.NewRouter(config),
	}

	routines := []struct {
		runForever   func()
		uponShutdown func()
	}{
		{
			runForever: func() {
				shared.LogLn(config, "listening to pubsub...")
				err := pubsub.ReceiveMessages(config, pubsubClient, pubsubCtx)
				cobra.CheckErr(err)
			},
			uponShutdown: func() {
				cancelPubsub()
			},
		},
		{
			runForever: func() {
				shared.LogLn(config, fmt.Sprintf("serving api on port %d...", config.Api.Port))
				if err := apiServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
					cobra.CheckErr(err)
				}
			},
			uponShutdown: func() {
				apiShutdownCtx, apiShutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer apiShutdownCancel()
				err := apiServer.Shutdown(apiShutdownCtx)
				cobra.CheckErr(err)
			},
		},
	}

	// Run continuous routines forever
	for _, routine := range routines {
		go routine.runForever()
	}

	// Block waiting for SIGINT/SIGTERM
	// We can't capture SIGKILL so no need to include
	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownChannel

	// Run shutdown routines "forever", use waitgroup to synchronize
	// (
	shared.LogLn(config, "shutting down...")
	var shutdownWaitGroup sync.WaitGroup
	for _, routine := range routines {
		shutdownWaitGroup.Add(1)
		// Intermediate variable to lock in routine referenced in func
		r := routine
		go func() {
			defer shutdownWaitGroup.Done()
			r.uponShutdown()
		}()
	}
	shutdownWaitGroup.Wait()
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
