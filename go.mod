module github.com/broadinstitute/revere

go 1.16

// List direct dependencies (here, not inline, so as to not confuse IDEs):
// Cobra is a CLI framework facilitating help text, commands, and error checking
// Httpmock works with Resty to mock APIs
// Mapstructure translates structures based on field names, helpful for API usage
// Resty simplifies REST API usage, remembering headers and unmarshalling to Go structs
// Validator recursively checks structs based on field tags
// Viper integrates with Cobra and handles configuration files
require (
	github.com/go-resty/resty/v2 v2.6.0
	github.com/jarcoal/httpmock v1.0.8
	github.com/mitchellh/mapstructure v1.4.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.0
	gopkg.in/go-playground/validator.v9 v9.31.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)
