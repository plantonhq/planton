package runner

import (
	"fmt"
	"testing"
)

// lanef writes one progress line of a lane through the lane's own test log,
// so lanes running side by side in one process keep their lines apart: go
// test labels each t.Log line with the test that wrote it, while a bare
// fmt.Printf from two lanes lands on stdout unattributed. A nil t (a unit
// test driving one helper directly) prints plainly. The format carries no
// trailing newline; t.Logf adds its own.
func lanef(t testing.TB, format string, args ...interface{}) {
	if t == nil {
		fmt.Printf(format+"\n", args...)
		return
	}
	t.Helper()
	t.Logf(format, args...)
}
