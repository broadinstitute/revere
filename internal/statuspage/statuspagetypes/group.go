package statuspagetypes

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
	"sort"
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
// Sorts component IDs.
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
	sort.Strings(componentIDs)
	apiGroup.Components = componentIDs
	return nil
}

// RequestGroup represents what Statuspage accepts as input for groups.
// This is necessary because the request structure is notably different from that
// received as a response.
// NOTE: As of 7/6/2021, the documentation for the request payload is wrong!
// The docs say that the description field should only occur on the outer object,
// and while providing it there does not error, it appears to be a no-op. No docs
// mention that the field can exist on the interior ComponentGroup object, but
// doing so doesn't error and correctly sets the fields.
// I'm leaving the field in both places, to hopefully be forwards/backwards
// compatible with whatever Atlassian does to fix this inconsistency. If they
// make one field start to error, we'd find out about it on app startup, not runtime.
type RequestGroup struct {
	ComponentGroup struct {
		Components  []string `json:"components"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
	} `json:"component_group"`
	Description string `json:"description"`
}

// ToRequest converts Group to RequestGroup (since the structures differ so much)
func (g *Group) ToRequest() RequestGroup {
	return RequestGroup{
		ComponentGroup: struct {
			Components  []string `json:"components"`
			Name        string   `json:"name"`
			Description string   `json:"description"`
		}{
			Components:  g.Components,
			Name:        g.Name,
			Description: g.Description,
		},
		Description: g.Description,
	}
}
