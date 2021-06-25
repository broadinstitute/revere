package statuspage

import (
	"fmt"
	"github.com/broadinstitute/terra-status-manager/internal/shared"
	"github.com/broadinstitute/terra-status-manager/pkg"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
)

// Component represents how Statuspage returns components in its API.
// This is an exact superset of what Statuspage accepts as input for components,
// which is in turn an exact superset of pkg.Component.
type Component struct {
	AutomationEmail    string `json:"automation_email,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	Description        string `json:"description,omitempty"`
	Group              bool   `json:"group"`
	GroupID            string `json:"group_id,omitempty"`
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name,omitempty"`
	OnlyShowIfDegraded bool   `json:"only_show_if_degraded"`
	PageID             string `json:"page_id,omitempty"`
	Position           int    `json:"position"`
	Showcase           bool   `json:"showcase,omitempty"`
	StartDate          string `json:"start_date,omitempty"`
	Status             string `json:"status,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

func ComponentConfigToApi(configComponent pkg.Component, apiComponent *Component) error {
	err := mapstructure.Decode(configComponent, apiComponent)
	if err != nil {
		return err
	}
	apiComponent.Showcase = !configComponent.HideUptime
	return nil
}

// requestComponent represents what Statuspage accepts as input for components.
// This is necessary because Statuspage errors if unexpected keys are present
// in request JSON (???) so we must reduce Component down to this type.
type requestComponent struct {
	Description        string `json:"description"`
	GroupID            string `json:"group_id,omitempty"`
	Name               string `json:"name,omitempty"`
	OnlyShowIfDegraded bool   `json:"only_show_if_degraded"`
	Showcase           bool   `json:"showcase"`
	StartDate          string `json:"start_date,omitempty"`
	Status             string `json:"status,omitempty"`
}

func (c *Component) toRequest() requestComponent {
	return requestComponent{
		Description:        c.Description,
		Status:             c.Status,
		Name:               c.Name,
		OnlyShowIfDegraded: c.OnlyShowIfDegraded,
		GroupID:            c.GroupID,
		Showcase:           c.Showcase,
		StartDate:          c.StartDate,
	}
}

func GetComponents(client *resty.Client, pageID string) (*[]Component, error) {
	resp, err := client.R().
		SetResult([]Component{}).
		Get(fmt.Sprintf("/pages/%s/components", pageID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*[]Component), nil
}

func PostComponent(client *resty.Client, pageID string, component Component) (*Component, error) {
	resp, err := client.R().
		SetResult(Component{}).
		SetBody(map[string]interface{}{"component": component.toRequest()}).
		Post(fmt.Sprintf("/pages/%s/components", pageID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*Component), nil
}

func PatchComponent(client *resty.Client, pageID string, componentID string, component Component) (*Component, error) {
	resp, err := client.R().
		SetResult(Component{}).
		SetBody(map[string]interface{}{"component": component.toRequest()}).
		Patch(fmt.Sprintf("/pages/%s/components/%s", pageID, componentID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*Component), nil
}

func DeleteComponent(client *resty.Client, pageID string, componentID string) error {
	resp, err := client.R().
		Delete(fmt.Sprintf("/pages/%s/components/%s", pageID, componentID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return err
	}
	return nil
}
