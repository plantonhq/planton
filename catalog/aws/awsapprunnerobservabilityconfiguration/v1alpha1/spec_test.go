package awsapprunnerobservabilityconfigurationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAwsAppRunnerObservabilityConfigurationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAppRunnerObservabilityConfiguration Validation Suite")
}

func strPtr(s string) *string { return &s }

func validEnvelope(spec *AwsAppRunnerObservabilityConfigurationSpec) *AwsAppRunnerObservabilityConfiguration {
	return &AwsAppRunnerObservabilityConfiguration{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsAppRunnerObservabilityConfiguration",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-oc"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("AwsAppRunnerObservabilityConfigurationSpec validations", func() {
	var spec *AwsAppRunnerObservabilityConfigurationSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsAppRunnerObservabilityConfigurationSpec{
			Region: "us-west-2",
			TraceConfiguration: &AwsAppRunnerObservabilityConfigurationTraceConfiguration{
				Vendor: strPtr("AWSXRAY"),
			},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a spec with X-Ray tracing", func() {
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts trace_configuration without an explicit vendor (default applies)", func() {
		spec.TraceConfiguration = &AwsAppRunnerObservabilityConfigurationTraceConfiguration{}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec without trace_configuration (inert but valid)", func() {
		spec.TraceConfiguration = nil
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

	ginkgo.It("fails when vendor is not a supported tracing vendor", func() {
		spec.TraceConfiguration.Vendor = strPtr("DATADOG")
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
