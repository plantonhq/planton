package awsapprunnerautoscalingconfigurationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAwsAppRunnerAutoScalingConfigurationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAppRunnerAutoScalingConfiguration Validation Suite")
}

func int32Ptr(i int32) *int32 { return &i }

func validEnvelope(spec *AwsAppRunnerAutoScalingConfigurationSpec) *AwsAppRunnerAutoScalingConfiguration {
	return &AwsAppRunnerAutoScalingConfiguration{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsAppRunnerAutoScalingConfiguration",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-asc"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("AwsAppRunnerAutoScalingConfigurationSpec validations", func() {
	var spec *AwsAppRunnerAutoScalingConfigurationSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsAppRunnerAutoScalingConfigurationSpec{
			Region: "us-west-2",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec (region only -- AWS defaults apply)", func() {
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully tuned configuration", func() {
		spec.MaxConcurrency = int32Ptr(50)
		spec.MaxSize = int32Ptr(10)
		spec.MinSize = int32Ptr(2)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts the concurrency-of-one serverless posture", func() {
		spec.MaxConcurrency = int32Ptr(1)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts max_size equal to min_size (fixed fleet)", func() {
		spec.MinSize = int32Ptr(5)
		spec.MaxSize = int32Ptr(5)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts the maximum concurrency dial (200)", func() {
		spec.MaxConcurrency = int32Ptr(200)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a configuration claiming the account default", func() {
		spec.MaxConcurrency = int32Ptr(80)
		spec.SetAsAccountDefault = true
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure cases
	// -------------------------------------------------------------------------

	ginkgo.It("fails when region is empty", func() {
		spec.Region = ""
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when max_concurrency is 0 (below range)", func() {
		spec.MaxConcurrency = int32Ptr(0)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when max_concurrency is 201 (above range)", func() {
		spec.MaxConcurrency = int32Ptr(201)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when min_size is 0 (below range)", func() {
		spec.MinSize = int32Ptr(0)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when max_size is 0 (below range)", func() {
		spec.MaxSize = int32Ptr(0)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when max_size is below min_size", func() {
		spec.MinSize = int32Ptr(10)
		spec.MaxSize = int32Ptr(5)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
