package cmd

import (
	"github.com/broadinstitute/revere/internal/actions"
	config2 "github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/statuspage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Configure Statuspage.io based on the configuration",
	Long: `This command diffs the Statuspage components and groups from
the configuration with what's present on the remote, matching
based on name alone. It then sequentially deletes,
creates, and patches resources such that a subsequent
diff would identify no changes.`,
	Run: Prepare,
}

func Prepare(*cobra.Command, []string) {
	config, err := config2.AssembleConfig(viper.GetViper())
	cobra.CheckErr(err)
	client := statuspage.Client(config)
	err = actions.ReconcileComponents(config, client)
	cobra.CheckErr(err)
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
