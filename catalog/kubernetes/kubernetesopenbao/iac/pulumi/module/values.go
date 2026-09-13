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

	switch locals.Mode {
	case modeDev:
		// Dev mode: in-memory, auto-unsealed, root token "root" (the
		// chart default — deliberately not configurable; the whole arm
		// is a documented never-production sandbox).
		server["dev"] = map[string]interface{}{"enabled": true}
	case modeStandalone:
		server["standalone"] = map[string]interface{}{
			"enabled": true,
			"config":  locals.BaoConfigHcl,
		}
	case modeHa:
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

	// Data volume. The chart consumes dataStorage only in standalone and
	// ha+raft (dev is in-memory) — rendered UNCONDITIONALLY (matching the
	// Terraform twin) for explicitness; the chart's own mode gates
	// decide.
	dataStorage := map[string]interface{}{"enabled": true}
	if ds := spec.GetServer().GetDataStorage(); ds != nil {
		if ds.Size != nil && ds.GetSize() != "" {
			dataStorage["size"] = ds.GetSize()
		}
		if ds.GetStorageClass().GetValue() != "" {
			dataStorage["storageClass"] = ds.GetStorageClass().GetValue()
		}
	}
	server["dataStorage"] = dataStorage

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

	// Seal credentials: identifiers ride plain env; credential material
	// rides the module-owned Secret (created before the release).
	if env := sealPlainEnv(spec); env != nil {
		server["extraEnvironmentVars"] = stringMapToInterface(env)
	}
	if data := sealSecretData(spec); data != nil {
		secretEnv := make([]interface{}, 0, len(data))
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		// Deterministic order — map iteration would diff the rendered
		// values between runs.
		sort.Strings(keys)
		for _, envName := range keys {
			secretEnv = append(secretEnv, map[string]interface{}{
				"envName":    envName,
				"secretName": locals.SealCredentialsSecretName,
				"secretKey":  envName,
			})
		}
		server["extraSecretEnvironmentVars"] = secretEnv
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
