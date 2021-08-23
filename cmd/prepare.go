package cmd

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/broadinstitute/revere/internal/statuspage"
	"github.com/broadinstitute/revere/internal/statuspage/statuspageapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare input and output services for Revere's operation",
	Long: `Configure and check input event sources and output communication
channels for Revere to subsequently run.

Contents:
	- Configure Statuspage.io to display Terra components as described in the
configuration file`,
	Run: Prepare,
}

func Prepare(*cobra.Command, []string) {
	config, err := configuration.AssembleConfig(viper.GetViper())
	cobra.CheckErr(err)
	client := statuspageapi.Client(config)
	shared.LogLn(config, "reconciling components...")
	err = statuspage.ReconcileComponents(config, client)
	cobra.CheckErr(err)
	shared.LogLn(config, "reconciling groups...")
	err = statuspage.ReconcileGroups(config, client)
	cobra.CheckErr(err)
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
