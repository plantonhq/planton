package alicloudceninstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	fkv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudCenInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudCenInstanceSpec Validation Tests")
}

func strRef(val string) *fkv1.StringValueOrRef {
	return &fkv1.StringValueOrRef{
		LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: val},
	}
}

func validMinimalSpec() *AliCloudCenInstanceSpec {
	return &AliCloudCenInstanceSpec{
		Region:          "cn-hangzhou",
		CenInstanceName: "my-cen",
	}
}

func validMinimalInput() *AliCloudCenInstance {
	return &AliCloudCenInstance{
		ApiVersion: "alicloud.planton.dev/v1",
		Kind:       "AliCloudCenInstance",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-cen"},
		Spec:       validMinimalSpec(),
	}
}

var _ = ginkgo.Describe("AliCloudCenInstanceSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields and no attachments", func() {
			input := validMinimalInput()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all top-level optional fields populated", func() {
			input := validMinimalInput()
			input.Spec.Description = "Multi-region backbone"
			input.Spec.ProtectionLevel = proto.String("REDUCED")
			input.Spec.ResourceGroupId = "rg-abc123"
			input.Spec.Tags = map[string]string{"team": "network", "env": "production"}
			input.Metadata.Org = "acme"
			input.Metadata.Env = "production"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a single VPC attachment", func() {
			input := validMinimalInput()
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceId:       strRef("vpc-abc123"),
					ChildInstanceRegionId: "cn-hangzhou",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with explicit VPC child_instance_type", func() {
			input := validMinimalInput()
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceId:       strRef("vpc-abc123"),
					ChildInstanceType:     proto.String("VPC"),
					ChildInstanceRegionId: "cn-hangzhou",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with VBR child_instance_type", func() {
			input := validMinimalInput()
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceId:       strRef("vbr-abc123"),
					ChildInstanceType:     proto.String("VBR"),
					ChildInstanceRegionId: "cn-shanghai",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with CCN child_instance_type", func() {
			input := validMinimalInput()
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceId:       strRef("ccn-abc123"),
					ChildInstanceType:     proto.String("CCN"),
					ChildInstanceRegionId: "cn-hangzhou",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with multiple cross-region attachments", func() {
			input := validMinimalInput()
			input.Spec.CenInstanceName = "multi-region-backbone"
			input.Spec.Description = "Connects VPCs across three regions"
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceId:       strRef("vpc-hangzhou"),
					ChildInstanceType:     proto.String("VPC"),
					ChildInstanceRegionId: "cn-hangzhou",
				},
				{
					ChildInstanceId:       strRef("vpc-shanghai"),
					ChildInstanceType:     proto.String("VPC"),
					ChildInstanceRegionId: "cn-shanghai",
				},
				{
					ChildInstanceId:       strRef("vpc-singapore"),
					ChildInstanceType:     proto.String("VPC"),
					ChildInstanceRegionId: "ap-southeast-1",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with protection_level set to REDUCED", func() {
			input := validMinimalInput()
			input.Spec.ProtectionLevel = proto.String("REDUCED")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with empty protection_level (default strict mode)", func() {
			input := validMinimalInput()
			input.Spec.ProtectionLevel = proto.String("")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with cen_instance_name at minimum length", func() {
			input := validMinimalInput()
			input.Spec.CenInstanceName = "ab"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := validMinimalInput()
			input.ApiVersion = "wrong/v1"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := validMinimalInput()
			input.Kind = "WrongKind"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := validMinimalInput()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := validMinimalInput()
			input.Spec = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := validMinimalInput()
			input.Spec.Region = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cen_instance_name is empty", func() {
			input := validMinimalInput()
			input.Spec.CenInstanceName = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cen_instance_name is too short", func() {
			input := validMinimalInput()
			input.Spec.CenInstanceName = "x"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when protection_level is invalid", func() {
			input := validMinimalInput()
			input.Spec.ProtectionLevel = proto.String("FULL")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when attachment child_instance_id is missing", func() {
			input := validMinimalInput()
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceRegionId: "cn-hangzhou",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when attachment child_instance_region_id is empty", func() {
			input := validMinimalInput()
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceId:       strRef("vpc-abc123"),
					ChildInstanceRegionId: "",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when attachment child_instance_type is invalid", func() {
			input := validMinimalInput()
			input.Spec.Attachments = []*AliCloudCenAttachment{
				{
					ChildInstanceId:       strRef("vpc-abc123"),
					ChildInstanceType:     proto.String("ECS"),
					ChildInstanceRegionId: "cn-hangzhou",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
