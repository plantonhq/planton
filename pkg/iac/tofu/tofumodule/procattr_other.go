//go:build !unix

// Non-unix fallback for IaC child-process control (keeps the CLI building on
// platforms without POSIX process groups). The runner only ships on unix, so the
// group-reaping path above is the one that matters operationally.
package tofumodule

import "os/exec"

// setProcessGroup is a no-op where POSIX process groups are unavailable.
func setProcessGroup(cmd *exec.Cmd) {}

// terminateProcessGroup falls back to killing just the leader process.
func terminateProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
