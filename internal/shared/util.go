package shared

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
)

// LogLn prints each non-empty string if configuration.Verbose, only the first string otherwise.
func LogLn(config *configuration.Config, alwaysPrint string, verbosePrint ...string) {
	if alwaysPrint != "" {
		fmt.Println(alwaysPrint)
	}
	if config.Verbose {
		for _, str := range verbosePrint {
			if str != "" {
				fmt.Println(str)
			}
		}
	}
}
