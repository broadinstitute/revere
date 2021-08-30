package cmd

import (
	"github.com/broadinstitute/revere/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get Revere's recorded build version",
	Long: `Get Revere's internal BuildVersion, usually set via LDFlags during
compilation.`,
	Run: func(cmd *cobra.Command, args []string) {
		println()
		println("version:")
		println(version.BuildVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
