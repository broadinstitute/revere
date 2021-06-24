package pkg

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/validator.v2"
)

/*
Config schema

All values:
- may be set in terra-status-manager.yaml, see cmd/root.go:
  ```yaml
  client:
    redirects: 2
  ```
- may be overridden via environment variables:
  ```bash
  TSM_CLIENT_REDIRECTS=2
  ```

Some values:
- may be overridden via command line flags, noted below.
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
		// NOTE: Required but sensitive, should be set via TSM_STATUSPAGE_APIKEY=...
		ApiKey string `validate:"nonzero"`

		// ID of the particular page to interact with
		PageID string `validate:"nonzero"`

		ApiRoot string // default: "https://api.statuspage.io/v1"

		Components []StatuspageComponent
	}
}

type StatuspageComponent struct {
	// Unique but user-readable component name
	Name               string `validate:"nonzero"`
	Description        string
	GroupID            string
	OnlyShowIfDegraded bool
	Showcase           bool
	StartDate          string
}

func newDefaultConfig() *Config {
	var c Config
	c.Client.Redirects = 3
	c.Client.Retries = 3
	c.Statuspage.ApiRoot = "https://api.statuspage.io/v1"
	return &c
}

func AssembleConfig(v *viper.Viper) (*Config, error) {
	c := newDefaultConfig()
	err := v.Unmarshal(c)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling Viper to config struct: %w", err)
	}
	err = validator.Validate(c)
	if err != nil {
		return nil, fmt.Errorf("errors validating config: %w", err)
	}
	return c, nil
}
