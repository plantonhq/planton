package module

import (
	kubernetestektonv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestekton/v1alpha1"
)

// tektonConfigSpecBody renders the TektonConfig CR spec from the typed
// fields. Field names are the operator API's own JSON keys (verified
// against the pinned operator source, pkg/apis/operator/v1alpha1) —
// values render ONLY when declared, so the operator's defaulting stays
// authoritative for everything the manifest leaves unsaid. The Terraform
// twin (locals.tf `tekton_config_spec`) renders the identical body —
// keep them in byte lockstep.
func tektonConfigSpecBody(locals *Locals) map[string]interface{} {
	spec := locals.Spec

	body := map[string]interface{}{
		"profile":         locals.Profile,
		"targetNamespace": locals.TargetNamespace,
	}

	if metadata := namespaceMetadataBody(spec.GetTargetNamespaceMetadata()); metadata != nil {
		body["targetNamespaceMetadata"] = metadata
	}
	if config := placementBody(spec.GetPlacement()); config != nil {
		body["config"] = config
	}
	if pipeline := pipelineBody(spec.GetPipeline()); pipeline != nil {
		body["pipeline"] = pipeline
	}
	if trigger := triggerBody(spec.GetTrigger()); trigger != nil {
		body["trigger"] = trigger
	}
	if dashboard := dashboardBody(spec.GetDashboard()); dashboard != nil {
		body["dashboard"] = dashboard
	}
	if chain := chainBody(spec.GetChain()); chain != nil {
		body["chain"] = chain
	}
	if pruner := prunerBody(spec.GetPruner()); pruner != nil {
		body["pruner"] = pruner
	}
	if params := paramsBody(spec.GetAdditionalParams()); params != nil {
		body["params"] = params
	}

	return body
}

func namespaceMetadataBody(metadata *kubernetestektonv1alpha1.KubernetesTektonNamespaceMetadata) map[string]interface{} {
	if metadata == nil {
		return nil
	}
	out := map[string]interface{}{}
	if len(metadata.GetLabels()) > 0 {
		out["labels"] = stringMapToInterface(metadata.GetLabels())
	}
	if len(metadata.GetAnnotations()) > 0 {
		out["annotations"] = stringMapToInterface(metadata.GetAnnotations())
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// placementBody renders spec.config — scheduling applied to every Tekton
// component pod the operator deploys.
func placementBody(placement *kubernetestektonv1alpha1.KubernetesTektonPlacement) map[string]interface{} {
	if placement == nil {
		return nil
	}
	out := map[string]interface{}{}
	if len(placement.GetNodeSelector()) > 0 {
		out["nodeSelector"] = stringMapToInterface(placement.GetNodeSelector())
	}
	if len(placement.GetTolerations()) > 0 {
		tolerations := []interface{}{}
		for _, toleration := range placement.GetTolerations() {
			entry := map[string]interface{}{}
			if toleration.GetKey() != "" {
				entry["key"] = toleration.GetKey()
			}
			if toleration.GetOperator() != "" {
				entry["operator"] = toleration.GetOperator()
			}
			if toleration.GetValue() != "" {
				entry["value"] = toleration.GetValue()
			}
			if toleration.GetEffect() != "" {
				entry["effect"] = toleration.GetEffect()
			}
			if toleration.TolerationSeconds != nil {
				entry["tolerationSeconds"] = int(toleration.GetTolerationSeconds())
			}
			tolerations = append(tolerations, entry)
		}
		out["tolerations"] = tolerations
	}
	if placement.GetPriorityClassName() != "" {
		out["priorityClassName"] = placement.GetPriorityClassName()
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// pipelineBody renders spec.pipeline — the feature flags, defaults,
// CloudEvents sink (as the config-events ConfigMap's data), metrics shape,
// resolver toggles and performance block. Tri-state booleans (optional in
// the proto) render only when set, keeping Tekton's own defaults
// authoritative.
func pipelineBody(pipeline *kubernetestektonv1alpha1.KubernetesTektonPipeline) map[string]interface{} {
	if pipeline == nil {
		return nil
	}
	out := map[string]interface{}{}

	// The sink lives in the config-events ConfigMap, whose `sink`
	// supersedes config-defaults' deprecated default-cloud-events-sink.
	// The TektonConfig reaches that ConfigMap only through
	// options.configMaps, which the operator merges into the one it
	// installs; tektonv1 is the one event format Tekton emits.
	if pipeline.GetCloudEventsSinkUrl() != "" {
		out["options"] = map[string]interface{}{
			"configMaps": map[string]interface{}{
				"config-events": map[string]interface{}{
					"data": map[string]interface{}{
						"sink":    pipeline.GetCloudEventsSinkUrl(),
						"formats": "tektonv1",
					},
				},
			},
		}
	}
	if pipeline.GetEnableApiFields() != "" {
		out["enable-api-fields"] = pipeline.GetEnableApiFields()
	}
	if pipeline.DefaultTimeoutMinutes != nil {
		out["default-timeout-minutes"] = int(pipeline.GetDefaultTimeoutMinutes())
	}
	if pipeline.GetDefaultServiceAccount() != "" {
		out["default-service-account"] = pipeline.GetDefaultServiceAccount()
	}

	if features := pipeline.GetFeatures(); features != nil {
		setBool(out, "disable-creds-init", features.DisableCredsInit)
		setBool(out, "await-sidecar-readiness", features.AwaitSidecarReadiness)
		setBool(out, "running-in-environment-with-injected-sidecars", features.RunningInEnvironmentWithInjectedSidecars)
		setBool(out, "require-git-ssh-secret-known-hosts", features.RequireGitSshSecretKnownHosts)
		setBool(out, "enable-custom-tasks", features.EnableCustomTasks)
		setBool(out, "keep-pod-on-cancel", features.KeepPodOnCancel)
		setBool(out, "enable-provenance-in-status", features.EnableProvenanceInStatus)
		setBool(out, "set-security-context", features.SetSecurityContext)
		setBool(out, "enable-cel-in-whenexpression", features.EnableCelInWhenexpression)
		setBool(out, "enable-step-actions", features.EnableStepActions)
		setBool(out, "enable-param-enum", features.EnableParamEnum)
		if features.GetResultsFrom() != "" {
			out["results-from"] = features.GetResultsFrom()
		}
		if features.MaxResultSize != nil {
			out["max-result-size"] = int(features.GetMaxResultSize())
		}
		if features.GetCoschedule() != "" {
			out["coschedule"] = features.GetCoschedule()
		}
	}

	if resolvers := pipeline.GetResolvers(); resolvers != nil {
		setBool(out, "enable-bundles-resolver", resolvers.EnableBundlesResolver)
		setBool(out, "enable-hub-resolver", resolvers.EnableHubResolver)
		setBool(out, "enable-git-resolver", resolvers.EnableGitResolver)
		setBool(out, "enable-cluster-resolver", resolvers.EnableClusterResolver)
	}

	if metrics := pipeline.GetMetrics(); metrics != nil {
		if metrics.GetTaskrunLevel() != "" {
			out["metrics.taskrun.level"] = metrics.GetTaskrunLevel()
		}
		if metrics.GetTaskrunDurationType() != "" {
			out["metrics.taskrun.duration-type"] = metrics.GetTaskrunDurationType()
		}
		if metrics.GetPipelinerunLevel() != "" {
			out["metrics.pipelinerun.level"] = metrics.GetPipelinerunLevel()
		}
		if metrics.GetPipelinerunDurationType() != "" {
			out["metrics.pipelinerun.duration-type"] = metrics.GetPipelinerunDurationType()
		}
		setBool(out, "metrics.count.enable-reason", metrics.CountWithReason)
	}

	if performance := performanceBody(pipeline.GetPerformance()); performance != nil {
		out["performance"] = performance
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func performanceBody(performance *kubernetestektonv1alpha1.KubernetesTektonPipelinePerformance) map[string]interface{} {
	if performance == nil {
		return nil
	}
	out := map[string]interface{}{}
	if performance.Replicas != nil {
		out["replicas"] = int(performance.GetReplicas())
	}
	if performance.Buckets != nil {
		out["buckets"] = int(performance.GetBuckets())
	}
	if performance.ThreadsPerController != nil {
		out["threads-per-controller"] = int(performance.GetThreadsPerController())
	}
	if performance.KubeApiQps != nil {
		out["kube-api-qps"] = int(performance.GetKubeApiQps())
	}
	if performance.KubeApiBurst != nil {
		out["kube-api-burst"] = int(performance.GetKubeApiBurst())
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func triggerBody(trigger *kubernetestektonv1alpha1.KubernetesTektonTrigger) map[string]interface{} {
	if trigger == nil {
		return nil
	}
	out := map[string]interface{}{}
	if trigger.GetEnableApiFields() != "" {
		out["enable-api-fields"] = trigger.GetEnableApiFields()
	}
	if trigger.GetDefaultServiceAccount() != "" {
		out["default-service-account"] = trigger.GetDefaultServiceAccount()
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func dashboardBody(dashboard *kubernetestektonv1alpha1.KubernetesTektonDashboard) map[string]interface{} {
	if dashboard == nil {
		return nil
	}
	out := map[string]interface{}{
		// `readonly` is a required (non-pointer) upstream field — render
		// it whenever the block is declared.
		"readonly": dashboard.GetReadonly(),
	}
	if dashboard.GetExternalLogs() != "" {
		out["external-logs"] = dashboard.GetExternalLogs()
	}
	return out
}

func chainBody(chain *kubernetestektonv1alpha1.KubernetesTektonChain) map[string]interface{} {
	if chain == nil {
		return nil
	}
	out := map[string]interface{}{
		"disabled": chain.GetDisabled(),
	}
	if chain.GetGenerateSigningSecret() {
		out["generateSigningSecret"] = true
	}
	return out
}

// prunerBody renders spec.pruner. An absent block renders nothing — no
// pruner cron is scheduled (Tekton's own default).
func prunerBody(pruner *kubernetestektonv1alpha1.KubernetesTektonPruner) map[string]interface{} {
	if pruner == nil {
		return nil
	}
	out := map[string]interface{}{
		"disabled":  false,
		"schedule":  pruner.GetSchedule(),
		"resources": stringSliceToInterface(pruner.GetResources()),
	}
	if pruner.Keep != nil {
		out["keep"] = int(pruner.GetKeep())
	}
	if pruner.KeepSince != nil {
		out["keep-since"] = int(pruner.GetKeepSince())
	}
	if pruner.GetPrunePerResource() {
		out["prune-per-resource"] = true
	}
	return out
}

func paramsBody(params []*kubernetestektonv1alpha1.KubernetesTektonParam) []interface{} {
	if len(params) == 0 {
		return nil
	}
	out := []interface{}{}
	for _, param := range params {
		out = append(out, map[string]interface{}{
			"name":  param.GetName(),
			"value": param.GetValue(),
		})
	}
	return out
}

func setBool(out map[string]interface{}, key string, value *bool) {
	if value != nil {
		out[key] = *value
	}
}

func stringMapToInterface(in map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for key, value := range in {
		out[key] = value
	}
	return out
}

func stringSliceToInterface(in []string) []interface{} {
	out := []interface{}{}
	for _, value := range in {
		out = append(out, value)
	}
	return out
}
