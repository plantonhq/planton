//go:build unix

package tofumodule

import (
	"context"
	"syscall"
	"testing"
	"time"
)

// TestNewReapableCommand_KillsProcessGroupOnCancel proves the load-bearing
// behavior of the cancellation-reaping fix WITHOUT needing a tofu binary:
// cancelling the context terminates the child's ENTIRE process group (the leader
// AND its descendants), not just the leader. This is exactly what stops a
// cancelled/superseded stack job from orphaning a tofu (or its provider plugins)
// that would keep holding the state lock.
//
// The command backgrounds a grandchild `sleep` in the same group so the group has
// more than one member; after cancel, no member may remain.
func TestNewReapableCommand_KillsProcessGroupOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := newReapableCommand(ctx, "sh", "-c", "sleep 60 & wait")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start command: %v", err)
	}

	// With Setpgid the leader's PID is the process-group id.
	pgid := cmd.Process.Pid
	if err := syscall.Kill(-pgid, 0); err != nil {
		t.Fatalf("expected process group %d to be alive before cancel, got: %v", pgid, err)
	}

	cancel()

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()
	select {
	case <-waitDone:
		// Wait returned after cancel -- the os/exec cancel path fired.
	case <-time.After(reapGraceWaitDelay + 5*time.Second):
		t.Fatalf("cmd.Wait did not return within the WaitDelay budget after cancel")
	}

	// The whole group must be gone. SIGTERM->exit is asynchronous, so poll briefly
	// until syscall.Kill(-pgid, 0) reports ESRCH (no such process group).
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := syscall.Kill(-pgid, 0); err == syscall.ESRCH {
			return // success: no process in the group remains
		}
		if time.Now().After(deadline) {
			t.Fatalf("process group %d still has live members after cancel", pgid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
