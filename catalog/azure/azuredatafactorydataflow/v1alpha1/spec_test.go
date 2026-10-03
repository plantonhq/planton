package azuredatafactorydataflowv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDataFactoryDataFlowSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataFactoryDataFlowSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testFactoryId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DataFactory/factories/app-df"

const testScript = "source(allowSchemaDrift: true, validateSchema: false) ~> source1\nsource1 sink(allowSchemaDrift: true, validateSchema: false) ~> sink1"

// validResource returns a valid mapping data flow that individual
// cases mutate into the shape under test.
func validResource() *AzureDataFactoryDataFlow {
	return &AzureDataFactoryDataFlow{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataFactoryDataFlow",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-adf-data-flow",
		},
		Spec: &AzureDataFactoryDataFlowSpec{
			DataFactoryId: literal(testFactoryId),
			Name:          "transform-orders",
			Script:        testScript,
			Sources: []*AzureDataFactoryDataFlowSource{
				{Name: "source1"},
			},
			Sinks: []*AzureDataFactoryDataFlowSink{
				{Name: "sink1"},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureDataFactoryDataFlowSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_data_factory_data_flow", func() {

			ginkgo.It("should not return a validation error for the minimal mapping shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a flowlet without sources and sinks", func() {
				input := validResource()
				input.Spec.Flowlet = true
				input.Spec.Sources = nil
				input.Spec.Sinks = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept script_lines in place of script", func() {
				input := validResource()
				input.Spec.Script = ""
				input.Spec.ScriptLines = []string{"source() ~> source1", "source1 sink() ~> sink1"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full optional surface", func() {
				input := validResource()
				input.Spec.Description = "cleans and reshapes daily orders"
				input.Spec.Annotations = []string{"team:data"}
				input.Spec.Folder = "transformations/daily"
				input.Spec.ScriptLines = []string{"// extra line"}
				input.Spec.Sources[0].Description = "raw orders"
				input.Spec.Sources[0].Dataset = &AzureDataFactoryDataFlowDatasetReference{
					Name:       literal("raw_orders"),
					Parameters: map[string]string{"path": "raw/orders"},
				}
				input.Spec.Sources[0].SchemaLinkedService = &AzureDataFactoryDataFlowLinkedServiceReference{
					Name: literal("schema-store"),
				}
				input.Spec.Sinks[0].LinkedService = &AzureDataFactoryDataFlowLinkedServiceReference{
					Name: literal("lakehouse"),
				}
				input.Spec.Sinks[0].RejectedLinkedService = &AzureDataFactoryDataFlowLinkedServiceReference{
					Name: literal("quarantine"),
				}
				input.Spec.Transformations = []*AzureDataFactoryDataFlowTransformation{
					{
						Name: "join1",
						Flowlet: &AzureDataFactoryDataFlowFlowletReference{
							Name:              literal("shared-cleanup"),
							Parameters:        map[string]string{"mode": "strict"},
							DatasetParameters: "path: 'raw/orders'",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a source embedding a flowlet by reference", func() {
				input := validResource()
				input.Spec.Sources[0].Flowlet = &AzureDataFactoryDataFlowFlowletReference{
					Name: literal("shared-cleanup"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_data_factory_data_flow", func() {

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

			ginkgo.It("should reject a data flow without script and script_lines", func() {
				input := validResource()
				input.Spec.Script = ""
				input.Spec.ScriptLines = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty script_lines entry", func() {
				input := validResource()
				input.Spec.ScriptLines = []string{""}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a mapping data flow without sources", func() {
				input := validResource()
				input.Spec.Sources = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a mapping data flow without sinks", func() {
				input := validResource()
				input.Spec.Sinks = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a source without a name", func() {
				input := validResource()
				input.Spec.Sources[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a sink without a name", func() {
				input := validResource()
				input.Spec.Sinks[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a transformation without a name", func() {
				input := validResource()
				input.Spec.Transformations = []*AzureDataFactoryDataFlowTransformation{{Name: ""}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a dataset reference without a name", func() {
				input := validResource()
				input.Spec.Sources[0].Dataset = &AzureDataFactoryDataFlowDatasetReference{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a flowlet reference without a name", func() {
				input := validResource()
				input.Spec.Sources[0].Flowlet = &AzureDataFactoryDataFlowFlowletReference{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a linked service reference without a name", func() {
				input := validResource()
				input.Spec.Sinks[0].LinkedService = &AzureDataFactoryDataFlowLinkedServiceReference{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
