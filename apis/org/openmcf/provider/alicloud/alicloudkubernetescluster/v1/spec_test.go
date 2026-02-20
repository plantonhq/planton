package alicloudkubernetesclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudKubernetesClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudKubernetesClusterSpec Validation Tests")
}

func strRef(s string) *fkv1.StringValueOrRef {
	return &fkv1.StringValueOrRef{
		LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: s},
	}
}

func minimalValidSpec() *AlicloudKubernetesClusterSpec {
	return &AlicloudKubernetesClusterSpec{
		Region:      "cn-hangzhou",
		VswitchIds:  []*fkv1.StringValueOrRef{strRef("vsw-aaa111"), strRef("vsw-bbb222")},
		ServiceCidr: "172.21.0.0/20",
	}
}

func minimalValidInput() *AlicloudKubernetesCluster {
	return &AlicloudKubernetesCluster{
		ApiVersion: "alicloud.openmcf.org/v1",
		Kind:       "AlicloudKubernetesCluster",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-cluster"},
		Spec:       minimalValidSpec(),
	}
}

var _ = ginkgo.Describe("AlicloudKubernetesClusterSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			err := protovalidate.Validate(minimalValidInput())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a single vswitch", func() {
			input := minimalValidInput()
			input.Spec.VswitchIds = []*fkv1.StringValueOrRef{strRef("vsw-aaa111")}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all core identity fields", func() {
			input := minimalValidInput()
			input.Spec.Name = "prod-ack-cluster"
			input.Spec.Version = "1.30"
			input.Spec.ClusterSpec = proto.String("ack.pro.small")
			input.Spec.ClusterDomain = "cluster.local"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Flannel networking (pod_cidr)", func() {
			input := minimalValidInput()
			input.Spec.PodCidr = "172.20.0.0/16"
			input.Spec.ProxyMode = proto.String("ipvs")
			input.Spec.NodeCidrMask = proto.Int32(24)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Terway networking (pod_vswitch_ids)", func() {
			input := minimalValidInput()
			input.Spec.PodVswitchIds = []*fkv1.StringValueOrRef{
				strRef("vsw-pod-a"),
				strRef("vsw-pod-b"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with security configuration", func() {
			input := minimalValidInput()
			input.Spec.SecurityGroupId = strRef("sg-abc123")
			input.Spec.EnableRrsa = proto.Bool(true)
			input.Spec.DeletionProtection = proto.Bool(true)
			input.Spec.EncryptionProviderKey = strRef("kms-key-id")
			input.Spec.CustomSan = "10.0.0.1,api.example.com"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with enterprise security group", func() {
			input := minimalValidInput()
			input.Spec.IsEnterpriseSecurityGroup = proto.Bool(true)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with addons", func() {
			input := minimalValidInput()
			input.Spec.Addons = []*AlicloudKubernetesAddon{
				{Name: "flannel"},
				{Name: "csi-plugin"},
				{Name: "logtail-ds", Config: `{"IngressDashboardEnabled":"true"}`},
				{Name: "metrics-server", Version: "0.6.4"},
				{Name: "nginx-ingress-controller", Disabled: true},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with logging configuration", func() {
			input := minimalValidInput()
			input.Spec.Logging = &AlicloudKubernetesClusterLogging{
				ControlPlaneLogProject:    strRef("my-sls-project"),
				ControlPlaneLogTtl:        proto.String("60"),
				ControlPlaneLogComponents: []string{"apiserver", "kcm", "scheduler"},
				AuditLogEnabled:           true,
				AuditLogSlsProject:        "audit-sls-project",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with maintenance window", func() {
			input := minimalValidInput()
			input.Spec.MaintenanceWindow = &AlicloudKubernetesClusterMaintenanceWindow{
				Enable:          true,
				MaintenanceTime: "2026-03-01T03:00:00+08:00",
				Duration:        "3h",
				WeeklyPeriod:    "Monday,Thursday",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with auto-upgrade enabled", func() {
			input := minimalValidInput()
			input.Spec.AutoUpgrade = &AlicloudKubernetesClusterAutoUpgrade{
				Enabled: true,
				Channel: proto.String("stable"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with resource management fields", func() {
			input := minimalValidInput()
			input.Spec.Tags = map[string]string{"team": "platform", "env": "prod"}
			input.Spec.ResourceGroupId = "rg-abc123"
			input.Spec.Timezone = "Asia/Shanghai"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with iptables proxy mode", func() {
			input := minimalValidInput()
			input.Spec.ProxyMode = proto.String("iptables")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with NAT gateway disabled", func() {
			input := minimalValidInput()
			input.Spec.NewNatGateway = proto.Bool(false)
			input.Spec.SlbInternetEnabled = proto.Bool(false)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with full production configuration", func() {
			input := minimalValidInput()
			input.Spec.Name = "production-cluster"
			input.Spec.Version = "1.30"
			input.Spec.ClusterSpec = proto.String("ack.pro.small")
			input.Spec.VswitchIds = []*fkv1.StringValueOrRef{
				strRef("vsw-aaa"), strRef("vsw-bbb"), strRef("vsw-ccc"),
			}
			input.Spec.PodVswitchIds = []*fkv1.StringValueOrRef{
				strRef("vsw-pod-a"), strRef("vsw-pod-b"), strRef("vsw-pod-c"),
			}
			input.Spec.ServiceCidr = "172.21.0.0/20"
			input.Spec.NodeCidrMask = proto.Int32(26)
			input.Spec.NewNatGateway = proto.Bool(false)
			input.Spec.EnableRrsa = proto.Bool(true)
			input.Spec.DeletionProtection = proto.Bool(true)
			input.Spec.EncryptionProviderKey = strRef("kms-key-prod")
			input.Spec.Addons = []*AlicloudKubernetesAddon{
				{Name: "terway-eniip"},
				{Name: "csi-plugin"},
				{Name: "csi-provisioner"},
				{Name: "logtail-ds", Config: `{"IngressDashboardEnabled":"true","sls_project_name":"prod-logs"}`},
				{Name: "arms-prometheus"},
				{Name: "metrics-server"},
			}
			input.Spec.Logging = &AlicloudKubernetesClusterLogging{
				ControlPlaneLogProject:    strRef("prod-logs"),
				ControlPlaneLogTtl:        proto.String("90"),
				ControlPlaneLogComponents: []string{"apiserver", "kcm", "scheduler", "ccm", "controlplane-events"},
				AuditLogEnabled:           true,
			}
			input.Spec.MaintenanceWindow = &AlicloudKubernetesClusterMaintenanceWindow{
				Enable:          true,
				MaintenanceTime: "2026-03-01T03:00:00+08:00",
				Duration:        "3h",
				WeeklyPeriod:    "Wednesday",
			}
			input.Spec.AutoUpgrade = &AlicloudKubernetesClusterAutoUpgrade{
				Enabled: true,
				Channel: proto.String("patch"),
			}
			input.Spec.Tags = map[string]string{"team": "platform"}
			input.Spec.Timezone = "Asia/Shanghai"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := minimalValidInput()
			input.ApiVersion = "wrong/v1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := minimalValidInput()
			input.Kind = "WrongKind"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := minimalValidInput()
			input.Metadata = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AlicloudKubernetesCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudKubernetesCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := minimalValidInput()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_ids is empty", func() {
			input := minimalValidInput()
			input.Spec.VswitchIds = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_ids exceeds 5", func() {
			input := minimalValidInput()
			input.Spec.VswitchIds = []*fkv1.StringValueOrRef{
				strRef("vsw-1"), strRef("vsw-2"), strRef("vsw-3"),
				strRef("vsw-4"), strRef("vsw-5"), strRef("vsw-6"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when service_cidr is empty", func() {
			input := minimalValidInput()
			input.Spec.ServiceCidr = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cluster_spec is invalid", func() {
			input := minimalValidInput()
			input.Spec.ClusterSpec = proto.String("ack.enterprise")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when proxy_mode is invalid", func() {
			input := minimalValidInput()
			input.Spec.ProxyMode = proto.String("nftables")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when node_cidr_mask is below range", func() {
			input := minimalValidInput()
			input.Spec.NodeCidrMask = proto.Int32(20)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when node_cidr_mask is above range", func() {
			input := minimalValidInput()
			input.Spec.NodeCidrMask = proto.Int32(30)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when name exceeds 63 characters", func() {
			input := minimalValidInput()
			input.Spec.Name = "this-cluster-name-is-way-too-long-and-exceeds-the-sixty-three-character-limit-imposed-by-the-spec"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when addon name is empty", func() {
			input := minimalValidInput()
			input.Spec.Addons = []*AlicloudKubernetesAddon{
				{Name: ""},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auto_upgrade channel is invalid", func() {
			input := minimalValidInput()
			input.Spec.AutoUpgrade = &AlicloudKubernetesClusterAutoUpgrade{
				Enabled: true,
				Channel: proto.String("bleeding-edge"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
