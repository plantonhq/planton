package awsplantonrunnerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsPlantonRunnerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsPlantonRunnerSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func stringPtr(v string) *string { return &v }

// minimalValidRunner is the common case: a runner placed on two private
// subnets with its runner token supplied (in real deployments the token
// arrives as a managed-secret reference; validation sees the resolved
// value).
func minimalValidRunner() *AwsPlantonRunner {
	return &AwsPlantonRunner{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsPlantonRunner",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "vpc-runner",
		},
		Spec: &AwsPlantonRunnerSpec{
			Region: "us-west-2",
			Subnets: []*foreignkeyv1.StringValueOrRef{
				literalRef("subnet-0123456789abcdef0"),
				literalRef("subnet-0fedcba9876543210"),
			},
			Token: "prt_FAKE_PLACEHOLDER_VALUE",
		},
	}
}

var _ = ginkgo.Describe("AwsPlantonRunnerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_planton_runner", func() {

			ginkgo.It("should not return a validation error for a minimal runner", func() {
				err := protovalidate.Validate(minimalValidRunner())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the default sizing pair (512 cpu / 1024 memory)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(512)
				input.Spec.Memory = int32Ptr(1024)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the smallest sizing pair (256 cpu / 512 memory)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(256)
				input.Spec.Memory = int32Ptr(512)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a sized-up runner (2048 cpu / 8192 memory)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(2048)
				input.Spec.Memory = int32Ptr(8192)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the largest sizing pair (16384 cpu / 122880 memory)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(16384)
				input.Spec.Memory = int32Ptr(122880)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept extra security groups and a public IP", func() {
				input := minimalValidRunner()
				input.Spec.SecurityGroups = []*foreignkeyv1.StringValueOrRef{
					literalRef("sg-0123456789abcdef0"),
				}
				input.Spec.AssignPublicIp = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a self-hosted control plane endpoint", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "planton.example.com:443"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a referenced runtime task role", func() {
				input := minimalValidRunner()
				input.Spec.TaskRole = literalRef("arn:aws:iam::123456789012:role/vpc-runner-runtime")
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

			ginkgo.It("should accept a supported log retention period", func() {
				input := minimalValidRunner()
				input.Spec.LogRetentionDays = int32Ptr(90)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_planton_runner", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalValidRunner()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when subnets are missing", func() {
				input := minimalValidRunner()
				input.Spec.Subnets = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when the token is missing", func() {
				input := minimalValidRunner()
				input.Spec.Token = ""
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

			ginkgo.It("should reject a cpu value outside the serverless sizes", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(300)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject memory below the cpu tier's floor", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(2048)
				input.Spec.Memory = int32Ptr(1024)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject memory above the cpu tier's ceiling", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(512)
				input.Spec.Memory = int32Ptr(8192)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject memory that is not a valid step for the tier", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(512)
				input.Spec.Memory = int32Ptr(1536)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a large tier memory off its 4096 step", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = int32Ptr(8192)
				input.Spec.Memory = int32Ptr(18432)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unsupported log retention period", func() {
				input := minimalValidRunner()
				input.Spec.LogRetentionDays = int32Ptr(45)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
