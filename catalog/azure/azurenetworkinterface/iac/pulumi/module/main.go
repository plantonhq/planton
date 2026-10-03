package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurenetworkinterfacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurenetworkinterface/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurenetworkinterfacev1alpha1.AzureNetworkInterfaceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureNetworkInterface.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - Name, region, and edge zone are the NIC's identity -- changing
	//   any of them replaces the NIC, detaching it from its VM.
	//   Everything else (configurations, DNS, acceleration, forwarding,
	//   associations, tags) updates in place.
	// - The MAC address is assigned when the NIC attaches to a running
	//   VM, not at creation.
	// - NSG and ASG memberships are separate ARM operations, realized
	//   below as association resources (Azure's own model) rather than
	//   inline NIC properties -- detaching is just removing the spec
	//   field.
	ipConfigurations := network.NetworkInterfaceIpConfigurationArray{}
	for _, config := range spec.IpConfigurations {
		// Unset enums apply Azure's defaults (Dynamic allocation, IPv4),
		// so an unspecified spec and Azure's default deploy identically
		// on both engines.
		allocation := "Dynamic"
		if config.PrivateIpAllocation == azurenetworkinterfacev1alpha1.AzureNetworkInterfacePrivateIpAllocation_STATIC {
			allocation = "Static"
		}
		version := "IPv4"
		if config.PrivateIpVersion == azurenetworkinterfacev1alpha1.AzureNetworkInterfacePrivateIpVersion_IPV6 {
			version = "IPv6"
		}

		configArgs := network.NetworkInterfaceIpConfigurationArgs{
			Name:                       pulumi.String(config.Name),
			PrivateIpAddressAllocation: pulumi.String(allocation),
			PrivateIpAddressVersion:    pulumi.String(version),
			Primary:                    pulumi.Bool(config.Primary),
		}
		if config.SubnetId.GetValue() != "" {
			configArgs.SubnetId = pulumi.String(config.SubnetId.GetValue())
		}
		if config.PrivateIpAddress != "" {
			configArgs.PrivateIpAddress = pulumi.String(config.PrivateIpAddress)
		}
		if config.PublicIpAddressId.GetValue() != "" {
			configArgs.PublicIpAddressId = pulumi.String(config.PublicIpAddressId.GetValue())
		}
		if config.GatewayLoadBalancerFrontendIpConfigurationId != "" {
			configArgs.GatewayLoadBalancerFrontendIpConfigurationId = pulumi.String(config.GatewayLoadBalancerFrontendIpConfigurationId)
		}
		ipConfigurations = append(ipConfigurations, configArgs)
	}

	networkInterfaceArgs := &network.NetworkInterfaceArgs{
		Name:                         pulumi.String(spec.Name),
		Location:                     pulumi.String(spec.Region),
		ResourceGroupName:            pulumi.String(locals.ResourceGroupName),
		IpConfigurations:             ipConfigurations,
		AcceleratedNetworkingEnabled: pulumi.Bool(spec.AcceleratedNetworkingEnabled),
		IpForwardingEnabled:          pulumi.Bool(spec.IpForwardingEnabled),
		Tags:                         pulumi.ToStringMap(locals.AzureTags),
	}

	if len(spec.DnsServers) > 0 {
		networkInterfaceArgs.DnsServers = pulumi.ToStringArray(spec.DnsServers)
	}
	if spec.InternalDnsNameLabel != "" {
		networkInterfaceArgs.InternalDnsNameLabel = pulumi.String(spec.InternalDnsNameLabel)
	}
	// NVA acceleration (preview; subscription must be enrolled). Nothing
	// is sent when unspecified -- the correct shape for every
	// non-appliance NIC.
	if locals.AuxiliaryMode != "" {
		networkInterfaceArgs.AuxiliaryMode = pulumi.String(locals.AuxiliaryMode)
	}
	if locals.AuxiliarySku != "" {
		networkInterfaceArgs.AuxiliarySku = pulumi.String(locals.AuxiliarySku)
	}
	if spec.EdgeZone != "" {
		networkInterfaceArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	createdNetworkInterface, err := network.NewNetworkInterface(ctx,
		spec.Name,
		networkInterfaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create network interface %s", spec.Name)
	}

	// Attach the NIC-level network security group. Its own ARM operation
	// (Azure's model), so filtering can change without touching the NIC.
	if locals.NetworkSecurityGroupId != "" {
		if _, err := network.NewNetworkInterfaceSecurityGroupAssociation(ctx,
			fmt.Sprintf("%s-nsg", spec.Name),
			&network.NetworkInterfaceSecurityGroupAssociationArgs{
				NetworkInterfaceId:     createdNetworkInterface.ID(),
				NetworkSecurityGroupId: pulumi.String(locals.NetworkSecurityGroupId),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate network security group with network interface %s", spec.Name)
		}
	}

	// Join the NIC to its application security groups so NSG rules can
	// target workload groups instead of IP ranges. ASG references resolve
	// to literal ARM IDs before the module runs, so GetValue() returns the
	// resolved id for both a literal and a valueFrom reference.
	for i, asg := range spec.ApplicationSecurityGroupIds {
		if _, err := network.NewNetworkInterfaceApplicationSecurityGroupAssociation(ctx,
			fmt.Sprintf("%s-asg-%d", spec.Name, i),
			&network.NetworkInterfaceApplicationSecurityGroupAssociationArgs{
				NetworkInterfaceId:         createdNetworkInterface.ID(),
				ApplicationSecurityGroupId: pulumi.String(asg.GetValue()),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate application security group %d with network interface %s", i, spec.Name)
		}
	}

	// Load-balancer and Application Gateway memberships: expressed from
	// the member side in Azure's model. Each membership is its own ARM
	// association, so joining/leaving never touches the NIC itself.
	for _, config := range spec.IpConfigurations {
		for i, poolId := range config.LoadBalancerBackendAddressPoolIds {
			if poolId.GetValue() == "" {
				continue
			}
			if _, err := network.NewNetworkInterfaceBackendAddressPoolAssociation(ctx,
				fmt.Sprintf("%s-%s-lb-pool-%d", spec.Name, config.Name, i),
				&network.NetworkInterfaceBackendAddressPoolAssociationArgs{
					NetworkInterfaceId:   createdNetworkInterface.ID(),
					IpConfigurationName:  pulumi.String(config.Name),
					BackendAddressPoolId: pulumi.String(poolId.GetValue()),
				},
				pulumi.Provider(azureProvider)); err != nil {
				return errors.Wrapf(err, "failed to associate load balancer backend pool %d with ip configuration %s", i, config.Name)
			}
		}
		// Single-target inbound NAT rules: the load balancer declares the
		// port forward, this association picks the receiving instance.
		for i, natRuleId := range config.LoadBalancerInboundNatRuleIds {
			if natRuleId.GetValue() == "" {
				continue
			}
			if _, err := network.NewNetworkInterfaceNatRuleAssociation(ctx,
				fmt.Sprintf("%s-%s-lb-nat-%d", spec.Name, config.Name, i),
				&network.NetworkInterfaceNatRuleAssociationArgs{
					NetworkInterfaceId:  createdNetworkInterface.ID(),
					IpConfigurationName: pulumi.String(config.Name),
					NatRuleId:           pulumi.String(natRuleId.GetValue()),
				},
				pulumi.Provider(azureProvider)); err != nil {
				return errors.Wrapf(err, "failed to associate inbound NAT rule %d with ip configuration %s", i, config.Name)
			}
		}
		for i, poolId := range config.ApplicationGatewayBackendAddressPoolIds {
			if poolId.GetValue() == "" {
				continue
			}
			if _, err := network.NewNetworkInterfaceApplicationGatewayBackendAddressPoolAssociation(ctx,
				fmt.Sprintf("%s-%s-agw-pool-%d", spec.Name, config.Name, i),
				&network.NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationArgs{
					NetworkInterfaceId:   createdNetworkInterface.ID(),
					IpConfigurationName:  pulumi.String(config.Name),
					BackendAddressPoolId: pulumi.String(poolId.GetValue()),
				},
				pulumi.Provider(azureProvider)); err != nil {
				return errors.Wrapf(err, "failed to associate application gateway backend pool %d with ip configuration %s", i, config.Name)
			}
		}
	}

	// Export outputs from the created resource.
	ctx.Export(OpNetworkInterfaceId, createdNetworkInterface.ID())
	ctx.Export(OpNetworkInterfaceName, createdNetworkInterface.Name)
	ctx.Export(OpPrivateIpAddress, createdNetworkInterface.PrivateIpAddress)
	ctx.Export(OpPrivateIpAddresses, createdNetworkInterface.PrivateIpAddresses)
	ctx.Export(OpMacAddress, createdNetworkInterface.MacAddress)
	ctx.Export(OpInternalDomainNameSuffix, createdNetworkInterface.InternalDomainNameSuffix)

	return nil
}
