package cloudmonitoring

import (
	"google.golang.org/genproto/googleapis/monitoring/v3"
	"testing"
)

func TestMonitoringIncident_HasEnded(t *testing.T) {
	type fields struct {
		IncidentID              string
		URL                     string
		State                   string
		StartedAt               int64
		EndedAt                 int64
		Summary                 string
		ApigeeURL               string
		Resource                *MonitoringResource
		ResourceTypeDisplayName string
		ResourceID              string
		ResourceDisplayName     string
		ResourceName            string
		Metric                  *MonitoringMetric
		PolicyName              string
		PolicyUserLabels        map[string]string
		Documentation           *monitoring.AlertPolicy_Documentation
		Condition               *monitoring.AlertPolicy_Condition
		ConditionName           string
		Errors                  []map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Open",
			fields: fields{State: "open"},
			want:   false,
		},
		{
			name:   "Closed",
			fields: fields{State: "closed"},
			want:   true,
		},
		{
			name:   "Fallback open",
			fields: fields{State: "foo", StartedAt: 12345},
			want:   false,
		},
		{
			name:   "Fallback closed",
			fields: fields{State: "foo", StartedAt: 12345, EndedAt: 12349},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &MonitoringIncident{
				IncidentID:              tt.fields.IncidentID,
				URL:                     tt.fields.URL,
				State:                   tt.fields.State,
				StartedAt:               tt.fields.StartedAt,
				EndedAt:                 tt.fields.EndedAt,
				Summary:                 tt.fields.Summary,
				ApigeeURL:               tt.fields.ApigeeURL,
				Resource:                tt.fields.Resource,
				ResourceTypeDisplayName: tt.fields.ResourceTypeDisplayName,
				ResourceID:              tt.fields.ResourceID,
				ResourceDisplayName:     tt.fields.ResourceDisplayName,
				ResourceName:            tt.fields.ResourceName,
				Metric:                  tt.fields.Metric,
				PolicyName:              tt.fields.PolicyName,
				PolicyUserLabels:        tt.fields.PolicyUserLabels,
				Documentation:           tt.fields.Documentation,
				Condition:               tt.fields.Condition,
				ConditionName:           tt.fields.ConditionName,
				Errors:                  tt.fields.Errors,
			}
			if got := i.HasEnded(); got != tt.want {
				t.Errorf("HasEnded() = %v, want %v", got, tt.want)
			}
		})
	}
}
