package state

import (
	"testing"
)

func dummyState() *State {
	s := &State{}
	s.Seed(map[string]string{
		"foo": "foo-id",
		"bar": "bar-id",
		"baz": "baz-id",
	})
	return s
}

func TestInMemoryComponentState_GetID(t *testing.T) {
	tests := []struct {
		name     string
		starting *State
		seed     map[string]string
		wantName string
		wantID   string
		wantErr  bool
	}{
		{
			name: "Basic",
			seed: map[string]string{
				"foo": "foo-id",
				"bar": "bar-id",
			},
			wantName: "foo",
			wantID:   "foo-id",
		},
		{
			name:     "Updating",
			starting: dummyState(),
			seed: map[string]string{
				"foo": "foo-id-2",
			},
			wantName: "foo",
			wantID:   "foo-id-2",
		},
		{
			name:     "Errors with blank",
			seed:     map[string]string{},
			wantName: "foo",
			wantErr:  true,
		},
		{
			name: "Errors with missing",
			seed: map[string]string{
				"bar": "bar-id",
			},
			wantName: "foo",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *State
			if tt.starting != nil {
				s = tt.starting
			} else {
				s = &State{}
			}
			s.Seed(tt.seed)
			var got string
			err := s.UseComponent(tt.wantName, func(c *ComponentState) error {
				got = c.GetID()
				return nil
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetID() error %v", err)
				return
			}
			if got != tt.wantID {
				t.Errorf("GetID() = %v, want %v", got, tt.wantID)
			}
		})
	}
}
