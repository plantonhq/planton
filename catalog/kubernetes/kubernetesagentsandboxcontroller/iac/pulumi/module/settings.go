package module

import (
	"strconv"
	"strings"

	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesagentsandboxcontrollerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandboxcontroller/v1alpha1"
)

// The resolved settings: the spec with every default filled in, so the
// module installs the same thing whether or not the platform's defaulting
// ran before it. Each default equals the spec's declared default, which
// equals upstream's own; a flag is written onto upstream's arguments only
// when its value differs, so upstream's defaults stay upstream's. The
// Terraform twin resolves the same values in locals.tf.

type spec = kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerSpec

type settings struct {
	// requestedNamespace is spec.namespace resolved to a name; the module
	// refuses any name but upstream's.
	requestedNamespace string
	createNamespace    bool
	// identityLabels go on the Namespace this module creates (set from the
	// resource's metadata, not the spec).
	identityLabels      map[string]string
	extensions          bool
	allowedLabelDomains []string
	clusterDomain       string
	imageRegistry       string
	imagePullSecrets    []string
	controller          controllerSettings
	claim               claimSettings
	warmPool            warmPoolSettings
	templateWorkers     int32
	router              *routerSettings
}

type controllerSettings struct {
	image                   string
	imagePullSecret         string
	replicas                int32
	resources               map[string]interface{}
	nodeSelector            map[string]string
	tolerations             []interface{}
	leaderElect             bool
	leaderElectionNamespace string
	metricsPort             int32
	restClientMetrics       bool
	metricsTLS              *metricsTLSSettings
	healthProbePort         int32
	sandboxWorkers          int32
	sandboxEventsDisabled   bool
	writeBehindWindow       string
	cacheLabelSelectors     bool
	qps                     *float64
	burst                   int32
	connections             int32
	separateWatchConnection bool
	tracing                 *telemetrySettings
	profilingCPU            bool
	profilingDebug          bool
	blockProfileRate        int32
	mutexProfileFraction    int32
	loggingDevelopment      bool
	loggingEncoder          string
	loggingLevel            string
	loggingStacktraceLevel  string
	loggingTimeEncoding     string
}

type metricsTLSSettings struct {
	certificateSecret string
	minVersion        string
	cipherSuites      []string
}

type telemetrySettings struct {
	endpoint      string
	insecure      bool
	headersSecret string
	headersKey    string
}

type claimSettings struct {
	workers                          int32
	warmCandidateGracePeriod         string
	eventsDisabled                   bool
	observabilityAnnotationsDisabled bool
}

type warmPoolSettings struct {
	workers                      int32
	maxBatchSize                 int32
	replenishDelay               string
	maxRefillRate                float64
	readinessGracePeriod         string
	unschedulableRecheckInterval string
	evictionEnabled              bool
}

type routerSettings struct {
	image                     string
	imagePullSecret           string
	replicas                  int32
	resources                 map[string]interface{}
	nodeSelector              map[string]string
	tolerations               []interface{}
	minAvailable              int32
	authzMode                 string
	tokenReviewTTL            string
	tokenReviewCacheSize      int32
	tokenReviewRequireToken   bool
	tokenReviewAudiences      []string
	scopedTokenV1             *secretKey
	scopedTokenV2             *secretKey
	scopedTokenV1AcceptUntil  string
	cookie                    *cookieSettings
	tls                       *routerTLSSettings
	httpEnabled               bool
	httpPort                  int32
	metricsPort               int32
	healthProbePort           int32
	proxyTimeout              string
	responseHeaderTimeout     string
	shutdownTimeout           string
	maxRequestBodyBytes       int64
	allowLoopbackPodIP        bool
	pathRoutingPrefix         string
	upstreamMaxRetries        int32
	upstreamRetryInitialDelay string
	upstreamRetryMaxDelay     string
	cacheEnabled              bool
	cacheNamespace            string
	telemetry                 *telemetrySettings
	tracesEnabled             bool
	otelMetricsEnabled        bool
	accessLog                 bool
	networkPolicy             *networkPolicySettings
}

type secretKey struct {
	secret string
	key    string
}

type cookieSettings struct {
	name                string
	queryParam          string
	sameSite            string
	insecure            bool
	allowedOrigins      []string
	trustForwardedProto bool
}

type routerTLSSettings struct {
	certificateSecret string
	port              int32
	mtlsMode          string
	clientCA          *secretKey
	minVersion        string
	cipherSuites      []string
}

type networkPolicySettings struct {
	clientNamespaceLabels map[string]string
	metricsNamespace      string
	sandboxPorts          []int32
}

// defaultSandboxPort is upstream's default sandbox port, the one the router's
// NetworkPolicy opens when the spec lists none.
const defaultSandboxPort = 8888

func resolveSettings(s *spec) *settings {
	r := &settings{
		requestedNamespace: s.GetNamespace().GetValue(),
		createNamespace:    s.GetCreateNamespace(),
		clusterDomain:      stringOr(s.ClusterDomain, "cluster.local"),
		imageRegistry:      s.GetImageRegistry(),
		imagePullSecrets:   s.GetImagePullSecrets(),
	}

	if ext := s.GetExtensions(); ext != nil && boolOr(ext.Enabled, true) {
		r.extensions = true
		r.allowedLabelDomains = ext.GetAllowedLabelDomains()
	}
	ext := orEmpty(s.GetExtensions())
	claim, pool := orEmpty(ext.GetClaim()), orEmpty(ext.GetWarmPool())
	r.claim = claimSettings{
		workers:                          int32Or(claim.ConcurrentWorkers, 50),
		warmCandidateGracePeriod:         stringOr(claim.WarmCandidateGracePeriod, "2s"),
		eventsDisabled:                   claim.GetEventsDisabled(),
		observabilityAnnotationsDisabled: claim.GetObservabilityAnnotationsDisabled(),
	}
	r.warmPool = warmPoolSettings{
		workers:                      int32Or(pool.ConcurrentWorkers, 1),
		maxBatchSize:                 int32Or(pool.MaxBatchSize, 300),
		replenishDelay:               stringOr(pool.ReplenishDelay, "0s"),
		maxRefillRate:                float64Or(pool.MaxRefillRate, 0),
		readinessGracePeriod:         stringOr(pool.ReadinessGracePeriod, "5m"),
		unschedulableRecheckInterval: stringOr(pool.UnschedulableRecheckInterval, "1m"),
		evictionEnabled:              boolOr(pool.EvictionEnabled, true),
	}
	r.templateWorkers = int32Or(orEmpty(ext.GetTemplate()).ConcurrentWorkers, 1)

	c := orEmpty(s.GetController())
	metrics, logging := orEmpty(c.GetMetrics()), orEmpty(c.GetLogging())
	leader, sandbox := orEmpty(c.GetLeaderElection()), orEmpty(c.GetSandbox())
	kubeAPI, profiling := orEmpty(c.GetKubeApi()), orEmpty(c.GetProfiling())
	r.controller = controllerSettings{
		image:                   imageOverride(c.GetImage()),
		imagePullSecret:         c.GetImage().GetPullSecretName(),
		replicas:                int32Or(c.Replicas, 1),
		resources:               containerResources(c.GetResources()),
		nodeSelector:            c.GetNodeSelector(),
		tolerations:             tolerations(c.GetTolerations()),
		leaderElect:             boolOr(leader.Enabled, true),
		leaderElectionNamespace: leader.GetNamespace(),
		metricsPort:             int32Or(metrics.Port, 8080),
		restClientMetrics:       boolOr(metrics.RestClientMetricsEnabled, true),
		healthProbePort:         int32Or(c.HealthProbePort, 8081),
		sandboxWorkers:          int32Or(sandbox.ConcurrentWorkers, 100),
		sandboxEventsDisabled:   sandbox.GetEventsDisabled(),
		writeBehindWindow:       stringOr(sandbox.WriteBehindWindow, "0s"),
		cacheLabelSelectors:     c.GetCacheLabelSelectors(),
		qps:                     kubeAPI.Qps,
		burst:                   int32Or(kubeAPI.Burst, 10),
		connections:             int32Or(kubeAPI.Connections, 1),
		separateWatchConnection: kubeAPI.GetSeparateWatchConnection(),
		profilingCPU:            profiling.GetCpuEnabled(),
		profilingDebug:          profiling.GetDebugEnabled(),
		blockProfileRate:        int32Or(profiling.BlockProfileRate, 1000000),
		mutexProfileFraction:    int32Or(profiling.MutexProfileFraction, 10),
		loggingDevelopment:      logging.GetDevelopment(),
		loggingLevel:            logging.GetLevel(),
	}
	if logging.GetEncoder() != 0 {
		r.controller.loggingEncoder = logging.GetEncoder().String()
	}
	if logging.GetStacktraceLevel() != 0 {
		r.controller.loggingStacktraceLevel = logging.GetStacktraceLevel().String()
	}
	if logging.GetTimeEncoding() != 0 {
		r.controller.loggingTimeEncoding = logging.GetTimeEncoding().String()
	}
	if tls := metrics.GetTls(); tls != nil && boolOr(tls.Enabled, true) {
		r.controller.metricsTLS = &metricsTLSSettings{
			certificateSecret: tls.GetCertificateSecret().GetValue(),
			minVersion:        tlsVersion(tls.GetMinVersion()),
			cipherSuites:      tls.GetCipherSuites(),
		}
	}
	if t := c.GetTracing(); t != nil {
		r.controller.tracing = telemetry(t.GetOtlpEndpoint(), t.GetInsecure(), t.GetHeaders())
	}

	if rt := s.GetRouter(); rt != nil && boolOr(rt.Enabled, true) {
		r.router = resolveRouter(rt)
	}
	return r
}

func resolveRouter(rt *kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerRouter) *routerSettings {
	proxy, cache, ports := orEmpty(rt.GetProxy()), orEmpty(rt.GetCache()), orEmpty(rt.GetPorts())
	r := &routerSettings{
		image:                     imageOverride(rt.GetImage()),
		imagePullSecret:           rt.GetImage().GetPullSecretName(),
		replicas:                  int32Or(rt.Replicas, 2),
		resources:                 containerResources(rt.GetResources()),
		nodeSelector:              rt.GetNodeSelector(),
		tolerations:               tolerations(rt.GetTolerations()),
		minAvailable:              int32Or(rt.MinAvailable, 1),
		authzMode:                 "allow-all",
		httpEnabled:               true,
		httpPort:                  int32Or(ports.Http, 8080),
		metricsPort:               int32Or(ports.Metrics, 9090),
		healthProbePort:           int32Or(ports.HealthProbe, 8081),
		proxyTimeout:              stringOr(proxy.Timeout, "180s"),
		responseHeaderTimeout:     stringOr(proxy.ResponseHeaderTimeout, "30s"),
		shutdownTimeout:           stringOr(proxy.ShutdownTimeout, "30s"),
		maxRequestBodyBytes:       proxy.GetMaxRequestBodyBytes(),
		allowLoopbackPodIP:        proxy.GetAllowLoopbackPodIp(),
		pathRoutingPrefix:         proxy.GetPathRoutingPrefix(),
		upstreamMaxRetries:        int32Or(proxy.UpstreamMaxRetries, 3),
		upstreamRetryInitialDelay: stringOr(proxy.UpstreamRetryInitialDelay, "200ms"),
		upstreamRetryMaxDelay:     stringOr(proxy.UpstreamRetryMaxDelay, "800ms"),
		cacheEnabled:              boolOr(cache.Enabled, true),
		cacheNamespace:            cache.GetNamespace(),
		accessLog:                 boolOr(rt.AccessLogEnabled, true),
	}

	authz := rt.GetAuthorization()
	switch {
	case authz.GetTokenReview() != nil:
		tr := authz.GetTokenReview()
		r.authzMode = "tokenreview"
		r.tokenReviewTTL = stringOr(tr.Ttl, "30s")
		r.tokenReviewCacheSize = int32Or(tr.CacheSize, 2048)
		r.tokenReviewRequireToken = tr.GetRequireToken()
		r.tokenReviewAudiences = tr.GetAudiences()
	case authz.GetScopedToken() != nil:
		st := authz.GetScopedToken()
		r.authzMode = "scoped-token"
		r.scopedTokenV1 = toSecretKey(st.GetV1Secret())
		r.scopedTokenV2 = toSecretKey(st.GetVerificationKeys())
		r.scopedTokenV1AcceptUntil = st.GetV1AcceptUntil()
	}
	if ck := authz.GetCookie(); ck != nil {
		r.cookie = &cookieSettings{
			name:                ck.GetName(),
			queryParam:          ck.GetQueryParam(),
			insecure:            ck.GetInsecure(),
			allowedOrigins:      ck.GetAllowedOrigins(),
			trustForwardedProto: ck.GetTrustForwardedProto(),
		}
		if ck.GetSameSite() != 0 {
			r.cookie.sameSite = ck.GetSameSite().String()
		}
	}

	if tls := rt.GetTls(); tls != nil && boolOr(tls.Enabled, true) {
		r.tls = &routerTLSSettings{
			certificateSecret: tls.GetCertificateSecret().GetValue(),
			port:              int32Or(tls.Port, 8443),
			clientCA:          toSecretKey(tls.GetClientCa()),
			minVersion:        tlsVersion(tls.GetMinVersion()),
			cipherSuites:      tls.GetCipherSuites(),
		}
		if mode := tls.GetMtlsMode(); mode != 0 {
			r.tls.mtlsMode = mode.String()
		}
		r.httpEnabled = boolOr(tls.HttpEnabled, true)
	}

	if t := rt.GetTelemetry(); t != nil {
		r.telemetry = telemetry(t.GetOtlpEndpoint(), t.GetInsecure(), t.GetHeaders())
		r.tracesEnabled = boolOr(t.TracesEnabled, true)
		r.otelMetricsEnabled = boolOr(t.MetricsEnabled, true)
	}

	if np := orEmpty(rt.GetNetworkPolicy()); boolOr(np.Enabled, true) {
		r.networkPolicy = &networkPolicySettings{
			clientNamespaceLabels: np.GetClientNamespaceLabels(),
			metricsNamespace:      stringOr(np.MetricsNamespace, "monitoring"),
			sandboxPorts:          np.GetSandboxPorts(),
		}
		if len(r.networkPolicy.sandboxPorts) == 0 {
			r.networkPolicy.sandboxPorts = []int32{defaultSandboxPort}
		}
	}
	return r
}

// imageOverride folds an image override into one reference: "repo:tag" when
// both are set, whichever is set otherwise, "" for none.
func imageOverride(image *kubernetes.ContainerImage) string {
	repo, tag := image.GetRepo(), image.GetTag()
	if repo != "" && tag != "" {
		return repo + ":" + tag
	}
	return repo + tag
}

func containerResources(res *kubernetes.ContainerResources) map[string]interface{} {
	if res == nil {
		return nil
	}
	out := map[string]interface{}{}
	for name, cm := range map[string]*kubernetes.CpuMemory{"requests": res.GetRequests(), "limits": res.GetLimits()} {
		values := map[string]interface{}{}
		if cm.GetCpu() != "" {
			values["cpu"] = cm.GetCpu()
		}
		if cm.GetMemory() != "" {
			values["memory"] = cm.GetMemory()
		}
		if len(values) > 0 {
			out[name] = values
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func tolerations(in []*kubernetes.WorkloadToleration) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, t := range in {
		toleration := map[string]interface{}{}
		for key, value := range map[string]string{"key": t.GetKey(), "operator": t.GetOperator(), "value": t.GetValue(), "effect": t.GetEffect()} {
			if value != "" {
				toleration[key] = value
			}
		}
		if t.TolerationSeconds != nil {
			toleration["tolerationSeconds"] = int64(t.GetTolerationSeconds())
		}
		out = append(out, toleration)
	}
	return out
}

func telemetry(endpoint string, insecure bool, headers *kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerSecretKey) *telemetrySettings {
	t := &telemetrySettings{endpoint: endpoint, insecure: insecure}
	if h := toSecretKey(headers); h != nil {
		t.headersSecret, t.headersKey = h.secret, h.key
	}
	return t
}

func toSecretKey(k *kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerSecretKey) *secretKey {
	if k == nil {
		return nil
	}
	return &secretKey{secret: k.GetSecret().GetValue(), key: k.GetKey()}
}

// tlsVersion maps the spec's version to the Go name both binaries accept.
func tlsVersion(v kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerTlsVersion) string {
	if v == 0 {
		return ""
	}
	return "VersionTLS1" + strings.TrimPrefix(v.String(), "tls1")
}

// orEmpty stands an empty message in for an absent block, so a presence-
// tracked field inside it reads as unset rather than dereferencing nil.
func orEmpty[T any](message *T) *T {
	if message == nil {
		return new(T)
	}
	return message
}

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

func boolOr(v *bool, def bool) bool {
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

// formatFloat writes a number the way the Terraform twin's tostring() does:
// no exponent, no trailing zeros.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
