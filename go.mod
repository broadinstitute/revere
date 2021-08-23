module github.com/broadinstitute/revere

go 1.16

// List direct dependencies (here, not inline, so as to not confuse IDEs):
// Cobra is a CLI framework facilitating help text, commands, and error checking
// Httpmock works with Resty to mock APIs
// Mapstructure translates structures based on field names, helpful for API usage
// Pubsub connects with Google Pub/Sub for input events
// Resty simplifies REST API usage, remembering headers and unmarshalling to Go structs
// Validator recursively checks structs based on field tags
// Viper integrates with Cobra and handles configuration files
require (
	cloud.google.com/go v0.93.3 // indirect
	cloud.google.com/go/kms v0.1.0 // indirect
	cloud.google.com/go/pubsub v1.15.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-resty/resty/v2 v2.6.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/jarcoal/httpmock v1.0.8
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.0
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210821163610-241b8fcbd6c8 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
)
