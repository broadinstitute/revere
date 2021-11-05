package statuspage

import (
	"github.com/broadinstitute/revere/internal/cloudmonitoring"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/pubsub/pubsubtypes"
	"github.com/broadinstitute/revere/internal/state"
	"github.com/broadinstitute/revere/internal/statuspage/statuspageapi"
	"github.com/go-resty/resty/v2"
)

// StatusUpdater returns a function to handle a possible update against a single component.
// The returned function is correctly typed to be called by pubsub.ReceiveMessages as a callback.
func StatusUpdater(config *configuration.Config, appState *state.State, client *resty.Client) pubsubtypes.PerComponentHandler {

	// StatusUpdater returns a function with arguments only for what changes per-component. Even though the function
	// takes advantage of config/appState/client, it has a narrow signature in line with what the pubsub package
	// parses from an incoming message.
	return func(componentName string, labels *cloudmonitoring.AlertLabels, incident *cloudmonitoring.MonitoringIncident) error {

		// Within the StatusUpdater's returned function, we wrap all work inside the appState.UseComponent hook.
		// 		This is a bit like a React useEffect hook! If that makes no sense, read on:
		//
		// 1. By wrapping the work here in a UseComponent hook, the appState gives us access to the ComponentState, so we
		// can record the incident and get the component's ID and status.
		// 2. The appState manages this access so that no two hooks will ever run simultaneously against the same
		// component.
		// 3. Because the entire body of this function is within the hook, this function does not need to worry about
		// concurrency control.
		// 4. **This eliminates a class of race conditions arising out of delay around status changes (both in-memory
		// __and__ in communicating with Statuspage.io)**
		return appState.UseComponent(componentName, func(c *state.ComponentState) error {
			var componentStatusChanged bool
			if incident.HasEnded() {
				componentStatusChanged = c.ResolveIncident(incident.IncidentID)
			} else {
				componentStatusChanged = c.LogIncident(incident.IncidentID, labels.AlertType)
			}
			if componentStatusChanged {
				_, err := statuspageapi.PatchComponentStatus(client, config.Statuspage.PageID, c.GetID(), c.GetDesiredStatus())
				return err
			}
			return nil
		})
	}
}
