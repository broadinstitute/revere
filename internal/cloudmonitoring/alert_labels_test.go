package cloudmonitoring

import (
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"reflect"
	"testing"
)

func TestMonitoringPacket_ParseLabels(t *testing.T) {
	type fields struct {
		Version  string
		Incident *MonitoringIncident
	}
	tests := []struct {
		name    string
		fields  fields
		want    *AlertLabels
		wantErr bool
	}{
		{
			name: "Parses properly",
			fields: fields{Incident: &MonitoringIncident{PolicyUserLabels: map[string]string{
				"revere-service-name":        "buffer",
				"revere-service-environment": "prod",
				"revere-alert-type":          "degraded-performance",
				"another-random-label":       "random-value",
			}}},
			want: &AlertLabels{
				ServiceName:        "buffer",
				ServiceEnvironment: "prod",
				AlertType:          statuspagetypes.DegradedPerformance,
			},
		},
		{
			name:    "Errors without labels",
			fields:  fields{Incident: &MonitoringIncident{}},
			wantErr: true,
		},
		{
			name: "Errors without name",
			fields: fields{Incident: &MonitoringIncident{PolicyUserLabels: map[string]string{
				"revere-service-environment": "prod",
				"revere-alert-type":          "degraded-performance",
				"another-random-label":       "random-value",
			}}},
			wantErr: true,
		},
		{
			name: "Errors without environment",
			fields: fields{Incident: &MonitoringIncident{PolicyUserLabels: map[string]string{
				"revere-service-name":  "buffer",
				"revere-alert-type":    "degraded-performance",
				"another-random-label": "random-value",
			}}},
			wantErr: true,
		},
		{
			name: "Errors without alert type",
			fields: fields{Incident: &MonitoringIncident{PolicyUserLabels: map[string]string{
				"revere-service-name":        "buffer",
				"revere-service-environment": "prod",
				"another-random-label":       "random-value",
			}}},
			wantErr: true,
		},
		{
			name: "Errors if alert type can't be parsed",
			fields: fields{Incident: &MonitoringIncident{PolicyUserLabels: map[string]string{
				"revere-service-name":        "buffer",
				"revere-service-environment": "prod",
				"revere-alert-type":          "nonsensical-value",
				"another-random-label":       "random-value",
			}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MonitoringPacket{
				Version:  tt.fields.Version,
				Incident: tt.fields.Incident,
			}
			got, err := p.ParseLabels()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}
