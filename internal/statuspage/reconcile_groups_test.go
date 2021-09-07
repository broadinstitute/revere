package statuspage

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/statuspage/statuspageapi"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagemocks"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/jarcoal/httpmock"
	"reflect"
	"testing"
)

var emptyTestConfig = configuration.Config{
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
	}{ApiKey: "foo", PageID: "bar", ApiRoot: "https://localhost"},
}

func TestReconcileGroups(t *testing.T) {
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
		}{
			ApiKey:  "key",
			PageID:  "foo",
			ApiRoot: "https://localhost",
			Components: []configuration.Component{
				{Name: "A component"},
				{Name: "B component"},
				{Name: "C component"},
				{Name: "D component"},
			},
			Groups: []configuration.ComponentGroup{
				{Name: "Modified group", ComponentNames: []string{"B component", "A component"}},
				{Name: "Created group", ComponentNames: []string{"C component"}},
				{Name: "Same group", ComponentNames: []string{"D component"}},
			},
		},
	}
	client := statuspageapi.Client(&config)
	tests := []struct {
		name    string
		args    args
		wantErr bool
		// ID-name component map to use for the mock
		components map[string]string
		// ID-group maps to use for the mock
		start map[string]statuspagetypes.Group
		end   map[string]statuspagetypes.Group
	}{
		{
			name: "Reconciles groups",
			args: args{
				config: &config,
				client: client,
			},
			components: map[string]string{
				"123": "A component",
				"456": "B component",
				"789": "C component",
				"111": "D component",
			},
			start: map[string]statuspagetypes.Group{
				"1": {ID: "1", PageID: "foo", Name: "Modified group", Components: []string{"456"}},
				"2": {ID: "2", PageID: "foo", Name: "Deleted group", Components: []string{"111"}},
				"3": {ID: "3", PageID: "foo", Name: "Same group", Components: []string{"111"}},
			},
			end: map[string]statuspagetypes.Group{
				"1": {ID: "1", PageID: "foo", Name: "Modified group", Components: []string{"123", "456"}},
				"3": {ID: "3", PageID: "foo", Name: "Same group", Components: []string{"111"}},
				"4": {ID: "4", PageID: "foo", Name: "Created group", Components: []string{"789"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureGroupMock(&config, tt.components, tt.start)
			err := ReconcileGroups(tt.args.config, tt.args.client)
			httpmock.DeactivateAndReset()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.end, tt.start); diff != "" {
				t.Errorf("ReconcileGroups() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_listGroupsToCreate(t *testing.T) {
	type args struct {
		configGroupMap    map[string]configuration.ComponentGroup
		remoteGroupMap    map[string]statuspagetypes.Group
		componentNameToID map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    []statuspagetypes.Group
		wantErr bool
	}{
		{
			name: "Create groups not on remote",
			args: args{
				configGroupMap: map[string]configuration.ComponentGroup{
					"Group A": {Name: "Group A", ComponentNames: []string{"Component A", "Component B"}},
				},
				remoteGroupMap: map[string]statuspagetypes.Group{},
				componentNameToID: map[string]string{
					"Component A": "123",
					"Component B": "456",
				},
			},
			want: []statuspagetypes.Group{
				{Name: "Group A", Components: []string{"123", "456"}},
			},
		},
		{
			name: "Ignore groups not related to creation",
			args: args{
				configGroupMap: map[string]configuration.ComponentGroup{
					"Group A": {Name: "Group A", ComponentNames: []string{"Component A", "Component B"}},
					"Group C": {Name: "Group C", Description: "Would be modified"},
				},
				remoteGroupMap: map[string]statuspagetypes.Group{
					"Group B": {Name: "Group B", Description: "Would be deleted"},
					"Group C": {Name: "Group C"},
				},
				componentNameToID: map[string]string{
					"Component A": "123",
					"Component B": "456",
				},
			},
			want: []statuspagetypes.Group{
				{Name: "Group A", Components: []string{"123", "456"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listGroupsToCreate(tt.args.configGroupMap, tt.args.remoteGroupMap, tt.args.componentNameToID)
			if (err != nil) != tt.wantErr {
				t.Errorf("listGroupsToCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listGroupsToCreate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_listGroupsToDelete(t *testing.T) {
	type args struct {
		configGroupMap map[string]configuration.ComponentGroup
		remoteGroupMap map[string]statuspagetypes.Group
	}
	tests := []struct {
		name string
		args args
		want []statuspagetypes.Group
	}{
		{
			name: "Deletes groups not present in config",
			args: args{
				configGroupMap: map[string]configuration.ComponentGroup{},
				remoteGroupMap: map[string]statuspagetypes.Group{
					"Group A": {Name: "Group A"},
				},
			},
			want: []statuspagetypes.Group{{Name: "Group A"}},
		},
		{
			name: "Ignores groups not related to deletion",
			args: args{
				configGroupMap: map[string]configuration.ComponentGroup{
					"Group B": {Name: "Group B", Description: "To be created"},
					"Group C": {Name: "Group C", Description: "To be modified"},
				},
				remoteGroupMap: map[string]statuspagetypes.Group{
					"Group A": {Name: "Group A"},
					"Group C": {Name: "Group C"},
				},
			},
			want: []statuspagetypes.Group{{Name: "Group A"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listGroupsToDelete(tt.args.configGroupMap, tt.args.remoteGroupMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listGroupsToDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listGroupsToModify(t *testing.T) {
	type args struct {
		configGroupMap    map[string]configuration.ComponentGroup
		remoteGroupMap    map[string]statuspagetypes.Group
		componentNameToID map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    []statuspagetypes.Group
		wantErr bool
	}{
		{
			name: "Modifies groups",
			args: args{
				configGroupMap: map[string]configuration.ComponentGroup{
					"Group A": {Name: "Group A", ComponentNames: []string{"Component B", "Component A"}},
				},
				remoteGroupMap: map[string]statuspagetypes.Group{
					"Group A": {Name: "Group A", ID: "111", Components: []string{"456"}},
				},
				componentNameToID: map[string]string{
					"Component A": "123",
					"Component B": "456",
				},
			},
			want: []statuspagetypes.Group{
				{Name: "Group A", ID: "111", Components: []string{"123", "456"}},
			},
		},
		{
			name: "Ignores groups not related to modification",
			args: args{
				configGroupMap: map[string]configuration.ComponentGroup{
					"Group A": {Name: "Group A", ComponentNames: []string{"Component B", "Component A"}},
					"Group B": {Name: "Group B"},
				},
				remoteGroupMap: map[string]statuspagetypes.Group{
					"Group A": {Name: "Group A", ID: "111", Components: []string{"456"}},
					"Group C": {Name: "Group C"},
				},
				componentNameToID: map[string]string{
					"Component A": "123",
					"Component B": "456",
				},
			},
			want: []statuspagetypes.Group{
				{Name: "Group A", ID: "111", Components: []string{"123", "456"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listGroupsToModify(tt.args.configGroupMap, tt.args.remoteGroupMap, tt.args.componentNameToID)
			if (err != nil) != tt.wantErr {
				t.Errorf("listGroupsToModify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listGroupsToModify() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_makeComponentMapping(t *testing.T) {
	type args struct {
		client *resty.Client
		pageID string
	}
	config := emptyTestConfig
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
		// ID-name component map to seed the mock with
		seed map[string]string
	}{
		{
			name: "Inverts server 'map'",
			args: args{
				client: statuspageapi.Client(&config),
				pageID: config.Statuspage.PageID,
			},
			seed: map[string]string{
				"123": "A component",
				"456": "B component",
			},
			want: map[string]string{
				"A component": "123",
				"B component": "456",
			},
		},
		{
			name: "Squashes same name components stably",
			args: args{
				client: statuspageapi.Client(&config),
				pageID: config.Statuspage.PageID,
			},
			seed: map[string]string{
				"123": "component",
				"456": "component",
			},
			want: map[string]string{
				"component": "456",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureGroupMock(&config, tt.seed, map[string]statuspagetypes.Group{})
			got, err := makeComponentMapping(tt.args.client, tt.args.pageID)
			httpmock.DeactivateAndReset()
			if (err != nil) != tt.wantErr {
				t.Errorf("makeComponentMapping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("makeComponentMapping() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_makeStatuspageGroupMapping(t *testing.T) {
	type args struct {
		client *resty.Client
		pageID string
	}
	config := emptyTestConfig
	tests := []struct {
		name    string
		args    args
		want    map[string]statuspagetypes.Group
		wantErr bool
		// ID-group map to seed the mock with
		seed map[string]statuspagetypes.Group
	}{
		{
			name: "Inverts the server 'map'",
			args: args{
				client: statuspageapi.Client(&config),
				pageID: config.Statuspage.PageID,
			},
			seed: map[string]statuspagetypes.Group{
				"123": {Name: "Group A", ID: "123", PageID: config.Statuspage.PageID},
				"456": {Name: "Group B", ID: "456", PageID: config.Statuspage.PageID},
			},
			want: map[string]statuspagetypes.Group{
				"Group A": {Name: "Group A", ID: "123", PageID: config.Statuspage.PageID},
				"Group B": {Name: "Group B", ID: "456", PageID: config.Statuspage.PageID},
			},
		},
		{
			name: "Squashes same name groups stably",
			args: args{
				client: statuspageapi.Client(&config),
				pageID: config.Statuspage.PageID,
			},
			seed: map[string]statuspagetypes.Group{
				"123": {Name: "Group", ID: "123", PageID: config.Statuspage.PageID},
				"456": {Name: "Group", ID: "456", PageID: config.Statuspage.PageID},
			},
			want: map[string]statuspagetypes.Group{
				"Group": {Name: "Group", ID: "456", PageID: config.Statuspage.PageID},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureGroupMock(&config, map[string]string{}, tt.seed)
			got, err := makeStatuspageGroupMapping(tt.args.client, tt.args.pageID)
			httpmock.DeactivateAndReset()
			if (err != nil) != tt.wantErr {
				t.Errorf("makeStatuspageGroupMapping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("makeStatuspageGroupMapping() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
