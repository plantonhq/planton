package module

import "testing"

// An unset peer address is the /30's other usable address, whichever end
// the interface takes; a declared one is kept, and no range means none.
func TestSessionPeerIPAddress(t *testing.T) {
	for _, c := range []struct{ declared, interfaceRange, want string }{
		{"", "169.254.101.2/30", "169.254.101.1"},
		{"", "169.254.100.1/30", "169.254.100.2"},
		{"169.254.100.9", "169.254.100.1/30", "169.254.100.9"},
		{"", "", ""},
	} {
		if got := sessionPeerIPAddress(c.declared, c.interfaceRange); got != c.want {
			t.Errorf("sessionPeerIPAddress(%q, %q) = %q, want %q", c.declared, c.interfaceRange, got, c.want)
		}
	}
}
