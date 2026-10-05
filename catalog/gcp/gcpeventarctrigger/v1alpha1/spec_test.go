package gcpeventarctriggerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpEventarcTriggerSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpEventarcTriggerSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The canonical trigger: Pub/Sub messagePublished to a Cloud Run
	// service.
	minimal := func() *GcpEventarcTrigger {
		return &GcpEventarcTrigger{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpEventarcTrigger",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-trigger",
			},
			Spec: &GcpEventarcTriggerSpec{
				Location: "us-central1",
				MatchingCriteria: []*GcpEventarcTriggerMatchingCriterion{
					{Attribute: "type", Value: "google.cloud.pubsub.topic.v1.messagePublished"},
				},
				Destination: &GcpEventarcTriggerDestination{
					CloudRunService: &GcpEventarcTriggerCloudRunDestination{
						Service: litRef("my-service"),
					},
				},
			},
		}
	}

	ginkgo.It("accepts the canonical Pub/Sub-to-Cloud-Run trigger", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.Context("location", func() {
		ginkgo.It("is required", func() {
			m := minimal()
			m.Spec.Location = ""
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("matching_criteria", func() {
		ginkgo.It("requires at least one criterion", func() {
			m := minimal()
			m.Spec.MatchingCriteria = nil
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("requires a 'type' criterion (the API's own rule)", func() {
			m := minimal()
			m.Spec.MatchingCriteria = []*GcpEventarcTriggerMatchingCriterion{
				{Attribute: "bucket", Value: "my-bucket"},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("requires attribute and value on every criterion", func() {
			m := minimal()
			m.Spec.MatchingCriteria = append(m.Spec.MatchingCriteria,
				&GcpEventarcTriggerMatchingCriterion{Attribute: "bucket"})
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("accepts only match-path-pattern as a non-empty operator", func() {
			m := minimal()
			m.Spec.MatchingCriteria[0].Operator = "match-path-pattern"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.MatchingCriteria[0].Operator = "equals"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("destination", func() {
		ginkgo.It("rejects a destination with no arm", func() {
			m := minimal()
			m.Spec.Destination = &GcpEventarcTriggerDestination{}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("rejects two arms at once", func() {
			m := minimal()
			m.Spec.Destination.Workflow = litRef("projects/p/locations/l/workflows/w")
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("accepts each single arm", func() {
			m := minimal()
			m.Spec.Destination = &GcpEventarcTriggerDestination{
				Workflow: litRef("projects/p/locations/l/workflows/w"),
			}
			// A workflow destination requires the trigger identity (API
			// truth, spec-enforced).
			m.Spec.ServiceAccount = litRef("sa@p.iam.gserviceaccount.com")
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "workflow arm")
			m.Spec.ServiceAccount = nil

			m.Spec.Destination = &GcpEventarcTriggerDestination{
				Gke: &GcpEventarcTriggerGkeDestination{
					Cluster:   litRef("my-cluster"),
					Location:  "us-central1",
					Namespace: "default",
					Service:   "my-svc",
				},
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "gke arm")

			m.Spec.Destination = &GcpEventarcTriggerDestination{
				HttpEndpoint: &GcpEventarcTriggerHttpEndpointDestination{
					Uri:               "https://svc.internal:8080/events",
					NetworkAttachment: "projects/p/regions/us-central1/networkAttachments/na",
				},
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "http arm")
		})

		ginkgo.It("http_endpoint requires its network attachment", func() {
			m := minimal()
			m.Spec.Destination = &GcpEventarcTriggerDestination{
				HttpEndpoint: &GcpEventarcTriggerHttpEndpointDestination{
					Uri: "https://svc.internal:8080/events",
				},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("gke arm requires cluster, location, namespace, and service", func() {
			m := minimal()
			m.Spec.Destination = &GcpEventarcTriggerDestination{
				Gke: &GcpEventarcTriggerGkeDestination{
					Cluster:  litRef("my-cluster"),
					Location: "us-central1",
				},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("workflow arm requires service_account (live-verified API truth)", func() {
			m := minimal()
			m.Spec.Destination = &GcpEventarcTriggerDestination{
				Workflow: litRef("projects/p/locations/l/workflows/w"),
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "bare workflow arm")

			m.Spec.ServiceAccount = litRef("sa@p.iam.gserviceaccount.com")
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "with identity")
		})

		ginkgo.It("cloud_run arm on a global trigger requires an explicit region", func() {
			m := minimal()
			m.Spec.Location = "global"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "global without region")

			m.Spec.Destination.CloudRunService.Region = "us-central1"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "global with region")
		})
	})

	ginkgo.Context("event_data_content_type", func() {
		ginkgo.It("is rejected on Pub/Sub triggers (live-verified API truth)", func() {
			m := minimal()
			m.Spec.EventDataContentType = "application/json"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "pubsub trigger with content type")
		})

		ginkgo.It("is accepted on non-Pub/Sub triggers", func() {
			m := minimal()
			m.Spec.MatchingCriteria = []*GcpEventarcTriggerMatchingCriterion{
				{Attribute: "type", Value: "google.cloud.storage.object.v1.finalized"},
				{Attribute: "bucket", Value: "my-bucket"},
			}
			m.Spec.EventDataContentType = "application/json"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})
	})

	ginkgo.Context("retry_max_attempts", func() {
		ginkgo.It("accepts only the value 1 (the provider's only valid value)", func() {
			m := minimal()
			m.Spec.RetryMaxAttempts = 1
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.RetryMaxAttempts = 3
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("is legal only with a Cloud Run destination", func() {
			m := minimal()
			m.Spec.Destination = &GcpEventarcTriggerDestination{
				Workflow: litRef("projects/p/locations/l/workflows/w"),
			}
			m.Spec.RetryMaxAttempts = 1
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("partner_channel", func() {
		ginkgo.It("requires the third-party provider", func() {
			m := minimal()
			m.Spec.PartnerChannel = &GcpEventarcTriggerPartnerChannel{}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())

			m.Spec.PartnerChannel.ThirdPartyProvider = "projects/p/locations/us-central1/providers/datadog"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})
	})

	ginkgo.Context("deletion_policy", func() {
		ginkgo.It("accepts the documented values and rejects others", func() {
			for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				m := minimal()
				m.Spec.DeletionPolicy = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m := minimal()
			m.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})
})
