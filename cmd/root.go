package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "os"

    "github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "terra-status-manager",
    Short: "Interact with Terra's production Statuspage",
    Long: `Requires a configuration file via --config, ./terra-status-manager.yaml,
or /etc/terra-status-manager/terra-status-manager.yaml.

To configure Statuspage.io based on the config file:
	$ terra-status-manager prepare`,
}

// Execute adds all child commands to the root command, only necessary for rootCmd
func Execute() {
    cobra.CheckErr(rootCmd.Execute())
}

// init configures flags (both persistent across child commands and local to root)
func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./terra-status-manager.yaml)")
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
        viper.AddConfigPath("/etc/terra-status-manager/")
        viper.SetConfigName("terra-status-manager")
    }

    viper.SetEnvPrefix("TSM")
    viper.AutomaticEnv()

    err := viper.ReadInConfig()
    if err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            _, err := fmt.Fprintln(os.Stderr, "Not using a configuration file!")
            cobra.CheckErr(err)
        }
        cobra.CheckErr(err)
    } else {
        _, err := fmt.Println("Using configuration file:", viper.ConfigFileUsed())
        cobra.CheckErr(err)
    }
}
