package module

import (
	"fmt"

	"github.com/pkg/errors"
	kuberneteslokiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesloki/v1alpha1"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace).
//
// MODE DISCIPLINE (load-bearing): exactly one deployment mode renders, and
// every OTHER mode's workloads are explicitly zeroed — the chart's own
// reference values files zero them by hand, and leaving that to the
// operator's care is how two modes end up half-running side by side.
//
// SECRET DISCIPLINE (load-bearing): the chart renders the Loki
// configuration into a Secret/ConfigMap visible to anyone with read access
// on the namespace's rendered values. Declared object-store credentials
// therefore NEVER appear in these values: they travel as environment
// variables sourced from the referenced Secrets (defaults.extraEnv
// secretKeyRefs), the config references them as ${VAR} placeholders, and
// -config.expand-env=true makes Loki expand them at process start.
//
// PARITY: the Terraform module reaches the same result natively — its
// helm_release passes values = [yamlencode(typed), helm_values,
// yamlencode(re-pin)] and the provider merges the documents in exactly
// this order. Keep every typed mapping below in lockstep with the
// Terraform module's locals.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec

	// The chart composes child names like `<fullname>-backend-headless`
	// and truncates the COMPOSED name at 63 characters — an over-long
	// resource name corrupts the naming contract the outputs promise.
	// Fail THIS deploy loudly instead (twin: the Terraform module's
	// plan-time precondition).
	if len(locals.ReleaseName) > vars.MaxNameLength {
		return nil, errors.Errorf(
			"metadata.name %q is %d characters — the loki chart's child-name budget allows at most %d "+
				"(it composes names like <name>-backend-headless within Kubernetes' 63-character cap)",
			locals.ReleaseName, len(locals.ReleaseName), vars.MaxNameLength)
	}

	values := map[string]interface{}{}

	// fullnameOverride pins loki.fullname to the resource name: the
	// gateway Service, the Loki Services and the memberlist name all
	// derive deterministically, and the exported outputs are built from
	// that contract.
	values["fullnameOverride"] = locals.ReleaseName

	// ---- storage backend resolution ------------------------------------
	// One derivation feeds three surfaces: loki.storage.type, the derived
	// schema's object_store, and the compactor's delete_request_store.
	backendType := storageBackendType(spec.GetStorage())

	// ---- deployment mode + workloads -----------------------------------
	// The write path's replica count also bounds the replication factor:
	// Loki refuses a replication_factor above the number of ingesting
	// replicas, and 3 is the HA ceiling worth defaulting to.
	writePathReplicas := 1
	if ssd := spec.GetSimpleScalable(); ssd != nil {
		values["deploymentMode"] = "SimpleScalable"
		writeReplicas := intOrDefault(ssd.WriteReplicas, 3)
		readReplicas := intOrDefault(ssd.ReadReplicas, 3)
		backendReplicas := intOrDefault(ssd.BackendReplicas, 3)
		writePathReplicas = writeReplicas

		diskSize := ssd.GetDiskSize()
		if diskSize == "" {
			diskSize = vars.DefaultDiskSize
		}
		persistence := persistenceMap(diskSize, ssd.GetStorageClass().GetValue())

		write := map[string]interface{}{"replicas": writeReplicas, "persistence": persistence}
		backend := map[string]interface{}{"replicas": backendReplicas, "persistence": persistence}
		read := map[string]interface{}{"replicas": readReplicas}
		if r := resourcesMap(ssd.GetResources()); r != nil {
			write["resources"] = r
			backend["resources"] = r
			read["resources"] = r
		}
		values["write"] = write
		values["backend"] = backend
		values["read"] = read
		// The other mode's workload, zeroed.
		values["singleBinary"] = map[string]interface{}{"replicas": 0}
	} else {
		// Monolithic — the default mode (an absent oneof means a
		// single-replica monolithic instance, the spec's documented
		// default).
		values["deploymentMode"] = "Monolithic"
		mono := spec.GetMonolithic()
		replicas := 1
		diskSize := vars.DefaultDiskSize
		storageClass := ""
		var resources map[string]interface{}
		if mono != nil {
			replicas = intOrDefault(mono.Replicas, 1)
			if mono.GetDiskSize() != "" {
				diskSize = mono.GetDiskSize()
			}
			storageClass = mono.GetStorageClass().GetValue()
			resources = resourcesMap(mono.GetResources())
		}
		writePathReplicas = replicas

		singleBinary := map[string]interface{}{
			"replicas":    replicas,
			"persistence": persistenceMap(diskSize, storageClass),
		}
		if resources != nil {
			singleBinary["resources"] = resources
		}
		values["singleBinary"] = singleBinary
		// The other mode's workloads, zeroed.
		values["write"] = map[string]interface{}{"replicas": 0}
		values["read"] = map[string]interface{}{"replicas": 0}
		values["backend"] = map[string]interface{}{"replicas": 0}
	}

	// The microservices-mode components, zeroed in EVERY rendering.
	for _, component := range distributedComponents {
		values[component] = map[string]interface{}{"replicas": 0}
	}

	// ---- the loki config block -----------------------------------------
	replicationFactor := writePathReplicas
	if replicationFactor > 3 {
		replicationFactor = 3
	}

	schemaFrom := spec.GetSchemaFromDate()
	if schemaFrom == "" {
		schemaFrom = vars.DefaultSchemaFromDate
	}

	loki := map[string]interface{}{
		// ALWAYS rendered: the component's single-tenant default
		// deliberately diverges from the chart's multi-tenant-on default
		// (auth_enabled: true), so relying on the chart default would
		// invert the spec's contract.
		"auth_enabled": spec.GetMultiTenancy().GetEnabled(),
		"commonConfig": map[string]interface{}{
			"replication_factor": replicationFactor,
		},
		// The derived index schema — TSDB on v13 with 24h periods, its
		// object_store matching the storage backend. Upstream makes
		// every user hand-author this block; here a new install never
		// writes one, and imports override only the start date.
		"schemaConfig": map[string]interface{}{
			"configs": []interface{}{
				map[string]interface{}{
					"from":         schemaFrom,
					"store":        vars.SchemaStore,
					"object_store": backendType,
					"schema":       vars.SchemaVersion,
					"index": map[string]interface{}{
						"prefix": vars.SchemaIndexPrefix,
						"period": vars.SchemaIndexPeriod,
					},
				},
			},
		},
	}

	// Anonymous usage reporting: OFF unless explicitly opted in (the
	// spec's documented divergence from Loki's report-by-default).
	if !boolOrDefault(spec.UsageReporting, false) {
		loki["analytics"] = map[string]interface{}{"reporting_enabled": false}
	}

	// ---- storage ---------------------------------------------------------
	storage := map[string]interface{}{"type": backendType}
	credentialEnv := []interface{}{}
	extraVolumes := []interface{}{}
	extraVolumeMounts := []interface{}{}
	switch {
	case locals.R2 != nil:
		// The r2 arm: R2 speaks S3, so loki.storage.type is "s3" and the
		// values are composed from R2's vocabulary (r2.go). The key pair
		// rides env expansion from the module-owned Secret.
		storage["bucketNames"] = locals.R2.BucketNames
		storage["s3"] = locals.R2.S3Values
		credentialEnv = append(credentialEnv, r2CredentialEnv(locals.R2CredentialsSecretName)...)
	case backendType == "s3":
		s3 := spec.GetStorage().GetS3()
		storage["bucketNames"] = bucketNames(s3.GetBucket(), s3.GetRulerBucket())
		s3Values := map[string]interface{}{}
		if s3.GetEndpoint() != "" {
			s3Values["endpoint"] = s3.GetEndpoint()
		}
		if s3.GetRegion() != "" {
			s3Values["region"] = s3.GetRegion()
		}
		if s3.GetForcePathStyle() {
			s3Values["s3ForcePathStyle"] = true
		}
		if s3.GetInsecure() {
			s3Values["insecure"] = true
		}
		// Declared credentials ride env expansion — never these values
		// (see the SECRET DISCIPLINE note). Keyless (credentials
		// absent) = the pod's ambient identity (IRSA), nothing rendered.
		if creds := s3.GetCredentials(); creds != nil {
			s3Values["accessKeyId"] = fmt.Sprintf("${%s}", vars.EnvS3AccessKeyId)
			s3Values["secretAccessKey"] = fmt.Sprintf("${%s}", vars.EnvS3SecretAccessKey)
			credentialEnv = append(credentialEnv,
				secretEnvVar(vars.EnvS3AccessKeyId, creds.GetAccessKeyIdSecret()),
				secretEnvVar(vars.EnvS3SecretAccessKey, creds.GetSecretAccessKeySecret()),
			)
		}
		if len(s3Values) > 0 {
			storage["s3"] = s3Values
		}
	case backendType == "gcs":
		gcs := spec.GetStorage().GetGcs()
		storage["bucketNames"] = bucketNames(gcs.GetBucket(), gcs.GetRulerBucket())
		// A declared service-account key is mounted from the referenced
		// Secret and named through GOOGLE_APPLICATION_CREDENTIALS — the
		// GCS client library's own contract. Keyless = GKE workload
		// identity, nothing rendered.
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
	case backendType == "azure":
		azure := spec.GetStorage().GetAzure()
		storage["bucketNames"] = bucketNames(azure.GetContainer(), azure.GetRulerContainer())
		azureValues := map[string]interface{}{
			"accountName": azure.GetAccountName(),
		}
		if key := azure.GetAccountKeySecret(); key != nil {
			azureValues["accountKey"] = fmt.Sprintf("${%s}", vars.EnvAzureAccountKey)
			credentialEnv = append(credentialEnv, secretEnvVar(vars.EnvAzureAccountKey, key))
		} else {
			// Keyless = AKS federated workload identity (the spec's
			// documented recommended posture).
			azureValues["useFederatedToken"] = true
		}
		storage["azure"] = azureValues
	}
	loki["storage"] = storage
	// Loki reads its object-store credentials only at start: the
	// fingerprint of the module-owned Secret changes every Loki pod's
	// template when the key rotates, so the next apply rolls them onto it.
	if locals.R2 != nil {
		loki["podAnnotations"] = map[string]interface{}{
			vars.CredentialsChecksumAnnotation: credentialsChecksum(locals.R2.SecretData),
		}
	}

	// ---- limits + retention ----------------------------------------------
	limitsConfig := map[string]interface{}{}
	if spec.GetRetentionPeriod() != "" {
		limitsConfig["retention_period"] = spec.GetRetentionPeriod()
		// Compactor-driven deletion: retention_enabled without a
		// delete_request_store is a startup error — the store follows
		// the chunk backend.
		loki["compactor"] = map[string]interface{}{
			"retention_enabled":    true,
			"delete_request_store": backendType,
		}
	}
	if limits := spec.GetLimits(); limits != nil {
		if limits.IngestionRateMb != nil {
			limitsConfig["ingestion_rate_mb"] = int(limits.GetIngestionRateMb())
		}
		if limits.IngestionBurstSizeMb != nil {
			limitsConfig["ingestion_burst_size_mb"] = int(limits.GetIngestionBurstSizeMb())
		}
		if limits.MaxGlobalStreamsPerUser != nil {
			limitsConfig["max_global_streams_per_user"] = int(limits.GetMaxGlobalStreamsPerUser())
		}
		if limits.GetMaxQueryLookback() != "" {
			limitsConfig["max_query_lookback"] = limits.GetMaxQueryLookback()
		}
	}
	if len(limitsConfig) > 0 {
		loki["limits_config"] = limitsConfig
	}

	// ---- ruler -------------------------------------------------------------
	// Rules are discovered by the chart's rules sidecar (on by default,
	// label loki_rule) into /rules — the ruler reads them from that local
	// directory and fires at the declared Alertmanager. Ruler storage
	// tuning beyond this rides helm_values.
	if ruler := spec.GetRuler(); ruler.GetEnabled() {
		loki["rulerConfig"] = map[string]interface{}{
			"alertmanager_url": ruler.GetAlertmanagerUrl().GetValue(),
			"storage": map[string]interface{}{
				"type":  "local",
				"local": map[string]interface{}{"directory": "/rules"},
			},
		}
	}

	// ---- multi-tenancy -------------------------------------------------------
	// The chart builds the gateway's htpasswd from loki.tenants (bcrypt
	// passwordHash entries — one-way hashes, the chart's own documented
	// pattern; the actual passwords never exist here). An existing
	// htpasswd Secret arm bypasses the list entirely.
	if mt := spec.GetMultiTenancy(); mt.GetEnabled() && len(mt.GetTenants()) > 0 {
		tenants := make([]interface{}, 0, len(mt.GetTenants()))
		for _, tenant := range mt.GetTenants() {
			tenants = append(tenants, map[string]interface{}{
				"name":         tenant.GetName(),
				"passwordHash": tenant.GetPasswordHash(),
			})
		}
		loki["tenants"] = tenants
	}

	values["loki"] = loki

	// ---- gateway ----------------------------------------------------------
	gateway := map[string]interface{}{}
	if !locals.GatewayEnabled {
		gateway["enabled"] = false
	}
	if gw := spec.GetGateway(); gw != nil {
		if gw.Replicas != nil && gw.GetReplicas() != 1 {
			gateway["replicas"] = int(gw.GetReplicas())
		}
		if r := resourcesMap(gw.GetResources()); r != nil {
			gateway["resources"] = r
		}
	}
	// Basic auth gates the gateway when multi-tenancy is on: the tenant
	// list feeds the chart's own htpasswd template; an existing Secret
	// (key `.htpasswd`) replaces it.
	if mt := spec.GetMultiTenancy(); mt.GetEnabled() && (len(mt.GetTenants()) > 0 || mt.GetExistingHtpasswdSecret() != "") {
		basicAuth := map[string]interface{}{"enabled": true}
		if mt.GetExistingHtpasswdSecret() != "" {
			basicAuth["existingSecret"] = mt.GetExistingHtpasswdSecret()
		}
		gateway["basicAuth"] = basicAuth
	}
	if len(gateway) > 0 {
		values["gateway"] = gateway
	}

	// ---- caches --------------------------------------------------------------
	if caching := spec.GetCaching(); caching != nil {
		if chunks := cacheValues(caching.ChunksCacheEnabled, caching.ChunksCacheMemoryMb); chunks != nil {
			values["chunksCache"] = chunks
		}
		if results := cacheValues(caching.ResultsCacheEnabled, caching.ResultsCacheMemoryMb); results != nil {
			values["resultsCache"] = results
		}
	}

	// ---- canary ----------------------------------------------------------------
	// The chart default is ON (a DaemonSet continuously proving the
	// write→read pipeline); rendered only when disabled. The chart's Helm
	// test reads the canary's metrics and refuses to render without it, so
	// it goes off with the canary.
	if !boolOrDefault(spec.CanaryEnabled, true) {
		values["lokiCanary"] = map[string]interface{}{"enabled": false}
		values["test"] = map[string]interface{}{"enabled": false}
	}

	// ---- observability ------------------------------------------------------------
	if spec.GetServiceMonitorEnabled() {
		values["monitoring"] = map[string]interface{}{
			"serviceMonitor": map[string]interface{}{"enabled": true},
		}
	}

	// ---- images -----------------------------------------------------------------------
	// The chart MIXES image forms: loki/gateway/canary/sidecar images are
	// SPLIT (registry + repository — global.imageRegistry overrides all
	// their registries at once) while the memcached caches run the
	// docker-library `memcached` image (repository-only, COMBINED form)
	// that the global override does not reach — its repository is
	// re-pointed at the mirror explicitly. Mirrors laying out library
	// images elsewhere (e.g. <registry>/library/memcached) override via
	// helm_values.
	if spec.GetImageRegistry() != "" {
		values["global"] = map[string]interface{}{"imageRegistry": spec.GetImageRegistry()}
		values["memcached"] = map[string]interface{}{
			"image": map[string]interface{}{
				"repository": fmt.Sprintf("%s/%s", spec.GetImageRegistry(), vars.MemcachedRepository),
			},
		}
	}
	if len(spec.GetImagePullSecrets()) > 0 {
		pullSecrets := make([]interface{}, 0, len(spec.GetImagePullSecrets()))
		for _, name := range spec.GetImagePullSecrets() {
			pullSecrets = append(pullSecrets, name)
		}
		values["imagePullSecrets"] = pullSecrets
	}

	// ---- the loki workloads' shared defaults ------------------------------------------
	// defaults.* applies to every Loki component (single binary or the
	// write/read/backend tiers) — exactly the spec's scheduling scope.
	// The gateway, caches and canary keep the chart's own scheduling
	// (steered via helm_values). Credential env/volumes land here too so
	// every Loki component can expand its config's ${VAR} placeholders.
	defaults := map[string]interface{}{}
	if len(credentialEnv) > 0 {
		defaults["extraEnv"] = credentialEnv
		// ${VAR} placeholders in the config are inert without expansion.
		defaults["extraArgs"] = []interface{}{"-config.expand-env=true"}
	}
	if len(extraVolumes) > 0 {
		defaults["extraVolumes"] = extraVolumes
		defaults["extraVolumeMounts"] = extraVolumeMounts
	}
	if sched := spec.GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			defaults["nodeSelector"] = stringMapToInterface(sched.GetNodeSelector())
		}
		if len(sched.GetTolerations()) > 0 {
			defaults["tolerations"] = tolerationsSlice(sched.GetTolerations())
		}
		if sched.GetPriorityClassName() != "" {
			defaults["priorityClassName"] = sched.GetPriorityClassName()
		}
	}
	if len(defaults) > 0 {
		values["defaults"] = defaults
	}

	// ---- escape hatch (merged LAST, helm -f semantics) ----------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	// fullnameOverride re-pinned AFTER the merge — the one deliberate
	// exception to the escape hatch's last-word contract (twin of the
	// Terraform module's third values document). Every Service name —
	// and the exported outputs built from them — derives from the
	// fullname; letting an override move it would break every output.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// storageBackendType resolves the spec's storage oneof onto the chart's
// loki.storage.type vocabulary; an absent block means filesystem (the
// spec's documented dev default).
func storageBackendType(storage *kuberneteslokiv1alpha1.KubernetesLokiStorage) string {
	switch {
	case storage.GetS3() != nil, storage.GetR2() != nil:
		return "s3"
	case storage.GetGcs() != nil:
		return "gcs"
	case storage.GetAzure() != nil:
		return "azure"
	default:
		return "filesystem"
	}
}

// bucketNames renders the chart's bucketNames block: the ruler shares the
// chunks bucket unless a dedicated one is declared.
func bucketNames(chunks, ruler string) map[string]interface{} {
	if ruler == "" {
		ruler = chunks
	}
	return map[string]interface{}{
		"chunks": chunks,
		"ruler":  ruler,
	}
}

// secretEnvVar renders one credential env var sourced from an existing
// Secret (a secretKeyRef the kubelet resolves — the value never exists in
// rendered values or state).
func secretEnvVar(name string, ref *kuberneteslokiv1alpha1.KubernetesLokiSecretKeyRef) map[string]interface{} {
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

// cacheValues renders one memcached cache block (chunksCache /
// resultsCache share the shape); nil when nothing diverges from the chart
// defaults.
func cacheValues(enabled *bool, memoryMb *int32) map[string]interface{} {
	out := map[string]interface{}{}
	if enabled != nil && !*enabled {
		out["enabled"] = false
	}
	if memoryMb != nil {
		out["allocatedMemory"] = int(*memoryMb)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// intOrDefault resolves an optional int32 to its proto default — the
// tfvars path flattens presence, so both engines treat the default value
// and absence identically (the cross-engine value-based contract).
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

// persistenceMap renders a workload persistence block.
func persistenceMap(size, storageClass string) map[string]interface{} {
	persistence := map[string]interface{}{
		"enabled": true,
		"size":    size,
	}
	if storageClass != "" {
		persistence["storageClass"] = storageClass
	}
	return persistence
}
