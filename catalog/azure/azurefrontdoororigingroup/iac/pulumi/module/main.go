package module

import (
	"github.com/pkg/errors"
	azurefrontdoororigingroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoororigingroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoororigingroupv1alpha1.AzureFrontDoorOriginGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorOriginGroup.Spec

	// Azure requires load-balancing settings on every origin group, so
	// the block is always sent -- with the spec's values when present,
	// otherwise Azure's own defaults (sample size 4, 3 successful samples
	// required, 50 ms additional latency), matching what an unset spec
	// block documents. IaC inputs never carry proto defaults, so each
	// field materializes its documented default here.
	loadBalancingArgs := &cdn.FrontdoorOriginGroupLoadBalancingArgs{
		SampleSize:                      pulumi.Int(4),
		SuccessfulSamplesRequired:       pulumi.Int(3),
		AdditionalLatencyInMilliseconds: pulumi.Int(50),
	}
	if spec.LoadBalancing != nil {
		if spec.LoadBalancing.SampleSize != nil {
			loadBalancingArgs.SampleSize = pulumi.Int(int(spec.LoadBalancing.GetSampleSize()))
		}
		if spec.LoadBalancing.SuccessfulSamplesRequired != nil {
			loadBalancingArgs.SuccessfulSamplesRequired = pulumi.Int(int(spec.LoadBalancing.GetSuccessfulSamplesRequired()))
		}
		if spec.LoadBalancing.AdditionalLatencyInMilliseconds != nil {
			loadBalancingArgs.AdditionalLatencyInMilliseconds = pulumi.Int(int(spec.LoadBalancing.GetAdditionalLatencyInMilliseconds()))
		}
	}

	// The origin group addresses its parent by the profile's full ARM id
	// -- the provider derives the resource group and profile name from
	// it. ARM does not support tags on origin groups.
	originGroupArgs := &cdn.FrontdoorOriginGroupArgs{
		Name:                  pulumi.String(spec.OriginGroupName),
		CdnFrontdoorProfileId: pulumi.String(locals.ProfileId),
		LoadBalancing:         loadBalancingArgs,
	}

	// The health probe is sent only when configured: Front Door treats
	// ABSENT probe settings as probing disabled (all origins assumed
	// healthy), so omitting the block is a real behavior, not a defaults
	// shortcut.
	if spec.HealthProbe != nil {
		requestType := healthProbeRequestTypeStrings[spec.HealthProbe.RequestType]
		if requestType == "" {
			requestType = "HEAD"
		}
		probePath := "/"
		if spec.HealthProbe.Path != nil {
			probePath = spec.HealthProbe.GetPath()
		}
		originGroupArgs.HealthProbe = &cdn.FrontdoorOriginGroupHealthProbeArgs{
			Protocol:          pulumi.String(healthProbeProtocolStrings[spec.HealthProbe.Protocol]),
			IntervalInSeconds: pulumi.Int(int(spec.HealthProbe.IntervalInSeconds)),
			RequestType:       pulumi.String(requestType),
			Path:              pulumi.String(probePath),
		}
	}

	// Sent only when explicitly disabled/changed: Azure's defaults are
	// session affinity on and a 10-minute traffic-restore ramp.
	if spec.SessionAffinityEnabled != nil {
		originGroupArgs.SessionAffinityEnabled = pulumi.Bool(spec.GetSessionAffinityEnabled())
	}
	if spec.RestoreTrafficTimeToHealedOrNewEndpointInMinutes != nil {
		originGroupArgs.RestoreTrafficTimeToHealedOrNewEndpointInMinutes = pulumi.Int(int(spec.GetRestoreTrafficTimeToHealedOrNewEndpointInMinutes()))
	}

	createdOriginGroup, err := cdn.NewFrontdoorOriginGroup(ctx,
		spec.OriginGroupName,
		originGroupArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door origin group %s", spec.OriginGroupName)
	}

	// Export outputs. origin_group_id is what AzureFrontDoorOrigin
	// (parent) and AzureFrontDoorRoute (destination) reference.
	ctx.Export(OpOriginGroupId, createdOriginGroup.ID())
	ctx.Export(OpOriginGroupName, createdOriginGroup.Name)

	return nil
}
