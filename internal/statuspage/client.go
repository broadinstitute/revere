package statuspage

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/go-resty/resty/v2"
)

func Client(config *configuration.Config) *resty.Client {
	return shared.BaseClient(config).
		SetHostURL(config.Statuspage.ApiRoot).
		SetAuthScheme("OAuth").
		SetAuthToken(config.Statuspage.ApiKey).
		SetHeader("Accept", "application/json")
}
