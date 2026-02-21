package alicloudvpngatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudVpnGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudVpnGatewaySpec Validation Tests")
}

func strRef(val string) *fkv1.StringValueOrRef {
	return &fkv1.StringValueOrRef{
		LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: val},
	}
}

func validMinimalSpec() *AliCloudVpnGatewaySpec {
	return &AliCloudVpnGatewaySpec{
		Region:         "cn-hangzhou",
		VpcId:          strRef("vpc-abc123"),
		VswitchId:      strRef("vsw-abc123"),
		VpnGatewayName: "my-vpn",
		Bandwidth:      10,
	}
}

func validMinimalInput() *AliCloudVpnGateway {
	return &AliCloudVpnGateway{
		ApiVersion: "ali-cloud.openmcf.org/v1",
		Kind:       "AliCloudVpnGateway",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-vpn"},
		Spec:       validMinimalSpec(),
	}
}

var _ = ginkgo.Describe("AliCloudVpnGatewaySpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields and no connections", func() {
			input := validMinimalInput()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all top-level optional fields populated", func() {
			input := validMinimalInput()
			input.Spec.Description = "Production VPN Gateway"
			input.Spec.PaymentType = proto.String("Subscription")
			input.Spec.EnableSsl = proto.Bool(true)
			input.Spec.SslConnections = proto.Int32(50)
			input.Spec.Tags = map[string]string{"team": "platform"}
			input.Spec.ResourceGroupId = "rg-abc123"
			input.Metadata.Org = "acme"
			input.Metadata.Env = "production"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a single connection using minimal fields", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "office-hq",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a fully-configured connection", func() {
			input := validMinimalInput()
			input.Spec.Bandwidth = 100
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:               "datacenter-primary",
					CustomerGatewayIp:  "198.51.100.1",
					CustomerGatewayAsn: "65001",
					LocalSubnets:       []string{"10.0.0.0/8", "172.16.0.0/12"},
					RemoteSubnets:      []string{"192.168.1.0/24", "192.168.2.0/24"},
					EnableDpd:          proto.Bool(true),
					EnableNatTraversal: proto.Bool(true),
					EffectImmediately:  proto.Bool(true),
					IkeConfig: &AliCloudIkeConfig{
						Psk:         "my-secret-key-123",
						IkeVersion:  proto.String("ikev2"),
						IkeMode:     proto.String("main"),
						IkeEncAlg:   proto.String("aes256"),
						IkeAuthAlg:  proto.String("sha256"),
						IkePfs:      proto.String("group14"),
						IkeLifetime: proto.Int32(86400),
					},
					IpsecConfig: &AliCloudIpsecConfig{
						IpsecEncAlg:   proto.String("aes256"),
						IpsecAuthAlg:  proto.String("sha256"),
						IpsecPfs:      proto.String("group14"),
						IpsecLifetime: proto.Int32(86400),
					},
					HealthCheckConfig: &AliCloudVpnHealthCheckConfig{
						Enable:   proto.Bool(true),
						Sip:      "10.0.0.1",
						Dip:      "192.168.1.1",
						Interval: proto.Int32(5),
						Retry:    proto.Int32(5),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with multiple connections", func() {
			input := validMinimalInput()
			input.Spec.Bandwidth = 200
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "office-hq",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.1.0/24"},
				},
				{
					Name:              "office-branch",
					CustomerGatewayIp: "203.0.113.2",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.2.0/24"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all valid bandwidth values", func() {
			for _, bw := range []int32{5, 10, 20, 50, 100, 200, 500, 1000} {
				input := validMinimalInput()
				input.Spec.Bandwidth = bw
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with IKEv1 and aggressive mode", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "legacy-site",
					CustomerGatewayIp: "198.51.100.5",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"172.16.0.0/12"},
					IkeConfig: &AliCloudIkeConfig{
						IkeVersion:  proto.String("ikev1"),
						IkeMode:     proto.String("aggressive"),
						IkeEncAlg:   proto.String("3des"),
						IkeAuthAlg:  proto.String("md5"),
						IkePfs:      proto.String("group1"),
						IkeLifetime: proto.Int32(3600),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with IPsec PFS disabled", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "no-pfs-site",
					CustomerGatewayIp: "198.51.100.10",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"172.16.0.0/12"},
					IpsecConfig: &AliCloudIpsecConfig{
						IpsecPfs: proto.String("disabled"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with zero lifetime (no expiry)", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "no-rekey",
					CustomerGatewayIp: "198.51.100.20",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"172.16.0.0/12"},
					IkeConfig: &AliCloudIkeConfig{
						IkeLifetime: proto.Int32(0),
					},
					IpsecConfig: &AliCloudIpsecConfig{
						IpsecLifetime: proto.Int32(0),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PayAsYouGo payment type", func() {
			input := validMinimalInput()
			input.Spec.PaymentType = proto.String("PayAsYouGo")
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

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := validMinimalInput()
			input.Spec.VpcId = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_id is missing", func() {
			input := validMinimalInput()
			input.Spec.VswitchId = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpn_gateway_name is too short", func() {
			input := validMinimalInput()
			input.Spec.VpnGatewayName = "x"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when bandwidth is invalid", func() {
			for _, bw := range []int32{0, 1, 3, 7, 15, 25, 75, 150, 300, 999, 2000} {
				input := validMinimalInput()
				input.Spec.Bandwidth = bw
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should fail when payment_type is invalid", func() {
			input := validMinimalInput()
			input.Spec.PaymentType = proto.String("Free")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when connection name is too short", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "x",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when customer_gateway_ip is empty", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:          "missing-ip",
					LocalSubnets:  []string{"10.0.0.0/8"},
					RemoteSubnets: []string{"192.168.0.0/16"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when local_subnets is empty", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "no-local",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{},
					RemoteSubnets:     []string{"192.168.0.0/16"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when remote_subnets is empty", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "no-remote",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ike_version is invalid", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-ike",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IkeConfig: &AliCloudIkeConfig{
						IkeVersion: proto.String("ikev3"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ike_mode is invalid", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-mode",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IkeConfig: &AliCloudIkeConfig{
						IkeMode: proto.String("hybrid"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ike_enc_alg is invalid", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-enc",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IkeConfig: &AliCloudIkeConfig{
						IkeEncAlg: proto.String("blowfish"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ike_auth_alg is invalid", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-auth",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IkeConfig: &AliCloudIkeConfig{
						IkeAuthAlg: proto.String("sha3"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ike_pfs is invalid", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-pfs",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IkeConfig: &AliCloudIkeConfig{
						IkePfs: proto.String("group99"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ike_lifetime exceeds maximum", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-lifetime",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IkeConfig: &AliCloudIkeConfig{
						IkeLifetime: proto.Int32(86401),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ipsec_enc_alg is invalid", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-ipsec-enc",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IpsecConfig: &AliCloudIpsecConfig{
						IpsecEncAlg: proto.String("rc4"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ipsec_pfs is invalid", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-ipsec-pfs",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IpsecConfig: &AliCloudIpsecConfig{
						IpsecPfs: proto.String("group99"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ipsec_lifetime exceeds maximum", func() {
			input := validMinimalInput()
			input.Spec.Connections = []*AliCloudVpnConnection{
				{
					Name:              "bad-ipsec-life",
					CustomerGatewayIp: "203.0.113.1",
					LocalSubnets:      []string{"10.0.0.0/8"},
					RemoteSubnets:     []string{"192.168.0.0/16"},
					IpsecConfig: &AliCloudIpsecConfig{
						IpsecLifetime: proto.Int32(100000),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
