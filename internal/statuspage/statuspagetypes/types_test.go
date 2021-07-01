package statuspagetypes

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"reflect"
	"testing"
)

func TestComponentConfigToApi(t *testing.T) {
	type args struct {
		configComponent configuration.Component
		apiComponent    *Component
	}
	tests := []struct {
		name string
		args args
		want *Component
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
			MergeConfigComponentToApi(tt.args.configComponent, tt.args.apiComponent)
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
		want   RequestComponent
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
			want: RequestComponent{
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
			if got := c.ToRequest(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
