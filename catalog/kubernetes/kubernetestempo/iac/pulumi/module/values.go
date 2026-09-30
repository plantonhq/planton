package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kubernetestempov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestempo/v1alpha1"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics, and re-pins fullnameOverride last.
//
// SECRET DISCIPLINE (load-bearing): the chart renders the Tempo config into
// a ConfigMap. Declared object-store credentials therefore NEVER appear in
// these values: they travel as environment variables sourced from the
// referenced Secrets (tempo.extraEnv secretKeyRefs), the config references
// them as ${VAR} placeholders, and -config.expand-env=true (tempo.extraArgs)
// makes Tempo expand them at process start.
//
// PARITY: the Terraform module reaches the same result natively. Keep every
// typed mapping below in lockstep with the Terraform module's locals.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec

	if len(locals.ReleaseName) > vars.MaxNameLength {
		return nil, errors.Errorf(
			"metadata.name %q is %d characters — the tempo chart's child-name budget allows at most %d "+
				"(it composes names within Kubernetes' 63-character cap)",
			locals.ReleaseName, len(locals.ReleaseName), vars.MaxNameLength)
	}

	values := map[string]interface{}{
		"fullnameOverride": locals.ReleaseName,
	}

	// ---- replicas ------------------------------------------------------
	values["replicas"] = intOrDefault(spec.Replicas, 1)

	// ---- persistence ---------------------------------------------------
	// The chart's own default is emptyDir (traces vanish on restart); this
	// component provisions a PVC by default. `ephemeral` restores the
	// chart posture.
	if spec.GetEphemeral() {
		values["persistence"] = map[string]interface{}{"enabled": false}
	} else {
		diskSize := spec.GetDiskSize()
		if diskSize == "" {
			diskSize = vars.DefaultDiskSize
		}
		persistence := map[string]interface{}{
			"enabled": true,
			"size":    diskSize,
		}
		if sc := spec.GetStorageClass().GetValue(); sc != "" {
			persistence["storageClassName"] = sc
		}
		values["persistence"] = persistence
	}

	// ---- the tempo block -----------------------------------------------
	retention := spec.GetRetention()
	if retention == "" {
		retention = vars.DefaultRetention
	}
	tempo := map[string]interface{}{
		"retention": retention,
	}

	// Anonymous usage reporting: OFF unless explicitly opted in.
	if !boolOrDefault(spec.UsageReporting, false) {
		tempo["reportingEnabled"] = false
	}
	if spec.GetMultiTenancyEnabled() {
		tempo["multitenancyEnabled"] = true
	}

	// ---- receivers -----------------------------------------------------
	// OTLP is always on (the 2026 wire standard). The four legacy Jaeger
	// protocols are opt-in — every closed receiver is one less ingest
	// surface. This narrows the chart's all-receivers default.
	receivers := map[string]interface{}{
		"otlp": map[string]interface{}{
			"protocols": map[string]interface{}{
				"grpc": map[string]interface{}{"endpoint": "0.0.0.0:4317"},
				"http": map[string]interface{}{"endpoint": "0.0.0.0:4318"},
			},
		},
	}
	if boolOrDefault(spec.JaegerReceiversEnabled, false) {
		receivers["jaeger"] = map[string]interface{}{
			"protocols": map[string]interface{}{
				"grpc":           map[string]interface{}{"endpoint": "0.0.0.0:14250"},
				"thrift_binary":  map[string]interface{}{"endpoint": "0.0.0.0:6832"},
				"thrift_compact": map[string]interface{}{"endpoint": "0.0.0.0:6831"},
				"thrift_http":    map[string]interface{}{"endpoint": "0.0.0.0:14268"},
			},
		}
	}
	tempo["receivers"] = receivers

	// ---- storage backend -----------------------------------------------
	trace := map[string]interface{}{
		"wal": map[string]interface{}{"path": vars.WalPath},
	}
	credentialEnv := []interface{}{}
	extraVolumes := []interface{}{}
	extraVolumeMounts := []interface{}{}

	switch {
	case locals.R2 != nil:
		// The r2 arm: R2 speaks S3, so the trace backend is "s3" and the
		// values are composed from R2's vocabulary (r2.go). The key pair
		// rides env expansion from the module-owned Secret.
		trace["backend"] = "s3"
		trace["s3"] = locals.R2.S3Values
		credentialEnv = append(credentialEnv, r2CredentialEnv(locals.R2CredentialsSecretName)...)
	case spec.GetStorage().GetS3() != nil:
		s3 := spec.GetStorage().GetS3()
		trace["backend"] = "s3"
		s3Values := map[string]interface{}{
			"bucket":   s3.GetBucket(),
			"endpoint": s3.GetEndpoint(),
		}
		if s3.GetRegion() != "" {
			s3Values["region"] = s3.GetRegion()
		}
		if s3.GetForcePathStyle() {
			s3Values["forcepathstyle"] = true
		}
		if s3.GetInsecure() {
			s3Values["insecure"] = true
		}
		if creds := s3.GetCredentials(); creds != nil {
			s3Values["access_key"] = fmt.Sprintf("${%s}", vars.EnvS3AccessKeyId)
			s3Values["secret_key"] = fmt.Sprintf("${%s}", vars.EnvS3SecretAccessKey)
			credentialEnv = append(credentialEnv,
				secretEnvVar(vars.EnvS3AccessKeyId, creds.GetAccessKeyIdSecret()),
				secretEnvVar(vars.EnvS3SecretAccessKey, creds.GetSecretAccessKeySecret()),
			)
		}
		trace["s3"] = s3Values
	case spec.GetStorage().GetGcs() != nil:
		gcs := spec.GetStorage().GetGcs()
		trace["backend"] = "gcs"
		trace["gcs"] = map[string]interface{}{"bucket_name": gcs.GetBucket()}
		if key := gcs.GetServiceAccountKeySecret(); key != nil {
			keyPath := fmt.Sprintf("%s/%s", vars.GcsKeyMountPath, key.GetKey())
			credentialEnv = append(credentialEnv, map[string]interface{}{
				"name":  "GOOGLE_APPLICATION_CREDENTIALS",
				"value": keyPath,
			})
			extraVolumes = append(extraVolumes, map[string]interface{}{
				"name":   vars.GcsKeyVolume,
				"secret": map[string]interface{}{"secretName": key.GetName()},
			})
			extraVolumeMounts = append(extraVolumeMounts, map[string]interface{}{
				"name":      vars.GcsKeyVolume,
				"mountPath": vars.GcsKeyMountPath,
				"readOnly":  true,
			})
		}
	case spec.GetStorage().GetAzure() != nil:
		azure := spec.GetStorage().GetAzure()
		trace["backend"] = "azure"
		azureValues := map[string]interface{}{
			"storage_account_name": azure.GetAccountName(),
			"container_name":       azure.GetContainer(),
		}
		if key := azure.GetAccountKeySecret(); key != nil {
			azureValues["storage_account_key"] = fmt.Sprintf("${%s}", vars.EnvAzureAccountKey)
			credentialEnv = append(credentialEnv, secretEnvVar(vars.EnvAzureAccountKey, key))
		} else {
			azureValues["use_federated_token"] = true
		}
		trace["azure"] = azureValues
	default:
		// local (the default) — trace blocks on the persistent volume.
		trace["backend"] = "local"
		trace["local"] = map[string]interface{}{"path": vars.TracesLocalPath}
	}
	tempo["storage"] = map[string]interface{}{"trace": trace}

	// ---- metrics generator ---------------------------------------------
	if mg := spec.GetMetricsGenerator(); mg.GetEnabled() {
		remoteWriteUrl := normalizeRemoteWriteUrl(mg.GetRemoteWriteUrl().GetValue())
		tempo["metricsGenerator"] = map[string]interface{}{
			"enabled":        true,
			"remoteWriteUrl": remoteWriteUrl,
		}
		// The active processor LIST is a per-tenant override; empty =
		// both processors (the spec's documented default).
		processors := processorList(mg.GetProcessors())
		tempo["overrides"] = map[string]interface{}{
			"defaults": map[string]interface{}{
				"metrics_generator": map[string]interface{}{
					"processors": processors,
				},
			},
		}
	}

	// ---- resources -----------------------------------------------------
	if r := resourcesMap(spec.GetResources()); r != nil {
		tempo["resources"] = r
	}

	// ---- credential env + expand-env -----------------------------------
	if len(credentialEnv) > 0 {
		tempo["extraEnv"] = credentialEnv
		// ${VAR} placeholders in the config are inert without expansion.
		tempo["extraArgs"] = map[string]interface{}{"config.expand-env": "true"}
	}
	if len(extraVolumeMounts) > 0 {
		tempo["extraVolumeMounts"] = extraVolumeMounts
	}

	// ---- image + pull secrets ------------------------------------------
	// The tempo image is SPLIT (registry + repository) and rides
	// global.imageRegistry; tempo-query is the COMBINED docker-library
	// form the global override does not reach — re-point it explicitly.
	if reg := spec.GetImageRegistry(); reg != "" {
		values["global"] = map[string]interface{}{"imageRegistry": reg}
	}
	if len(spec.GetImagePullSecrets()) > 0 {
		pullSecrets := make([]interface{}, 0, len(spec.GetImagePullSecrets()))
		for _, name := range spec.GetImagePullSecrets() {
			pullSecrets = append(pullSecrets, name)
		}
		tempo["pullSecrets"] = pullSecrets
	}

	values["tempo"] = tempo

	// Tempo reads its object-store credentials only at start: the
	// fingerprint of the module-owned Secret changes the pod template when
	// the key rotates, so the next apply rolls Tempo onto it.
	if locals.R2 != nil {
		values["podAnnotations"] = map[string]interface{}{
			vars.CredentialsChecksumAnnotation: credentialsChecksum(locals.R2.SecretData),
		}
	}

	// ---- tempo-query sidecar -------------------------------------------
	if spec.GetTempoQueryEnabled() {
		tempoQuery := map[string]interface{}{"enabled": true}
		if reg := spec.GetImageRegistry(); reg != "" {
			tempoQuery["repository"] = fmt.Sprintf("%s/%s", reg, vars.TempoQueryRepository)
		}
		values["tempoQuery"] = tempoQuery
	}

	// ---- observability -------------------------------------------------
	if spec.GetServiceMonitorEnabled() {
		values["serviceMonitor"] = map[string]interface{}{"enabled": true}
	}

	// ---- scheduling (top-level on this chart) --------------------------
	if sched := spec.GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			values["nodeSelector"] = stringMapToInterface(sched.GetNodeSelector())
		}
		if len(sched.GetTolerations()) > 0 {
			values["tolerations"] = tolerationsSlice(sched.GetTolerations())
		}
		if sched.GetPriorityClassName() != "" {
			values["priorityClassName"] = sched.GetPriorityClassName()
		}
	}

	// ---- gcs key volume (top-level pod volume) -------------------------
	if len(extraVolumes) > 0 {
		values["extraVolumes"] = extraVolumes
	}

	// ---- escape hatch (merged LAST) ------------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	// fullnameOverride re-pinned AFTER the merge — every child name (and
	// the exported outputs built from them) derives from the fullname.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// normalizeRemoteWriteUrl appends Prometheus' standard remote-write path
// when the declared URL carries none (a bare Service endpoint like the
// stack's prometheus_endpoint output). Twin: the Terraform module's
// remote_write_url coalesce.
func normalizeRemoteWriteUrl(url string) string {
	if url == "" {
		return url
	}
	// A path is present when a "/" appears past the scheme's host.
	rest := url
	if idx := strings.Index(url, "://"); idx >= 0 {
		rest = url[idx+len("://"):]
	}
	if strings.Contains(rest, "/") {
		return url
	}
	return fmt.Sprintf("%s/api/v1/write", url)
}

// processorList maps the typed processor enum onto Tempo's hyphenated
// override vocabulary; empty = both (the documented default).
func processorList(processors []kubernetestempov1alpha1.KubernetesTempoMetricsGeneratorProcessor) []interface{} {
	if len(processors) == 0 {
		return []interface{}{"service-graphs", "span-metrics"}
	}
	out := make([]interface{}, 0, len(processors))
	for _, p := range processors {
		switch p {
		case kubernetestempov1alpha1.KubernetesTempoMetricsGeneratorProcessor_service_graphs:
			out = append(out, "service-graphs")
		case kubernetestempov1alpha1.KubernetesTempoMetricsGeneratorProcessor_span_metrics:
			out = append(out, "span-metrics")
		}
	}
	return out
}

// intOrDefault resolves an optional int32 to its proto default.
func intOrDefault(v *int32, def int) int {
	if v == nil {
		return def
	}
	return int(*v)
}

// boolOrDefault resolves an optional bool to its proto default.
func boolOrDefault(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

// secretEnvVar renders one credential env var sourced from an existing
// Secret (the value never exists in rendered values or state).
func secretEnvVar(name string, ref *kubernetestempov1alpha1.KubernetesTempoSecretKeyRef) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
		"valueFrom": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{
				"name": ref.GetName(),
				"key":  ref.GetKey(),
			},
		},
	}
}
