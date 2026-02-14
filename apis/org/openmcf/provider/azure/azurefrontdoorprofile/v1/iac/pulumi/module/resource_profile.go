package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createProfile(ctx *pulumi.Context, locals *Locals, azureProvider *azure.Provider) (*cdn.FrontdoorProfile, error) {
	spec := locals.AzureFrontDoorProfile.Spec

	profile, err := cdn.NewFrontdoorProfile(ctx,
		spec.Name,
		&cdn.FrontdoorProfileArgs{
			Name:                   pulumi.String(spec.Name),
			ResourceGroupName:      pulumi.String(locals.ResourceGroupName),
			SkuName:                pulumi.String(spec.GetSku()),
			ResponseTimeoutSeconds: pulumi.Int(int(spec.GetResponseTimeoutSeconds())),
			Tags:                   pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Front Door profile %s", spec.Name)
	}

	return profile, nil
}
