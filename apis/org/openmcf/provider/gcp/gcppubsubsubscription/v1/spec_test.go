package gcppubsubsubscriptionv1

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
	ginkgo.RunSpecs(t, "GcpPubSubSubscriptionSpec Suite")
}

var _ = ginkgo.Describe("GcpPubSubSubscriptionSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpPubSubSubscription.
	minimal := func() *GcpPubSubSubscription {
		return &GcpPubSubSubscription{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpPubSubSubscription",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-pubsub-subscription",
			},
			Spec: &GcpPubSubSubscriptionSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				SubscriptionName: "my-test-subscription",
				Topic: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/my-gcp-project/topics/my-topic",
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (project_id + subscription_name + topic)", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept subscription_name at minimum boundary (3 chars)", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept subscription_name at maximum boundary (255 chars)", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "a" + strings.Repeat("b", 254)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept subscription_name with hyphens, underscores, dots, tildes", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "events.v2~staging-sub_01"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with ack_deadline_seconds at minimum (10)", func() {
		msg := minimal()
		msg.Spec.AckDeadlineSeconds = 10
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with ack_deadline_seconds at maximum (600)", func() {
		msg := minimal()
		msg.Spec.AckDeadlineSeconds = 600
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with ack_deadline_seconds at zero (default)", func() {
		msg := minimal()
		msg.Spec.AckDeadlineSeconds = 0
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with message_retention_duration", func() {
		msg := minimal()
		msg.Spec.MessageRetentionDuration = "604800s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with retain_acked_messages", func() {
		msg := minimal()
		msg.Spec.RetainAckedMessages = true
		msg.Spec.MessageRetentionDuration = "604800s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with expiration_policy (finite TTL)", func() {
		msg := minimal()
		msg.Spec.ExpirationPolicy = &GcpPubSubSubscriptionExpirationPolicy{
			Ttl: "2592000s",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with expiration_policy (never expires)", func() {
		msg := minimal()
		msg.Spec.ExpirationPolicy = &GcpPubSubSubscriptionExpirationPolicy{
			Ttl: "",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with filter", func() {
		msg := minimal()
		msg.Spec.Filter = `attributes.type = "important"`
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with enable_message_ordering", func() {
		msg := minimal()
		msg.Spec.EnableMessageOrdering = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with enable_exactly_once_delivery", func() {
		msg := minimal()
		msg.Spec.EnableExactlyOnceDelivery = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with push_config and OIDC token", func() {
		msg := minimal()
		msg.Spec.PushConfig = &GcpPubSubSubscriptionPushConfig{
			PushEndpoint: "https://example.com/push",
			OidcToken: &GcpPubSubSubscriptionPushConfigOidcToken{
				ServiceAccountEmail: "push-sa@my-gcp-project.iam.gserviceaccount.com",
				Audience:            "https://example.com",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with push_config and no_wrapper", func() {
		msg := minimal()
		msg.Spec.PushConfig = &GcpPubSubSubscriptionPushConfig{
			PushEndpoint: "https://example.com/webhook",
			NoWrapper: &GcpPubSubSubscriptionPushConfigNoWrapper{
				WriteMetadata: true,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with bigquery_config using topic schema", func() {
		msg := minimal()
		msg.Spec.BigqueryConfig = &GcpPubSubSubscriptionBigQueryConfig{
			Table:          "my-project.my_dataset.my_table",
			UseTopicSchema: true,
			WriteMetadata:  true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with bigquery_config using table schema", func() {
		msg := minimal()
		msg.Spec.BigqueryConfig = &GcpPubSubSubscriptionBigQueryConfig{
			Table:             "my-project.my_dataset.my_table",
			UseTableSchema:    true,
			DropUnknownFields: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with cloud_storage_config", func() {
		msg := minimal()
		msg.Spec.CloudStorageConfig = &GcpPubSubSubscriptionCloudStorageConfig{
			Bucket: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-archive-bucket",
				},
			},
			FilenamePrefix: "pubsub/",
			FilenameSuffix: ".json",
			MaxBytes:       10485760,
			MaxDuration:    "300s",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with cloud_storage_config and avro_config", func() {
		msg := minimal()
		msg.Spec.CloudStorageConfig = &GcpPubSubSubscriptionCloudStorageConfig{
			Bucket: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-avro-bucket",
				},
			},
			AvroConfig: &GcpPubSubSubscriptionCloudStorageConfigAvroConfig{
				UseTopicSchema: true,
				WriteMetadata:  true,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with dead_letter_policy", func() {
		msg := minimal()
		msg.Spec.DeadLetterPolicy = &GcpPubSubSubscriptionDeadLetterPolicy{
			DeadLetterTopic: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/my-gcp-project/topics/my-dead-letter-topic",
				},
			},
			MaxDeliveryAttempts: 10,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with dead_letter_policy at min attempts (5)", func() {
		msg := minimal()
		msg.Spec.DeadLetterPolicy = &GcpPubSubSubscriptionDeadLetterPolicy{
			DeadLetterTopic: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/my-gcp-project/topics/dlq",
				},
			},
			MaxDeliveryAttempts: 5,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with dead_letter_policy at max attempts (100)", func() {
		msg := minimal()
		msg.Spec.DeadLetterPolicy = &GcpPubSubSubscriptionDeadLetterPolicy{
			DeadLetterTopic: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/my-gcp-project/topics/dlq",
				},
			},
			MaxDeliveryAttempts: 100,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with retry_policy", func() {
		msg := minimal()
		msg.Spec.RetryPolicy = &GcpPubSubSubscriptionRetryPolicy{
			MinimumBackoff: "10s",
			MaximumBackoff: "600s",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully-loaded pull subscription", func() {
		msg := minimal()
		msg.Spec.AckDeadlineSeconds = 60
		msg.Spec.MessageRetentionDuration = "1209600s"
		msg.Spec.RetainAckedMessages = true
		msg.Spec.ExpirationPolicy = &GcpPubSubSubscriptionExpirationPolicy{Ttl: ""}
		msg.Spec.Filter = `attributes.priority = "high"`
		msg.Spec.EnableMessageOrdering = true
		msg.Spec.EnableExactlyOnceDelivery = true
		msg.Spec.DeadLetterPolicy = &GcpPubSubSubscriptionDeadLetterPolicy{
			DeadLetterTopic: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/topics/dlq",
				},
			},
			MaxDeliveryAttempts: 20,
		}
		msg.Spec.RetryPolicy = &GcpPubSubSubscriptionRetryPolicy{
			MinimumBackoff: "30s",
			MaximumBackoff: "300s",
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

	ginkgo.It("should reject when subscription_name is empty", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject subscription_name shorter than 3 characters", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject subscription_name longer than 255 characters", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "a" + strings.Repeat("b", 255)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject subscription_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "1my-sub"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject subscription_name starting with a hyphen", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "-my-sub"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject subscription_name with spaces", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "my invalid sub"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject subscription_name with special characters (hash)", func() {
		msg := minimal()
		msg.Spec.SubscriptionName = "my#sub"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when topic is missing", func() {
		msg := minimal()
		msg.Spec.Topic = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject ack_deadline_seconds below minimum (9)", func() {
		msg := minimal()
		msg.Spec.AckDeadlineSeconds = 9
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject ack_deadline_seconds above maximum (601)", func() {
		msg := minimal()
		msg.Spec.AckDeadlineSeconds = 601
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject filter longer than 256 bytes", func() {
		msg := minimal()
		msg.Spec.Filter = strings.Repeat("a", 257)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject max_delivery_attempts below minimum (4)", func() {
		msg := minimal()
		msg.Spec.DeadLetterPolicy = &GcpPubSubSubscriptionDeadLetterPolicy{
			DeadLetterTopic: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/topics/dlq",
				},
			},
			MaxDeliveryAttempts: 4,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject max_delivery_attempts above maximum (101)", func() {
		msg := minimal()
		msg.Spec.DeadLetterPolicy = &GcpPubSubSubscriptionDeadLetterPolicy{
			DeadLetterTopic: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/topics/dlq",
				},
			},
			MaxDeliveryAttempts: 101,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject push_config without push_endpoint", func() {
		msg := minimal()
		msg.Spec.PushConfig = &GcpPubSubSubscriptionPushConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject push_config with OIDC token missing service_account_email", func() {
		msg := minimal()
		msg.Spec.PushConfig = &GcpPubSubSubscriptionPushConfig{
			PushEndpoint: "https://example.com/push",
			OidcToken: &GcpPubSubSubscriptionPushConfigOidcToken{
				Audience: "https://example.com",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject bigquery_config without table", func() {
		msg := minimal()
		msg.Spec.BigqueryConfig = &GcpPubSubSubscriptionBigQueryConfig{
			UseTopicSchema: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject bigquery_config with both use_topic_schema and use_table_schema", func() {
		msg := minimal()
		msg.Spec.BigqueryConfig = &GcpPubSubSubscriptionBigQueryConfig{
			Table:          "my-project.dataset.table",
			UseTopicSchema: true,
			UseTableSchema: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cloud_storage_config without bucket", func() {
		msg := minimal()
		msg.Spec.CloudStorageConfig = &GcpPubSubSubscriptionCloudStorageConfig{
			FilenamePrefix: "prefix/",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject push_config and bigquery_config both set (mutual exclusion)", func() {
		msg := minimal()
		msg.Spec.PushConfig = &GcpPubSubSubscriptionPushConfig{
			PushEndpoint: "https://example.com/push",
		}
		msg.Spec.BigqueryConfig = &GcpPubSubSubscriptionBigQueryConfig{
			Table: "project.dataset.table",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject push_config and cloud_storage_config both set (mutual exclusion)", func() {
		msg := minimal()
		msg.Spec.PushConfig = &GcpPubSubSubscriptionPushConfig{
			PushEndpoint: "https://example.com/push",
		}
		msg.Spec.CloudStorageConfig = &GcpPubSubSubscriptionCloudStorageConfig{
			Bucket: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-bucket",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject bigquery_config and cloud_storage_config both set (mutual exclusion)", func() {
		msg := minimal()
		msg.Spec.BigqueryConfig = &GcpPubSubSubscriptionBigQueryConfig{
			Table: "project.dataset.table",
		}
		msg.Spec.CloudStorageConfig = &GcpPubSubSubscriptionCloudStorageConfig{
			Bucket: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-bucket",
				},
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
