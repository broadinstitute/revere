package statuspage

import (
    "encoding/json"
    "fmt"
    "github.com/broadinstitute/revere/internal/configuration"
    "github.com/jarcoal/httpmock"
    "github.com/mitchellh/mapstructure"
    "math/rand"
    "net/http"
)

//goland:noinspection SpellCheckingInspection
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

// randStringRunes makes a random alphabetic string of the given length
func randStringRunes(length int) string {
    b := make([]rune, length)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

// ComponentFactory helps make a unique component based on name alone.
// Creates an ID for the component to simplify testing, so that tests don't
// need to preface tests with information-gathering GETs to the mock.
// API functions themselves will strip out the ID before the component goes to
// the server.
func ComponentFactory(name string) *Component {
    return &Component{
        ID:     randStringRunes(8),
        Name:   name,
        Status: "operational",
    }
}

// ComponentMapFactory helps make a set of arbitrary components suitable for ConfigureComponentMock
func ComponentMapFactory(length int) map[string]Component {
    componentMap := make(map[string]Component)
    for idx := 0; idx < length; idx++ {
        component := ComponentFactory(randStringRunes(8))
        componentMap[component.ID] = *component
    }
    return componentMap
}

// validatePageID returns a 404 response if the pageID wasn't present as the first regex of the request URL,
// and nil otherwise
func validatePageID(pageID string, request *http.Request) *http.Response {
    reqPageID := httpmock.MustGetSubmatch(request, 1)
    if reqPageID != pageID {
        return httpmock.NewStringResponse(
            404,
            fmt.Sprintf("Page ID %s was not correct (should be %s)", reqPageID, pageID))
    }
    return nil
}

// validateComponentID returns a 404 response if a component ID wasn't present as the second regex of the request URL,
// and nil otherwise
func validateComponentID(components map[string]Component, request *http.Request) *http.Response {
    reqComponentID := httpmock.MustGetSubmatch(request, 2)
    for id, _ := range components {
        if id == reqComponentID {
            return nil
        }
    }
    return httpmock.NewStringResponse(
        404,
        fmt.Sprintf("Component ID %s was not found in components map", reqComponentID))
}

// ConfigureComponentMock mimics the behavior of Statuspage's component API via the given backing map.
// Any components given in the initial map or created via the mock will have their page ID properly set.
// The caller is responsible for activating/deactivating/resetting httpmock.
func ConfigureComponentMock(config *configuration.Config, components map[string]Component) {
    pageID := config.Statuspage.PageID
    apiRoot := config.Statuspage.ApiRoot
    for _, component := range components {
        component.PageID = pageID
    }
    httpmock.RegisterResponder("GET", fmt.Sprintf(`=~^%s/pages/(\w+)/components`, apiRoot),
        func(request *http.Request) (*http.Response, error) {
            if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
                return pageNotFound, nil
            }
            componentSlice := make([]Component, 0, len(components))
            for _, component := range components {
                componentSlice = append(componentSlice, component)
            }
            resp, err := httpmock.NewJsonResponse(200, componentSlice)
            if err != nil {
                return httpmock.NewStringResponse(500, err.Error()), nil
            }
            return resp, nil
        })
    httpmock.RegisterResponder("POST", fmt.Sprintf(`=~^%s/pages/(\w+)/co`, apiRoot),
        func(request *http.Request) (*http.Response, error) {
            if pageNotFound := validatePageID(pageID, request); pageNotFound != nil {
                return pageNotFound, nil
            }
            var incomingBody struct{Component Component}
            if err := json.NewDecoder(request.Body).Decode(&incomingBody); err != nil {
                return httpmock.NewStringResponse(400, err.Error()), nil
            }
            component := incomingBody.Component
            component.ID = randStringRunes(8)
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
            var incomingBody struct{Component Component}
            if err := json.NewDecoder(request.Body).Decode(&incomingBody); err != nil {
                return httpmock.NewStringResponse(400, err.Error()), nil
            }
            existingComponent := components[httpmock.MustGetSubmatch(request, 2)]
            if err := mapstructure.Decode(incomingBody.Component, &existingComponent); err != nil {
                return httpmock.NewStringResponse(500, err.Error()), nil
            }
            existingComponent.ID = httpmock.MustGetSubmatch(request, 2)
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
            return httpmock.NewStringResponse(204, "deleted"), nil
        })
}