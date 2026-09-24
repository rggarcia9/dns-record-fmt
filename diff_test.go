package dnsfmt

import (
	"reflect"
	"testing"
)

func TestDiffLines(t *testing.T) {
	cases := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{
			"identical input produces no changes",
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			[]string{" a", " b", " c"},
		},
		{
			"both empty",
			nil,
			nil,
			[]string{},
		},
		{
			"a empty, everything added",
			nil,
			[]string{"a", "b"},
			[]string{"+a", "+b"},
		},
		{
			"b empty, everything removed",
			[]string{"a", "b"},
			nil,
			[]string{"-a", "-b"},
		},
		{
			"one line changed in the middle",
			[]string{"a", "b", "c"},
			[]string{"a", "x", "c"},
			[]string{" a", "-b", "+x", " c"},
		},
		{
			"a line was added",
			[]string{"a", "c"},
			[]string{"a", "b", "c"},
			[]string{" a", "+b", " c"},
		},
		{
			"a line was removed",
			[]string{"a", "b", "c"},
			[]string{"a", "c"},
			[]string{" a", "-b", " c"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := DiffLines(c.a, c.b)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("DiffLines(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}
