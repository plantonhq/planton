package azuremachinelearningbatchdeploymentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMachineLearningBatchDeploymentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningBatchDeploymentSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testEndpointId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/batchEndpoints/nightly-scoring"
	testComputeId   = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/computes/cpu-pool"
	testModelId     = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/models/churn/versions/3"
	testComponentId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/components/nightly-pipeline/versions/1"
	testDatastoreId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/datastores/models"
	testJobId       = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/jobs/train-run-42"
)

// validResource returns a minimal valid Model-type recipe. Model and
// compute are both present because the service REQUIRES them on a
// Model-type create (live-proven: 400 UserError / ArgumentNullOrEmpty
// on each when missing) -- a bare deployment is schema-legal but
// undeployable, and the spec's CEL mirrors the service.
func validResource() *AzureMachineLearningBatchDeployment {
	return &AzureMachineLearningBatchDeployment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningBatchDeployment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ml-batch-deployment",
		},
		Spec: &AzureMachineLearningBatchDeploymentSpec{
			EndpointId: literal(testEndpointId),
			Name:       "production",
			Region:     "eastus",
			ComputeId:  literal(testComputeId),
			Model: &AzureMachineLearningBatchDeploymentModel{
				Id: &AzureMachineLearningBatchDeploymentModelIdReference{AssetId: testModelId},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningBatchDeploymentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_batch_deployment", func() {

			ginkgo.It("should not return a validation error for a minimal model recipe", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a full model recipe on a compute pool", func() {
				input := validResource()
				input.Spec.ComputeId = literal(testComputeId)
				input.Spec.Model = &AzureMachineLearningBatchDeploymentModel{
					Id: &AzureMachineLearningBatchDeploymentModelIdReference{AssetId: testModelId},
				}
				input.Spec.Resources = &AzureMachineLearningBatchDeploymentResources{
					InstanceCount: int32Ptr(4),
					InstanceType:  "Standard_DS3_v2",
				}
				input.Spec.RetrySettings = &AzureMachineLearningBatchDeploymentRetrySettings{
					MaxRetries: int32Ptr(5),
					Timeout:    "PT1M",
				}
				input.Spec.MiniBatchSize = int64Ptr(50)
				input.Spec.MaxConcurrencyPerInstance = int32Ptr(2)
				input.Spec.ErrorThreshold = int32Ptr(10)
				input.Spec.OutputAction = "AppendRow"
				input.Spec.OutputFileName = "scores.csv"
				input.Spec.LoggingLevel = "Debug"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept each model reference arm alone", func() {
				arms := []*AzureMachineLearningBatchDeploymentModel{
					{Id: &AzureMachineLearningBatchDeploymentModelIdReference{AssetId: testModelId}},
					{DataPath: &AzureMachineLearningBatchDeploymentModelDataPathReference{DatastoreId: testDatastoreId, Path: "models/churn"}},
					{OutputPath: &AzureMachineLearningBatchDeploymentModelOutputPathReference{JobId: testJobId, Path: "outputs/model"}},
				}
				for _, arm := range arms {
					input := validResource()
					input.Spec.Model = arm
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a scoring-code recipe", func() {
				input := validResource()
				input.Spec.CodeConfiguration = &AzureMachineLearningBatchDeploymentCodeConfiguration{
					CodeId:        "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/codes/scoring/versions/2",
					ScoringScript: "score.py",
				}
				input.Spec.EnvironmentId = "azureml://registries/azureml/environments/sklearn-1.5/versions/1"
				input.Spec.EnvironmentVariables = map[string]string{"LOG_LEVEL": "info"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pipeline-component recipe with no model or compute", func() {
				// A pipeline recipe carries its own steps -- the service's
				// model/compute validators apply to Model-type creates only,
				// and the spec's CEL mirrors that split.
				input := validResource()
				input.Spec.Model = nil
				input.Spec.ComputeId = nil
				input.Spec.PipelineComponent = &AzureMachineLearningBatchDeploymentPipelineComponent{
					ComponentId:    testComponentId,
					Settings:       map[string]string{"default_compute": "cpu-pool", "continue_on_step_failure": "false"},
					JobTags:        map[string]string{"trigger": "batch-endpoint"},
					JobDescription: "nightly scoring pipeline",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the error threshold's sentinel -1 (ignore all failures)", func() {
				input := validResource()
				input.Spec.ErrorThreshold = int32Ptr(-1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept names on ARM's rule: digit start, underscores and hyphens", func() {
				for _, name := range []string{"0deployment", "production_v2-blue", "a"} {
					input := validResource()
					input.Spec.Name = name
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_batch_deployment", func() {

			ginkgo.It("should reject a missing endpoint reference", func() {
				input := validResource()
				input.Spec.EndpointId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a bare model-type recipe (the service's own boundary)", func() {
				// Live-proven: the batch service rejects a Model-type create
				// naming neither a model nor a compute with 400 UserError /
				// ArgumentNullOrEmpty on both properties.
				input := validResource()
				input.Spec.Model = nil
				input.Spec.ComputeId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a model-type recipe without a compute", func() {
				input := validResource()
				input.Spec.ComputeId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a model-type recipe without a model", func() {
				input := validResource()
				input.Spec.Model = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty model block (no reference arm)", func() {
				input := validResource()
				input.Spec.Model = &AzureMachineLearningBatchDeploymentModel{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a model block carrying two reference arms", func() {
				input := validResource()
				input.Spec.Model = &AzureMachineLearningBatchDeploymentModel{
					Id:       &AzureMachineLearningBatchDeploymentModelIdReference{AssetId: testModelId},
					DataPath: &AzureMachineLearningBatchDeploymentModelDataPathReference{DatastoreId: testDatastoreId},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an id model reference without an asset id", func() {
				input := validResource()
				input.Spec.Model = &AzureMachineLearningBatchDeploymentModel{
					Id: &AzureMachineLearningBatchDeploymentModelIdReference{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a code configuration without a scoring script", func() {
				input := validResource()
				input.Spec.CodeConfiguration = &AzureMachineLearningBatchDeploymentCodeConfiguration{
					CodeId: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/codes/scoring/versions/2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a pipeline component without the component id (the recorded tightening)", func() {
				input := validResource()
				input.Spec.PipelineComponent = &AzureMachineLearningBatchDeploymentPipelineComponent{
					Settings: map[string]string{"default_compute": "cpu-pool"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an out-of-vocabulary output action", func() {
				input := validResource()
				input.Spec.OutputAction = "Append"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an out-of-vocabulary logging level", func() {
				input := validResource()
				input.Spec.LoggingLevel = "Verbose"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an error threshold below the -1 sentinel", func() {
				input := validResource()
				input.Spec.ErrorThreshold = int32Ptr(-2)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero mini-batch size", func() {
				input := validResource()
				input.Spec.MiniBatchSize = int64Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero instance count", func() {
				input := validResource()
				input.Spec.Resources = &AzureMachineLearningBatchDeploymentResources{
					InstanceCount: int32Ptr(0),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative retry count", func() {
				input := validResource()
				input.Spec.RetrySettings = &AzureMachineLearningBatchDeploymentRetrySettings{
					MaxRetries: int32Ptr(-1),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a retry timeout that is not an ISO-8601 duration", func() {
				input := validResource()
				input.Spec.RetrySettings = &AzureMachineLearningBatchDeploymentRetrySettings{
					Timeout: "30s",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names ARM rejects", func() {
				for _, name := range []string{"-deployment", "_deployment", "prod deployment", "prod.deployment"} {
					input := validResource()
					input.Spec.Name = name
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				}
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind literal", func() {
				input := validResource()
				input.Kind = "AzureMachineLearningDeployment"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

func int32Ptr(v int32) *int32 { return &v }
func int64Ptr(v int64) *int64 { return &v }
