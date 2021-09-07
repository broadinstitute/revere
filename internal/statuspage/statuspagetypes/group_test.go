package statuspagetypes

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
)

func TestGroup_ToRequest(t *testing.T) {
	type fields struct {
		Components  []string
		CreatedAt   string
		Description string
		ID          string
		Name        string
		PageID      string
		Position    int
		UpdatedAt   string
	}
	tests := []struct {
		name   string
		fields fields
		want   RequestGroup
	}{
		{
			name: "Converts simple cases",
			fields: fields{
				Components:  []string{"foo"},
				Description: "bar",
				Name:        "baz",
			},
			want: RequestGroup{
				ComponentGroup: struct {
					Components  []string `json:"components"`
					Name        string   `json:"name"`
					Description string   `json:"description"`
				}{Components: []string{"foo"}, Name: "baz", Description: "bar"},
				Description: "bar",
			},
		},
		{
			name: "Strips other fields",
			fields: fields{
				Components:  []string{"1", "2"},
				CreatedAt:   "3",
				Description: "4",
				ID:          "5",
				Name:        "6",
				PageID:      "7",
				Position:    8,
				UpdatedAt:   "9",
			},
			want: RequestGroup{
				ComponentGroup: struct {
					Components  []string `json:"components"`
					Name        string   `json:"name"`
					Description string   `json:"description"`
				}{Components: []string{"1", "2"}, Name: "6", Description: "4"},
				Description: "4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Group{
				Components:  tt.fields.Components,
				CreatedAt:   tt.fields.CreatedAt,
				Description: tt.fields.Description,
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				PageID:      tt.fields.PageID,
				Position:    tt.fields.Position,
				UpdatedAt:   tt.fields.UpdatedAt,
			}
			if got := g.ToRequest(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeConfigGroupToApi(t *testing.T) {
	type args struct {
		configGroup       configuration.ComponentGroup
		apiGroup          *Group
		componentNameToID map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    *Group
		wantErr bool
	}{
		{
			name: "Translates name to ID and sorts",
			args: args{
				configGroup: configuration.ComponentGroup{
					Name:           "Group",
					Description:    "A group",
					ComponentNames: []string{"Another component", "Some component"},
				},
				apiGroup: &Group{
					ID: "123",
				},
				componentNameToID: map[string]string{
					"Another component": "789",
					"Some component":    "456",
				},
			},
			want: &Group{
				Components:  []string{"456", "789"},
				Description: "A group",
				ID:          "123",
				Name:        "Group",
			},
		},
		{
			name: "Errors if component name not found",
			args: args{
				configGroup: configuration.ComponentGroup{
					ComponentNames: []string{"Nonexistent component"},
				},
				apiGroup:          &Group{},
				componentNameToID: map[string]string{},
			},
			want:    &Group{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MergeConfigGroupToApi(tt.args.configGroup, tt.args.apiGroup, tt.args.componentNameToID); (err != nil) != tt.wantErr {
				t.Errorf("MergeConfigGroupToApi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, tt.args.apiGroup); diff != "" {
				t.Errorf("MergeConfigGroupToApi() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
