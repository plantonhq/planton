package ocifunctionsapplicationv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestOciFunctionsApplicationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciFunctionsApplicationSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func newValueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name: name,
			},
		},
	}
}

func minimalValidApp() *OciFunctionsApplication {
	return &OciFunctionsApplication{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciFunctionsApplication",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-functions-app",
		},
		Spec: &OciFunctionsApplicationSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				newStringValueOrRef("ocid1.subnet.oc1..example"),
			},
		},
	}
}

var _ = ginkgo.Describe("OciFunctionsApplicationSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_functions_application", func() {

			ginkgo.It("should not return a validation error for minimal application", func() {
				input := minimalValidApp()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidApp()
				input.Spec.DisplayName = "my-serverless-app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with shape generic_x86", func() {
				input := minimalValidApp()
				input.Spec.Shape = OciFunctionsApplicationSpec_generic_x86
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with shape generic_arm", func() {
				input := minimalValidApp()
				input.Spec.Shape = OciFunctionsApplicationSpec_generic_arm
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with shape generic_x86_arm", func() {
				input := minimalValidApp()
				input.Spec.Shape = OciFunctionsApplicationSpec_generic_x86_arm
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with config map", func() {
				input := minimalValidApp()
				input.Spec.Config = map[string]string{
					"DB_HOST": "10.0.1.100",
					"DB_PORT": "5432",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with network security group IDs", func() {
				input := minimalValidApp()
				input.Spec.NetworkSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1..example1"),
					newStringValueOrRef("ocid1.networksecuritygroup.oc1..example2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with syslog_url", func() {
				input := minimalValidApp()
				input.Spec.SyslogUrl = "tcp://logserver.example.com:514"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with trace_config enabled", func() {
				input := minimalValidApp()
				input.Spec.TraceConfig = &OciFunctionsApplicationSpec_TraceConfig{
					IsEnabled: proto.Bool(true),
					DomainId:  "ocid1.apmdomain.oc1..example",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with trace_config disabled", func() {
				input := minimalValidApp()
				input.Spec.TraceConfig = &OciFunctionsApplicationSpec_TraceConfig{
					IsEnabled: proto.Bool(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with image_policy_config enabled and keys", func() {
				input := minimalValidApp()
				input.Spec.ImagePolicyConfig = &OciFunctionsApplicationSpec_ImagePolicyConfig{
					IsPolicyEnabled: true,
					KeyDetails: []*OciFunctionsApplicationSpec_ImagePolicyKeyDetail{
						{KmsKeyId: newStringValueOrRef("ocid1.key.oc1..example")},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with image_policy_config disabled and no keys", func() {
				input := minimalValidApp()
				input.Spec.ImagePolicyConfig = &OciFunctionsApplicationSpec_ImagePolicyConfig{
					IsPolicyEnabled: false,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple subnets", func() {
				input := minimalValidApp()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.subnet.oc1..subnet1"),
					newStringValueOrRef("ocid1.subnet.oc1..subnet2"),
					newStringValueOrRef("ocid1.subnet.oc1..subnet3"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidApp()
				input.Spec.CompartmentId = newValueFromRef("my-compartment")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subnet_ids via valueFrom ref", func() {
				input := minimalValidApp()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
					newValueFromRef("my-subnet"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidApp()
				input.Spec.DisplayName = "production-functions-app"
				input.Spec.Shape = OciFunctionsApplicationSpec_generic_arm
				input.Spec.Config = map[string]string{"APP_ENV": "production"}
				input.Spec.NetworkSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1..example"),
				}
				input.Spec.SyslogUrl = "tcp+tls://logs.example.com:6514"
				input.Spec.ImagePolicyConfig = &OciFunctionsApplicationSpec_ImagePolicyConfig{
					IsPolicyEnabled: true,
					KeyDetails: []*OciFunctionsApplicationSpec_ImagePolicyKeyDetail{
						{KmsKeyId: newStringValueOrRef("ocid1.key.oc1..example1")},
						{KmsKeyId: newStringValueOrRef("ocid1.key.oc1..example2")},
					},
				}
				input.Spec.TraceConfig = &OciFunctionsApplicationSpec_TraceConfig{
					IsEnabled: proto.Bool(true),
					DomainId:  "ocid1.apmdomain.oc1..example",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_functions_application", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidApp()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidApp()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidApp()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciFunctionsApplication{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciFunctionsApplication",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidApp()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_ids is nil", func() {
				input := minimalValidApp()
				input.Spec.SubnetIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_ids is empty", func() {
				input := minimalValidApp()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image_policy is enabled with empty key_details", func() {
				input := minimalValidApp()
				input.Spec.ImagePolicyConfig = &OciFunctionsApplicationSpec_ImagePolicyConfig{
					IsPolicyEnabled: true,
					KeyDetails:      []*OciFunctionsApplicationSpec_ImagePolicyKeyDetail{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image_policy is enabled with nil key_details", func() {
				input := minimalValidApp()
				input.Spec.ImagePolicyConfig = &OciFunctionsApplicationSpec_ImagePolicyConfig{
					IsPolicyEnabled: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kms_key_id is missing in key_details", func() {
				input := minimalValidApp()
				input.Spec.ImagePolicyConfig = &OciFunctionsApplicationSpec_ImagePolicyConfig{
					IsPolicyEnabled: true,
					KeyDetails: []*OciFunctionsApplicationSpec_ImagePolicyKeyDetail{
						{KmsKeyId: nil},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
