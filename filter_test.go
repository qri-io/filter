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

		{"[]", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}

	for _, c := range fieldCases {
		t.Run(fmt.Sprintf("Filter_%s", c.filter), func(t *testing.T) {
			got, err := Apply(c.filter, c.source)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if diff := cmp.Diff(c.value, got); diff != "" {
				t.Errorf("value mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
