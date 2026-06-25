package cloudflared1databasev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func validD1() *CloudflareD1Database {
	return &CloudflareD1Database{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareD1Database",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-d1-database"},
		Spec: &CloudflareD1DatabaseSpec{
			AccountId:    validAccountID,
			DatabaseName: "test-database",
		},
	}
}

func TestCloudflareD1DatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareD1DatabaseSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareD1DatabaseSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts minimal valid fields", func() {
			gomega.Expect(protovalidate.Validate(validD1())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a region hint", func() {
			in := validD1()
			in.Spec.Region = CloudflareD1Region_weur
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts read replication auto and disabled", func() {
			for _, m := range []CloudflareD1ReadReplicationMode{
				CloudflareD1ReadReplicationMode_auto,
				CloudflareD1ReadReplicationMode_disabled,
			} {
				in := validD1()
				in.Spec.ReadReplication = &CloudflareD1ReadReplication{Mode: m}
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts a jurisdiction (without a region)", func() {
			in := validD1()
			in.Spec.Jurisdiction = "eu"
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts database_name at the 64-char limit", func() {
			in := validD1()
			in.Spec.DatabaseName = "a234567890123456789012345678901234567890123456789012345678901234"
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every region", func() {
			for _, r := range []CloudflareD1Region{
				CloudflareD1Region_weur, CloudflareD1Region_eeur, CloudflareD1Region_apac,
				CloudflareD1Region_oc, CloudflareD1Region_wnam, CloudflareD1Region_enam,
			} {
				in := validD1()
				in.Spec.Region = r
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing account_id", func() {
			in := validD1()
			in.Spec.AccountId = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-hex account_id", func() {
			in := validD1()
			in.Spec.AccountId = "test-account-123"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing database_name", func() {
			in := validD1()
			in.Spec.DatabaseName = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects database_name over 64 chars", func() {
			in := validD1()
			in.Spec.DatabaseName = "a2345678901234567890123456789012345678901234567890123456789012345"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects read_replication with an unspecified mode", func() {
			in := validD1()
			in.Spec.ReadReplication = &CloudflareD1ReadReplication{}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid jurisdiction", func() {
			in := validD1()
			in.Spec.Jurisdiction = "antarctica"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects region and jurisdiction set together", func() {
			in := validD1()
			in.Spec.Region = CloudflareD1Region_weur
			in.Spec.Jurisdiction = "eu"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
