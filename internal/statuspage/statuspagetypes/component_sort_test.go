package statuspagetypes

import (
	"reflect"
	"sort"
	"testing"
)

func TestStableSort(t *testing.T) {
	tests := []struct {
		name  string
		given []Component
		want  []Component
	}{
		{
			name:  "No-op for single items",
			given: []Component{{Name: "foo"}},
			want:  []Component{{Name: "foo"}},
		},
		{
			name:  "Sorts when necessary",
			given: []Component{{ID: "123"}, {Name: "foo"}, {Position: 3}},
			want:  []Component{{Position: 3}, {Name: "foo"}, {ID: "123"}},
		},
		{
			name:  "Doesn't sort when stable",
			given: []Component{{Position: 3}, {Name: "foo"}, {ID: "123"}},
			want:  []Component{{Position: 3}, {Name: "foo"}, {ID: "123"}},
		},
		{
			name:  "Sorts ID/Name specifically",
			given: []Component{{ID: "a"}, {ID: "c"}, {Name: "a"}, {Name: "b", ID: "b"}},
			want:  []Component{{Name: "a"}, {ID: "a"}, {Name: "b", ID: "b"}, {ID: "c"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(ComponentSort(tt.given))
			if !reflect.DeepEqual(tt.given, tt.want) {
				t.Errorf("Sorted = %v, want %v", tt.given, tt.want)
			}
		})
	}
}
