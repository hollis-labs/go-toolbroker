package broker

import (
	"context"
	"testing"
)

func TestNopEnricher_AlwaysMisses(t *testing.T) {
	var enr Enricher = NopEnricher{}

	cases := []string{"", "any_tool", "something_else"}
	for _, name := range cases {
		h, ok, err := enr.LookupByToolName(context.Background(), name)
		if err != nil {
			t.Errorf("LookupByToolName(%q): unexpected error: %v", name, err)
		}
		if ok {
			t.Errorf("LookupByToolName(%q): expected ok=false, got ok=true", name)
		}
		if !h.IsEmpty() {
			t.Errorf("LookupByToolName(%q): expected zero-value Hints, got %#v", name, h)
		}
	}
}
