package kubernetesflagdv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesFlagd(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesFlagd Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
		ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
	}}
}

func violations(err error) []string {
	gomega.Expect(err).NotTo(gomega.BeNil())
	var verr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &verr)).To(gomega.BeTrue())
	out := []string{}
	for _, v := range verr.Violations {
		out = append(out, v.Proto.GetRuleId(), protovalidate.FieldPathString(v.Proto.GetField()))
	}
	return out
}

func configMapSource() *KubernetesFlagdSource {
	return &KubernetesFlagdSource{Kind: &KubernetesFlagdSource_ConfigMap{ConfigMap: &KubernetesFlagdConfigMapSource{
		ConfigMapName: valueFrom(catalogkind.CatalogKind_KubernetesFlagdFlagFile, "flags", "status.outputs.config_map_name"),
		Key:           literal("flags.flagd.json"),
	}}}
}

var _ = ginkgo.Describe("KubernetesFlagd Validation Tests", func() {
	var input *KubernetesFlagd

	ginkgo.BeforeEach(func() {
		input = &KubernetesFlagd{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesFlagd",
			Metadata:   &shared.CatalogObjectMetadata{Name: "flagd"},
			Spec: &KubernetesFlagdSpec{
				Namespace: literal("feature-flags"),
				Sources:   []*KubernetesFlagdSource{configMapSource()},
			},
		}
	})

	invalid := func(want string) {
		gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement(gomega.ContainSubstring(want)))
	}

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec (one ConfigMap source) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a maximal spec (every source kind and every block) should be valid", func() {
			cm := configMapSource()
			cm.GetConfigMap().Watcher = "fileinfo"
			input.Spec = &KubernetesFlagdSpec{
				Namespace:       valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "feature-flags", "spec.name"),
				CreateNamespace: true,
				Image:           &KubernetesFlagdImage{Repository: strPtr("ghcr.io/open-feature/flagd"), Tag: strPtr("v0.17.0"), Fips: true, PullPolicy: strPtr("IfNotPresent"), PullSecretNames: []string{"mirror"}},
				Replicas:        int32Ptr(2),
				Resources:       &kubernetes.ContainerResources{Limits: &kubernetes.CpuMemory{Cpu: "1", Memory: "512Mi"}},
				Hpa:             &KubernetesFlagdHpa{Enabled: true, MinReplicas: int32Ptr(2), MaxReplicas: int32Ptr(4), TargetCpuUtilizationPercent: int32Ptr(60), TargetMemoryUtilizationPercent: int32Ptr(70)},
				Pdb:             &KubernetesFlagdPdb{Enabled: true, MaxUnavailable: "1"},
				Sources: []*KubernetesFlagdSource{
					cm,
					{Kind: &KubernetesFlagdSource_Http{Http: &KubernetesFlagdHttpSource{
						Url: "https://flags.example.com/flags.json", TimeoutSeconds: int32Ptr(10),
						Oauth:   &KubernetesFlagdHttpOAuth{ClientId: "flagd", ClientSecret: "s", TokenUrl: "https://idp.example.com/token"},
						Headers: map[string]string{"Accept": "application/json"}, SensitiveHeaders: map[string]string{"X-Api-Key": "k"},
						IntervalSeconds: int32Ptr(30), IntervalSeed: "flagd",
					}}},
					{Kind: &KubernetesFlagdSource_Grpc{Grpc: &KubernetesFlagdGrpcSource{
						Target: "flag-service:8015", Tls: true, CaCertSecret: &kubernetes.KubernetesSecretKey{Name: "flag-service-ca", Key: "ca.crt"},
						ProviderId: "flagd-web", MaxMsgSize: int32Ptr(5242880), IncrementalUpdates: true,
						Headers: map[string]string{"x-env": "prod"}, SensitiveHeaders: map[string]string{"authorization": "Bearer t"}, Selector: "flagSetId=web",
					}}},
					{Kind: &KubernetesFlagdSource_FeatureFlag{FeatureFlag: &KubernetesFlagdFeatureFlagSource{Namespace: "apps", Name: "web-flags"}}},
					{Kind: &KubernetesFlagdSource_GoogleStorage{GoogleStorage: &KubernetesFlagdObjectSource{Bucket: "flags", Object: "flags.json", IntervalSeconds: int32Ptr(60)}}},
					{Kind: &KubernetesFlagdSource_AzureBlob{AzureBlob: &KubernetesFlagdObjectSource{Bucket: "flags", Object: "flags.json"}}},
					{Kind: &KubernetesFlagdSource_S3{S3: &KubernetesFlagdObjectSource{Bucket: "flags", Object: "flags.yaml", IntervalSeed: "x"}}},
				},
				Server:                   &KubernetesFlagdServer{Port: int32Ptr(8013), ManagementPort: int32Ptr(8014), SyncPort: int32Ptr(8015), OfrepPort: int32Ptr(8016), TlsSecretName: "flagd-tls", ServiceType: strPtr("ClusterIP")},
				Evaluation:               &KubernetesFlagdEvaluation{ContextValues: map[string]string{"env": "prod"}, ContextFromHeader: map[string]string{"X-Tenant": "tenant"}, CorsOrigins: []string{"https://app.example.com"}},
				OfrepSse:                 &KubernetesFlagdOfrepSse{Enabled: boolPtr(true), InactivityDelaySeconds: int32Ptr(60), PublicUrl: "https://flags.example.com"},
				Sync:                     &KubernetesFlagdSync{HttpEnabled: boolPtr(false), DisableMetadata: true, StreamDeadline: "1h30m"},
				Log:                      &KubernetesFlagdLog{Format: strPtr("console"), Debug: true},
				Telemetry:                &KubernetesFlagdTelemetry{MetricsExporter: "otel", OtelCollectorUri: "otel-collector:4317", OtelCaCertSecret: &kubernetes.KubernetesSecretKey{Name: "otel-ca", Key: "ca.crt"}, OtelClientTlsSecretName: "otel-client-tls", OtelReloadInterval: "30m"},
				Limits:                   &KubernetesFlagdLimits{MaxRequestBodyBytes: int64Ptr(2000000), MaxRequestHeaderBytes: int64Ptr(0)},
				Scheduling:               &kubernetes.WorkloadScheduling{NodeSelector: map[string]string{"pool": "system"}, SchedulerName: "default-scheduler"},
				ServiceAccount:           &KubernetesFlagdServiceAccount{Annotations: map[string]string{"eks.amazonaws.com/role-arn": "arn:aws:iam::123:role/flagd"}},
				PodAnnotations:           map[string]string{"a": "b"},
				PodLabels:                map[string]string{"team": "platform"},
				PodSecurityContext:       &kubernetes.WorkloadPodSecurityContext{RunAsNonRoot: boolPtr(true)},
				ContainerSecurityContext: &kubernetes.WorkloadContainerSecurityContext{ReadOnlyRootFilesystem: boolPtr(true)},
				ExtraEnv:                 map[string]string{"AWS_REGION": "us-east-1"},
				ExtraEnvFromSecret:       map[string]*kubernetes.KubernetesSecretKey{"AZURE_STORAGE_KEY": {Name: "azure", Key: "key"}},
				Metrics:                  &KubernetesFlagdMetrics{ServiceMonitorEnabled: true, ServiceMonitorLabels: map[string]string{"release": "prom"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			invalid("namespace")
		})

		ginkgo.It("no sources should fail", func() {
			input.Spec.Sources = nil
			invalid("sources")
		})

		ginkgo.It("a source with no kind should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{}}
			invalid("kind")
		})

		ginkgo.It("a ConfigMap source key without a parser extension should fail", func() {
			s := configMapSource()
			s.GetConfigMap().Key = literal("flags.txt")
			input.Spec.Sources = []*KubernetesFlagdSource{s}
			invalid("spec.sources.config_map.key")
		})

		ginkgo.It("an unknown ConfigMap watcher should fail", func() {
			s := configMapSource()
			s.GetConfigMap().Watcher = "inotify"
			input.Spec.Sources = []*KubernetesFlagdSource{s}
			invalid("spec.sources.config_map.watcher")
		})

		ginkgo.It("an HTTP source polling over a day should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_Http{Http: &KubernetesFlagdHttpSource{Url: "https://x.example.com/f.json", IntervalSeconds: int32Ptr(86401)}}}}
			invalid("interval_seconds")
		})

		ginkgo.It("a gRPC CA certificate without TLS should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_Grpc{Grpc: &KubernetesFlagdGrpcSource{
				Target: "svc:8015", CaCertSecret: &kubernetes.KubernetesSecretKey{Name: "ca", Key: "ca.crt"},
			}}}}
			invalid("spec.sources.grpc.ca_needs_tls")
		})

		ginkgo.It("a FeatureFlag source without a name should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_FeatureFlag{FeatureFlag: &KubernetesFlagdFeatureFlagSource{}}}}
			invalid("name")
		})

		ginkgo.It("an object source without an object should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_S3{S3: &KubernetesFlagdObjectSource{Bucket: "b"}}}}
			invalid("object")
		})

		ginkgo.It("two ports with the same number should fail", func() {
			input.Spec.Server = &KubernetesFlagdServer{SyncPort: int32Ptr(8013)}
			invalid("spec.server.distinct_ports")
		})

		ginkgo.It("an unknown service type should fail", func() {
			input.Spec.Server = &KubernetesFlagdServer{ServiceType: strPtr("ExternalName")}
			invalid("spec.server.service_type")
		})

		ginkgo.It("the otel metrics exporter without a collector should fail", func() {
			input.Spec.Telemetry = &KubernetesFlagdTelemetry{MetricsExporter: "otel"}
			invalid("spec.telemetry.otel_needs_collector")
		})

		ginkgo.It("an unknown metrics exporter should fail", func() {
			input.Spec.Telemetry = &KubernetesFlagdTelemetry{MetricsExporter: "statsd"}
			invalid("spec.telemetry.metrics_exporter")
		})

		ginkgo.It("a malformed stream deadline should fail", func() {
			input.Spec.Sync = &KubernetesFlagdSync{StreamDeadline: "an hour"}
			invalid("spec.sync.stream_deadline")
		})

		ginkgo.It("an unknown log format should fail", func() {
			input.Spec.Log = &KubernetesFlagdLog{Format: strPtr("text")}
			invalid("spec.log.format")
		})

		ginkgo.It("a relative OFREP SSE public URL should fail", func() {
			input.Spec.OfrepSse = &KubernetesFlagdOfrepSse{PublicUrl: "flags.example.com"}
			invalid("spec.ofrep_sse.public_url")
		})

		ginkgo.It("a negative request body limit should fail", func() {
			input.Spec.Limits = &KubernetesFlagdLimits{MaxRequestBodyBytes: int64Ptr(-1)}
			invalid("max_request_body_bytes")
		})

		ginkgo.It("a FLAGD_ extra variable should fail", func() {
			input.Spec.ExtraEnv = map[string]string{"FLAGD_PORT": "9000"}
			invalid("spec.extra_env.no_flagd_overrides")
		})

		ginkgo.It("a FLAGD_ variable from a Secret should fail", func() {
			input.Spec.ExtraEnvFromSecret = map[string]*kubernetes.KubernetesSecretKey{"FLAGD_SOURCES": {Name: "s", Key: "k"}}
			invalid("spec.extra_env_from_secret.no_flagd_overrides")
		})

		ginkgo.It("a disruption budget minimum that is not a count or percentage should fail", func() {
			input.Spec.Pdb = &KubernetesFlagdPdb{Enabled: true, MinAvailable: "half"}
			invalid("spec.pdb.min_available")
		})

		ginkgo.It("a disruption budget maximum above 100% should fail", func() {
			input.Spec.Pdb = &KubernetesFlagdPdb{Enabled: true, MaxUnavailable: "150%"}
			invalid("spec.pdb.max_unavailable")
		})

		ginkgo.It("an OAuth token URL that is not http(s) should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_Http{Http: &KubernetesFlagdHttpSource{
				Url:   "https://x.example.com/f.json",
				Oauth: &KubernetesFlagdHttpOAuth{ClientId: "c", ClientSecret: "s", TokenUrl: "idp.example.com/token"},
			}}}}
			invalid("spec.sources.http.oauth.token_url")
		})

		ginkgo.It("a header mapping with an equals sign should fail", func() {
			input.Spec.Evaluation = &KubernetesFlagdEvaluation{ContextFromHeader: map[string]string{"X-Tenant": "tenant=id"}}
			invalid("spec.evaluation.context_from_header")
		})

		ginkgo.It("an OpenTelemetry reload interval that is not a duration should fail", func() {
			input.Spec.Telemetry = &KubernetesFlagdTelemetry{OtelReloadInterval: "hourly"}
			invalid("spec.telemetry.otel_reload_interval")
		})

		ginkgo.It("an object-storage source without a flag-file extension should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_S3{S3: &KubernetesFlagdObjectSource{Bucket: "b", Object: "flags.txt"}}}}
			invalid("spec.sources.object.extension")
		})

		ginkgo.It("an object-storage source without an extension should pass", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_GoogleStorage{GoogleStorage: &KubernetesFlagdObjectSource{Bucket: "b", Object: "flags.d/prod"}}}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an object-storage source ending in .YAML should pass", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_S3{S3: &KubernetesFlagdObjectSource{Bucket: "b", Object: "flags/prod.YAML"}}}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an HTTP source with both an auth header and OAuth should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_Http{Http: &KubernetesFlagdHttpSource{
				Url: "https://x.example.com/f.json", AuthHeader: "Bearer t",
				Oauth: &KubernetesFlagdHttpOAuth{ClientId: "c", ClientSecret: "s", TokenUrl: "https://idp.example.com/token"},
			}}}}
			invalid("spec.sources.http.one_auth")
		})

		ginkgo.It("a non-http source URL should fail", func() {
			input.Spec.Sources = []*KubernetesFlagdSource{{Kind: &KubernetesFlagdSource_Http{Http: &KubernetesFlagdHttpSource{Url: "ftp://x.example.com/f.json"}}}}
			invalid("spec.sources.http.url")
		})

		ginkgo.It("a context value with a comma should fail", func() {
			input.Spec.Evaluation = &KubernetesFlagdEvaluation{ContextValues: map[string]string{"regions": "us,eu"}}
			invalid("spec.evaluation.context_values")
		})

		ginkgo.It("the same variable in both env maps should fail", func() {
			input.Spec.ExtraEnv = map[string]string{"AWS_REGION": "us-east-1"}
			input.Spec.ExtraEnvFromSecret = map[string]*kubernetes.KubernetesSecretKey{"AWS_REGION": {Name: "s", Key: "k"}}
			invalid("spec.extra_env.unique_names")
		})

		ginkgo.It("an autoscaler minimum above the default maximum should fail", func() {
			input.Spec.Hpa = &KubernetesFlagdHpa{Enabled: true, MinReplicas: int32Ptr(20)}
			invalid("spec.hpa.replica_bounds")
		})

		ginkgo.It("an autoscaler minimum above its maximum should fail", func() {
			input.Spec.Hpa = &KubernetesFlagdHpa{Enabled: true, MinReplicas: int32Ptr(5), MaxReplicas: int32Ptr(2)}
			invalid("spec.hpa.replica_bounds")
		})

		ginkgo.It("a PDB with both bounds should fail", func() {
			input.Spec.Pdb = &KubernetesFlagdPdb{Enabled: true, MinAvailable: "1", MaxUnavailable: "1"}
			invalid("spec.pdb.one_bound")
		})

		ginkgo.It("an unknown image pull policy should fail", func() {
			input.Spec.Image = &KubernetesFlagdImage{PullPolicy: strPtr("Sometimes")}
			invalid("spec.image.pull_policy")
		})
	})
})
