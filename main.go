package main

import (
	"github.com/broadinstitute/revere/cmd"
	"github.com/broadinstitute/revere/internal/version"
)

// BuildVersion is intended for use with Go's LDFlags compiler option, to
// set this value at compile time
var BuildVersion = "development"

func main() {
	// Short-form LDFlags only work for top-level files, but we can only
	// import from interior packages later on, so manually set an interior
	// variable
	version.BuildVersion = BuildVersion

	cmd.Execute()
}
