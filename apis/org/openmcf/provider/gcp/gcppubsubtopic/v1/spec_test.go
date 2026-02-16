package gcppubsubtopicv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
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

	// Helper to build a minimal valid GcpPubSubTopic.
	minimal := func() *GcpPubSubTopic {
		return &GcpPubSubTopic{
			ApiVersion: "gcp.openmcf.org/v1",
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
			Schema:   "projects/my-project/schemas/my-schema",
			Encoding: "JSON",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with schema settings (BINARY encoding)", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema:   "projects/my-project/schemas/my-schema",
			Encoding: "BINARY",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with schema settings (encoding omitted)", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema: "projects/my-project/schemas/my-schema",
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
				TextFormat: &GcpPubSubTopicIngestionCloudStorageTextFormat{
					Delimiter: ",",
				},
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
				MatchGlob:  "*.avro",
				AvroFormat: &GcpPubSubTopicIngestionCloudStorageAvroFormat{},
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
				GcpServiceAccount: "kinesis-ingestion@my-project.iam.gserviceaccount.com",
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
				GcpServiceAccount: "confluent-ingestion@my-project.iam.gserviceaccount.com",
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
				TextFormat: &GcpPubSubTopicIngestionCloudStorageTextFormat{},
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
			Schema:   "projects/p/schemas/s",
			Encoding: "JSON",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when project_id is missing", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
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

	ginkgo.It("should reject invalid schema encoding", func() {
		msg := minimal()
		msg.Spec.SchemaSettings = &GcpPubSubTopicSchemaSettings{
			Schema:   "projects/p/schemas/s",
			Encoding: "XML",
		}
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

	ginkgo.It("should reject Cloud Storage ingestion without bucket", func() {
		msg := minimal()
		msg.Spec.IngestionDataSourceSettings = &GcpPubSubTopicIngestionDataSourceSettings{
			CloudStorage: &GcpPubSubTopicIngestionCloudStorage{
				TextFormat: &GcpPubSubTopicIngestionCloudStorageTextFormat{},
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
				GcpServiceAccount: "sa@project.iam.gserviceaccount.com",
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
				GcpServiceAccount: "sa@project.iam.gserviceaccount.com",
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
