package module

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	v1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandboxcontroller/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The composition is pinned here against upstream's own v1.0.7 files,
// copied verbatim into testdata/v1.0.7 (the release assets without their
// resource definitions, which the kept-CRD primitive applies, and the
// router's deployment files). TestFixturesAreThePinnedRelease proves the
// copies still equal what upstream serves. The pins: an empty spec installs
// upstream as published plus only the documented additions; every typed
// setting reaches exactly its argument, port, mount or rule; and the router's
// files follow the settings.

func fixture(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", vars.Release, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func fixtureFiles(t *testing.T, s *settings) upstreamFiles {
	t.Helper()
	files := upstreamFiles{release: fixture(t, releaseAsset(s.extensions)), router: map[string]string{}}
	for _, file := range routerFilesFor(s) {
		files.router[file] = fixture(t, filepath.Join("router", file))
	}
	return files
}

func composeSpec(t *testing.T, sp *v1.KubernetesAgentSandboxControllerSpec) (*settings, *composition) {
	t.Helper()
	s := resolveSettings(sp)
	c, err := compose(s, fixtureFiles(t, s))
	if err != nil {
		t.Fatal(err)
	}
	return s, c
}

func byID(c *composition) map[string]document {
	out := map[string]document{}
	for _, d := range c.workloads {
		out[documentID(d)] = d
	}
	return out
}

func asJSON(v interface{}) string {
	raw, _ := json.MarshalIndent(v, "", "  ")
	return string(raw)
}

func controllerContainer(t *testing.T, c *composition) document {
	t.Helper()
	d := byID(c)["apps/v1//Deployment//agent-sandbox-controller//agent-sandbox-system"]
	container, err := containerNamed(lookup(d, "spec", "template", "spec").(map[string]interface{}), vars.ControllerContainerName)
	if err != nil {
		t.Fatal(err)
	}
	return container
}

func routerContainer(t *testing.T, c *composition) document {
	t.Helper()
	d := byID(c)["apps/v1//Deployment//sandbox-router//agent-sandbox-system"]
	container, err := containerNamed(lookup(d, "spec", "template", "spec").(map[string]interface{}), vars.RouterContainerName)
	if err != nil {
		t.Fatal(err)
	}
	return container
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func secretKeyOf(name, key string) *v1.KubernetesAgentSandboxControllerSecretKey {
	return &v1.KubernetesAgentSandboxControllerSecretKey{Secret: literal(name), Key: key}
}

func i32(i int32) *int32     { return &i }
func b(v bool) *bool         { return &v }
func sp(v string) *string    { return &v }
func f64(f float64) *float64 { return &f }

func TestCreatedNamespaceIsUpstreamsWithTheIdentityLabels(t *testing.T) {
	s := resolveSettings(&v1.KubernetesAgentSandboxControllerSpec{Namespace: literal(vars.Namespace), CreateNamespace: true})
	s.identityLabels = identityLabels(&shared.CatalogObjectMetadata{Name: "agent-sandbox", Org: "acme", Env: "dev"})
	c, err := compose(s, fixtureFiles(t, s))
	if err != nil {
		t.Fatal(err)
	}
	if documentID(c.namespace) != "v1//Namespace//agent-sandbox-system" {
		t.Fatalf("namespace document: %s", documentID(c.namespace))
	}
	want := map[string]string{
		"planton.ai/resource":     "true",
		"planton.ai/name":         "agent-sandbox",
		"planton.ai/kind":         kindName,
		"planton.ai/organization": "acme",
		"planton.ai/environment":  "dev",
	}
	for key, value := range want {
		if lookup(c.namespace, "metadata", "labels", key) != value {
			t.Errorf("the created namespace must carry %s=%s:\n%s", key, value, asJSON(c.namespace))
		}
	}
}

func TestEmptySpecInstallsUpstreamAsPublished(t *testing.T) {
	_, c := composeSpec(t, &v1.KubernetesAgentSandboxControllerSpec{})
	upstream, err := splitDocuments(fixture(t, "sandbox.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	var expectedIDs, gotIDs []string
	for _, d := range upstream {
		if str(d, "kind") != "Namespace" {
			expectedIDs = append(expectedIDs, documentID(d))
		}
	}
	for _, d := range c.workloads {
		gotIDs = append(gotIDs, documentID(d))
	}
	if !reflect.DeepEqual(expectedIDs, gotIDs) {
		t.Fatalf("an empty spec must apply exactly upstream's objects:\nwant %v\ngot  %v", expectedIDs, gotIDs)
	}
	if c.namespace != nil {
		t.Fatalf("without create_namespace the namespace is its owner's, never applied here: %s", asJSON(c.namespace))
	}

	// Every object but the controller Deployment is upstream's, byte for byte.
	got := byID(c)
	for _, want := range upstream {
		id := documentID(want)
		if str(want, "kind") == "Namespace" || id == "apps/v1//Deployment//agent-sandbox-controller//agent-sandbox-system" {
			continue
		}
		if asJSON(want) != asJSON(got[id]) {
			t.Errorf("%s changed with an empty spec:\n%s", id, asJSON(got[id]))
		}
	}

	// The controller Deployment differs only by the documented additions:
	// the ownership label and the probes on upstream's own health endpoints.
	container := controllerContainer(t, c)
	if !reflect.DeepEqual(stringList(container["args"]), []string{"--leader-elect=true"}) {
		t.Fatalf("an empty spec must keep upstream's arguments: %v", container["args"])
	}
	if container["image"] != "registry.k8s.io/agent-sandbox/agent-sandbox-controller:v1.0.7" {
		t.Fatalf("image: %v", container["image"])
	}
	if lookup(container, "readinessProbe", "httpGet", "path") != "/readyz" || lookup(container, "livenessProbe", "httpGet", "path") != "/healthz" {
		t.Fatalf("the controller must carry probes on its own health endpoints: %s", asJSON(container))
	}
	deployment := got["apps/v1//Deployment//agent-sandbox-controller//agent-sandbox-system"]
	if lookup(deployment, "metadata", "labels", "planton.ai/kind") != "KubernetesAgentSandboxController" {
		t.Fatal("the controller Deployment must carry the ownership label the ownership check reads")
	}
	if _, ok := container["env"]; ok {
		t.Fatal("no environment without tracing")
	}
}

func TestExtensionsSelectTheirAssetAndTheLabelDomainConfigMap(t *testing.T) {
	_, c := composeSpec(t, &v1.KubernetesAgentSandboxControllerSpec{
		Extensions: &v1.KubernetesAgentSandboxControllerExtensions{AllowedLabelDomains: []string{"sandbox.users.io", "example.com"}},
	})
	container := controllerContainer(t, c)
	if !reflect.DeepEqual(stringList(container["args"]), []string{"--leader-elect=true", "--extensions"}) {
		t.Fatalf("extensions with default tuning keep upstream's arguments: %v", container["args"])
	}
	cm := byID(c)["v1//ConfigMap//agent-sandbox-config//agent-sandbox-system"]
	if lookup(cm, "data", "allowed-label-domains") != "sandbox.users.io\nexample.com" {
		t.Fatalf("the label domains must reach the ConfigMap upstream reads: %s", asJSON(cm))
	}
	deployment := byID(c)["apps/v1//Deployment//agent-sandbox-controller//agent-sandbox-system"]
	if lookup(deployment, "spec", "template", "metadata", "annotations", vars.ConfigChecksumAnnotation) != configDigest("sandbox.users.io\nexample.com") {
		t.Fatal("a change to the label domains must change the pod template, or upstream (which reads them at startup) never sees it")
	}

	_, withoutDomains := composeSpec(t, &v1.KubernetesAgentSandboxControllerSpec{Extensions: &v1.KubernetesAgentSandboxControllerExtensions{}})
	if _, ok := byID(withoutDomains)["v1//ConfigMap//agent-sandbox-config//agent-sandbox-system"]; ok {
		t.Fatal("without label domains the ConfigMap is upstream's default (absent)")
	}

	_, off := composeSpec(t, &v1.KubernetesAgentSandboxControllerSpec{
		Extensions: &v1.KubernetesAgentSandboxControllerExtensions{Enabled: b(false), AllowedLabelDomains: []string{"example.com"}},
	})
	if strings.Contains(asJSON(controllerContainer(t, off)["args"]), "--extensions") || byID(off)["v1//ConfigMap//agent-sandbox-config//agent-sandbox-system"] != nil {
		t.Fatal("extensions turned off install the core asset and no ConfigMap")
	}
}

func fullSpec() *v1.KubernetesAgentSandboxControllerSpec {
	return &v1.KubernetesAgentSandboxControllerSpec{
		ImageRegistry:    "mirror.example.com/k8s",
		ImagePullSecrets: []string{"mirror-pull"},
		Extensions: &v1.KubernetesAgentSandboxControllerExtensions{
			AllowedLabelDomains: []string{"sandbox.users.io"},
			Claim: &v1.KubernetesAgentSandboxControllerClaimSettings{
				ConcurrentWorkers: i32(80), WarmCandidateGracePeriod: sp("500ms"), EventsDisabled: true, ObservabilityAnnotationsDisabled: true,
			},
			WarmPool: &v1.KubernetesAgentSandboxControllerWarmPoolSettings{
				ConcurrentWorkers: i32(4), MaxBatchSize: i32(100), ReplenishDelay: sp("10s"), MaxRefillRate: f64(2.5),
				ReadinessGracePeriod: sp("10m"), UnschedulableRecheckInterval: sp("30s"), EvictionEnabled: b(false),
			},
			Template: &v1.KubernetesAgentSandboxControllerTemplateSettings{ConcurrentWorkers: i32(2)},
		},
		Controller: &v1.KubernetesAgentSandboxControllerManager{
			Replicas:       i32(2),
			Resources:      &kubernetes.ContainerResources{Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"}},
			NodeSelector:   map[string]string{"role": "platform"},
			Tolerations:    []*kubernetes.WorkloadToleration{{Key: "platform", Operator: "Exists", Effect: "NoSchedule"}},
			LeaderElection: &v1.KubernetesAgentSandboxControllerLeaderElection{Namespace: "agent-sandbox-system"},
			Metrics: &v1.KubernetesAgentSandboxControllerMetrics{
				Port: i32(8443), RestClientMetricsEnabled: b(false),
				Tls: &v1.KubernetesAgentSandboxControllerMetricsTls{
					CertificateSecret: literal("metrics-tls"),
					MinVersion:        v1.KubernetesAgentSandboxControllerTlsVersion_tls12,
					CipherSuites:      []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
				},
			},
			Sandbox:             &v1.KubernetesAgentSandboxControllerSandboxSettings{ConcurrentWorkers: i32(200), EventsDisabled: true, WriteBehindWindow: sp("200ms")},
			CacheLabelSelectors: true,
			KubeApi:             &v1.KubernetesAgentSandboxControllerKubeApi{Qps: f64(50), Burst: i32(100), Connections: i32(4), SeparateWatchConnection: true},
			Tracing: &v1.KubernetesAgentSandboxControllerTracing{
				OtlpEndpoint: "http://otel-collector.observability:4317", Insecure: true, Headers: secretKeyOf("otel", "headers"),
			},
			Profiling: &v1.KubernetesAgentSandboxControllerProfiling{CpuEnabled: true, DebugEnabled: true, BlockProfileRate: i32(1), MutexProfileFraction: i32(100)},
			Logging: &v1.KubernetesAgentSandboxControllerLogging{
				Development: true, Encoder: v1.KubernetesAgentSandboxControllerLogging_console, Level: "4",
				StacktraceLevel: v1.KubernetesAgentSandboxControllerLogging_error, TimeEncoding: v1.KubernetesAgentSandboxControllerLogging_rfc3339nano,
			},
		},
		Router: &v1.KubernetesAgentSandboxControllerRouter{
			Image:        &kubernetes.ContainerImage{Repo: "mirror.example.com/custom-router", Tag: "v1.0.7-patched", PullSecretName: "router-pull"},
			Replicas:     i32(3),
			MinAvailable: i32(2),
			Authorization: &v1.KubernetesAgentSandboxControllerRouterAuthorization{
				Mode: &v1.KubernetesAgentSandboxControllerRouterAuthorization_ScopedToken{ScopedToken: &v1.KubernetesAgentSandboxControllerRouterScopedToken{
					V1Secret: secretKeyOf("router-hmac", "secret"), VerificationKeys: secretKeyOf("router-keys", "keys.json"), V1AcceptUntil: "2026-12-31T00:00:00Z",
				}},
				Cookie: &v1.KubernetesAgentSandboxControllerRouterCookie{
					Name: "sandbox-token", QueryParam: "token", SameSite: v1.KubernetesAgentSandboxControllerRouterCookie_strict,
					AllowedOrigins: []string{"https://ide.example.com"}, TrustForwardedProto: true,
				},
			},
			Tls: &v1.KubernetesAgentSandboxControllerRouterTls{
				CertificateSecret: literal("router-tls"), HttpEnabled: b(false),
				MtlsMode: v1.KubernetesAgentSandboxControllerRouterTls_required, ClientCa: secretKeyOf("router-tls", "ca.crt"),
				MinVersion: v1.KubernetesAgentSandboxControllerTlsVersion_tls13,
			},
			Proxy: &v1.KubernetesAgentSandboxControllerRouterProxy{
				Timeout: sp("10m"), ResponseHeaderTimeout: sp("1m"), ShutdownTimeout: sp("0s"), MaxRequestBodyBytes: 10485760,
				AllowLoopbackPodIp: true, PathRoutingPrefix: "/sandboxes", UpstreamMaxRetries: i32(0),
				UpstreamRetryInitialDelay: sp("100ms"), UpstreamRetryMaxDelay: sp("2s"),
			},
			Cache:            &v1.KubernetesAgentSandboxControllerRouterCache{Namespace: "sandboxes"},
			Telemetry:        &v1.KubernetesAgentSandboxControllerRouterTelemetry{OtlpEndpoint: "http://otel-collector.observability:4317", MetricsEnabled: b(false)},
			AccessLogEnabled: b(false),
			NetworkPolicy: &v1.KubernetesAgentSandboxControllerRouterNetworkPolicy{
				ClientNamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "gateway"},
				MetricsNamespace:      sp("observability"),
				SandboxPorts:          []int32{8888, 3000},
			},
		},
	}
}

func TestEveryControllerSettingReachesItsArgument(t *testing.T) {
	_, c := composeSpec(t, fullSpec())
	container := controllerContainer(t, c)
	want := []string{
		"--leader-elect=true",
		"--extensions",
		"--leader-election-namespace=agent-sandbox-system",
		"--metrics-bind-address=:8443",
		"--enable-rest-client-metrics=false",
		"--metrics-secure-serving=true",
		"--metrics-cert-dir=/etc/agent-sandbox/metrics-tls",
		"--tls-min-version=VersionTLS12",
		"--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"--sandbox-concurrent-workers=200",
		"--disable-sandbox-events=true",
		"--sandbox-write-behind-window=200ms",
		"--cache-label-selectors=true",
		"--kube-api-qps=50",
		"--kube-api-burst=100",
		"--api-connections=4",
		"--separate-watch-connection=true",
		"--enable-tracing=true",
		"--enable-pprof=true",
		"--enable-pprof-debug=true",
		"--pprof-block-profile-rate=1",
		"--pprof-mutex-profile-fraction=100",
		"--zap-devel=true",
		"--zap-encoder=console",
		"--zap-log-level=4",
		"--zap-stacktrace-level=error",
		"--zap-time-encoding=rfc3339nano",
		"--sandbox-claim-concurrent-workers=80",
		"--sandbox-claim-warm-candidate-grace-period=500ms",
		"--disable-claim-events=true",
		"--disable-claim-observability-annotations=true",
		"--sandbox-warm-pool-concurrent-workers=4",
		"--sandbox-warm-pool-max-batch-size=100",
		"--sandbox-warm-pool-replenish-delay=10s",
		"--sandbox-warm-pool-max-refill-rate=2.5",
		"--sandbox-warm-pool-readiness-grace-period=10m",
		"--sandbox-warm-pool-unschedulable-recheck-interval=30s",
		"--enable-warm-pool-eviction=false",
		"--sandbox-template-concurrent-workers=2",
	}
	if got := stringList(container["args"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("controller arguments:\nwant %s\ngot  %s", asJSON(want), asJSON(got))
	}
	if container["image"] != "mirror.example.com/k8s/agent-sandbox/agent-sandbox-controller:v1.0.7" {
		t.Fatalf("the mirror must move the controller image: %v", container["image"])
	}
	env := asJSON(container["env"])
	for _, want := range []string{`"OTEL_EXPORTER_OTLP_ENDPOINT"`, `"http://otel-collector.observability:4317"`, `"OTEL_EXPORTER_OTLP_INSECURE"`, `"OTEL_EXPORTER_OTLP_HEADERS"`, `"name": "otel"`, `"key": "headers"`} {
		if !strings.Contains(env, want) {
			t.Errorf("tracing environment lacks %s: %s", want, env)
		}
	}
	ports := asJSON(container["ports"])
	if !strings.Contains(ports, `"containerPort": 8443`) {
		t.Errorf("the metrics port must move the container port: %s", ports)
	}
	if !strings.Contains(asJSON(container["volumeMounts"]), vars.MetricsTLSMountPath) {
		t.Error("the metrics certificate must be mounted where --metrics-cert-dir reads it")
	}
	deployment := byID(c)["apps/v1//Deployment//agent-sandbox-controller//agent-sandbox-system"]
	podSpec := lookup(deployment, "spec", "template", "spec").(map[string]interface{})
	if lookup(deployment, "spec", "replicas") != int64(2) || podSpec["nodeSelector"] == nil || podSpec["tolerations"] == nil {
		t.Errorf("replicas and placement must reach the Deployment: %s", asJSON(deployment))
	}
	if asJSON(podSpec["imagePullSecrets"]) != asJSON([]interface{}{document{"name": "mirror-pull"}}) {
		t.Errorf("pull secrets: %s", asJSON(podSpec["imagePullSecrets"]))
	}
	service := byID(c)["v1//Service//agent-sandbox-controller//agent-sandbox-system"]
	if !strings.Contains(asJSON(service), `"port": 8443`) {
		t.Errorf("the metrics Service must follow the metrics port: %s", asJSON(service))
	}
}

func TestEveryRouterSettingReachesItsArgumentPortMountAndRule(t *testing.T) {
	_, c := composeSpec(t, fullSpec())
	container := routerContainer(t, c)
	want := []string{
		"--http-bind-address=",
		"--metrics-bind-address=:9090",
		"--health-probe-bind-address=:8081",
		"--cluster-domain=cluster.local",
		"--proxy-timeout=10m",
		"--upstream-max-retries=0",
		"--cache-enabled=true",
		"--https-bind-address=:8443",
		"--tls-cert-file=/etc/sandbox-router/tls/tls.crt",
		"--tls-key-file=/etc/sandbox-router/tls/tls.key",
		"--mtls-mode=required",
		"--tls-client-ca-file=/etc/sandbox-router/client-ca/ca.crt",
		"--tls-min-version=VersionTLS13",
		"--response-header-timeout=1m",
		"--shutdown-timeout=0s",
		"--max-request-body-bytes=10485760",
		"--allow-loopback-pod-ip=true",
		"--path-routing-prefix=/sandboxes",
		"--upstream-retry-initial-delay=100ms",
		"--upstream-retry-max-delay=2s",
		"--enable-tracing=true",
		"--enable-otel-metrics=false",
		"--access-log=false",
		"--authz-mode=scoped-token",
		"--authz-scoped-token-secret-file=/etc/sandbox-router/scoped-token-v1/secret",
		"--authz-scoped-token-verification-keys-file=/etc/sandbox-router/scoped-token-v2/keys.json",
		"--authz-scoped-token-v1-accept-until=2026-12-31T00:00:00Z",
		"--authz-cookie-name=sandbox-token",
		"--authz-cookie-query-param=token",
		"--authz-cookie-samesite=strict",
		"--authz-cookie-allowed-origins=https://ide.example.com",
		"--authz-trust-forwarded-proto=true",
		"--cache-namespace=sandboxes",
	}
	if got := stringList(container["args"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("router arguments:\nwant %s\ngot  %s", asJSON(want), asJSON(got))
	}
	if container["image"] != "mirror.example.com/custom-router:v1.0.7-patched" {
		t.Fatalf("an image override wins over the mirror: %v", container["image"])
	}
	if got := asJSON(container["ports"]); !strings.Contains(got, `"proxy-tls"`) || strings.Contains(got, `"name": "proxy"`) {
		t.Errorf("TLS-only serves proxy-tls and no plain proxy port: %s", got)
	}
	mounts := asJSON(container["volumeMounts"])
	for _, path := range []string{vars.RouterTLSMountPath, vars.RouterClientCAMountPath, vars.RouterScopedTokenV1Mount, vars.RouterScopedTokenV2Mount} {
		if !strings.Contains(mounts, path) {
			t.Errorf("router mounts lack %s: %s", path, mounts)
		}
	}
	deployment := byID(c)["apps/v1//Deployment//sandbox-router//agent-sandbox-system"]
	if lookup(deployment, "spec", "template", "spec", "terminationGracePeriodSeconds") != int64(15) {
		t.Errorf("the grace period is the drain budget plus 15s: %v", lookup(deployment, "spec", "template", "spec", "terminationGracePeriodSeconds"))
	}
	if asJSON(lookup(deployment, "spec", "template", "spec", "imagePullSecrets")) != asJSON([]interface{}{document{"name": "mirror-pull"}, document{"name": "router-pull"}}) {
		t.Errorf("router pull secrets: %s", asJSON(lookup(deployment, "spec", "template", "spec", "imagePullSecrets")))
	}
	if lookup(deployment, "spec", "replicas") != int64(3) || lookup(deployment, "metadata", "labels", "planton.ai/kind") != "KubernetesAgentSandboxController" {
		t.Error("router replicas and ownership label")
	}

	service := byID(c)["v1//Service//sandbox-router-svc//agent-sandbox-system"]
	if got := asJSON(lookup(service, "spec", "ports")); got != asJSON([]interface{}{document{"name": "proxy-tls", "port": int64(8443), "targetPort": "proxy-tls", "protocol": "TCP"}}) {
		t.Errorf("TLS-only Service ports: %s", got)
	}
	if lookup(byID(c)["policy/v1//PodDisruptionBudget//sandbox-router//agent-sandbox-system"], "spec", "minAvailable") != int64(2) {
		t.Error("disruption budget")
	}

	np := byID(c)["networking.k8s.io/v1//NetworkPolicy//sandbox-router//agent-sandbox-system"]
	ingress := lookup(np, "spec", "ingress").([]interface{})
	egress := lookup(np, "spec", "egress").([]interface{})
	if got := asJSON(ingress[0]); !strings.Contains(got, `"kubernetes.io/metadata.name": "gateway"`) || !strings.Contains(got, `"port": 8443`) || strings.Contains(got, `"port": 8080`) {
		t.Errorf("proxy ingress must admit only the client namespaces on the served listeners: %s", got)
	}
	if got := asJSON(ingress[1]); !strings.Contains(got, `"observability"`) {
		t.Errorf("metrics ingress namespace: %s", got)
	}
	if got := asJSON(lookup(egress[1].(map[string]interface{}), "ports")); got != asJSON([]interface{}{document{"protocol": "TCP", "port": int64(8888)}, document{"protocol": "TCP", "port": int64(3000)}}) {
		t.Errorf("sandbox egress ports: %s", got)
	}
	upstreamPolicy, _ := splitDocuments(fixture(t, "router/networkpolicy.yaml"))
	for _, i := range []int{0, 2, 3} {
		if asJSON(egress[i]) != asJSON(lookup(upstreamPolicy[0], "spec", "egress").([]interface{})[i]) {
			t.Errorf("egress rule %d is not the module's to change", i)
		}
	}
}

func TestRouterFilesFollowTheSettings(t *testing.T) {
	defaults := resolveSettings(&v1.KubernetesAgentSandboxControllerSpec{Router: &v1.KubernetesAgentSandboxControllerRouter{
		Authorization: &v1.KubernetesAgentSandboxControllerRouterAuthorization{Mode: &v1.KubernetesAgentSandboxControllerRouterAuthorization_AllowAll{AllowAll: &v1.KubernetesAgentSandboxControllerRouterAllowAll{}}},
	}})
	if got := routerFilesFor(defaults); !reflect.DeepEqual(got, []string{"serviceaccount.yaml", "rbac.yaml", "deployment.yaml", "service.yaml", "pdb.yaml", "networkpolicy.yaml"}) {
		t.Fatalf("default router files: %v", got)
	}
	lean := resolveSettings(&v1.KubernetesAgentSandboxControllerSpec{Router: &v1.KubernetesAgentSandboxControllerRouter{
		MinAvailable:  i32(0),
		Authorization: &v1.KubernetesAgentSandboxControllerRouterAuthorization{Mode: &v1.KubernetesAgentSandboxControllerRouterAuthorization_TokenReview{TokenReview: &v1.KubernetesAgentSandboxControllerRouterTokenReview{}}},
		Cache:         &v1.KubernetesAgentSandboxControllerRouterCache{Enabled: b(false)},
		NetworkPolicy: &v1.KubernetesAgentSandboxControllerRouterNetworkPolicy{Enabled: b(false)},
	}})
	if got := routerFilesFor(lean); !reflect.DeepEqual(got, []string{"serviceaccount.yaml", "rbac-tokenreview.yaml", "deployment.yaml", "service.yaml"}) {
		t.Fatalf("no cache RBAC without the cache, the auth-delegator binding with token review, no budget at 0, no policy when off: %v", got)
	}
	if routerFilesFor(resolveSettings(&v1.KubernetesAgentSandboxControllerSpec{})) != nil {
		t.Fatal("no router, no router files")
	}

	_, c := composeSpec(t, &v1.KubernetesAgentSandboxControllerSpec{Router: &v1.KubernetesAgentSandboxControllerRouter{
		Authorization: &v1.KubernetesAgentSandboxControllerRouterAuthorization{Mode: &v1.KubernetesAgentSandboxControllerRouterAuthorization_AllowAll{AllowAll: &v1.KubernetesAgentSandboxControllerRouterAllowAll{}}},
	}})
	container := routerContainer(t, c)
	upstreamDeployment, _ := splitDocuments(fixture(t, "router/deployment.yaml"))
	upstreamContainer, _ := containerNamed(lookup(upstreamDeployment[0], "spec", "template", "spec").(map[string]interface{}), vars.RouterContainerName)
	if !reflect.DeepEqual(container["args"], upstreamContainer["args"]) {
		t.Fatalf("a router with defaults keeps upstream's arguments:\n%s", asJSON(container["args"]))
	}
	if container["image"] != "registry.k8s.io/agent-sandbox/sandbox-router-go:v1.0.7" {
		t.Fatalf("the router image is pinned to the release, never :latest: %v", container["image"])
	}
}

func TestApplyFlagsReplacesInPlaceAndAppendsInOrder(t *testing.T) {
	got := applyFlags([]string{"--leader-elect=true", "--extensions"}, []flag{{"zap-devel", "true"}, {"leader-elect", "false"}, {"cache-label-selectors", "true"}})
	want := []string{"--leader-elect=false", "--extensions", "--zap-devel=true", "--cache-label-selectors=true"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTerminationGraceAddsUpstreamsMargin(t *testing.T) {
	for in, want := range map[string]int64{"0s": 15, "30s": 45, "1m30s": 105, "500ms": 16} {
		got, err := terminationGraceSeconds(in)
		if err != nil || got != want {
			t.Errorf("%s: got %d, want %d (%v)", in, got, want, err)
		}
	}
}

func TestOutputsNameTheRouterEndpointConsumersDial(t *testing.T) {
	plain := outputValues(resolveSettings(&v1.KubernetesAgentSandboxControllerSpec{Router: &v1.KubernetesAgentSandboxControllerRouter{}}))
	if plain[OpRouterEndpoint] != "http://sandbox-router-svc.agent-sandbox-system.svc.cluster.local:8080" || plain[OpRouterServiceName] != "sandbox-router-svc" {
		t.Fatalf("plain router outputs: %v", plain)
	}
	tlsOnly := outputValues(resolveSettings(fullSpec()))
	if tlsOnly[OpRouterEndpoint] != "https://sandbox-router-svc.agent-sandbox-system.svc.cluster.local:8443" || tlsOnly[OpMetricsPort] != 8443 {
		t.Fatalf("TLS-only router outputs: %v", tlsOnly)
	}
	none := outputValues(resolveSettings(&v1.KubernetesAgentSandboxControllerSpec{}))
	if none[OpRouterEndpoint] != "" || none[OpRouterServiceName] != "" || none[OpNamespace] != "agent-sandbox-system" {
		t.Fatalf("no-router outputs: %v", none)
	}
}

// TestTerraformTwinPinsTheSameRelease holds the upstream contract equal
// across the engines offline; the gated parity test proves the rest.
func TestTerraformTwinPinsTheSameRelease(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "tf", "locals.tf"))
	if err != nil {
		t.Fatal(err)
	}
	locals := string(raw)
	for name, value := range map[string]string{
		"release":                    vars.Release,
		"bundle_name":                vars.BundleName,
		"namespace":                  vars.Namespace,
		"upstream_registry":          vars.UpstreamRegistry,
		"router_image_repo":          vars.RouterImageRepo,
		"config_checksum_annotation": vars.ConfigChecksumAnnotation,
		"metrics_tls_mount_path":     vars.MetricsTLSMountPath,
		"router_tls_mount":           vars.RouterTLSMountPath,
		"router_client_ca_mount":     vars.RouterClientCAMountPath,
		"router_scoped_v1_mount":     vars.RouterScopedTokenV1Mount,
		"router_scoped_v2_mount":     vars.RouterScopedTokenV2Mount,
	} {
		found := false
		for _, line := range strings.Split(locals, "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 3 && fields[0] == name && fields[1] == "=" {
				found = strings.Trim(fields[2], `"`) == value
				break
			}
		}
		if !found {
			t.Errorf("the Terraform module's local %s must equal the Pulumi module's %q", name, value)
		}
	}
}

// TestFixturesAreThePinnedRelease proves testdata/v1.0.7 still equals what
// upstream serves at the pin, so the hermetic tests above test the real
// install. It fetches from GitHub, so it runs when asked:
// AGENT_SANDBOX_UPSTREAM_PIN=1.
func TestFixturesAreThePinnedRelease(t *testing.T) {
	if os.Getenv("AGENT_SANDBOX_UPSTREAM_PIN") != "1" {
		t.Skip("set AGENT_SANDBOX_UPSTREAM_PIN=1 to compare the fixtures with the pinned upstream files")
	}
	for _, extensions := range []bool{false, true} {
		live, err := fetchText(ReleaseAssetURL(extensions))
		if err != nil {
			t.Fatal(err)
		}
		liveDocuments, err := splitDocuments(live)
		if err != nil {
			t.Fatal(err)
		}
		var withoutDefinitions []document
		definitions := 0
		for _, d := range liveDocuments {
			if str(d, "kind") == "CustomResourceDefinition" {
				definitions++
				continue
			}
			withoutDefinitions = append(withoutDefinitions, d)
		}
		fixtureDocuments, _ := splitDocuments(fixture(t, releaseAsset(extensions)))
		if !reflect.DeepEqual(withoutDefinitions, fixtureDocuments) {
			t.Errorf("%s at %s no longer matches the fixture", releaseAsset(extensions), vars.Release)
		}
		wantDefinitions := 1
		if extensions {
			wantDefinitions = 4
		}
		if definitions != wantDefinitions {
			t.Errorf("%s carries %d definitions, want %d", releaseAsset(extensions), definitions, wantDefinitions)
		}
	}
	for _, file := range routerFiles {
		live, err := fetchText(RouterFileURL(file))
		if err != nil {
			t.Fatal(err)
		}
		if live != fixture(t, filepath.Join("router", file)) {
			t.Errorf("router file %s at %s no longer matches the fixture", file, vars.Release)
		}
	}
}
