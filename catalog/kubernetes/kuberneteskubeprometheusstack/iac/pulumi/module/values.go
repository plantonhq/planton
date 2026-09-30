package module

import (
	"fmt"

	"github.com/pkg/errors"
	kuberneteskubeprometheusstackv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskubeprometheusstack/v1alpha1"
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
	values := map[string]interface{}{}

	// fullnameOverride pins kube-prometheus-stack.fullname to the resource
	// name: every child name (`-prometheus`, `-alertmanager`, `-grafana`,
	// `-operator`, the exporter subcharts) derives deterministically, and
	// the exported outputs are built from that contract. The 26-character
	// budget is fail-loud-checked in Resources() — the chart would
	// otherwise TRUNCATE silently.
	values["fullnameOverride"] = locals.ReleaseName

	// ---- CRD lifecycle ------------------------------------------------------
	// The CRDs ship in the chart's local `crds` subchart whose crds/
	// directory Helm installs ONCE: upgrades never touch them, uninstall
	// keeps them (ServiceMonitors and rules across the cluster survive
	// removal of the stack). skip_crds is the bring-your-own-CRDs arm; the
	// upgradeJob is the chart's own pre-install/pre-upgrade hook that
	// server-side-applies the bundle across operator versions.
	if spec.GetSkipCrds() {
		values["crds"] = map[string]interface{}{"enabled": false}
	} else if spec.GetCrdUpgradeJob() {
		values["crds"] = map[string]interface{}{
			"upgradeJob": map[string]interface{}{"enabled": true},
		}
	}

	// ---- global image plumbing ------------------------------------------------
	// imageRegistry replaces the registry of EVERY image the stack pulls
	// (the air-gap path); imagePullSecrets must reach every workload —
	// operator, prometheus, alertmanager, exporters, grafana, the certgen
	// hook — which is exactly what the chart's global block does. The
	// Terraform twin renders the same shape (a silent single-engine
	// ImagePullBackOff class otherwise).
	global := map[string]interface{}{}
	if spec.GetImageRegistry() != "" {
		global["imageRegistry"] = spec.GetImageRegistry()
	}
	if len(spec.GetImagePullSecrets()) > 0 {
		pullSecrets := make([]interface{}, 0, len(spec.GetImagePullSecrets()))
		for _, s := range spec.GetImagePullSecrets() {
			pullSecrets = append(pullSecrets, map[string]interface{}{"name": s})
		}
		global["imagePullSecrets"] = pullSecrets
	}
	if len(global) > 0 {
		values["global"] = global
	}

	// ---- prometheus -------------------------------------------------------------
	prometheusSpec, err := buildPrometheusSpec(locals)
	if err != nil {
		return nil, err
	}
	values["prometheus"] = map[string]interface{}{
		"prometheusSpec": prometheusSpec,
	}

	// ---- alertmanager --------------------------------------------------------------
	alertmanager, err := buildAlertmanagerValues(locals)
	if err != nil {
		return nil, err
	}
	values["alertmanager"] = alertmanager

	// ---- bundled grafana ---------------------------------------------------------------
	values["grafana"] = buildGrafanaValues(locals)

	// ---- operator -------------------------------------------------------------------------
	// Pruned when empty so the rendered values document matches the
	// Terraform twin key-for-key (Helm treats an empty map as a no-op
	// either way).
	if operator := buildOperatorValues(locals); len(operator) > 0 {
		values["prometheusOperator"] = operator
	}

	// ---- exporters ---------------------------------------------------------------------------
	// The toggle keys (kubeStateMetrics/nodeExporter) gate the subcharts;
	// the subchart config keys (kube-state-metrics/prometheus-node-exporter)
	// carry their resources.
	exporters := spec.GetExporters()
	ksmEnabled := true
	neEnabled := true
	if exporters != nil {
		if exporters.KubeStateMetricsEnabled != nil {
			ksmEnabled = exporters.GetKubeStateMetricsEnabled()
		}
		if exporters.NodeExporterEnabled != nil {
			neEnabled = exporters.GetNodeExporterEnabled()
		}
	}
	values["kubeStateMetrics"] = map[string]interface{}{"enabled": ksmEnabled}
	values["nodeExporter"] = map[string]interface{}{"enabled": neEnabled}
	if r := resourcesValue(exporters.GetKubeStateMetricsResources()); r != nil {
		values["kube-state-metrics"] = map[string]interface{}{"resources": r}
	}
	if r := resourcesValue(exporters.GetNodeExporterResources()); r != nil {
		values["prometheus-node-exporter"] = map[string]interface{}{"resources": r}
	}

	// ---- control-plane scrapers ------------------------------------------------------------------
	// All chart-default-on; rendered only when the manifest disables one
	// (the managed-cloud posture — see the spec's MANAGED CLOUDS note and
	// the per-cloud presets).
	if cps := spec.GetControlPlaneScrapers(); cps != nil {
		renderScraperToggle(values, "kubeApiServer", cps.KubeApiServer)
		renderScraperToggle(values, "kubelet", cps.Kubelet)
		renderScraperToggle(values, "kubeControllerManager", cps.KubeControllerManager)
		renderScraperToggle(values, "coreDns", cps.CoreDns)
		renderScraperToggle(values, "kubeEtcd", cps.KubeEtcd)
		renderScraperToggle(values, "kubeScheduler", cps.KubeScheduler)
		renderScraperToggle(values, "kubeProxy", cps.KubeProxy)
	}

	// ---- default rules --------------------------------------------------------------------------------
	if dr := spec.GetDefaultRules(); dr != nil {
		defaultRules := map[string]interface{}{}
		if dr.Enabled != nil && !dr.GetEnabled() {
			defaultRules["create"] = false
		}
		if len(dr.GetDisabledGroups()) > 0 {
			rules := map[string]interface{}{}
			for _, group := range dr.GetDisabledGroups() {
				rules[group] = false
			}
			defaultRules["rules"] = rules
		}
		if len(defaultRules) > 0 {
			values["defaultRules"] = defaultRules
		}
	}

	// ---- escape hatch (merged LAST, helm -f semantics) --------------------------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	// fullnameOverride re-pinned AFTER the merge — the one deliberate
	// exception to the escape hatch's last-word contract (twin of the
	// Terraform module's third values document). Every child service name
	// — and the exported outputs built from them — derives from the
	// fullname; letting an override move it would break every output.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// buildPrometheusSpec renders the spec's prometheus block into the chart's
// prometheus.prometheusSpec value (the Prometheus CR the operator
// reconciles).
func buildPrometheusSpec(locals *Locals) (map[string]interface{}, error) {
	prom := locals.Spec.GetPrometheus()
	prometheusSpec := map[string]interface{}{}

	replicas := int32(1)
	retention := ""
	if prom != nil {
		if prom.Replicas != nil {
			replicas = prom.GetReplicas()
		}
		retention = prom.GetRetention()
	}
	prometheusSpec["replicas"] = int(replicas)
	if retention != "" {
		prometheusSpec["retention"] = retention
	}
	if prom.GetRetentionSize() != "" {
		prometheusSpec["retentionSize"] = prom.GetRetentionSize()
	}

	// Storage: a PVC per replica BY DEFAULT — the chart's own default is
	// an emptyDir that loses every metric on restart, which the spec
	// deliberately inverts. The ephemeral arm restores the chart default
	// for throwaway clusters.
	if !prom.GetEphemeral() {
		diskSize := prom.GetDiskSize()
		if diskSize == "" {
			diskSize = vars.DefaultPrometheusDiskSize
		}
		prometheusSpec["storageSpec"] = volumeClaimTemplate(diskSize, prom.GetStorageClass().GetValue())
	}

	if r := resourcesValue(prom.GetResources()); r != nil {
		prometheusSpec["resources"] = r
	}
	if len(prom.GetExternalLabels()) > 0 {
		prometheusSpec["externalLabels"] = stringMapToInterface(prom.GetExternalLabels())
	}
	if prom.GetScrapeInterval() != "" {
		prometheusSpec["scrapeInterval"] = prom.GetScrapeInterval()
	}
	if prom.GetEvaluationInterval() != "" {
		prometheusSpec["evaluationInterval"] = prom.GetEvaluationInterval()
	}
	if prom.GetEnableRemoteWriteReceiver() {
		prometheusSpec["enableRemoteWriteReceiver"] = true
	}

	// Discovery: the component default is cluster-wide (`all_monitors`) —
	// every catalog kind's service_monitor toggle and any user-authored
	// monitor lights up without extra wiring. The chart's own default is
	// release-fenced; `release_managed_only` restores it by rendering
	// nothing. The five NilUsesHelmValues switches are the chart's
	// documented mechanism for "select everything".
	discovery := kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackMonitorDiscovery_all_monitors
	if prom != nil && prom.Discovery != nil {
		discovery = prom.GetDiscovery()
	}
	if discovery == kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackMonitorDiscovery_all_monitors ||
		discovery == kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackMonitorDiscovery_kubernetes_kube_prometheus_stack_monitor_discovery_unspecified {
		prometheusSpec["serviceMonitorSelectorNilUsesHelmValues"] = false
		prometheusSpec["podMonitorSelectorNilUsesHelmValues"] = false
		prometheusSpec["ruleSelectorNilUsesHelmValues"] = false
		prometheusSpec["probeSelectorNilUsesHelmValues"] = false
		prometheusSpec["scrapeConfigSelectorNilUsesHelmValues"] = false
	}

	// Remote write: each entry maps onto the Prometheus CRD's remoteWrite
	// shape. The CRD reads BOTH basic-auth halves from Secrets; the
	// declared plain-string username rides the module-owned
	// `<name>-remote-write-auth` Secret (key `username-<i>`), the password
	// stays in the user's Secret — mixed-source selectors are the CRD's
	// normal shape.
	if len(prom.GetRemoteWrite()) > 0 {
		remoteWrites := make([]interface{}, 0, len(prom.GetRemoteWrite()))
		for i, rw := range prom.GetRemoteWrite() {
			entry := map[string]interface{}{"url": rw.GetUrl()}
			if rw.GetName() != "" {
				entry["name"] = rw.GetName()
			}
			switch auth := rw.GetAuth().(type) {
			case *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackRemoteWrite_BasicAuth:
				entry["basicAuth"] = map[string]interface{}{
					"username": map[string]interface{}{
						"name": locals.RemoteWriteAuthSecretName,
						"key":  remoteWriteUsernameKey(i),
					},
					"password": map[string]interface{}{
						"name": auth.BasicAuth.GetPasswordSecret().GetName(),
						"key":  auth.BasicAuth.GetPasswordSecret().GetKey(),
					},
				}
			case *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackRemoteWrite_BearerTokenSecret:
				entry["authorization"] = map[string]interface{}{
					"type": "Bearer",
					"credentials": map[string]interface{}{
						"name": auth.BearerTokenSecret.GetName(),
						"key":  auth.BearerTokenSecret.GetKey(),
					},
				}
			case *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackRemoteWrite_Sigv4:
				sigv4 := map[string]interface{}{"region": auth.Sigv4.GetRegion()}
				if auth.Sigv4.GetRoleArn() != "" {
					sigv4["roleArn"] = auth.Sigv4.GetRoleArn()
				}
				if ak := auth.Sigv4.GetAccessKeySecret(); ak != nil {
					sigv4["accessKey"] = map[string]interface{}{
						"name": ak.GetName(),
						"key":  ak.GetKey(),
					}
				}
				if sk := auth.Sigv4.GetSecretKeySecret(); sk != nil {
					sigv4["secretKey"] = map[string]interface{}{
						"name": sk.GetName(),
						"key":  sk.GetKey(),
					}
				}
				entry["sigv4"] = sigv4
			case *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackRemoteWrite_AzureAd:
				azureAd := map[string]interface{}{
					"managedIdentity": map[string]interface{}{
						"clientId": auth.AzureAd.GetManagedIdentityClientId(),
					},
				}
				if auth.AzureAd.GetCloud() != "" {
					azureAd["cloud"] = auth.AzureAd.GetCloud()
				}
				entry["azureAd"] = azureAd
			}
			remoteWrites = append(remoteWrites, entry)
		}
		prometheusSpec["remoteWrite"] = remoteWrites
	}

	// Raw scrape_config seam: the chart accepts an inline LIST of
	// scrape_config entries. Entries here bypass the operator's
	// validation — the spec comment carries the warning.
	if prom.GetAdditionalScrapeConfigs() != "" {
		var scrapeConfigs []interface{}
		if err := yaml.Unmarshal([]byte(prom.GetAdditionalScrapeConfigs()), &scrapeConfigs); err != nil {
			return nil, errors.Wrap(err, "failed to parse additional_scrape_configs as a YAML list of scrape_config entries")
		}
		prometheusSpec["additionalScrapeConfigs"] = scrapeConfigs
	}

	applyScheduling(prometheusSpec, prom.GetScheduling())
	return prometheusSpec, nil
}

// buildAlertmanagerValues renders the spec's alertmanager block into the
// chart's alertmanager value.
func buildAlertmanagerValues(locals *Locals) (map[string]interface{}, error) {
	am := locals.Spec.GetAlertmanager()
	alertmanager := map[string]interface{}{"enabled": locals.AlertmanagerEnabled}
	if !locals.AlertmanagerEnabled {
		return alertmanager, nil
	}

	alertmanagerSpec := map[string]interface{}{}
	replicas := int32(1)
	retention := ""
	if am != nil {
		if am.Replicas != nil {
			replicas = am.GetReplicas()
		}
		retention = am.GetRetention()
	}
	alertmanagerSpec["replicas"] = int(replicas)
	if retention != "" {
		alertmanagerSpec["retention"] = retention
	}

	// Storage: a small PVC per replica BY DEFAULT (silences and the
	// notification log survive restarts); the ephemeral arm restores the
	// chart's emptyDir default.
	if !am.GetEphemeral() {
		diskSize := am.GetDiskSize()
		if diskSize == "" {
			diskSize = vars.DefaultAlertmanagerDiskSize
		}
		alertmanagerSpec["storage"] = volumeClaimTemplate(diskSize, am.GetStorageClass().GetValue())
	}

	if r := resourcesValue(am.GetResources()); r != nil {
		alertmanagerSpec["resources"] = r
	}
	applyScheduling(alertmanagerSpec, am.GetScheduling())
	alertmanager["alertmanagerSpec"] = alertmanagerSpec

	// Typed notifications (notifications.go): the routing tree and
	// receivers, the message template, and the credentials Secret mounted
	// for the receivers' `_file` fields. Mutually exclusive with
	// config_yaml (a spec rule).
	if n := locals.Notifications; n != nil {
		alertmanager["config"] = n.Config
		alertmanager["templateFiles"] = n.TemplateFiles
		if len(n.SecretData) > 0 {
			alertmanagerSpec["secrets"] = []interface{}{locals.NotificationsSecretName}
		}
	}

	// The raw alerting configuration document (route/receivers). The chart
	// value is a MAP (rendered into the Alertmanager Secret), so the
	// spec's YAML seam parses here; empty = the chart's null-receiver +
	// Watchdog default. Credential discipline lives on the spec field
	// comment (use _file receiver fields over a Secret the user mounts,
	// never inline tokens).
	if am.GetConfigYaml() != "" {
		config := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(am.GetConfigYaml()), &config); err != nil {
			return nil, errors.Wrap(err, "failed to parse alertmanager config_yaml as a YAML document")
		}
		alertmanager["config"] = config
	}

	return alertmanager, nil
}

// buildGrafanaValues renders the spec's bundled-grafana block into the
// grafana subchart's values.
func buildGrafanaValues(locals *Locals) map[string]interface{} {
	g := locals.Spec.GetGrafana()
	grafana := map[string]interface{}{"enabled": locals.GrafanaEnabled}
	if !locals.GrafanaEnabled {
		return grafana
	}

	if g != nil && g.DefaultDashboardsEnabled != nil && !g.GetDefaultDashboardsEnabled() {
		grafana["defaultDashboardsEnabled"] = false
	}

	// Existing-secret arm: the subchart wires the referenced Secret's
	// keys into env at pod start — it must exist BEFORE the install.
	// Generate arm (admin_secret absent): the subchart creates its own
	// `<name>-grafana` Secret ONCE (lookup-stable across upgrades).
	if existing := g.GetAdminSecret(); existing != nil {
		userKey := existing.GetUserKey()
		if userKey == "" {
			userKey = "admin-user"
		}
		passwordKey := existing.GetPasswordKey()
		if passwordKey == "" {
			passwordKey = "admin-password"
		}
		grafana["admin"] = map[string]interface{}{
			"existingSecret": existing.GetName(),
			"userKey":        userKey,
			"passwordKey":    passwordKey,
		}
	}

	// Rendered only when declared: the subchart default is ephemeral —
	// honest for dashboards-as-code, wrong for hand-authored dashboards
	// (the spec comment carries the trade).
	if storage := g.GetStorage(); storage != nil {
		size := storage.GetSize()
		if size == "" {
			size = vars.DefaultGrafanaStorageSize
		}
		persistence := map[string]interface{}{
			"enabled": true,
			"size":    size,
		}
		if sc := storage.GetStorageClass().GetValue(); sc != "" {
			persistence["storageClassName"] = sc
		}
		grafana["persistence"] = persistence
	}

	if r := resourcesValue(g.GetResources()); r != nil {
		grafana["resources"] = r
	}

	return grafana
}

// buildOperatorValues renders the spec's operator block into the chart's
// prometheusOperator value.
func buildOperatorValues(locals *Locals) map[string]interface{} {
	op := locals.Spec.GetOperator()
	operator := map[string]interface{}{}

	if r := resourcesValue(op.GetResources()); r != nil {
		operator["resources"] = r
	}

	// Admission webhooks: chart-default ON with the self-contained
	// certgen hook Job. The cert-manager arm swaps the certificate
	// machinery (an Issuer + Certificates the chart renders — requires
	// KubernetesCertManager); the disabled arm turns validation off
	// entirely (rules then fail at config-reload time, not admission).
	if aw := op.GetAdmissionWebhooks(); aw != nil {
		admissionWebhooks := map[string]interface{}{}
		if aw.GetDisabled() {
			admissionWebhooks["enabled"] = false
			// The certgen patch job has nothing to provision certificates
			// for when the webhooks are off.
			admissionWebhooks["patch"] = map[string]interface{}{"enabled": false}
		}
		if aw.GetCertManager() {
			admissionWebhooks["certManager"] = map[string]interface{}{"enabled": true}
		}
		if len(admissionWebhooks) > 0 {
			operator["admissionWebhooks"] = admissionWebhooks
		}
	}

	if sched := op.GetScheduling(); sched != nil {
		applyScheduling(operator, sched)
	}

	return operator
}

// volumeClaimTemplate renders the operator's storage shape (the
// volumeClaimTemplate both the Prometheus and Alertmanager CRs take).
func volumeClaimTemplate(size, storageClass string) map[string]interface{} {
	pvcSpec := map[string]interface{}{
		"accessModes": []interface{}{"ReadWriteOnce"},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{"storage": size},
		},
	}
	if storageClass != "" {
		pvcSpec["storageClassName"] = storageClass
	}
	return map[string]interface{}{
		"volumeClaimTemplate": map[string]interface{}{
			"spec": pvcSpec,
		},
	}
}

// applyScheduling folds the shared scheduling block into one component's
// spec map (nodeSelector/tolerations/priorityClassName are the same keys
// on prometheusSpec, alertmanagerSpec and the operator's top level).
func applyScheduling(target map[string]interface{}, sched *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackScheduling) {
	if sched == nil {
		return
	}
	if len(sched.GetNodeSelector()) > 0 {
		target["nodeSelector"] = stringMapToInterface(sched.GetNodeSelector())
	}
	if len(sched.GetTolerations()) > 0 {
		target["tolerations"] = tolerationsSlice(sched.GetTolerations())
	}
	if sched.GetPriorityClassName() != "" {
		target["priorityClassName"] = sched.GetPriorityClassName()
	}
}

// renderScraperToggle renders one control-plane scraper's enabled value —
// only when the manifest carries an explicit decision (proto optional
// presence), so chart defaults stay untouched otherwise.
func renderScraperToggle(values map[string]interface{}, chartKey string, toggle *bool) {
	if toggle == nil {
		return
	}
	values[chartKey] = map[string]interface{}{"enabled": *toggle}
}

// remoteWriteUsernameKey is the deterministic key of one remote-write
// entry's username inside the module-owned auth Secret. Both engines
// derive it identically.
func remoteWriteUsernameKey(index int) string {
	return fmt.Sprintf("username-%d", index)
}
