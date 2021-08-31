package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "revere",
	Short: "Communicate Terra's production status and uptime",
	Long: `Bridges input events about Terra components and output communication
channels to notify the public about Terra's status and uptime.

Current input event sources:
	- Google Cloud Monitoring via Google Cloud Pub/Sub
Current output communication channels:
	- Atlassian Statuspage.io

Requires a configuration file via --configuration, ./revere.yaml,
or /etc/revere/revere.yaml.

To prepare input and output services for Revere's operation:
	$ revere prepare

To run Revere continuously:
	$ revere serve

See subcommand help for more information.`,
}

// Execute adds all child commands to the root command, only necessary for rootCmd
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// init configures flags (both persistent across child commands and local to root)
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "configuration", "", "configuration file (default is ./revere.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable more verbose console output")
	err := viper.BindPFlags(rootCmd.Flags())
	cobra.CheckErr(err)
}

// initConfig reads in configuration file (flag or default) and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/revere/")
		viper.SetConfigName("revere")
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Don't error out since "version" command requires no config file
			fmt.Printf("not using a configuration file! %v\n", err)
		} else {
			cobra.CheckErr(err)
		}
	} else {
		_, err := fmt.Println("using configuration file:", viper.ConfigFileUsed())
		cobra.CheckErr(err)
	}
}
