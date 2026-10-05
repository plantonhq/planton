package gcpplantonrunnerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpPlantonRunnerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpPlantonRunnerSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func stringPtr(v string) *string { return &v }

// minimalValidRunner is the common case: a runner in a region with its
// runner token supplied (in real deployments the token arrives as a
// managed-secret reference; validation sees the resolved value).
func minimalValidRunner() *GcpPlantonRunner {
	return &GcpPlantonRunner{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpPlantonRunner",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "vpc-runner",
		},
		Spec: &GcpPlantonRunnerSpec{
			Region: "us-central1",
			Token:  "prt_FAKE_PLACEHOLDER_VALUE",
		},
	}
}

var _ = ginkgo.Describe("GcpPlantonRunnerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_planton_runner", func() {

			ginkgo.It("should not return a validation error for a minimal runner", func() {
				err := protovalidate.Validate(minimalValidRunner())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit project reference", func() {
				input := minimalValidRunner()
				input.Spec.ProjectId = literalRef("my-project-123456")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the default sizing pair (1 cpu / 512Mi)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = stringPtr("1")
				input.Spec.Memory = stringPtr("512Mi")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a sized-up runner (4 cpu / 4Gi)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = stringPtr("4")
				input.Spec.Memory = stringPtr("4Gi")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a referenced runtime service account", func() {
				input := minimalValidRunner()
				input.Spec.ServiceAccount = literalRef("runner@my-project.iam.gserviceaccount.com")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept direct VPC egress placement with tags", func() {
				input := minimalValidRunner()
				input.Spec.VpcAccess = &GcpPlantonRunnerVpcAccess{
					Network:    literalRef("my-vpc"),
					Subnetwork: literalRef("my-subnet"),
					Tags:       []string{"planton-runner"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a self-hosted control plane endpoint", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "planton.example.com:443"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned runner version and a mirrored repository", func() {
				input := minimalValidRunner()
				input.Spec.RunnerVersion = stringPtr("v0.3.5")
				input.Spec.ImageRepository = stringPtr("mirror.example.com/planton/runner")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("gcp_planton_runner", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalValidRunner()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a region that is not a GCP region shape", func() {
				input := minimalValidRunner()
				input.Spec.Region = "US-Central-1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when the token is missing", func() {
				input := minimalValidRunner()
				input.Spec.Token = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a cpu value outside the Cloud Run sizes", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = stringPtr("3")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject memory without a Mi/Gi unit", func() {
				input := minimalValidRunner()
				input.Spec.Memory = stringPtr("512")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject VPC access without a network", func() {
				input := minimalValidRunner()
				input.Spec.VpcAccess = &GcpPlantonRunnerVpcAccess{
					Subnetwork: literalRef("my-subnet"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject VPC access without a subnetwork", func() {
				input := minimalValidRunner()
				input.Spec.VpcAccess = &GcpPlantonRunnerVpcAccess{
					Network: literalRef("my-vpc"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid network tag", func() {
				input := minimalValidRunner()
				input.Spec.VpcAccess = &GcpPlantonRunnerVpcAccess{
					Network:    literalRef("my-vpc"),
					Subnetwork: literalRef("my-subnet"),
					Tags:       []string{"Not_A_Tag"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a control plane endpoint carrying a scheme prefix", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "https://planton.example.com:443"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a control plane endpoint without a port", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "planton.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
