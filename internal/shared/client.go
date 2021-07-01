package shared

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"

	"github.com/go-resty/resty/v2"
)

// BaseClient should configure Resty "globally", not in any service-dependent way
func BaseClient(config *configuration.Config) *resty.Client {
	return resty.New().
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(config.Client.Redirects)).
		SetRetryCount(config.Client.Retries)
}

// CheckResponse returns an error if the response wasn't successful.
// This helps us safely call response.Result() when err is nil, because a non-200-series response
// normally does not produce any sort of error but the Result() will be empty.
func CheckResponse(response *resty.Response, err error) error {
	if err != nil {
		return err
	} else if response.StatusCode() < 200 || response.StatusCode() > 299 {
		return fmt.Errorf("%d from %s, response: %s", response.StatusCode(), response.Request.URL, response.String())
	} else {
		return nil
	}
}
