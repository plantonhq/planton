package module

import (
	"fmt"

	"github.com/pkg/errors"
	cloudflarezerotrustaccessapplicationv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarezerotrustaccessapplication/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// application provisions the Access Policy and the Access Application that
// references it. In v5 a policy is a standalone, account-scoped object; the
// application links it through its policies list.
func application(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.ZeroTrustAccessApplication, error) {

	spec := locals.CloudflareZeroTrustAccessApplication.Spec

	// Resolve zone_id from literal value or reference.
	zoneId := ""
	if spec.ZoneId != nil {
		zoneId = spec.ZoneId.GetValue()
	}
	if zoneId == "" {
		return nil, errors.New("zone_id is required")
	}

	// Lookup zone to get account ID (required for account-scoped resources).
	zone := cloudflare.LookupZoneOutput(ctx, cloudflare.LookupZoneOutputArgs{
		ZoneId: pulumi.String(zoneId),
	}, pulumi.Provider(cloudflareProvider))
	accountId := zone.Account().Id()

	// --- Access Policy -------------------------------------------------------
	var includes cloudflare.ZeroTrustAccessPolicyIncludeArray
	for _, e := range spec.AllowedEmails {
		includes = append(includes, &cloudflare.ZeroTrustAccessPolicyIncludeArgs{
			Email: &cloudflare.ZeroTrustAccessPolicyIncludeEmailArgs{
				Email: pulumi.String(e),
			},
		})
	}
	for _, g := range spec.AllowedGoogleGroups {
		includes = append(includes, &cloudflare.ZeroTrustAccessPolicyIncludeArgs{
			Group: &cloudflare.ZeroTrustAccessPolicyIncludeGroupArgs{
				Id: pulumi.String(g),
			},
		})
	}

	var requires cloudflare.ZeroTrustAccessPolicyRequireArray
	if spec.RequireMfa {
		requires = append(requires, &cloudflare.ZeroTrustAccessPolicyRequireArgs{
			AuthMethod: &cloudflare.ZeroTrustAccessPolicyRequireAuthMethodArgs{
				AuthMethod: pulumi.String("mfa"),
			},
		})
	}

	decision := "allow"
	if spec.PolicyType == cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustPolicyType_BLOCK {
		decision = "deny"
	}

	createdAccessPolicy, err := cloudflare.NewZeroTrustAccessPolicy(
		ctx,
		"access_policy",
		&cloudflare.ZeroTrustAccessPolicyArgs{
			AccountId: accountId,
			Name:      pulumi.String("default-policy"),
			Decision:  pulumi.String(decision),
			Includes:  includes,
			Requires:  requires,
		},
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create access policy")
	}

	// --- Access Application --------------------------------------------------
	appArgs := &cloudflare.ZeroTrustAccessApplicationArgs{
		AccountId: accountId,
		Name:      pulumi.String(spec.ApplicationName),
		Domain:    pulumi.String(spec.Hostname),
		Type:      pulumi.StringPtr("self_hosted"),
		Policies: cloudflare.ZeroTrustAccessApplicationPolicyArray{
			&cloudflare.ZeroTrustAccessApplicationPolicyArgs{
				Id:         createdAccessPolicy.ID(),
				Precedence: pulumi.IntPtr(1),
			},
		},
	}

	if spec.SessionDurationMinutes > 0 {
		appArgs.SessionDuration = pulumi.StringPtr(fmt.Sprintf("%dm", spec.SessionDurationMinutes))
	}

	createdAccessApplication, err := cloudflare.NewZeroTrustAccessApplication(
		ctx,
		"access_application",
		appArgs,
		pulumi.Provider(cloudflareProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAccessPolicy}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create access application")
	}

	// --- Stack Outputs -------------------------------------------------------
	ctx.Export(OpApplicationId, createdAccessApplication.ID())
	ctx.Export(OpPublicHostname, pulumi.String(spec.Hostname))
	ctx.Export(OpPolicyId, createdAccessPolicy.ID())

	return createdAccessApplication, nil
}
