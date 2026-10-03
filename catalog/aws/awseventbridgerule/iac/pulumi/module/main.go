package module

import (
	"github.com/pkg/errors"
	awseventbridgerulev1alpha1 "github.com/plantonhq/planton/catalog/aws/awseventbridgerule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates EventBridge rule and target creation, then exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awseventbridgerulev1alpha1.AwsEventBridgeRuleIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdRule, err := rule(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "event bridge rule")
	}

	if err := targets(ctx, locals, createdRule, provider); err != nil {
		return errors.Wrap(err, "event bridge targets")
	}

	return nil
}

// rule creates the EventBridge rule and exports rule-level outputs.
func rule(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*cloudwatch.EventRule, error) {
	spec := locals.Spec

	args := &cloudwatch.EventRuleArgs{
		Name: pulumi.StringPtr(locals.Target.Metadata.Name),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Description
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Event bus (defaults to "default" when empty)
	if spec.EventBusName.GetValue() != "" {
		args.EventBusName = pulumi.StringPtr(spec.EventBusName.GetValue())
	}

	// Event pattern — serialize google.protobuf.Struct to JSON
	if spec.EventPattern != nil {
		patternJSON, err := serializeStruct(spec.EventPattern)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize event_pattern")
		}
		args.EventPattern = pulumi.StringPtr(patternJSON)
	}

	// Schedule expression
	if spec.ScheduleExpression != "" {
		args.ScheduleExpression = pulumi.StringPtr(spec.ScheduleExpression)
	}

	// State (defaults to ENABLED when not set)
	if spec.State != "" {
		args.State = pulumi.StringPtr(spec.State)
	}

	// Rule-level invocation role (per-target role_arn takes precedence for
	// its own target).
	if spec.RoleArn.GetValue() != "" {
		args.RoleArn = pulumi.StringPtr(spec.RoleArn.GetValue())
	}

	// AWS refuses to delete a rule that still has targets unless forced. The
	// module removes its own targets first, so force only matters when an
	// out-of-band consumer attached extra targets to this rule.
	if spec.ForceDestroy {
		args.ForceDestroy = pulumi.BoolPtr(true)
	}

	createdRule, err := cloudwatch.NewEventRule(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create EventBridge rule")
	}

	// Export rule-level outputs
	ctx.Export(OpRuleArn, createdRule.Arn)
	ctx.Export(OpRuleName, createdRule.Name)

	return createdRule, nil
}
