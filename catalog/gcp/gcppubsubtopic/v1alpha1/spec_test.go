package gcppubsubtopicv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpPubSubTopicSpec Suite")
}

var _ = ginkgo.Describe("GcpPubSubTopicSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper to build a minimal valid GcpPubSubTopic.
	minimal := func() *GcpPubSubTopic {
		return &GcpPubSubTopic{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPubSubTopic",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-pubsub-topic",
			},
			Spec: &GcpPubSubTopicSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				TopicName: "my-test-topic",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (project_id + topic_name)", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept topic_name at minimum boundary (3 chars)", func() {
		msg := minimal()
		msg.Spec.TopicName = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept topic_name at maximum boundary (255 chars)", func() {
		msg := minimal()
		msg.Spec.TopicName = "a" + strings.Repeat("b", 254)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept topic_name with hyphens", func() {
		msg := minimal()
		msg.Spec.TopicName = "my-analytics-topic"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept topic_name with underscores", func() {
		msg := minimal()
		msg.Spec.TopicName = "my_analytics_topic"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept topic_name with dots, tildes, plus, percent", func() {
		msg := minimal()
		msg.Spec.TopicName = "events.v2~staging+prod%test"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept topic_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.TopicName = "MyAnalyticsTopic"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with CMEK encryption", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with message retention duration", func() {
		msg := minimal()
		msg.Spec.MessageRetentionDuration = "604800s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with message storage policy (single region)", func() {
		msg := minimal()
		msg.Spec.MessageStoragePolicy = &GcpPubSubTopicMessageStoragePolicy{
			AllowedPersistenceRegions: []string{"us-central1"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with message storage policy and enforce_in_transit", func() {
		msg := minimal()
		msg.Spec.MessageStoragePolicy = &GcpPubSubTopicMessageStoragePolicy{
			AllowedPersistenceRegions: []string{"us-central1", "us-east1"},
			EnforceInTransit:          true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with schema settings (JSON encoding)", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema:   svr("projects/my-project/schemas/my-schema"),
			Encoding: "JSON",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with schema settings (BINARY encoding)", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema:   svr("projects/my-project/schemas/my-schema"),
			Encoding: "BINARY",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with schema settings (encoding omitted)", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema: svr("projects/my-project/schemas/my-schema"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with Cloud Storage ingestion (text format)", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			CloudStorage: &GcpPubSubTopicIngestionCloudStorage{
				Bucket: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-ingestion-bucket",
					},
				},
				InputFormat: &GcpPubSubTopicIngestionCloudStorage_TextFormat{TextFormat: &GcpPubSubTopicIngestionCloudStorageTextFormat{
					Delimiter: ",",
				}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with Cloud Storage ingestion (avro format)", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			CloudStorage: &GcpPubSubTopicIngestionCloudStorage{
				Bucket: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-avro-bucket",
					},
				},
				MatchGlob:   "*.avro",
				InputFormat: &GcpPubSubTopicIngestionCloudStorage_AvroFormat{AvroFormat: &GcpPubSubTopicIngestionCloudStorageAvroFormat{}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with AWS Kinesis ingestion", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			AwsKinesis: &GcpPubSubTopicIngestionAwsKinesis{
				StreamArn:         "arn:aws:kinesis:us-east-1:123456789012:stream/my-stream",
				ConsumerArn:       "arn:aws:kinesis:us-east-1:123456789012:stream/my-stream/consumer/my-consumer:1234567890",
				AwsRoleArn:        "arn:aws:iam::123456789012:role/pubsub-kinesis-role",
				GcpServiceAccount: svr("kinesis-ingestion@my-project.iam.gserviceaccount.com"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with Confluent Cloud ingestion", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			ConfluentCloud: &GcpPubSubTopicIngestionConfluentCloud{
				BootstrapServer:   "pkc-12345.us-central1.gcp.confluent.cloud:9092",
				Topic:             "my-confluent-topic",
				IdentityPoolId:    "projects/123456/locations/global/workloadIdentityPools/confluent-pool/providers/confluent-provider",
				GcpServiceAccount: svr("confluent-ingestion@my-project.iam.gserviceaccount.com"),
				ClusterId:         "lkc-12345",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with ingestion and platform logs", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			CloudStorage: &GcpPubSubTopicIngestionCloudStorage{
				Bucket: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-bucket",
					},
				},
				InputFormat: &GcpPubSubTopicIngestionCloudStorage_TextFormat{TextFormat: &GcpPubSubTopicIngestionCloudStorageTextFormat{}},
			},
			PlatformLogsSettings: &GcpPubSubTopicIngestionPlatformLogsSettings{
				Severity: "INFO",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with all core fields set", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/p/locations/us/keyRings/r/cryptoKeys/k",
			},
		}
		msg.Spec.MessageRetentionDuration = "86400s"
		msg.Spec.MessageStoragePolicy = &GcpPubSubTopicMessageStoragePolicy{
			AllowedPersistenceRegions: []string{"us-central1"},
			EnforceInTransit:          true,
		}
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema:   svr("projects/p/schemas/s"),
			Encoding: "JSON",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept user labels", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{
			"team":        "payments",
			"cost-center": "cc-1234",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a topic name containing goog when not a prefix", func() {
		msg := minimal()
		msg.Spec.TopicName = "my-google-events"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a message transform (javascript udf)", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{
				JavascriptUdf: &GcpPubSubTopicMessageTransformJavascriptUdf{
					FunctionName: "redactSsn",
					Code:         "function redactSsn(message, metadata) { return message; }",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a disabled message transform", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{
				JavascriptUdf: &GcpPubSubTopicMessageTransformJavascriptUdf{
					FunctionName: "normalize",
					Code:         "function normalize(message, metadata) { return message; }",
				},
				Disabled: true,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when topic_name is empty", func() {
		msg := minimal()
		msg.Spec.TopicName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject topic_name shorter than 3 characters", func() {
		msg := minimal()
		msg.Spec.TopicName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject topic_name longer than 255 characters", func() {
		msg := minimal()
		msg.Spec.TopicName = "a" + strings.Repeat("b", 255)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject topic_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.TopicName = "1my-topic"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject topic_name starting with a hyphen", func() {
		msg := minimal()
		msg.Spec.TopicName = "-my-topic"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject topic_name with spaces", func() {
		msg := minimal()
		msg.Spec.TopicName = "my invalid topic"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject topic_name with special characters (at sign)", func() {
		msg := minimal()
		msg.Spec.TopicName = "my@topic"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject topic_name with special characters (hash)", func() {
		msg := minimal()
		msg.Spec.TopicName = "my#topic"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a topic name with the reserved goog prefix", func() {
		msg := minimal()
		msg.Spec.TopicName = "goog-events"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid schema encoding", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema:   svr("projects/p/schemas/s"),
			Encoding: "XML",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a message transform missing function_name", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{
				JavascriptUdf: &GcpPubSubTopicMessageTransformJavascriptUdf{
					Code: "function f(message, metadata) { return message; }",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a message transform missing code", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{
				JavascriptUdf: &GcpPubSubTopicMessageTransformJavascriptUdf{
					FunctionName: "f",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a message transform with no transform arm", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{Disabled: true},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an ai_inference message transform", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{
				AiInference: &GcpPubSubTopicMessageTransformAiInference{
					Endpoint:            svr("projects/my-project/locations/us-central1/endpoints/1234567890"),
					ServiceAccountEmail: svr("sa@project.iam.gserviceaccount.com"),
					UnstructuredInference: &GcpPubSubTopicMessageTransformAiInferenceUnstructuredInference{
						Parameters: map[string]string{"temperature": "0.2"},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a message transform with both javascript_udf and ai_inference", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{
				JavascriptUdf: &GcpPubSubTopicMessageTransformJavascriptUdf{
					FunctionName: "f",
					Code:         "function f(message, metadata) { return message; }",
				},
				AiInference: &GcpPubSubTopicMessageTransformAiInference{
					Endpoint: svr("projects/my-project/locations/us-central1/endpoints/1234567890"),
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an ai_inference transform missing endpoint", func() {
		msg := minimal()
		msg.Spec.MessageTransforms = []*GcpPubSubTopicMessageTransform{
			{
				AiInference: &GcpPubSubTopicMessageTransformAiInference{
					ServiceAccountEmail: svr("sa@project.iam.gserviceaccount.com"),
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept schema settings pinned to a revision range", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema:          svr("projects/my-project/schemas/order-events"),
			Encoding:        "JSON",
			FirstRevisionId: svr("abc123de"),
			LastRevisionId:  svr("abc123de"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept resource manager tags", func() {
		msg := minimal()
		msg.Spec.ResourceManagerTags = map[string]string{
			"tagKeys/123456789": "tagValues/987654321",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each valid deletion_policy value", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "deletion_policy %q should be accepted", policy)
		}
	})

	ginkgo.It("should reject an invalid deletion_policy value", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject schema settings without schema field", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Encoding: "JSON",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject message storage policy with empty regions list", func() {
		msg := minimal()
		msg.Spec.MessageStoragePolicy = &GcpPubSubTopicMessageStoragePolicy{
			AllowedPersistenceRegions: []string{},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject Cloud Storage ingestion with no input format", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			CloudStorage: &GcpPubSubTopicIngestionCloudStorage{
				Bucket: svr("my-bucket"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("carries exactly one input format by construction (setting a second arm replaces the first)", func() {
		cs := &GcpPubSubTopicIngestionCloudStorage{
			Bucket:      svr("my-bucket"),
			InputFormat: &GcpPubSubTopicIngestionCloudStorage_AvroFormat{AvroFormat: &GcpPubSubTopicIngestionCloudStorageAvroFormat{}},
		}
		cs.InputFormat = &GcpPubSubTopicIngestionCloudStorage_TextFormat{TextFormat: &GcpPubSubTopicIngestionCloudStorageTextFormat{}}
		gomega.Expect(cs.GetAvroFormat()).To(gomega.BeNil())
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{CloudStorage: cs}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject Cloud Storage ingestion without bucket", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			CloudStorage: &GcpPubSubTopicIngestionCloudStorage{
				InputFormat: &GcpPubSubTopicIngestionCloudStorage_TextFormat{TextFormat: &GcpPubSubTopicIngestionCloudStorageTextFormat{}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject AWS Kinesis ingestion missing stream_arn", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			AwsKinesis: &GcpPubSubTopicIngestionAwsKinesis{
				ConsumerArn:       "arn:aws:kinesis:us-east-1:123456789012:stream/s/consumer/c:1234567890",
				AwsRoleArn:        "arn:aws:iam::123456789012:role/r",
				GcpServiceAccount: svr("sa@project.iam.gserviceaccount.com"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject Confluent Cloud ingestion missing bootstrap_server", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			ConfluentCloud: &GcpPubSubTopicIngestionConfluentCloud{
				Topic:             "my-topic",
				IdentityPoolId:    "pool-id",
				GcpServiceAccount: svr("sa@project.iam.gserviceaccount.com"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid platform logs severity", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			PlatformLogsSettings: &GcpPubSubTopicIngestionPlatformLogsSettings{
				Severity: "CRITICAL",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when metadata is missing", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when spec is missing", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
