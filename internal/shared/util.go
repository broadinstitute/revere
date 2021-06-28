package shared

import (
	"github.com/broadinstitute/revere/internal/configuration"
)

// LogLn prints each non-empty string if configuration.Verbose, only the first string otherwise.
func LogLn(config *configuration.Config, always string, verbose ...string) {
	if always != "" {
		println(always)
	}
	if config.Verbose {
		for _, s := range verbose {
			if s != "" {
				println(s)
			}
		}
	}
}
