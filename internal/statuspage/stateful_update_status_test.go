package statuspage

import (
	"github.com/broadinstitute/revere/internal/cloudmonitoring"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/state"
	"github.com/broadinstitute/revere/internal/statuspage/statuspageapi"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagemocks"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/jarcoal/httpmock"
	"testing"
)

func makeConfigHelper(components []configuration.Component, serviceToComponentMapping []configuration.ServiceToComponentMapping) *configuration.Config {
	return &configuration.Config{
		Verbose: false,
		Client: struct {
			Redirects int
			Retries   int
		}{Redirects: 3, Retries: 3},
		Statuspage: struct {
			ApiKey     string `validate:"required"`
			PageID     string `validate:"required"`
			ApiRoot    string
			Components []configuration.Component      `validate:"unique=Name,dive"`
			Groups     []configuration.ComponentGroup `validate:"unique=Name,dive"`
		}{
			ApiKey: "foo", PageID: "bar", ApiRoot: "https://localhost",
			Components: components,
		},
		ServiceToComponentMapping: serviceToComponentMapping,
	}
}

func TestStatusUpdater(t *testing.T) {
	type args struct {
		config       *configuration.Config
		appStateSeed map[string]string
		mockState    map[string]statuspagetypes.Component
	}
	type resultArgs struct {
		componentName string
		labels        *cloudmonitoring.AlertLabels
		incident      *cloudmonitoring.MonitoringIncident
	}
	tests := []struct {
		name string
		// Args passed to the StatusUpdater function
		args args
		// Extra modifications made to the state (so we can pretend there are already incidents there)
		stateModifications func(appState *state.State)
		// Args passed to StatusUpdater's result when we call it for effect
		resultArgs resultArgs
		// Desired status of the component (in-memory and in-mock) after calling StatusUpdater's result
		wantStatus statuspagetypes.Status
		wantErr    bool
	}{
		{
			name: "plain update",
			args: args{
				config: makeConfigHelper(
					[]configuration.Component{
						{Name: "a component"},
					},
					[]configuration.ServiceToComponentMapping{
						{
							ServiceName: "a service", ServiceEnvironment: "an environment",
							AffectsComponentsNamed: []string{"a component"},
						},
					}),
				appStateSeed: map[string]string{
					"a component": "a-component-id",
				},
				mockState: map[string]statuspagetypes.Component{
					"a-component-id": {Name: "a component", ID: "a-component-id", Status: "operational"},
				},
			},
			resultArgs: resultArgs{
				componentName: "a component",
				labels:        &cloudmonitoring.AlertLabels{AlertType: statuspagetypes.MajorOutage},
				incident: &cloudmonitoring.MonitoringIncident{
					IncidentID: "an-incident-id",
					State:      "open",
				},
			},
			wantStatus: statuspagetypes.MajorOutage,
		},
		{
			name: "plain resolve",
			args: args{
				config: makeConfigHelper(
					[]configuration.Component{
						{Name: "a component"},
					},
					[]configuration.ServiceToComponentMapping{
						{
							ServiceName: "a service", ServiceEnvironment: "an environment",
							AffectsComponentsNamed: []string{"a component"},
						},
					}),
				appStateSeed: map[string]string{
					"a component": "a-component-id",
				},
				mockState: map[string]statuspagetypes.Component{
					"a-component-id": {Name: "a component", ID: "a-component-id", Status: "major_outage"},
				},
			},
			stateModifications: func(appState *state.State) {
				_ = appState.UseComponent("a component", func(c *state.ComponentState) error {
					c.LogIncident("an-incident-id", statuspagetypes.MajorOutage)
					return nil
				})
			},
			resultArgs: resultArgs{
				componentName: "a component",
				labels:        &cloudmonitoring.AlertLabels{AlertType: statuspagetypes.MajorOutage},
				incident: &cloudmonitoring.MonitoringIncident{
					IncidentID: "an-incident-id",
					State:      "closed",
				},
			},
			wantStatus: statuspagetypes.Operational,
		},
		{
			name: "no-op update (duplicate incident log)",
			args: args{
				config: makeConfigHelper(
					[]configuration.Component{
						{Name: "a component"},
					},
					[]configuration.ServiceToComponentMapping{
						{
							ServiceName: "a service", ServiceEnvironment: "an environment",
							AffectsComponentsNamed: []string{"a component"},
						},
					}),
				appStateSeed: map[string]string{
					"a component": "a-component-id",
				},
				mockState: map[string]statuspagetypes.Component{
					"a-component-id": {Name: "a component", ID: "a-component-id", Status: "major_outage"},
				},
			},
			stateModifications: func(appState *state.State) {
				_ = appState.UseComponent("a component", func(c *state.ComponentState) error {
					// Force the state to already contain this incident
					c.LogIncident("an-incident-id", statuspagetypes.MajorOutage)
					return nil
				})
			},
			resultArgs: resultArgs{
				componentName: "a component",
				labels:        &cloudmonitoring.AlertLabels{AlertType: statuspagetypes.MajorOutage},
				incident: &cloudmonitoring.MonitoringIncident{
					IncidentID: "an-incident-id",
					State:      "open",
				},
			},
			wantStatus: statuspagetypes.MajorOutage,
		},
		{
			name: "no-op update (duplicate incident resolve)",
			args: args{
				config: makeConfigHelper(
					[]configuration.Component{
						{Name: "a component"},
					},
					[]configuration.ServiceToComponentMapping{
						{
							ServiceName: "a service", ServiceEnvironment: "an environment",
							AffectsComponentsNamed: []string{"a component"},
						},
					}),
				appStateSeed: map[string]string{
					"a component": "a-component-id",
				},
				mockState: map[string]statuspagetypes.Component{
					"a-component-id": {Name: "a component", ID: "a-component-id", Status: "operational"},
				},
			},
			resultArgs: resultArgs{
				componentName: "a component",
				labels:        &cloudmonitoring.AlertLabels{AlertType: statuspagetypes.MajorOutage},
				incident: &cloudmonitoring.MonitoringIncident{
					IncidentID: "an-incident-id",
					State:      "closed",
				},
			},
			wantStatus: statuspagetypes.Operational,
		},
		{
			name: "upgrade with new incident",
			args: args{
				config: makeConfigHelper(
					[]configuration.Component{
						{Name: "a component"},
					},
					[]configuration.ServiceToComponentMapping{
						{
							ServiceName: "a service", ServiceEnvironment: "an environment",
							AffectsComponentsNamed: []string{"a component"},
						},
					}),
				appStateSeed: map[string]string{
					"a component": "a-component-id",
				},
				mockState: map[string]statuspagetypes.Component{
					"a-component-id": {Name: "a component", ID: "a-component-id", Status: "partial_outage"},
				},
			},
			stateModifications: func(appState *state.State) {
				_ = appState.UseComponent("a component", func(c *state.ComponentState) error {
					// Force the state to already contain a lesser incident
					c.LogIncident("another-incident-id", statuspagetypes.PartialOutage)
					return nil
				})
			},
			resultArgs: resultArgs{
				componentName: "a component",
				labels:        &cloudmonitoring.AlertLabels{AlertType: statuspagetypes.MajorOutage},
				incident: &cloudmonitoring.MonitoringIncident{
					IncidentID: "an-incident-id",
					State:      "open",
				},
			},
			wantStatus: statuspagetypes.MajorOutage,
		},
		{
			name: "no-op with new lesser incident",
			args: args{
				config: makeConfigHelper(
					[]configuration.Component{
						{Name: "a component"},
					},
					[]configuration.ServiceToComponentMapping{
						{
							ServiceName: "a service", ServiceEnvironment: "an environment",
							AffectsComponentsNamed: []string{"a component"},
						},
					}),
				appStateSeed: map[string]string{
					"a component": "a-component-id",
				},
				mockState: map[string]statuspagetypes.Component{
					"a-component-id": {Name: "a component", ID: "a-component-id", Status: "partial_outage"},
				},
			},
			stateModifications: func(appState *state.State) {
				_ = appState.UseComponent("a component", func(c *state.ComponentState) error {
					// Force the state to already contain an incident
					c.LogIncident("another-incident-id", statuspagetypes.PartialOutage)
					return nil
				})
			},
			resultArgs: resultArgs{
				componentName: "a component",
				labels:        &cloudmonitoring.AlertLabels{AlertType: statuspagetypes.DegradedPerformance},
				incident: &cloudmonitoring.MonitoringIncident{
					IncidentID: "an-incident-id",
					State:      "open",
				},
			},
			wantStatus: statuspagetypes.PartialOutage,
		},
		{
			name: "downgrade to lesser incident",
			args: args{
				config: makeConfigHelper(
					[]configuration.Component{
						{Name: "a component"},
					},
					[]configuration.ServiceToComponentMapping{
						{
							ServiceName: "a service", ServiceEnvironment: "an environment",
							AffectsComponentsNamed: []string{"a component"},
						},
					}),
				appStateSeed: map[string]string{
					"a component": "a-component-id",
				},
				mockState: map[string]statuspagetypes.Component{
					"a-component-id": {Name: "a component", ID: "a-component-id", Status: "partial_outage"},
				},
			},
			stateModifications: func(appState *state.State) {
				_ = appState.UseComponent("a component", func(c *state.ComponentState) error {
					// Force the state to already contain an incident
					c.LogIncident("another-incident-id", statuspagetypes.PartialOutage)
					c.LogIncident("an-incident-id", statuspagetypes.DegradedPerformance)
					return nil
				})
			},
			resultArgs: resultArgs{
				componentName: "a component",
				labels:        &cloudmonitoring.AlertLabels{AlertType: statuspagetypes.PartialOutage},
				incident: &cloudmonitoring.MonitoringIncident{
					IncidentID: "another-incident-id",
					State:      "closed",
				},
			},
			wantStatus: statuspagetypes.DegradedPerformance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appState := &state.State{}
			appState.Seed(tt.args.appStateSeed)
			if tt.stateModifications != nil {
				tt.stateModifications(appState)
			}
			statuspageClient := statuspageapi.Client(tt.args.config)
			httpmock.ActivateNonDefault(statuspageClient.GetClient())
			statuspagemocks.ConfigureComponentMock(tt.args.config, tt.args.mockState)
			callback := StatusUpdater(tt.args.config, appState, statuspageClient)
			if err := callback(tt.resultArgs.componentName, tt.resultArgs.labels, tt.resultArgs.incident); (err != nil) != tt.wantErr {
				t.Errorf("callback error %v", err)
				return
			}
			err := appState.UseComponent(tt.resultArgs.componentName, func(c *state.ComponentState) error {
				// Check that the status got updated in the in-memory state
				if c.GetDesiredStatus() != tt.wantStatus {
					t.Errorf("%s status in-memory was %s, wanted %s",
						tt.resultArgs.componentName,
						c.GetDesiredStatus().ToString(),
						tt.wantStatus.ToString())
				}
				// Check that the status got updated in the API mock's state
				if tt.args.mockState[c.GetID()].Status != tt.wantStatus.ToSnakeCase() {
					t.Errorf("%s status in mock was %s, wanted %s",
						tt.resultArgs.componentName,
						tt.args.mockState[c.GetID()].Status,
						tt.wantStatus.ToSnakeCase())
				}
				return nil
			})
			if err != nil {
				t.Errorf("unexpected UseComponent error %v", err)
				return
			}
		})
	}
}
