package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurefrontdoorprofilev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurefrontdoorprofile/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createdEndpoint holds the Pulumi resource and its hostname output.
type createdEndpoint struct {
	Resource *cdn.FrontdoorEndpoint
}

func createEndpoints(
	ctx *pulumi.Context,
	locals *Locals,
	azureProvider *azure.Provider,
	profile *cdn.FrontdoorProfile,
) (map[string]*createdEndpoint, error) {
	spec := locals.AzureFrontDoorProfile.Spec
	result := make(map[string]*createdEndpoint)

	for _, ep := range spec.GetEndpoints() {
		created, err := createEndpoint(ctx, ep, profile, azureProvider)
		if err != nil {
			return nil, err
		}
		result[ep.Name] = created
	}

	return result, nil
}

func createEndpoint(
	ctx *pulumi.Context,
	ep *azurefrontdoorprofilev1.AzureFrontDoorEndpoint,
	profile *cdn.FrontdoorProfile,
	azureProvider *azure.Provider,
) (*createdEndpoint, error) {
	args := &cdn.FrontdoorEndpointArgs{
		Name:                  pulumi.String(ep.Name),
		CdnFrontdoorProfileId: profile.ID(),
		Enabled:               pulumi.Bool(ep.GetEnabled()),
		Tags:                  nil,
	}

	endpoint, err := cdn.NewFrontdoorEndpoint(ctx,
		fmt.Sprintf("endpoint-%s", ep.Name),
		args,
		pulumi.Provider(azureProvider),
		pulumi.DependsOn([]pulumi.Resource{profile}))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Front Door endpoint %s", ep.Name)
	}

	return &createdEndpoint{Resource: endpoint}, nil
}
