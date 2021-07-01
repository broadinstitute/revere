package statuspagetypes

import "github.com/broadinstitute/revere/internal/configuration"

// Component represents how Statuspage returns components in its API.
// This is an exact superset of what Statuspage accepts as input for components,
// which is in turn an exact superset of pkg.Component.
type Component struct {
	AutomationEmail    string `json:"automation_email"`
	CreatedAt          string `json:"created_at"`
	Description        string `json:"description"`
	Group              bool   `json:"group"`
	GroupID            string `json:"group_id"`
	ID                 string `json:"id"`
	Name               string `json:"name"`
	OnlyShowIfDegraded bool   `json:"only_show_if_degraded"`
	PageID             string `json:"page_id"`
	Position           int    `json:"position"`
	Showcase           bool   `json:"showcase"`
	StartDate          string `json:"start_date"`
	Status             string `json:"status"`
	UpdatedAt          string `json:"updated_at"`
}

// MergeConfigComponentToApi overwrites fields of the Component with that of the one from the config.
// Does not create a new Component so that it can be used for merging and difference finding.
func MergeConfigComponentToApi(configComponent configuration.Component, apiComponent *Component) {
	apiComponent.Name = configComponent.Name
	apiComponent.Description = configComponent.Description
	apiComponent.OnlyShowIfDegraded = configComponent.OnlyShowIfDegraded
	apiComponent.StartDate = configComponent.StartDate
	apiComponent.Showcase = !configComponent.HideUptime
}

// RequestComponent represents what Statuspage accepts as input for components.
// This is necessary because Statuspage errors if unexpected keys are present
// in request JSON (???) so we must reduce Component down to this type.
type RequestComponent struct {
	Description        string `json:"description"`
	GroupID            string `json:"group_id,omitempty"`
	Name               string `json:"name,omitempty"`
	OnlyShowIfDegraded bool   `json:"only_show_if_degraded"`
	Showcase           bool   `json:"showcase"`
	StartDate          string `json:"start_date,omitempty"`
	Status             string `json:"status,omitempty"`
}

// ToRequest converts Component to requestComponent in a type-safe way to avoid
// needing to handle mapstructure.decode(...) errors
func (c *Component) ToRequest() RequestComponent {
	return RequestComponent{
		Description:        c.Description,
		Status:             c.Status,
		Name:               c.Name,
		OnlyShowIfDegraded: c.OnlyShowIfDegraded,
		GroupID:            c.GroupID,
		Showcase:           c.Showcase,
		StartDate:          c.StartDate,
	}
}
