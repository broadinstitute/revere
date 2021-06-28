package shared

import (
	"github.com/broadinstitute/revere/internal/configuration"
)

// LogLn prints each non-empty string if configuration.Verbose, only the first string otherwise.
func LogLn(config *configuration.Config, alwaysPrint string, verbosePrint ...string) {
	if alwaysPrint != "" {
		println(alwaysPrint)
	}
	if config.Verbose {
		for _, str := range verbosePrint {
			if str != "" {
				println(str)
			}
		}
	}
}
