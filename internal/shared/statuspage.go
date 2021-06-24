package shared

import (
	"github.com/broadinstitute/terra-status-manager/pkg"
	"github.com/go-resty/resty/v2"
)

type StatuspageRequestComponent struct {
	Description        string
	GroupID            string `json:"group_id"`
	Name               string
	OnlyShowIfDegraded bool `json:"only_show_if_degraded"`
	Showcase           bool
	StartDate          string `json:"start_date"`
	Status             string
}

type StatuspageResponseComponent struct {
	AutomationEmail    string `json:"automation_email"`
	CreatedAt          string `json:"created_at"`
	Description        string
	GroupID            string `json:"group_id"`
	ID                 string
	Name               string
	OnlyShowIfDegraded bool   `json:"only_show_if_degraded"`
	PageID             string `json:"page_id"`
	Position           int
	Showcase           bool
	StartDate          string `json:"start_date"`
	Status             string
	UpdatedAt          string `json:"updated_at"`
}

func StatuspageClient(config *pkg.Config) *resty.Client {
	return baseClient(config).
		SetHostURL(config.Statuspage.ApiRoot).
		SetAuthScheme("OAuth").
		SetAuthToken(config.Statuspage.ApiKey).
		SetHeader("Accept", "application/json")
}
