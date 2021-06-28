package shared

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"

	"github.com/go-resty/resty/v2"
)

func BaseClient(config *configuration.Config) *resty.Client {
	return resty.New().
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(config.Client.Redirects)).
		SetRetryCount(config.Client.Retries)
}

func CheckResponse(response *resty.Response, err error) error {
	if err != nil {
		return err
	} else if response.StatusCode() < 200 || response.StatusCode() > 299 {
		return fmt.Errorf("%d from %s, response: %s", response.StatusCode(), response.Request.URL, response.String())
	} else {
		return nil
	}
}
