package shared

import "github.com/broadinstitute/terra-status-manager/pkg"

// LogLn prints each non-empty string if config.Verbose, only the first string otherwise.
func LogLn(config *pkg.Config, always string, verbose ...string) {
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
