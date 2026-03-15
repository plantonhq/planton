package module

import (
	"fmt"

	"github.com/pkg/errors"
	cloudflarezerotrustaccessapplicationv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarezerotrustaccessapplication/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// application provisions the Access Application and its default policy.
func application(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.AccessApplication, error) {

	// --- Access Application --------------------------------------------------
	// Resolve zone_id from literal value or reference.
	zoneId := ""
	if locals.CloudflareZeroTrustAccessApplication.Spec.ZoneId != nil {
		zoneId = locals.CloudflareZeroTrustAccessApplication.Spec.ZoneId.GetValue()
	}
	if zoneId == "" {
		return nil, errors.New("zone_id is required")
	}

	appArgs := &cloudflare.AccessApplicationArgs{
		Name:   pulumi.String(locals.CloudflareZeroTrustAccessApplication.Spec.ApplicationName),
		ZoneId: pulumi.String(zoneId),
		Domain: pulumi.String(locals.CloudflareZeroTrustAccessApplication.Spec.Hostname),
		Type:   pulumi.StringPtr("self_hosted"),
	}

	if locals.CloudflareZeroTrustAccessApplication.Spec.SessionDurationMinutes > 0 {
		appArgs.SessionDuration = pulumi.StringPtr(
			fmt.Sprintf("%dm", locals.CloudflareZeroTrustAccessApplication.Spec.SessionDurationMinutes),
		)
	}

	createdAccessApplication, err := cloudflare.NewAccessApplication(
		ctx,
		"access_application",
		appArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create access application")
	}

	// --- Access Policy -------------------------------------------------------
	// Lookup zone to get account ID (required for AccessPolicy in v6)
	zone := cloudflare.LookupZoneOutput(ctx, cloudflare.LookupZoneOutputArgs{
		ZoneId: pulumi.String(zoneId),
	}, pulumi.Provider(cloudflareProvider))

	accountId := zone.Account().Id()

	var includeBlocks cloudflare.AccessPolicyIncludeArray
	// In v6, email and group are nested structures, not arrays
	// Each email/group needs to be a separate include block
	if len(locals.CloudflareZeroTrustAccessApplication.Spec.AllowedEmails) > 0 {
		for _, e := range locals.CloudflareZeroTrustAccessApplication.Spec.AllowedEmails {
			includeBlocks = append(includeBlocks, &cloudflare.AccessPolicyIncludeArgs{
				Email: &cloudflare.AccessPolicyIncludeEmailArgs{
					Email: pulumi.String(e),
				},
			})
		}
	}

	if len(locals.CloudflareZeroTrustAccessApplication.Spec.AllowedGoogleGroups) > 0 {
		for _, g := range locals.CloudflareZeroTrustAccessApplication.Spec.AllowedGoogleGroups {
			includeBlocks = append(includeBlocks, &cloudflare.AccessPolicyIncludeArgs{
				Group: &cloudflare.AccessPolicyIncludeGroupArgs{
					Id: pulumi.String(g),
				},
			})
		}
	}

	var requireBlocks cloudflare.AccessPolicyRequireArray
	if locals.CloudflareZeroTrustAccessApplication.Spec.RequireMfa {
		requireBlocks = append(requireBlocks, &cloudflare.AccessPolicyRequireArgs{
			AuthMethod: &cloudflare.AccessPolicyRequireAuthMethodArgs{
				AuthMethod: pulumi.String("mfa"),
			},
		})
	}

	decision := "allow"
	if locals.CloudflareZeroTrustAccessApplication.Spec.PolicyType ==
		cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustPolicyType_BLOCK {
		decision = "deny"
	}

	createdAccessPolicy, err := cloudflare.NewAccessPolicy(
		ctx,
		"access_policy",
		&cloudflare.AccessPolicyArgs{
			AccountId: accountId,
			Name:      pulumi.String("default-policy"),
			Decision:  pulumi.String(decision),
			Includes:  includeBlocks,
			Requires:  requireBlocks,
		},
		pulumi.Provider(cloudflareProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAccessApplication}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create access policy")
	}

	// --- Stack Outputs -------------------------------------------------------
	ctx.Export(OpApplicationId, createdAccessApplication.ID())
	ctx.Export(OpPublicHostname, pulumi.String(locals.CloudflareZeroTrustAccessApplication.Spec.Hostname))
	ctx.Export(OpPolicyId, createdAccessPolicy.ID())

	return createdAccessApplication, nil
}
