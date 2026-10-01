package gcppubsubtopiciammemberv1alpha1

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
	ginkgo.RunSpecs(t, "GcpPubSubTopicIamMemberSpec Suite")
}

func valueOf(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

func refTo(name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
			Name:      name,
			FieldPath: fieldPath,
		}},
	}
}

var _ = ginkgo.Describe("GcpPubSubTopicIamMemberSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpPubSubTopicIamMember {
		return &GcpPubSubTopicIamMember{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPubSubTopicIamMember",
			Metadata:   &shared.CloudResourceMetadata{Name: "sink-publisher"},
			Spec: &GcpPubSubTopicIamMemberSpec{
				Topic:  valueOf("projects/my-project/topics/audit-logs"),
				Role:   valueOf("roles/pubsub.publisher"),
				Member: valueOf("serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com"),
			},
		}
	}

	ginkgo.It("accepts a publisher grant to a service agent", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("accepts references for topic and member", func() {
		msg := minimal()
		msg.Spec.Topic = refTo("audit-logs", "status.outputs.topic_id")
		msg.Spec.Member = refTo("org-audit", "status.outputs.writer_identity")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("accepts a custom role", func() {
		msg := minimal()
		msg.Spec.Role = valueOf("projects/my-project/roles/topicPublisherLite")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("refuses a bare topic name", func() {
		msg := minimal()
		msg.Spec.Topic = valueOf("audit-logs")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("refuses a missing topic, role, or member", func() {
		for _, mutate := range []func(*GcpPubSubTopicIamMemberSpec){
			func(s *GcpPubSubTopicIamMemberSpec) { s.Topic = nil },
			func(s *GcpPubSubTopicIamMemberSpec) { s.Role = nil },
			func(s *GcpPubSubTopicIamMemberSpec) { s.Member = nil },
		} {
			msg := minimal()
			mutate(msg.Spec)
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("refuses a role that is not a role name", func() {
		msg := minimal()
		msg.Spec.Role = valueOf("pubsub.publisher")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("refuses a member without a type prefix and a deleted principal", func() {
		for _, member := range []string{"svc@my-project.iam.gserviceaccount.com", "deleted:serviceAccount:old@p.iam.gserviceaccount.com?uid=1"} {
			msg := minimal()
			msg.Spec.Member = valueOf(member)
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), member)
		}
	})
})
