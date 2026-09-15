package module

import (
	"sort"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values. The
// Terraform twin (locals.tf `helm_values`) renders the byte-identical
// document — keep them in lockstep.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec

	// ------------------------------ server --------------------------------
	server := map[string]interface{}{}

	// THE CHART IS TOLD A MODE, NEVER AN ENGINE. It offers `dev`,
	// `standalone`, and `ha`, plus a raw configuration string per mode;
	// this module drives `ha` for every storage engine and puts the
	// engine's stanza in the string it already writes (bao_config.go).
	// Raft: `ha.raft.enabled` on, the string in `ha.raft.config`, one data
	// volume per replica. PostgreSQL: `ha.raft.enabled` off, the string in
	// `ha.config` (the path the chart reads when Raft is off), no data
	// volume. The chart's `standalone` mode (file storage) is never driven.
	switch {
	case locals.Dev:
		// Dev mode: in-memory, auto-unsealed, root token "root" (the
		// chart default — deliberately not configurable; the whole arm
		// is a documented never-production sandbox).
		server["dev"] = map[string]interface{}{"enabled": true}
	case locals.Storage == storagePostgresql:
		server["ha"] = map[string]interface{}{
			"enabled":  true,
			"replicas": locals.Replicas,
			"raft":     map[string]interface{}{"enabled": false},
			"config":   locals.BaoConfigHcl,
			// On PostgreSQL the database holds the data and one live node
			// is availability, so every other replica may be disrupted.
			"disruptionBudget": disruptionBudgetBlock(locals.Replicas, locals.Replicas-1),
		}
	default:
		server["ha"] = map[string]interface{}{
			"enabled":  true,
			"replicas": locals.Replicas,
			"raft": map[string]interface{}{
				"enabled": true,
				// Stable, human-readable Raft node IDs = pod names
				// (without this the server generates a GUID — persisted
				// on the data PVC, but opaque in every peer listing).
				"setNodeId": true,
				"config":    locals.BaoConfigHcl,
			},
			// On Raft the chart's own quorum arithmetic ((n/2)-1) is the
			// right budget; only the one-replica case is overridden.
			"disruptionBudget": disruptionBudgetBlock(locals.Replicas, 0),
		}
	}

	if s := spec.GetServer(); s != nil {
		if r := resourcesBlock(s.GetResources()); r != nil {
			server["resources"] = r
		}
		if s.LogLevel != nil && s.GetLogLevel() != "" {
			server["logLevel"] = s.GetLogLevel()
		}
		if s.LogFormat != nil && s.GetLogFormat() != "" {
			server["logFormat"] = s.GetLogFormat()
		}
		if sched := s.GetScheduling(); sched != nil {
			if len(sched.GetNodeSelector()) > 0 {
				server["nodeSelector"] = stringMapToInterface(sched.GetNodeSelector())
			}
			if len(sched.GetTolerations()) > 0 {
				server["tolerations"] = tolerationsSlice(sched.GetTolerations())
			}
		}

		if as := s.GetAuditStorage(); as != nil {
			auditStorage := map[string]interface{}{"enabled": true}
			if as.Size != nil && as.GetSize() != "" {
				auditStorage["size"] = as.GetSize()
			}
			if as.GetStorageClass().GetValue() != "" {
				auditStorage["storageClass"] = as.GetStorageClass().GetValue()
			}
			server["auditStorage"] = auditStorage
		}
	}

	// Data volume: Raft's alone. The spec carries it inside `server.raft`
	// because no other engine has one; the chart claims it only when
	// `ha.raft.enabled` is on, so PostgreSQL is told `enabled: false`
	// explicitly (the reader of the rendered values should not have to
	// know the chart's gate) and dev renders no key at all (in-memory).
	switch {
	case locals.Dev:
	case locals.Storage == storagePostgresql:
		server["dataStorage"] = map[string]interface{}{"enabled": false}
	default:
		dataStorage := map[string]interface{}{"enabled": true}
		if ds := spec.GetServer().GetRaft().GetDataStorage(); ds != nil {
			if ds.Size != nil && ds.GetSize() != "" {
				dataStorage["size"] = ds.GetSize()
			}
			if ds.GetStorageClass().GetValue() != "" {
				dataStorage["storageClass"] = ds.GetStorageClass().GetValue()
			}
		}
		server["dataStorage"] = dataStorage
	}

	// TLS: mount the certificate Secret where the synthesized listener
	// config expects it. `global.tlsDisable` alone changes ONLY probe
	// schemes and derived URLs — the listener lines in bao_config.go are
	// the other half of the composite switch.
	if locals.TlsEnabled {
		server["volumes"] = []interface{}{
			map[string]interface{}{
				"name": "tls",
				"secret": map[string]interface{}{
					"secretName": locals.TlsSecretName,
				},
			},
		}
		server["volumeMounts"] = []interface{}{
			map[string]interface{}{
				"name":      "tls",
				"mountPath": vars.TlsMountPath,
				"readOnly":  true,
			},
		}
	}

	// The two environment seams the chart offers, each fed by every
	// consumer that needs it. Plain: the seal's identifiers and the
	// PostgreSQL connection facts. Secret-backed: the seal's credential
	// material from the module-owned Secret, and the PostgreSQL password
	// from the REFERENCED Secret (the module never holds it). Both sorted
	// by variable name — map iteration would diff the rendered values
	// between runs.
	plainEnv := map[string]string{}
	for k, v := range sealPlainEnv(spec) {
		plainEnv[k] = v
	}
	for k, v := range locals.PgPlainEnv {
		plainEnv[k] = v
	}
	if len(plainEnv) > 0 {
		server["extraEnvironmentVars"] = stringMapToInterface(plainEnv)
	}
	secretEnv := []secretEnvRef{}
	for envName := range sealSecretData(spec) {
		secretEnv = append(secretEnv, secretEnvRef{envName: envName, secretName: locals.SealCredentialsSecretName, secretKey: envName})
	}
	if locals.PgPasswordSecretName != "" {
		secretEnv = append(secretEnv, secretEnvRef{envName: envPgPassword, secretName: locals.PgPasswordSecretName, secretKey: locals.PgPasswordSecretKey})
	}
	if len(secretEnv) > 0 {
		sort.Slice(secretEnv, func(i, j int) bool { return secretEnv[i].envName < secretEnv[j].envName })
		rendered := make([]interface{}, 0, len(secretEnv))
		for _, ref := range secretEnv {
			rendered = append(rendered, map[string]interface{}{
				"envName":    ref.envName,
				"secretName": ref.secretName,
				"secretKey":  ref.secretKey,
			})
		}
		server["extraSecretEnvironmentVars"] = rendered
	}

	// ServiceAccount identity (the workload-identity seam) + the
	// Kubernetes-auth TokenReview binding. The GCP seal arm's declared
	// workload identity contributes its own annotation (the spec field
	// promises exactly this); explicit service_account.annotations win
	// on conflict. NOTE dev mode drops SA annotations (chart behavior —
	// taught on the spec field).
	saAnnotations := map[string]string{}
	if email := spec.GetAutoUnseal().GetGcpKms().GetWorkloadIdentityServiceAccount().GetValue(); email != "" {
		saAnnotations["iam.gke.io/gcp-service-account"] = email
	}
	for k, v := range spec.GetServiceAccount().GetAnnotations() {
		saAnnotations[k] = v
	}
	if len(saAnnotations) > 0 {
		server["serviceAccount"] = map[string]interface{}{
			"annotations": stringMapToInterface(saAnnotations),
		}
	}
	if sa := spec.GetServiceAccount(); sa != nil && sa.AuthDelegatorEnabled != nil {
		server["authDelegator"] = map[string]interface{}{
			"enabled": sa.GetAuthDelegatorEnabled(),
		}
	}

	if spec.GetNetworkPolicyEnabled() {
		server["networkPolicy"] = map[string]interface{}{"enabled": true}
	}

	// ------------------------------ top level -----------------------------
	uiEnabled := true
	if spec.UiEnabled != nil {
		uiEnabled = spec.GetUiEnabled()
	}

	values := map[string]interface{}{
		"global": map[string]interface{}{
			"tlsDisable": !locals.TlsEnabled,
		},
		"server": server,
		// The ui Service toggle; the listener-side `ui = true` lives in
		// the synthesized config — one spec field drives both.
		"ui": map[string]interface{}{"enabled": uiEnabled},
		// THE INJECTOR IS OPT-IN — a deliberate divergence from the
		// chart default (which installs a cluster-wide mutating webhook
		// on every install); rendered explicitly either way.
		"injector": injectorBlock(locals),
	}

	if spec.GetMetrics().GetServiceMonitorEnabled() {
		values["serverTelemetry"] = map[string]interface{}{
			"serviceMonitor": map[string]interface{}{"enabled": true},
		}
	}

	// ------------------------- helm_values escape --------------------------
	// Merged LAST with Helm -f semantics; fullnameOverride is re-pinned
	// after the merge — resource naming is the module's identity
	// contract and cannot be overridden.
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "helm_values is not valid YAML")
		}
		values = mergeMaps(values, overrides)
	}
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// secretEnvRef is one entry of the chart's extraSecretEnvironmentVars: an
// environment variable the server reads, sourced from a key of a Secret.
// The seal's entries point at the module-owned credentials Secret; the
// PostgreSQL password points at the Secret the manifest referenced.
type secretEnvRef struct {
	envName    string
	secretName string
	secretKey  string
}

// disruptionBudgetBlock renders the chart's `ha.disruptionBudget` values
// from the declared replica count. The chart enables its
// PodDisruptionBudget by default and hard-codes `maxUnavailable: 0` for
// one replica — a budget that permits zero voluntary disruptions blocks
// every node drain and cluster scale-down forever while protecting
// nothing, so at one replica the budget is disabled. Above one,
// maxUnavailable > 0 overrides the chart's arithmetic (PostgreSQL: all
// but one, the database holds the data); 0 leaves the chart's own
// quorum formula in charge (Raft). The Terraform twin renders the same
// object.
func disruptionBudgetBlock(replicas int, maxUnavailable int) map[string]interface{} {
	if replicas <= 1 {
		return map[string]interface{}{"enabled": false}
	}
	block := map[string]interface{}{"enabled": true}
	if maxUnavailable > 0 {
		block["maxUnavailable"] = maxUnavailable
	}
	return block
}

// injectorBlock renders the injector values (explicitly disabled unless
// opted in — see the spec's blast-radius comment).
func injectorBlock(locals *Locals) map[string]interface{} {
	inj := locals.Spec.GetInjector()
	if inj == nil || !inj.GetEnabled() {
		return map[string]interface{}{"enabled": false}
	}
	block := map[string]interface{}{"enabled": true}
	if inj.Replicas != nil {
		block["replicas"] = int(inj.GetReplicas())
	}
	if inj.FailurePolicy != nil && inj.GetFailurePolicy() != "" {
		block["webhook"] = map[string]interface{}{
			"failurePolicy": inj.GetFailurePolicy(),
		}
	}
	if r := resourcesBlock(inj.GetResources()); r != nil {
		block["resources"] = r
	}
	return block
}
