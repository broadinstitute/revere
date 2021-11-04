package state

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"sync"
)

// State contains information necessary for continuous operation that's derived throughout
// the course of operation. Information not meeting that constraint should exist elsewhere
// (like configuration.Config).
// Right now, the only information meeting this criteria is per-component state.
//
// This object is responsible for making sure that concurrent users don't step on each
// other.
type State struct {
	componentNameToState *sync.Map
}

// Seed the State with the component ID information obtained from Statuspage.
func (s *State) Seed(componentNamesToIDs map[string]string) {
	if s.componentNameToState == nil {
		s.componentNameToState = &sync.Map{}
	}
	for name, id := range componentNamesToIDs {
		uncastedComponentState, _ := s.componentNameToState.LoadOrStore(name, &ComponentState{
			lock:          &sync.Mutex{},
			openIncidents: map[string]statuspagetypes.Status{},
		})
		componentState := uncastedComponentState.(*ComponentState)
		componentState.lock.Lock()
		componentState.id = id
		componentState.lock.Unlock()
	}
}

// UseComponent runs a hook function with the state of some component. This function should
// ensure that hooks never run simultaneously against the same component so long as callers
// never copy the reference to the ComponentState object.
func (s *State) UseComponent(componentName string, hook func(c *ComponentState) error) error {
	uncastedComponentState, found := s.componentNameToState.Load(componentName)
	if !found {
		return fmt.Errorf("did not find component named %s", componentName)
	}
	componentState := uncastedComponentState.(*ComponentState)
	componentState.lock.Lock()
	err := hook(componentState)
	componentState.lock.Unlock()
	return err
}
