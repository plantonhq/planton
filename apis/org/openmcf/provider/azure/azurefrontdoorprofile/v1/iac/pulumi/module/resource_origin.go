package module

import (
	"fmt"

	azurefrontdoorprofilev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurefrontdoorprofile/v1"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createdOrigin holds the Pulumi resource for a created origin.
type createdOrigin struct {
	Resource *cdn.FrontdoorOrigin
}

// createAllOrigins iterates over all origin groups and creates origins for each.
// Returns a flat map keyed by "originGroupName/originName" for downstream lookups.
func createAllOrigins(
	ctx *pulumi.Context,
	azureProvider *azure.Provider,
	originGroups map[string]*createdOriginGroup,
) (map[string]*createdOrigin, error) {
	result := make(map[string]*createdOrigin)

	for ogName, og := range originGroups {
		for _, origin := range og.Spec.GetOrigins() {
			created, err := createOrigin(ctx, origin, ogName, og.Resource, azureProvider)
			if err != nil {
				return nil, err
			}
			key := fmt.Sprintf("%s/%s", ogName, origin.Name)
			result[key] = created
		}
	}

	return result, nil
}

func createOrigin(
	ctx *pulumi.Context,
	origin *azurefrontdoorprofilev1.AzureFrontDoorOrigin,
	originGroupName string,
	originGroup *cdn.FrontdoorOriginGroup,
	azureProvider *azure.Provider,
) (*createdOrigin, error) {
	args := &cdn.FrontdoorOriginArgs{
		Name:                        pulumi.String(origin.Name),
		CdnFrontdoorOriginGroupId:   originGroup.ID(),
		HostName:                    pulumi.String(origin.GetHostName()),
		CertificateNameCheckEnabled: pulumi.Bool(origin.GetCertificateNameCheckEnabled()),
		HttpPort:                    pulumi.IntPtr(int(origin.GetHttpPort())),
		HttpsPort:                   pulumi.IntPtr(int(origin.GetHttpsPort())),
		Priority:                    pulumi.IntPtr(int(origin.GetPriority())),
		Weight:                      pulumi.IntPtr(int(origin.GetWeight())),
		Enabled:                     pulumi.Bool(origin.GetEnabled()),
	}

	// Optional origin host header.
	if origin.OriginHostHeader != nil {
		args.OriginHostHeader = pulumi.StringPtr(origin.GetOriginHostHeader())
	}

	// Optional private link configuration.
	if origin.GetPrivateLink() != nil {
		pl := origin.GetPrivateLink()
		plArgs := cdn.FrontdoorOriginPrivateLinkArgs{
			Location:              pulumi.String(pl.GetLocation()),
			PrivateLinkTargetId:   pulumi.String(pl.GetPrivateLinkTargetId()),
			RequestMessage:        pulumi.StringPtr(pl.GetRequestMessage()),
		}
		if pl.TargetType != nil {
			plArgs.TargetType = pulumi.StringPtr(pl.GetTargetType())
		}
		args.PrivateLink = &plArgs
	}

	resource, err := cdn.NewFrontdoorOrigin(ctx,
		fmt.Sprintf("origin-%s-%s", originGroupName, origin.Name),
		args,
		pulumi.Provider(azureProvider),
		pulumi.DependsOn([]pulumi.Resource{originGroup}))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Front Door origin %s in group %s", origin.Name, originGroupName)
	}

	return &createdOrigin{Resource: resource}, nil
}
