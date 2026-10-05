package module

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sfn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func stateMachine(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	// The ASL definition arrives as a protobuf Struct; the AWS API takes it
	// as a JSON document. ASL key casing (StartAt, States, ...) survives the
	// round trip because Struct preserves map keys verbatim.
	definitionMap := spec.Definition.AsMap()
	definitionJSON, err := json.Marshal(definitionMap)
	if err != nil {
		return errors.Wrap(err, "failed to serialize state machine definition to JSON")
	}

	// STANDARD when unset. Changing the type replaces the state machine
	// (AWS ForceNew), which the spec documents.
	smType := "STANDARD"
	if spec.Type != "" {
		smType = spec.Type
	}

	args := &sfn.StateMachineArgs{
		// The cloud name is metadata.name -- the same basis the Terraform
		// module uses, so both engines create the same physical identity.
		Name:       pulumi.StringPtr(locals.Target.Metadata.Name),
		Definition: pulumi.String(string(definitionJSON)),
		RoleArn:    pulumi.String(spec.RoleArn.GetValue()),
		Type:       pulumi.StringPtr(smType),
		// Publish an immutable version on create and on every configuration
		// change. The latest version's ARN is exported as an output so
		// consumers can pin executions to a snapshot instead of the mutable
		// state machine.
		Publish: pulumi.BoolPtr(spec.Publish),
		Tags:    pulumi.ToStringMap(locals.AwsTags),
	}

	// X-Ray tracing is a tri-state toggle; the role must be able to put
	// trace segments (xray:PutTraceSegments / PutTelemetryRecords). Unset
	// sends no block (AWS default: off); an explicit true or false sends
	// the block -- the explicit false is what turns tracing OFF on a
	// machine that had it on (block removal alone is suppressed by the
	// provider and reverts nothing).
	if spec.TracingEnabled != nil {
		args.TracingConfiguration = &sfn.StateMachineTracingConfigurationArgs{
			Enabled: pulumi.BoolPtr(*spec.TracingEnabled),
		}
	}

	// Execution-history logging. Rendered for ANY configured level,
	// including an explicit OFF -- the OFF block is the disable send that
	// turns logging off on a machine that had it on (an absent block is
	// suppressed by the provider and reverts nothing).
	if spec.Logging != nil && spec.Logging.Level != "" {
		logArgs := &sfn.StateMachineLoggingConfigurationArgs{
			Level:                pulumi.StringPtr(spec.Logging.Level),
			IncludeExecutionData: pulumi.BoolPtr(spec.Logging.IncludeExecutionData),
		}

		if spec.Logging.LogDestination.GetValue() != "" {
			logDest := spec.Logging.LogDestination.GetValue()
			// AWS requires the CloudWatch log group ARN to end with ":*".
			// Referenced log-group ARNs arrive without the suffix, so append
			// it -- users should never have to know about this quirk.
			if !strings.HasSuffix(logDest, ":*") {
				logDest = logDest + ":*"
			}
			logArgs.LogDestination = pulumi.StringPtr(logDest)
		}

		args.LoggingConfiguration = logArgs
	}

	// Customer-managed KMS encryption. AWS's other arm (AWS_OWNED_KEY) is the
	// no-block default, so the spec models exactly one honest shape: a block
	// with a key means customer-managed.
	if spec.Encryption != nil && spec.Encryption.KmsKeyId.GetValue() != "" {
		encArgs := &sfn.StateMachineEncryptionConfigurationArgs{
			Type:     pulumi.StringPtr("CUSTOMER_MANAGED_KMS_KEY"),
			KmsKeyId: pulumi.StringPtr(spec.Encryption.KmsKeyId.GetValue()),
		}
		// A zero reuse period means "let AWS default" (300s); AWS accepts
		// 60-900.
		if spec.Encryption.KmsDataKeyReusePeriodSeconds != 0 {
			encArgs.KmsDataKeyReusePeriodSeconds = pulumi.IntPtr(int(spec.Encryption.KmsDataKeyReusePeriodSeconds))
		}
		args.EncryptionConfiguration = encArgs
	}

	sm, err := sfn.NewStateMachine(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Step Functions state machine")
	}

	// Folded aliases: an alias's identity IS this state machine (one alias
	// set per machine), so aliases live here rather than as their own kind.
	// Each entry is keyed by its name -- adding, renaming, or removing one
	// alias never touches its siblings -- and routes 100% of traffic to the
	// version THIS deployment published (spec CEL guarantees publish: true
	// whenever aliases exist). Weighted canary routing between two specific
	// versions is an imperative deployment-shift operation and is
	// deliberately not modeled.
	aliasArns := pulumi.StringMap{}
	for _, aliasSpec := range spec.Aliases {
		aliasArgs := &sfn.AliasArgs{
			Name: pulumi.StringPtr(aliasSpec.Name),
			RoutingConfigurations: sfn.AliasRoutingConfigurationArray{
				&sfn.AliasRoutingConfigurationArgs{
					StateMachineVersionArn: sm.StateMachineVersionArn,
					Weight:                 pulumi.Int(100),
				},
			},
		}
		if aliasSpec.Description != "" {
			aliasArgs.Description = pulumi.StringPtr(aliasSpec.Description)
		}
		createdAlias, err := sfn.NewAlias(ctx,
			locals.Target.Metadata.Name+"-alias-"+aliasSpec.Name,
			aliasArgs, pulumi.Provider(provider), pulumi.Parent(sm))
		if err != nil {
			return errors.Wrapf(err, "failed to create alias %s", aliasSpec.Name)
		}
		aliasArns[aliasSpec.Name] = createdAlias.Arn
	}

	// Export outputs matching AwsStepFunctionOutputs.
	ctx.Export(OpStateMachineArn, sm.Arn)
	ctx.Export(OpStateMachineName, sm.Name)
	ctx.Export(OpStateMachineVersionArn, sm.StateMachineVersionArn)
	ctx.Export(OpRevisionId, sm.RevisionId)
	ctx.Export(OpStatus, sm.Status)
	ctx.Export(OpCreationDate, sm.CreationDate)
	ctx.Export(OpAliasArns, aliasArns)

	return nil
}
