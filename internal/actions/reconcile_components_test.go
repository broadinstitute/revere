package actions

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/statuspage"
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
			Components []configuration.Component
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
		start   map[string]statuspage.Component
		end     map[string]statuspage.Component
		wantErr bool
	}{
		{
			name: "Reconciles components",
			args: args{
				config: &config,
				client: client,
			},
			start: map[string]statuspage.Component{
				"1": {Name: "Same", Description: "Same description", Showcase: true, Status: "operational", ID: "1", PageID: "foo"},
				"2": {Name: "Modified", Description: "Old description", Showcase: true, Status: "operational", ID: "2", PageID: "foo"},
				"3": {Name: "Deleted", Description: "To be deleted", Showcase: true, Status: "operational", ID: "3", PageID: "foo"},
			},
			end: map[string]statuspage.Component{
				"1": {Name: "Same", Description: "Same description", Showcase: true, Status: "operational", ID: "1", PageID: "foo"},
				"2": {Name: "Modified", Description: "New description", Showcase: true, Status: "operational", ID: "2", PageID: "foo"},
				"4": {Name: "New", Description: "A new component", Showcase: true, Status: "operational", ID: "4", PageID: "foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspage.ConfigureComponentMock(&config, tt.start)
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
		remoteComponentMap map[string]statuspage.Component
	}
	tests := []struct {
		name    string
		args    args
		want    []statuspage.Component
		wantErr bool
	}{
		{
			name: "Create components when not on remote",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"new": {Name: "new"},
				},
				remoteComponentMap: map[string]statuspage.Component{},
			},
			want: []statuspage.Component{{Name: "new", Showcase: true, Status: "operational"}},
		},
		{
			name: "Ignore components not slated for creation",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"name": {Name: "name", Description: "foo"},
					"new":  {Name: "new"},
				},
				remoteComponentMap: map[string]statuspage.Component{
					"name": {Name: "name", Description: "baz"},
					"old":  {Name: "old"},
				},
			},
			want: []statuspage.Component{{Name: "new", Showcase: true, Status: "operational"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listComponentsToCreate(tt.args.configComponentMap, tt.args.remoteComponentMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("listComponentsToCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Sort(statuspage.ComponentSort(got))
			sort.Sort(statuspage.ComponentSort(tt.want))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listComponentsToCreate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listComponentsToDelete(t *testing.T) {
	type args struct {
		configComponentMap map[string]configuration.Component
		remoteComponentMap map[string]statuspage.Component
	}
	tests := []struct {
		name string
		args args
		want []statuspage.Component
	}{
		{
			name: "Deletes components when not on remote",
			args: args{
				configComponentMap: map[string]configuration.Component{},
				remoteComponentMap: map[string]statuspage.Component{
					"a component": {ID: "123", Name: "a component"},
					"foobar":      {ID: "456", Name: "foobar"},
				},
			},
			want: []statuspage.Component{{ID: "123", Name: "a component"}, {ID: "456", Name: "foobar"}},
		},
		{
			name: "Ignores components not slated for deletion",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"name": {Name: "name", Description: "foo"},
					"new":  {Name: "new"},
				},
				remoteComponentMap: map[string]statuspage.Component{
					"name": {Name: "name", Description: "baz"},
					"old":  {Name: "old"},
				},
			},
			want: []statuspage.Component{{Name: "old"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listComponentsToDelete(tt.args.configComponentMap, tt.args.remoteComponentMap)
			sort.Sort(statuspage.ComponentSort(got))
			sort.Sort(statuspage.ComponentSort(tt.want))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listComponentsToDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listComponentsToModify(t *testing.T) {
	type args struct {
		configComponentMap map[string]configuration.Component
		remoteComponentMap map[string]statuspage.Component
	}
	tests := []struct {
		name    string
		args    args
		want    []statuspage.Component
		wantErr bool
	}{
		{
			name: "Modifies components that differ on the remote",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"same name": {Name: "same name", Description: "new description", OnlyShowIfDegraded: false},
				},
				remoteComponentMap: map[string]statuspage.Component{
					"same name": {Name: "same name", Description: "old description", Showcase: true, Status: "operational"},
				},
			},
			want: []statuspage.Component{{Name: "same name", Description: "new description", Showcase: true, Status: "operational"}},
		},
		{
			name: "Translates the showcase field",
			args: args{
				configComponentMap: map[string]configuration.Component{
					"same name": {Name: "same name", OnlyShowIfDegraded: false},
				},
				remoteComponentMap: map[string]statuspage.Component{
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
				remoteComponentMap: map[string]statuspage.Component{
					"name": {Name: "name", Description: "baz", Showcase: true},
					"old":  {Name: "old"},
				},
			},
			want: []statuspage.Component{{Name: "name", Description: "foo", Showcase: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listComponentsToModify(tt.args.configComponentMap, tt.args.remoteComponentMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("listComponentsToModify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Sort(statuspage.ComponentSort(got))
			sort.Sort(statuspage.ComponentSort(tt.want))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listComponentsToModify() got = %v, want %v", got, tt.want)
			}
		})
	}
}
