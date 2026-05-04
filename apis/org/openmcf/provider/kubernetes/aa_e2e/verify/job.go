package verify

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// JobVerifier checks that a batch/v1 Job exists and reaches completion,
// or is absent after destroy. Jobs run immediately upon creation, so
// waiting for the Complete condition validates that the job spec
// (image, command, resources) was correct — not just that the resource exists.
type JobVerifier struct {
	Namespace string
	Name      string
}

func (v *JobVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceExists(ctx, kubeconfig, "job", v.Name, v.Namespace); err != nil {
		return err
	}
	return KubectlJobComplete(ctx, kubeconfig, v.Name, v.Namespace)
}

func (v *JobVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "job", v.Name, v.Namespace)
}

// KubectlJobComplete waits for a Job to reach the Complete condition.
// Uses 10 attempts with 5-second progressive backoff. Busybox jobs typically
// complete in <5s; this gives up to ~275s for heavier workloads.
func KubectlJobComplete(ctx context.Context, kubeconfig, name, namespace string) error {
	args := []string{
		"get", "job", name,
		"-o", `jsonpath={.status.conditions[?(@.type=="Complete")].status}`,
	}
	if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	var lastErr error
	for attempt := 0; attempt < 10; attempt++ {
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			lastErr = errors.Wrapf(err, "kubectl get job %s condition: %s", name, stderr.String())
		} else {
			output := strings.TrimSpace(stdout.String())
			if output == "True" {
				return nil
			}
			lastErr = errors.Errorf("job %s not complete yet (condition status: %q)", name, output)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 5 * time.Second):
		}
	}

	return errors.Wrapf(lastErr, "job %s in namespace %s did not complete after 10 attempts", name, namespace)
}
