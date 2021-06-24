package cmd

import (
	"github.com/broadinstitute/terra-status-manager/internal"
	"github.com/broadinstitute/terra-status-manager/internal/shared"
	"github.com/broadinstitute/terra-status-manager/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := pkg.AssembleConfig(viper.GetViper())
		cobra.CheckErr(err)
		client := shared.StatuspageClient(config)
		err = internal.InstantiateComponents(config, client)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
