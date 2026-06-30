package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurenetworksecuritygroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurenetworksecuritygroupv1.AzureNetworkSecurityGroupStackInput) error {
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

	spec := locals.AzureNetworkSecurityGroup.Spec

	// Create the Network Security Group (shell -- rules are created as separate resources)
	nsg, err := network.NewNetworkSecurityGroup(ctx,
		spec.Name,
		&network.NetworkSecurityGroupArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Network Security Group %s", spec.Name)
	}

	// Create a separate NetworkSecurityRule for each rule in the spec.
	// Using separate resources (not inline) provides per-rule lifecycle management,
	// better error messages, and explicit resource naming in state.
	for i, rule := range spec.SecurityRules {
		ruleArgs := &network.NetworkSecurityRuleArgs{
			Name:                     pulumi.String(rule.Name),
			ResourceGroupName:        pulumi.String(locals.ResourceGroupName),
			NetworkSecurityGroupName: nsg.Name,
			Priority:                 pulumi.Int(int(rule.Priority)),
			Direction:                pulumi.String(rule.Direction),
			Access:                   pulumi.String(rule.Access),
			Protocol:                 pulumi.String(rule.Protocol),
			SourcePortRange:          pulumi.String(rule.GetSourcePortRange()),
			DestinationPortRange:     pulumi.String(rule.DestinationPortRange),
		}

		// Set description if provided
		if rule.Description != "" {
			ruleArgs.Description = pulumi.String(rule.Description)
		}

		// Address prefix handling: plural takes precedence over singular.
		// If plural is non-empty, use it. Otherwise use singular (default "*").
		if len(rule.SourceAddressPrefixes) > 0 {
			prefixes := pulumi.StringArray{}
			for _, p := range rule.SourceAddressPrefixes {
				prefixes = append(prefixes, pulumi.String(p))
			}
			ruleArgs.SourceAddressPrefixes = prefixes
		} else {
			ruleArgs.SourceAddressPrefix = pulumi.String(rule.GetSourceAddressPrefix())
		}

		if len(rule.DestinationAddressPrefixes) > 0 {
			prefixes := pulumi.StringArray{}
			for _, p := range rule.DestinationAddressPrefixes {
				prefixes = append(prefixes, pulumi.String(p))
			}
			ruleArgs.DestinationAddressPrefixes = prefixes
		} else {
			ruleArgs.DestinationAddressPrefix = pulumi.String(rule.GetDestinationAddressPrefix())
		}

		_, err := network.NewNetworkSecurityRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, rule.Name),
			ruleArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{nsg}))
		if err != nil {
			return errors.Wrapf(err, "failed to create security rule %d (%s)", i, rule.Name)
		}
	}

	// Export stack outputs
	ctx.Export(OpNsgId, nsg.ID())
	ctx.Export(OpNsgName, nsg.Name)

	return nil
}
