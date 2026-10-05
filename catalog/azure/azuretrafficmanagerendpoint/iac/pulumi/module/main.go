package module

import (
	"github.com/pkg/errors"
	azuretrafficmanagerendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuretrafficmanagerendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one Traffic Manager endpoint -- a destination the
// referenced profile steers traffic to.
//
// The endpoint type is whichever variant block the spec carries
// (validation guarantees exactly one), so exactly one branch below
// runs. Shared fields (weight, priority, enabled, geo/subnet claims,
// probe headers) feed whichever resource is created; which of them
// MATTER depends on the profile's routing method (Azure evaluates them
// there).
//
// Endpoints carry no ARM tags on any engine. Priority is sent only
// when set -- unset lets Azure assign the next free value in creation
// order (the service owns that default). Weight defaults to 1 and is
// always sent, so both engines send identical wire shapes.
func Resources(ctx *pulumi.Context, iacInput *azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureTrafficManagerEndpoint.Spec

	weight := 1
	if spec.Weight != nil {
		weight = int(*spec.Weight)
	}

	enabled := true
	if spec.Enabled != nil {
		enabled = *spec.Enabled
	}

	// Exactly one endpoint resource materializes; its ARM id and name
	// flatten onto the same outputs regardless of type.
	var endpointId pulumi.StringOutput
	var endpointName pulumi.StringOutput

	switch {
	case spec.Azure != nil:
		args := &network.TrafficManagerAzureEndpointArgs{
			Name:      pulumi.String(spec.Name),
			ProfileId: pulumi.String(locals.ProfileId),
			// The target reference is pre-resolved by the platform
			// middleware; Azure reads the target's address (and its
			// region, for Performance routing) from the resource itself,
			// so this type has no endpoint-location argument.
			TargetResourceId:   pulumi.String(spec.Azure.TargetResourceId.GetValue()),
			Weight:             pulumi.Int(weight),
			Enabled:            pulumi.Bool(enabled),
			AlwaysServeEnabled: pulumi.Bool(spec.Azure.AlwaysServeEnabled != nil && *spec.Azure.AlwaysServeEnabled),
		}
		if spec.Priority != nil {
			args.Priority = pulumi.Int(int(*spec.Priority))
		}
		if len(spec.GeoMappings) > 0 {
			args.GeoMappings = pulumi.ToStringArray(spec.GeoMappings)
		}
		args.Subnets = azureEndpointSubnets(spec.Subnets)
		args.CustomHeaders = azureEndpointHeaders(spec.CustomHeaders)

		created, err := network.NewTrafficManagerAzureEndpoint(ctx, spec.Name, args, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create azure endpoint %s", spec.Name)
		}
		endpointId = created.ID().ToStringOutput()
		endpointName = created.Name

	case spec.External != nil:
		args := &network.TrafficManagerExternalEndpointArgs{
			Name:      pulumi.String(spec.Name),
			ProfileId: pulumi.String(locals.ProfileId),
			// The target is a StringValueOrRef; the platform resolves
			// valueFrom references before the module runs, so GetValue()
			// is the resolved literal DNS name or IP.
			Target:             pulumi.String(spec.External.Target.GetValue()),
			Weight:             pulumi.Int(weight),
			Enabled:            pulumi.Bool(enabled),
			AlwaysServeEnabled: pulumi.Bool(spec.External.AlwaysServeEnabled != nil && *spec.External.AlwaysServeEnabled),
		}
		// REQUIRED by the service when the profile routes by Performance
		// (external targets carry no discoverable region); ignored
		// otherwise -- enforced apply-time.
		if spec.External.EndpointLocation != "" {
			args.EndpointLocation = pulumi.String(spec.External.EndpointLocation)
		}
		if spec.Priority != nil {
			args.Priority = pulumi.Int(int(*spec.Priority))
		}
		if len(spec.GeoMappings) > 0 {
			args.GeoMappings = pulumi.ToStringArray(spec.GeoMappings)
		}
		args.Subnets = externalEndpointSubnets(spec.Subnets)
		args.CustomHeaders = externalEndpointHeaders(spec.CustomHeaders)

		created, err := network.NewTrafficManagerExternalEndpoint(ctx, spec.Name, args, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create external endpoint %s", spec.Name)
		}
		endpointId = created.ID().ToStringOutput()
		endpointName = created.Name

	case spec.Nested != nil:
		// This type carries no always-serve switch (the provider exposes
		// none -- child health IS the point of nesting).
		args := &network.TrafficManagerNestedEndpointArgs{
			Name:                  pulumi.String(spec.Name),
			ProfileId:             pulumi.String(locals.ProfileId),
			TargetResourceId:      pulumi.String(spec.Nested.TargetProfileId.GetValue()),
			MinimumChildEndpoints: pulumi.Int(int(spec.Nested.GetMinimumChildEndpoints())),
			Weight:                pulumi.Int(weight),
			Enabled:               pulumi.Bool(enabled),
		}
		// The IPv4/IPv6 health floors pass through only when positive,
		// mirroring the provider's own send-when-set behavior.
		if spec.Nested.MinimumRequiredChildEndpointsIpv4 != nil && *spec.Nested.MinimumRequiredChildEndpointsIpv4 > 0 {
			args.MinimumRequiredChildEndpointsIpv4 = pulumi.Int(int(*spec.Nested.MinimumRequiredChildEndpointsIpv4))
		}
		if spec.Nested.MinimumRequiredChildEndpointsIpv6 != nil && *spec.Nested.MinimumRequiredChildEndpointsIpv6 > 0 {
			args.MinimumRequiredChildEndpointsIpv6 = pulumi.Int(int(*spec.Nested.MinimumRequiredChildEndpointsIpv6))
		}
		if spec.Nested.EndpointLocation != "" {
			args.EndpointLocation = pulumi.String(spec.Nested.EndpointLocation)
		}
		if spec.Priority != nil {
			args.Priority = pulumi.Int(int(*spec.Priority))
		}
		if len(spec.GeoMappings) > 0 {
			args.GeoMappings = pulumi.ToStringArray(spec.GeoMappings)
		}
		args.Subnets = nestedEndpointSubnets(spec.Subnets)
		args.CustomHeaders = nestedEndpointHeaders(spec.CustomHeaders)

		created, err := network.NewTrafficManagerNestedEndpoint(ctx, spec.Name, args, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create nested endpoint %s", spec.Name)
		}
		endpointId = created.ID().ToStringOutput()
		endpointName = created.Name

	default:
		// Unreachable: spec validation requires exactly one variant.
		return errors.New("no endpoint variant present in spec")
	}

	ctx.Export(OpEndpointId, endpointId)
	ctx.Export(OpEndpointName, endpointName)

	return nil
}

// The three endpoint resources carry structurally identical subnet and
// custom-header blocks, but the SDK generates a distinct Go type per
// resource -- hence one small builder per variant instead of one shared
// one.

func azureEndpointSubnets(subnets []*azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointSubnet) network.TrafficManagerAzureEndpointSubnetArrayInput {
	if len(subnets) == 0 {
		return nil
	}
	out := network.TrafficManagerAzureEndpointSubnetArray{}
	for _, subnet := range subnets {
		entry := &network.TrafficManagerAzureEndpointSubnetArgs{First: pulumi.String(subnet.First)}
		if subnet.Last != "" {
			entry.Last = pulumi.String(subnet.Last)
		}
		if subnet.Scope != nil {
			entry.Scope = pulumi.Int(int(*subnet.Scope))
		}
		out = append(out, entry)
	}
	return out
}

func externalEndpointSubnets(subnets []*azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointSubnet) network.TrafficManagerExternalEndpointSubnetArrayInput {
	if len(subnets) == 0 {
		return nil
	}
	out := network.TrafficManagerExternalEndpointSubnetArray{}
	for _, subnet := range subnets {
		entry := &network.TrafficManagerExternalEndpointSubnetArgs{First: pulumi.String(subnet.First)}
		if subnet.Last != "" {
			entry.Last = pulumi.String(subnet.Last)
		}
		if subnet.Scope != nil {
			entry.Scope = pulumi.Int(int(*subnet.Scope))
		}
		out = append(out, entry)
	}
	return out
}

func nestedEndpointSubnets(subnets []*azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointSubnet) network.TrafficManagerNestedEndpointSubnetArrayInput {
	if len(subnets) == 0 {
		return nil
	}
	out := network.TrafficManagerNestedEndpointSubnetArray{}
	for _, subnet := range subnets {
		entry := &network.TrafficManagerNestedEndpointSubnetArgs{First: pulumi.String(subnet.First)}
		if subnet.Last != "" {
			entry.Last = pulumi.String(subnet.Last)
		}
		if subnet.Scope != nil {
			entry.Scope = pulumi.Int(int(*subnet.Scope))
		}
		out = append(out, entry)
	}
	return out
}

func azureEndpointHeaders(headers []*azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointCustomHeader) network.TrafficManagerAzureEndpointCustomHeaderArrayInput {
	if len(headers) == 0 {
		return nil
	}
	out := network.TrafficManagerAzureEndpointCustomHeaderArray{}
	for _, header := range headers {
		out = append(out, &network.TrafficManagerAzureEndpointCustomHeaderArgs{
			Name:  pulumi.String(header.Name),
			Value: pulumi.String(header.Value),
		})
	}
	return out
}

func externalEndpointHeaders(headers []*azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointCustomHeader) network.TrafficManagerExternalEndpointCustomHeaderArrayInput {
	if len(headers) == 0 {
		return nil
	}
	out := network.TrafficManagerExternalEndpointCustomHeaderArray{}
	for _, header := range headers {
		out = append(out, &network.TrafficManagerExternalEndpointCustomHeaderArgs{
			Name:  pulumi.String(header.Name),
			Value: pulumi.String(header.Value),
		})
	}
	return out
}

func nestedEndpointHeaders(headers []*azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointCustomHeader) network.TrafficManagerNestedEndpointCustomHeaderArrayInput {
	if len(headers) == 0 {
		return nil
	}
	out := network.TrafficManagerNestedEndpointCustomHeaderArray{}
	for _, header := range headers {
		out = append(out, &network.TrafficManagerNestedEndpointCustomHeaderArgs{
			Name:  pulumi.String(header.Name),
			Value: pulumi.String(header.Value),
		})
	}
	return out
}
