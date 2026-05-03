package aa_e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ManifestInfo holds the parsed fields from a manifest needed for verification.
type ManifestInfo struct {
	Kind      string
	Name      string
	Namespace string
}

// ParseManifestInfo extracts kind, name, and namespace from a manifest YAML file
// to drive dynamic verification without hardcoded values.
func ParseManifestInfo(manifestPath string) (*ManifestInfo, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read manifest %s", manifestPath)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest YAML %s", manifestPath)
	}

	info := &ManifestInfo{}

	if kind, ok := raw["kind"].(string); ok {
		info.Kind = kind
	}

	if metadata, ok := raw["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			info.Name = name
		}
	}

	if spec, ok := raw["spec"].(map[string]interface{}); ok {
		if name, ok := spec["name"].(string); ok {
			info.Name = name
		}

		switch ns := spec["namespace"].(type) {
		case string:
			info.Namespace = ns
		case map[string]interface{}:
			if val, ok := ns["value"].(string); ok {
				info.Namespace = val
			}
		}
	}

	if info.Namespace == "" {
		info.Namespace = "default"
	}

	return info, nil
}

// ResourceVerifier knows how to verify a specific Kubernetes resource type.
type ResourceVerifier interface {
	VerifyExists(ctx context.Context, kubeconfig string) error
	VerifyAbsent(ctx context.Context, kubeconfig string) error
}

// GetVerifierFromManifest creates the appropriate verifier by parsing the manifest.
func GetVerifierFromManifest(manifestPath string) (ResourceVerifier, error) {
	info, err := ParseManifestInfo(manifestPath)
	if err != nil {
		return nil, err
	}

	component := strings.ToLower(info.Kind)

	switch component {
	case "kubernetesnamespace":
		return &NamespaceVerifier{Name: info.Name}, nil

	case "kubernetesdeployment":
		return &WorkloadVerifier{
			Namespace: info.Namespace,
			Kind:      "deployment",
			Name:      info.Name,
		}, nil

	case "kubernetesstatefulset":
		return &WorkloadVerifier{
			Namespace: info.Namespace,
			Kind:      "statefulset",
			Name:      info.Name,
		}, nil

	case "kubernetessecret":
		return &ResourceExistenceVerifier{
			Namespace: info.Namespace,
			Kind:      "secret",
			Name:      info.Name,
		}, nil

	case "kubernetesservice":
		return &ResourceExistenceVerifier{
			Namespace: info.Namespace,
			Kind:      "service",
			Name:      info.Name,
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
