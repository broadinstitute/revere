package state

import (
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/google/go-cmp/cmp"
	"sync"
	"testing"
)

func TestComponentState_GetDesiredStatus(t *testing.T) {
	type fields struct {
		openIncidents map[string]statuspagetypes.Status
		desiredStatus statuspagetypes.Status
		id            string
		lock          *sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   statuspagetypes.Status
	}{
		{
			name: "Basic",
			fields: fields{
				desiredStatus: statuspagetypes.Operational,
			},
			want: statuspagetypes.Operational,
		},
		{
			name: "Does not recalculate on its own!",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"foo": statuspagetypes.PartialOutage,
				},
				desiredStatus: statuspagetypes.Operational,
			},
			want: statuspagetypes.Operational,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComponentState{
				openIncidents: tt.fields.openIncidents,
				desiredStatus: tt.fields.desiredStatus,
				id:            tt.fields.id,
				lock:          tt.fields.lock,
			}
			if got := c.GetDesiredStatus(); got != tt.want {
				t.Errorf("GetDesiredStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponentState_GetID(t *testing.T) {
	type fields struct {
		openIncidents map[string]statuspagetypes.Status
		desiredStatus statuspagetypes.Status
		id            string
		lock          *sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Basic",
			fields: fields{
				id: "foo",
			},
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComponentState{
				openIncidents: tt.fields.openIncidents,
				desiredStatus: tt.fields.desiredStatus,
				id:            tt.fields.id,
				lock:          tt.fields.lock,
			}
			if got := c.GetID(); got != tt.want {
				t.Errorf("GetID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponentState_LogIncident(t *testing.T) {
	type fields struct {
		openIncidents map[string]statuspagetypes.Status
		desiredStatus statuspagetypes.Status
		id            string
		lock          *sync.Mutex
	}
	type args struct {
		incidentID      string
		componentStatus statuspagetypes.Status
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		want              bool
		wantIncidents     map[string]statuspagetypes.Status
		wantDesiredStatus statuspagetypes.Status
	}{
		{
			name: "New incident",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{},
				desiredStatus: statuspagetypes.Operational,
			},
			args: args{
				incidentID:      "abc",
				componentStatus: statuspagetypes.MajorOutage,
			},
			want: true,
			wantIncidents: map[string]statuspagetypes.Status{
				"abc": statuspagetypes.MajorOutage,
			},
			wantDesiredStatus: statuspagetypes.MajorOutage,
		},
		{
			name: "Upgrade existing incident",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.PartialOutage,
				},
				desiredStatus: statuspagetypes.PartialOutage,
			},
			args: args{
				incidentID:      "abc",
				componentStatus: statuspagetypes.MajorOutage,
			},
			want: true,
			wantIncidents: map[string]statuspagetypes.Status{
				"abc": statuspagetypes.MajorOutage,
			},
			wantDesiredStatus: statuspagetypes.MajorOutage,
		},
		{
			name: "Downgrade existing incident",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.MajorOutage,
				},
				desiredStatus: statuspagetypes.MajorOutage,
			},
			args: args{
				incidentID:      "abc",
				componentStatus: statuspagetypes.PartialOutage,
			},
			want: true,
			wantIncidents: map[string]statuspagetypes.Status{
				"abc": statuspagetypes.PartialOutage,
			},
			wantDesiredStatus: statuspagetypes.PartialOutage,
		},
		{
			name: "No-op existing incident",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.MajorOutage,
				},
				desiredStatus: statuspagetypes.MajorOutage,
			},
			args: args{
				incidentID:      "abc",
				componentStatus: statuspagetypes.MajorOutage,
			},
			want: false,
			wantIncidents: map[string]statuspagetypes.Status{
				"abc": statuspagetypes.MajorOutage,
			},
			wantDesiredStatus: statuspagetypes.MajorOutage,
		},
		{
			name: "Add additional incident",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.PartialOutage,
				},
				desiredStatus: statuspagetypes.PartialOutage,
			},
			args: args{
				incidentID:      "def",
				componentStatus: statuspagetypes.MajorOutage,
			},
			want: true,
			wantIncidents: map[string]statuspagetypes.Status{
				"abc": statuspagetypes.PartialOutage,
				"def": statuspagetypes.MajorOutage,
			},
			wantDesiredStatus: statuspagetypes.MajorOutage,
		},
		{
			name: "Add additional incident without effect",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.PartialOutage,
				},
				desiredStatus: statuspagetypes.PartialOutage,
			},
			args: args{
				incidentID:      "def",
				componentStatus: statuspagetypes.DegradedPerformance,
			},
			want: false,
			wantIncidents: map[string]statuspagetypes.Status{
				"abc": statuspagetypes.PartialOutage,
				"def": statuspagetypes.DegradedPerformance,
			},
			wantDesiredStatus: statuspagetypes.PartialOutage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComponentState{
				openIncidents: tt.fields.openIncidents,
				desiredStatus: tt.fields.desiredStatus,
				id:            tt.fields.id,
				lock:          tt.fields.lock,
			}
			if got := c.LogIncident(tt.args.incidentID, tt.args.componentStatus); got != tt.want {
				t.Errorf("LogIncident() = %v, want %v", got, tt.want)
			}
			if diff := cmp.Diff(tt.wantIncidents, c.openIncidents); diff != "" {
				t.Errorf("LogIncident() bad effect: %s", diff)
			}
			if c.desiredStatus != tt.wantDesiredStatus {
				t.Errorf("LogIncident() bad effect: got desiredStatus %v, want %v", c.desiredStatus, tt.wantDesiredStatus)
			}
		})
	}
}

func TestComponentState_ResolveIncident(t *testing.T) {
	type fields struct {
		openIncidents map[string]statuspagetypes.Status
		desiredStatus statuspagetypes.Status
		id            string
		lock          *sync.Mutex
	}
	type args struct {
		incidentID string
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		want              bool
		wantIncidents     map[string]statuspagetypes.Status
		wantDesiredStatus statuspagetypes.Status
	}{
		{
			name: "Non-existent",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{},
				desiredStatus: statuspagetypes.Operational,
			},
			args:              args{incidentID: "foo"},
			want:              false,
			wantIncidents:     map[string]statuspagetypes.Status{},
			wantDesiredStatus: statuspagetypes.Operational,
		},
		{
			name: "Basic removal",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.DegradedPerformance,
				},
				desiredStatus: statuspagetypes.DegradedPerformance,
			},
			args:              args{incidentID: "abc"},
			want:              true,
			wantIncidents:     map[string]statuspagetypes.Status{},
			wantDesiredStatus: statuspagetypes.Operational,
		},
		{
			name: "Downgrade removal",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.DegradedPerformance,
					"def": statuspagetypes.MajorOutage,
				},
				desiredStatus: statuspagetypes.MajorOutage,
			},
			args: args{incidentID: "def"},
			want: true,
			wantIncidents: map[string]statuspagetypes.Status{
				"abc": statuspagetypes.DegradedPerformance,
			},
			wantDesiredStatus: statuspagetypes.DegradedPerformance,
		},
		{
			name: "No effect removal",
			fields: fields{
				openIncidents: map[string]statuspagetypes.Status{
					"abc": statuspagetypes.DegradedPerformance,
					"def": statuspagetypes.MajorOutage,
				},
				desiredStatus: statuspagetypes.MajorOutage,
			},
			args: args{incidentID: "abc"},
			want: false,
			wantIncidents: map[string]statuspagetypes.Status{
				"def": statuspagetypes.MajorOutage,
			},
			wantDesiredStatus: statuspagetypes.MajorOutage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComponentState{
				openIncidents: tt.fields.openIncidents,
				desiredStatus: tt.fields.desiredStatus,
				id:            tt.fields.id,
				lock:          tt.fields.lock,
			}
			if got := c.ResolveIncident(tt.args.incidentID); got != tt.want {
				t.Errorf("ResolveIncident() = %v, want %v", got, tt.want)
			}
			if diff := cmp.Diff(tt.wantIncidents, c.openIncidents); diff != "" {
				t.Errorf("ResolveIncident() bad effect: %s", diff)
			}
			if c.desiredStatus != tt.wantDesiredStatus {
				t.Errorf("ResolveIncident() bad effect: got desiredStatus %v, want %v", c.desiredStatus, tt.wantDesiredStatus)
			}
		})
	}
}
