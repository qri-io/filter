package filter

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type goodCase struct {
	filter string
	source interface{}
	value  interface{}
}

type foo struct {
	Bar string `json:"bar"`
	Baz int    `json:"baz,omitempty"`
}

func TestApply(t *testing.T) {
	fieldCases := []goodCase{
		{".", map[string]interface{}{"a": "b"}, map[string]interface{}{"a": "b"}},
		{`"swoosh"`, map[string]interface{}{"a": "b"}, "swoosh"},
		{".a", map[string]interface{}{"a": "b"}, "b"},
		{".bar", foo{Bar: "b", Baz: 10}, "b"},
		{".a.bar", map[string]interface{}{"a": foo{"b", 0}}, "b"},
		{".a | length", map[string]interface{}{"a": foo{"b", 0}}, 2},

		{".[:]", []string{"a", "b", "c"}, []interface{}{"a", "b", "c"}},
		{"[1]", []string{"a", "b", "c"}, "b"},
		{".[0:2]", []string{"a", "b", "c"}, []interface{}{"a", "b"}},
		{".bar.[0:2]", map[string]interface{}{"bar": []string{"a", "b", "c"}}, []interface{}{"a", "b"}},

		{".bar.a",
			map[string]interface{}{
				"bar": []interface{}{
					map[string]string{"a": "a"},
					map[string]string{"a": "b"},
					map[string]string{"a": "c"}}}, []interface{}{"a", "b", "c"}},

		{".foo, .bar", map[string]string{"bar": "a", "foo": "b", "camp": "lucky"}, []interface{}{"b", "a"}},

		{".bar * 5", map[string]interface{}{"bar": 5}, float64(25)},
		// {"( .bar | length ) x 5", map[string]interface{}{ "bar": []string{"a","b","c"} }, 15},
	}

	for _, c := range fieldCases {
		t.Run(fmt.Sprintf("Filter_%s", c.filter), func(t *testing.T) {
			got, err := Apply(c.filter, c.source)
			if err != nil {
				t.Fatalf("error: %s", err)
			}
			if diff := cmp.Diff(c.value, got); diff != "" {
				t.Errorf("value mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
