package tofumodule

import (
	"context"
	"os/exec"
	"time"
)

// reapGraceWaitDelay bounds how long exec.Cmd.Wait lingers after ctx cancellation
// (and the SIGTERM that cmd.Cancel sends to the process group) before the os/exec
// machinery force-kills the child. It is a safety net: tofu normally exits within
// a second of SIGTERM, but a wedged child (or one whose stdout pipe is still held
// by a lingering grandchild) must not block Wait forever.
const reapGraceWaitDelay = 10 * time.Second

// newReapableCommand builds an exec.Cmd bound to ctx whose ENTIRE process group is
// terminated when ctx is cancelled.
//
// Why this exists: a stack job that is cancelled/superseded (e.g. an undeploy
// supersedes a deploy whose tofu is still polling ACM cert validation) must not
// leave an orphaned tofu behind. On the local backend that orphan keeps holding
// the state flock and wedges the next operation; on remote backends it holds the
// remote lock. Binding the command to ctx + putting it in its own process group +
// signalling the group on cancel guarantees the holder dies and the lock frees.
func newReapableCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	setProcessGroup(cmd)
	// Override exec's default cancel (which kills only the leader) to take down the
	// whole group, so provider-plugin children are reaped too.
	cmd.Cancel = func() error { return terminateProcessGroup(cmd) }
	cmd.WaitDelay = reapGraceWaitDelay
	return cmd
}
