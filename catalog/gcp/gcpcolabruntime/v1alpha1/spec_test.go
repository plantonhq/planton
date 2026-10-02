package gcpcolabruntimev1alpha1

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
	ginkgo.RunSpecs(t, "GcpColabRuntimeSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpColabRuntimeSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpColabRuntime {
		return &GcpColabRuntime{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpColabRuntime",
			Metadata:   &shared.CloudResourceMetadata{Name: "alice-gpu"},
			Spec: &GcpColabRuntimeSpec{
				Location:        "us-central1",
				RuntimeTemplate: litRef("projects/ml-project/locations/us-central1/notebookRuntimeTemplates/gpu-t4"),
				RuntimeUser:     "alice@example.com",
			},
		}
	}

	ginkgo.It("should accept the smallest runtime", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.RuntimeId = "alice-gpu-01"
		msg.Spec.DisplayName = "Alice's GPU runtime"
		msg.Spec.Description = "Vision experiments"
		msg.Spec.DesiredState = "STOPPED"
		msg.Spec.AutoUpgrade = true
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a template and a user email", func() {
		msg := minimal()
		msg.Spec.RuntimeTemplate = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.RuntimeUser = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.RuntimeUser = "alice"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept only RUNNING and STOPPED as a desired state", func() {
		msg := minimal()
		msg.Spec.DesiredState = "ACTIVE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.DesiredState = "RUNNING"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject an upper-case runtime id and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.RuntimeId = "Alice"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
