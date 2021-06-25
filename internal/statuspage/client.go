package statuspage

import (
	"github.com/broadinstitute/terra-status-manager/internal/shared"
	"github.com/broadinstitute/terra-status-manager/pkg"
	"github.com/go-resty/resty/v2"
)

func StatuspageClient(config *pkg.Config) *resty.Client {
	return shared.BaseClient(config).
		SetHostURL(config.Statuspage.ApiRoot).
		SetAuthScheme("OAuth").
		SetAuthToken(config.Statuspage.ApiKey).
		SetHeader("Accept", "application/json")
}
