package azuremachinelearningonlinedeploymentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMachineLearningOnlineDeploymentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningOnlineDeploymentSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testEndpointId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/onlineEndpoints/fraud-scoring"
	testModelId    = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/models/fraud-model/versions/3"
)

// validResource returns a minimal valid managed deployment that
// individual cases mutate into the shape under test.
func validResource() *AzureMachineLearningOnlineDeployment {
	return &AzureMachineLearningOnlineDeployment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningOnlineDeployment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ml-online-deployment",
		},
		Spec: &AzureMachineLearningOnlineDeploymentSpec{
			EndpointId: literal(testEndpointId),
			Name:       "blue",
			Region:     "eastus",
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningOnlineDeploymentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_online_deployment", func() {

			ginkgo.It("should not return a validation error for a minimal managed deployment", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a registered model with instance type and count", func() {
				input := validResource()
				input.Spec.Model = testModelId
				input.Spec.InstanceType = "Standard_DS3_v2"
				count := int32(3)
				input.Spec.InstanceCount = &count
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept scoring code with all three probes on ISO-8601 durations", func() {
				input := validResource()
				input.Spec.CodeConfiguration = &AzureMachineLearningOnlineDeploymentCodeConfiguration{
					CodeId:        "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/codes/scoring/versions/1",
					ScoringScript: "score.py",
				}
				probe := &AzureMachineLearningOnlineDeploymentProbeSettings{
					InitialDelay: "PT10S",
					Period:       "PT10S",
					Timeout:      "PT2S",
				}
				input.Spec.LivenessProbe = probe
				input.Spec.ReadinessProbe = probe
				input.Spec.StartupProbe = probe
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept request settings with a raised concurrency", func() {
				input := validResource()
				concurrency := int32(4)
				input.Spec.RequestSettings = &AzureMachineLearningOnlineDeploymentRequestSettings{
					MaxConcurrentRequestsPerInstance: &concurrency,
					RequestTimeout:                   "PT30S",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a data collector capturing inputs and outputs", func() {
				input := validResource()
				rate := 0.5
				input.Spec.DataCollector = &AzureMachineLearningOnlineDeploymentDataCollector{
					Collections: map[string]*AzureMachineLearningOnlineDeploymentDataCollection{
						"model_inputs":  {Enabled: true, SamplingRate: &rate},
						"model_outputs": {Enabled: true},
					},
					RollingRate: "Day",
					RequestLogging: &AzureMachineLearningOnlineDeploymentRequestLogging{
						CaptureHeaders: []string{"x-request-id"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept secure egress with environment configuration", func() {
				input := validResource()
				disabled := false
				input.Spec.EgressPublicNetworkAccessEnabled = &disabled
				input.Spec.EnvironmentId = "azureml://registries/azureml/environments/sklearn-1.5/versions/12"
				input.Spec.EnvironmentVariables = map[string]string{"LOG_LEVEL": "info"}
				input.Spec.AppInsightsEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_online_deployment", func() {

			ginkgo.It("should reject a missing endpoint reference", func() {
				input := validResource()
				input.Spec.EndpointId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an instance count below 1 (no scale-to-zero)", func() {
				input := validResource()
				count := int32(0)
				input.Spec.InstanceCount = &count
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject scoring code without the script", func() {
				input := validResource()
				input.Spec.CodeConfiguration = &AzureMachineLearningOnlineDeploymentCodeConfiguration{
					CodeId: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace/codes/scoring/versions/1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a probe delay that is not an ISO-8601 duration", func() {
				input := validResource()
				input.Spec.LivenessProbe = &AzureMachineLearningOnlineDeploymentProbeSettings{
					InitialDelay: "10 seconds",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a request timeout that is not an ISO-8601 duration", func() {
				input := validResource()
				input.Spec.RequestSettings = &AzureMachineLearningOnlineDeploymentRequestSettings{
					RequestTimeout: "5000ms",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero concurrency", func() {
				input := validResource()
				concurrency := int32(0)
				input.Spec.RequestSettings = &AzureMachineLearningOnlineDeploymentRequestSettings{
					MaxConcurrentRequestsPerInstance: &concurrency,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data collector without collections (ARM's own contract)", func() {
				input := validResource()
				input.Spec.DataCollector = &AzureMachineLearningOnlineDeploymentDataCollector{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a sampling rate above 1", func() {
				input := validResource()
				rate := 1.5
				input.Spec.DataCollector = &AzureMachineLearningOnlineDeploymentDataCollector{
					Collections: map[string]*AzureMachineLearningOnlineDeploymentDataCollection{
						"model_inputs": {Enabled: true, SamplingRate: &rate},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a rolling rate outside the service vocabulary", func() {
				input := validResource()
				input.Spec.DataCollector = &AzureMachineLearningOnlineDeploymentDataCollector{
					Collections: map[string]*AzureMachineLearningOnlineDeploymentDataCollection{
						"model_inputs": {Enabled: true},
					},
					RollingRate: "Weekly",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names ARM rejects", func() {
				for _, name := range []string{"-blue", "_blue", "blue green", "blue.green"} {
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
