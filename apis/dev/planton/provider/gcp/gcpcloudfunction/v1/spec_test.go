package gcpcloudfunctionv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpCloudFunctionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudFunctionSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpCloudFunctionSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid HTTP function", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python311",
							EntryPoint: "hello_http",
							Source: &GcpCloudFunctionSource{
								Bucket: "my-code-bucket",
								Object: "function.zip",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("valid HTTP function with service config", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "api-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-east1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "nodejs20",
							EntryPoint: "handleRequest",
							Source: &GcpCloudFunctionSource{
								Bucket: "code-bucket",
								Object: "api-v1.zip",
							},
						},
						ServiceConfig: &GcpCloudFunctionServiceConfig{
							ServiceAccountEmail:           "sa@test-project-123.iam.gserviceaccount.com",
							AvailableMemoryMb:             512,
							TimeoutSeconds:                120,
							MaxInstanceRequestConcurrency: 80,
							EnvironmentVariables: map[string]string{
								"LOG_LEVEL": "info",
							},
							SecretEnvironmentVariables: map[string]string{
								"API_KEY": "api-key-secret",
							},
							Scaling: &GcpCloudFunctionScalingConfig{
								MinInstanceCount: 1,
								MaxInstanceCount: 10,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("valid event-driven function with Pub/Sub trigger", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pubsub-worker",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "go122",
							EntryPoint: "ProcessMessage",
							Source: &GcpCloudFunctionSource{
								Bucket: "code-bucket",
								Object: "worker-v1.zip",
							},
						},
						Trigger: &GcpCloudFunctionTrigger{
							TriggerType: GcpCloudFunctionTriggerType_EVENT_TRIGGER,
							EventTrigger: &GcpCloudFunctionEventTrigger{
								EventType:   "google.cloud.pubsub.topic.v1.messagePublished",
								PubsubTopic: "projects/test-project-123/topics/job-queue",
								RetryPolicy: GcpCloudFunctionRetryPolicy_RETRY_POLICY_RETRY,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("valid function with VPC connector", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "vpc-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python312",
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "code-bucket",
								Object: "function.zip",
							},
						},
						ServiceConfig: &GcpCloudFunctionServiceConfig{
							AvailableMemoryMb:             256,
							TimeoutSeconds:                60,
							MaxInstanceRequestConcurrency: 80,
							VpcConnector:                  "projects/test-project-123/locations/us-central1/connectors/my-connector",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required project_id", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python311",
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "bucket",
								Object: "code.zip",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("valid project_id using value_from reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name:      "main-project",
									FieldPath: "status.outputs.project_id",
								},
							},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python311",
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "bucket",
								Object: "code.zip",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid runtime", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python27", // Invalid: deprecated runtime
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "bucket",
								Object: "code.zip",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing build_config", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid memory value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python311",
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "bucket",
								Object: "code.zip",
							},
						},
						ServiceConfig: &GcpCloudFunctionServiceConfig{
							AvailableMemoryMb: 300, // Invalid: not in allowed list
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid scaling: min > max", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python311",
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "bucket",
								Object: "code.zip",
							},
						},
						ServiceConfig: &GcpCloudFunctionServiceConfig{
							Scaling: &GcpCloudFunctionScalingConfig{
								MinInstanceCount: 10, // Invalid: min > max
								MaxInstanceCount: 5,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("event trigger without event_trigger config", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python311",
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "bucket",
								Object: "code.zip",
							},
						},
						Trigger: &GcpCloudFunctionTrigger{
							TriggerType: GcpCloudFunctionTriggerType_EVENT_TRIGGER,
							// Missing EventTrigger config - should fail validation
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid VPC connector format", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpCloudFunction{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudFunction",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-function",
					},
					Spec: &GcpCloudFunctionSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Region: "us-central1",
						BuildConfig: &GcpCloudFunctionBuildConfig{
							Runtime:    "python311",
							EntryPoint: "main",
							Source: &GcpCloudFunctionSource{
								Bucket: "bucket",
								Object: "code.zip",
							},
						},
						ServiceConfig: &GcpCloudFunctionServiceConfig{
							VpcConnector: "invalid-format", // Invalid format
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
