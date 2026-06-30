package cloudflareworkerskvpairv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validPair() *CloudflareWorkersKvPair {
	return &CloudflareWorkersKvPair{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareWorkersKvPair",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-kv-pair"},
		Spec: &CloudflareWorkersKvPairSpec{
			AccountId:   validAccountID,
			NamespaceId: value("0f1e2d3c4b5a69788796a5b4c3d2e1f0"),
			KeyName:     "feature.new-dashboard",
			Value:       "true",
		},
	}
}

func TestCloudflareWorkersKvPairSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareWorkersKvPairSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareWorkersKvPairSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal valid entry", func() {
			gomega.Expect(protovalidate.Validate(validPair())).To(gomega.BeNil())
		})

		ginkgo.It("accepts metadata and a namespace reference", func() {
			in := validPair()
			in.Spec.Metadata = `{"owner":"platform"}`
			in.Spec.NamespaceId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "app-config", FieldPath: "status.outputs.namespace_id"},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a key_name at the 512-byte limit", func() {
			in := validPair()
			in.Spec.KeyName = strings.Repeat("a", 512)
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validPair()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing namespace_id", func() {
			in := validPair()
			in.Spec.NamespaceId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty namespace_id reference", func() {
			in := validPair()
			in.Spec.NamespaceId = value("")
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing key_name", func() {
			in := validPair()
			in.Spec.KeyName = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a key_name over 512 bytes", func() {
			in := validPair()
			in.Spec.KeyName = strings.Repeat("a", 513)
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing value", func() {
			in := validPair()
			in.Spec.Value = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
