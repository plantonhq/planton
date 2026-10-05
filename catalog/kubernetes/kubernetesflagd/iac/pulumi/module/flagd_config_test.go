package module

import (
	"encoding/json"
	"strings"
	"testing"

	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesflagdv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflagd/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func lit(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func TestConfigMapSourcesMountAsDirectoriesAndDefaultsServeJSONLogs(t *testing.T) {
	spec := &kubernetesflagdv1alpha1.KubernetesFlagdSpec{
		Namespace: lit("flags"),
		Sources: []*kubernetesflagdv1alpha1.KubernetesFlagdSource{
			{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_ConfigMap{ConfigMap: &kubernetesflagdv1alpha1.KubernetesFlagdConfigMapSource{
				ConfigMapName: lit("release-flags"), Key: lit("flags.flagd.json"),
			}}},
		},
	}
	cfg, err := buildFlagdConfig(spec, "flags")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"start", "--port=8013", "--management-port=8014", "--sync-port=8015", "--ofrep-port=8016", "--log-format=json"}
	if strings.Join(cfg.Args, " ") != strings.Join(want, " ") {
		t.Fatalf("args = %v, want %v", cfg.Args, want)
	}
	if cfg.SourcesJSON != `[{"provider":"file","uri":"/etc/flagd/sources/0/flags.flagd.json"}]` {
		t.Fatalf("sources = %s", cfg.SourcesJSON)
	}
	if len(cfg.ConfigMapMounts) != 1 || cfg.ConfigMapMounts[0].MountPath != "/etc/flagd/sources/0" || cfg.ConfigMapMounts[0].ConfigMap != "release-flags" {
		t.Fatalf("a ConfigMap source mounts as a directory (never subPath): %+v", cfg.ConfigMapMounts)
	}
	if len(cfg.SourcesChecksum) != 64 {
		t.Fatalf("checksum must be a sha256 hex digest, got %q", cfg.SourcesChecksum)
	}
}

func TestEverySourceKindAndSettingRenders(t *testing.T) {
	interval := int32(30)
	timeout := int32(10)
	maxMsg := int32(5242880)
	enabled := false
	delay := int32(60)
	httpOn := false
	body := int64(2000000)
	spec := &kubernetesflagdv1alpha1.KubernetesFlagdSpec{
		Namespace: lit("flags"),
		Sources: []*kubernetesflagdv1alpha1.KubernetesFlagdSource{
			{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_Http{Http: &kubernetesflagdv1alpha1.KubernetesFlagdHttpSource{
				Url: "https://x.example.com/f.json", AuthHeader: "Bearer t", Headers: map[string]string{"Accept": "application/json"},
				SensitiveHeaders: map[string]string{"X-Api-Key": "k"}, IntervalSeconds: &interval, IntervalSeed: "seed", TimeoutSeconds: &timeout,
			}}},
			{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_Grpc{Grpc: &kubernetesflagdv1alpha1.KubernetesFlagdGrpcSource{
				Target: "svc:8015", Tls: true, CaCertSecret: &kubernetesprovider.KubernetesSecretKey{Name: "ca", Key: "ca.crt"},
				ProviderId: "p", MaxMsgSize: &maxMsg, IncrementalUpdates: true, Selector: "flagSetId=web",
			}}},
			{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_FeatureFlag{FeatureFlag: &kubernetesflagdv1alpha1.KubernetesFlagdFeatureFlagSource{Name: "web-flags"}}},
			{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_GoogleStorage{GoogleStorage: &kubernetesflagdv1alpha1.KubernetesFlagdObjectSource{Bucket: "b", Object: "f.json"}}},
			{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_AzureBlob{AzureBlob: &kubernetesflagdv1alpha1.KubernetesFlagdObjectSource{Bucket: "c", Object: "f.json"}}},
			{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_S3{S3: &kubernetesflagdv1alpha1.KubernetesFlagdObjectSource{Bucket: "b", Object: "f.yaml"}}},
		},
		Server:     &kubernetesflagdv1alpha1.KubernetesFlagdServer{TlsSecretName: "flagd-tls"},
		Evaluation: &kubernetesflagdv1alpha1.KubernetesFlagdEvaluation{ContextValues: map[string]string{"env": "prod"}, ContextFromHeader: map[string]string{"X-Tenant": "tenant"}, CorsOrigins: []string{"https://a.example.com", "https://b.example.com"}},
		OfrepSse:   &kubernetesflagdv1alpha1.KubernetesFlagdOfrepSse{Enabled: &enabled, InactivityDelaySeconds: &delay, PublicUrl: "https://flags.example.com"},
		Sync:       &kubernetesflagdv1alpha1.KubernetesFlagdSync{HttpEnabled: &httpOn, DisableMetadata: true, StreamDeadline: "1h"},
		Log:        &kubernetesflagdv1alpha1.KubernetesFlagdLog{Debug: true},
		Telemetry:  &kubernetesflagdv1alpha1.KubernetesFlagdTelemetry{MetricsExporter: "otel", OtelCollectorUri: "otel:4317", OtelCaCertSecret: &kubernetesprovider.KubernetesSecretKey{Name: "otel-ca", Key: "ca.crt"}, OtelClientTlsSecretName: "otel-tls", OtelReloadInterval: "30m"},
		Limits:     &kubernetesflagdv1alpha1.KubernetesFlagdLimits{MaxRequestBodyBytes: &body},
	}
	cfg, err := buildFlagdConfig(spec, "flags")
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cfg.Args, " ")
	for _, want := range []string{
		"--debug", "--server-cert-path=/etc/flagd/tls/tls.crt", "--server-key-path=/etc/flagd/tls/tls.key",
		"--context-value=env=prod", "--context-from-header=X-Tenant=tenant", "--cors-origin=https://a.example.com,https://b.example.com",
		"--ofrep-sse-enabled=false", "--ofrep-sse-inactivity-delay=60", "--ofrep-sse-public-url=https://flags.example.com",
		"--sync-http-enabled=false", "--disable-sync-metadata", "--stream-deadline=1h",
		"--metrics-exporter=otel", "--otel-collector-uri=otel:4317", "--otel-ca-path=/etc/flagd/otel-ca/ca.crt",
		"--otel-cert-path=/etc/flagd/otel-tls/tls.crt", "--otel-key-path=/etc/flagd/otel-tls/tls.key", "--otel-reload-interval=30m",
		"--max-request-body=2000000",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("args miss %q: %s", want, args)
		}
	}
	var sources []map[string]interface{}
	if err := json.Unmarshal([]byte(cfg.SourcesJSON), &sources); err != nil {
		t.Fatal(err)
	}
	http := sources[0]
	if http["provider"] != "http" || http["authHeader"] != "Bearer t" || http["interval"].(float64) != 30 || http["intervalSeed"] != "seed" || http["timeoutS"].(float64) != 10 {
		t.Errorf("http source wrong: %v", http)
	}
	if h := http["headers"].(map[string]interface{}); h["Accept"] != "application/json" || h["X-Api-Key"] != "k" {
		t.Errorf("plain and credential headers merge: %v", h)
	}
	grpc := sources[1]
	if grpc["tls"] != true || grpc["certPath"] != "/etc/flagd/grpc-ca/1/ca.crt" || grpc["providerID"] != "p" || grpc["incrementalUpdates"] != true || grpc["selector"] != "flagSetId=web" {
		t.Errorf("grpc source wrong: %v", grpc)
	}
	if sources[2]["uri"] != "flags/web-flags" || sources[2]["provider"] != "kubernetes" {
		t.Errorf("a FeatureFlag source defaults to flagd's namespace: %v", sources[2])
	}
	if sources[3]["uri"] != "gs://b/f.json" || sources[4]["uri"] != "azblob://c/f.json" || sources[5]["uri"] != "s3://b/f.yaml" {
		t.Errorf("object sources wrong: %v %v %v", sources[3], sources[4], sources[5])
	}
	if len(cfg.FeatureFlagNamespaces) != 1 || cfg.FeatureFlagNamespaces[0] != "flags" {
		t.Errorf("feature-flag grants wrong: %v", cfg.FeatureFlagNamespaces)
	}
	if len(cfg.GrpcCaMounts) != 1 || cfg.GrpcCaMounts[0].Secret != "ca" {
		t.Errorf("grpc CA mount wrong: %+v", cfg.GrpcCaMounts)
	}
}

func TestAHeaderDeclaredPlainAndSensitiveIsRefused(t *testing.T) {
	spec := &kubernetesflagdv1alpha1.KubernetesFlagdSpec{
		Namespace: lit("flags"),
		Sources: []*kubernetesflagdv1alpha1.KubernetesFlagdSource{{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_Http{Http: &kubernetesflagdv1alpha1.KubernetesFlagdHttpSource{
			Url: "https://x.example.com/f.json", Headers: map[string]string{"authorization": "a"}, SensitiveHeaders: map[string]string{"Authorization": "b"},
		}}}},
	}
	if _, err := buildFlagdConfig(spec, "flags"); err == nil || !strings.Contains(err.Error(), "both plain and sensitive") {
		t.Fatalf("got %v", err)
	}
}

func TestTheServiceMonitorSelectsTheService(t *testing.T) {
	locals, err := initializeLocals(&kubernetesflagdv1alpha1.KubernetesFlagdIacInput{Target: &kubernetesflagdv1alpha1.KubernetesFlagd{
		Metadata: &shared.CatalogObjectMetadata{Name: "flagd"},
		Spec: &kubernetesflagdv1alpha1.KubernetesFlagdSpec{
			Namespace: lit("flags"),
			Sources: []*kubernetesflagdv1alpha1.KubernetesFlagdSource{
				{Kind: &kubernetesflagdv1alpha1.KubernetesFlagdSource_ConfigMap{ConfigMap: &kubernetesflagdv1alpha1.KubernetesFlagdConfigMapSource{
					ConfigMapName: lit("release-flags"), Key: lit("flags.flagd.json"),
				}}},
			},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	labels := serviceLabels(locals)
	for k, v := range locals.SelectorLabels {
		if labels[k] != v {
			t.Fatalf("the ServiceMonitor selects the Service by %s=%s, which the Service does not carry: %v", k, v, labels)
		}
	}
	for k, v := range locals.Labels {
		if labels[k] != v {
			t.Fatalf("the Service lost its Planton label %s: %v", k, labels)
		}
	}
}
