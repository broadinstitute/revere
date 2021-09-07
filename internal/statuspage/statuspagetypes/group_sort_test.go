package statuspagetypes

import (
	"github.com/google/go-cmp/cmp"
	"sort"
	"testing"
)

func TestGroupSort(t *testing.T) {
	tests := []struct {
		name  string
		given []Group
		want  []Group
	}{
		{
			name:  "No-op for single items",
			given: []Group{{Name: "foo"}},
			want:  []Group{{Name: "foo"}},
		},
		{
			name:  "Sorts when necessary",
			given: []Group{{ID: "123"}, {Name: "foo"}, {Position: 3}},
			want:  []Group{{Position: 3}, {Name: "foo"}, {ID: "123"}},
		},
		{
			name:  "Doesn't sort when stable",
			given: []Group{{Position: 3}, {Name: "foo"}, {ID: "123"}},
			want:  []Group{{Position: 3}, {Name: "foo"}, {ID: "123"}},
		},
		{
			name:  "Sorts ID/Name specifically",
			given: []Group{{ID: "a"}, {ID: "c"}, {Name: "a"}, {Name: "b", ID: "b"}},
			want:  []Group{{Name: "a"}, {ID: "a"}, {Name: "b", ID: "b"}, {ID: "c"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(GroupSort(tt.given))
			if diff := cmp.Diff(tt.want, tt.given); diff != "" {
				t.Errorf("Sorted mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
