package module

import "testing"

func TestParseMembershipName(t *testing.T) {
	for name, want := range map[string][2]string{
		"projects/p/locations/us-east1/memberships/m":                         {"us-east1", "m"},
		"//gkehub.googleapis.com/projects/p/locations/us-east1/memberships/m": {"us-east1", "m"},
	} {
		location, id, err := parseMembershipName(name)
		if err != nil || location != want[0] || id != want[1] {
			t.Errorf("parseMembershipName(%q) = %q, %q, %v; want %q, %q", name, location, id, err, want[0], want[1])
		}
	}
	if _, _, err := parseMembershipName("memberships/m"); err == nil {
		t.Error("parseMembershipName accepted a name with no project or location")
	}
}
