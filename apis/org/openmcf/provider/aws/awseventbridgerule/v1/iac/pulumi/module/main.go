package module

import (
	"github.com/pkg/errors"
	awseventbridgerulev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awseventbridgerule/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates EventBridge rule and target creation, then exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awseventbridgerulev1.AwsEventBridgeRuleStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
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

	createdRule, err := cloudwatch.NewEventRule(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create EventBridge rule")
	}

	// Export rule-level outputs
	ctx.Export(OpRuleArn, createdRule.Arn)
	ctx.Export(OpRuleName, createdRule.Name)

	return createdRule, nil
}
