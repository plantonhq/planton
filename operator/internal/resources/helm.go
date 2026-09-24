package resources

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
)

// RenderHelmChart loads a Helm chart from an in-memory .tgz archive, renders
// it with the given values, and returns the resulting Kubernetes objects as
// unstructured resources ready for Server-Side Apply.
//
// Rendering uses the Helm action package (equivalent to `helm template`) rather
// than the raw engine, so subchart dependency conditions defined in Chart.yaml
// (e.g., cassandra.enabled, postgresql.enabled) are correctly evaluated. Charts
// with disabled subcharts will not produce resources for those subcharts.
//
// The releaseName and namespace are used in chart template rendering (they
// populate .Release.Name and .Release.Namespace respectively). Only manifests
// that decode to valid Kubernetes objects are returned; empty documents, NOTES
// files, and Helm test resources are filtered out.
func RenderHelmChart(chartData []byte, releaseName, namespace string, values map[string]any) ([]*unstructured.Unstructured, error) {
	chart, err := loader.LoadArchive(bytes.NewReader(chartData))
	if err != nil {
		return nil, fmt.Errorf("loading helm chart archive: %w", err)
	}

	client := action.NewInstall(&action.Configuration{})
	client.DryRun = true
	client.ClientOnly = true
	client.ReleaseName = releaseName
	client.Namespace = namespace
	client.Replace = true
	client.IncludeCRDs = false
	client.KubeVersion = &chartutil.KubeVersion{
		Version: "v1.31.0",
		Major:   "1",
		Minor:   "31",
	}

	rel, err := client.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("rendering helm chart templates: %w", err)
	}

	var allManifests []string
	if m := strings.TrimSpace(rel.Manifest); m != "" {
		allManifests = append(allManifests, m)
	}
	for _, hook := range rel.Hooks {
		if isTestHook(hook.Events) {
			continue
		}
		if m := strings.TrimSpace(hook.Manifest); m != "" {
			allManifests = append(allManifests, m)
		}
	}

	if len(allManifests) == 0 {
		return nil, nil
	}

	var result []*unstructured.Unstructured
	for _, manifest := range allManifests {
		objs, err := parseMultiDocYAML([]byte(manifest))
		if err != nil {
			return nil, fmt.Errorf("parsing rendered manifests: %w", err)
		}
		for _, obj := range objs {
			if obj.GetNamespace() == "" && IsNamespacedKind(obj.GetKind()) {
				obj.SetNamespace(namespace)
			}
			result = append(result, obj)
		}
	}

	return result, nil
}

// shouldSkipRenderedFile returns true for Helm output files that should not
// be applied as Kubernetes resources (NOTES.txt, test manifests).
func shouldSkipRenderedFile(name string) bool {
	if strings.HasSuffix(name, "NOTES.txt") {
		return true
	}
	if strings.Contains(name, "/tests/") {
		return true
	}
	return false
}

// isTestHook returns true if the hook events include only "test" events.
// Test hooks should not be applied as regular resources.
func isTestHook(events []release.HookEvent) bool {
	return slices.Contains(events, release.HookTest)
}

// IsNamespacedKind reports whether objects of this kind live in a namespace.
// It is the operator's one list of the cluster-scoped kinds a chart or a
// vendored manifest can render: the render step uses it to place namespaced
// objects in the release namespace, and the platform-owned apply uses it to
// refuse ownership of anything a namespaced owner cannot garbage-collect.
func IsNamespacedKind(kind string) bool {
	switch kind {
	case "Namespace", "ClusterRole", "ClusterRoleBinding",
		"CustomResourceDefinition", "PriorityClass",
		"PersistentVolume", "StorageClass", "IngressClass",
		"ValidatingWebhookConfiguration", "MutatingWebhookConfiguration",
		"APIService", "CSIDriver", "RuntimeClass", "GatewayClass":
		return false
	default:
		return true
	}
}

// helmResourceValues turns a typed ResourceRequirements into the values shape
// every chart the operator renders reads under its `resources` key
// ({requests: {cpu, memory}, limits: {cpu, memory}}). Unset quantities are
// omitted, never rendered as empty strings, so "no CPU limit" stays exactly
// that -- the house pattern is requests for both, a memory limit, and no CPU
// limit. One conversion, so a component's sizing is declared once as typed
// Kubernetes quantities whether it renders a Deployment or a chart.
func helmResourceValues(r corev1.ResourceRequirements) map[string]any {
	toValues := func(list corev1.ResourceList) map[string]any {
		out := map[string]any{}
		if q, ok := list[corev1.ResourceCPU]; ok && !q.IsZero() {
			out["cpu"] = q.String()
		}
		if q, ok := list[corev1.ResourceMemory]; ok && !q.IsZero() {
			out["memory"] = q.String()
		}
		return out
	}
	values := map[string]any{}
	if requests := toValues(r.Requests); len(requests) > 0 {
		values["requests"] = requests
	}
	if limits := toValues(r.Limits); len(limits) > 0 {
		values["limits"] = limits
	}
	return values
}
