package statuspage

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
)

// GetGroups provides a slice of all groups on the remote page
func GetGroups(client *resty.Client, pageID string) (*[]statuspagetypes.Group, error) {
	resp, err := client.R().
		SetResult([]statuspagetypes.Group{}).
		Get(fmt.Sprintf("/pages/%s/component-groups", pageID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*[]statuspagetypes.Group), nil
}

// PostGroup creates a new group on the remote page
func PostGroup(client *resty.Client, pageID string, group statuspagetypes.Group) (*statuspagetypes.Group, error) {
	resp, err := client.R().
		SetResult(statuspagetypes.Group{}).
		SetBody(group.ToRequest()).
		Post(fmt.Sprintf("/pages/%s/component-groups", pageID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*statuspagetypes.Group), nil
}

// PatchGroup updates an existing group on the remote page by the group's ID, not name
func PatchGroup(client *resty.Client, pageID string, groupID string, group statuspagetypes.Group) (*statuspagetypes.Group, error) {
	resp, err := client.R().
		SetResult(statuspagetypes.Group{}).
		SetBody(group.ToRequest()).
		Patch(fmt.Sprintf("/pages/%s/component-groups/%s", pageID, groupID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*statuspagetypes.Group), nil
}

// DeleteGroup deletes an existing group on the remote page by the group's ID, not name
func DeleteGroup(client *resty.Client, pageID string, groupID string) error {
	resp, err := client.R().
		Delete(fmt.Sprintf("/pages/%s/component-groups/%s", pageID, groupID))
	if err = shared.CheckResponse(resp, err); err != nil {
		return err
	}
	return nil
}
