package shared

import "github.com/broadinstitute/revere/internal/configuration"

func ExampleLogLn_verbose() {
	LogLn(&configuration.Config{
		Verbose: true,
	}, "always", "verbose 1", "verbose 2")
	// Output:
	// always
	// verbose 1
	// verbose 2
}

func ExampleLogLn_concise() {
	LogLn(&configuration.Config{
		Verbose: false,
	}, "always", "verbose 1", "verbose 2")
	// Output:
	// always
}
