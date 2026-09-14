package azurevirtualmachinescalesetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualMachineScaleSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualMachineScaleSetSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testSubnetId = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/app"
	testPoolId   = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb/backendAddressPools/web"
	testProbeId  = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb/probes/http-health"
	testUaiId    = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ci"
)

func intPtr(v int32) *int32 { return &v }

func boolPtr(v bool) *bool { return &v }

// healthExtension returns the application health extension rolling
// upgrades, instance repair, and platform patching key on.
func healthExtension() *AzureVirtualMachineScaleSetExtension {
	return &AzureVirtualMachineScaleSetExtension{
		Name:               "health",
		Publisher:          "Microsoft.ManagedServices",
		Type:               "ApplicationHealthLinux",
		TypeHandlerVersion: "1.0",
		Settings:           `{"protocol":"tcp","port":22}`,
	}
}

// flexibleLinux returns a minimal valid FLEXIBLE Linux scale set the
// cases mutate one contract at a time.
func flexibleLinux() *AzureVirtualMachineScaleSet {
	return &AzureVirtualMachineScaleSet{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualMachineScaleSet",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-vmss",
		},
		Spec: &AzureVirtualMachineScaleSetSpec{
			Region:                   "eastus",
			ResourceGroup:            literal("test-rg"),
			Name:                     "web-fleet",
			SkuName:                  "Standard_B2s",
			Instances:                intPtr(2),
			PlatformFaultDomainCount: intPtr(1),
			OsProfile: &AzureVirtualMachineScaleSetOsProfile{
				Linux: &AzureVirtualMachineScaleSetLinuxProfile{
					AdminUsername: "azureuser",
					SshPublicKeys: []*AzureVirtualMachineScaleSetSshPublicKey{
						{PublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest"},
					},
				},
			},
			OsDisk: &AzureVirtualMachineScaleSetOsDisk{
				Caching:            AzureVirtualMachineScaleSetDiskCaching_READ_WRITE,
				StorageAccountType: AzureVirtualMachineScaleSetOsDiskStorageAccountType_STANDARD_LRS,
			},
			SourceImageReference: &AzureVirtualMachineScaleSetSourceImageReference{
				Publisher: "Canonical",
				Offer:     "ubuntu-24_04-lts",
				Sku:       "server",
				Version:   "latest",
			},
			NetworkInterfaces: []*AzureVirtualMachineScaleSetNetworkInterface{
				{
					Name:    "primary",
					Primary: true,
					IpConfigurations: []*AzureVirtualMachineScaleSetIpConfiguration{
						{
							Name:     "internal",
							Primary:  true,
							SubnetId: literal(testSubnetId),
						},
					},
				},
			},
		},
	}
}

// uniformLinux returns a minimal valid UNIFORM Linux scale set.
func uniformLinux() *AzureVirtualMachineScaleSet {
	input := flexibleLinux()
	input.Spec.OrchestrationMode = AzureVirtualMachineScaleSetOrchestrationMode_UNIFORM
	input.Spec.PlatformFaultDomainCount = nil
	return input
}

var _ = ginkgo.Describe("AzureVirtualMachineScaleSetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal FLEXIBLE Linux scale set", func() {
			err := protovalidate.Validate(flexibleLinux())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept boot diagnostics declared and switched off", func() {
			input := flexibleLinux()
			input.Spec.BootDiagnostics = &AzureVirtualMachineScaleSetBootDiagnostics{Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a minimal UNIFORM Linux scale set", func() {
			err := protovalidate.Validate(uniformLinux())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept UNIFORM-only capabilities on a UNIFORM set", func() {
			input := uniformLinux()
			input.Spec.Overprovision = boolPtr(false)
			input.Spec.ScaleIn = &AzureVirtualMachineScaleSetScaleIn{
				Rule: AzureVirtualMachineScaleSetScaleInRule_OLDEST_VM,
			}
			input.Spec.DoNotRunExtensionsOnOverprovisionedMachines = boolPtr(true)
			input.Spec.Security = &AzureVirtualMachineScaleSetSecurity{
				SecureBootEnabled: true,
				VtpmEnabled:       true,
			}
			input.Spec.Placement = &AzureVirtualMachineScaleSetPlacement{
				HostGroupId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/hostGroups/hg",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a UNIFORM rolling upgrade gated by a load-balancer probe", func() {
			input := uniformLinux()
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode: AzureVirtualMachineScaleSetUpgradeMode_ROLLING,
				Rolling: &AzureVirtualMachineScaleSetRollingUpgradePolicy{
					MaxBatchInstancePercent:             20,
					MaxUnhealthyInstancePercent:         20,
					MaxUnhealthyUpgradedInstancePercent: 20,
					PauseTimeBetweenBatches:             "PT30S",
				},
				AutomaticOsUpgrade: &AzureVirtualMachineScaleSetAutomaticOsUpgrade{
					Enabled: true,
				},
				HealthProbeId: literal(testProbeId),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a FLEXIBLE rolling upgrade gated by the health extension", func() {
			input := flexibleLinux()
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{healthExtension()}
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode: AzureVirtualMachineScaleSetUpgradeMode_ROLLING,
				Rolling: &AzureVirtualMachineScaleSetRollingUpgradePolicy{
					MaxBatchInstancePercent:             20,
					MaxUnhealthyInstancePercent:         20,
					MaxUnhealthyUpgradedInstancePercent: 20,
					PauseTimeBetweenBatches:             "PT30S",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a FLEXIBLE mixed-SKU spot fleet with a priority mix", func() {
			input := flexibleLinux()
			input.Spec.SkuName = "Mix"
			input.Spec.SkuProfile = &AzureVirtualMachineScaleSetSkuProfile{
				AllocationStrategy: AzureVirtualMachineScaleSetAllocationStrategy_CAPACITY_OPTIMIZED,
				VmSizes: []*AzureVirtualMachineScaleSetVmSize{
					{Name: "Standard_D2s_v3"},
					{Name: "Standard_D2as_v4"},
				},
			}
			input.Spec.Spot = &AzureVirtualMachineScaleSetSpot{
				EvictionPolicy: AzureVirtualMachineScaleSetEvictionPolicy_DELETE,
				PriorityMix: &AzureVirtualMachineScaleSetPriorityMix{
					BaseRegularCount:           intPtr(2),
					RegularPercentageAboveBase: intPtr(25),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a UNIFORM spot fleet with restore", func() {
			input := uniformLinux()
			input.Spec.Spot = &AzureVirtualMachineScaleSetSpot{
				EvictionPolicy: AzureVirtualMachineScaleSetEvictionPolicy_DEALLOCATE,
				Restore: &AzureVirtualMachineScaleSetSpotRestore{
					Timeout: "PT1H",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a FLEXIBLE Windows fleet with platform patching and hotpatching", func() {
			input := flexibleLinux()
			healthWin := healthExtension()
			healthWin.Type = "ApplicationHealthWindows"
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{healthWin}
			input.Spec.OsProfile = &AzureVirtualMachineScaleSetOsProfile{
				ComputerNamePrefix: "webw",
				Windows: &AzureVirtualMachineScaleSetWindowsProfile{
					AdminUsername:      "azureadmin",
					AdminPassword:      literal("S3cure!Passw0rd"),
					PatchMode:          AzureVirtualMachineScaleSetWindowsPatchMode_WINDOWS_AUTOMATIC_BY_PLATFORM,
					HotpatchingEnabled: true,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept load-balancer pool membership, zones, and user-assigned identity", func() {
			input := flexibleLinux()
			input.Spec.Zones = []string{"1", "2", "3"}
			input.Spec.ZoneBalance = boolPtr(true)
			input.Spec.NetworkInterfaces[0].IpConfigurations[0].LoadBalancerBackendAddressPoolIds = []*foreignkeyv1.StringValueOrRef{literal(testPoolId)}
			input.Spec.Identity = &AzureVirtualMachineScaleSetIdentity{
				Type:        AzureVirtualMachineScaleSetIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testUaiId)},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept data disks with dialed Ultra performance", func() {
			iops := int64(4000)
			mbps := int64(200)
			input := flexibleLinux()
			input.Spec.DataDisks = []*AzureVirtualMachineScaleSetDataDisk{
				{
					Lun:                       intPtr(0),
					Caching:                   AzureVirtualMachineScaleSetDiskCaching_NONE,
					DiskSizeGb:                256,
					StorageAccountType:        AzureVirtualMachineScaleSetDataDiskStorageAccountType_ULTRA_SSD_LRS,
					UltraSsdDiskIopsReadWrite: &iops,
					UltraSsdDiskMbpsReadWrite: &mbps,
				},
			}
			input.Spec.AdditionalCapabilities = &AzureVirtualMachineScaleSetAdditionalCapabilities{
				UltraSsdEnabled: true,
			}
			input.Spec.Zones = []string{"1"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept per-instance public IPs and an ephemeral OS disk", func() {
			input := flexibleLinux()
			input.Spec.NetworkInterfaces[0].IpConfigurations[0].PublicIpAddress = &AzureVirtualMachineScaleSetPublicIpAddress{
				Name:            "instance-ip",
				DomainNameLabel: "web-fleet",
			}
			input.Spec.OsDisk = &AzureVirtualMachineScaleSetOsDisk{
				Caching:            AzureVirtualMachineScaleSetDiskCaching_READ_ONLY,
				StorageAccountType: AzureVirtualMachineScaleSetOsDiskStorageAccountType_STANDARD_LRS,
				DiffDiskSettings: &AzureVirtualMachineScaleSetDiffDiskSettings{
					Placement: AzureVirtualMachineScaleSetDiffDiskPlacement_CACHE_DISK,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept UNIFORM pool-style NAT rule membership", func() {
			input := uniformLinux()
			input.Spec.NetworkInterfaces[0].IpConfigurations[0].LoadBalancerInboundNatRuleIds = []*foreignkeyv1.StringValueOrRef{
				literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb/inboundNatRules/per-instance-ssh"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing image source", func() {
			input := flexibleLinux()
			input.Spec.SourceImageReference = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject two image sources", func() {
			input := flexibleLinux()
			input.Spec.SourceImageId = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/galleries/g/images/i"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a Linux fleet with no credentials", func() {
			input := flexibleLinux()
			input.Spec.OsProfile.Linux.SshPublicKeys = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a Windows fleet without a password", func() {
			input := flexibleLinux()
			input.Spec.OsProfile = &AzureVirtualMachineScaleSetOsProfile{
				Windows: &AzureVirtualMachineScaleSetWindowsProfile{
					AdminUsername: "azureadmin",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject both OS profiles", func() {
			input := flexibleLinux()
			input.Spec.OsProfile.Windows = &AzureVirtualMachineScaleSetWindowsProfile{
				AdminUsername: "azureadmin",
				AdminPassword: literal("S3cure!Passw0rd"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a FLEXIBLE set without platform_fault_domain_count", func() {
			input := flexibleLinux()
			input.Spec.PlatformFaultDomainCount = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a system-assigned identity on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.Identity = &AzureVirtualMachineScaleSetIdentity{
				Type: AzureVirtualMachineScaleSetIdentityType_SYSTEM_ASSIGNED,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject overprovision on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.Overprovision = boolPtr(true)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject scale_in on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.ScaleIn = &AzureVirtualMachineScaleSetScaleIn{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject secure boot on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.Security = &AzureVirtualMachineScaleSetSecurity{SecureBootEnabled: true}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject gallery applications on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.GalleryApplications = []*AzureVirtualMachineScaleSetGalleryApplication{
				{VersionId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/galleries/g/applications/a/versions/1.0.0"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a health probe on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{healthExtension()}
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				HealthProbeId: literal(testProbeId),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spot restore on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.Spot = &AzureVirtualMachineScaleSetSpot{
				EvictionPolicy: AzureVirtualMachineScaleSetEvictionPolicy_DELETE,
				Restore:        &AzureVirtualMachineScaleSetSpotRestore{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a sku_profile on a UNIFORM set", func() {
			input := uniformLinux()
			input.Spec.SkuName = "Mix"
			input.Spec.SkuProfile = &AzureVirtualMachineScaleSetSkuProfile{
				AllocationStrategy: AzureVirtualMachineScaleSetAllocationStrategy_LOWEST_PRICE,
				VmSizes:            []*AzureVirtualMachineScaleSetVmSize{{Name: "Standard_D2s_v3"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a priority mix on a UNIFORM set", func() {
			input := uniformLinux()
			input.Spec.Spot = &AzureVirtualMachineScaleSetSpot{
				EvictionPolicy: AzureVirtualMachineScaleSetEvictionPolicy_DELETE,
				PriorityMix:    &AzureVirtualMachineScaleSetPriorityMix{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject network_api_version on a UNIFORM set", func() {
			input := uniformLinux()
			input.Spec.NetworkApiVersion = "2022-11-01"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject patch modes on a UNIFORM set", func() {
			input := uniformLinux()
			input.Spec.OsProfile.Linux.PatchMode = AzureVirtualMachineScaleSetLinuxPatchMode_LINUX_AUTOMATIC_BY_PLATFORM
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{healthExtension()}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject NAT rule membership on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.NetworkInterfaces[0].IpConfigurations[0].LoadBalancerInboundNatRuleIds = []*foreignkeyv1.StringValueOrRef{
				literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb/inboundNatRules/r"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a named data disk on a FLEXIBLE set", func() {
			input := flexibleLinux()
			input.Spec.DataDisks = []*AzureVirtualMachineScaleSetDataDisk{
				{
					Lun:                intPtr(0),
					Caching:            AzureVirtualMachineScaleSetDiskCaching_READ_ONLY,
					DiskSizeGb:         64,
					StorageAccountType: AzureVirtualMachineScaleSetDataDiskStorageAccountType_DATA_STANDARD_SSD_LRS,
					Name:               "explicit-name",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject extension automatic major-version upgrades on a FLEXIBLE set", func() {
			input := flexibleLinux()
			ext := healthExtension()
			ext.AutomaticUpgradeEnabled = true
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{ext}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject extension failure suppression on a UNIFORM set", func() {
			input := uniformLinux()
			ext := healthExtension()
			ext.FailureSuppressionEnabled = true
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{ext}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject sku_name Mix without a sku_profile", func() {
			input := flexibleLinux()
			input.Spec.SkuName = "Mix"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject zone_balance without zones", func() {
			input := flexibleLinux()
			input.Spec.ZoneBalance = boolPtr(true)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject ROLLING mode without a rolling policy", func() {
			input := uniformLinux()
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode:          AzureVirtualMachineScaleSetUpgradeMode_ROLLING,
				HealthProbeId: literal(testProbeId),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject ROLLING mode without health monitoring", func() {
			input := uniformLinux()
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode: AzureVirtualMachineScaleSetUpgradeMode_ROLLING,
				Rolling: &AzureVirtualMachineScaleSetRollingUpgradePolicy{
					MaxBatchInstancePercent:             20,
					MaxUnhealthyInstancePercent:         20,
					MaxUnhealthyUpgradedInstancePercent: 20,
					PauseTimeBetweenBatches:             "PT30S",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rolling policy on MANUAL mode", func() {
			input := uniformLinux()
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode: AzureVirtualMachineScaleSetUpgradeMode_MANUAL,
				Rolling: &AzureVirtualMachineScaleSetRollingUpgradePolicy{
					MaxBatchInstancePercent:             20,
					MaxUnhealthyInstancePercent:         20,
					MaxUnhealthyUpgradedInstancePercent: 20,
					PauseTimeBetweenBatches:             "PT30S",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject automatic OS upgrades on MANUAL mode", func() {
			input := uniformLinux()
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode:               AzureVirtualMachineScaleSetUpgradeMode_MANUAL,
				AutomaticOsUpgrade: &AzureVirtualMachineScaleSetAutomaticOsUpgrade{Enabled: true},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject cross-zone upgrades without zones", func() {
			input := uniformLinux()
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode: AzureVirtualMachineScaleSetUpgradeMode_ROLLING,
				Rolling: &AzureVirtualMachineScaleSetRollingUpgradePolicy{
					MaxBatchInstancePercent:             20,
					MaxUnhealthyInstancePercent:         20,
					MaxUnhealthyUpgradedInstancePercent: 20,
					PauseTimeBetweenBatches:             "PT30S",
					CrossZoneUpgradesEnabled:            true,
				},
				HealthProbeId: literal(testProbeId),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject surge upgrades without overprovision explicitly false", func() {
			input := uniformLinux()
			input.Spec.UpgradePolicy = &AzureVirtualMachineScaleSetUpgradePolicy{
				Mode: AzureVirtualMachineScaleSetUpgradeMode_ROLLING,
				Rolling: &AzureVirtualMachineScaleSetRollingUpgradePolicy{
					MaxBatchInstancePercent:             20,
					MaxUnhealthyInstancePercent:         20,
					MaxUnhealthyUpgradedInstancePercent: 20,
					PauseTimeBetweenBatches:             "PT30S",
					MaximumSurgeInstancesEnabled:        true,
				},
				HealthProbeId: literal(testProbeId),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject instance repair without health monitoring", func() {
			input := flexibleLinux()
			input.Spec.AutomaticInstanceRepair = &AzureVirtualMachineScaleSetAutomaticInstanceRepair{
				Enabled: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject hotpatching without platform patching", func() {
			input := flexibleLinux()
			healthWin := healthExtension()
			healthWin.Type = "ApplicationHealthWindows"
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{healthWin}
			input.Spec.OsProfile = &AzureVirtualMachineScaleSetOsProfile{
				Windows: &AzureVirtualMachineScaleSetWindowsProfile{
					AdminUsername:      "azureadmin",
					AdminPassword:      literal("S3cure!Passw0rd"),
					HotpatchingEnabled: true,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject platform patching without the health extension", func() {
			input := flexibleLinux()
			input.Spec.OsProfile.Linux.PatchMode = AzureVirtualMachineScaleSetLinuxPatchMode_LINUX_AUTOMATIC_BY_PLATFORM
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject eviction policy without spot presence semantics", func() {
			input := flexibleLinux()
			input.Spec.Spot = &AzureVirtualMachineScaleSetSpot{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject dialed Ultra performance on a standard data disk", func() {
			iops := int64(4000)
			input := flexibleLinux()
			input.Spec.DataDisks = []*AzureVirtualMachineScaleSetDataDisk{
				{
					Lun:                       intPtr(0),
					Caching:                   AzureVirtualMachineScaleSetDiskCaching_NONE,
					DiskSizeGb:                256,
					StorageAccountType:        AzureVirtualMachineScaleSetDataDiskStorageAccountType_DATA_PREMIUM_LRS,
					UltraSsdDiskIopsReadWrite: &iops,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an ephemeral OS disk without READ_ONLY caching", func() {
			input := flexibleLinux()
			input.Spec.OsDisk.DiffDiskSettings = &AzureVirtualMachineScaleSetDiffDiskSettings{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject confidential encryption without vTPM", func() {
			input := uniformLinux()
			input.Spec.OsDisk.SecurityEncryptionType = AzureVirtualMachineScaleSetSecurityEncryptionType_VM_GUEST_STATE_ONLY
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HTTPS WinRM listener without a certificate", func() {
			input := flexibleLinux()
			input.Spec.OsProfile = &AzureVirtualMachineScaleSetOsProfile{
				Windows: &AzureVirtualMachineScaleSetWindowsProfile{
					AdminUsername: "azureadmin",
					AdminPassword: literal("S3cure!Passw0rd"),
					WinrmListeners: []*AzureVirtualMachineScaleSetWinrmListener{
						{Protocol: AzureVirtualMachineScaleSetWinrmProtocol_HTTPS},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject user-assigned identity type without identity ids", func() {
			input := flexibleLinux()
			input.Spec.Identity = &AzureVirtualMachineScaleSetIdentity{
				Type: AzureVirtualMachineScaleSetIdentityType_USER_ASSIGNED,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a capacity reservation without single_placement_group false", func() {
			input := uniformLinux()
			input.Spec.Placement = &AzureVirtualMachineScaleSetPlacement{
				CapacityReservationGroupId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/capacityReservationGroups/crg",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject inline and Key-Vault protected settings together", func() {
			input := flexibleLinux()
			ext := healthExtension()
			ext.ProtectedSettings = `{"key":"secret"}`
			ext.ProtectedSettingsFromKeyVault = &AzureVirtualMachineScaleSetExtensionProtectedSettingsFromKeyVault{
				SecretUrl:     "https://vault.vault.azure.net/secrets/ext/1",
				SourceVaultId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/vault"),
			}
			input.Spec.Extensions = []*AzureVirtualMachineScaleSetExtension{ext}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject multiple NICs when the first is not primary", func() {
			input := flexibleLinux()
			input.Spec.NetworkInterfaces[0].Primary = false
			input.Spec.NetworkInterfaces = append(input.Spec.NetworkInterfaces, &AzureVirtualMachineScaleSetNetworkInterface{
				Name:    "data",
				Primary: true,
				IpConfigurations: []*AzureVirtualMachineScaleSetIpConfiguration{
					{Name: "data", Primary: true, SubnetId: literal(testSubnetId)},
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject sku_profile ranks without the PRIORITIZED strategy", func() {
			input := flexibleLinux()
			input.Spec.SkuName = "Mix"
			input.Spec.SkuProfile = &AzureVirtualMachineScaleSetSkuProfile{
				AllocationStrategy: AzureVirtualMachineScaleSetAllocationStrategy_LOWEST_PRICE,
				VmSizes: []*AzureVirtualMachineScaleSetVmSize{
					{Name: "Standard_D2s_v3", Rank: intPtr(0)},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
