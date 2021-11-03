package statuspageapi

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagemocks"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/jarcoal/httpmock"
	"testing"
)

func testConfig() *configuration.Config {
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
		}{ApiKey: "foo", PageID: "baz", ApiRoot: "https://localhost"},
	}
}

func TestDeleteComponent(t *testing.T) {
	type args struct {
		client      *resty.Client
		pageID      string
		componentID string
	}
	config := testConfig()
	component := statuspagemocks.ComponentFactory("to delete")
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Succeeds on 204",
			args: args{
				client:      Client(config),
				pageID:      config.Statuspage.PageID,
				componentID: component.ID,
			},
		},
		{
			name: "Fails on 404",
			args: args{
				client:      Client(config),
				pageID:      config.Statuspage.PageID,
				componentID: "nonexistentID",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureComponentMock(config, map[string]statuspagetypes.Component{component.ID: *component})
			// test for function returning an error
			if err := DeleteComponent(tt.args.client, tt.args.pageID, tt.args.componentID); (err != nil) != tt.wantErr {
				t.Errorf("DeleteComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
			httpmock.DeactivateAndReset()
		})
	}
}

func TestGetComponents(t *testing.T) {
	type args struct {
		client *resty.Client
		pageID string
	}
	config := testConfig()
	component := statuspagemocks.ComponentFactory("to be returned")
	group := statuspagemocks.ComponentFactory("a group component that shouldn't be returned")
	component.PageID = config.Statuspage.PageID
	group.PageID = config.Statuspage.PageID
	group.Group = true
	tests := []struct {
		name    string
		args    args
		want    *[]statuspagetypes.Component
		wantErr bool
	}{
		{
			name: "Returns parsed component list",
			args: args{
				client: Client(config),
				pageID: config.Statuspage.PageID,
			},
			want: &[]statuspagetypes.Component{*component},
		},
		{
			name: "Fails on 404",
			args: args{
				client: Client(config),
				pageID: "nonexistentID",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureComponentMock(config, map[string]statuspagetypes.Component{
				component.ID: *component,
				group.ID:     *group,
			})
			got, err := GetComponents(tt.args.client, tt.args.pageID)
			httpmock.DeactivateAndReset()
			// test for function returning an error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test for function mutating components
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetComponents() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPatchComponent(t *testing.T) {
	type args struct {
		client      *resty.Client
		pageID      string
		componentID string
		component   statuspagetypes.Component
	}
	config := testConfig()
	baseComponent := statuspagemocks.ComponentFactory("to be edited")
	baseComponent.PageID = config.Statuspage.PageID
	modifiedComponent := statuspagemocks.ComponentFactory("edited component")
	modifiedComponent.ID = baseComponent.ID
	modifiedComponent.PageID = config.Statuspage.PageID
	tests := []struct {
		name    string
		args    args
		want    *statuspagetypes.Component
		wantErr bool
	}{
		{
			name: "Modifies the component if found",
			args: args{
				client:      Client(config),
				pageID:      config.Statuspage.PageID,
				componentID: baseComponent.ID,
				component:   *modifiedComponent,
			},
			want: modifiedComponent,
		},
		{
			name: "Fails on page 404",
			args: args{
				client:      Client(config),
				pageID:      "nonexistentID",
				componentID: baseComponent.ID,
				component:   *modifiedComponent,
			},
			wantErr: true,
		},
		{
			name: "Fails on component 404",
			args: args{
				client:      Client(config),
				pageID:      config.Statuspage.PageID,
				componentID: "nonexistentID",
				component:   *modifiedComponent,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureComponentMock(config, map[string]statuspagetypes.Component{baseComponent.ID: *baseComponent})
			got, err := PatchComponent(tt.args.client, tt.args.pageID, tt.args.componentID, tt.args.component)
			httpmock.DeactivateAndReset()
			// test for function returning an error
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test for function mutating components
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("PatchComponent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPostComponent(t *testing.T) {
	type args struct {
		client    *resty.Client
		pageID    string
		component statuspagetypes.Component
	}
	config := testConfig()
	newComponent := statuspagemocks.ComponentFactory("to be created")
	tests := []struct {
		name    string
		args    args
		want    *statuspagetypes.Component
		wantErr bool
	}{
		{
			name: "Creates the component and returns it",
			args: args{
				client:    Client(config),
				pageID:    config.Statuspage.PageID,
				component: *newComponent,
			},
			want: newComponent,
		},
		{
			name: "Errors on 404",
			args: args{
				client:    Client(config),
				pageID:    "nonexistentID",
				component: *newComponent,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			componentMap := map[string]statuspagetypes.Component{}
			statuspagemocks.ConfigureComponentMock(config, componentMap)
			got, err := PostComponent(tt.args.client, tt.args.pageID, tt.args.component)
			httpmock.DeactivateAndReset()
			// test for function returning an error
			if (err != nil) != tt.wantErr {
				t.Errorf("PostComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test for function mutating components
			for id, component := range componentMap {
				// One component, but we lack some info on it
				tt.want.ID = id
				tt.want.PageID = component.PageID
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("PostComponent() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestPatchComponentStatus(t *testing.T) {
	type args struct {
		client      *resty.Client
		pageID      string
		componentID string
		newStatus   statuspagetypes.Status
	}
	config := testConfig()
	baseComponent := statuspagemocks.ComponentFactory("to be edited")
	baseComponent.PageID = config.Statuspage.PageID
	baseComponent.Status = statuspagetypes.Operational.ToSnakeCase()
	modifiedComponent := statuspagemocks.ComponentFactory("to be edited")
	modifiedComponent.ID = baseComponent.ID
	modifiedComponent.Status = statuspagetypes.MajorOutage.ToSnakeCase()
	modifiedComponent.PageID = config.Statuspage.PageID
	tests := []struct {
		name    string
		args    args
		want    *statuspagetypes.Component
		wantErr bool
	}{
		{
			name: "Modifies the component if found",
			args: args{
				client:      Client(config),
				pageID:      config.Statuspage.PageID,
				componentID: baseComponent.ID,
				newStatus:   statuspagetypes.MajorOutage,
			},
			want: modifiedComponent,
		},
		{
			name: "Fails on page 404",
			args: args{
				client:      Client(config),
				pageID:      "nonexistentID",
				componentID: baseComponent.ID,
				newStatus:   statuspagetypes.MajorOutage,
			},
			wantErr: true,
		},
		{
			name: "Fails on component 404",
			args: args{
				client:      Client(config),
				pageID:      config.Statuspage.PageID,
				componentID: "nonexistentID",
				newStatus:   statuspagetypes.MajorOutage,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureComponentMock(config, map[string]statuspagetypes.Component{baseComponent.ID: *baseComponent})
			got, err := PatchComponentStatus(tt.args.client, tt.args.pageID, tt.args.componentID, tt.args.newStatus)
			httpmock.DeactivateAndReset()
			// test for function returning an error
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test for function mutating components
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("PatchComponent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
