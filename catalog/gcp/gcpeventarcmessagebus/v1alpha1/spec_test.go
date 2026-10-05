package gcpeventarcmessagebusv1alpha1

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
	ginkgo.RunSpecs(t, "GcpEventarcMessageBusSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpEventarcMessageBusSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpEventarcMessageBus {
		return &GcpEventarcMessageBus{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpEventarcMessageBus",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-bus",
			},
			Spec: &GcpEventarcMessageBusSpec{
				Location: "us-central1",
			},
		}
	}

	topicPipeline := func(id string) *GcpEventarcMessageBusPipeline {
		return &GcpEventarcMessageBusPipeline{
			PipelineId: id,
			Destination: &GcpEventarcMessageBusPipelineDestination{
				Topic: litRef("projects/p/topics/t"),
			},
		}
	}

	ginkgo.It("accepts the minimal bus", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.Context("location", func() {
		ginkgo.It("is required", func() {
			m := minimal()
			m.Spec.Location = ""
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("message_bus_id", func() {
		ginkgo.It("enforces the API's id format", func() {
			for _, v := range []string{"", "my-bus", "a", "bus-2"} {
				m := minimal()
				m.Spec.MessageBusId = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			for _, v := range []string{"My-Bus", "-bus", "bus-", "2bus"} {
				m := minimal()
				m.Spec.MessageBusId = v
				gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "value %q", v)
			}
		})
	})

	ginkgo.Context("log_severity", func() {
		ginkgo.It("accepts the provider's ValidateEnum values and rejects others", func() {
			for _, v := range []string{"", "NONE", "DEBUG", "INFO", "NOTICE", "WARNING", "ERROR", "CRITICAL", "ALERT", "EMERGENCY"} {
				m := minimal()
				m.Spec.LogSeverity = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m := minimal()
			m.Spec.LogSeverity = "VERBOSE"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("google_api_sources", func() {
		ginkgo.It("requires a format-valid source_id", func() {
			m := minimal()
			m.Spec.GoogleApiSources = []*GcpEventarcMessageBusGoogleApiSource{{}}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "missing id")

			m.Spec.GoogleApiSources[0].SourceId = "audit-logs"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.GoogleApiSources[0].SourceId = "Audit_Logs"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "bad format")
		})
	})

	ginkgo.Context("enrollments", func() {
		ginkgo.It("must reference a pipeline defined in this spec", func() {
			m := minimal()
			m.Spec.Pipelines = []*GcpEventarcMessageBusPipeline{topicPipeline("deliver-topic")}
			m.Spec.Enrollments = []*GcpEventarcMessageBusEnrollment{
				{EnrollmentId: "route-all", CelMatch: "true", Pipeline: "deliver-topic"},
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.Enrollments[0].Pipeline = "no-such-pipeline"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "dangling pipeline id")
		})

		ginkgo.It("requires enrollment_id, cel_match, and pipeline", func() {
			m := minimal()
			m.Spec.Pipelines = []*GcpEventarcMessageBusPipeline{topicPipeline("p1")}
			m.Spec.Enrollments = []*GcpEventarcMessageBusEnrollment{
				{EnrollmentId: "route-all", Pipeline: "p1"},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "missing cel_match")
		})
	})

	ginkgo.Context("pipelines", func() {
		withPipeline := func(p *GcpEventarcMessageBusPipeline) *GcpEventarcMessageBus {
			m := minimal()
			m.Spec.Pipelines = []*GcpEventarcMessageBusPipeline{p}
			return m
		}

		ginkgo.It("accepts each single destination target", func() {
			gomega.Expect(validator.Validate(withPipeline(topicPipeline("p1")))).To(gomega.Succeed(), "topic")

			p := &GcpEventarcMessageBusPipeline{
				PipelineId: "p2",
				Destination: &GcpEventarcMessageBusPipelineDestination{
					Workflow: litRef("projects/p/locations/l/workflows/w"),
				},
			}
			gomega.Expect(validator.Validate(withPipeline(p))).To(gomega.Succeed(), "workflow")

			p = &GcpEventarcMessageBusPipeline{
				PipelineId: "p3",
				Destination: &GcpEventarcMessageBusPipelineDestination{
					MessageBus: "projects/p/locations/l/messageBuses/other",
				},
			}
			gomega.Expect(validator.Validate(withPipeline(p))).To(gomega.Succeed(), "message_bus")

			p = &GcpEventarcMessageBusPipeline{
				PipelineId: "p4",
				Destination: &GcpEventarcMessageBusPipelineDestination{
					HttpEndpoint: &GcpEventarcMessageBusHttpEndpoint{
						Uri:               "https://svc.internal:8080/events",
						NetworkAttachment: "projects/p/regions/us-central1/networkAttachments/na",
					},
				},
			}
			gomega.Expect(validator.Validate(withPipeline(p))).To(gomega.Succeed(), "http_endpoint")
		})

		ginkgo.It("rejects zero and two destination targets", func() {
			p := &GcpEventarcMessageBusPipeline{
				PipelineId:  "p1",
				Destination: &GcpEventarcMessageBusPipelineDestination{},
			}
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "no target")

			p = topicPipeline("p1")
			p.Destination.Workflow = litRef("projects/p/locations/l/workflows/w")
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "two targets")
		})

		ginkgo.It("requires https and a network attachment on http endpoints", func() {
			p := &GcpEventarcMessageBusPipeline{
				PipelineId: "p1",
				Destination: &GcpEventarcMessageBusPipelineDestination{
					HttpEndpoint: &GcpEventarcMessageBusHttpEndpoint{
						Uri: "http://svc.internal:8080/events",
					},
				},
			}
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "plain http")

			p.Destination.HttpEndpoint.Uri = "https://svc.internal:8080/events"
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "missing network attachment")
		})

		ginkgo.It("allows at most one authentication mechanism", func() {
			p := topicPipeline("p1")
			p.Authentication = &GcpEventarcMessageBusPipelineAuthentication{
				GoogleOidc: &GcpEventarcMessageBusOidcAuth{ServiceAccount: litRef("sa@p.iam.gserviceaccount.com")},
				OauthToken: &GcpEventarcMessageBusOauthAuth{ServiceAccount: litRef("sa@p.iam.gserviceaccount.com")},
			}
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed())

			p.Authentication.OauthToken = nil
			gomega.Expect(validator.Validate(withPipeline(p))).To(gomega.Succeed())
		})

		ginkgo.It("requires exactly one payload format form", func() {
			p := topicPipeline("p1")
			p.InputPayloadFormat = &GcpEventarcMessageBusPayloadFormat{}
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "no form")

			p.InputPayloadFormat = &GcpEventarcMessageBusPayloadFormat{Json: true}
			gomega.Expect(validator.Validate(withPipeline(p))).To(gomega.Succeed(), "json")

			p.InputPayloadFormat = &GcpEventarcMessageBusPayloadFormat{
				Json: true,
				Avro: &GcpEventarcMessageBusSchemaFormat{SchemaDefinition: "{}"},
			}
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "two forms")
		})

		ginkgo.It("bounds the retry policy per the API's documented ranges", func() {
			p := topicPipeline("p1")
			p.RetryPolicy = &GcpEventarcMessageBusPipelineRetryPolicy{
				MaxAttempts:   100,
				MinRetryDelay: "5s",
				MaxRetryDelay: "60s",
			}
			gomega.Expect(validator.Validate(withPipeline(p))).To(gomega.Succeed())

			p.RetryPolicy.MaxAttempts = 101
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "attempts over 100")

			p.RetryPolicy.MaxAttempts = 5
			p.RetryPolicy.MinRetryDelay = "5"
			gomega.Expect(validator.Validate(withPipeline(p))).ToNot(gomega.Succeed(), "delay missing the s suffix")
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
