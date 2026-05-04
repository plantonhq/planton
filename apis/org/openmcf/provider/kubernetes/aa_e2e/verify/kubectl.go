package verify

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// KubectlPodsRunningInNamespace waits for at least one Pod in the namespace to
// reach Running phase. Uses progressive backoff with 15 attempts because
// Helm charts and CRD workloads need time to pull images and start.
func KubectlPodsRunningInNamespace(ctx context.Context, kubeconfig, namespace string) error {
	args := []string{"get", "pods", "-n", namespace,
		"--field-selector=status.phase=Running",
		"-o", "jsonpath={.items}",
		"--no-headers",
	}
	if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}

	var lastErr error
	for attempt := 0; attempt < 15; attempt++ {
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			lastErr = errors.Wrapf(err, "kubectl get pods: %s", stderr.String())
		} else {
			output := strings.TrimSpace(stdout.String())
			if output != "" && output != "[]" {
				return nil
			}
			lastErr = errors.Errorf("no running pods in namespace %s (output: %q)", namespace, output)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 3 * time.Second):
		}
	}

	return errors.Wrapf(lastErr, "no running pods in namespace %s after 15 attempts", namespace)
}

// KubectlServicesExistInNamespace checks that at least one Service exists in the namespace.
func KubectlServicesExistInNamespace(ctx context.Context, kubeconfig, namespace string) error {
	args := []string{"get", "svc", "-n", namespace,
		"-o", "jsonpath={.items[*].metadata.name}",
		"--no-headers",
	}
	if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}

	var lastErr error
	for attempt := 0; attempt < 10; attempt++ {
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			lastErr = errors.Wrapf(err, "kubectl get svc: %s", stderr.String())
		} else {
			output := strings.TrimSpace(stdout.String())
			if output != "" {
				return nil
			}
			lastErr = errors.Errorf("no services in namespace %s", namespace)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
		}
	}

	return errors.Wrapf(lastErr, "no services in namespace %s after 10 attempts", namespace)
}

// KubectlResourceExists waits for a specific resource to exist, retrying up to 5 times.
func KubectlResourceExists(ctx context.Context, kubeconfig, kind, name, namespace string) error {
	args := BuildKubectlArgs(kubeconfig, "get", kind, name, namespace)

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			return nil
		} else {
			lastErr = errors.Wrapf(err, "kubectl get %s %s: %s", kind, name, stderr.String())
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
		}
	}

	return errors.Wrapf(lastErr, "resource %s/%s not found after 5 attempts", kind, name)
}

// KubectlResourceAbsent waits for a specific resource to be gone, retrying up to 10 times.
func KubectlResourceAbsent(ctx context.Context, kubeconfig, kind, name, namespace string) error {
	args := BuildKubectlArgs(kubeconfig, "get", kind, name, namespace)

	for attempt := 0; attempt < 10; attempt++ {
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			stderrStr := stderr.String()
			if strings.Contains(stderrStr, "NotFound") || strings.Contains(stderrStr, "not found") {
				return nil
			}
		}

		if err == nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
			}
			continue
		}

		return nil
	}

	return errors.Errorf("resource %s/%s still exists after 10 verification attempts", kind, name)
}

// BuildKubectlArgs constructs kubectl command arguments with optional kubeconfig and namespace.
func BuildKubectlArgs(kubeconfig, verb, kind, name, namespace string) []string {
	args := []string{verb, kind, name}
	if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return args
}
