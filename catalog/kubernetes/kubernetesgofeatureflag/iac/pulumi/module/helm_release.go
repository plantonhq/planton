package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sigs.k8s.io/yaml"
)

// helmRelease installs the relay from the official chart as a real Helm
// release (helm.v3 Release - Helm's own lifecycle and rollback; never the
// client-side Chart resource).
func helmRelease(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) (*helmv3.Release, error) {
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the relay to become Ready (Helm --wait), stated
		// explicitly to mirror the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependsOn...)

	release, err := helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to install go-feature-flag relay-proxy helm release")
	}
	return release, nil
}

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace), then re-pins fullnameOverride.
//
// PARITY: the Terraform module reaches the same result natively - its
// helm_release passes values = [yamlencode(typed values), helm_values,
// yamlencode(fullnameOverride re-pin)]. Keep every typed mapping below in
// lockstep with the Terraform module's locals.
//
// CHART FACTS (verified at chart 1.56.0) the mapping relies on:
//   - relayproxy.config is a string the chart writes into its ConfigMap
//     (see RelayConfig.Document for its exact shape);
//   - env is a MAP of variable name -> {value} or {valueFrom};
//   - the PodDisruptionBudget selects the label `name: <fullname>`, which
//     the chart never puts on pods - podLabels carries it here so the
//     budget protects the relay;
//   - the HorizontalPodAutoscaler renders a memory target unless
//     targetMemoryUtilizationPercentage is explicitly null.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec
	values := map[string]interface{}{
		"fullnameOverride": locals.ReleaseName,
		"relayproxy":       map[string]interface{}{"config": locals.Relay.Document},
		"service": map[string]interface{}{
			"port": locals.Port,
		},
	}

	if st := spec.GetServer().GetServiceType(); st != "" {
		values["service"].(map[string]interface{})["type"] = st
	}

	// ---- image -------------------------------------------------------------
	image := map[string]interface{}{}
	if img := spec.GetImage(); img != nil {
		setString(image, "repository", img.GetRepository())
		setString(image, "tag", img.GetTag())
		setString(image, "pullPolicy", img.GetPullPolicy())
		setTrue(image, "fips", img.GetFips())
		if len(img.GetPullSecretNames()) > 0 {
			secrets := []interface{}{}
			for _, n := range img.GetPullSecretNames() {
				secrets = append(secrets, map[string]interface{}{"name": n})
			}
			values["imagePullSecrets"] = secrets
		}
	}
	if len(image) > 0 {
		values["image"] = image
	}

	// ---- replicas and autoscaling -------------------------------------------
	if hpa := spec.GetHpa(); hpa.GetEnabled() {
		autoscaling := map[string]interface{}{"enabled": true}
		if hpa.MinReplicas != nil {
			autoscaling["minReplicas"] = int(hpa.GetMinReplicas())
		}
		if hpa.MaxReplicas != nil {
			autoscaling["maxReplicas"] = int(hpa.GetMaxReplicas())
		}
		if hpa.TargetCpuUtilizationPercent != nil {
			autoscaling["targetCPUUtilizationPercentage"] = int(hpa.GetTargetCpuUtilizationPercent())
		}
		if hpa.TargetMemoryUtilizationPercent != nil {
			autoscaling["targetMemoryUtilizationPercentage"] = int(hpa.GetTargetMemoryUtilizationPercent())
		} else {
			autoscaling["targetMemoryUtilizationPercentage"] = nil
		}
		values["autoscaling"] = autoscaling
	} else if spec.Replicas != nil {
		values["replicaCount"] = int(spec.GetReplicas())
	}

	// ---- disruption budget -----------------------------------------------------
	if pdb := spec.GetPdb(); pdb.GetEnabled() {
		budget := map[string]interface{}{"enable": true}
		if pdb.GetMaxUnavailable() != "" {
			budget["maxUnavailable"] = intOrString(pdb.GetMaxUnavailable())
		} else if pdb.GetMinAvailable() != "" {
			budget["minAvailable"] = intOrString(pdb.GetMinAvailable())
		}
		values["pdb"] = budget
	}

	// ---- pods ---------------------------------------------------------------------
	podLabels := map[string]interface{}{}
	for k, v := range spec.GetPodLabels() {
		podLabels[k] = v
	}
	for k, v := range locals.Labels {
		podLabels[k] = v
	}
	podLabels["name"] = locals.ReleaseName
	values["podLabels"] = podLabels

	// The env Secret's checksum rolls the pods when a secret value changes:
	// env is read once at process start, so a rotated key would otherwise
	// reach the relay only on the next restart.
	podAnnotations := map[string]interface{}{}
	for k, v := range spec.GetPodAnnotations() {
		podAnnotations[k] = v
	}
	if len(locals.Relay.SecretEnv) > 0 {
		sum, err := secretEnvChecksum(locals.Relay.SecretEnv)
		if err != nil {
			return nil, err
		}
		podAnnotations[vars.EnvSecretChecksumAnnotation] = sum
	}
	if len(podAnnotations) > 0 {
		values["podAnnotations"] = podAnnotations
	}
	if len(spec.GetCommonLabels()) > 0 {
		values["commonLabels"] = stringMap(spec.GetCommonLabels())
	}
	if psc := podSecurityContextBody(spec.GetPodSecurityContext()); len(psc) > 0 {
		values["podSecurityContext"] = psc
	}
	if csc := containerSecurityContextBody(spec.GetContainerSecurityContext()); len(csc) > 0 {
		values["securityContext"] = csc
	}
	if resources := resourcesBody(spec.GetResources()); resources != nil {
		values["resources"] = resources
	}

	// ---- environment -----------------------------------------------------------------
	// Secret values by secretKeyRef into the module-owned Secret; extra
	// variables as given; the OpenTelemetry protocol and sampler variables,
	// which the relay's exporter and sampler read from the environment only.
	env := map[string]interface{}{}
	for name, value := range spec.GetExtraEnv() {
		env[name] = map[string]interface{}{"value": value}
	}
	for name, ref := range spec.GetExtraEnvFromSecret() {
		env[name] = map[string]interface{}{"valueFrom": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{"name": ref.GetName(), "key": ref.GetKey()},
		}}
	}
	if p := spec.GetTelemetry().GetOtlpProtocol(); p != "" {
		env["OTEL_EXPORTER_OTLP_PROTOCOL"] = map[string]interface{}{"value": p}
	}
	// Every sampler but jaeger_remote is built by the OpenTelemetry SDK from
	// its own environment.
	if sampler := spec.GetTelemetry().GetTracesSampler(); sampler != "" && sampler != "jaeger_remote" {
		env["OTEL_TRACES_SAMPLER"] = map[string]interface{}{"value": sampler}
	}
	if arg := spec.GetTelemetry().GetTracesSamplerArg(); arg != "" {
		env["OTEL_TRACES_SAMPLER_ARG"] = map[string]interface{}{"value": arg}
	}
	secretNames := make([]string, 0, len(locals.Relay.SecretEnv))
	for name := range locals.Relay.SecretEnv {
		secretNames = append(secretNames, name)
	}
	sort.Strings(secretNames)
	for _, name := range secretNames {
		if _, clash := env[name]; clash {
			return nil, errors.Errorf("extra environment variable %q collides with a variable the module generates for a secret value", name)
		}
		env[name] = map[string]interface{}{"valueFrom": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{"name": locals.EnvSecretName, "key": name},
		}}
	}
	if len(env) > 0 {
		values["env"] = env
	}

	// ---- scheduling ------------------------------------------------------------------------
	if s := spec.GetScheduling(); s != nil {
		if len(s.GetNodeSelector()) > 0 {
			values["nodeSelector"] = stringMap(s.GetNodeSelector())
		}
		if len(s.GetTolerations()) > 0 {
			values["tolerations"] = tolerationsBody(s.GetTolerations())
		}
		if affinity := affinityBody(s.GetNodeAffinity(), s.GetPodAffinity(), s.GetPodAntiAffinity()); affinity != nil {
			values["affinity"] = affinity
		}
	}

	// ---- service account -----------------------------------------------------------------------
	if existing := spec.GetServiceAccount().GetExistingName(); existing != "" {
		values["serviceAccount"] = map[string]interface{}{"create": false, "name": existing}
	} else if len(spec.GetServiceAccount().GetAnnotations()) > 0 {
		values["serviceAccount"] = map[string]interface{}{
			"annotations": stringMap(spec.GetServiceAccount().GetAnnotations()),
		}
	}

	// ---- escape hatch (merged LAST, Helm -f semantics) -----------------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as YAML")
		}
		values = mergeMaps(values, overrides)
	}

	// fullnameOverride re-pinned AFTER the merge - the one deliberate
	// exception to the escape hatch's last-word contract: the Service,
	// ServiceAccount and every output derive from the fullname.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// intOrString renders an integer-or-percentage string as the YAML scalar
// Kubernetes expects (a bare integer stays a number).
func intOrString(s string) interface{} {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return s
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// mergeMaps deep-merges override into base with Helm `-f` semantics: maps
// merge recursively, every other value (lists included) replaces.
func mergeMaps(base, override map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(base))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		if bv, ok := out[k].(map[string]interface{}); ok {
			if ov, ok := v.(map[string]interface{}); ok {
				out[k] = mergeMaps(bv, ov)
				continue
			}
		}
		out[k] = v
	}
	return out
}

// secretEnvChecksum is the sha256 of the env Secret's data as compact JSON
// with sorted keys (the Terraform twin hashes jsonencode of the same map).
func secretEnvChecksum(secretEnv map[string]string) (string, error) {
	b, err := json.Marshal(secretEnv)
	if err != nil {
		return "", errors.Wrap(err, "failed to serialize the env secret for its checksum")
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
