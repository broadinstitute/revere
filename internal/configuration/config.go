package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/validator.v2"
)

/*
Config schema

All values:
- may be set in revere.yaml, see cmd/root.go:
  ```yaml
  client:
    redirects: 2
  ```

Some values:
- may be overridden via command line flags, noted below.
- may be overridden via environment variables, noted below and set in readEnvironmentVariables().
- may have non-"zero" default values, noted below and set in newDefaultConfig().
- may be required to be non-"zero", noted below and validated in AssembleConfig().
*/
type Config struct {
	// Whether to be more verbose with console output
	// NOTE: May be set via --verbose / -v command line flags
	Verbose bool

	Client struct {
		// Number of 300-series redirects to follow
		Redirects int // default: 3
		// Number of exponential-backoff retries to make
		Retries int // default: 3
	}

	Statuspage struct {
		// API key to communicate with Statuspage.io
		// NOTE: May be set via REVERE_STATUSPAGE_APIKEY in environment
		ApiKey string `validate:"nonzero"`
		// ID of the particular page to interact with
		PageID     string `validate:"nonzero"`
		ApiRoot    string // default: "https://api.statuspage.io/v1"
		Components []Component
	}
}

// Component configuration--note that leaving any of the below unfilled will use Go's "zero" value (false/empty)
type Component struct {
	// Unique but user-readable component name
	Name        string `validate:"nonzero"`
	Description string
	// If the component should be hidden to users while operational
	OnlyShowIfDegraded bool
	// If uptime data should be hidden and go unrecorded
	HideUptime bool
	// Date the component existed from, in the form YYYY-MM-DD
	StartDate string `validate:"nonzero"`
}

// newDefaultConfig sets config defaults only as described above
func newDefaultConfig() *Config {
	var config Config
	config.Client.Redirects = 3
	config.Client.Retries = 3
	config.Statuspage.ApiRoot = "https://api.statuspage.io/v1"
	return &config
}

// readEnvironmentVariables sets config values from the environment specifically only as described above
func readEnvironmentVariables(config *Config) {
	apiKey, present := os.LookupEnv("REVERE_STATUSPAGE_APIKEY")
	if present {
		config.Statuspage.ApiKey = apiKey
	}
}

// AssembleConfig creates a default config, reads values from Viper's config file,
// reads applies overrides from the environment, and validates the config before returning
func AssembleConfig(v *viper.Viper) (*Config, error) {
	config := newDefaultConfig()
	err := v.Unmarshal(config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling Viper to configuration struct: %w", err)
	}
	readEnvironmentVariables(config)
	err = validator.Validate(config)
	if err != nil {
		return nil, fmt.Errorf("errors validating configuration: %w", err)
	}
	return config, nil
}
