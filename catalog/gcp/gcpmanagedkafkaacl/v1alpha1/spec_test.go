package gcpmanagedkafkaaclv1alpha1

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
	ginkgo.RunSpecs(t, "GcpManagedKafkaAclSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpManagedKafkaAclSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpManagedKafkaAcl {
		return &GcpManagedKafkaAcl{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpManagedKafkaAcl",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders-topic"},
			Spec: &GcpManagedKafkaAclSpec{
				Location: "us-central1",
				Cluster:  litRef("events"),
				AclId:    "topic/orders",
				AclEntries: []*GcpManagedKafkaAclEntry{
					{Principal: "User:orders-api@p.iam.gserviceaccount.com", Operation: "WRITE"},
				},
			},
		}
	}

	ginkgo.It("should accept one entry on one topic", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every resource pattern Google defines", func() {
		for _, id := range []string{"cluster", "topic/*", "consumerGroup/orders-consumer", "transactionalId/tx",
			"topicPrefixed/orders.", "consumerGroupPrefixed/orders-", "transactionalIdPrefixed/tx-"} {
			msg := minimal()
			msg.Spec.AclId = id
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), id)
		}
	})

	ginkgo.It("should accept every entry field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("data-project")
		msg.Spec.AclEntries = append(msg.Spec.AclEntries,
			&GcpManagedKafkaAclEntry{Principal: "User:*", Operation: "DESCRIBE", PermissionType: "DENY", Host: "*"})
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a pattern outside Google's grammar", func() {
		for _, id := range []string{"", "topics/orders", "group/orders", "topic/", "Cluster"} {
			msg := minimal()
			msg.Spec.AclId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
	})

	ginkgo.It("should require entries and cap them at 100", func() {
		msg := minimal()
		msg.Spec.AclEntries = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		for i := 0; i < 100; i++ {
			msg.Spec.AclEntries = append(msg.Spec.AclEntries, &GcpManagedKafkaAclEntry{Principal: "User:*", Operation: "READ"})
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a principal without User:, unknown values, and a host other than *", func() {
		msg := minimal()
		msg.Spec.AclEntries[0].Principal = "orders-api@p.iam.gserviceaccount.com"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AclEntries[0].Operation = "PRODUCE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AclEntries[0].PermissionType = "BLOCK"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AclEntries[0].Host = "10.0.0.1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a cluster and reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.Cluster = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
