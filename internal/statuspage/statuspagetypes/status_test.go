package statuspagetypes

import "testing"

func TestStatus_ToString(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want string
	}{
		{
			name: "Operational output",
			s:    Operational,
			want: "Operational",
		},
		{
			name: "DegradedPerformance output",
			s:    DegradedPerformance,
			want: "Degraded Performance",
		},
		{
			name: "PartialOutage output",
			s:    PartialOutage,
			want: "Partial Outage",
		},
		{
			name: "MajorOutage output",
			s:    MajorOutage,
			want: "Major Outage",
		},
		{
			name: "UnderMaintenance output",
			s:    UnderMaintenance,
			want: "Under Maintenance",
		},
		{
			name: "Invalid status output",
			s:    -1,
			want: "Invalid Status -1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ToString(); got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_ToSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want string
	}{
		{
			name: "Operational output",
			s:    Operational,
			want: "operational",
		},
		{
			name: "DegradedPerformance output",
			s:    DegradedPerformance,
			want: "degraded_performance",
		},
		{
			name: "PartialOutage output",
			s:    PartialOutage,
			want: "partial_outage",
		},
		{
			name: "MajorOutage output",
			s:    MajorOutage,
			want: "major_outage",
		},
		{
			name: "UnderMaintenance output",
			s:    UnderMaintenance,
			want: "under_maintenance",
		},
		{
			name: "Invalid status output",
			s:    -1,
			want: "invalid_status_-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ToSnakeCase(); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusFromKebabCase(t *testing.T) {
	type args struct {
		kebabCaseString string
	}
	tests := []struct {
		name    string
		args    args
		want    Status
		wantErr bool
	}{
		{
			name: "operational parse",
			args: args{kebabCaseString: "operational"},
			want: Operational,
		},
		{
			name: "degraded-performance parse",
			args: args{kebabCaseString: "degraded-performance"},
			want: DegradedPerformance,
		},
		{
			name: "partial-outage parse",
			args: args{kebabCaseString: "partial-outage"},
			want: PartialOutage,
		},
		{
			name: "major-outage parse",
			args: args{kebabCaseString: "major-outage"},
			want: MajorOutage,
		},
		{
			name: "under-maintenance parse",
			args: args{kebabCaseString: "under-maintenance"},
			want: UnderMaintenance,
		},
		{
			name:    "invalid parse",
			args:    args{kebabCaseString: "invalid"},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StatusFromKebabCase(tt.args.kebabCaseString)
			if (err != nil) != tt.wantErr {
				t.Errorf("StatusFromKebabCase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StatusFromKebabCase() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_WorstWith(t *testing.T) {
	type args struct {
		other Status
	}
	tests := []struct {
		name string
		s    Status
		args args
		want Status
	}{
		{
			name: "Same",
			s:    MajorOutage,
			args: args{other: MajorOutage},
			want: MajorOutage,
		},
		{
			name: "Self worst",
			s:    MajorOutage,
			args: args{other: PartialOutage},
			want: MajorOutage,
		},
		{
			name: "Other worst",
			s:    Operational,
			args: args{other: DegradedPerformance},
			want: DegradedPerformance,
		},
		{
			name: "Maintenance takes precedence",
			s:    UnderMaintenance,
			args: args{other: MajorOutage},
			want: UnderMaintenance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.WorstWith(tt.args.other); got != tt.want {
				t.Errorf("WorstWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
