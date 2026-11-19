package aa_e2e

import "testing"

// A list output reads by index too, the form list-walking verifiers use.
func TestStringOutputsIndexLists(t *testing.T) {
	got := stringOutputs(map[string]interface{}{
		"router_name":  "r",
		"tunnel_names": []interface{}{"t0", "t1"},
	})
	if got["router_name"] != "r" || got["tunnel_names.0"] != "t0" || got["tunnel_names.1"] != "t1" || got["tunnel_names"] == "" {
		t.Fatalf("outputs = %v", got)
	}
}
