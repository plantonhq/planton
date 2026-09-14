package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sesv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// featureStatus maps a bool onto SES's ENABLED/DISABLED vocabulary.
func featureStatus(enabled bool) pulumi.String {
	if enabled {
		return pulumi.String("ENABLED")
	}
	return pulumi.String("DISABLED")
}

// accountSettings manages the region's SES account object -- the
// suppression list and VDM posture -- and exports outputs.
//
// Lifecycle facts the render below depends on:
//   - each arm renders ONLY when its spec message is present (an
//     omitted arm leaves the account's current setting untouched --
//     that omission is meaningful and deliberate);
//   - suppression switched off (enabled: false) is a real posture: it
//     writes the EMPTY reason list, which turns account-level
//     auto-suppression OFF;
//   - destroy semantics DIFFER per arm: suppression PERSISTS after
//     destroy (the provider's delete is a no-op; the last-applied
//     reasons stay), while the VDM resource's delete resets VDM to
//     DISABLED;
//   - VDM's dashboard/guardian sub-toggles are presence-typed: unset
//     renders nothing (AWS keeps its default), set maps to the
//     ENABLED/DISABLED FeatureStatus strings.
func accountSettings(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	if spec.Suppression != nil {
		// Suppression that is switched off is written as the empty reason
		// list (SES's spelling of "no auto-suppression"); the API already
		// refuses a switched-on arm with no events.
		reasons := pulumi.StringArray{}
		if spec.Suppression.GetEnabled() {
			for _, reason := range spec.Suppression.Reasons {
				reasons = append(reasons, pulumi.String(reason))
			}
		}
		if _, err := sesv2.NewAccountSuppressionAttributes(ctx, "suppression-attributes",
			&sesv2.AccountSuppressionAttributesArgs{
				SuppressedReasons: reasons,
			}, pulumi.Provider(provider)); err != nil {
			return errors.Wrap(err, "put account suppression attributes")
		}
	}

	if spec.Vdm != nil {
		args := &sesv2.AccountVdmAttributesArgs{
			VdmEnabled: featureStatus(spec.Vdm.GetEnabled()),
		}
		if spec.Vdm.EngagementMetrics != nil {
			args.DashboardAttributes = &sesv2.AccountVdmAttributesDashboardAttributesArgs{
				EngagementMetrics: featureStatus(spec.Vdm.GetEngagementMetrics()),
			}
		}
		if spec.Vdm.OptimizedSharedDelivery != nil {
			args.GuardianAttributes = &sesv2.AccountVdmAttributesGuardianAttributesArgs{
				OptimizedSharedDelivery: featureStatus(spec.Vdm.GetOptimizedSharedDelivery()),
			}
		}
		if _, err := sesv2.NewAccountVdmAttributes(ctx, "vdm-attributes", args,
			pulumi.Provider(provider)); err != nil {
			return errors.Wrap(err, "put account vdm attributes")
		}
	}

	// The account id feeds the output regardless of which arms render
	// (both upstream resources are account-scoped).
	callerIdentity, err := aws.GetCallerIdentity(ctx, &aws.GetCallerIdentityArgs{}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "get caller identity")
	}
	ctx.Export(OpAccountId, pulumi.String(callerIdentity.AccountId))

	return nil
}
