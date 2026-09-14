package azureaksclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureAksClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureAksClusterSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

const (
	testSubnetId    = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/nodes"
	testWorkspaceId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.OperationalInsights/workspaces/my-law"
	testIdentityId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/my-uai"
)

// validResource returns a minimal valid AzureAksCluster that individual
// cases then mutate into the shape under test.
func validResource() *AzureAksCluster {
	return &AzureAksCluster{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureAksCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-cluster",
		},
		Spec: &AzureAksClusterSpec{
			ResourceGroup: ref("my-rg"),
			Region:        "eastus",
			Name:          "my-aks",
			DefaultNodePool: &AzureAksClusterDefaultNodePool{
				Name:      "system",
				VmSize:    "Standard_D4s_v5",
				NodeCount: 1,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureAksClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			err := protovalidate.Validate(validResource())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an autoscaled default pool without node_count", func() {
			input := validResource()
			input.Spec.DefaultNodePool.NodeCount = 0
			input.Spec.DefaultNodePool.AutoScalingEnabled = true
			input.Spec.DefaultNodePool.MinCount = 3
			input.Spec.DefaultNodePool.MaxCount = 5
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a zoned production default pool with subnet, labels, and rotation", func() {
			input := validResource()
			input.Spec.DefaultNodePool.Zones = []string{"1", "2", "3"}
			input.Spec.DefaultNodePool.VnetSubnetId = literal(testSubnetId)
			input.Spec.DefaultNodePool.NodeLabels = map[string]string{"tier": "system"}
			input.Spec.DefaultNodePool.OnlyCriticalAddonsEnabled = true
			input.Spec.DefaultNodePool.TemporaryNameForRotation = "systemtmp"
			input.Spec.DefaultNodePool.OsSku = AzureAksClusterOsSku_AZURE_LINUX
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a private cluster with a private DNS zone and private dns prefix", func() {
			input := validResource()
			input.Spec.PrivateClusterEnabled = true
			input.Spec.DnsPrefixPrivateCluster = "myaks"
			input.Spec.PrivateDnsZoneId = literal("System")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept AAD RBAC with azure RBAC and admin groups plus disabled local accounts", func() {
			input := validResource()
			input.Spec.LocalAccountDisabled = true
			input.Spec.AzureActiveDirectoryRoleBasedAccessControl = &AzureAksClusterAadRbac{
				AzureRbacEnabled:    true,
				AdminGroupObjectIds: []string{"11111111-2222-3333-4444-555555555555"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a user-assigned cluster identity with identity ids", func() {
			input := validResource()
			input.Spec.Identity = &AzureAksClusterIdentity{
				Type:        AzureAksClusterIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full network profile with cilium dataplane and overlay", func() {
			input := validResource()
			input.Spec.NetworkProfile = &AzureAksClusterNetworkProfile{
				NetworkPlugin:     AzureAksClusterNetworkPlugin_AZURE_CNI,
				NetworkPluginMode: AzureAksClusterNetworkPluginMode_OVERLAY,
				NetworkPolicy:     AzureAksClusterNetworkPolicy_NETWORK_POLICY_CILIUM,
				NetworkDataPlane:  AzureAksClusterNetworkDataPlane_DATA_PLANE_CILIUM,
				PodCidr:           "10.244.0.0/16",
				ServiceCidr:       "10.0.0.0/16",
				DnsServiceIp:      "10.0.0.10",
				OutboundType:      AzureAksClusterOutboundType_MANAGED_NAT_GATEWAY,
				NatGatewayProfile: &AzureAksClusterNatGatewayProfile{
					IdleTimeoutInMinutes:   10,
					ManagedOutboundIpCount: 2,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a load balancer profile with managed outbound IPs", func() {
			input := validResource()
			input.Spec.NetworkProfile = &AzureAksClusterNetworkProfile{
				LoadBalancerProfile: &AzureAksClusterLoadBalancerProfile{
					ManagedOutboundIpCount: 4,
					IdleTimeoutInMinutes:   30,
					OutboundPortsAllocated: 8000,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept oms agent, defender, and monitor metrics addons", func() {
			input := validResource()
			input.Spec.OmsAgent = &AzureAksClusterOmsAgent{
				LogAnalyticsWorkspaceId:     literal(testWorkspaceId),
				MsiAuthForMonitoringEnabled: true,
			}
			input.Spec.MicrosoftDefender = &AzureAksClusterMicrosoftDefender{
				LogAnalyticsWorkspaceId: ref("my-law"),
			}
			input.Spec.MonitorMetrics = &AzureAksClusterMonitorMetrics{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept add-on blocks declared and switched off", func() {
			input := validResource()
			input.Spec.KeyVaultSecretsProvider = &AzureAksClusterKeyVaultSecretsProvider{Enabled: proto.Bool(false), SecretRotationEnabled: true}
			input.Spec.MonitorMetrics = &AzureAksClusterMonitorMetrics{Enabled: proto.Bool(false)}
			input.Spec.ConfidentialComputing = &AzureAksClusterConfidentialComputing{Enabled: proto.Bool(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept maintenance windows with schedules", func() {
			input := validResource()
			input.Spec.MaintenanceWindow = &AzureAksClusterMaintenanceWindow{
				Allowed: []*AzureAksClusterMaintenanceWindowAllowed{
					{Day: AzureAksClusterWeekDay_SUNDAY, Hours: []int32{1, 2, 3}},
				},
			}
			input.Spec.MaintenanceWindowAutoUpgrade = &AzureAksClusterMaintenanceWindowSchedule{
				Frequency: AzureAksClusterMaintenanceFrequency_WEEKLY,
				Interval:  1,
				Duration:  4,
				DayOfWeek: AzureAksClusterWeekDay_SUNDAY,
				StartTime: "02:00",
				UtcOffset: "+05:30",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an istio service mesh with ingress gateways", func() {
			input := validResource()
			input.Spec.ServiceMeshProfile = &AzureAksClusterServiceMeshProfile{
				Mode:                          AzureAksClusterServiceMeshMode_ISTIO,
				Revisions:                     []string{"asm-1-24"},
				InternalIngressGatewayEnabled: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept cost analysis on the standard tier", func() {
			input := validResource()
			input.Spec.SkuTier = AzureAksClusterSkuTier_STANDARD
			input.Spec.CostAnalysisEnabled = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept LTS support plan on premium tier", func() {
			input := validResource()
			input.Spec.SkuTier = AzureAksClusterSkuTier_PREMIUM
			input.Spec.SupportPlan = AzureAksClusterSupportPlan_AKS_LONG_TERM_SUPPORT
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept windows profile with gmsa and hybrid licensing", func() {
			input := validResource()
			input.Spec.WindowsProfile = &AzureAksClusterWindowsProfile{
				AdminUsername: "azureadmin",
				AdminPassword: "S3cur3P@ssw0rd!",
				License:       AzureAksClusterWindowsLicense_WINDOWS_SERVER,
				Gmsa: &AzureAksClusterWindowsGmsa{
					DnsServer:  "10.0.0.4",
					RootDomain: "corp.example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a kubelet and linux os config on the default pool", func() {
			input := validResource()
			input.Spec.DefaultNodePool.KubeletConfig = &AzureAksClusterKubeletConfig{
				CpuManagerPolicy:      AzureAksClusterCpuManagerPolicy_CPU_MANAGER_STATIC,
				ImageGcHighThreshold:  85,
				ImageGcLowThreshold:   80,
				ContainerLogMaxSizeMb: 50,
				ContainerLogMaxFiles:  5,
			}
			input.Spec.DefaultNodePool.LinuxOsConfig = &AzureAksClusterLinuxOsConfig{
				SysctlConfig: &AzureAksClusterSysctlConfig{
					VmMaxMapCount:           262144,
					NetCoreSomaxconn:        65535,
					NetIpv4TcpKeepaliveTime: 600,
					FsInotifyMaxUserWatches: 1048576,
				},
				TransparentHugePage: AzureAksClusterTransparentHugePage_THP_MADVISE,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept upgrade settings with surge and undrainable behavior", func() {
			input := validResource()
			input.Spec.DefaultNodePool.UpgradeSettings = &AzureAksClusterDefaultNodePoolUpgradeSettings{
				MaxSurge:                  "10%",
				DrainTimeoutInMinutes:     30,
				NodeSoakDurationInMinutes: 5,
				UndrainableNodeBehavior:   AzureAksClusterUndrainableNodeBehavior_CORDON,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept bootstrap cache with a container registry", func() {
			input := validResource()
			input.Spec.BootstrapProfile = &AzureAksClusterBootstrapProfile{
				ArtifactSource:      AzureAksClusterBootstrapArtifactSource_CACHE,
				ContainerRegistryId: ref("my-acr"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept node auto-provisioning in auto mode", func() {
			input := validResource()
			input.Spec.NodeProvisioningProfile = &AzureAksClusterNodeProvisioningProfile{
				Mode: AzureAksClusterNodeProvisioningMode_AUTO,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept image cleaner with interval and user tags", func() {
			input := validResource()
			input.Spec.ImageCleanerEnabled = true
			input.Spec.ImageCleanerIntervalHours = 48
			input.Spec.Tags = map[string]string{"cost-center": "platform"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the AGIC addon anchored to an existing gateway", func() {
			input := validResource()
			input.Spec.IngressApplicationGateway = &AzureAksClusterIngressApplicationGateway{
				GatewayId: ref("my-agw"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept oidc issuer explicitly disabled without workload identity", func() {
			input := validResource()
			input.Spec.OidcIssuerEnabled = proto.Bool(false)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should return an error when resource_group is missing", func() {
			input := validResource()
			input.Spec.ResourceGroup = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when region is missing", func() {
			input := validResource()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when name is missing", func() {
			input := validResource()
			input.Spec.Name = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a name ending with a hyphen", func() {
			input := validResource()
			input.Spec.Name = "bad-name-"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when default_node_pool is missing", func() {
			input := validResource()
			input.Spec.DefaultNodePool = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an invalid default pool name", func() {
			input := validResource()
			input.Spec.DefaultNodePool.Name = "Bad-Name"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when the default pool has neither node_count nor autoscaling", func() {
			input := validResource()
			input.Spec.DefaultNodePool.NodeCount = 0
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when autoscaling bounds are inverted on the default pool", func() {
			input := validResource()
			input.Spec.DefaultNodePool.AutoScalingEnabled = true
			input.Spec.DefaultNodePool.MinCount = 5
			input.Spec.DefaultNodePool.MaxCount = 3
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when min/max are set without autoscaling", func() {
			input := validResource()
			input.Spec.DefaultNodePool.MinCount = 1
			input.Spec.DefaultNodePool.MaxCount = 3
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a Windows os_sku on the default pool", func() {
			input := validResource()
			input.Spec.DefaultNodePool.OsSku = AzureAksClusterOsSku_WINDOWS_2022
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when both dns prefixes are set", func() {
			input := validResource()
			input.Spec.DnsPrefix = "public"
			input.Spec.DnsPrefixPrivateCluster = "private"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a private dns prefix on a public cluster", func() {
			input := validResource()
			input.Spec.DnsPrefixPrivateCluster = "private"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for workload identity with oidc issuer disabled", func() {
			input := validResource()
			input.Spec.WorkloadIdentityEnabled = true
			input.Spec.OidcIssuerEnabled = proto.Bool(false)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for disabled local accounts without AAD RBAC", func() {
			input := validResource()
			input.Spec.LocalAccountDisabled = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for cost analysis on the free tier", func() {
			input := validResource()
			input.Spec.CostAnalysisEnabled = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for LTS without premium tier", func() {
			input := validResource()
			input.Spec.SkuTier = AzureAksClusterSkuTier_STANDARD
			input.Spec.SupportPlan = AzureAksClusterSupportPlan_AKS_LONG_TERM_SUPPORT
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an image cleaner interval without the cleaner enabled", func() {
			input := validResource()
			input.Spec.ImageCleanerIntervalHours = 48
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an out-of-range image cleaner interval", func() {
			input := validResource()
			input.Spec.ImageCleanerEnabled = true
			input.Spec.ImageCleanerIntervalHours = 10
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a user-assigned identity without identity ids", func() {
			input := validResource()
			input.Spec.Identity = &AzureAksClusterIdentity{
				Type: AzureAksClusterIdentityType_USER_ASSIGNED,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for identity ids on a system-assigned identity", func() {
			input := validResource()
			input.Spec.Identity = &AzureAksClusterIdentity{
				Type:        AzureAksClusterIdentityType_SYSTEM_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for cilium network policy without the cilium dataplane", func() {
			input := validResource()
			input.Spec.NetworkProfile = &AzureAksClusterNetworkProfile{
				NetworkPolicy: AzureAksClusterNetworkPolicy_NETWORK_POLICY_CILIUM,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for the cilium dataplane on kubenet", func() {
			input := validResource()
			input.Spec.NetworkProfile = &AzureAksClusterNetworkProfile{
				NetworkPlugin:    AzureAksClusterNetworkPlugin_KUBENET,
				NetworkDataPlane: AzureAksClusterNetworkDataPlane_DATA_PLANE_CILIUM,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for overlay mode on kubenet", func() {
			input := validResource()
			input.Spec.NetworkProfile = &AzureAksClusterNetworkProfile{
				NetworkPlugin:     AzureAksClusterNetworkPlugin_KUBENET,
				NetworkPluginMode: AzureAksClusterNetworkPluginMode_OVERLAY,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for two outbound IP strategies on the load balancer", func() {
			input := validResource()
			input.Spec.NetworkProfile = &AzureAksClusterNetworkProfile{
				LoadBalancerProfile: &AzureAksClusterLoadBalancerProfile{
					ManagedOutboundIpCount: 2,
					OutboundIpAddressIds:   []*foreignkeyv1.StringValueOrRef{ref("egress-ip")},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an out-of-range load balancer idle timeout", func() {
			input := validResource()
			input.Spec.NetworkProfile = &AzureAksClusterNetworkProfile{
				LoadBalancerProfile: &AzureAksClusterLoadBalancerProfile{
					IdleTimeoutInMinutes: 200,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an invalid authorized ip range", func() {
			input := validResource()
			input.Spec.ApiServerAccessProfile = &AzureAksClusterApiServerAccessProfile{
				AuthorizedIpRanges: []string{"not-a-cidr"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a non-uuid AAD admin group id", func() {
			input := validResource()
			input.Spec.AzureActiveDirectoryRoleBasedAccessControl = &AzureAksClusterAadRbac{
				AdminGroupObjectIds: []string{"not-a-uuid"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for the AGIC addon without an anchor", func() {
			input := validResource()
			input.Spec.IngressApplicationGateway = &AzureAksClusterIngressApplicationGateway{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for the AGIC addon with two anchors", func() {
			input := validResource()
			input.Spec.IngressApplicationGateway = &AzureAksClusterIngressApplicationGateway{
				GatewayId:  ref("my-agw"),
				SubnetCidr: "10.225.0.0/24",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a maintenance schedule with a short duration", func() {
			input := validResource()
			input.Spec.MaintenanceWindowAutoUpgrade = &AzureAksClusterMaintenanceWindowSchedule{
				Frequency: AzureAksClusterMaintenanceFrequency_DAILY,
				Interval:  1,
				Duration:  2,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a malformed maintenance start_time", func() {
			input := validResource()
			input.Spec.MaintenanceWindowAutoUpgrade = &AzureAksClusterMaintenanceWindowSchedule{
				Frequency: AzureAksClusterMaintenanceFrequency_DAILY,
				Interval:  1,
				Duration:  4,
				StartTime: "2am",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an istio mesh without revisions", func() {
			input := validResource()
			input.Spec.ServiceMeshProfile = &AzureAksClusterServiceMeshProfile{
				Mode: AzureAksClusterServiceMeshMode_ISTIO,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for bootstrap cache without a registry", func() {
			input := validResource()
			input.Spec.BootstrapProfile = &AzureAksClusterBootstrapProfile{
				ArtifactSource: AzureAksClusterBootstrapArtifactSource_CACHE,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a short windows admin password", func() {
			input := validResource()
			input.Spec.WindowsProfile = &AzureAksClusterWindowsProfile{
				AdminUsername: "azureadmin",
				AdminPassword: "short",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for gmsa with only one field set", func() {
			input := validResource()
			input.Spec.WindowsProfile = &AzureAksClusterWindowsProfile{
				AdminUsername: "azureadmin",
				AdminPassword: "S3cur3P@ssw0rd!",
				Gmsa: &AzureAksClusterWindowsGmsa{
					DnsServer: "10.0.0.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a node public ip prefix without node public ips", func() {
			input := validResource()
			input.Spec.DefaultNodePool.NodePublicIpPrefixId = ref("my-prefix")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an out-of-range sysctl value", func() {
			input := validResource()
			input.Spec.DefaultNodePool.LinuxOsConfig = &AzureAksClusterLinuxOsConfig{
				SysctlConfig: &AzureAksClusterSysctlConfig{
					VmMaxMapCount: 100,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for upgrade settings without max_surge on the default pool", func() {
			input := validResource()
			input.Spec.DefaultNodePool.UpgradeSettings = &AzureAksClusterDefaultNodePoolUpgradeSettings{
				DrainTimeoutInMinutes: 10,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an invalid zone value", func() {
			input := validResource()
			input.Spec.DefaultNodePool.Zones = []string{"4"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for more than 10 custom CA certificates", func() {
			input := validResource()
			certs := make([]string, 11)
			for i := range certs {
				certs[i] = "Y2VydA=="
			}
			input.Spec.CustomCaTrustCertificatesBase64 = certs
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
