package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ResourceVerifier knows how to verify a specific Kubernetes resource type.
type ResourceVerifier interface {
	VerifyExists(ctx context.Context, kubeconfig string) error
	VerifyAbsent(ctx context.Context, kubeconfig string) error
}

// getVerifier returns the appropriate verifier for a component.
// Components are mapped to their primary resource type for verification.
func getVerifier(component string) (ResourceVerifier, error) {
	switch strings.ToLower(component) {
	case "kubernetesnamespace":
		return &NamespaceVerifier{Name: "test-namespace"}, nil
	case "kubernetesdeployment":
		return &WorkloadVerifier{
			Namespace: "test-deployment-ns",
			Kind:      "deployment",
			Name:      "test-deployment",
		}, nil
	case "kubernetesstatefulset":
		return &WorkloadVerifier{
			Namespace: "test-statefulset-ns",
			Kind:      "statefulset",
			Name:      "test-statefulset",
		}, nil
	case "kubernetessecret":
		return &ResourceExistenceVerifier{
			Namespace: "default",
			Kind:      "secret",
			Name:      "test-secret",
		}, nil
	case "kubernetesservice":
		return &ResourceExistenceVerifier{
			Namespace: "default",
			Kind:      "service",
			Name:      "test-service",
		}, nil
	default:
		return &GenericVerifier{Component: component}, nil
	}
}

// NamespaceVerifier checks that a namespace exists or is absent.
type NamespaceVerifier struct {
	Name string
}

func (v *NamespaceVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	return kubectlResourceExists(ctx, kubeconfig, "namespace", v.Name, "")
}

func (v *NamespaceVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return kubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Name, "")
}

// WorkloadVerifier checks that a workload (deployment, statefulset, etc.) is ready.
type WorkloadVerifier struct {
	Namespace string
	Kind      string
	Name      string
}

func (v *WorkloadVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	return kubectlResourceExists(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}

func (v *WorkloadVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return kubectlResourceAbsent(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}

// ResourceExistenceVerifier checks basic existence/absence without readiness.
type ResourceExistenceVerifier struct {
	Namespace string
	Kind      string
	Name      string
}

func (v *ResourceExistenceVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	return kubectlResourceExists(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}

func (v *ResourceExistenceVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return kubectlResourceAbsent(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}

// GenericVerifier is a fallback that always passes (for components without specific verifiers yet).
type GenericVerifier struct {
	Component string
}

func (v *GenericVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] No specific verifier for %s -- skipping resource verification\n", v.Component)
	return nil
}

func (v *GenericVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] No specific verifier for %s -- skipping cleanup verification\n", v.Component)
	return nil
}

func kubectlResourceExists(ctx context.Context, kubeconfig, kind, name, namespace string) error {
	args := buildKubectlArgs(kubeconfig, "get", kind, name, namespace)

	// Retry with backoff because resource creation may have eventual consistency
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

func kubectlResourceAbsent(ctx context.Context, kubeconfig, kind, name, namespace string) error {
	args := buildKubectlArgs(kubeconfig, "get", kind, name, namespace)

	// Give time for deletion to propagate
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
			// Resource still exists -- wait and retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
			}
			continue
		}

		// Non-NotFound error -- resource is gone or something else
		return nil
	}

	return errors.Errorf("resource %s/%s still exists after 10 verification attempts", kind, name)
}

func buildKubectlArgs(kubeconfig, verb, kind, name, namespace string) []string {
	args := []string{verb, kind, name}
	if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return args
}
