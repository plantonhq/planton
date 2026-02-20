package alicloudcontainerregistryv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudContainerRegistrySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudContainerRegistrySpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudContainerRegistrySpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-registry",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-registry",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:          "cn-shanghai",
					InstanceName:    "prod-registry",
					InstanceType:    "Advanced",
					PaymentType:     proto.String("Subscription"),
					Period:          12,
					Password:        "MyStr0ng!Pass",
					ResourceGroupId: "rg-abc123",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Standard instance type", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "standard-reg",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "us-west-1",
					InstanceName: "standard-registry",
					InstanceType: "Standard",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PayAsYouGo payment type", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "payg-reg",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "dev-registry",
					InstanceType: "Basic",
					PaymentType:  proto.String("PayAsYouGo"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with namespaces configured", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ns-reg",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "ns-registry",
					InstanceType: "Standard",
					Namespaces: []*AlicloudContainerRegistryNamespace{
						{
							Name:              "platform",
							AutoCreate:        proto.Bool(true),
							DefaultVisibility: proto.String("PRIVATE"),
						},
						{
							Name:              "frontend",
							AutoCreate:        proto.Bool(false),
							DefaultVisibility: proto.String("PUBLIC"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with namespace using only required fields", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "minimal-ns-reg",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "minimal-ns-registry",
					InstanceType: "Basic",
					Namespaces: []*AlicloudContainerRegistryNamespace{
						{
							Name: "backend",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					InstanceName: "my-registry",
					InstanceType: "Basic",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_name is missing", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceType: "Basic",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_type is missing", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_type has invalid value", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Enterprise",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when payment_type has invalid value", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
					PaymentType:  proto.String("Spot"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when namespace name is too short", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
					Namespaces: []*AlicloudContainerRegistryNamespace{
						{
							Name: "x",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when namespace name exceeds max length", func() {
			longName := ""
			for i := 0; i < 121; i++ {
				longName += "a"
			}
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
					Namespaces: []*AlicloudContainerRegistryNamespace{
						{
							Name: longName,
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when namespace default_visibility has invalid value", func() {
			input := &AlicloudContainerRegistry{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudContainerRegistry",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudContainerRegistrySpec{
					Region:       "cn-hangzhou",
					InstanceName: "my-registry",
					InstanceType: "Basic",
					Namespaces: []*AlicloudContainerRegistryNamespace{
						{
							Name:              "test-ns",
							DefaultVisibility: proto.String("INTERNAL"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
