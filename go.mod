module github.com/broadinstitute/revere

go 1.16

require (
	// Resty simplifies REST API usage, remembering headers and unmarshalling to Go structs
	github.com/go-resty/resty/v2 v2.6.0
	// Mapstructure translates structures based on field names, helpful for API usage
	github.com/mitchellh/mapstructure v1.4.1
	// Cobra is a CLI framework facilitating help text, commands, and error checking
	github.com/spf13/cobra v1.1.3
	// Viper integrates with Cobra and handles configuration files
	github.com/spf13/viper v1.8.0
	// Validator recursively checks structs based on field tags
	gopkg.in/go-playground/validator.v9 v9.31.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
)
