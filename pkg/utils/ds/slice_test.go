package ds

import (
	"reflect"
	"testing"
)

func TestUniq(t *testing.T) {
	cases := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			"non unique",
			[]int{0, 1, 2, 3, 4},
			[]int{0, 1, 2, 3, 4},
		},
		{
			"want unique",
			[]int{0, 0, 1, 2, 2, 3, 4},
			[]int{0, 1, 2, 3, 4},
		},
		{
			"want unique 2",
			[]int{0, 0, 1, 2, 2, 3, 4, 4},
			[]int{0, 1, 2, 3, 4},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Uniq(c.input); !reflect.DeepEqual(got, c.want) {
				t.Errorf("Uniq() = %v, want %v", got, c.want)
			}
		})
	}
}
