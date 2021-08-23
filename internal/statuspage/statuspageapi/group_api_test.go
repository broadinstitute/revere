package statuspageapi

import (
	"github.com/broadinstitute/revere/internal/statuspage/statuspagemocks"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"reflect"
	"testing"
)

func TestDeleteGroup(t *testing.T) {
	type args struct {
		client  *resty.Client
		pageID  string
		groupID string
	}
	config := testConfig()
	group := statuspagemocks.GroupFactory("to delete")
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Succeeds on 204",
			args: args{
				client:  Client(config),
				pageID:  config.Statuspage.PageID,
				groupID: group.ID,
			},
		},
		{
			name: "Fails on 404",
			args: args{
				client:  Client(config),
				pageID:  config.Statuspage.PageID,
				groupID: "nonexistentID",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureGroupMock(config, map[string]string{}, map[string]statuspagetypes.Group{
				group.ID: *group,
			})
			// test for function returning an error
			if err := DeleteGroup(tt.args.client, tt.args.pageID, tt.args.groupID); (err != nil) != tt.wantErr {
				t.Errorf("DeleteGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
			httpmock.DeactivateAndReset()
		})
	}
}

func TestGetGroups(t *testing.T) {
	type args struct {
		client *resty.Client
		pageID string
	}
	config := testConfig()
	group := statuspagemocks.GroupFactory("to be returned")
	group.PageID = config.Statuspage.PageID
	tests := []struct {
		name    string
		args    args
		want    *[]statuspagetypes.Group
		wantErr bool
	}{
		{
			name: "Returns parsed group list",
			args: args{
				client: Client(config),
				pageID: config.Statuspage.PageID,
			},
			want: &[]statuspagetypes.Group{*group},
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
			statuspagemocks.ConfigureGroupMock(config, map[string]string{}, map[string]statuspagetypes.Group{
				group.ID: *group,
			})
			got, err := GetGroups(tt.args.client, tt.args.pageID)
			httpmock.DeactivateAndReset()
			// test for function returning an error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test for function mutating groups
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGroups() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatchGroup(t *testing.T) {
	type args struct {
		client  *resty.Client
		pageID  string
		groupID string
		group   statuspagetypes.Group
	}
	config := testConfig()
	baseGroup := statuspagemocks.GroupFactory("to be edited")
	baseGroup.PageID = config.Statuspage.PageID
	modifiedGroup := statuspagetypes.Group{
		Name:       baseGroup.Name,
		ID:         baseGroup.ID,
		PageID:     baseGroup.PageID,
		Components: []string{"1234"},
	}
	tests := []struct {
		name    string
		args    args
		want    *statuspagetypes.Group
		wantErr bool
	}{
		{
			name: "Modifies the group if found",
			args: args{
				client:  Client(config),
				pageID:  config.Statuspage.PageID,
				groupID: baseGroup.ID,
				group:   modifiedGroup,
			},
			want: &modifiedGroup,
		},
		{
			name: "Fails on page 404",
			args: args{
				client:  Client(config),
				pageID:  "nonexistentID",
				groupID: baseGroup.ID,
				group:   modifiedGroup,
			},
			wantErr: true,
		},
		{
			name: "Fails on group 404",
			args: args{
				client:  Client(config),
				pageID:  config.Statuspage.PageID,
				groupID: "nonexistentID",
				group:   modifiedGroup,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			statuspagemocks.ConfigureGroupMock(config, map[string]string{}, map[string]statuspagetypes.Group{
				baseGroup.ID: *baseGroup,
			})
			got, err := PatchGroup(tt.args.client, tt.args.pageID, tt.args.groupID, tt.args.group)
			httpmock.DeactivateAndReset()
			// test for function returning an error
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test for function mutating groups
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PatchGroup() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostGroup(t *testing.T) {
	type args struct {
		client *resty.Client
		pageID string
		group  statuspagetypes.Group
	}
	config := testConfig()
	newGroup := statuspagemocks.GroupFactory("to be created")
	tests := []struct {
		name    string
		args    args
		want    *statuspagetypes.Group
		wantErr bool
	}{
		{
			name: "Creates the group and returns it",
			args: args{
				client: Client(config),
				pageID: config.Statuspage.PageID,
				group:  *newGroup,
			},
			want: newGroup,
		},
		{
			name: "Errors on 404",
			args: args{
				client: Client(config),
				pageID: "nonexistentID",
				group:  *newGroup,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.client.GetClient())
			groupMap := map[string]statuspagetypes.Group{}
			statuspagemocks.ConfigureGroupMock(config, map[string]string{}, groupMap)
			got, err := PostGroup(tt.args.client, tt.args.pageID, tt.args.group)
			httpmock.DeactivateAndReset()
			// test for function returning an error
			if (err != nil) != tt.wantErr {
				t.Errorf("PostGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test for function mutating groups
			for id, group := range groupMap {
				// One group, but we lack some info on it
				tt.want.ID = id
				tt.want.PageID = group.PageID
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PostGroup() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
