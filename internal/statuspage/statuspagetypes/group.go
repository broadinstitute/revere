package statuspagetypes

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
)

// Group represents how Statuspage represents component groups in its API.
// This is not an exact subset of configuration.ComponentGroup, since components
// here are represented by their IDs instead of by their names.
type Group struct {
	Components  []string `json:"components"`
	CreatedAt   string   `json:"created_at"`
	Description string   `json:"description"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	PageID      string   `json:"page_id"`
	Position    int      `json:"position"`
	UpdatedAt   string   `json:"updated_at"`
}

// MergeConfigGroupToApi overwrites fields of the Group with what of the one from the config.
// Does not create a new Group so that it can be used for merging and difference finding.
func MergeConfigGroupToApi(configGroup configuration.ComponentGroup, apiGroup *Group, componentNameToID map[string]string) error {
	apiGroup.Name = configGroup.Name
	apiGroup.Description = configGroup.Description
	componentIDs := make([]string, 0, len(configGroup.ComponentNames))
	for _, name := range configGroup.ComponentNames {
		id, present := componentNameToID[name]
		if !present {
			return fmt.Errorf("component %s only declared in groups section", name)
		}
		componentIDs = append(componentIDs, id)
	}
	apiGroup.Components = componentIDs
	return nil
}

// RequestGroup represents what Statuspage accepts as input for groups.
// This is necessary because the request structure is notably different from that
// received as a response.
type RequestGroup struct {
	ComponentGroup struct {
		Components []string `json:"components"`
		Name       string   `json:"name"`
	} `json:"component_group"`
	Description string `json:"description"`
}

// ToRequest converts Group to RequestGroup (since the structures differ so much)
func (g *Group) ToRequest() RequestGroup {
	return RequestGroup{
		ComponentGroup: struct {
			Components []string `json:"components"`
			Name       string   `json:"name"`
		}{
			Components: g.Components,
			Name:       g.Name,
		},
		Description: g.Description,
	}
}
