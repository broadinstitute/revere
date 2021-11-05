package state

import (
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"sync"
)

// ComponentState records information about components that's derived during continuous operation.
// Its fields shouldn't be operated on in parallel; it contains a sync.Mutex to help state.State
// manage attempts at concurrent access.
type ComponentState struct {
	openIncidents map[string]statuspagetypes.Status
	desiredStatus statuspagetypes.Status
	id            string
	lock          *sync.Mutex
}

// recalculateDesiresStatus updates the cached desiresStatus and returns a bool representing if the value changed.
func (c *ComponentState) recalculateDesiredStatus() bool {
	worstStatusSoFar := statuspagetypes.Operational
	for _, status := range c.openIncidents {
		worstStatusSoFar = worstStatusSoFar.WorstWith(status)
	}
	if worstStatusSoFar != c.desiredStatus {
		c.desiredStatus = worstStatusSoFar
		return true
	}
	return false
}

// GetID returns the Statuspage ID correlating to this component.
func (c *ComponentState) GetID() string {
	return c.id
}

// GetDesiredStatus returns the status that the component should have. This can be cached
// so long as it is recalculated when a component's incidents change.
func (c *ComponentState) GetDesiredStatus() statuspagetypes.Status {
	return c.desiredStatus
}

// LogIncident notes a new/updated incident affecting the status of the component.
// The returned bool represents if the component's entire status changed based on the new incident.
func (c *ComponentState) LogIncident(incidentID string, componentStatus statuspagetypes.Status) bool {
	c.openIncidents[incidentID] = componentStatus
	return c.recalculateDesiredStatus()
}

// ResolveIncident notes than an incident is no longer affecting the status of the component.
// Has no effect if the incident has already been resolved or never existed/
// The return bool represents if the component's entire status changed based on the resolved incident.
func (c *ComponentState) ResolveIncident(incidentID string) bool {
	delete(c.openIncidents, incidentID)
	return c.recalculateDesiredStatus()
}
