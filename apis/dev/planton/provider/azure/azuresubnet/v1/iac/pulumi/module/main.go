package module

import (
	"github.com/pkg/errors"
	azuresubnetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azuresubnet/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azuresubnetv1.AzureSubnetStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureSubnet.Spec

	// Build subnet arguments
	subnetArgs := &network.SubnetArgs{
		Name:               pulumi.String(spec.Name),
		ResourceGroupName:  pulumi.String(locals.ResourceGroupName),
		VirtualNetworkName: pulumi.String(locals.VnetName),
		AddressPrefixes: pulumi.StringArray{
			pulumi.String(spec.AddressPrefix),
		},
	}

	// Set service endpoints if specified
	if len(spec.ServiceEndpoints) > 0 {
		endpoints := pulumi.StringArray{}
		for _, ep := range spec.ServiceEndpoints {
			endpoints = append(endpoints, pulumi.String(ep))
		}
		subnetArgs.ServiceEndpoints = endpoints
	}

	// Set delegation if specified
	if spec.Delegation != nil {
		delegationArgs := network.SubnetDelegationArgs{
			Name: pulumi.String(spec.Delegation.Name),
			ServiceDelegation: network.SubnetDelegationServiceDelegationArgs{
				Name: pulumi.String(spec.Delegation.ServiceName),
			},
		}

		// Set delegation actions if specified
		if len(spec.Delegation.Actions) > 0 {
			actions := pulumi.StringArray{}
			for _, action := range spec.Delegation.Actions {
				actions = append(actions, pulumi.String(action))
			}
			delegationArgs.ServiceDelegation = network.SubnetDelegationServiceDelegationArgs{
				Name:    pulumi.String(spec.Delegation.ServiceName),
				Actions: actions,
			}
		}

		subnetArgs.Delegations = network.SubnetDelegationArray{delegationArgs}
	}

	// Set private endpoint network policies (uses default from proto if not explicitly set)
	subnetArgs.PrivateEndpointNetworkPolicies = pulumi.String(spec.GetPrivateEndpointNetworkPolicies())

	// Set private link service network policies
	subnetArgs.PrivateLinkServiceNetworkPoliciesEnabled = pulumi.Bool(spec.GetPrivateLinkServiceNetworkPoliciesEnabled())

	// Create the subnet
	subnet, err := network.NewSubnet(ctx,
		spec.Name,
		subnetArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create subnet %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpSubnetId, subnet.ID())
	ctx.Export(OpSubnetName, subnet.Name)
	ctx.Export(OpAddressPrefix, pulumi.String(spec.AddressPrefix))

	return nil
}
