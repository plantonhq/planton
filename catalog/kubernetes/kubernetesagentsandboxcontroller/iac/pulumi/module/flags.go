package module

import (
	"strconv"
	"strings"
)

// flag is one command-line setting, written as --name=value.
type flag struct {
	name  string
	value string
}

// applyFlags writes typed settings onto upstream's own argument list. A flag
// upstream already passes keeps its position and takes the new value; every
// other flag is appended in the order given. Upstream's own arguments that no
// setting names are kept verbatim, so what upstream decides stays upstream's.
// The Terraform twin builds the identical list in locals.tf.
func applyFlags(upstream []string, flags []flag) []string {
	byName := make(map[string]string, len(flags))
	for _, f := range flags {
		byName[f.name] = f.value
	}
	used := map[string]bool{}
	out := make([]string, 0, len(upstream)+len(flags))
	for _, arg := range upstream {
		name := strings.TrimPrefix(strings.SplitN(arg, "=", 2)[0], "--")
		if value, ok := byName[name]; ok {
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
	return out
}

// flagList collects flags whose condition holds, in order.
type flagList []flag

func (l *flagList) add(condition bool, name string, value func() string) {
	if condition {
		*l = append(*l, flag{name: name, value: value()})
	}
}

func is(value string) func() string { return func() string { return value } }

func itoa(i int32) func() string { return func() string { return strconv.Itoa(int(i)) } }

func port(i int32) func() string { return func() string { return ":" + strconv.Itoa(int(i)) } }

func boolean(b bool) func() string { return func() string { return strconv.FormatBool(b) } }

func joined(items []string) func() string { return func() string { return strings.Join(items, ",") } }

// controllerFlags are the controller's settings that differ from upstream's
// defaults, in the order the Terraform twin writes them.
func controllerFlags(s *settings) []flag {
	c := s.controller
	var l flagList
	l.add(!c.leaderElect, "leader-elect", is("false"))
	l.add(c.leaderElectionNamespace != "", "leader-election-namespace", is(c.leaderElectionNamespace))
	l.add(s.clusterDomain != "cluster.local", "cluster-domain", is(s.clusterDomain))
	l.add(c.metricsPort != 8080, "metrics-bind-address", port(c.metricsPort))
	l.add(c.healthProbePort != 8081, "health-probe-bind-address", port(c.healthProbePort))
	l.add(!c.restClientMetrics, "enable-rest-client-metrics", is("false"))
	if t := c.metricsTLS; t != nil {
		l.add(true, "metrics-secure-serving", is("true"))
		l.add(t.certificateSecret != "", "metrics-cert-dir", is(vars.MetricsTLSMountPath))
		l.add(t.minVersion != "", "tls-min-version", is(t.minVersion))
		l.add(len(t.cipherSuites) > 0, "tls-cipher-suites", joined(t.cipherSuites))
	}
	l.add(c.sandboxWorkers != 100, "sandbox-concurrent-workers", itoa(c.sandboxWorkers))
	l.add(c.sandboxEventsDisabled, "disable-sandbox-events", is("true"))
	l.add(c.writeBehindWindow != "0s", "sandbox-write-behind-window", is(c.writeBehindWindow))
	l.add(c.cacheLabelSelectors, "cache-label-selectors", is("true"))
	l.add(c.qps != nil, "kube-api-qps", func() string { return formatFloat(*c.qps) })
	l.add(c.burst != 10, "kube-api-burst", itoa(c.burst))
	l.add(c.connections != 1, "api-connections", itoa(c.connections))
	l.add(c.separateWatchConnection, "separate-watch-connection", is("true"))
	l.add(c.tracing != nil, "enable-tracing", is("true"))
	l.add(c.profilingCPU, "enable-pprof", is("true"))
	l.add(c.profilingDebug, "enable-pprof-debug", is("true"))
	l.add(c.blockProfileRate != 1000000, "pprof-block-profile-rate", itoa(c.blockProfileRate))
	l.add(c.mutexProfileFraction != 10, "pprof-mutex-profile-fraction", itoa(c.mutexProfileFraction))
	l.add(c.loggingDevelopment, "zap-devel", is("true"))
	l.add(c.loggingEncoder != "", "zap-encoder", is(c.loggingEncoder))
	l.add(c.loggingLevel != "", "zap-log-level", is(c.loggingLevel))
	l.add(c.loggingStacktraceLevel != "", "zap-stacktrace-level", is(c.loggingStacktraceLevel))
	l.add(c.loggingTimeEncoding != "", "zap-time-encoding", is(c.loggingTimeEncoding))
	if s.extensions {
		l.add(s.claim.workers != 50, "sandbox-claim-concurrent-workers", itoa(s.claim.workers))
		l.add(s.claim.warmCandidateGracePeriod != "2s", "sandbox-claim-warm-candidate-grace-period", is(s.claim.warmCandidateGracePeriod))
		l.add(s.claim.eventsDisabled, "disable-claim-events", is("true"))
		l.add(s.claim.observabilityAnnotationsDisabled, "disable-claim-observability-annotations", is("true"))
		l.add(s.warmPool.workers != 1, "sandbox-warm-pool-concurrent-workers", itoa(s.warmPool.workers))
		l.add(s.warmPool.maxBatchSize != 300, "sandbox-warm-pool-max-batch-size", itoa(s.warmPool.maxBatchSize))
		l.add(s.warmPool.replenishDelay != "0s", "sandbox-warm-pool-replenish-delay", is(s.warmPool.replenishDelay))
		l.add(s.warmPool.maxRefillRate != 0, "sandbox-warm-pool-max-refill-rate", func() string { return formatFloat(s.warmPool.maxRefillRate) })
		l.add(s.warmPool.readinessGracePeriod != "5m", "sandbox-warm-pool-readiness-grace-period", is(s.warmPool.readinessGracePeriod))
		l.add(s.warmPool.unschedulableRecheckInterval != "1m", "sandbox-warm-pool-unschedulable-recheck-interval", is(s.warmPool.unschedulableRecheckInterval))
		l.add(!s.warmPool.evictionEnabled, "enable-warm-pool-eviction", is("false"))
		l.add(s.templateWorkers != 1, "sandbox-template-concurrent-workers", itoa(s.templateWorkers))
	}
	return l
}

// routerFlags are the router's settings. The first seven are arguments
// upstream's deployment file already passes, so they always take the
// resolved value (which equals upstream's when nothing changed it); the rest
// are written only when they differ from the binary's defaults.
func routerFlags(s *settings) []flag {
	r := s.router
	var l flagList
	httpBind := ""
	if r.httpEnabled {
		httpBind = ":" + strconv.Itoa(int(r.httpPort))
	}
	l.add(true, "http-bind-address", is(httpBind))
	l.add(true, "metrics-bind-address", port(r.metricsPort))
	l.add(true, "health-probe-bind-address", port(r.healthProbePort))
	l.add(true, "cluster-domain", is(s.clusterDomain))
	l.add(true, "proxy-timeout", is(r.proxyTimeout))
	l.add(true, "upstream-max-retries", itoa(r.upstreamMaxRetries))
	l.add(true, "cache-enabled", boolean(r.cacheEnabled))
	if t := r.tls; t != nil {
		l.add(true, "https-bind-address", port(t.port))
		l.add(true, "tls-cert-file", is(vars.RouterTLSMountPath+"/tls.crt"))
		l.add(true, "tls-key-file", is(vars.RouterTLSMountPath+"/tls.key"))
		l.add(t.mtlsMode != "" && t.mtlsMode != "off", "mtls-mode", is(t.mtlsMode))
		l.add(t.clientCA != nil, "tls-client-ca-file", func() string { return vars.RouterClientCAMountPath + "/" + t.clientCA.key })
		l.add(t.minVersion != "", "tls-min-version", is(t.minVersion))
		l.add(len(t.cipherSuites) > 0, "tls-cipher-suites", joined(t.cipherSuites))
	}
	l.add(r.responseHeaderTimeout != "30s", "response-header-timeout", is(r.responseHeaderTimeout))
	l.add(r.shutdownTimeout != "30s", "shutdown-timeout", is(r.shutdownTimeout))
	l.add(r.maxRequestBodyBytes != 0, "max-request-body-bytes", is(strconv.FormatInt(r.maxRequestBodyBytes, 10)))
	l.add(r.allowLoopbackPodIP, "allow-loopback-pod-ip", is("true"))
	l.add(r.pathRoutingPrefix != "", "path-routing-prefix", is(r.pathRoutingPrefix))
	l.add(r.upstreamRetryInitialDelay != "200ms", "upstream-retry-initial-delay", is(r.upstreamRetryInitialDelay))
	l.add(r.upstreamRetryMaxDelay != "800ms", "upstream-retry-max-delay", is(r.upstreamRetryMaxDelay))
	// Both are stated whenever telemetry is configured: upstream turns them
	// on by itself once an OTLP endpoint is in its environment, so an unset
	// flag would not mean off.
	l.add(r.telemetry != nil, "enable-tracing", boolean(r.tracesEnabled))
	l.add(r.telemetry != nil, "enable-otel-metrics", boolean(r.otelMetricsEnabled))
	l.add(!r.accessLog, "access-log", is("false"))
	l.add(r.authzMode != "allow-all", "authz-mode", is(r.authzMode))
	if r.authzMode == "tokenreview" {
		l.add(r.tokenReviewTTL != "30s", "authz-tokenreview-ttl", is(r.tokenReviewTTL))
		l.add(r.tokenReviewCacheSize != 2048, "authz-tokenreview-cache-size", itoa(r.tokenReviewCacheSize))
		l.add(r.tokenReviewRequireToken, "authz-tokenreview-require-token", is("true"))
		l.add(len(r.tokenReviewAudiences) > 0, "authz-tokenreview-audiences", joined(r.tokenReviewAudiences))
	}
	l.add(r.scopedTokenV1 != nil, "authz-scoped-token-secret-file", func() string { return vars.RouterScopedTokenV1Mount + "/" + r.scopedTokenV1.key })
	l.add(r.scopedTokenV2 != nil, "authz-scoped-token-verification-keys-file", func() string { return vars.RouterScopedTokenV2Mount + "/" + r.scopedTokenV2.key })
	l.add(r.scopedTokenV1AcceptUntil != "", "authz-scoped-token-v1-accept-until", is(r.scopedTokenV1AcceptUntil))
	if ck := r.cookie; ck != nil {
		l.add(true, "authz-cookie-name", is(ck.name))
		l.add(ck.queryParam != "", "authz-cookie-query-param", is(ck.queryParam))
		l.add(ck.sameSite != "" && ck.sameSite != "lax", "authz-cookie-samesite", is(ck.sameSite))
		l.add(ck.insecure, "authz-cookie-insecure", is("true"))
		l.add(len(ck.allowedOrigins) > 0, "authz-cookie-allowed-origins", joined(ck.allowedOrigins))
		l.add(ck.trustForwardedProto, "authz-trust-forwarded-proto", is("true"))
	}
	l.add(r.cacheNamespace != "", "cache-namespace", is(r.cacheNamespace))
	return l
}
