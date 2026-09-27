package module

import (
	"strconv"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sigs.k8s.io/yaml"
)

// helmRelease installs OpenFGA from the official chart as a real Helm
// release (helm.v3 Release — Helm's own lifecycle, hooks and rollback;
// never the client-side Chart resource).
func helmRelease(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) error {
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
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
		// Wait for the server to become Ready. Safe here ONLY because
		// migrations run as an init container (buildHelmValues): the
		// chart's default hook-Job mode would deadlock this wait — the
		// Deployment's wait-for-migration init container waits on a
		// post-install hook Job that Helm only runs AFTER --wait
		// returns. SkipAwait false is Helm --wait, stated explicitly to
		// mirror the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependsOn...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install openfga helm release")
	}

	return nil
}

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace).
//
// PARITY: the Terraform module reaches the same result natively — its
// helm_release passes values = [yamlencode(typed values), helm_values,
// yamlencode(fullnameOverride re-pin)] and the provider merges the
// documents in exactly this order. Keep every typed mapping below in
// lockstep with the Terraform module's locals.
//
// ESCAPE-HATCH REALITY (verified at chart 0.3.10): the chart ships a
// CLOSED values schema (values.schema.json additionalProperties: false) —
// a key the chart does not define fails the install outright, so
// helm_values can only override EXISTING chart values (extraEnvVars for
// the ~50 server flags without values paths, TLS file wiring, sidecars),
// never invent new ones.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec
	values := map[string]interface{}{}

	// fullnameOverride pins openfga.fullname to the resource name: the
	// Service, ServiceAccount, `-migrate` Job and `-datastore-secret`
	// all derive deterministically, and the exported endpoints are built
	// from it.
	values["fullnameOverride"] = locals.ReleaseName

	// ---- replicas ------------------------------------------------------------
	// Skipped under autoscaling (the chart omits the Deployment replicas
	// field entirely when autoscaling.enabled — the HPA owns the count).
	// The memory arm renders 1 EXPLICITLY: the chart forces it anyway
	// (ternary on engine == "memory" — each replica would hold its own
	// divergent authorization world), rendered here so the manifest
	// states the intent.
	if locals.DatastoreEngine == "memory" {
		values["replicaCount"] = 1
	} else if !spec.GetHpa().GetEnabled() && spec.Replicas != nil {
		values["replicaCount"] = int(spec.GetReplicas())
	}

	// ---- datastore ------------------------------------------------------------
	// MIGRATIONS RUN AS AN INIT CONTAINER, ALWAYS (`openfga migrate` is
	// idempotent with embedded migrations). The chart's default hook-Job
	// mode is deliberately not used: (1) it DEADLOCKS engines that wait
	// on rollout readiness — Helm --wait waits for the Deployment, whose
	// wait-for-migration init container waits for a post-install hook
	// Job that Helm only runs after --wait; (2) its hook list includes
	// post-delete, which dials the database during uninstall.
	//
	// SECRET DISCIPLINE (chart precedence, _helpers.tpl envConfig): each
	// credential field resolves its own branch independently, so mixing
	// a plain username with existingSecret+passwordKey is legal and
	// exact — username lands in the chart-owned `<fullname>-datastore-
	// secret` (usernames are not secrets), while the password rides a
	// secretKeyRef into the referenced existing Secret. The URI carries
	// no userinfo; the server prefers flag-supplied credentials over
	// URI-embedded ones. Nothing credential-bearing lands in these
	// values.
	datastore := map[string]interface{}{
		"engine":          locals.DatastoreEngine,
		"applyMigrations": true,
		"migrationType":   "initContainer",
	}
	if locals.DatastoreUri != "" {
		datastore["uri"] = locals.DatastoreUri
	}
	if locals.DatastoreUsername != "" {
		datastore["username"] = locals.DatastoreUsername
	}
	if locals.PasswordSecretName != "" {
		datastore["existingSecret"] = locals.PasswordSecretName
		datastore["secretKeys"] = map[string]interface{}{
			"passwordKey": locals.PasswordSecretKey,
		}
	}
	if ds := spec.GetDatastore(); ds != nil {
		if ds.MaxOpenConns != nil {
			datastore["maxOpenConns"] = int(ds.GetMaxOpenConns())
		}
		if ds.MaxIdleConns != nil {
			datastore["maxIdleConns"] = int(ds.GetMaxIdleConns())
		}
		if ds.GetConnMaxIdleTime() != "" {
			datastore["connMaxIdleTime"] = ds.GetConnMaxIdleTime()
		}
		if ds.GetConnMaxLifetime() != "" {
			datastore["connMaxLifetime"] = ds.GetConnMaxLifetime()
		}
	}
	values["datastore"] = datastore

	// migrate.timeout is consumed by the initContainer branch as
	// OPENFGA_TIMEOUT (chart-truth: deployment.yaml renders it into the
	// migrate-database init container's env) — how long `openfga
	// migrate` retries an unreachable database before failing the pod.
	// Meaningless on the memory arm (no migrations run).
	if locals.DatastoreEngine != "memory" && spec.GetDatastore().GetMigrationTimeout() != "" {
		values["migrate"] = map[string]interface{}{
			"timeout": spec.GetDatastore().GetMigrationTimeout(),
		}
	}

	// ---- playground: ALWAYS OFF -------------------------------------------------
	// The chart ships its demo playground ENABLED by default; this
	// module unconditionally disables it. Verified at OpenFGA v1.18.1:
	// upstream turned the playground off by default for security
	// (GHSA-68m9-983m-f3v5), the server REFUSES TO START when the
	// playground combines with ANY authn method, and at this version it
	// binds 127.0.0.1 pod-local — the chart's playground Service port
	// cannot reach it anyway.
	values["playground"] = map[string]interface{}{"enabled": false}

	// Every OpenFGA pod carries the resource's identity labels (the
	// migration runs as an init container in the same pod).
	values["podExtraLabels"] = stringMapToInterface(locals.Labels)

	// ---- authn --------------------------------------------------------------------
	// Unset renders NOTHING (server default: no authentication). Keys
	// reach the server through a Kubernetes Secret in both preshared
	// arms — authn.preshared.keys (plaintext keys into the Deployment
	// manifest) is never rendered.
	if authn := spec.GetAuthn(); authn != nil {
		if authn.GetPreshared() != nil {
			values["authn"] = map[string]interface{}{
				"method": "preshared",
				"preshared": map[string]interface{}{
					"keysSecret": locals.PresharedKeysSecretRef,
				},
			}
		}
		if oidc := authn.GetOidc(); oidc != nil {
			values["authn"] = map[string]interface{}{
				"method": "oidc",
				"oidc": map[string]interface{}{
					"issuer":   oidc.GetIssuer(),
					"audience": oidc.GetAudience(),
				},
			}
		}
	}

	// ---- telemetry -------------------------------------------------------------------
	// metrics.enabled is rendered EXPLICITLY in both states (chart
	// default true) so the manifest states the intent; the
	// ServiceMonitor requires the Prometheus Operator CRDs — the
	// install fails without them.
	metricsEnabled := true
	if m := spec.GetMetrics(); m != nil && m.Enabled != nil {
		metricsEnabled = m.GetEnabled()
	}
	telemetryMetrics := map[string]interface{}{"enabled": metricsEnabled}
	if spec.GetMetrics().GetServiceMonitorEnabled() {
		telemetryMetrics["serviceMonitor"] = map[string]interface{}{"enabled": true}
	}
	if spec.GetMetrics().GetEnableRpcHistograms() {
		telemetryMetrics["enableRPCHistograms"] = true
	}
	telemetry := map[string]interface{}{"metrics": telemetryMetrics}

	if tracing := spec.GetTracing(); tracing.GetEnabled() {
		trace := map[string]interface{}{
			"enabled": true,
			"otlp":    map[string]interface{}{"endpoint": tracing.GetOtlpEndpoint()},
		}
		if tracing.GetSampleRatio() != "" {
			// The chart schema types sampleRatio as a NUMBER — a string
			// would fail the closed-schema validation. The proto pattern
			// guarantees a parseable 0.0–1.0 (Terraform twin: tonumber).
			ratio, err := strconv.ParseFloat(tracing.GetSampleRatio(), 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse tracing.sample_ratio as a number")
			}
			trace["sampleRatio"] = ratio
		}
		telemetry["trace"] = trace
	}
	values["telemetry"] = telemetry

	// ---- log ------------------------------------------------------------------------
	if log := spec.GetLog(); log != nil && (log.GetLevel() != "" || log.GetFormat() != "") {
		logBlock := map[string]interface{}{}
		if log.GetLevel() != "" {
			logBlock["level"] = log.GetLevel()
		}
		if log.GetFormat() != "" {
			logBlock["format"] = log.GetFormat()
		}
		values["log"] = logBlock
	}

	// ---- tuning (top-level chart values, 1:1) ------------------------------------------
	// Each renders ONLY when set — the chart guards them with `if`, so
	// absent means the server's own default. The two MaxResults fields
	// are `ne nil` guards in the chart (an explicit 0 = unlimited is
	// expressible), hence the pointer checks that pass 0 through.
	if tuning := spec.GetTuning(); tuning != nil {
		if tuning.MaxTuplesPerWrite != nil {
			values["maxTuplesPerWrite"] = int(tuning.GetMaxTuplesPerWrite())
		}
		if tuning.MaxTypesPerAuthorizationModel != nil {
			values["maxTypesPerAuthorizationModel"] = int(tuning.GetMaxTypesPerAuthorizationModel())
		}
		if tuning.MaxChecksPerBatchCheck != nil {
			values["maxChecksPerBatchCheck"] = int(tuning.GetMaxChecksPerBatchCheck())
		}
		if tuning.GetListObjectsDeadline() != "" {
			values["listObjectsDeadline"] = tuning.GetListObjectsDeadline()
		}
		if tuning.ListObjectsMaxResults != nil {
			values["listObjectsMaxResults"] = int(tuning.GetListObjectsMaxResults())
		}
		if tuning.GetListUsersDeadline() != "" {
			values["listUsersDeadline"] = tuning.GetListUsersDeadline()
		}
		if tuning.ListUsersMaxResults != nil {
			values["listUsersMaxResults"] = int(tuning.GetListUsersMaxResults())
		}
		if tuning.GetRequestTimeout() != "" {
			values["requestTimeout"] = tuning.GetRequestTimeout()
		}
		if cqc := tuning.GetCheckQueryCache(); cqc != nil {
			checkQueryCache := map[string]interface{}{"enabled": cqc.GetEnabled()}
			if cqc.Limit != nil {
				checkQueryCache["limit"] = int(cqc.GetLimit())
			}
			if cqc.GetTtl() != "" {
				checkQueryCache["ttl"] = cqc.GetTtl()
			}
			values["checkQueryCache"] = checkQueryCache
		}
		// Server contract at v1.18.1: this list REPLACES the server's
		// own default experimental set — the spec comment teaches it.
		if len(tuning.GetExperimentals()) > 0 {
			values["experimentals"] = toInterfaceSlice(tuning.GetExperimentals())
		}
	}

	// ---- container resources -------------------------------------------------------------
	if resources := resourcesBlock(spec.GetResources()); resources != nil {
		values["resources"] = resources
	}

	// ---- autoscaling ------------------------------------------------------------------------
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
		}
		values["autoscaling"] = autoscaling
	}

	// ---- scheduling ------------------------------------------------------------------------------
	if scheduling := spec.GetScheduling(); scheduling != nil {
		if len(scheduling.GetNodeSelector()) > 0 {
			values["nodeSelector"] = stringMapToInterface(scheduling.GetNodeSelector())
		}
		if len(scheduling.GetTolerations()) > 0 {
			values["tolerations"] = tolerationsSlice(scheduling.GetTolerations())
		}
	}

	// ---- service account -----------------------------------------------------------------------------
	// Only the annotations seam (cloud workload identity) is modeled.
	// serviceAccount.create=false is deliberately unsupported: it would
	// silently drop the Job-status RBAC the wait-for-migration init
	// container needs if anyone flips migrationType back to "job".
	if len(spec.GetServiceAccountAnnotations()) > 0 {
		values["serviceAccount"] = map[string]interface{}{
			"annotations": stringMapToInterface(spec.GetServiceAccountAnnotations()),
		}
	}

	// ---- escape hatch (merged LAST, Helm -f semantics) ----------------------------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as YAML")
		}
		values = mergeMaps(values, overrides)
	}

	// fullnameOverride re-pinned AFTER the merge — the one deliberate
	// exception to the escape hatch's last-word contract (twin of the
	// Terraform module's third values document). The Service /
	// `-migrate` Job / `-datastore-secret` names — and the exported
	// endpoints built from them — all derive from the fullname; letting
	// an override move it would break every output.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// resourcesBlock renders a ContainerResources into the chart's resources
// shape (nil when nothing is declared).
func resourcesBlock(r *kubernetesprovider.ContainerResources) map[string]interface{} {
	if r == nil {
		return nil
	}
	resources := map[string]interface{}{}
	if q := r.GetRequests(); q != nil && (q.GetCpu() != "" || q.GetMemory() != "") {
		requests := map[string]interface{}{}
		if q.GetCpu() != "" {
			requests["cpu"] = q.GetCpu()
		}
		if q.GetMemory() != "" {
			requests["memory"] = q.GetMemory()
		}
		resources["requests"] = requests
	}
	if l := r.GetLimits(); l != nil && (l.GetCpu() != "" || l.GetMemory() != "") {
		limits := map[string]interface{}{}
		if l.GetCpu() != "" {
			limits["cpu"] = l.GetCpu()
		}
		if l.GetMemory() != "" {
			limits["memory"] = l.GetMemory()
		}
		resources["limits"] = limits
	}
	if len(resources) == 0 {
		return nil
	}
	return resources
}

// tolerationsSlice renders the shared WorkloadToleration list into the
// chart's tolerations shape.
func tolerationsSlice(tolerations []*kubernetesprovider.WorkloadToleration) []interface{} {
	out := make([]interface{}, 0, len(tolerations))
	for _, t := range tolerations {
		tol := map[string]interface{}{}
		if t.GetKey() != "" {
			tol["key"] = t.GetKey()
		}
		if t.GetOperator() != "" {
			tol["operator"] = t.GetOperator()
		}
		if t.GetValue() != "" {
			tol["value"] = t.GetValue()
		}
		if t.GetEffect() != "" {
			tol["effect"] = t.GetEffect()
		}
		if t.TolerationSeconds != nil {
			tol["tolerationSeconds"] = t.GetTolerationSeconds()
		}
		out = append(out, tol)
	}
	return out
}

// mergeMaps deep-merges b over a with Helm's `-f` semantics: nested maps
// merge recursively with b winning per key; everything else (scalars,
// lists) is replaced by b's value.
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if bChild, ok := v.(map[string]interface{}); ok {
			if aChild, ok := out[k].(map[string]interface{}); ok {
				out[k] = mergeMaps(aChild, bChild)
				continue
			}
		}
		out[k] = v
	}
	return out
}

// stringMapToInterface converts a map[string]string into the
// map[string]interface{} YAML rendering expects.
func stringMapToInterface(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// toInterfaceSlice converts a string slice into the []interface{} YAML
// rendering expects.
func toInterfaceSlice(in []string) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, v := range in {
		out = append(out, v)
	}
	return out
}
