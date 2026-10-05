package kubernetesgofeatureflagv1alpha1

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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestKubernetesGoFeatureFlag(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGoFeatureFlag Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
		},
	}
}

// violations lists every violation of a validation error as both its rule id
// and its field path, so a test can name either.
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

func mustStruct(m map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(m)
	gomega.Expect(err).To(gomega.BeNil())
	return s
}

func configMapRetriever() *KubernetesGoFeatureFlagRetriever {
	return &KubernetesGoFeatureFlagRetriever{
		Kind: &KubernetesGoFeatureFlagRetriever_ConfigMap{ConfigMap: &KubernetesGoFeatureFlagConfigMapRetriever{
			ConfigMapName: valueFrom(catalogkind.CatalogKind_KubernetesGoFeatureFlagFlagFile, "flags", "status.outputs.config_map_name"),
			Key:           literal("flags.goff.yaml"),
		}},
	}
}

func source(retrievers ...*KubernetesGoFeatureFlagRetriever) *KubernetesGoFeatureFlagFlagSource {
	return &KubernetesGoFeatureFlagFlagSource{Retrievers: retrievers}
}

var _ = ginkgo.Describe("KubernetesGoFeatureFlag Validation Tests", func() {
	var input *KubernetesGoFeatureFlag

	ginkgo.BeforeEach(func() {
		input = &KubernetesGoFeatureFlag{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesGoFeatureFlag",
			Metadata:   &shared.CatalogObjectMetadata{Name: "flags"},
			Spec: &KubernetesGoFeatureFlagSpec{
				Namespace: literal("feature-flags"),
				Mode:      &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(configMapRetriever())},
			},
		}
	})

	invalid := func(want string) {
		gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement(gomega.ContainSubstring(want)))
	}

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec (one ConfigMap retriever, everything else defaulted) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "feature-flags", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a maximal flag_source spec (every retriever, notifier and exporter kind) should be valid", func() {
			src := source(
				configMapRetriever(),
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_Http{Http: &KubernetesGoFeatureFlagHttpRetriever{
					Url: "https://flags.example.com/flags.yaml", Method: "REPORT", Body: "{}",
					Headers:          map[string]string{"Accept": "application/yaml"},
					SensitiveHeaders: map[string]string{"Authorization": "Bearer t"},
					TimeoutMs:        int32Ptr(5000),
				}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_Github{Github: &KubernetesGoFeatureFlagGitRetriever{
					RepositorySlug: "acme/flags", Path: "flags.yaml", Branch: "release", Token: "ghp", BaseUrl: "https://github.example.com/api/v3", TimeoutMs: int32Ptr(3000),
				}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_Gitlab{Gitlab: &KubernetesGoFeatureFlagGitRetriever{RepositorySlug: "acme/platform/flags", Path: "flags.yaml"}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_Bitbucket{Bitbucket: &KubernetesGoFeatureFlagGitRetriever{RepositorySlug: "acme/flags", Path: "flags.yaml"}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_S3{S3: &KubernetesGoFeatureFlagS3Retriever{Bucket: "flags", Item: "flags.yaml"}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_GoogleStorage{GoogleStorage: &KubernetesGoFeatureFlagGcsRetriever{Bucket: "flags", Object: "flags.yaml"}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_AzureBlobStorage{AzureBlobStorage: &KubernetesGoFeatureFlagAzureBlobRetriever{AccountName: "acme", AccountKey: "k", Container: "flags", Object: "flags.yaml"}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_Mongodb{Mongodb: &KubernetesGoFeatureFlagMongoDbRetriever{Uri: "mongodb://u:p@mongo:27017", Database: "flags", Collection: "flags"}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_Redis{Redis: &KubernetesGoFeatureFlagRedisRetriever{
					Options: &KubernetesGoFeatureFlagRedisOptions{
						Addr: "redis:6379", Network: "tcp", Username: "flags", Password: "p", Db: 1, TlsEnabled: true, Protocol: int32Ptr(3),
						ClientName: "relay", IdentitySuffix: "x", DisableIdentity: true, MaxRetries: int32Ptr(-1), MinRetryBackoffMs: int64Ptr(-1),
						MaxRetryBackoffMs: int64Ptr(512), DialTimeoutMs: int64Ptr(5000), ReadTimeoutMs: int64Ptr(-2), WriteTimeoutMs: int64Ptr(3000),
						ContextTimeoutEnabled: true, PoolFifo: true, PoolSize: int32Ptr(20), PoolTimeoutMs: int64Ptr(4000), MinIdleConns: 1, MaxIdleConns: 5,
						ConnMaxIdleTimeMs: int64Ptr(-1), ConnMaxLifetimeMs: int64Ptr(600000),
					},
					Prefix: "goff:",
				}}},
				&KubernetesGoFeatureFlagRetriever{Kind: &KubernetesGoFeatureFlagRetriever_Postgresql{Postgresql: &KubernetesGoFeatureFlagPostgresRetriever{
					Uri: "postgres://u:p@pg:5432/flags", Table: "flags", Columns: map[string]string{"flag_name": "name"},
				}}},
			)
			src.Notifiers = []*KubernetesGoFeatureFlagNotifier{
				{Kind: &KubernetesGoFeatureFlagNotifier_Slack{Slack: &KubernetesGoFeatureFlagWebhookUrlNotifier{WebhookUrl: "https://hooks.slack.com/services/x"}}},
				{Kind: &KubernetesGoFeatureFlagNotifier_MicrosoftTeams{MicrosoftTeams: &KubernetesGoFeatureFlagWebhookUrlNotifier{WebhookUrl: "https://acme.webhook.office.com/x"}}},
				{Kind: &KubernetesGoFeatureFlagNotifier_Discord{Discord: &KubernetesGoFeatureFlagWebhookUrlNotifier{WebhookUrl: "https://discord.com/api/webhooks/x"}}},
				{Kind: &KubernetesGoFeatureFlagNotifier_Webhook{Webhook: &KubernetesGoFeatureFlagWebhookNotifier{
					EndpointUrl: "https://ops.example.com/flag-changes", Secret: "s", Meta: map[string]string{"env": "prod"},
					Headers: map[string]string{"X-Source": "relay"}, SensitiveHeaders: map[string]string{"Authorization": "Bearer t"},
				}}},
			}
			file := &KubernetesGoFeatureFlagEventFileFormat{Format: "Parquet", Filename: "events-{{ .Timestamp}}.{{ .Format}}", CsvTemplate: "{{ .Key}}\n", ParquetCompressionCodec: "ZSTD"}
			src.Exporters = []*KubernetesGoFeatureFlagExporter{
				{Kind: &KubernetesGoFeatureFlagExporter_Webhook{Webhook: &KubernetesGoFeatureFlagWebhookExporter{EndpointUrl: "https://events.example.com", Secret: "s"}}, FlushIntervalMs: int64Ptr(10000), MaxEventInMemory: int64Ptr(5000), EventType: "tracking"},
				{Kind: &KubernetesGoFeatureFlagExporter_Log{Log: &KubernetesGoFeatureFlagLogExporter{LogFormat: "{{ .Key}}={{ .Value}}"}}},
				{Kind: &KubernetesGoFeatureFlagExporter_S3{S3: &KubernetesGoFeatureFlagS3Exporter{Bucket: "events", Path: "goff/", File: file}}},
				{Kind: &KubernetesGoFeatureFlagExporter_GoogleStorage{GoogleStorage: &KubernetesGoFeatureFlagGcsExporter{Bucket: "events", File: file}}},
				{Kind: &KubernetesGoFeatureFlagExporter_AzureBlobStorage{AzureBlobStorage: &KubernetesGoFeatureFlagAzureBlobExporter{AccountName: "acme", Container: "events", File: file}}},
				{Kind: &KubernetesGoFeatureFlagExporter_Sqs{Sqs: &KubernetesGoFeatureFlagSqsExporter{QueueUrl: "https://sqs.us-east-1.amazonaws.com/123/events"}}},
				{Kind: &KubernetesGoFeatureFlagExporter_Kinesis{Kinesis: &KubernetesGoFeatureFlagKinesisExporter{StreamName: "events"}}},
				{Kind: &KubernetesGoFeatureFlagExporter_Pubsub{Pubsub: &KubernetesGoFeatureFlagPubSubExporter{ProjectId: "acme", Topic: "events"}}},
				{Kind: &KubernetesGoFeatureFlagExporter_Bigquery{Bigquery: &KubernetesGoFeatureFlagBigQueryExporter{ProjectId: "acme", DatasetId: "flags", TableName: "events", GoogleCredentials: "{}", AutoMigrate: true}}},
				{Kind: &KubernetesGoFeatureFlagExporter_Kafka{Kafka: &KubernetesGoFeatureFlagKafkaExporter{
					Topic: "events", Addresses: []string{"kafka:9092"},
					Config:       mustStruct(map[string]any{"net": map[string]any{"sasl": map[string]any{"enable": true, "mechanism": "SCRAM-SHA-512", "user": "events"}}}),
					SaslPassword: "p",
				}}},
				{Kind: &KubernetesGoFeatureFlagExporter_Opentelemetry{Opentelemetry: &KubernetesGoFeatureFlagOpenTelemetryExporter{TracerName: "goff"}}},
			}
			src.FileFormat = "yaml"
			src.PollingIntervalMs = int32Ptr(15000)
			src.StartWithRetrieverError = boolPtr(true)
			src.EnablePollingJitter = true
			src.DisableNotifierOnInit = true
			src.EvaluationContextEnrichment = mustStruct(map[string]any{"region": "us-east-1"})

			input.Spec = &KubernetesGoFeatureFlagSpec{
				Namespace:        literal("feature-flags"),
				CreateNamespace:  true,
				ChartVersion:     strPtr("1.56.0"),
				Image:            &KubernetesGoFeatureFlagImage{Repository: strPtr("ghcr.io/go-feature-flag/go-feature-flag"), Tag: "v1.56.0", Fips: true, PullPolicy: strPtr("Always"), PullSecretNames: []string{"mirror"}},
				Replicas:         int32Ptr(3),
				Resources:        &kubernetes.ContainerResources{Requests: &kubernetes.CpuMemory{Cpu: "200m", Memory: "256Mi"}},
				Hpa:              &KubernetesGoFeatureFlagHpa{Enabled: true, MinReplicas: int32Ptr(2), MaxReplicas: int32Ptr(6), TargetCpuUtilizationPercent: int32Ptr(70), TargetMemoryUtilizationPercent: int32Ptr(75)},
				Pdb:              &KubernetesGoFeatureFlagPdb{Enabled: true, MinAvailable: "50%"},
				Server:           &KubernetesGoFeatureFlagServer{Port: int32Ptr(1031), MonitoringPort: int32Ptr(1032), ServiceType: strPtr("ClusterIP")},
				Log:              &KubernetesGoFeatureFlagLog{Level: strPtr("debug"), Format: strPtr("logfmt")},
				AuthorizedKeys:   &KubernetesGoFeatureFlagAuthorizedKeys{Admin: []string{"admin-key"}, Evaluation: []string{"eval-key"}},
				Mode:             &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: src},
				OfrepEventStream: &KubernetesGoFeatureFlagOfrepEventStream{BaseUrl: "https://flags.example.com", InactivityDelaySec: 120},
				Telemetry: &KubernetesGoFeatureFlagTelemetry{
					OtlpEndpoint: "http://otel-collector:4318", OtlpProtocol: "http/protobuf", ServiceName: "flags", TracesSampler: "jaeger_remote",
					ResourceAttributes: map[string]string{"deployment.environment": "prod"},
					JaegerSampler:      &KubernetesGoFeatureFlagJaegerSampler{ManagerHostPort: "http://jaeger:5778/sampling", RefreshInterval: "1m30s", MaxOperations: 100},
				},
				Swagger: &KubernetesGoFeatureFlagSwagger{Enabled: true, Host: "flags.example.com"},
				Runtime: &KubernetesGoFeatureFlagRuntime{HideBanner: true, EnablePprof: true, DisableVersionHeader: true, EnableBulkMetricFlagNames: true, DisableFlagDetailsInStream: true, ExporterCleanQueueInterval: "1m30s", EnvVariablePrefix: strPtr("GOFF_")},
				Metrics: &KubernetesGoFeatureFlagMetrics{ServiceMonitorEnabled: true, ServiceMonitorLabels: map[string]string{"release": "kube-prometheus-stack"}},
				Scheduling: &KubernetesGoFeatureFlagScheduling{
					NodeSelector: map[string]string{"pool": "system"},
					Tolerations:  []*kubernetes.WorkloadToleration{{Key: "dedicated", Operator: "Equal", Value: "system", Effect: "NoSchedule"}},
				},
				ServiceAccount:           &KubernetesGoFeatureFlagServiceAccount{Annotations: map[string]string{"iam.gke.io/gcp-service-account": "flags@acme.iam.gserviceaccount.com"}},
				PodAnnotations:           map[string]string{"sidecar.istio.io/inject": "false"},
				PodLabels:                map[string]string{"team": "platform"},
				CommonLabels:             map[string]string{"app.kubernetes.io/part-of": "flags"},
				PodSecurityContext:       &kubernetes.WorkloadPodSecurityContext{RunAsNonRoot: boolPtr(true)},
				ContainerSecurityContext: &kubernetes.WorkloadContainerSecurityContext{ReadOnlyRootFilesystem: boolPtr(true)},
				ExtraEnv:                 map[string]string{"AWS_REGION": "us-east-1"},
				ExtraEnvFromSecret:       map[string]*kubernetes.KubernetesSecretKey{"AWS_ACCESS_KEY_ID": {Name: "aws", Key: "id"}},
				HelmValues:               "extraManifests: []\n",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("flag_sets mode with keys per set should be valid", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSets{FlagSets: &KubernetesGoFeatureFlagFlagSets{Items: []*KubernetesGoFeatureFlagFlagSet{
				{Name: "web", ApiKeys: []string{"web-key"}, Source: source(configMapRetriever())},
				{Name: "mobile", ApiKeys: []string{"mobile-key-1", "mobile-key-2"}, Source: source(configMapRetriever())},
			}}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a ConfigMap retriever in another namespace should be valid", func() {
			r := configMapRetriever()
			r.GetConfigMap().Namespace = literal("flag-files")
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(r)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a negative polling interval (polling disabled) should be valid", func() {
			input.Spec.GetFlagSource().PollingIntervalMs = int32Ptr(-1)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			invalid("namespace")
		})

		ginkgo.It("no mode should fail", func() {
			input.Spec.Mode = nil
			invalid("mode")
		})

		ginkgo.It("a flag source without retrievers should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: &KubernetesGoFeatureFlagFlagSource{}}
			invalid("retrievers")
		})

		ginkgo.It("an empty flag_sets list should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSets{FlagSets: &KubernetesGoFeatureFlagFlagSets{}}
			invalid("items")
		})

		ginkgo.It("a flag set without API keys should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSets{FlagSets: &KubernetesGoFeatureFlagFlagSets{Items: []*KubernetesGoFeatureFlagFlagSet{
				{Name: "web", Source: source(configMapRetriever())},
			}}}
			invalid("api_keys")
		})

		ginkgo.It("a flag set without a source should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSets{FlagSets: &KubernetesGoFeatureFlagFlagSets{Items: []*KubernetesGoFeatureFlagFlagSet{
				{Name: "web", ApiKeys: []string{"k"}},
			}}}
			invalid("source")
		})

		ginkgo.It("a retriever with no kind should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(&KubernetesGoFeatureFlagRetriever{})}
			invalid("kind")
		})

		ginkgo.It("a ConfigMap retriever without a key should fail", func() {
			r := configMapRetriever()
			r.GetConfigMap().Key = nil
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(r)}
			invalid("key")
		})

		ginkgo.It("an HTTP retriever with a lower-case method should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(&KubernetesGoFeatureFlagRetriever{
				Kind: &KubernetesGoFeatureFlagRetriever_Http{Http: &KubernetesGoFeatureFlagHttpRetriever{Url: "https://x.example.com/f.yaml", Method: "get"}},
			})}
			invalid("spec.retriever.http.method")
		})

		ginkgo.It("a git retriever without a path should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(&KubernetesGoFeatureFlagRetriever{
				Kind: &KubernetesGoFeatureFlagRetriever_Github{Github: &KubernetesGoFeatureFlagGitRetriever{RepositorySlug: "acme/flags"}},
			})}
			invalid("path")
		})

		ginkgo.It("a git retriever with a relative base_url should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(&KubernetesGoFeatureFlagRetriever{
				Kind: &KubernetesGoFeatureFlagRetriever_Gitlab{Gitlab: &KubernetesGoFeatureFlagGitRetriever{RepositorySlug: "acme/flags", Path: "f.yaml", BaseUrl: "gitlab.local"}},
			})}
			invalid("spec.retriever.git.base_url")
		})

		ginkgo.It("a Redis retriever with RESP protocol 4 should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(&KubernetesGoFeatureFlagRetriever{
				Kind: &KubernetesGoFeatureFlagRetriever_Redis{Redis: &KubernetesGoFeatureFlagRedisRetriever{Options: &KubernetesGoFeatureFlagRedisOptions{Addr: "redis:6379", Protocol: int32Ptr(4)}}},
			})}
			invalid("spec.retriever.redis.protocol")
		})

		ginkgo.It("a Redis retriever on an unknown network should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(&KubernetesGoFeatureFlagRetriever{
				Kind: &KubernetesGoFeatureFlagRetriever_Redis{Redis: &KubernetesGoFeatureFlagRedisRetriever{Options: &KubernetesGoFeatureFlagRedisOptions{Addr: "redis:6379", Network: "udp"}}},
			})}
			invalid("spec.retriever.redis.network")
		})

		ginkgo.It("a polling interval under one second should fail", func() {
			input.Spec.GetFlagSource().PollingIntervalMs = int32Ptr(500)
			invalid("spec.flag_source.polling_interval_ms")
		})

		ginkgo.It("an unknown file format should fail", func() {
			input.Spec.GetFlagSource().FileFormat = "xml"
			invalid("spec.flag_source.file_format")
		})

		ginkgo.It("a notifier with no kind should fail", func() {
			input.Spec.GetFlagSource().Notifiers = []*KubernetesGoFeatureFlagNotifier{{}}
			invalid("kind")
		})

		ginkgo.It("a Slack notifier without a webhook URL should fail", func() {
			input.Spec.GetFlagSource().Notifiers = []*KubernetesGoFeatureFlagNotifier{{Kind: &KubernetesGoFeatureFlagNotifier_Slack{Slack: &KubernetesGoFeatureFlagWebhookUrlNotifier{}}}}
			invalid("webhook_url")
		})

		ginkgo.It("an exporter with an unknown event type should fail", func() {
			input.Spec.GetFlagSource().Exporters = []*KubernetesGoFeatureFlagExporter{{Kind: &KubernetesGoFeatureFlagExporter_Log{Log: &KubernetesGoFeatureFlagLogExporter{}}, EventType: "all"}}
			invalid("spec.exporter.event_type")
		})

		ginkgo.It("an exporter file format other than JSON, CSV or Parquet should fail", func() {
			input.Spec.GetFlagSource().Exporters = []*KubernetesGoFeatureFlagExporter{{Kind: &KubernetesGoFeatureFlagExporter_S3{S3: &KubernetesGoFeatureFlagS3Exporter{Bucket: "b", File: &KubernetesGoFeatureFlagEventFileFormat{Format: "Avro"}}}}}
			invalid("spec.exporter.file.format")
		})

		ginkgo.It("an unknown Parquet codec should fail", func() {
			input.Spec.GetFlagSource().Exporters = []*KubernetesGoFeatureFlagExporter{{Kind: &KubernetesGoFeatureFlagExporter_S3{S3: &KubernetesGoFeatureFlagS3Exporter{Bucket: "b", File: &KubernetesGoFeatureFlagEventFileFormat{ParquetCompressionCodec: "XZ"}}}}}
			invalid("spec.exporter.file.parquet_compression_codec")
		})

		ginkgo.It("a Kinesis exporter without a stream should fail", func() {
			input.Spec.GetFlagSource().Exporters = []*KubernetesGoFeatureFlagExporter{{Kind: &KubernetesGoFeatureFlagExporter_Kinesis{Kinesis: &KubernetesGoFeatureFlagKinesisExporter{}}}}
			invalid("spec.exporter.kinesis.stream")
		})

		ginkgo.It("a Kafka exporter without brokers should fail", func() {
			input.Spec.GetFlagSource().Exporters = []*KubernetesGoFeatureFlagExporter{{Kind: &KubernetesGoFeatureFlagExporter_Kafka{Kafka: &KubernetesGoFeatureFlagKafkaExporter{Topic: "t"}}}}
			invalid("addresses")
		})

		ginkgo.It("an unknown image pull policy should fail", func() {
			input.Spec.Image = &KubernetesGoFeatureFlagImage{PullPolicy: strPtr("Sometimes")}
			invalid("spec.image.pull_policy")
		})

		ginkgo.It("zero replicas should fail", func() {
			input.Spec.Replicas = int32Ptr(0)
			invalid("replicas")
		})

		ginkgo.It("an autoscaler minimum above its maximum should fail", func() {
			input.Spec.Hpa = &KubernetesGoFeatureFlagHpa{Enabled: true, MinReplicas: int32Ptr(5), MaxReplicas: int32Ptr(2)}
			invalid("spec.hpa.replica_bounds")
		})

		ginkgo.It("a PDB with both bounds should fail", func() {
			input.Spec.Pdb = &KubernetesGoFeatureFlagPdb{Enabled: true, MinAvailable: "1", MaxUnavailable: "1"}
			invalid("spec.pdb.one_bound")
		})

		ginkgo.It("a malformed PDB bound should fail", func() {
			input.Spec.Pdb = &KubernetesGoFeatureFlagPdb{Enabled: true, MinAvailable: "half"}
			invalid("spec.pdb.min_available")
		})

		ginkgo.It("the same evaluation and monitoring port should fail", func() {
			input.Spec.Server = &KubernetesGoFeatureFlagServer{Port: int32Ptr(1031), MonitoringPort: int32Ptr(1031)}
			invalid("spec.server.distinct_ports")
		})

		ginkgo.It("an unknown service type should fail", func() {
			input.Spec.Server = &KubernetesGoFeatureFlagServer{ServiceType: strPtr("ExternalName")}
			invalid("spec.server.service_type")
		})

		ginkgo.It("an unknown log level should fail", func() {
			input.Spec.Log = &KubernetesGoFeatureFlagLog{Level: strPtr("trace")}
			invalid("spec.log.level")
		})

		ginkgo.It("an unknown log format should fail", func() {
			input.Spec.Log = &KubernetesGoFeatureFlagLog{Format: strPtr("text")}
			invalid("spec.log.format")
		})

		ginkgo.It("a relative OFREP event-stream base URL should fail", func() {
			input.Spec.OfrepEventStream = &KubernetesGoFeatureFlagOfrepEventStream{BaseUrl: "flags.example.com"}
			invalid("spec.ofrep_event_stream.base_url")
		})

		ginkgo.It("an unknown OTLP protocol should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{OtlpProtocol: "udp"}
			invalid("spec.telemetry.otlp_protocol")
		})

		ginkgo.It("a malformed Jaeger refresh interval should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{JaegerSampler: &KubernetesGoFeatureFlagJaegerSampler{RefreshInterval: "1 minute"}}
			invalid("spec.telemetry.jaeger_sampler.refresh_interval")
		})

		ginkgo.It("a malformed exporter clean-queue interval should fail", func() {
			input.Spec.Runtime = &KubernetesGoFeatureFlagRuntime{ExporterCleanQueueInterval: "soon"}
			invalid("spec.runtime.exporter_clean_queue_interval")
		})

		ginkgo.It("an OTLP protocol the exporter does not support should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{OtlpProtocol: "http/json"}
			invalid("spec.telemetry.otlp_protocol")
		})

		ginkgo.It("an unknown sampler should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{TracesSampler: "sometimes"}
			invalid("spec.telemetry.traces_sampler")
		})

		ginkgo.It("a Jaeger sampler without the jaeger_remote sampler should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{TracesSampler: "always_on", JaegerSampler: &KubernetesGoFeatureFlagJaegerSampler{ManagerHostPort: "http://jaeger:5778/sampling"}}
			invalid("spec.telemetry.jaeger_needs_jaeger_remote")
		})

		ginkgo.It("a Jaeger manager given as host:port should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{TracesSampler: "jaeger_remote", JaegerSampler: &KubernetesGoFeatureFlagJaegerSampler{ManagerHostPort: "jaeger:5778"}}
			invalid("spec.telemetry.jaeger_sampler.manager_host_port")
		})

		ginkgo.It("a sampler argument without a ratio sampler should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{TracesSampler: "always_on", TracesSamplerArg: "0.5"}
			invalid("spec.telemetry.sampler_arg_needs_ratio_sampler")
		})

		ginkgo.It("a sampler argument above 1 should fail", func() {
			input.Spec.Telemetry = &KubernetesGoFeatureFlagTelemetry{TracesSampler: "traceidratio", TracesSamplerArg: "1.5"}
			invalid("spec.telemetry.traces_sampler_arg")
		})

		ginkgo.It("two flag sets with the same name should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSets{FlagSets: &KubernetesGoFeatureFlagFlagSets{Items: []*KubernetesGoFeatureFlagFlagSet{
				{Name: "web", ApiKeys: []string{"a"}, Source: source(configMapRetriever())},
				{Name: "web", ApiKeys: []string{"b"}, Source: source(configMapRetriever())},
			}}}
			invalid("spec.flag_sets.unique_names")
		})

		ginkgo.It("a flag set named default should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSets{FlagSets: &KubernetesGoFeatureFlagFlagSets{Items: []*KubernetesGoFeatureFlagFlagSet{
				{Name: "default", ApiKeys: []string{"a"}, Source: source(configMapRetriever())},
			}}}
			invalid("spec.flag_sets.reserved_name")
		})

		ginkgo.It("a non-http OFREP event-stream URL should fail", func() {
			input.Spec.OfrepEventStream = &KubernetesGoFeatureFlagOfrepEventStream{BaseUrl: "ftp://flags.example.com"}
			invalid("spec.ofrep_event_stream.base_url")
		})

		ginkgo.It("an autoscaler minimum above the default maximum should fail", func() {
			input.Spec.Hpa = &KubernetesGoFeatureFlagHpa{Enabled: true, MinReplicas: int32Ptr(20)}
			invalid("spec.hpa.replica_bounds")
		})

		ginkgo.It("a PDB minimum of zero should fail (the chart renders it as 1)", func() {
			input.Spec.Pdb = &KubernetesGoFeatureFlagPdb{Enabled: true, MinAvailable: "0"}
			invalid("spec.pdb.min_available")
		})

		ginkgo.It("a PDB percentage above 100 should fail", func() {
			input.Spec.Pdb = &KubernetesGoFeatureFlagPdb{Enabled: true, MaxUnavailable: "150%"}
			invalid("spec.pdb.max_unavailable")
		})

		ginkgo.It("an unknown Postgres column mapping should fail", func() {
			input.Spec.Mode = &KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: source(&KubernetesGoFeatureFlagRetriever{
				Kind: &KubernetesGoFeatureFlagRetriever_Postgresql{Postgresql: &KubernetesGoFeatureFlagPostgresRetriever{Uri: "postgres://x", Table: "t", Columns: map[string]string{"name": "n"}}},
			})}
			invalid("spec.retriever.postgresql.columns")
		})

		ginkgo.It("a Helm template in a pod annotation should fail", func() {
			input.Spec.PodAnnotations = map[string]string{"a": "{{ .Release.Name }}"}
			invalid("spec.pod_annotations.no_templates")
		})

		ginkgo.It("the same variable in both env maps should fail", func() {
			input.Spec.ExtraEnv = map[string]string{"AWS_REGION": "us-east-1"}
			input.Spec.ExtraEnvFromSecret = map[string]*kubernetes.KubernetesSecretKey{"AWS_REGION": {Name: "s", Key: "k"}}
			invalid("spec.extra_env.unique_names")
		})

		ginkgo.It("a lower-case env variable prefix should fail", func() {
			input.Spec.Runtime = &KubernetesGoFeatureFlagRuntime{EnvVariablePrefix: strPtr("goff_")}
			invalid("spec.runtime.env_variable_prefix")
		})
	})
})
