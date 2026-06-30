package module

import (
	"github.com/pkg/errors"
	azurevirtualmachinev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurevirtualmachine/v1"
	"github.com/pulumi/pulumi-azure-native-sdk/compute/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v3"
	azurenative "github.com/pulumi/pulumi-azure-native-sdk/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurevirtualmachinev1.AzureVirtualMachineStackInput) error {
	azureProviderConfig := stackInput.ProviderConfig

	// Create Azure provider using the credentials from the input
	provider, err := azurenative.NewProvider(ctx,
		"azure",
		&azurenative.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// Get inputs
	locals := initializeLocals(ctx, stackInput)
	target := stackInput.Target
	spec := target.Spec
	metadata := target.Metadata

	// Get VM size (with default)
	vmSize := "Standard_D2s_v3"
	if spec.VmSize != nil && *spec.VmSize != "" {
		vmSize = *spec.VmSize
	}

	// Get admin username (with default)
	adminUsername := "azureuser"
	if spec.AdminUsername != nil && *spec.AdminUsername != "" {
		adminUsername = *spec.AdminUsername
	}

	// Create Network Interface
	nicName := metadata.Name + "-nic"
	nicArgs := &network.NetworkInterfaceArgs{
		NetworkInterfaceName: pulumi.String(nicName),
		ResourceGroupName:    pulumi.String(locals.ResourceGroupName),
		Location:             pulumi.String(spec.Region),
		IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
			&network.NetworkInterfaceIPConfigurationArgs{
				Name:                      pulumi.String("primary"),
				Primary:                   pulumi.Bool(true),
				PrivateIPAllocationMethod: pulumi.String("Dynamic"),
				Subnet: &network.SubnetTypeArgs{
					Id: pulumi.String(spec.SubnetId.GetValue()),
				},
			},
		},
	}

	// Configure accelerated networking if specified
	if spec.Network != nil {
		if spec.Network.EnableAcceleratedNetworking != nil && *spec.Network.EnableAcceleratedNetworking {
			nicArgs.EnableAcceleratedNetworking = pulumi.Bool(true)
		}

		// Associate NSG if specified
		if spec.Network.NetworkSecurityGroupId != nil {
			nsgValue := spec.Network.NetworkSecurityGroupId.GetValue()
			if nsgValue != "" {
				nicArgs.NetworkSecurityGroup = &network.NetworkSecurityGroupTypeArgs{
					Id: pulumi.String(nsgValue),
				}
			}
		}
	}

	nic, err := network.NewNetworkInterface(ctx, nicName, nicArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create network interface")
	}

	// Create Public IP if enabled
	var publicIp *network.PublicIPAddress
	if spec.Network != nil && spec.Network.EnablePublicIp {
		pipName := metadata.Name + "-pip"
		pipSku := "Standard"
		if spec.Network.PublicIpSku != nil {
			if *spec.Network.PublicIpSku == azurevirtualmachinev1.AzureVirtualMachineNetworkConfig_basic {
				pipSku = "Basic"
			}
		}

		pipAllocation := "Static"
		if spec.Network.PublicIpAllocation != nil {
			if *spec.Network.PublicIpAllocation == azurevirtualmachinev1.AzureVirtualMachineNetworkConfig_public_dynamic {
				pipAllocation = "Dynamic"
			}
		}

		publicIp, err = network.NewPublicIPAddress(ctx, pipName, &network.PublicIPAddressArgs{
			PublicIpAddressName:      pulumi.String(pipName),
			ResourceGroupName:        pulumi.String(locals.ResourceGroupName),
			Location:                 pulumi.String(spec.Region),
			PublicIPAllocationMethod: pulumi.String(pipAllocation),
			Sku: &network.PublicIPAddressSkuArgs{
				Name: pulumi.String(pipSku),
			},
			Zones: getAvailabilityZoneArray(spec.AvailabilityZone),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create public IP")
		}
	}

	// Build VM arguments
	vmArgs := &compute.VirtualMachineArgs{
		VmName:            pulumi.String(metadata.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Location:          pulumi.String(spec.Region),
		HardwareProfile: &compute.HardwareProfileArgs{
			VmSize: pulumi.String(vmSize),
		},
		NetworkProfile: &compute.NetworkProfileArgs{
			NetworkInterfaces: compute.NetworkInterfaceReferenceArray{
				&compute.NetworkInterfaceReferenceArgs{
					Id:      nic.ID(),
					Primary: pulumi.Bool(true),
				},
			},
		},
	}

	// Configure OS Profile
	osProfile := &compute.OSProfileArgs{
		ComputerName:  pulumi.String(metadata.Name),
		AdminUsername: pulumi.String(adminUsername),
	}

	// Configure authentication
	if spec.SshPublicKey != "" {
		osProfile.LinuxConfiguration = &compute.LinuxConfigurationArgs{
			DisablePasswordAuthentication: pulumi.Bool(true),
			Ssh: &compute.SshConfigurationArgs{
				PublicKeys: compute.SshPublicKeyTypeArray{
					&compute.SshPublicKeyTypeArgs{
						Path:    pulumi.Sprintf("/home/%s/.ssh/authorized_keys", adminUsername),
						KeyData: pulumi.String(spec.SshPublicKey),
					},
				},
			},
		}
	}

	if spec.AdminPassword != nil {
		passwordValue := spec.AdminPassword.GetValue()
		if passwordValue != "" {
			osProfile.AdminPassword = pulumi.String(passwordValue)
		}
	}

	vmArgs.OsProfile = osProfile

	// Configure storage profile (image and OS disk)
	storageProfile := &compute.StorageProfileArgs{}

	// Configure image reference
	if spec.Image != nil {
		if spec.Image.CustomImageId != "" {
			storageProfile.ImageReference = &compute.ImageReferenceArgs{
				Id: pulumi.String(spec.Image.CustomImageId),
			}
		} else {
			imageVersion := "latest"
			if spec.Image.Version != nil && *spec.Image.Version != "" {
				imageVersion = *spec.Image.Version
			}
			storageProfile.ImageReference = &compute.ImageReferenceArgs{
				Publisher: pulumi.String(spec.Image.Publisher),
				Offer:     pulumi.String(spec.Image.Offer),
				Sku:       pulumi.String(spec.Image.Sku),
				Version:   pulumi.String(imageVersion),
			}
		}
	}

	// Configure OS disk
	osDiskName := metadata.Name + "-osdisk"
	storageProfile.OsDisk = &compute.OSDiskArgs{
		Name:         pulumi.String(osDiskName),
		CreateOption: pulumi.String("FromImage"),
		Caching:      compute.CachingTypesReadWrite,
		ManagedDisk: &compute.ManagedDiskParametersArgs{
			StorageAccountType: pulumi.String("Premium_LRS"),
		},
		DeleteOption: compute.DiskDeleteOptionTypesDelete,
	}

	vmArgs.StorageProfile = storageProfile

	// Configure availability zone
	if spec.AvailabilityZone != "" {
		vmArgs.Zones = pulumi.StringArray{pulumi.String(spec.AvailabilityZone)}
	}

	// Configure identity
	if spec.EnableSystemAssignedIdentity {
		vmArgs.Identity = &compute.VirtualMachineIdentityArgs{
			Type: compute.ResourceIdentityTypeSystemAssigned,
		}
	}

	// Configure spot instance
	if spec.IsSpotInstance {
		vmArgs.Priority = pulumi.String("Spot")
		vmArgs.EvictionPolicy = compute.VirtualMachineEvictionPolicyTypesDeallocate
		if spec.SpotMaxPrice > 0 || spec.SpotMaxPrice == -1 {
			vmArgs.BillingProfile = &compute.BillingProfileArgs{
				MaxPrice: pulumi.Float64(spec.SpotMaxPrice),
			}
		}
	}

	// Configure boot diagnostics
	enableBootDiag := true
	if spec.EnableBootDiagnostics != nil {
		enableBootDiag = *spec.EnableBootDiagnostics
	}
	if enableBootDiag {
		vmArgs.DiagnosticsProfile = &compute.DiagnosticsProfileArgs{
			BootDiagnostics: &compute.BootDiagnosticsArgs{
				Enabled: pulumi.Bool(true),
			},
		}
	}

	// Configure tags
	if len(spec.Tags) > 0 {
		tags := pulumi.StringMap{}
		for k, v := range spec.Tags {
			tags[k] = pulumi.String(v)
		}
		vmArgs.Tags = tags
	}

	// Create the Virtual Machine
	vm, err := compute.NewVirtualMachine(ctx, metadata.Name, vmArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create virtual machine")
	}

	// Export outputs
	ctx.Export(OpVmId, vm.ID())
	ctx.Export(OpVmName, vm.Name)
	ctx.Export(OpNetworkInterfaceId, nic.ID())
	ctx.Export(OpPrivateIpAddress, nic.IpConfigurations.Index(pulumi.Int(0)).PrivateIPAddress())

	if publicIp != nil {
		ctx.Export(OpPublicIpAddress, publicIp.IpAddress)
	}

	if spec.AvailabilityZone != "" {
		ctx.Export(OpAvailabilityZone, pulumi.String(spec.AvailabilityZone))
	}

	ctx.Export(OpComputerName, pulumi.String(metadata.Name))

	// Export identity principal ID if system-assigned
	if spec.EnableSystemAssignedIdentity {
		ctx.Export(OpSystemAssignedIdentityPrincipalId, vm.Identity.PrincipalId())
	}

	return nil
}

func getAvailabilityZoneArray(zone string) pulumi.StringArray {
	if zone == "" {
		return nil
	}
	return pulumi.StringArray{pulumi.String(zone)}
}
