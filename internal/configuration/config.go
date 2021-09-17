package configuration

import (
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"os"
	"strconv"

	"github.com/spf13/viper"
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
		ApiKey string `validate:"required"`
		// ID of the particular page to interact with
		PageID     string           `validate:"required"`
		ApiRoot    string           // default: "https://api.statuspage.io/v1"
		Components []Component      `validate:"unique=Name,dive"`
		Groups     []ComponentGroup `validate:"unique=Name,dive"`
	}

	Pubsub struct {
		// Non-numeric ID of the GCP project containing the subscription
		ProjectID string `validate:"required"`
		// ID of the Cloud Pub/Sub subscription to use to pull messages
		SubscriptionID string `validate:"required"`
	}

	Api struct {
		// Port to host Revere's web server on
		// NOTE: May be set via REVERE_API_PORT in environment
		Port int // default: 8080
		// Print debugging information from the server library
		Debug bool
		// Forcibly silence the request log
		Silent bool
	}

	// Correlate developed services to user-facing components
	ServiceToComponentMapping []ServiceToComponentMapping
}

// Component configuration--note that leaving any of the below unfilled will use Go's "zero" value (false/empty)
type Component struct {
	// Unique but user-readable component name
	Name        string `validate:"required"`
	Description string
	// If the component should be hidden to users while operational
	OnlyShowIfDegraded bool
	// If uptime data should be hidden and go unrecorded
	HideUptime bool
	// Date the component existed from, in the form YYYY-MM-DD
	StartDate string `validate:"required"`
}

// ComponentGroup configuration--note that leaving any of the below unfilled will use Go's "zero" value (false/empty)
type ComponentGroup struct {
	// Unique but user-readable group name
	Name        string `validate:"required"`
	Description string
	// Exact names of components to include in the group (components should never exist in more than one group)
	ComponentNames []string `validate:"required,unique"`
}

// ServiceToComponentMapping correlates developed services ("Rawls", "Leonardo") in particular environments ("prod")
// to user-facing components ("Notebooks", "Terra UI")
type ServiceToComponentMapping struct {
	ServiceName            string   `validate:"required"`
	EnvironmentName        string   `validate:"required"`
	AffectsComponentsNamed []string `validate:"unique"`
}

// newDefaultConfig sets config defaults only as described above
func newDefaultConfig() *Config {
	var config Config
	config.Client.Redirects = 3
	config.Client.Retries = 3
	config.Statuspage.ApiRoot = "https://api.statuspage.io/v1"
	config.Api.Port = 8080
	return &config
}

// readEnvironmentVariables sets config values from the environment specifically only as described above
func readEnvironmentVariables(config *Config) error {
	apiKey, present := os.LookupEnv("REVERE_STATUSPAGE_APIKEY")
	if present {
		config.Statuspage.ApiKey = apiKey
	}
	stringPort, present := os.LookupEnv("REVERE_API_PORT")
	if present {
		intPort, err := strconv.Atoi(stringPort)
		if err != nil {
			return err
		}
		config.Api.Port = intPort
	}
	return nil
}

// secondaryConfigValidation performs logical validation that can't be captured by struct tags
func secondaryConfigValidation(config *Config) error {
	// Go compiler optimized to use map[string]struct{} like a Set (no alloc for values)
	componentNames := make(map[string]struct{})
	for _, component := range config.Statuspage.Components {
		componentNames[component.Name] = struct{}{}
	}
	for _, serviceMapping := range config.ServiceToComponentMapping {
		for _, componentName := range serviceMapping.AffectsComponentsNamed {
			if _, present := componentNames[componentName]; !present {
				return fmt.Errorf("mapping for service %s affects non-existent component %s",
					serviceMapping.ServiceName, componentName)
			}
		}
	}
	return nil
}

// AssembleConfig creates a default config, reads values from Viper's config file,
// reads applies overrides from the environment, and validates the config before returning
func AssembleConfig(v *viper.Viper) (*Config, error) {
	config := newDefaultConfig()
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshalling Viper to configuration struct: %w", err)
	}
	if err := readEnvironmentVariables(config); err != nil {
		return nil, fmt.Errorf("error reading environment variables: %w", err)
	}
	if err := validator.New().Struct(config); err != nil {
		return nil, fmt.Errorf("errors validating configuration: %w", err)
	}
	if err := secondaryConfigValidation(config); err != nil {
		return nil, fmt.Errorf("errors during secondary configuration validation: %w", err)
	}
	return config, nil
}
