package module

import (
	"strings"

	"github.com/pkg/errors"
	kubernetesvalkeyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesvalkey/v1alpha1"
	"sigs.k8s.io/yaml"
)

// defaultAclPermissions is the full-access ACL rule an ACL user falls back
// to when the spec leaves permissions unset — mirror of the proto field's
// default option and the Terraform module's coalesce. The chart REQUIRES a
// permissions value on every aclUsers entry (it fails the install
// otherwise), so the module always renders one.
const defaultAclPermissions = "~* &* +@all"

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

	// Pin the chart's fullname to the release name (= metadata.name): every
	// chart object then carries a deterministic, manifest-derived name —
	// the write Service renders as `<name>`, pod discovery as
	// `<name>-headless`, and the replication read Service as `<name>-read`,
	// which is exactly what the stack outputs promise and what lets several
	// Valkey instances coexist in one cluster.
	values["fullnameOverride"] = locals.ReleaseName

	// Every pod names its organization and environment (locals.Labels).
	values["podLabels"] = stringMapToInterface(locals.Labels)

	// ---- image -----------------------------------------------------------
	// Chart defaults: docker.io / valkey/valkey at the chart's app version.
	if img := spec.GetImage(); img != nil {
		image := map[string]interface{}{}
		if img.GetRegistry() != "" {
			image["registry"] = img.GetRegistry()
		}
		if img.GetRepository() != "" {
			image["repository"] = img.GetRepository()
		}
		if img.GetTag() != "" {
			image["tag"] = img.GetTag()
		}
		if len(image) > 0 {
			values["image"] = image
		}
	}

	// The chart consumes pull secrets as a plain list of Secret NAMES (its
	// imagePullSecrets helper renders `- name: <entry>` per string entry).
	if len(spec.GetImagePullSecrets()) > 0 {
		values["imagePullSecrets"] = toInterfaceSlice(spec.GetImagePullSecrets())
	}

	// Server log verbosity is a first-class chart value (valkeyLogLevel,
	// injected as the VALKEY_LOG_LEVEL env var) — NOT a valkey.conf
	// directive, so it deliberately stays out of valkeyConfig below.
	if spec.LogLevel != nil && spec.GetLogLevel() != "" {
		values["valkeyLogLevel"] = spec.GetLogLevel()
	}

	// ---- valkey.conf (module-owned rendering) ------------------------------
	// The chart accepts valkey.conf only as ONE raw string (valkeyConfig,
	// mounted via ConfigMap and appended after the chart's own generated
	// base config). The typed config block renders that string
	// deterministically; omitted/empty config renders no key at all.
	if rendered := renderValkeyConfig(spec.GetConfig()); rendered != "" {
		values["valkeyConfig"] = rendered
	}

	// ---- auth --------------------------------------------------------------
	// Declared ACL users render WITHOUT passwords: the chart reads each
	// user's password from the usersExistingSecret key named after the
	// username (its init script's get_user_password falls back
	// passwordKey -> username; the module leaves passwordKey unset). The
	// module materializes that Secret (secrets.go), so credentials never
	// appear in rendered chart values.
	if locals.AuthEnabled {
		values["auth"] = map[string]interface{}{
			"enabled":             true,
			"usersExistingSecret": locals.AuthSecretName,
			"aclUsers":            aclUsersMap(spec.GetAuth()),
		}
	}

	// ---- topology ------------------------------------------------------------
	if repl := spec.GetReplication(); repl != nil {
		// Replication: one primary plus N replicas from a StatefulSet.
		// Every scalar renders with its resolved default so both engines
		// emit the identical replica block.
		replicas := int32(2)
		if repl.Replicas != nil {
			replicas = repl.GetReplicas()
		}
		replicationUser := repl.GetReplicationUser()
		if replicationUser == "" {
			replicationUser = "default"
		}
		minReplicasMaxLag := int32(10)
		if repl.MinReplicasMaxLag != nil {
			minReplicasMaxLag = repl.GetMinReplicasMaxLag()
		}

		persistence := map[string]interface{}{
			"size": repl.GetPersistence().GetSize(),
		}
		if repl.GetPersistence().GetStorageClass().GetValue() != "" {
			persistence["storageClass"] = repl.GetPersistence().GetStorageClass().GetValue()
		}

		readServiceType := "ClusterIP"
		if rs := repl.GetReadService(); rs != nil && rs.GetType() != "" {
			readServiceType = rs.GetType()
		}
		readService := map[string]interface{}{
			"enabled": locals.ReadServiceEnabled,
			"type":    readServiceType,
		}
		if rs := repl.GetReadService(); rs != nil && len(rs.GetAnnotations()) > 0 {
			readService["annotations"] = stringMapToInterface(rs.GetAnnotations())
		}

		values["replica"] = map[string]interface{}{
			"enabled":            true,
			"replicas":           replicas,
			"replicationUser":    replicationUser,
			"disklessSync":       repl.GetDisklessSync(),
			"minReplicasToWrite": repl.GetMinReplicasToWrite(),
			"minReplicasMaxLag":  minReplicasMaxLag,
			"persistence":        persistence,
			"service":            readService,
		}
	} else if p := spec.GetPersistence(); p != nil {
		// Standalone persistence: the chart's dataStorage PVC (only read
		// by the standalone Deployment; replication uses
		// volumeClaimTemplates above instead).
		dataStorage := map[string]interface{}{
			"enabled":       true,
			"requestedSize": p.GetSize(),
			"keepPvc":       p.GetKeepOnUninstall(),
		}
		if p.GetStorageClass().GetValue() != "" {
			dataStorage["className"] = p.GetStorageClass().GetValue()
		}
		values["dataStorage"] = dataStorage
	}

	// ---- tls -----------------------------------------------------------------
	// The chart's key-name defaults (server.crt/server.key/ca.crt) predate
	// the kubernetes.io/tls convention; cert-manager Certificates store
	// their material as tls.crt/tls.key/ca.crt. The spec's certificate
	// seam is cert-manager, so the module pins the chart's key names to
	// the kubernetes.io/tls layout whenever TLS is enabled.
	if tls := spec.GetTls(); tls.GetEnabled() {
		values["tls"] = map[string]interface{}{
			"enabled":                  true,
			"existingSecret":           tls.GetCertificateSecret().GetValue(),
			"requireClientCertificate": tls.GetRequireClientCertificate(),
			"serverPublicKey":          "tls.crt",
			"serverKey":                "tls.key",
			"caPublicKey":              "ca.crt",
		}
	}

	// ---- probes ----------------------------------------------------------------
	// The chart ships readinessProbe disabled, so a pod that is starting or
	// still loading its dataset would join the Service at once; it is always
	// on here. Every probe the chart renders execs `valkey-cli ping`, which
	// presents no client certificate: under mutual TLS (tls-auth-clients
	// yes) each probe fails the handshake and the kubelet restarts the pod
	// in a loop. There all three probes are a TCP connect to the container
	// port the chart names `tcp`. The chart takes customProbe as the whole
	// probe, dropping its timing values, so the probes run on Kubernetes'
	// defaults, which are the chart's.
	readinessProbe := map[string]interface{}{"enabled": true}
	if tls := spec.GetTls(); tls.GetEnabled() && tls.GetRequireClientCertificate() {
		values["startupProbe"] = map[string]interface{}{"customProbe": tcpSocketProbe()}
		values["livenessProbe"] = map[string]interface{}{"customProbe": tcpSocketProbe()}
		readinessProbe["customProbe"] = tcpSocketProbe()
	}
	values["readinessProbe"] = readinessProbe

	// ---- write service ---------------------------------------------------------
	if svc := spec.GetService(); svc != nil {
		serviceType := svc.GetType()
		if serviceType == "" {
			serviceType = "ClusterIP"
		}
		service := map[string]interface{}{
			"type": serviceType,
			"port": locals.ServicePort,
		}
		if len(svc.GetAnnotations()) > 0 {
			service["annotations"] = stringMapToInterface(svc.GetAnnotations())
		}
		values["service"] = service
	}

	// ---- sizing ------------------------------------------------------------------
	if r := resourcesMap(spec.GetResources()); r != nil {
		values["resources"] = r
	}

	// ---- observability ----------------------------------------------------------
	// The ServiceMonitor toggle lives at metrics.serviceMonitor.enabled;
	// the chart additionally gates it on metrics.service.enabled, which
	// defaults to true — the metrics Service (`<name>-metrics`) always
	// accompanies the exporter here.
	if m := spec.GetMetrics(); m.GetEnabled() {
		metrics := map[string]interface{}{"enabled": true}
		if m.GetServiceMonitorEnabled() {
			metrics["serviceMonitor"] = map[string]interface{}{"enabled": true}
		}
		values["metrics"] = metrics
	}

	// ---- scheduling ----------------------------------------------------------------
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

	// ---- pod disruption budget --------------------------------------------------------
	// Rendered ONLY in replication mode: the chart's PDB template is gated
	// on replica.enabled, so a standalone PDB declaration would be a
	// silent no-op in the release — the module omits it instead of
	// rendering dead values. Exactly one bound is set (spec CEL rule);
	// with neither, the chart's own default (maxUnavailable: 1) applies.
	if pdb := spec.GetPodDisruptionBudget(); pdb.GetEnabled() && locals.ReplicationEnabled {
		podDisruptionBudget := map[string]interface{}{"enabled": true}
		if pdb.GetMaxUnavailable() > 0 {
			podDisruptionBudget["maxUnavailable"] = pdb.GetMaxUnavailable()
		}
		if pdb.GetMinAvailable() > 0 {
			podDisruptionBudget["minAvailable"] = pdb.GetMinAvailable()
		}
		values["podDisruptionBudget"] = podDisruptionBudget
	}

	// ---- escape hatch (merged LAST, helm -f semantics) --------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	return values, nil
}

// renderValkeyConfig renders the typed config block into the chart's single
// valkeyConfig string, deterministically ordered: appendonly, save points
// (or the disable directive), maxmemory, maxmemory-policy, then
// extra_directives verbatim. Returns "" (no key rendered) when the block is
// absent or renders nothing. The line order and joining MUST stay identical
// to the Terraform module's valkey_config local.
func renderValkeyConfig(config *kubernetesvalkeyv1alpha1.KubernetesValkeyConfig) string {
	if config == nil {
		return ""
	}
	var lines []string
	if config.GetAppendOnly() {
		lines = append(lines, "appendonly yes")
	}
	for _, savePoint := range config.GetRdbSavePoints() {
		lines = append(lines, "save "+savePoint)
	}
	if config.GetSnapshotsDisabled() {
		lines = append(lines, `save ""`)
	}
	if config.GetMaxMemory() != "" {
		lines = append(lines, "maxmemory "+config.GetMaxMemory())
	}
	if config.GetMaxMemoryPolicy() != "" {
		lines = append(lines, "maxmemory-policy "+config.GetMaxMemoryPolicy())
	}
	if extra := strings.TrimSpace(config.GetExtraDirectives()); extra != "" {
		lines = append(lines, extra)
	}
	return strings.Join(lines, "\n")
}

// aclUsersMap renders the declared ACL users into the chart's aclUsers map
// ({username: {permissions: ...}}) — permissions only, never passwords
// (those live in the module-materialized Secret the chart consumes via
// usersExistingSecret).
func aclUsersMap(auth *kubernetesvalkeyv1alpha1.KubernetesValkeyAuth) map[string]interface{} {
	users := map[string]interface{}{}
	for _, user := range auth.GetUsers() {
		permissions := user.GetPermissions()
		if permissions == "" {
			permissions = defaultAclPermissions
		}
		users[user.GetName()] = map[string]interface{}{"permissions": permissions}
	}
	return users
}

// tcpSocketProbe is the probe the chart renders from customProbe under
// mutual TLS: a TCP connect to the container port named `tcp`, which needs
// no client certificate. A fresh map per probe, so a helm_values merge into
// one probe never reaches the others.
func tcpSocketProbe() map[string]interface{} {
	return map[string]interface{}{
		"tcpSocket": map[string]interface{}{"port": "tcp"},
	}
}
