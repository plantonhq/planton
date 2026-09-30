package module

import (
	"github.com/pkg/errors"
	kuberneteskedav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskeda/v1alpha1"
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
//
// No fullnameOverride: the chart names its components keda-operator /
// keda-operator-metrics-apiserver / keda-admission-webhooks independent of
// the release name — there is nothing for an override to pin.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec

	values := map[string]interface{}{}

	// ---- CRD lifecycle ------------------------------------------------------
	// crds.install matches the chart's own default (true) — rendered only on
	// explicit opt-out. keep_on_uninstall has NO chart knob: the chart
	// templates its CRDs and Helm would DELETE them on uninstall, cascading
	// to every ScaledObject/ScaledJob/TriggerAuthentication in the cluster.
	// Planton default keeps them via the standard Helm resource-policy
	// annotation, which the chart forwards onto the CRDs
	// (crds.additionalAnnotations) — the ESO-family precedent. The keep
	// annotation only makes sense when this release owns the CRDs, so it
	// rides along only when install && keep.
	crdsInstall := true
	if spec.GetCrds() != nil && spec.GetCrds().Install != nil {
		crdsInstall = spec.GetCrds().GetInstall()
	}
	crdsKeep := true
	if spec.GetCrds() != nil && spec.GetCrds().KeepOnUninstall != nil {
		crdsKeep = spec.GetCrds().GetKeepOnUninstall()
	}
	crds := map[string]interface{}{}
	if !crdsInstall {
		crds["install"] = false
	}
	if crdsInstall && crdsKeep {
		crds["additionalAnnotations"] = map[string]interface{}{
			"helm.sh/resource-policy": "keep",
		}
	}
	if len(crds) > 0 {
		values["crds"] = crds
	}

	// ---- watch scope ---------------------------------------------------------
	if spec.GetWatchNamespace() != "" {
		values["watchNamespace"] = spec.GetWatchNamespace()
	}

	// ---- component sizing ------------------------------------------------------
	// The chart's key layout is ASYMMETRIC: replica counts live under each
	// component (operator.replicaCount, ...) while resources are grouped
	// under a shared top-level resources block keyed per component — and
	// the metrics server's key there is "metricServer" (SINGULAR), unlike
	// the "metricsServer" component block. Render both halves here and
	// attach the shared resources block once at the end.
	resources := map[string]interface{}{}

	if o := spec.GetOperator(); o != nil {
		if o.Replicas != nil {
			values["operator"] = map[string]interface{}{
				"replicaCount": int(o.GetReplicas()),
			}
		}
		if r := resourcesMap(o.GetResources()); r != nil {
			resources["operator"] = r
		}
	}

	if m := spec.GetMetricsServer(); m != nil {
		if m.Replicas != nil {
			values["metricsServer"] = map[string]interface{}{
				"replicaCount": int(m.GetReplicas()),
			}
		}
		if r := resourcesMap(m.GetResources()); r != nil {
			resources["metricServer"] = r
		}
	}

	// ---- admission webhooks -----------------------------------------------------
	if w := spec.GetWebhooks(); w != nil {
		webhooks := map[string]interface{}{}
		// enabled matches the chart's own default (true) — rendered only
		// on explicit opt-out.
		if w.Enabled != nil && !w.GetEnabled() {
			webhooks["enabled"] = false
		}
		if w.FailurePolicy != nil && w.GetFailurePolicy() != "" {
			webhooks["failurePolicy"] = w.GetFailurePolicy()
		}
		if w.Replicas != nil {
			webhooks["replicaCount"] = int(w.GetReplicas())
		}
		if len(webhooks) > 0 {
			values["webhooks"] = webhooks
		}
		if r := resourcesMap(w.GetResources()); r != nil {
			resources["webhooks"] = r
		}
	}

	if len(resources) > 0 {
		values["resources"] = resources
	}

	// ---- pod identity ---------------------------------------------------------
	// The arms configure independent chart blocks — enabled cross-cloud
	// combinations render side by side.
	if pi := spec.GetPodIdentity(); pi != nil {
		podIdentity := map[string]interface{}{}
		if irsa := pi.GetAwsIrsa(); irsa != nil && irsa.GetEnabled() {
			podIdentity["aws"] = map[string]interface{}{
				"irsa": map[string]interface{}{
					"enabled": true,
					"roleArn": irsa.GetRoleArn(),
				},
			}
		}
		if wi := pi.GetAzureWorkloadIdentity(); wi != nil && wi.GetEnabled() {
			podIdentity["azureWorkload"] = map[string]interface{}{
				"enabled":  true,
				"clientId": wi.GetClientId(),
				"tenantId": wi.GetTenantId(),
			}
		}
		if wi := pi.GetGcpWorkloadIdentity(); wi != nil && wi.GetEnabled() {
			podIdentity["gcp"] = map[string]interface{}{
				"enabled":              true,
				"gcpIAMServiceAccount": wi.GetServiceAccountEmail(),
			}
		}
		if len(podIdentity) > 0 {
			values["podIdentity"] = podIdentity
		}
	}

	// ---- internal TLS certificates ------------------------------------------------
	if c := certificatesMap(spec.GetCertificates()); c != nil {
		values["certificates"] = c
	}

	// ---- scaler HTTP timeout ---------------------------------------------------------
	if spec.HttpTimeoutMs != nil {
		values["http"] = map[string]interface{}{
			"timeout": int(spec.GetHttpTimeoutMs()),
		}
	}

	// ---- scheduling -----------------------------------------------------------------
	if spec.GetPriorityClassName() != "" {
		values["priorityClassName"] = spec.GetPriorityClassName()
	}
	if len(spec.GetNodeSelector()) > 0 {
		values["nodeSelector"] = stringMapToInterface(spec.GetNodeSelector())
	}
	if len(spec.GetTolerations()) > 0 {
		values["tolerations"] = tolerationsSlice(spec.GetTolerations())
	}

	// ---- own telemetry ------------------------------------------------------------------
	// KEDA exposes its own /metrics per component — the chart mirrors the
	// per-component layout, so one spec flag fans out to the operator and
	// metrics-server blocks identically.
	if p := spec.GetPrometheus(); p.GetEnabled() {
		operatorTelemetry := map[string]interface{}{"enabled": true}
		metricServerTelemetry := map[string]interface{}{"enabled": true}
		if p.GetServiceMonitor() {
			operatorTelemetry["serviceMonitor"] = map[string]interface{}{"enabled": true}
			metricServerTelemetry["serviceMonitor"] = map[string]interface{}{"enabled": true}
		}
		values["prometheus"] = map[string]interface{}{
			"operator":     operatorTelemetry,
			"metricServer": metricServerTelemetry,
		}
	}

	// ---- escape hatch (merged LAST, helm -f semantics) ------------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	return values, nil
}

// certificatesMap renders the internal-TLS configuration. Returns nil for
// the default ("operator" — the KEDA operator self-generates certificates
// and patches the APIService caBundle): the chart's own default needs no
// values.
func certificatesMap(c *kuberneteskedav1alpha1.KubernetesKedaCertificates) map[string]interface{} {
	if c == nil || c.GetType() != "cert_manager" {
		return nil
	}
	certManager := map[string]interface{}{"enabled": true}
	// With no issuer reference the chart generates its own self-signed
	// CA + Issuer chain — leave the issuer block absent.
	if issuer := c.GetCertManagerIssuer(); issuer != nil {
		kind, name := issuer.CRDKind()
		certManager["issuer"] = map[string]interface{}{
			"generate": false,
			"name":     name,
			"kind":     kind,
			"group":    "cert-manager.io",
		}
	}
	return map[string]interface{}{"certManager": certManager}
}
