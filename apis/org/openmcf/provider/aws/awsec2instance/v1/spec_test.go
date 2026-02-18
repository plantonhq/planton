package awsec2instancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	fk "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsEc2InstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEc2InstanceSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsEc2InstanceSpec validations", func() {
	var spec *AwsEc2InstanceSpec

	newSubnet := func(id string) *fk.StringValueOrRef {
		return &fk.StringValueOrRef{LiteralOrRef: &fk.StringValueOrRef_Value{Value: id}}
	}

	newSg := func(id string) *fk.StringValueOrRef {
		return &fk.StringValueOrRef{LiteralOrRef: &fk.StringValueOrRef_Value{Value: id}}
	}

	newIamProfile := func(arn string) *fk.StringValueOrRef {
		return &fk.StringValueOrRef{LiteralOrRef: &fk.StringValueOrRef_Value{Value: arn}}
	}

	ginkgo.BeforeEach(func() {
		spec = &AwsEc2InstanceSpec{
			Region:       "us-west-2",
			InstanceName: "web-1",
			AmiId:        "ami-0123456789abcdef0",
			InstanceType: "t3.small",
			SubnetId:     newSubnet("subnet-aaa111"),
			SecurityGroupIds: []*fk.StringValueOrRef{
				newSg("sg-000111222"),
			},
			ConnectionMethod:      func() *AwsEc2InstanceConnectionMethod { v := AwsEc2InstanceConnectionMethod_SSM; return &v }(),
			IamInstanceProfileArn: newIamProfile("arn:aws:iam::123456789012:instance-profile/ssm"),
			RootVolumeSizeGb:      proto.Int32(30),
			UserData:              "#!/bin/bash\necho hello",
		}
	})

	ginkgo.It("accepts a valid SSM spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when instance_name is empty", func() {
		spec.InstanceName = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ami_id does not start with ami-", func() {
		spec.AmiId = "image-123"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when instance_type is empty", func() {
		spec.InstanceType = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when security_group_ids is empty", func() {
		spec.SecurityGroupIds = []*fk.StringValueOrRef{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when root_volume_size_gb is not greater than 0", func() {
		spec.RootVolumeSizeGb = proto.Int32(0)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when connection_method is an undefined enum value", func() {
		spec.ConnectionMethod = func() *AwsEc2InstanceConnectionMethod { v := AwsEc2InstanceConnectionMethod(99); return &v }()
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when connection_method is SSM and iam_instance_profile_arn is not set (CEL)", func() {
		spec.ConnectionMethod = func() *AwsEc2InstanceConnectionMethod { v := AwsEc2InstanceConnectionMethod_SSM; return &v }()
		spec.IamInstanceProfileArn = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when connection_method is BASTION and key_name is empty (CEL)", func() {
		spec.ConnectionMethod = func() *AwsEc2InstanceConnectionMethod { v := AwsEc2InstanceConnectionMethod_BASTION; return &v }()
		spec.IamInstanceProfileArn = nil
		spec.KeyName = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a valid BASTION spec with key_name set", func() {
		spec.ConnectionMethod = func() *AwsEc2InstanceConnectionMethod { v := AwsEc2InstanceConnectionMethod_BASTION; return &v }()
		spec.IamInstanceProfileArn = nil
		spec.KeyName = "my-key"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})
})
