package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kubernetesmetricsserverv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmetricsserver/v1alpha1"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace).
//
// PARITY: the Terraform module reaches the same result natively — its
// helm_release passes values = [yamlencode(typed values), helm_values] and
// the provider merges the documents in exactly this order. Keep every typed
// mapping below in lockstep with the Terraform module's locals.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec

	values := map[string]interface{}{
		// Pin the chart's fullname to the (fixed) release name: chart
		// objects get deterministic names ("metrics-server", the Service
		// the APIService routes to) — what verification and imports key
		// off.
		"fullnameOverride": vars.ReleaseName,
	}

	if spec.Replicas != nil {
		values["replicas"] = spec.GetReplicas()
	}

	// ---- kubelet scrape flags ---------------------------------------------
	// The chart concatenates defaultArgs + args into the container command
	// line, and metrics-server's flag parsing rejects nothing on
	// duplicates — the LAST occurrence silently wins. Rendering a
	// duplicate --metric-resolution into args would therefore work but
	// leave a confusing double flag in the pod spec. Instead the module
	// OWNS defaultArgs: it re-renders the chart's default list with the
	// typed substitutions applied, keeping the pod spec canonical.
	defaultArgs := []interface{}{
		"--cert-dir=/tmp",
		fmt.Sprintf("--kubelet-preferred-address-types=%s", kubeletAddressTypes(spec)),
		"--kubelet-use-node-status-port",
		fmt.Sprintf("--metric-resolution=%s", metricResolution(spec)),
	}
	values["defaultArgs"] = defaultArgs

	if spec.GetKubeletInsecureTls() {
		values["args"] = []interface{}{"--kubelet-insecure-tls"}
	}

	if spec.GetHostNetwork() {
		values["hostNetwork"] = map[string]interface{}{"enabled": true}
	}

	// ---- APIService ---------------------------------------------------------
	if a := spec.GetApiService(); a != nil {
		apiService := map[string]interface{}{}
		if a.Create != nil && !a.GetCreate() {
			apiService["create"] = false
		}
		if a.InsecureSkipTlsVerify != nil && !a.GetInsecureSkipTlsVerify() {
			apiService["insecureSkipTLSVerify"] = false
		}
		if a.GetCaBundle() != "" {
			apiService["caBundle"] = a.GetCaBundle()
		}
		if len(apiService) > 0 {
			values["apiService"] = apiService
		}
	}

	// ---- serving certificate -------------------------------------------------
	if tlsValues := tlsMap(spec.GetTls()); tlsValues != nil {
		values["tls"] = tlsValues
	}

	// ---- sizing / scheduling ---------------------------------------------------
	if r := resourcesMap(spec.GetResources()); r != nil {
		values["resources"] = r
	}
	if len(spec.GetNodeSelector()) > 0 {
		values["nodeSelector"] = stringMapToInterface(spec.GetNodeSelector())
	}
	if len(spec.GetTolerations()) > 0 {
		values["tolerations"] = tolerationsSlice(spec.GetTolerations())
	}
	if spec.PriorityClassName != nil && spec.GetPriorityClassName() != "" {
		values["priorityClassName"] = spec.GetPriorityClassName()
	}

	if spec.GetPodDisruptionBudget() {
		values["podDisruptionBudget"] = map[string]interface{}{
			"enabled":      true,
			"minAvailable": 1,
		}
	}

	// ---- own telemetry -----------------------------------------------------------
	if p := spec.GetPrometheus(); p.GetEnabled() {
		values["metrics"] = map[string]interface{}{"enabled": true}
		if p.GetServiceMonitor() {
			sm := map[string]interface{}{"enabled": true}
			if p.GetServiceMonitorInterval() != "" {
				sm["interval"] = p.GetServiceMonitorInterval()
			}
			if len(p.GetServiceMonitorLabels()) > 0 {
				sm["additionalLabels"] = stringMapToInterface(p.GetServiceMonitorLabels())
			}
			values["serviceMonitor"] = sm
		}
	}

	// ---- image --------------------------------------------------------------------
	if img := spec.GetImage(); img != nil && (img.GetRepository() != "" || img.GetTag() != "") {
		image := map[string]interface{}{}
		if img.GetRepository() != "" {
			image["repository"] = img.GetRepository()
		}
		if img.GetTag() != "" {
			image["tag"] = img.GetTag()
		}
		values["image"] = image
	}

	// ---- escape hatch (merged LAST, helm -f semantics) ------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	return values, nil
}

// kubeletAddressTypes resolves the preferred-address-types flag value —
// the spec list when set, otherwise the chart's default order.
func kubeletAddressTypes(spec *kubernetesmetricsserverv1alpha1.KubernetesMetricsServerSpec) string {
	if types := spec.GetKubeletPreferredAddressTypes(); len(types) > 0 {
		return strings.Join(types, ",")
	}
	return "InternalIP,ExternalIP,Hostname"
}

// metricResolution resolves the scrape interval — the spec value when set,
// otherwise the chart's default.
func metricResolution(spec *kubernetesmetricsserverv1alpha1.KubernetesMetricsServerSpec) string {
	if spec.MetricResolution != nil && spec.GetMetricResolution() != "" {
		return spec.GetMetricResolution()
	}
	return "15s"
}

// tlsMap renders the serving-certificate configuration. Returns nil for the
// default (metrics-server self-signed) with no issuer/secret set — the
// chart's own default needs no values.
func tlsMap(tls *kubernetesmetricsserverv1alpha1.KubernetesMetricsServerTls) map[string]interface{} {
	if tls == nil {
		return nil
	}
	switch tls.GetType() {
	case kubernetesmetricsserverv1alpha1.KubernetesMetricsServerTlsType_helm:
		return map[string]interface{}{"type": "helm"}
	case kubernetesmetricsserverv1alpha1.KubernetesMetricsServerTlsType_cert_manager:
		out := map[string]interface{}{"type": "cert-manager"}
		if issuer := tls.GetCertManagerIssuer(); issuer != nil {
			kind, name := issuer.CRDKind()
			out["certManager"] = map[string]interface{}{
				"existingIssuer": map[string]interface{}{
					"enabled": true,
					"kind":    kind,
					"name":    name,
				},
			}
		}
		return out
	case kubernetesmetricsserverv1alpha1.KubernetesMetricsServerTlsType_existing_secret:
		return map[string]interface{}{
			"type": "existingSecret",
			"existingSecret": map[string]interface{}{
				"name": tls.GetExistingSecretName().GetValue(),
			},
		}
	default:
		// self_signed — the chart default ("metrics-server"); nothing to
		// render.
		return nil
	}
}
