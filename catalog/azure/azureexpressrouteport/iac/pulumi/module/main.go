package module

import (
	"github.com/pkg/errors"
	azureexpressrouteportv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureexpressrouteport/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureexpressrouteportv1alpha1.AzureExpressRoutePortIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureExpressRoutePort.Spec

	// Create the ExpressRoute Port -- the physical port pair on a
	// Microsoft edge router at the peering location. The port bills its
	// full monthly rate FROM CREATION, whether or not any cross-connect
	// exists, and some subscriptions need Microsoft enrollment for
	// ExpressRoute Direct before ARM accepts this create.
	portArgs := &network.ExpressRoutePortArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		PeeringLocation:   pulumi.String(spec.PeeringLocation),
		BandwidthInGbps:   pulumi.Int(int(spec.BandwidthInGbps)),
		Encapsulation:     pulumi.String(encapsulationWireValue(spec.Encapsulation)),
		BillingType:       pulumi.String(billingTypeWireValue(spec.BillingType)),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// The port's managed identity -- what MACsec uses to read the CAK/CKN
	// secrets from Key Vault (the spec's contracts guarantee a
	// user-assigned identity is present whenever MACsec keys are set).
	if spec.Identity != nil {
		identityIds := make([]string, 0, len(spec.Identity.IdentityIds))
		for _, identityId := range spec.Identity.IdentityIds {
			identityIds = append(identityIds, identityId.GetValue())
		}
		portArgs.Identity = &network.ExpressRoutePortIdentityArgs{
			Type:        pulumi.String(identityTypeWireValue(spec.Identity.Type)),
			IdentityIds: pulumi.ToStringArray(identityIds),
		}
	}

	// ARM always creates the link pair with the port; these blocks only
	// MANIPULATE the existing links (admin state, MACsec). The provider
	// applies link configuration in a second call after the port exists.
	if spec.Link1 != nil {
		portArgs.Link1 = expressRoutePortLink1Args(spec.Link1)
	}
	if spec.Link2 != nil {
		portArgs.Link2 = expressRoutePortLink2Args(spec.Link2)
	}

	createdPort, err := network.NewExpressRoutePort(ctx,
		spec.Name,
		portArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create express route port %s", spec.Name)
	}

	// The composed authorizations: standalone ARM children of the port,
	// deployed after it (the provider serializes them against the port
	// with its own lock -- ARM allows one port mutation at a time). ARM
	// GENERATES each key; the name-keyed map surfaces them (as secrets)
	// so a circuit in another subscription can redeem one.
	authorizationKeys := pulumi.Map{}
	for _, authorization := range spec.Authorizations {
		createdAuthorization, err := network.NewExpressRoutePortAuthorization(ctx,
			spec.Name+"-"+authorization.Name,
			&network.ExpressRoutePortAuthorizationArgs{
				Name:                 pulumi.String(authorization.Name),
				ExpressRoutePortName: createdPort.Name,
				ResourceGroupName:    pulumi.String(locals.ResourceGroupName),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdPort))
		if err != nil {
			return errors.Wrapf(err, "failed to create authorization %s", authorization.Name)
		}
		authorizationKeys[authorization.Name] = createdAuthorization.AuthorizationKey
	}

	ctx.Export(OpExpressRoutePortId, createdPort.ID())
	ctx.Export(OpExpressRoutePortName, createdPort.Name)
	ctx.Export(OpGuid, createdPort.Guid)
	ctx.Export(OpEthertype, createdPort.Ethertype)
	ctx.Export(OpMtu, createdPort.Mtu)
	// Empty when the identity type does not include SYSTEM_ASSIGNED --
	// mirrors the Terraform module's try(..., "").
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdPort.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	// The per-link facility facts (router, interface, patch panel, rack)
	// are the letter-of-authorization data handed to the colocation
	// facility to order the physical cross-connects. Elem() dereferences
	// to the zero value ("") if ARM has not populated a fact yet.
	ctx.Export(OpLink1Id, createdPort.Link1.Id().Elem())
	ctx.Export(OpLink1RouterName, createdPort.Link1.RouterName().Elem())
	ctx.Export(OpLink1InterfaceName, createdPort.Link1.InterfaceName().Elem())
	ctx.Export(OpLink1PatchPanelId, createdPort.Link1.PatchPanelId().Elem())
	ctx.Export(OpLink1RackId, createdPort.Link1.RackId().Elem())
	ctx.Export(OpLink1ConnectorType, createdPort.Link1.ConnectorType().Elem())
	ctx.Export(OpLink2Id, createdPort.Link2.Id().Elem())
	ctx.Export(OpLink2RouterName, createdPort.Link2.RouterName().Elem())
	ctx.Export(OpLink2InterfaceName, createdPort.Link2.InterfaceName().Elem())
	ctx.Export(OpLink2PatchPanelId, createdPort.Link2.PatchPanelId().Elem())
	ctx.Export(OpLink2RackId, createdPort.Link2.RackId().Elem())
	ctx.Export(OpLink2ConnectorType, createdPort.Link2.ConnectorType().Elem())

	ctx.Export(OpAuthorizationKeys, pulumi.ToSecret(authorizationKeys))

	return nil
}

// expressRoutePortLink1Args builds the SDK's link1 block from the spec's
// link message. The SDK types link1 and link2 separately (they map to
// the fixed physical pair), so the two builders differ only in type.
func expressRoutePortLink1Args(link *azureexpressrouteportv1alpha1.AzureExpressRoutePortLink) *network.ExpressRoutePortLink1Args {
	args := &network.ExpressRoutePortLink1Args{
		AdminEnabled:     pulumi.Bool(link.AdminEnabled),
		MacsecCipher:     pulumi.String(macsecCipherWireValue(link.MacsecCipher)),
		MacsecSciEnabled: pulumi.Bool(link.MacsecSciEnabled),
	}
	// The two keys travel together (spec-validated); ARM rejects a
	// one-sided MACsec configuration.
	if link.MacsecCknKeyvaultSecretId != "" {
		args.MacsecCknKeyvaultSecretId = pulumi.String(link.MacsecCknKeyvaultSecretId)
		args.MacsecCakKeyvaultSecretId = pulumi.String(link.MacsecCakKeyvaultSecretId)
	}
	return args
}

// expressRoutePortLink2Args mirrors expressRoutePortLink1Args for the
// second link of the pair.
func expressRoutePortLink2Args(link *azureexpressrouteportv1alpha1.AzureExpressRoutePortLink) *network.ExpressRoutePortLink2Args {
	args := &network.ExpressRoutePortLink2Args{
		AdminEnabled:     pulumi.Bool(link.AdminEnabled),
		MacsecCipher:     pulumi.String(macsecCipherWireValue(link.MacsecCipher)),
		MacsecSciEnabled: pulumi.Bool(link.MacsecSciEnabled),
	}
	if link.MacsecCknKeyvaultSecretId != "" {
		args.MacsecCknKeyvaultSecretId = pulumi.String(link.MacsecCknKeyvaultSecretId)
		args.MacsecCakKeyvaultSecretId = pulumi.String(link.MacsecCakKeyvaultSecretId)
	}
	return args
}
