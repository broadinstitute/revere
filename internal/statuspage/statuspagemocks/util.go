package statuspagemocks

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/jarcoal/httpmock"
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
func ComponentFactory(name string) *statuspagetypes.Component {
	return &statuspagetypes.Component{
		ID:     randStringRunes(8),
		Name:   name,
		Status: "operational",
	}
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
func validateComponentID(components map[string]statuspagetypes.Component, request *http.Request) *http.Response {
	reqComponentID := httpmock.MustGetSubmatch(request, 2)
	for id := range components {
		if id == reqComponentID {
			return nil
		}
	}
	return httpmock.NewStringResponse(
		404,
		fmt.Sprintf("Component ID %s was not found in components map", reqComponentID))
}
