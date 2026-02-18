package awsclientvpnv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsClientVpnSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsClientVpnSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsClientVpnSpec validations", func() {
	var spec *AwsClientVpnSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsClientVpnSpec{
			Region: "us-west-2",
			VpcId:  &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-12345678"}},
			Subnets: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"}},
			},
			ClientCidrBlock:    "10.0.0.0/22",
			AuthenticationType: AwsClientVpnAuthenticationType(0),
			ServerCertificateArn: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm:us-east-1:123456789012:certificate/abc"},
			},
			VpnPort:           proto.Int32(443),
			TransportProtocol: AwsClientVpnTransportProtocol(2),
		}
	})

	ginkgo.It("accepts a valid spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when vpc_id is missing", func() {
		spec.VpcId = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subnets are empty", func() {
		spec.Subnets = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// Note: uniqueness on repeated message fields is not enforced by protovalidate built-ins
	// and we avoid brittle CEL for cross-item equality on oneofs. Skipping duplicate check here.

	ginkgo.It("fails when client_cidr_block is invalid", func() {
		spec.ClientCidrBlock = "10.0.0.0"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when authentication_type is not certificate", func() {
		spec.AuthenticationType = AwsClientVpnAuthenticationType(2)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when server_certificate_arn is missing", func() {
		spec.ServerCertificateArn = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when cidr_authorization_rules contains invalid CIDR", func() {
		spec.CidrAuthorizationRules = []string{"10.0.0.0/33"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when dns_servers has more than two entries", func() {
		spec.DnsServers = []string{"10.0.0.2", "10.0.0.3", "10.0.0.4"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when dns_servers entry is not a valid IPv4 address", func() {
		spec.DnsServers = []string{"a.b.c.d"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when transport_protocol and vpn_port do not match", func() {
		spec.TransportProtocol = AwsClientVpnTransportProtocol(1)
		spec.VpnPort = proto.Int32(443)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// Note: skipping duplicate check for security_groups for the same reason as subnets.
})
