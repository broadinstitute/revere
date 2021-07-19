package actions

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/statuspage"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagemocks"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"reflect"
	"sort"
	"testing"
)

func TestReconcileComponents(t *testing.T) {
	type args struct {
		config *configuration.Config
		client *resty.Client
	}
	config := configuration.Config{
		Verbose: false,
		Client: struct {
			Redirects int
			Retries   int
		}{Redirects: 0, Retries: 0},
		Statuspage: struct {
			ApiKey     string `validate:"required"`
			PageID     string `validate:"required"`
			ApiRoot    string
			Components []configuration.Component      `validate:"unique=Name,dive"`
			Groups     []configuration.ComponentGroup `validate:"unique=Name,dive"`
		}{ApiKey: "key", PageID: "foo", ApiRoot: "https://localhost",
			Components: []configuration.Component{
				{Name: "Same", Description: "Same description"},
				{Name: "Modified", Description: "New description"},
				{Name: "New", Description: "A new component"},
			},
		},
	}
	client := statuspage.Client(&config)
	tests := []struct {
		name string
		args args
		// ID-component maps to use for the mock
		start   map[string]statuspagetypes.Component
		end     map[string]statuspagetypes.Component
		wantErr bool
	}{
		{
			name: "Reconciles components",
			args: args{
				config: &config,
				client: client,
			},
			start: map[string]statuspagetypes.Component{
				"1": {Name: "Same", Description: "Same description", Showcase: true, Status: "operational", ID: "1", PageID: "foo"},
				"2": {Name: "Modified", Description: "Old description", Showcase: true, Status: "operational", ID: "2", PageID: "foo"},
				"3": {Name: "Deleted", Description: "To be deleted", Showcase: true, Status: "operational", ID: "3", PageID: "foo"},
			},
			end: map[string]statuspagetypes.Component{
				"1": {Name: "Same", Description: "Same description", Showcase: true, Status: "operational", ID: "1", PageID: "foo"},
				"2": {Name: "Modified", Description: "New description", Showcase: true, Status: "operational", ID: "2", PageID: "foo"},
				"4": {Name: "New", Description: "A new component", Showcase: true, Status: "operational", ID: "4", PageID: "foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureComponentMock(&config, tt.start)
			err := ReconcileComponents(tt.args.config, tt.args.client)
			httpmock.DeactivateAndReset()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.start, tt.end) {
				t.Errorf("ReconcileComponents() mutated %v, want %v", tt.start, tt.end)
			}
		})
	}
}

func Test_listComponentsToCreate(t *testing.T) {
	type args struct {
		configComponentMap map[string]configuration.Component
		remoteComponentMap map[string]statuspagetypes.Component
	}
	tests := []struct {
		name    string
		args    args
		want    []statuspagetypes.Component
		wantErr bool
	}{
		{
			name: "Create components when not on remote",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"new": {Name: "new"},
				},
				remoteComponentMap: map[string]statuspagetypes.Component{},
			},
			want: []statuspagetypes.Component{{Name: "new", Showcase: true, Status: "operational"}},
		},
		{
			name: "Ignore components not slated for creation",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"name": {Name: "name", Description: "foo"},
					"new":  {Name: "new"},
				},
				remoteComponentMap: map[string]statuspagetypes.Component{
					"name": {Name: "name", Description: "baz"},
					"old":  {Name: "old"},
				},
			},
			want: []statuspagetypes.Component{{Name: "new", Showcase: true, Status: "operational"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listComponentsToCreate(tt.args.configComponentMap, tt.args.remoteComponentMap)
			sort.Sort(statuspagetypes.ComponentSort(got))
			sort.Sort(statuspagetypes.ComponentSort(tt.want))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listComponentsToCreate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listComponentsToDelete(t *testing.T) {
	type args struct {
		configComponentMap map[string]configuration.Component
		remoteComponentMap map[string]statuspagetypes.Component
	}
	tests := []struct {
		name string
		args args
		want []statuspagetypes.Component
	}{
		{
			name: "Deletes components when not on remote",
			args: args{
				configComponentMap: map[string]configuration.Component{},
				remoteComponentMap: map[string]statuspagetypes.Component{
					"a component": {ID: "123", Name: "a component"},
					"foobar":      {ID: "456", Name: "foobar"},
				},
			},
			want: []statuspagetypes.Component{{ID: "123", Name: "a component"}, {ID: "456", Name: "foobar"}},
		},
		{
			name: "Ignores components not slated for deletion",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"name": {Name: "name", Description: "foo"},
					"new":  {Name: "new"},
				},
				remoteComponentMap: map[string]statuspagetypes.Component{
					"name": {Name: "name", Description: "baz"},
					"old":  {Name: "old"},
				},
			},
			want: []statuspagetypes.Component{{Name: "old"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listComponentsToDelete(tt.args.configComponentMap, tt.args.remoteComponentMap)
			sort.Sort(statuspagetypes.ComponentSort(got))
			sort.Sort(statuspagetypes.ComponentSort(tt.want))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listComponentsToDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listComponentsToModify(t *testing.T) {
	type args struct {
		configComponentMap map[string]configuration.Component
		remoteComponentMap map[string]statuspagetypes.Component
	}
	tests := []struct {
		name    string
		args    args
		want    []statuspagetypes.Component
		wantErr bool
	}{
		{
			name: "Modifies components that differ on the remote",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"same name": {Name: "same name", Description: "new description", OnlyShowIfDegraded: false},
				},
				remoteComponentMap: map[string]statuspagetypes.Component{
					"same name": {Name: "same name", Description: "old description", Showcase: true, Status: "operational"},
				},
			},
			want: []statuspagetypes.Component{{Name: "same name", Description: "new description", Showcase: true, Status: "operational"}},
		},
		{
			name: "Translates the showcase field",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"same name": {Name: "same name", OnlyShowIfDegraded: false},
				},
				remoteComponentMap: map[string]statuspagetypes.Component{
					"same name": {Name: "same name", Showcase: true, Status: "operational"},
				},
			},
			want: nil,
		},
		{
			name: "Ignores components not requiring modification",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"name": {Name: "name", Description: "foo"},
					"new":  {Name: "new"},
				},
				remoteComponentMap: map[string]statuspagetypes.Component{
					"name": {Name: "name", Description: "baz", Showcase: true},
					"old":  {Name: "old"},
				},
			},
			want: []statuspagetypes.Component{{Name: "name", Description: "foo", Showcase: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listComponentsToModify(tt.args.configComponentMap, tt.args.remoteComponentMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("listComponentsToModify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Sort(statuspagetypes.ComponentSort(got))
			sort.Sort(statuspagetypes.ComponentSort(tt.want))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listComponentsToModify() got = %v, want %v", got, tt.want)
			}
		})
	}
}
