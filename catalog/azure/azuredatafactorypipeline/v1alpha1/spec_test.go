package azuredatafactorypipelinev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureDataFactoryPipelineSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataFactoryPipelineSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testFactoryId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DataFactory/factories/app-df"

// validResource returns a valid pipeline that individual cases mutate
// into the shape under test.
func validResource() *AzureDataFactoryPipeline {
	return &AzureDataFactoryPipeline{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataFactoryPipeline",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-adf-pipeline",
		},
		Spec: &AzureDataFactoryPipelineSpec{
			DataFactoryId: literal(testFactoryId),
			Name:          "ingest-daily",
		},
	}
}

var _ = ginkgo.Describe("AzureDataFactoryPipelineSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_data_factory_pipeline", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept names starting with a letter, number, or underscore", func() {
				input := validResource()
				input.Spec.Name = "_staging"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = "1st-load"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = "ingest daily (v2)"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full optional surface", func() {
				input := validResource()
				input.Spec.Description = "copies yesterday's orders into the lakehouse"
				input.Spec.Parameters = map[string]string{"window": "P1D"}
				input.Spec.Variables = map[string]string{"cursor": ""}
				input.Spec.ActivitiesJson = `[{"name": "wait", "type": "Wait", "typeProperties": {"waitTimeInSeconds": 10}}]`
				input.Spec.Annotations = []string{"team:data", "tier:bronze"}
				input.Spec.Concurrency = proto.Int32(4)
				input.Spec.Folder = "ingest/daily"
				input.Spec.MonitorMetricsAfterDuration = "0.00:30:00"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept concurrency at both bounds", func() {
				input := validResource()
				input.Spec.Concurrency = proto.Int32(1)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Concurrency = proto.Int32(50)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_data_factory_pipeline", func() {

			ginkgo.It("should reject a missing data factory id", func() {
				input := validResource()
				input.Spec.DataFactoryId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name starting with a forbidden character", func() {
				input := validResource()
				input.Spec.Name = "-ingest"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = ".ingest"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names containing Azure's forbidden characters", func() {
				input := validResource()
				for _, name := range []string{"inge<st", "inge>st", "inge*st", "inge#st", "inge.st", "inge%st", "inge&st", "inge:st", "inge\\st", "inge+st", "inge?st", "inge/st"} {
					input.Spec.Name = name
					gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "expected %q to be rejected", name)
				}
			})

			ginkgo.It("should reject concurrency outside 1-50", func() {
				input := validResource()
				input.Spec.Concurrency = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Concurrency = proto.Int32(51)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
