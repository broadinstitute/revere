package statuspagemocks

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/jarcoal/httpmock"
	"net/http"
	"strconv"
)

// ConfigureGroupMock mimics the behavior of Statuspage's group API via the given backing map.
// Any groups given in the initial map or created via the mock will have their page ID properly set.
// Created component IDs are incremented based on component map size and the number of deleted components.
// The caller is responsible for activating/deactivating/resetting httpmock.
// NOTE: This mock mimics Statuspage's undocumented group payload behavior, see statuspagetypes.RequestGroup.
func ConfigureGroupMock(config *configuration.Config, componentIDtoName map[string]string, groupIDtoGroup map[string]statuspagetypes.Group) {
	deletedGroupCount := 0
	pageID := config.Statuspage.PageID
	apiRoot := config.Statuspage.ApiRoot
	for id, group := range groupIDtoGroup {
		group.PageID = pageID
		groupIDtoGroup[id] = group
	}

	// ReconcileGroups requires component name to ID, so it will query this component endpoint
	httpmock.RegisterResponder("GET", fmt.Sprintf(`=~^%s/pages/(\w+)/components`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			componentSlice := make([]statuspagetypes.Component, 0, len(componentIDtoName))
			for id, name := range componentIDtoName {
				componentSlice = append(componentSlice, statuspagetypes.Component{
					ID:   id,
					Name: name,
				})
			}
			resp, err := httpmock.NewJsonResponse(200, componentSlice)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})

	httpmock.RegisterResponder("GET", fmt.Sprintf(`=~^%s/pages/(\w+)/component-groups`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			groupSlice := make([]statuspagetypes.Group, 0, len(groupIDtoGroup))
			for _, group := range groupIDtoGroup {
				groupSlice = append(groupSlice, group)
			}
			resp, err := httpmock.NewJsonResponse(200, groupSlice)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
	// NOTE: this handler mimics the UNDOCUMENTED behavior of this endpoint relating to the description field.
	// See statuspagetypes.RequestGroup for more information.
	httpmock.RegisterResponder("POST", fmt.Sprintf(`=~^%s/pages/(\w+)/component-groups`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			var incomingBody struct {
				ComponentGroup statuspagetypes.Group `json:"component_group"`
			}
			if err := json.NewDecoder(request.Body).Decode(&incomingBody); err != nil {
				return httpmock.NewStringResponse(400, err.Error()), nil
			}
			group := incomingBody.ComponentGroup
			group.ID = strconv.Itoa(len(groupIDtoGroup) + deletedGroupCount + 1)
			group.PageID = pageID
			groupIDtoGroup[group.ID] = group
			resp, err := httpmock.NewJsonResponse(201, group)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
	// NOTE: this handler mimics the UNDOCUMENTED behavior of this endpoint relating to the description field.
	// See statuspagetypes.RequestGroup for more information.
	httpmock.RegisterResponder("PATCH", fmt.Sprintf(`=~^%s/pages/(\w+)/component-groups/(\w+)`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			if groupNotFound := validateGroupID(groupIDtoGroup, request); groupNotFound != nil {
				return groupNotFound, nil
			}
			var incomingBody struct {
				ComponentGroup statuspagetypes.Group `json:"component_group"`
			}
			if err := json.NewDecoder(request.Body).Decode(&incomingBody); err != nil {
				return httpmock.NewStringResponse(400, err.Error()), nil
			}
			existingGroup := groupIDtoGroup[httpmock.MustGetSubmatch(request, 2)]
			existingGroup.Name = incomingBody.ComponentGroup.Name
			existingGroup.Description = incomingBody.ComponentGroup.Description
			existingGroup.Components = incomingBody.ComponentGroup.Components
			groupIDtoGroup[httpmock.MustGetSubmatch(request, 2)] = existingGroup
			resp, err := httpmock.NewJsonResponse(200, &existingGroup)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, err
		})
	httpmock.RegisterResponder("DELETE", fmt.Sprintf(`=~^%s/pages/(\w+)/component-groups/(\w+)`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			if groupNotFound := validateGroupID(groupIDtoGroup, request); groupNotFound != nil {
				return groupNotFound, nil
			}
			delete(groupIDtoGroup, httpmock.MustGetSubmatch(request, 2))
			deletedGroupCount += 1
			return httpmock.NewStringResponse(204, "deleted"), nil
		})
}
