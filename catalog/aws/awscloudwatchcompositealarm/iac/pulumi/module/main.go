package module

import (
	"github.com/pkg/errors"
	awscloudwatchcompositealarmv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchcompositealarm/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	fkv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one CloudWatch composite alarm. The composite has no
// metrics or thresholds of its own — it re-evaluates its boolean alarm_rule
// whenever any referenced alarm changes state, which is why the module maps
// the spec 1:1 with no evaluation plumbing.
func Resources(ctx *pulumi.Context, iacInput *awscloudwatchcompositealarmv1alpha1.AwsCloudwatchCompositeAlarmIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AwsCloudwatchCompositeAlarm.Spec

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web
	// identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	args := &cloudwatch.CompositeAlarmArgs{
		// The alarm's cloud name is the resource's metadata.name — the same
		// basis the Terraform module uses, and the identity other composite
		// alarms address in their own rule expressions.
		AlarmName: pulumi.String(locals.AwsCloudwatchCompositeAlarm.Metadata.Name),
		// The boolean expression over other alarms' states. Referenced alarms
		// are addressed by NAME (compose from AwsCloudwatchAlarm's exported
		// alarm_name output); the rule text passes through verbatim.
		AlarmRule: pulumi.String(spec.AlarmRule),
		Tags:      pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.AlarmDescription != "" {
		args.AlarmDescription = pulumi.StringPtr(spec.AlarmDescription)
	}

	// Actions enabled is presence-aware: unset lets AWS default to true; an
	// explicit false is a real choice. Note this flag is create-time-only on
	// composite alarms (the provider marks it ForceNew) — flipping it replaces
	// the alarm, unlike on metric alarms.
	if spec.ActionsEnabled != nil {
		args.ActionsEnabled = pulumi.BoolPtr(spec.GetActionsEnabled())
	}

	if len(spec.AlarmActions) > 0 {
		args.AlarmActions = buildActionArns(spec.AlarmActions)
	}
	if len(spec.OkActions) > 0 {
		args.OkActions = buildActionArns(spec.OkActions)
	}
	if len(spec.InsufficientDataActions) > 0 {
		args.InsufficientDataActions = buildActionArns(spec.InsufficientDataActions)
	}

	// The actions suppressor silences actions (never state transitions) while
	// the designated suppressor alarm is in ALARM — the maintenance-window
	// mechanism. The suppressor is addressed by alarm NAME per the CloudWatch
	// API contract.
	if spec.ActionsSuppressor != nil {
		args.ActionsSuppressor = &cloudwatch.CompositeAlarmActionsSuppressorArgs{
			Alarm:           pulumi.String(spec.ActionsSuppressor.Alarm.GetValue()),
			WaitPeriod:      pulumi.Int(int(spec.ActionsSuppressor.WaitPeriod)),
			ExtensionPeriod: pulumi.Int(int(spec.ActionsSuppressor.ExtensionPeriod)),
		}
	}

	createdCompositeAlarm, err := cloudwatch.NewCompositeAlarm(ctx,
		locals.AwsCloudwatchCompositeAlarm.Metadata.Name,
		args,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create composite alarm")
	}

	ctx.Export(OpAlarmArn, createdCompositeAlarm.Arn)
	ctx.Export(OpAlarmName, createdCompositeAlarm.AlarmName)

	return nil
}

// buildActionArns converts a slice of StringValueOrRef into a
// pulumi.StringArray suitable for alarm/ok/insufficient-data action ARN
// fields. References arrive pre-resolved to literal values by the
// orchestrator.
func buildActionArns(actions []*fkv1.StringValueOrRef) pulumi.StringArray {
	result := pulumi.StringArray{}
	for _, action := range actions {
		if action.GetValue() != "" {
			result = append(result, pulumi.String(action.GetValue()))
		}
	}
	return result
}
