package module

import (
	"github.com/pkg/errors"
	azureexpressroutecircuitv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureexpressroutecircuit/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureexpressroutecircuitv1alpha1.AzureExpressRouteCircuitIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureExpressRouteCircuit.Spec

	circuitArgs := &network.ExpressRouteCircuitArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku: &network.ExpressRouteCircuitSkuArgs{
			Tier:   pulumi.String(skuTierWireValue(spec.SkuTier)),
			Family: pulumi.String(skuFamilyWireValue(spec.SkuFamily)),
		},
		AllowClassicOperations: pulumi.Bool(spec.AllowClassicOperations),
		RateLimitingEnabled:    pulumi.Bool(spec.RateLimitingEnabled),
		Tags:                   pulumi.ToStringMap(locals.AzureTags),
	}

	// Exactly one provisioning mode (spec-validated): the
	// service-provider trio, or the ExpressRoute Direct pair. ARM treats
	// the two property sets as mutually exclusive shapes of the same
	// create call.
	if spec.ServiceProviderName != "" {
		circuitArgs.ServiceProviderName = pulumi.String(spec.ServiceProviderName)
		circuitArgs.PeeringLocation = pulumi.String(spec.PeeringLocation)
		circuitArgs.BandwidthInMbps = pulumi.Int(int(spec.BandwidthInMbps))
	} else {
		// The port reference is pre-resolved by the platform middleware;
		// GetValue() is the resolved literal ARM id (an
		// AzureExpressRoutePort's express_route_port_id output, or a
		// literal for ports managed outside Planton).
		circuitArgs.ExpressRoutePortId = pulumi.String(spec.ExpressRoutePortId.GetValue())
		circuitArgs.BandwidthInGbps = pulumi.Float64(spec.BandwidthInGbps)
	}

	// The key this circuit REDEEMS (capacity someone else owns) -- not
	// the keys it issues. ARM never returns it on reads; the provider
	// writes it in a follow-up call after the circuit exists.
	if spec.AuthorizationKey.GetValue() != "" {
		circuitArgs.AuthorizationKey = pulumi.String(spec.AuthorizationKey.GetValue())
	}

	createdCircuit, err := network.NewExpressRouteCircuit(ctx,
		spec.Name,
		circuitArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create express route circuit %s", spec.Name)
	}

	// The composed authorizations: standalone ARM children of the
	// circuit, deployed after it. ARM GENERATES each key; the name-keyed
	// map surfaces them (as secrets) so a far-side gateway in another
	// subscription can redeem one.
	authorizationKeys := pulumi.Map{}
	for _, authorization := range spec.Authorizations {
		createdAuthorization, err := network.NewExpressRouteCircuitAuthorization(ctx,
			spec.Name+"-"+authorization.Name,
			&network.ExpressRouteCircuitAuthorizationArgs{
				Name:                    pulumi.String(authorization.Name),
				ExpressRouteCircuitName: createdCircuit.Name,
				ResourceGroupName:       pulumi.String(locals.ResourceGroupName),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdCircuit))
		if err != nil {
			return errors.Wrapf(err, "failed to create authorization %s", authorization.Name)
		}
		authorizationKeys[authorization.Name] = createdAuthorization.AuthorizationKey
	}

	ctx.Export(OpExpressRouteCircuitId, createdCircuit.ID())
	ctx.Export(OpExpressRouteCircuitName, createdCircuit.Name)
	// The service key is the circuit's provisioning credential -- secret
	// in both engines (the SDK already types it secret; ToSecret keeps
	// the contract explicit).
	ctx.Export(OpServiceKey, pulumi.ToSecret(createdCircuit.ServiceKey))
	ctx.Export(OpServiceProviderProvisioningState, createdCircuit.ServiceProviderProvisioningState)
	ctx.Export(OpAuthorizationKeys, pulumi.ToSecret(authorizationKeys))

	return nil
}
