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

// ConfigureComponentMock mimics the behavior of Statuspage's component API via the given backing map.
// Any components given in the initial map or created via the mock will have their page ID properly set.
// Created component IDs are incremented based on component map size and the number of deleted components.
// The caller is responsible for activating/deactivating/resetting httpmock.
func ConfigureComponentMock(config *configuration.Config, components map[string]statuspagetypes.Component) {
	deletedComponentCount := 0
	pageID := config.Statuspage.PageID
	apiRoot := config.Statuspage.ApiRoot
	for id, component := range components {
		component.PageID = pageID
		components[id] = component
	}
	httpmock.RegisterResponder("GET", fmt.Sprintf(`=~^%s/pages/(\w+)/components`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			componentSlice := make([]statuspagetypes.Component, 0, len(components))
			for _, component := range components {
				componentSlice = append(componentSlice, component)
			}
			resp, err := httpmock.NewJsonResponse(200, componentSlice)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
	httpmock.RegisterResponder("POST", fmt.Sprintf(`=~^%s/pages/(\w+)/components`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			var incomingBody struct{ Component statuspagetypes.Component }
			if err := json.NewDecoder(request.Body).Decode(&incomingBody); err != nil {
				return httpmock.NewStringResponse(400, err.Error()), nil
			}
			component := incomingBody.Component
			component.ID = strconv.Itoa(len(components) + deletedComponentCount + 1)
			component.PageID = pageID
			components[component.ID] = component
			resp, err := httpmock.NewJsonResponse(201, component)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
	httpmock.RegisterResponder("PATCH", fmt.Sprintf(`=~^%s/pages/(\w+)/components/(\w+)`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			if componentNotFound := validateComponentID(components, request); componentNotFound != nil {
				return componentNotFound, nil
			}
			var incomingBody struct{ Component statuspagetypes.Component }
			if err := json.NewDecoder(request.Body).Decode(&incomingBody); err != nil {
				return httpmock.NewStringResponse(400, err.Error()), nil
			}
			existingComponent := components[httpmock.MustGetSubmatch(request, 2)]
			// mimic Statuspage's more flexible json behavior as best we can
			if incomingBody.Component.Name != "" {
				existingComponent.Name = incomingBody.Component.Name
			}
			existingComponent.Description = incomingBody.Component.Description
			existingComponent.OnlyShowIfDegraded = incomingBody.Component.OnlyShowIfDegraded
			existingComponent.Showcase = incomingBody.Component.Showcase
			existingComponent.StartDate = incomingBody.Component.StartDate
			if incomingBody.Component.Status != "" {
				existingComponent.Status = incomingBody.Component.Status
			}
			components[httpmock.MustGetSubmatch(request, 2)] = existingComponent
			resp, err := httpmock.NewJsonResponse(200, &existingComponent)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
	httpmock.RegisterResponder("DELETE", fmt.Sprintf(`=~^%s/pages/(\w+)/components/(\w+)`, apiRoot),
		func(request *http.Request) (*http.Response, error) {
			if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
				return pageNotFound, nil
			}
			if componentNotFound := validateComponentID(components, request); componentNotFound != nil {
				return componentNotFound, nil
			}
			delete(components, httpmock.MustGetSubmatch(request, 2))
			deletedComponentCount += 1
			return httpmock.NewStringResponse(204, "deleted"), nil
		})
}
