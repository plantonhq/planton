package aa_e2e

import "testing"

// A zone-scoped token passes the connectivity check only when the account's
// zone list it reads names a zone.
func TestZoneListHasResult(t *testing.T) {
	cases := map[string]bool{
		`{"success":true,"result":[{"id":"46f5a0897d1e2732300d6bf70d10f38e"}]}`: true,
		`{"success":true,"result":[]}`:                                          false,
		`{"success":false,"errors":[{"code":9109}]}`:                            false,
		`not json`: false,
	}
	for body, want := range cases {
		if got := zoneListHasResult([]byte(body)); got != want {
			t.Errorf("zoneListHasResult(%s) = %v, want %v", body, got, want)
		}
	}
}
