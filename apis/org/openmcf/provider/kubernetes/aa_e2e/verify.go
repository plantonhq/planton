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

// operatorKinds lists manifest kind values (lowercased) for operator/controller
// components. Operators install CRD controllers that watch resources but typically
// do not expose a Kubernetes Service. Verification checks namespace + running
// pods only (no service requirement).
var operatorKinds = map[string]bool{
	"kubernetesperconamongooperator":    true,
	"kubernetesperconamysqloperator":    true,
	"kubernetesperconapostgresoperator": true,
	"kubernetessolroperator":            true,
	"kubernetesaltinityoperator":        true,
	"kuberneteselasticoperator":         true,
	"kubernetesstrimzikafkaoperator":    true,
	"kuberneteszalandopostgresoperator": true,
}

// crdWorkloadKinds lists manifest kind values (lowercased) for Tier 3
// operator-dependent components. These create Custom Resources (e.g.,
// Zalando Postgresql, Strimzi Kafka, ECK Elasticsearch) that are
// reconciled by their prerequisite operator into pods and services.
// Verification checks namespace + running pods + at least one service.
var crdWorkloadKinds = map[string]bool{
	"kubernetespostgres":      true,
	"kuberneteskafka":         true,
	"kuberneteselasticsearch": true,
	"kubernetesmongodb":       true,
	"kubernetessolr":          true,
	"kubernetesclickhouse":    true,
}

// helmTier2Kinds lists all manifest kind values (lowercased) for Helm-based
// Kubernetes components (Tier 2) that deploy applications with Services.
// These must match the CloudResourceKind enum names from cloud_resource_kind.proto
// (case-insensitive via lowercasing).
//
// Historical hack manifests use inconsistent kind names (e.g., "RedisKubernetes"
// instead of "KubernetesRedis"). E2E manifests must use the enum name. Both
// conventions are included here so the verifier works with either.
var helmTier2Kinds = map[string]bool{
	"kubernetesredis":     true,
	"kubernetesnats":      true,
	"kubernetesgrafana":   true,
	"kubernetesneo4j":     true,
	"kubernetesopenbao":   true,
	"kubernetesopenfga":   true,
	"kubernetesjenkins":   true,
	"kubernetestemporal":  true,
	"kubernetesargocd":    true,
	"kubernetesharbor":    true,
	"kubernetesgitlab":    true,
	"kuberneteslocust":  true,
	"kubernetessignoz":  true,
	"kuberneteskeycloak": true,
	// Legacy hack manifest kind names (lowercased)
	"rediskubernetes":    true,
	"harborkubernetes":   true,
	"locustkubernetes":   true,
	"signozkubernetes":   true,
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
		if operatorKinds[component] {
			return &OperatorComponentVerifier{
				Namespace:     info.Namespace,
				ComponentName: info.Name,
			}, nil
		}
		if crdWorkloadKinds[component] {
			return &CRDWorkloadVerifier{
				Namespace:     info.Namespace,
				ComponentName: info.Name,
			}, nil
		}
		if helmTier2Kinds[component] {
			return &HelmComponentVerifier{
				Namespace:     info.Namespace,
				ComponentName: info.Name,
			}, nil
		}
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

// OperatorComponentVerifier checks operator/controller components by verifying
// that the namespace exists and at least one Pod is Running. Operators install
// CRD controllers that do not expose Services, so service checks are omitted.
type OperatorComponentVerifier struct {
	Namespace     string
	ComponentName string
}

func (v *OperatorComponentVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] Operator component %q in namespace %q\n", v.ComponentName, v.Namespace)

	if err := kubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for operator component %q", v.Namespace, v.ComponentName)
	}

	if err := kubectlPodsRunningInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no running pods in namespace %q for operator component %q", v.Namespace, v.ComponentName)
	}

	return nil
}

func (v *OperatorComponentVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return kubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Namespace, "")
}

// CRDWorkloadVerifier checks Tier 3 operator-dependent components. These
// create Custom Resources (e.g., Zalando Postgresql, Strimzi Kafka) that an
// operator reconciles into pods and services. Verification checks namespace
// exists, at least one pod Running, and at least one service present. Uses
// the same retry windows as HelmComponentVerifier because CRD reconciliation
// takes comparable time to Helm chart startup.
type CRDWorkloadVerifier struct {
	Namespace     string
	ComponentName string
}

func (v *CRDWorkloadVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] CRD workload %q in namespace %q\n", v.ComponentName, v.Namespace)

	if err := kubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for CRD workload %q", v.Namespace, v.ComponentName)
	}

	if err := kubectlPodsRunningInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no running pods in namespace %q for CRD workload %q", v.Namespace, v.ComponentName)
	}

	if err := kubectlServicesExistInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no services in namespace %q for CRD workload %q", v.Namespace, v.ComponentName)
	}

	return nil
}

func (v *CRDWorkloadVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return kubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Namespace, "")
}

// HelmComponentVerifier checks Helm-based (Tier 2) components by verifying
// that the namespace exists, at least one Pod is Running, and at least one
// Service is present. This avoids coupling to chart-internal resource names.
type HelmComponentVerifier struct {
	Namespace     string
	ComponentName string
}

func (v *HelmComponentVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] Helm component %q in namespace %q\n", v.ComponentName, v.Namespace)

	if err := kubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for helm component %q", v.Namespace, v.ComponentName)
	}

	if err := kubectlPodsRunningInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no running pods in namespace %q for helm component %q", v.Namespace, v.ComponentName)
	}

	if err := kubectlServicesExistInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no services in namespace %q for helm component %q", v.Namespace, v.ComponentName)
	}

	return nil
}

func (v *HelmComponentVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return kubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Namespace, "")
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

// kubectlPodsRunningInNamespace waits for at least one Pod in the namespace to
// reach Running phase. Helm-based components need longer to start because they
// pull container images and may run init containers.
func kubectlPodsRunningInNamespace(ctx context.Context, kubeconfig, namespace string) error {
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

// kubectlServicesExistInNamespace checks that at least one Service exists in the namespace.
func kubectlServicesExistInNamespace(ctx context.Context, kubeconfig, namespace string) error {
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
