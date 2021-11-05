package statuspageapi

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
)

// GetComponents provides a slice of all components (NOT groups) on the remote page
func GetComponents(client *resty.Client, pageID string) (*[]statuspagetypes.Component, error) {
	resp, err := client.R().
		SetResult([]statuspagetypes.Component{}).
		Get(fmt.Sprintf("/pages/%s/components", pageID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	componentsWithGroups := resp.Result().(*[]statuspagetypes.Component)
	componentsWithoutGroups := make([]statuspagetypes.Component, 0, len(*componentsWithGroups))
	for _, component := range *componentsWithGroups {
		if !component.Group {
			componentsWithoutGroups = append(componentsWithoutGroups, component)
		}
	}
	return &componentsWithoutGroups, nil
}

// PostComponent creates a new component on the remote page
func PostComponent(client *resty.Client, pageID string, component statuspagetypes.Component) (*statuspagetypes.Component, error) {
	resp, err := client.R().
		SetResult(statuspagetypes.Component{}).
		SetBody(map[string]interface{}{"component": component.ToRequest()}).
		Post(fmt.Sprintf("/pages/%s/components", pageID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*statuspagetypes.Component), nil
}

// DeleteComponent deletes an existing component on the remote page by the component's ID, not name
func DeleteComponent(client *resty.Client, pageID string, componentID string) error {
	resp, err := client.R().
		Delete(fmt.Sprintf("/pages/%s/components/%s", pageID, componentID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return err
	}
	return nil
}

// PatchComponent updates an existing component on the remote page by the component's ID, not name
func PatchComponent(client *resty.Client, pageID string, componentID string, component statuspagetypes.Component) (*statuspagetypes.Component, error) {
	return patchHelper(client, pageID, componentID, map[string]interface{}{"component": component.ToRequest()})
}

// PatchComponentStatus updates an existing component's status only
func PatchComponentStatus(client *resty.Client, pageID string, componentID string, newStatus statuspagetypes.Status) (*statuspagetypes.Component, error) {
	return patchHelper(client, pageID, componentID, map[string]interface{}{"component": map[string]string{"status": newStatus.ToSnakeCase()}})
}

func patchHelper(client *resty.Client, pageID string, componentID string, body map[string]interface{}) (*statuspagetypes.Component, error) {
	resp, err := client.R().
		SetResult(statuspagetypes.Component{}).
		SetBody(body).
		Patch(fmt.Sprintf("/pages/%s/components/%s", pageID, componentID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*statuspagetypes.Component), nil
}
