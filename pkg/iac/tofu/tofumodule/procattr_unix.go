//go:build unix

// Process-group control for the IaC child process on unix (linux, darwin).
// Splitting this by build tag keeps the planton CLI compiling on every platform
// while letting the runner reap the whole tofu process tree on cancellation.
package tofumodule

import (
	"os/exec"
	"syscall"
)

// setProcessGroup puts the child in its OWN process group. tofu spawns provider
// plugins as child processes; without a dedicated group, signalling only the
// leader would orphan the plugins (and the leader holds the state lock). With
// Setpgid the whole tree can be signalled together via the group.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// terminateProcessGroup graceful-stops the child's entire process group. The
// NEGATIVE pid targets the group whose leader is cmd.Process.Pid (established by
// setProcessGroup). SIGTERM lets tofu finish its current atomic state write and
// release the state lock; exec.Cmd.WaitDelay escalates to SIGKILL if the group
// ignores it. A vanished process (ESRCH) means it already exited -- treat as
// success so a benign cancel/exit race is not reported as an error.
func terminateProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		return err
	}
	return nil
}
