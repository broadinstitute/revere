package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "revere",
	Short: "Interact with Terra's production Statuspage",
	Long: `Interact with Terra's production Statuspage.

Requires a configuration file via --config, ./revere.yaml,
or /etc/revere/revere.yaml.

To configure Statuspage.io based on the config file:
	$ revere prepare`,
}

// Execute adds all child commands to the root command, only necessary for rootCmd
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// init configures flags (both persistent across child commands and local to root)
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./revere.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable more verbose console output")
	err := viper.BindPFlags(rootCmd.Flags())
	cobra.CheckErr(err)
}

// initConfig reads in config file (flag or default) and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/revere/")
		viper.SetConfigName("revere")
	}

	// example: statuspage.apiKey overridden by env var REVERE_STATUSPAGE_APIKEY
	viper.SetEnvPrefix("REVERE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			_, err := fmt.Fprintln(os.Stderr, "Not using a configuration file!")
			// err here is Fprintln's, so probably nil--we intentionally don't always exit
			// here since Viper can be configured other ways and we validate config later
			cobra.CheckErr(err)
		} else {
			cobra.CheckErr(err)
		}
	} else {
		_, err := fmt.Println("Using configuration file:", viper.ConfigFileUsed())
		cobra.CheckErr(err)
	}
}
