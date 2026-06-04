package tofumodule

import (
	"bufio"
	"fmt"
	"os/exec"
	"runtime/debug"

	"github.com/pkg/errors"
)

// maxScanTokenSize bounds a single stdout line read from the IaC binary.
// tofu/terraform `-json` lines (notably plan output with large resource diffs)
// routinely exceed bufio.Scanner's 64KB default, which would otherwise surface
// as a sporadic "bufio.Scanner: token too long" failure on the same streaming
// path as the close/read race fixed here.
const maxScanTokenSize = 10 * 1024 * 1024

// streamCommandJSONOutput starts an already-configured command, streams each
// stdout line to lines (or prints to stdout when lines is nil), and returns only
// after the process has exited.
//
// WHY the read must complete before Wait: exec.Cmd.Wait closes the read end of
// the StdoutPipe as soon as the process exits. If a scanner read is still in
// flight when that happens, the read fails with "read |0: file already closed"
// and a successful tofu/terraform run is reported as a failure. The fix is to
// drain the reader goroutine fully (block on errChan, which the goroutine closes
// at EOF) and only then call Wait. See the exec.Cmd.StdoutPipe docs.
//
// binaryName is used only to label the read error ("tofu" or "terraform").
func streamCommandJSONOutput(binaryName string, cmd *exec.Cmd, lines chan<- string) (err error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to create stdout pipe")
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "failed to start %s command %s", binaryName, cmd.String())
	}

	// errChan carries at most one error (a panic or a scanner error). Its close
	// is the signal that the reader has drained the pipe to EOF.
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf(
					"panic recovered in %s stdout reader: %v\nstack trace:\n%s",
					binaryName, r, string(debug.Stack()),
				)
			}
			close(errChan)
		}()

		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxScanTokenSize)
		for scanner.Scan() {
			line := scanner.Text()
			if lines != nil {
				lines <- line
			} else {
				fmt.Println(line)
			}
		}
		if scanErr := scanner.Err(); scanErr != nil {
			errChan <- fmt.Errorf("error reading %s output: %v", binaryName, scanErr)
		}
	}()

	// Block until every read has completed; only then is it safe to Wait.
	readErr := <-errChan

	// Always Wait so the process is reaped even on a read error. A non-nil Wait
	// (non-zero exit) takes precedence over a read error, matching prior behavior.
	if waitErr := cmd.Wait(); waitErr != nil {
		return errors.Wrapf(waitErr, "failed to execute %s command %s", binaryName, cmd.String())
	}

	return readErr
}
