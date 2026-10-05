package gcpkmsautokeyconfigv1alpha1

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
	ginkgo.RunSpecs(t, "GcpKmsAutokeyConfigSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpKmsAutokeyConfigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpKmsAutokeyConfig {
		return &GcpKmsAutokeyConfig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpKmsAutokeyConfig",
			Metadata:   &shared.CatalogObjectMetadata{Name: "project-autokey"},
			Spec:       &GcpKmsAutokeyConfigSpec{KeyProjectResolutionMode: "RESOURCE_PROJECT"},
		}
	}

	folder := func() *GcpKmsAutokeyConfig {
		msg := minimal()
		msg.Spec.Scope = &GcpKmsAutokeyConfigScope{FolderId: litRef("123456789012")}
		msg.Spec.KeyProjectResolutionMode = "DEDICATED_KEY_PROJECT"
		msg.Spec.KeyProject = litRef("security-keys")
		return msg
	}

	ginkgo.It("should accept same-project storage on the provider's project", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept dedicated-project storage on a folder", func() {
		msg := folder()
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project switching Autokey off under an enabled folder", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpKmsAutokeyConfigScope{ProjectId: litRef("sandbox")}
		msg.Spec.KeyProjectResolutionMode = "DISABLED"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject both scope arms", func() {
		msg := folder()
		msg.Spec.Scope.ProjectId = litRef("p")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a key project on a project configuration", func() {
		msg := minimal()
		msg.Spec.KeyProject = litRef("security-keys")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject dedicated-project storage on a project", func() {
		msg := minimal()
		msg.Spec.KeyProjectResolutionMode = "DEDICATED_KEY_PROJECT"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject dedicated-project storage without a key project", func() {
		msg := folder()
		msg.Spec.KeyProject = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown resolution mode or deletion policy", func() {
		msg := minimal()
		msg.Spec.KeyProjectResolutionMode = "CENTRALIZED"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
