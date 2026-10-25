package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// The configuration the module derives from the spec, in the exact shapes
// upstream's installer writes (cmd/ate-setup/internal/steps): the API
// server's environment ConfigMap and Secret, its configuration and
// authentication files, the credential provider's namespace policy, and the
// OpenTelemetry ConfigMap.
// Upstream writes the authentication file and the policy only when absent,
// so a later change silently does nothing; the module reconciles both from
// the spec, and rolls the workloads that read them (configHash).

// The ConfigMap and Secret names upstream's manifests read.
const (
	configMapAPIEnv         = "ate-api-server-envvars"
	secretAPIEnv            = "ate-api-server-secret-envvars"
	configMapAPIConfig      = "ate-api-config"
	configMapAuthentication = "ate-api-authentication"
	configMapOtel           = "ate-otel-config"
	configMapPolicy         = "k8s-credential-provider-namespace-policy"
	configMapPostgres       = "postgres-config"
	secretPostgresServerCA  = "postgres-server-ca"
)

// The bundled database's fixed login users and host
// (pkg/postgressetup, cmd/ate-setup/internal/steps/postgres.go).
const (
	bundledReadWriteUser = "substrate_readwrite_user"
	bundledOwnerUser     = "substrate_owner_user"
	bundledPostgresHost  = "postgres.ate-system.svc:5432"
	size10PoolParams     = "&pool_max_conns=64&pool_min_conns=4"
)

// bundledDSN is the connection string for one bundled login: TLS verified
// against the service-DNS bundle, the client certificate the pod-certificate
// controller issues for that user.
func bundledDSN(user string) string {
	bundle := "/run/postgres.podcert.ate.dev/" + user + ".pem"
	return fmt.Sprintf("postgresql://%s@%s/atepg?sslmode=verify-full&sslrootcert=/run/servicedns.podcert.ate.dev/trust-bundle.pem&sslcert=%s&sslkey=%s", user, bundledPostgresHost, bundle, bundle)
}

// cloudSQLDSN logs in through the proxy as the API server's Google service
// account: passwordless, the proxy injects the IAM token.
func cloudSQLDSN(gsa string) string {
	return fmt.Sprintf("user=%s host=127.0.0.1 port=5432 dbname=atepg sslmode=disable", strings.TrimSuffix(gsa, ".gserviceaccount.com"))
}

// apiEnvConfigMap is ate-api-server-envvars: the roles, the pool size, the
// actor JWT issuer, the snapshot location and transfer service account, and
// with Cloud SQL the proxy's settings. The transfer service account is
// written empty off GCS, as upstream writes it, because the object-store
// sidecar names the key in its arguments.
func apiEnvConfigMap(s *settings) map[string]string {
	data := map[string]string{
		"ATE_API_POSTGRES_READ_WRITE_ROLE":          s.database.roleReadWrite,
		"ATE_API_POSTGRES_OWNER_ROLE":               s.database.roleOwner,
		"ATE_API_ACTOR_JWT_ISSUER":                  s.apiServer.actorIssuer,
		"ATE_API_SNAPSHOT_LOCATION":                 s.storage.location(),
		"ATE_API_SNAPSHOT_TRANSFER_SERVICE_ACCOUNT": s.transferGSA,
	}
	if s.database.poolMaxConns > 0 {
		data["ATE_API_POSTGRES_POOL_MAX_CONNS"] = strconv.Itoa(int(s.database.poolMaxConns))
	}
	if s.database.kind == "cloudsql" {
		data["ATE_API_POSTGRES_CLOUDSQL_INSTANCE"] = s.database.cloudSQLInstance
		data["CSQL_PROXY_AUTO_IAM_AUTHN"] = strconv.FormatBool(s.database.cloudSQLIAM)
		data["CSQL_PROXY_PORT"] = "5432"
		data["CSQL_PROXY_HEALTH_CHECK"] = "true"
		data["CSQL_PROXY_HTTP_ADDRESS"] = "0.0.0.0"
		data["CSQL_PROXY_HTTP_PORT"] = "9801"
		data["CSQL_PROXY_STRUCTURED_LOGS"] = "true"
		switch s.database.cloudSQLIPType {
		case "private":
			data["CSQL_PROXY_PRIVATE_IP"] = "true"
		case "psc":
			data["CSQL_PROXY_PSC"] = "true"
		}
	}
	return data
}

// apiEnvSecret is ate-api-server-secret-envvars: the connection strings the
// module composes (bundled, Cloud SQL with IAM) and the schema. Connection
// strings you supply stay in your Secret and reach the API server as
// secretKeyRef environment entries, which override this Secret.
func apiEnvSecret(s *settings) map[string]string {
	data := map[string]string{"ATE_API_POSTGRES_SCHEMA": s.database.schema}
	switch {
	case s.database.kind == "bundled":
		readWrite := bundledDSN(bundledReadWriteUser)
		if s.size10 {
			readWrite += size10PoolParams
		}
		data["ATE_API_POSTGRES_READ_WRITE_CONNECTION_STRING"] = readWrite
		data["ATE_API_POSTGRES_OWNER_CONNECTION_STRING"] = bundledDSN(bundledOwnerUser)
	case s.database.kind == "cloudsql" && s.database.cloudSQLIAM:
		dsn := cloudSQLDSN(s.apiServerGSA)
		data["ATE_API_POSTGRES_READ_WRITE_CONNECTION_STRING"] = dsn
		data["ATE_API_POSTGRES_OWNER_CONNECTION_STRING"] = dsn
	}
	return data
}

// apiIssuer is the issuer the API server trusts for Kubernetes tokens: GKE's
// own for the cluster, or the one the generic platform names.
func apiIssuer(s *settings) string {
	if s.gke {
		return "https://container.googleapis.com/v1/" + s.gkeClusterID
	}
	return s.serviceAccountIssuer
}

// apiConfigText is the API server's configuration file (apiconfig.textproto,
// upstream's ateconfigpb.APIConfig in the protobuf text format): its log
// level. The server reads the file as it changes, so no rollout follows a
// change. Upstream marshals it with prototext, whose spacing is deliberately
// unstable; this is its canonical form.
func apiConfigText(s *settings) string {
	return fmt.Sprintf("logging: {\n  level: %q\n}\n", s.apiServer.logLevel)
}

// authenticationYAML is authentication.yaml, byte for byte upstream's
// buildAuthenticationConfig: one JWT provider for Kubernetes tokens with the
// API's audience, verified against the cluster's own CA and token for the
// two in-cluster issuer forms.
func authenticationYAML(s *settings) string {
	issuer := apiIssuer(s)
	config := fmt.Sprintf("jwtProviders:\n- name: kubernetes\n  issuer: %s\n  audiences: [%s]\n", issuer, vars.APIServerHost)
	if issuer == inClusterIssuer || issuer == inClusterIssuer+".cluster.local" {
		config += "  certificateAuthorityFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt\n" +
			"  discoveryTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token\n"
	}
	return strings.TrimRight(config, "\n")
}

// policyYAML is the credential provider's namespace policy: which
// namespaces' Secrets each atespace may resolve (upstream's
// namespace-policy.yaml format). No grants resolves nothing.
func policyYAML(grants []namespaceGrant) string {
	if len(grants) == 0 {
		return "policies: []\n"
	}
	var b strings.Builder
	b.WriteString("policies:\n")
	for _, g := range grants {
		fmt.Fprintf(&b, "- atespace: %s\n  allowedNamespaces:\n", jsonQuote(g.atespace))
		for _, ns := range g.namespaces {
			fmt.Fprintf(&b, "  - %s\n", jsonQuote(ns))
		}
	}
	return b.String()
}

// otelData is ate-otel-config's data: upstream's endpoint unless the spec
// names one or the platform has no default collector, plus every exporter,
// sampler and interval setting the spec gives. On a generic cluster with no
// endpoint nothing is pushed: traces and logs go nowhere, metrics stay on
// each component's Prometheus endpoint, so no component dials a collector
// that is not there. Worker pods take their own exporter selection
// (ATE_WORKER_OTEL_*_EXPORTER, which the controller hands them) and export
// through their node agent, which forwards only to an endpoint it has.
func otelData(s *settings, upstream map[string]string) map[string]string {
	data := map[string]string{}
	for k, v := range upstream {
		data[k] = v
	}
	t := s.telemetry
	pushOff := !s.gke && t.endpoint == ""
	switch {
	case t.endpoint != "":
		data["OTEL_EXPORTER_OTLP_ENDPOINT"] = t.endpoint
	case pushOff:
		delete(data, "OTEL_EXPORTER_OTLP_ENDPOINT")
	}
	traces, metrics := t.traces, t.metrics
	if pushOff && len(traces) == 0 {
		traces = []string{"none"}
	}
	if pushOff && len(metrics) == 0 {
		metrics = []string{"prometheus"}
	}
	if len(traces) > 0 {
		data["OTEL_TRACES_EXPORTER"] = strings.Join(traces, ",")
	}
	if len(metrics) > 0 {
		data["OTEL_METRICS_EXPORTER"] = strings.Join(metrics, ",")
	}
	if len(t.logs) > 0 {
		data["OTEL_LOGS_EXPORTER"] = strings.Join(t.logs, ",")
	}
	if len(t.workerTraces) > 0 {
		data["ATE_WORKER_OTEL_TRACES_EXPORTER"] = strings.Join(t.workerTraces, ",")
	}
	if len(t.workerMetrics) > 0 {
		data["ATE_WORKER_OTEL_METRICS_EXPORTER"] = strings.Join(t.workerMetrics, ",")
	}
	if len(t.workerLogs) > 0 {
		data["ATE_WORKER_OTEL_LOGS_EXPORTER"] = strings.Join(t.workerLogs, ",")
	}
	if t.sampler != "" {
		data["OTEL_TRACES_SAMPLER"] = t.sampler
	}
	if t.samplerRatio != nil {
		data["OTEL_TRACES_SAMPLER_ARG"] = strconv.FormatFloat(*t.samplerRatio, 'f', -1, 64)
	}
	if t.exportInterval > 0 {
		data["OTEL_METRIC_EXPORT_INTERVAL"] = strconv.Itoa(int(t.exportInterval))
	}
	if t.exportTimeout > 0 {
		data["OTEL_METRIC_EXPORT_TIMEOUT"] = strconv.Itoa(int(t.exportTimeout))
	}
	return data
}

// jsonQuote quotes a string as JSON does, which YAML reads as a plain
// string; the Terraform twin quotes with jsonencode.
func jsonQuote(value string) string {
	quoted, _ := json.Marshal(value)
	return string(quoted)
}

// sortedLines renders a map as "k=v\n" lines sorted by key.
func sortedLines(data map[string]string) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k + "=" + data[k] + "\n")
	}
	return b.String()
}

// envHash is upstream's ate.dev/env-hash (cmd/ate-setup/internal/steps
// apienv.go): the digest of the API server's environment ConfigMap and
// Secret, so a change to either rolls it.
func envHash(configMap, secret map[string]string) string {
	sum := sha256.Sum256([]byte("configmap\n" + sortedLines(configMap) + "secret\n" + sortedLines(secret)))
	return hex.EncodeToString(sum[:])
}

// configHash is the digest of the named ConfigMaps a workload reads only at
// start, in the order given, so a change to any of them rolls it.
func configHash(names []string, configMaps map[string]map[string]string) string {
	var b strings.Builder
	for _, name := range names {
		b.WriteString("configmap " + name + "\n" + sortedLines(configMaps[name]))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// configReaders names, per workload, the spec-derived ConfigMaps it reads
// only at start (envFrom or a mounted file the binary never re-reads).
// A composition test holds each entry to the workload's own references.
var configReaders = map[string][]string{
	vars.APIServerName:          {configMapAuthentication, configMapOtel},
	vars.ControllerName:         {configMapOtel},
	vars.NodeAgentName:          {configMapOtel},
	vars.RouterName:             {configMapOtel},
	vars.EgressName:             {configMapOtel},
	vars.CredentialProviderName: {configMapPolicy},
	vars.PostgresName:           {configMapPostgres, vars.PostgresSetupConfigMap},
}
