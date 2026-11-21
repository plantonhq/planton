package verify

import "testing"

func TestProjectInSelfLink(t *testing.T) {
	for link, want := range map[string]string{
		"https://www.googleapis.com/compute/v1/projects/second/global/networks/n": "second",
		"projects/second/regions/us-east1/subnetworks/s":                          "second",
		"": "harness",
		"https://www.googleapis.com/compute/v1/global/networks/n": "harness",
	} {
		if got := projectInSelfLink(link, "harness"); got != want {
			t.Errorf("projectInSelfLink(%q) = %q, want %q", link, got, want)
		}
	}
}
