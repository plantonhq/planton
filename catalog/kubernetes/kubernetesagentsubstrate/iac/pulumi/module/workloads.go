package module

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// The patches applied onto upstream's own objects. Each names the object it
// expects by upstream's fixed name and refuses, rather than guesses, when the
// pinned files no longer have that shape. A setting is written only when it
// differs from upstream's own value, so an empty spec section installs
// upstream as published. The Terraform twin applies the same patches in
// locals.tf.

// kindName is the ownership label's value on every workload the module
// applies.
var kindName = catalogkind.CatalogKind_KubernetesAgentSubstrate.String()

const ownerLabel = failure.WorkloadOwnerLabel

// flag is one command-line setting, written as --name=value.
type flag struct {
	name  string
	value string
}

// flagList collects flags whose condition holds, in order.
type flagList []flag

func (l *flagList) add(condition bool, name, value string) {
	if condition {
		*l = append(*l, flag{name: name, value: value})
	}
}

// applyFlags writes typed settings onto upstream's own argument list. A flag
// upstream already passes keeps its position and takes the new value; every
// other flag is appended in the order given. Upstream's own arguments that no
// setting names are kept verbatim.
func applyFlags(container document, flags []flag) {
	if len(flags) == 0 {
		return
	}
	byName := make(map[string]string, len(flags))
	for _, f := range flags {
		byName[f.name] = f.value
	}
	used := map[string]bool{}
	upstream := stringList(container["args"])
	out := make([]string, 0, len(upstream)+len(flags))
	for _, arg := range upstream {
		name := strings.TrimPrefix(strings.SplitN(arg, "=", 2)[0], "--")
		if value, ok := byName[name]; ok && strings.HasPrefix(arg, "--") {
			out = append(out, "--"+name+"="+value)
			used[name] = true
			continue
		}
		out = append(out, arg)
	}
	for _, f := range flags {
		if !used[f.name] {
			out = append(out, "--"+f.name+"="+f.value)
		}
	}
	container["args"] = toInterfaces(out)
}

func itoa(i int32) string { return strconv.Itoa(int(i)) }

func formatFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

// setResources writes a container's resources when the spec gives any.
func setResources(container document, r *kubernetesprovider.ContainerResources) {
	if r == nil {
		return
	}
	resources := document{}
	for field, cm := range map[string]*kubernetesprovider.CpuMemory{"requests": r.GetRequests(), "limits": r.GetLimits()} {
		values := document{}
		if cm.GetCpu() != "" {
			values["cpu"] = cm.GetCpu()
		}
		if cm.GetMemory() != "" {
			values["memory"] = cm.GetMemory()
		}
		if len(values) > 0 {
			resources[field] = values
		}
	}
	if len(resources) > 0 {
		container["resources"] = resources
	}
}

// setEnv sets (or replaces) one environment entry.
func setEnv(container document, entry document) {
	env, _ := container["env"].([]interface{})
	for i, e := range env {
		if m, ok := e.(map[string]interface{}); ok && m["name"] == entry["name"] {
			env[i] = entry
			container["env"] = env
			return
		}
	}
	container["env"] = append(env, entry)
}

func envValue(name, value string) document { return document{"name": name, "value": value} }

func envSecret(name string, key *secretKey) document {
	return document{"name": name, "valueFrom": document{"secretKeyRef": document{"name": key.name, "key": key.key}}}
}

// storageEnv points an objectstore-plugin sidecar at the chosen store
// (pkg/objectstorage reads ATE_STORAGE_BACKEND and the AWS SDK's own
// variables; GCS uses the pod's Workload Identity).
func storageEnv(sidecar document, st storageSettings) {
	if !st.isS3() {
		setEnv(sidecar, envValue("ATE_STORAGE_BACKEND", "gcs"))
		return
	}
	setEnv(sidecar, envValue("ATE_STORAGE_BACKEND", "s3"))
	setEnv(sidecar, envValue("AWS_REGION", st.s3Region))
	setEnv(sidecar, envValue("AWS_ENDPOINT_URL", st.s3Endpoint))
	if st.s3PathStyle {
		setEnv(sidecar, envValue("AWS_S3_USE_PATH_STYLE", "true"))
	}
	setEnv(sidecar, envSecret("AWS_ACCESS_KEY_ID", &secretKey{name: st.s3Secret, key: st.s3IDKey}))
	setEnv(sidecar, envSecret("AWS_SECRET_ACCESS_KEY", &secretKey{name: st.s3Secret, key: st.s3SecretKey}))
}

// patchPodCertificates writes the pod-certificate controller's settings. The
// manifest reads --workers-per-signer from WORKERS_PER_SIGNER.
func patchPodCertificates(d document, s *settings) error {
	container, err := containerNamed(podSpec(d), "controller", false)
	if err != nil {
		return err
	}
	if s.podCerts.workersPerSigner != 1 {
		setEnv(container, envValue("WORKERS_PER_SIGNER", itoa(s.podCerts.workersPerSigner)))
	}
	var l flagList
	l.add(s.podCerts.qps > 0, "kube-api-qps", formatFloat(s.podCerts.qps))
	l.add(s.podCerts.burst > 0, "kube-api-burst", itoa(s.podCerts.burst))
	applyFlags(container, l)
	setResources(container, s.podCerts.resources)
	return nil
}

// applySize10PostgresConfig replaces the bundled database's postgresql.conf
// with upstream's size10 one.
func applySize10PostgresConfig(d document, patchText string) error {
	deleted, err := applyPatchText(d, patchText)
	if err != nil || deleted {
		return errors.Wrap(err, "applying the size10 postgresql.conf")
	}
	return nil
}

// patchPostgres writes the bundled database's volume, resources and setup.
func patchPostgres(d document, s *settings) error {
	spec := child(d, "spec")
	claims, _ := spec["volumeClaimTemplates"].([]interface{})
	claim := namedItemInMeta(claims, "data")
	if claim == nil {
		return errors.Errorf("Agent Substrate %s's database StatefulSet has no data volume claim; the module's pin needs a catalog fix", vars.Release)
	}
	claimSpec := child(claim, "spec")
	if s.database.storageClass != "" {
		claimSpec["storageClassName"] = s.database.storageClass
	}
	child(child(claimSpec, "resources"), "requests")["storage"] = s.database.volumeSize
	labelWith(child(claim, "metadata"), map[string]string{vars.MinorLabel: minorOf(s.version)})

	container, err := containerNamed(podSpec(d), vars.PostgresName, false)
	if err != nil {
		return err
	}
	if s.size10 {
		resources := child(container, "resources")
		delete(child(resources, "limits"), "cpu")
		child(resources, "requests")["cpu"] = "80"
		child(resources, "requests")["memory"] = "140Gi"
		child(resources, "limits")["memory"] = "140Gi"
	}
	setResources(container, s.database.resources)

	// The installer runs upstream's setup script by exec after every deploy.
	// The database's own start hook runs it instead, as the local superuser
	// (the only login local trust admits), with upstream's exact variables, in
	// one transaction; the script is idempotent.
	exec := child(child(child(container, "lifecycle"), "postStart"), "exec")
	command := stringList(exec["command"])
	if len(command) != 3 {
		return errors.Errorf("Agent Substrate %s's database start hook changed shape; the module's pin needs a catalog fix", vars.Release)
	}
	command[2] = strings.TrimRight(command[2], "\n") + "\n" + setupCommand + "\n"
	exec["command"] = toInterfaces(command)
	mounts, _ := container["volumeMounts"].([]interface{})
	container["volumeMounts"] = append(mounts, document{"name": vars.PostgresSetupConfigMap, "mountPath": vars.PostgresSetupMountPath, "readOnly": true})
	volumes, _ := podSpec(d)["volumes"].([]interface{})
	podSpec(d)["volumes"] = append(volumes, document{"name": vars.PostgresSetupConfigMap, "configMap": document{"name": vars.PostgresSetupConfigMap}})
	return nil
}

// setupCommand is upstream's setup invocation (postgressetup.PSQLArgs with
// the default identities and empty passwords), reading the mounted script.
const setupCommand = `psql --no-psqlrc --set=ON_ERROR_STOP=1 --single-transaction --username "${POSTGRES_USER}" --dbname atepg --set=substrate_schema=substrate --set=substrate_owner_role=substrate_owner --set=substrate_owner_user=substrate_owner_user --set=substrate_owner_password= --set=substrate_readwrite_role=substrate_readwrite --set=substrate_readwrite_user=substrate_readwrite_user --set=substrate_readwrite_password= --file=/etc/postgres-setup/setup.sql`

// namedItemInMeta finds a list item by metadata.name (claim templates).
func namedItemInMeta(list []interface{}, name string) document {
	for _, item := range list {
		if m, ok := item.(map[string]interface{}); ok && str(m, "metadata", "name") == name {
			return m
		}
	}
	return nil
}

// patchSystemDocument dispatches the system phase's patches by object.
func patchSystemDocument(d document, s *settings) error {
	name := str(d, "metadata", "name")
	switch str(d, "kind") {
	case "Deployment":
		switch name {
		case vars.APIServerName:
			return patchAPIServer(d, s)
		case vars.IdentityServerName:
			return patchIdentityServer(d, s)
		case vars.ControllerName:
			return patchController(d, s)
		case vars.RouterName:
			return patchRouter(d, s)
		case vars.EgressName:
			return patchEgress(d, s)
		case vars.CredentialProviderName:
			return patchCredentialProvider(d, s)
		}
	case "DaemonSet":
		if name == vars.NodeAgentName {
			return patchNodeAgent(d, s)
		}
	case "ServiceAccount":
		switch {
		case name == vars.APIServerName && s.apiServerGSA != "":
			child(child(d, "metadata"), "annotations")["iam.gke.io/gcp-service-account"] = s.apiServerGSA
		case name == vars.NodeAgentName && s.nodeAgentGSA != "":
			child(child(d, "metadata"), "annotations")["iam.gke.io/gcp-service-account"] = s.nodeAgentGSA
		}
	case "ConfigMap":
		switch {
		case name == vars.EgressName && s.egress.extprocService != "":
			return spliceExtProc(d, s)
		case name == agentgatewayEgressConfig:
			return spliceAgentgatewayProviders(d, s)
		}
	}
	return nil
}

func patchAPIServer(d document, s *settings) error {
	a := s.apiServer
	child(d, "spec")["replicas"] = int64(a.replicas)
	container, err := containerNamed(podSpec(d), vars.APIServerName, false)
	if err != nil {
		return err
	}
	// The bootstrap owners are always written, as upstream's installer
	// writes them: the API server refuses to start without one. The log level
	// is the configuration file's (apiConfigText).
	var l flagList
	l.add(true, "authz-bootstrap-owners", strings.Join(a.bootstrapOwners, ","))
	l.add(a.drainDelay != "13s", "drain-delay", a.drainDelay)
	l.add(a.drainTimeout != "15s", "drain-timeout", a.drainTimeout)
	l.add(a.resyncInterval != "20s", "template-resync-interval", a.resyncInterval)
	applyFlags(container, l)
	setResources(container, a.resources)
	if s.database.externalPasswords {
		setEnv(container, envSecret("ATE_API_POSTGRES_READ_WRITE_CONNECTION_STRING", s.database.readWrite))
		setEnv(container, envSecret("ATE_API_POSTGRES_OWNER_CONNECTION_STRING", s.database.owner))
	}
	sidecar, err := containerNamed(podSpec(d), "objectstore-plugin", true)
	if err != nil {
		return err
	}
	storageEnv(sidecar, s.storage)
	return nil
}

// patchIdentityServer writes the identity server's settings. The issuer is
// always written, as upstream's installer writes it (setIDPServerIssuer): the
// server refuses to start without the one actor JWTs carry.
func patchIdentityServer(d document, s *settings) error {
	id := s.identity
	child(d, "spec")["replicas"] = int64(id.replicas)
	container, err := containerNamed(podSpec(d), vars.IdentityServerName, false)
	if err != nil {
		return err
	}
	var l flagList
	l.add(true, "issuer", s.apiServer.actorIssuer)
	l.add(id.logLevel != "info", "log-level", id.logLevel)
	l.add(id.drainGrace != "5s", "drain-grace", id.drainGrace)
	applyFlags(container, l)
	setResources(container, id.resources)
	return nil
}

func patchController(d document, s *settings) error {
	container, err := containerNamed(podSpec(d), vars.ControllerName, false)
	if err != nil {
		return err
	}
	var l flagList
	l.add(s.controller.logLevel != "info", "log-level", s.controller.logLevel)
	l.add(s.controller.workerSync != 2, "worker-sync-concurrency", itoa(s.controller.workerSync))
	applyFlags(container, l)
	setResources(container, s.controller.resources)
	return nil
}

func patchNodeAgent(d document, s *settings) error {
	n := s.nodeAgent
	spec := podSpec(d)
	container, err := containerNamed(spec, vars.NodeAgentName, false)
	if err != nil {
		return err
	}
	var l flagList
	l.add(n.logLevel != "info", "log-level", n.logLevel)
	l.add(n.statsPoll != "1m", "actor-stats-poll-interval", n.statsPoll)
	l.add(n.drainDelay != "0s", "drain-delay", n.drainDelay)
	l.add(n.drainTimeout != "5m", "drain-timeout", n.drainTimeout)
	l.add(n.gcPeriod != "5m", "image-cache-gc-period", n.gcPeriod)
	l.add(n.highPercent != 85, "image-cache-high-percent", itoa(n.highPercent))
	l.add(n.lowPercent != 80, "image-cache-low-percent", itoa(n.lowPercent))
	l.add(n.maxBytes != 0, "image-cache-max-bytes", strconv.FormatInt(n.maxBytes, 10))
	l.add(n.minAge != "2m", "image-cache-min-age", n.minAge)
	l.add(n.gcDryRun, "image-cache-gc-dry-run", "true")
	snap := n.snapshots
	l.add(snap.uploadDelay != "5m", "snapshot-upload-delay", snap.uploadDelay)
	l.add(snap.detachDelay != "30m", "snapshot-detach-delay", snap.detachDelay)
	l.add(snap.uploadTimeout != "2m", "snapshot-upload-timeout", snap.uploadTimeout)
	l.add(snap.uploadMiB != 100, "snapshot-upload-est-throughput-mib", itoa(snap.uploadMiB))
	l.add(snap.uploadWorkers != 2, "snapshot-upload-workers", itoa(snap.uploadWorkers))
	l.add(snap.detachWorkers != 2, "snapshot-detach-workers", itoa(snap.detachWorkers))
	l.add(snap.reportWorkers != 1, "snapshot-report-workers", itoa(snap.reportWorkers))
	l.add(snap.reportGiveUp != "10m", "snapshot-report-give-up-after", snap.reportGiveUp)
	l.addFileCache("sandbox-asset-cache", "sandbox-asset-fetch-timeout", n.assetCache, "1h", "5m")
	l.addFileCache("snapshot-cache", "snapshot-cache-fetch-timeout", n.snapCache, "10m", "10m")
	// An empty cache directory turns the snapshot cache off.
	l.add(n.snapCache.disabled, "snapshot-cache-dir", "")
	applyFlags(container, l)
	setResources(container, n.resources)

	if !s.gke {
		// Upstream's kind overlay: no node image credential provider off GKE,
		// so neither its two flags nor its host mounts.
		var args []string
		for _, arg := range stringList(container["args"]) {
			if !strings.HasPrefix(arg, "--image-credential-provider-") {
				args = append(args, arg)
			}
		}
		container["args"] = toInterfaces(args)
		container["volumeMounts"] = withoutItems(container["volumeMounts"], "mountPath", "/run/image-credential-provider/bin", "/run/image-credential-provider/config.yaml")
		spec["volumes"] = withoutItems(spec["volumes"], "name", "image-credential-provider-bin", "image-credential-provider-config")
	}

	sidecar, err := containerNamed(spec, "objectstore-plugin", true)
	if err != nil {
		return err
	}
	var sl flagList
	sl.add(n.stagingGiB != 8, "asset-staging-capacity", itoa(n.stagingGiB)+"Gi")
	sl.add(snap.uploadMemory != "512Mi", "upload-memory-budget", snap.uploadMemory)
	applyFlags(sidecar, sl)
	if n.stagingGiB != 8 {
		staging := namedItem(spec["volumes"], "name", "asset-staging")
		if staging == nil {
			return errors.Errorf("Agent Substrate %s's node agent has no asset-staging volume; the module's pin needs a catalog fix", vars.Release)
		}
		child(staging, "emptyDir")["sizeLimit"] = itoa(n.stagingGiB+1) + "Gi"
	}
	storageEnv(sidecar, s.storage)

	if len(n.nodeSelector) > 0 {
		selector := child(spec, "nodeSelector")
		for k, v := range n.nodeSelector {
			selector[k] = v
		}
	}
	if len(n.tolerations) > 0 {
		existing, _ := spec["tolerations"].([]interface{})
		for _, t := range n.tolerations {
			existing = append(existing, toleration(t))
		}
		spec["tolerations"] = existing
	}
	return nil
}

// addFileCache adds one node file cache's flags where they differ from
// upstream's: its age and eviction flags under the cache's prefix (upstream's
// filecachegc flags), and its fetch timeout, which upstream names apart.
func (l *flagList) addFileCache(prefix, fetchTimeoutFlag string, c fileCacheSettings, minAgeDefault, fetchTimeoutDefault string) {
	l.add(c.minAge != minAgeDefault, prefix+"-min-age", c.minAge)
	l.add(c.fetchTimeout != fetchTimeoutDefault, fetchTimeoutFlag, c.fetchTimeout)
	l.add(c.gcPeriod != "5m", prefix+"-gc-period", c.gcPeriod)
	l.add(c.maxBytes != 0, prefix+"-max-bytes", strconv.FormatInt(c.maxBytes, 10))
	l.add(c.gcDryRun, prefix+"-gc-dry-run", "true")
}

// toleration renders a WorkloadToleration as Kubernetes writes it.
func toleration(t *kubernetesprovider.WorkloadToleration) document {
	out := document{}
	if t.GetKey() != "" {
		out["key"] = t.GetKey()
	}
	if t.GetOperator() != "" {
		out["operator"] = t.GetOperator()
	}
	if t.GetValue() != "" {
		out["value"] = t.GetValue()
	}
	if t.GetEffect() != "" {
		out["effect"] = t.GetEffect()
	}
	if t.TolerationSeconds != nil {
		out["tolerationSeconds"] = t.GetTolerationSeconds()
	}
	return out
}

// withoutItems drops the list items whose key field is one of values.
func withoutItems(list interface{}, key string, values ...string) []interface{} {
	items, _ := list.([]interface{})
	drop := map[string]bool{}
	for _, v := range values {
		drop[v] = true
	}
	var out []interface{}
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			if v, _ := m[key].(string); drop[v] {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

func patchRouter(d document, s *settings) error {
	r := s.router
	child(d, "spec")["replicas"] = int64(r.replicas)
	containers, _ := podSpec(d)["containers"].([]interface{})
	if len(containers) == 0 {
		return errors.Errorf("Agent Substrate %s's router has no containers; the module's pin needs a catalog fix", vars.Release)
	}
	if r.agentgateway {
		// agentgateway replaced the router binary; its configuration is
		// upstream's static one.
		setResources(containers[0].(map[string]interface{}), r.resources)
		return nil
	}
	container, err := containerNamed(podSpec(d), vars.RouterName, false)
	if err != nil {
		return err
	}
	var l flagList
	l.add(r.logLevel != "info", "log-level", r.logLevel)
	l.add(len(r.spiffeIDs) > 0, "ingress-allowed-spiffe-ids", strings.Join(r.spiffeIDs, ","))
	l.add(r.routeTimeout != "5m", "route-timeout", r.routeTimeout)
	l.add(r.parkBudget != "5s", "parked-request-budget", r.parkBudget)
	l.add(r.parkMax != 1024, "parked-request-max", itoa(r.parkMax))
	l.add(r.parkRetryInterval != "100ms", "parked-request-retry-interval", r.parkRetryInterval)
	l.add(r.parkRetryFactor != 1.1, "parked-request-retry-factor", formatFloat(r.parkRetryFactor))
	l.add(r.parkRetryJitter != 0.1, "parked-request-retry-jitter", formatFloat(r.parkRetryJitter))
	l.add(r.extprocMax != 0, "extproc-max-requests", itoa(r.extprocMax))
	l.add(r.drainDelay != "13s", "drain-delay", r.drainDelay)
	l.add(r.drainTimeout != "", "drain-timeout", r.drainTimeout)
	applyFlags(container, l)
	setResources(container, r.resources)
	return nil
}

func patchEgress(d document, s *settings) error {
	e := s.egress
	child(d, "spec")["replicas"] = int64(e.replicas)
	if s.router.agentgateway {
		containers, _ := podSpec(d)["containers"].([]interface{})
		if len(containers) > 0 {
			setResources(containers[0].(map[string]interface{}), e.resources)
		}
		return nil
	}
	spec := podSpec(d)
	extProc, err := containerNamed(spec, "ext-proc", false)
	if err != nil {
		return err
	}
	var l flagList
	l.add(e.logLevel != "info", "log-level", e.logLevel)
	l.add(e.policyTTL != "10s", "egress-policy-cache-ttl", e.policyTTL)
	l.add(e.drainDelay != "13s", "drain-delay", e.drainDelay)
	if e.provider != "disabled" {
		// Upstream's #ATE_EGRESS_INJECT_FLAGS block.
		l.add(true, "credential-provider-name", e.providerName)
		l.add(true, "credential-provider-address", e.providerAddress)
		l.add(true, "credential-provider-ca-file", "/run/servicedns.podcert.ate.dev/trust-bundle.pem")
		l.add(true, "credential-provider-client-cert", "/run/podidentity.podcert.ate.dev/credential-bundle.pem")
		l.add(true, "credential-provider-server-name", e.providerServerName())
	}
	applyFlags(extProc, l)
	setResources(extProc, e.resources)

	minter, err := containerNamed(spec, "sdsmint", true)
	if err != nil {
		return err
	}
	var m flagList
	m.add(e.interceptTTL != "15m", "leaf-cert-ttl", e.interceptTTL)
	m.add(e.logLevel != "info", "log-level", e.logLevel)
	applyFlags(minter, m)

	if e.envoyConcurrency > 0 {
		// Upstream's JSON patch: Envoy is container 0.
		envoy, err := containerNamed(spec, "envoy", false)
		if err != nil {
			return err
		}
		envoy["args"] = append(toInterfaces(stringList(envoy["args"])), "--concurrency", itoa(e.envoyConcurrency))
	}
	return nil
}

func patchCredentialProvider(d document, s *settings) error {
	e := s.egress
	container, err := containerNamed(podSpec(d), vars.CredentialProviderName, false)
	if err != nil {
		return err
	}
	var l flagList
	l.add(e.providerLogLevel != "info", "log-level", e.providerLogLevel)
	l.add(e.providerQPS != 50, "kube-api-qps", formatFloat(e.providerQPS))
	l.add(e.providerBurst != 100, "kube-api-burst", itoa(e.providerBurst))
	l.add(e.providerDrain != "5s", "drain-grace", e.providerDrain)
	applyFlags(container, l)
	setResources(container, e.providerResources)
	return nil
}

// minorOf is "0.6" for "v0.6.0".
func minorOf(version string) string {
	parts := strings.SplitN(strings.TrimPrefix(version, "v"), ".", 3)
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}
