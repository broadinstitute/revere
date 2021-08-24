package cmd

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/pubsub"
	"github.com/broadinstitute/revere/internal/pubsub/pubsubapi"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sync"
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
	client, err := pubsubapi.Client(config)
	cobra.CheckErr(err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		shared.LogLn(config, "listening to pubsub...")
		err = pubsub.ReceiveMessages(config, client)
		cobra.CheckErr(err)
		wg.Done()
	}()
	wg.Wait()
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
