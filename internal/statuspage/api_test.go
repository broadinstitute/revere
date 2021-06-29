package statuspage

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"reflect"
	"testing"
)

func TestComponentConfigToApi(t *testing.T) {
	type args struct {
		configComponent configuration.Component
		apiComponent    *Component
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *Component
	}{
		{
			name: "Inverts showcase field",
			args: args{
				configComponent: configuration.Component{
					HideUptime: false,
				},
				apiComponent: &Component{},
			},
			want: &Component{
				Showcase: true,
			},
		},
		{
			name: "Translates fields",
			args: args{
				configComponent: configuration.Component{
					Name:               "Foo",
					Description:        "Bar",
					OnlyShowIfDegraded: true,
					HideUptime:         true,
					StartDate:          "Baz",
				},
				apiComponent: &Component{},
			},
			want: &Component{
				Name:               "Foo",
				Description:        "Bar",
				OnlyShowIfDegraded: true,
				Showcase:           false,
				StartDate:          "Baz",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ComponentConfigToApi(tt.args.configComponent, tt.args.apiComponent); (err != nil) != tt.wantErr {
				t.Errorf("ComponentConfigToApi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err := ComponentConfigToApi(tt.args.configComponent, tt.args.apiComponent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComponentConfigToApi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.apiComponent, tt.want) {
				t.Errorf("GetComponents() mutated = %v, want %v", tt.args.apiComponent, tt.want)
			}
		})
	}
}

func TestComponent_toRequest(t *testing.T) {
	type fields struct {
		AutomationEmail    string
		CreatedAt          string
		Description        string
		Group              bool
		GroupID            string
		ID                 string
		Name               string
		OnlyShowIfDegraded bool
		PageID             string
		Position           int
		Showcase           bool
		StartDate          string
		Status             string
		UpdatedAt          string
	}
	tests := []struct {
		name   string
		fields fields
		want   requestComponent
	}{
		{
			name: "Translates fields",
			fields: fields{
				AutomationEmail:    "a",
				CreatedAt:          "b",
				Description:        "c",
				Group:              true,
				GroupID:            "d",
				ID:                 "e",
				Name:               "f",
				OnlyShowIfDegraded: true,
				PageID:             "g",
				Position:           1,
				Showcase:           true,
				StartDate:          "h",
				Status:             "i",
				UpdatedAt:          "j",
			},
			want: requestComponent{
				Description:        "c",
				GroupID:            "d",
				Name:               "f",
				OnlyShowIfDegraded: true,
				Showcase:           true,
				StartDate:          "h",
				Status:             "i",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Component{
				AutomationEmail:    tt.fields.AutomationEmail,
				CreatedAt:          tt.fields.CreatedAt,
				Description:        tt.fields.Description,
				Group:              tt.fields.Group,
				GroupID:            tt.fields.GroupID,
				ID:                 tt.fields.ID,
				Name:               tt.fields.Name,
				OnlyShowIfDegraded: tt.fields.OnlyShowIfDegraded,
				PageID:             tt.fields.PageID,
				Position:           tt.fields.Position,
				Showcase:           tt.fields.Showcase,
				StartDate:          tt.fields.StartDate,
				Status:             tt.fields.Status,
				UpdatedAt:          tt.fields.UpdatedAt,
			}
			if got := c.toRequest(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			Components []configuration.Component
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
	component := componentFactory("to delete")
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
			ConfigureComponentMock(config, map[string]Component{component.ID: *component})
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
	component := componentFactory("to be returned")
	tests := []struct {
		name    string
		args    args
		want    *[]Component
		wantErr bool
	}{
		{
			name: "Returns parsed component list",
			args: args{
				client: Client(config),
				pageID: config.Statuspage.PageID,
			},
			want: &[]Component{*component},
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
			ConfigureComponentMock(config, map[string]Component{component.ID: *component})
			got, err := GetComponents(tt.args.client, tt.args.pageID)
			httpmock.DeactivateAndReset()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatchComponent(t *testing.T) {
	type args struct {
		client      *resty.Client
		pageID      string
		componentID string
		component   Component
	}
	config := testConfig()
	baseComponent := componentFactory("to be edited")
	modifiedComponent := componentFactory("edited component")
	modifiedComponent.ID = baseComponent.ID
	tests := []struct {
		name    string
		args    args
		want    *Component
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
			ConfigureComponentMock(config, map[string]Component{baseComponent.ID: *baseComponent})
			got, err := PatchComponent(tt.args.client, tt.args.pageID, tt.args.componentID, tt.args.component)
			httpmock.DeactivateAndReset()
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PatchComponent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostComponent(t *testing.T) {
	type args struct {
		client    *resty.Client
		pageID    string
		component Component
	}
	config := testConfig()
	newComponent := componentFactory("to be created")
	tests := []struct {
		name    string
		args    args
		want    *Component
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
			componentMap := map[string]Component{}
			ConfigureComponentMock(config, componentMap)
			got, err := PostComponent(tt.args.client, tt.args.pageID, tt.args.component)
			httpmock.DeactivateAndReset()
			if (err != nil) != tt.wantErr {
				t.Errorf("PostComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for id, component := range componentMap {
				// One component, but we lack some info on it
				tt.want.ID = id
				tt.want.PageID = component.PageID
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PostComponent() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
