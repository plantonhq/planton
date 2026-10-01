package tofumodule

import (
	"fmt"
	"strings"
)

// maxStderrTail bounds the engine text a failure carries. The tail is where
// the engine's own sentence sits, and the error travels further than this
// process (a runner reports it in a size-limited workflow signal), so a
// noisy stderr never crowds out the rest of the report.
const maxStderrTail = 8 * 1024

// withStderr adds the engine's own words to a failed command. In JSON mode
// the engine reports most failures as JSON diagnostics on stdout, but the
// ones it raises before that stream starts -- a flag it refuses, a module it
// cannot parse at init, a crash -- exist only on stderr; without them the
// failure reads "exit status 1" and the person is left to find a log they
// may not be able to see. The text follows a "stderr:" label, the same
// shape the Pulumi Automation API gives its errors, so one reader serves
// both engines.
func withStderr(err error, stderr string) error {
	if err == nil {
		return nil
	}
	tail := strings.TrimSpace(stderr)
	if tail == "" {
		return err
	}
	if len(tail) > maxStderrTail {
		tail = tail[len(tail)-maxStderrTail:]
		if i := strings.IndexByte(tail, '\n'); i >= 0 && i < len(tail)-1 {
			tail = tail[i+1:]
		}
	}
	return fmt.Errorf("%w\nstderr: %s", err, tail)
}
