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
	"github.com/broadinstitute/revere/internal/state"
	"github.com/broadinstitute/revere/internal/statuspage"
	"github.com/broadinstitute/revere/internal/statuspage/statuspageapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
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

	shared.LogLn(config, "preparing statuspage..")
	statuspageClient := statuspageapi.Client(config)
	statuspageComponents, err := statuspageapi.GetComponents(statuspageClient, config.Statuspage.PageID)
	cobra.CheckErr(err)
	componentNamesToIDs := make(map[string]string)
	for _, component := range *statuspageComponents {
		componentNamesToIDs[component.Name] = component.ID
	}

	shared.LogLn(config, "preparing pubsub...")
	pubsubClient, err := pubsubapi.Client(config)
	cobra.CheckErr(err)
	pubsubCtx, cancelPubsub := context.WithCancel(context.Background())

	shared.LogLn(config, "preparing api...")
	apiServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Api.Port),
		Handler: api.NewRouter(config),
	}

	shared.LogLn(config, "deriving state...")
	appState := &state.State{}
	appState.Seed(componentNamesToIDs)

	// Routines to run in parallel
	routines := []struct {
		runForever   func()
		uponShutdown func() error
	}{
		{
			runForever: func() {
				shared.LogLn(config, "listening to pubsub...")
				err := pubsub.ReceiveMessages(config, pubsubClient, pubsubCtx,
					// StatusUpdater returns a function to update the status for one component;
					// ReceiveMessages will call that function as messages are handled
					statuspage.StatusUpdater(config, appState, statuspageClient))
				cobra.CheckErr(err)
			},
			uponShutdown: func() error {
				cancelPubsub()
				return nil
			},
		},
		{
			runForever: func() {
				shared.LogLn(config, fmt.Sprintf("serving api on port %d...", config.Api.Port))
				if err := apiServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
					cobra.CheckErr(err)
				}
			},
			uponShutdown: func() error {
				apiShutdownCtx, apiShutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer apiShutdownCancel()
				err := apiServer.Shutdown(apiShutdownCtx)
				return err
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

	// Run shutdown routines "forever", use errgroup to synchronize
	// Errgroup collects errors instead of exiting immediately
	shared.LogLn(config, "shutting down...")
	shutdownErrorGroup := new(errgroup.Group)
	for _, routine := range routines {
		shutdownErrorGroup.Go(routine.uponShutdown)
	}
	cobra.CheckErr(shutdownErrorGroup.Wait())
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
