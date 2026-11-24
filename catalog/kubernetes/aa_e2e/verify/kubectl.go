package verify

import (
	"bytes"
	"context"
	"fmt"
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

// podUid reads a pod's uid. Capture it BEFORE a deliberate pod deletion:
// the uid is the only property a same-named StatefulSet replacement can
// never share with the pod it replaces.
func podUid(ctx context.Context, kubeconfig, namespace, pod string) (string, error) {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", "pod", pod, "-n", namespace, "-o", "jsonpath={.metadata.uid}").Output()
	if err != nil {
		return "", errors.Wrapf(err, "reading pod %s/%s uid", namespace, pod)
	}
	return strings.TrimSpace(string(out)), nil
}

// waitForPodReplaced waits until the named pod exists with a uid DIFFERENT
// from oldUid — the only honest signal that a deleted StatefulSet pod's
// replacement is present. A terminating pod keeps reporting phase Running
// (and can remain Ready) for its whole grace period, and the StatefulSet's
// own status lags the deletion, so any wait keyed on phase/readiness alone
// returns instantly against the DYING pod (verified live — the follow-up
// probe then raced the old pod).
//
// requireReady additionally gates on the Ready condition. Pass false for
// pods whose readiness is deliberately gated on out-of-band bootstrap
// (e.g. a sealed-by-design secrets server that only turns Ready after an
// unseal the caller performs next): those replacements reach phase Running
// but are NEVER Ready on their own.
func waitForPodReplaced(ctx context.Context, kubeconfig, namespace, pod, oldUid string, requireReady bool, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "pod", pod, "-n", namespace,
			"-o", "jsonpath={.metadata.uid} {.status.phase} {.status.conditions[?(@.type=='Ready')].status}").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 2 && fields[0] != oldUid {
				if !requireReady && fields[1] == "Running" {
					return nil
				}
				if requireReady && len(fields) >= 3 && fields[2] == "True" {
					return nil
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
	gate := "Running"
	if requireReady {
		gate = "Ready"
	}
	return errors.Errorf("pod %s/%s never came back %s with a new uid within %s", namespace, pod, gate, budget)
}

// kubectlGetJSONPath reads one jsonpath expression off a live object,
// retrying while the object is still materializing. Returns the raw
// (untrimmed) kubectl output.
func kubectlGetJSONPath(ctx context.Context, kubeconfig, kind, name, namespace, jsonPath string) (string, error) {
	args := BuildKubectlArgs(kubeconfig, "get", kind, name, namespace)
	args = append(args, "-o", "jsonpath="+jsonPath)

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			return stdout.String(), nil
		} else {
			lastErr = errors.Wrapf(err, "kubectl get %s %s: %s", kind, name, stderr.String())
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
		}
	}

	return "", errors.Wrapf(lastErr, "resource %s/%s not readable after 5 attempts", kind, name)
}

// plantedStandInImage is the image every e2e fixture runs when it stands in
// for another installation, or for an earlier release of the kind, in a
// lane whose deploy must be refused.
const plantedStandInImage = "registry.k8s.io/pause:"

// KubectlDeploymentLeftAsPlanted proves that after a refused deploy a
// Deployment of this name is gone, or is still the planted stand-in: every
// container still runs the stand-in's image, so the refused deploy applied
// nothing over it. The stand-in stays until its own prerequisite is torn
// down.
func KubectlDeploymentLeftAsPlanted(ctx context.Context, kubeconfig, name, namespace string) error {
	args := append(BuildKubectlArgs(kubeconfig, "get", "deployment", name, namespace),
		"--ignore-not-found", "-o", "jsonpath={.spec.template.spec.containers[*].image}")
	out, err := exec.CommandContext(ctx, "kubectl", args...).Output()
	if err != nil {
		return errors.Wrapf(err, "reading deployment %s/%s after the refused deploy", namespace, name)
	}
	images := strings.Fields(string(out))
	if len(images) == 0 {
		return nil
	}
	for _, image := range images {
		if !strings.HasPrefix(image, plantedStandInImage) {
			return errors.Errorf("deployment %s/%s runs %s after a refused deploy, which must apply nothing: the planted stand-in was replaced", namespace, name, image)
		}
	}
	fmt.Printf("  [verify] DESTROY: deployment %s/%s is still the planted stand-in, left as it was\n", namespace, name)
	return nil
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
	// An empty name addresses every object of the kind (a list), never an
	// empty-name argument kubectl refuses ("resource name may not be empty").
	args := []string{verb, kind}
	if name != "" {
		args = append(args, name)
	}
	if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return args
}
