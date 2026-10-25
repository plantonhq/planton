package module

import (
	"strings"

	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesagentsubstratev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsubstrate/v1alpha1"
)

// settings is the spec resolved once: every optional field carries
// upstream's default, so the composition compares against upstream's own
// values and writes a flag or a field only when it differs. Terraform twin:
// the locals in locals.tf, resolved with the same defaults.
type settings struct {
	// version is the release installed: spec.version, or vars.Release.
	version string

	requestedNamespace string
	createNamespace    bool
	identityLabels     map[string]string

	gke                  bool
	gkeClusterID         string
	apiServerGSA         string
	nodeAgentGSA         string
	transferGSA          string
	serviceAccountIssuer string

	database    databaseSettings
	storage     storageSettings
	apiServer   apiServerSettings
	identity    identityServerSettings
	controller  controllerSettings
	nodeAgent   nodeAgentSettings
	router      routerSettings
	egress      egressSettings
	podCerts    podCertificateSettings
	telemetry   telemetrySettings
	preview     []string
	size10      bool
	cordon      bool
	registry    string
	pullSecrets []string
	crdsInstall bool
	crdsKeep    bool
}

type secretKey struct {
	name string
	key  string
}

type databaseSettings struct {
	kind              string // bundled | external | cloudsql
	poolMaxConns      int32  // 0 = unset
	storageClass      string
	volumeSize        string
	resources         *kubernetesprovider.ContainerResources
	readWrite         *secretKey
	owner             *secretKey
	roleReadWrite     string
	roleOwner         string
	schema            string
	serverCA          string
	cloudSQLInstance  string
	cloudSQLIAM       bool
	cloudSQLIPType    string // private | public | psc
	externalPasswords bool   // the connection strings come from Secrets
}

type storageSettings struct {
	prefix      string
	gcsBucket   string
	s3Endpoint  string
	s3Bucket    string
	s3Region    string
	s3PathStyle bool
	s3Secret    string
	s3IDKey     string
	s3SecretKey string
}

func (s storageSettings) isS3() bool { return s.s3Bucket != "" }

// location is where actor snapshots are stored, the API server's snapshot
// location: the bucket, under the prefix when one is set.
func (s storageSettings) location() string {
	location := "gs://" + s.gcsBucket
	if s.isS3() {
		location = "s3://" + s.s3Bucket
	}
	if s.prefix != "" {
		location += "/" + s.prefix
	}
	return location
}

type apiServerSettings struct {
	replicas        int32
	resources       *kubernetesprovider.ContainerResources
	logLevel        string
	bootstrapOwners []string
	jwtAlgorithm    string // ES256 | RS256
	actorIssuer     string
	drainDelay      string
	drainTimeout    string
	resyncInterval  string
}

type identityServerSettings struct {
	replicas   int32
	resources  *kubernetesprovider.ContainerResources
	logLevel   string
	drainGrace string
}

type controllerSettings struct {
	resources  *kubernetesprovider.ContainerResources
	logLevel   string
	workerSync int32
}

type nodeAgentSettings struct {
	resources    *kubernetesprovider.ContainerResources
	logLevel     string
	nodeSelector map[string]string
	tolerations  []*kubernetesprovider.WorkloadToleration
	statsPoll    string
	drainDelay   string
	drainTimeout string
	gcPeriod     string
	highPercent  int32
	lowPercent   int32
	maxBytes     int64
	minAge       string
	gcDryRun     bool
	stagingGiB   int32
	snapshots    nodeSnapshotSettings
	assetCache   fileCacheSettings
	snapCache    fileCacheSettings
}

type nodeSnapshotSettings struct {
	uploadDelay   string
	detachDelay   string
	uploadTimeout string
	uploadMiB     int32
	uploadWorkers int32
	detachWorkers int32
	reportWorkers int32
	reportGiveUp  string
	uploadMemory  string
}

// fileCacheSettings is one of the node's file caches: the sandbox asset
// cache or the snapshot cache, whose flags share a shape under their own
// prefix.
type fileCacheSettings struct {
	disabled     bool
	minAge       string
	fetchTimeout string
	gcPeriod     string
	maxBytes     int64
	gcDryRun     bool
}

type routerSettings struct {
	agentgateway      bool
	replicas          int32
	resources         *kubernetesprovider.ContainerResources
	logLevel          string
	spiffeIDs         []string
	routeTimeout      string
	parkBudget        string
	parkMax           int32
	parkRetryInterval string
	parkRetryFactor   float64
	parkRetryJitter   float64
	extprocMax        int32
	drainDelay        string
	drainTimeout      string
}

type namespaceGrant struct {
	atespace   string
	namespaces []string
}

type egressSettings struct {
	provider          string // disabled | kubernetes | external
	providerName      string
	providerAddress   string
	grants            []namespaceGrant
	providerResources *kubernetesprovider.ContainerResources
	providerLogLevel  string
	providerQPS       float64
	providerBurst     int32
	providerDrain     string
	replicas          int32
	resources         *kubernetesprovider.ContainerResources
	logLevel          string
	policyTTL         string
	interceptTTL      string
	envoyConcurrency  int32 // 0 = Envoy's default
	extprocNamespace  string
	extprocService    string
	extprocPort       int32
	drainDelay        string
}

type podCertificateSettings struct {
	resources        *kubernetesprovider.ContainerResources
	workersPerSigner int32
	qps              float64 // 0 = unset
	burst            int32   // 0 = unset
}

type telemetrySettings struct {
	endpoint       string
	traces         []string
	metrics        []string
	logs           []string
	workerTraces   []string
	workerMetrics  []string
	workerLogs     []string
	sampler        string
	samplerRatio   *float64
	exportInterval int32
	exportTimeout  int32
}

// defaultBootstrapOwner is upstream's bootstrap owner when the spec names
// none: ate-system's ate-client service account (installdefaults
// ClientPrincipal).
const defaultBootstrapOwner = "system:serviceaccount:ate-system:ate-client"

// The API server's in-cluster issuer forms: for these the API server
// verifies tokens against the cluster itself, with its own projected CA and
// token (upstream's buildAuthenticationConfig).
const inClusterIssuer = "https://kubernetes.default.svc"

// resolveSettings fills every default. The spec's validation has already
// held each value to its form, so nothing here re-validates.
func resolveSettings(spec *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateSpec) *settings {
	s := &settings{
		version:            stringOr(spec.Version, vars.Release),
		requestedNamespace: spec.GetNamespace().GetValue(),
		createNamespace:    spec.GetCreateNamespace(),
		size10:             spec.GetClusterSize() == kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateSpec_size10,
		cordon:             spec.GetCordonControlPlane(),
		registry:           spec.GetImageRegistry(),
		pullSecrets:        spec.GetImagePullSecrets(),
		preview:            spec.GetPreviewGates(),
		crdsInstall:        spec.GetCrds() == nil || spec.GetCrds().Install == nil || spec.GetCrds().GetInstall(),
		crdsKeep:           spec.GetCrds() == nil || spec.GetCrds().KeepOnUninstall == nil || spec.GetCrds().GetKeepOnUninstall(),
	}
	if g := spec.GetPlatform().GetGke(); g != nil {
		s.gke = true
		s.gkeClusterID = g.GetClusterId().GetValue()
		s.apiServerGSA = g.GetApiServerServiceAccount().GetValue()
		s.nodeAgentGSA = g.GetNodeAgentServiceAccount().GetValue()
		s.transferGSA = g.GetSnapshotTransferServiceAccount().GetValue()
	} else {
		s.serviceAccountIssuer = spec.GetPlatform().GetGeneric().GetServiceAccountIssuer()
	}
	s.database = resolveDatabase(spec.GetDatabase())
	s.storage = resolveStorage(spec.GetStorage())
	s.apiServer = resolveAPIServer(spec.GetApiServer())
	identity := orEmpty(spec.GetIdentityServer())
	s.identity = identityServerSettings{
		replicas:   int32Or(identity.Replicas, 2),
		resources:  identity.GetResources(),
		logLevel:   logLevel(identity.GetLogLevel()),
		drainGrace: stringOr(identity.DrainGrace, "5s"),
	}
	controller := orEmpty(spec.GetController())
	s.controller = controllerSettings{
		resources:  controller.GetResources(),
		logLevel:   logLevel(controller.GetLogLevel()),
		workerSync: int32Or(controller.WorkerSyncConcurrency, 2),
	}
	s.nodeAgent = resolveNodeAgent(spec.GetNodeAgent())
	s.router = resolveRouter(spec.GetRouter())
	s.egress = resolveEgress(spec.GetEgress())
	podCerts := orEmpty(spec.GetPodCertificates())
	s.podCerts = podCertificateSettings{
		resources:        podCerts.GetResources(),
		workersPerSigner: int32Or(podCerts.WorkersPerSigner, 1),
		qps:              spec.GetPodCertificates().GetKubeApiQps(),
		burst:            spec.GetPodCertificates().GetKubeApiBurst(),
	}
	t := orEmpty(spec.GetTelemetry())
	s.telemetry = telemetrySettings{
		endpoint:       t.GetOtlpEndpoint(),
		traces:         t.GetTracesExporters(),
		metrics:        t.GetMetricsExporters(),
		logs:           t.GetLogsExporters(),
		workerTraces:   t.GetWorkerTracesExporters(),
		workerMetrics:  t.GetWorkerMetricsExporters(),
		workerLogs:     t.GetWorkerLogsExporters(),
		sampler:        t.GetTracesSampler(),
		samplerRatio:   t.TracesSamplerRatio,
		exportInterval: t.GetMetricExportIntervalMs(),
		exportTimeout:  t.GetMetricExportTimeoutMs(),
	}
	return s
}

func resolveDatabase(d *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateDatabase) databaseSettings {
	out := databaseSettings{
		poolMaxConns:  d.GetPoolMaxConnections(),
		roleReadWrite: "substrate_readwrite",
		roleOwner:     "substrate_owner",
		schema:        "substrate",
	}
	roles := func(r *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateDatabaseRoles) {
		r = orEmpty(r)
		out.roleReadWrite = stringOr(r.ReadWrite, out.roleReadWrite)
		out.roleOwner = stringOr(r.Owner, out.roleOwner)
		out.schema = stringOr(r.Schema, out.schema)
	}
	switch {
	case d.GetExternal() != nil:
		e := d.GetExternal()
		out.kind = "external"
		out.externalPasswords = true
		out.readWrite = toSecretKey(e.GetReadWriteConnectionString())
		out.owner = toSecretKey(e.GetOwnerConnectionString())
		out.serverCA = e.GetServerCaCertificate()
		roles(e.GetRoles())
	case d.GetCloudSql() != nil:
		c := d.GetCloudSql()
		out.kind = "cloudsql"
		out.cloudSQLInstance = c.GetInstanceConnectionName().GetValue()
		out.cloudSQLIAM = c.IamAuthentication == nil || c.GetIamAuthentication()
		out.cloudSQLIPType = strings.TrimSuffix(c.GetIpType().String(), "_ip")
		if !out.cloudSQLIAM {
			out.externalPasswords = true
			out.readWrite = toSecretKey(c.GetReadWriteConnectionString())
			out.owner = toSecretKey(c.GetOwnerConnectionString())
		}
		roles(c.GetRoles())
	default:
		b := orEmpty(d.GetBundled())
		out.kind = "bundled"
		out.storageClass = b.GetStorageClass()
		out.volumeSize = stringOr(b.VolumeSize, "500Gi")
		out.resources = b.GetResources()
	}
	if out.owner == nil {
		out.owner = out.readWrite
	}
	return out
}

func resolveStorage(st *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateStorage) storageSettings {
	if g := st.GetGcs(); g != nil {
		return storageSettings{prefix: st.GetSnapshotPrefix(), gcsBucket: g.GetBucket().GetValue()}
	}
	s3 := orEmpty(st.GetS3())
	c := orEmpty(s3.GetCredentials())
	return storageSettings{
		prefix:      st.GetSnapshotPrefix(),
		s3Endpoint:  s3.GetEndpoint().GetValue(),
		s3Bucket:    s3.GetBucket(),
		s3Region:    stringOr(s3.Region, "us-east-1"),
		s3PathStyle: s3.ForcePathStyle == nil || s3.GetForcePathStyle(),
		s3Secret:    c.GetSecretName().GetValue(),
		s3IDKey:     stringOr(c.AccessKeyIdKey, "access_key_id"),
		s3SecretKey: stringOr(c.SecretAccessKeyKey, "secret_access_key"),
	}
}

func resolveAPIServer(a *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateApiServer) apiServerSettings {
	a = orEmpty(a)
	out := apiServerSettings{
		replicas:       int32Or(a.Replicas, 2),
		resources:      a.GetResources(),
		logLevel:       logLevel(a.GetLogLevel()),
		jwtAlgorithm:   strings.ToUpper(a.GetActorJwt().GetAlgorithm().String()),
		actorIssuer:    a.GetActorJwt().GetIssuer(),
		drainDelay:     stringOr(a.DrainDelay, "13s"),
		drainTimeout:   stringOr(a.DrainTimeout, "15s"),
		resyncInterval: stringOr(a.TemplateResyncInterval, "20s"),
	}
	if out.actorIssuer == "" {
		out.actorIssuer = "https://idp." + vars.Namespace + ".svc"
	}
	out.bootstrapOwners = a.GetAuthorization().GetBootstrapOwners()
	if len(out.bootstrapOwners) == 0 {
		out.bootstrapOwners = []string{defaultBootstrapOwner}
	}
	return out
}

func resolveNodeAgent(n *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateNodeAgent) nodeAgentSettings {
	n = orEmpty(n)
	c := orEmpty(n.GetImageCache())
	snapshots := orEmpty(n.GetSnapshots())
	assets := orEmpty(n.GetSandboxAssetCache())
	snapCache := orEmpty(n.GetSnapshotCache())
	return nodeAgentSettings{
		resources:    n.GetResources(),
		logLevel:     logLevel(n.GetLogLevel()),
		nodeSelector: n.GetNodeSelector(),
		tolerations:  n.GetTolerations(),
		statsPoll:    stringOr(n.ActorStatsPollInterval, "1m"),
		drainDelay:   stringOr(n.DrainDelay, "0s"),
		drainTimeout: stringOr(n.DrainTimeout, "5m"),
		gcPeriod:     stringOr(c.GcPeriod, "5m"),
		highPercent:  int32Or(c.HighPercent, 85),
		lowPercent:   int32Or(c.LowPercent, 80),
		maxBytes:     int64Or(c.MaxBytes, 0),
		minAge:       stringOr(c.MinAge, "2m"),
		gcDryRun:     c.GetDryRun(),
		stagingGiB:   int32Or(n.AssetStagingCapacityGib, 8),
		snapshots: nodeSnapshotSettings{
			uploadDelay:   stringOr(snapshots.UploadDelay, "5m"),
			detachDelay:   stringOr(snapshots.DetachDelay, "30m"),
			uploadTimeout: stringOr(snapshots.UploadTimeout, "2m"),
			uploadMiB:     int32Or(snapshots.UploadThroughputMib, 100),
			uploadWorkers: int32Or(snapshots.UploadWorkers, 2),
			detachWorkers: int32Or(snapshots.DetachWorkers, 2),
			reportWorkers: int32Or(snapshots.ReportWorkers, 1),
			reportGiveUp:  stringOr(snapshots.ReportGiveUpAfter, "10m"),
			uploadMemory:  stringOr(snapshots.UploadMemoryBudget, "512Mi"),
		},
		assetCache: resolveFileCache(false, assets.MinAge, "1h", assets.FetchTimeout, "5m", assets.GetEviction()),
		snapCache:  resolveFileCache(snapCache.GetDisabled(), snapCache.MinAge, "10m", snapCache.FetchTimeout, "10m", snapCache.GetEviction()),
	}
}

// resolveFileCache fills one file cache's defaults; the two caches differ
// only in their ages and timeouts.
func resolveFileCache(disabled bool, minAge *string, minAgeDefault string, fetchTimeout *string, fetchTimeoutDefault string, e *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateFileCacheEviction) fileCacheSettings {
	e = orEmpty(e)
	return fileCacheSettings{
		disabled:     disabled,
		minAge:       stringOr(minAge, minAgeDefault),
		fetchTimeout: stringOr(fetchTimeout, fetchTimeoutDefault),
		gcPeriod:     stringOr(e.GcPeriod, "5m"),
		maxBytes:     int64Or(e.MaxBytes, 0),
		gcDryRun:     e.GetDryRun(),
	}
}

func resolveRouter(r *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateRouter) routerSettings {
	r = orEmpty(r)
	p := orEmpty(r.GetParkedRequests())
	return routerSettings{
		agentgateway:      r.GetDataplane() == kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateRouter_agentgateway,
		replicas:          int32Or(r.Replicas, 1),
		resources:         r.GetResources(),
		logLevel:          logLevel(r.GetLogLevel()),
		spiffeIDs:         r.GetAllowedClientSpiffeIds(),
		routeTimeout:      stringOr(r.RouteTimeout, "5m"),
		parkBudget:        stringOr(p.Budget, "5s"),
		parkMax:           int32Or(p.Max, 1024),
		parkRetryInterval: stringOr(p.RetryInterval, "100ms"),
		parkRetryFactor:   float64Or(p.RetryFactor, 1.1),
		parkRetryJitter:   float64Or(p.RetryJitter, 0.1),
		extprocMax:        int32Or(r.ExtProcMaxRequests, 0),
		drainDelay:        stringOr(r.DrainDelay, "13s"),
		drainTimeout:      r.GetDrainTimeout(),
	}
}

func resolveEgress(e *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateEgress) egressSettings {
	e = orEmpty(e)
	out := egressSettings{
		provider:         "disabled",
		replicas:         int32Or(e.Replicas, 1),
		resources:        e.GetResources(),
		logLevel:         logLevel(e.GetLogLevel()),
		policyTTL:        stringOr(e.PolicyCacheTtl, "10s"),
		interceptTTL:     stringOr(e.InterceptCertificateTtl, "15m"),
		envoyConcurrency: e.GetEnvoyConcurrency(),
		drainDelay:       stringOr(e.DrainDelay, "13s"),
	}
	if x := e.GetAdditionalExtProcService(); x != nil {
		out.extprocNamespace, out.extprocService, out.extprocPort = x.GetNamespace(), x.GetService(), x.GetPort()
	}
	switch p := e.GetCredentialProvider(); {
	case p.GetKubernetesSecrets() != nil:
		k := p.GetKubernetesSecrets()
		out.provider = "kubernetes"
		out.providerName = "k8s.io"
		out.providerAddress = vars.CredentialProviderName + "." + vars.Namespace + ".svc:50051"
		for _, g := range k.GetGrants() {
			out.grants = append(out.grants, namespaceGrant{atespace: g.GetAtespace(), namespaces: g.GetAllowedNamespaces()})
		}
		out.providerResources = k.GetResources()
		out.providerLogLevel = logLevel(k.GetLogLevel())
		out.providerQPS = float64Or(k.KubeApiQps, 50)
		out.providerBurst = int32Or(k.KubeApiBurst, 100)
		out.providerDrain = stringOr(k.DrainGrace, "5s")
	case p.GetExternal() != nil:
		out.provider = "external"
		out.providerName = p.GetExternal().GetName()
		out.providerAddress = p.GetExternal().GetAddress()
	}
	return out
}

// providerServerName is the name the gateway verifies a provider as: its
// address without the port (upstream's CredentialProvider.ServerName).
func (e egressSettings) providerServerName() string {
	if i := strings.LastIndex(e.providerAddress, ":"); i >= 0 {
		return e.providerAddress[:i]
	}
	return e.providerAddress
}

// orEmpty stands an empty message in for an absent one, so optional fields
// read as unset rather than through a nil message.
func orEmpty[T any](m *T) *T {
	if m == nil {
		return new(T)
	}
	return m
}

func toSecretKey(r *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateSecretKeyRef) *secretKey {
	if r == nil {
		return nil
	}
	return &secretKey{name: r.GetName().GetValue(), key: r.GetKey()}
}

// logLevel is the spec's level as the binaries spell it; info is every
// binary's default.
func logLevel(l kubernetesagentsubstratev1alpha1.LogLevel) string { return l.String() }

func stringOr(v *string, def string) string {
	if v == nil {
		return def
	}
	return *v
}

func int32Or(v *int32, def int32) int32 {
	if v == nil {
		return def
	}
	return *v
}

func int64Or(v *int64, def int64) int64 {
	if v == nil {
		return def
	}
	return *v
}

func float64Or(v *float64, def float64) float64 {
	if v == nil {
		return def
	}
	return *v
}
