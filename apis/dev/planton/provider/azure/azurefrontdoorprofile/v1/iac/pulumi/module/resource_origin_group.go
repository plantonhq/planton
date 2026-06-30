package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurefrontdoorprofilev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurefrontdoorprofile/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createdOriginGroup holds the Pulumi resource and associated origins spec.
type createdOriginGroup struct {
	Resource *cdn.FrontdoorOriginGroup
	Spec     *azurefrontdoorprofilev1.AzureFrontDoorOriginGroup
}

func createOriginGroups(
	ctx *pulumi.Context,
	locals *Locals,
	azureProvider *azure.Provider,
	profile *cdn.FrontdoorProfile,
) (map[string]*createdOriginGroup, error) {
	spec := locals.AzureFrontDoorProfile.Spec
	result := make(map[string]*createdOriginGroup)

	for _, og := range spec.GetOriginGroups() {
		created, err := createOriginGroup(ctx, og, profile, azureProvider)
		if err != nil {
			return nil, err
		}
		result[og.Name] = created
	}

	return result, nil
}

func createOriginGroup(
	ctx *pulumi.Context,
	og *azurefrontdoorprofilev1.AzureFrontDoorOriginGroup,
	profile *cdn.FrontdoorProfile,
	azureProvider *azure.Provider,
) (*createdOriginGroup, error) {
	args := &cdn.FrontdoorOriginGroupArgs{
		Name:                   pulumi.String(og.Name),
		CdnFrontdoorProfileId:  profile.ID(),
		SessionAffinityEnabled: pulumi.Bool(og.GetSessionAffinityEnabled()),
		LoadBalancing:          buildLoadBalancing(og.GetLoadBalancing()),
	}

	// Health probe is optional -- only configure if provided.
	if og.GetHealthProbe() != nil {
		args.HealthProbe = buildHealthProbe(og.GetHealthProbe())
	}

	originGroup, err := cdn.NewFrontdoorOriginGroup(ctx,
		fmt.Sprintf("og-%s", og.Name),
		args,
		pulumi.Provider(azureProvider),
		pulumi.DependsOn([]pulumi.Resource{profile}))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Front Door origin group %s", og.Name)
	}

	return &createdOriginGroup{
		Resource: originGroup,
		Spec:     og,
	}, nil
}

func buildLoadBalancing(lb *azurefrontdoorprofilev1.AzureFrontDoorLoadBalancing) cdn.FrontdoorOriginGroupLoadBalancingArgs {
	args := cdn.FrontdoorOriginGroupLoadBalancingArgs{}

	if lb != nil {
		args.SampleSize = pulumi.IntPtr(int(lb.GetSampleSize()))
		args.SuccessfulSamplesRequired = pulumi.IntPtr(int(lb.GetSuccessfulSamplesRequired()))
		args.AdditionalLatencyInMilliseconds = pulumi.IntPtr(int(lb.GetAdditionalLatencyInMilliseconds()))
	}

	return args
}

func buildHealthProbe(hp *azurefrontdoorprofilev1.AzureFrontDoorHealthProbe) cdn.FrontdoorOriginGroupHealthProbePtrInput {
	return &cdn.FrontdoorOriginGroupHealthProbeArgs{
		Protocol:          pulumi.String(hp.GetProtocol()),
		Path:              pulumi.StringPtr(hp.GetPath()),
		RequestType:       pulumi.StringPtr(hp.GetRequestType()),
		IntervalInSeconds: pulumi.Int(int(hp.GetIntervalInSeconds())),
	}
}
