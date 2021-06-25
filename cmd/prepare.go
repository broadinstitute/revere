package cmd

import (
	"github.com/broadinstitute/terra-status-manager/internal"
	"github.com/broadinstitute/terra-status-manager/internal/statuspage"
	"github.com/broadinstitute/terra-status-manager/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Configure Statuspage.io based on the config",
	Long: `This command diffs the Statuspage components and groups from
the config with what's present on the remote, matching
based on name alone. It then sequentially deletes,
creates, and patches resources such that a subsequent
diff would identify no changes.`,
	Run: Prepare,
}

func Prepare(*cobra.Command, []string) {
	config, err := pkg.AssembleConfig(viper.GetViper())
	cobra.CheckErr(err)
	client := statuspage.Client(config)
	err = internal.ReconcileComponents(config, client)
	cobra.CheckErr(err)
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
