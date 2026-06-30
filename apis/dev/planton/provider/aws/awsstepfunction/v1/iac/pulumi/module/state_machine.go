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

	// -------------------------------------------------------------------
	// Serialize the ASL definition from protobuf Struct to JSON.
	// -------------------------------------------------------------------

	definitionMap := spec.Definition.AsMap()
	definitionJSON, err := json.Marshal(definitionMap)
	if err != nil {
		return errors.Wrap(err, "failed to serialize state machine definition to JSON")
	}

	// -------------------------------------------------------------------
	// State machine type (default to STANDARD when not specified).
	// -------------------------------------------------------------------

	smType := "STANDARD"
	if spec.Type != "" {
		smType = spec.Type
	}

	args := &sfn.StateMachineArgs{
		Name:       pulumi.StringPtr(locals.Target.Metadata.Name),
		Definition: pulumi.String(string(definitionJSON)),
		RoleArn:    pulumi.String(spec.RoleArn.GetValue()),
		Type:       pulumi.StringPtr(smType),
		Tags:       pulumi.ToStringMap(locals.AwsTags),
	}

	// -------------------------------------------------------------------
	// Description
	// -------------------------------------------------------------------

	// Note: The SFN Pulumi resource does not have a top-level Description
	// input property. Description is read-only (computed by AWS). We skip
	// setting it here; it is surfaced in the AWS Console automatically.

	// -------------------------------------------------------------------
	// Tracing configuration
	// -------------------------------------------------------------------

	if spec.TracingEnabled {
		args.TracingConfiguration = &sfn.StateMachineTracingConfigurationArgs{
			Enabled: pulumi.BoolPtr(true),
		}
	}

	// -------------------------------------------------------------------
	// Logging configuration
	// -------------------------------------------------------------------

	if spec.Logging != nil && spec.Logging.Level != "" && spec.Logging.Level != "OFF" {
		logArgs := &sfn.StateMachineLoggingConfigurationArgs{
			Level:                pulumi.StringPtr(spec.Logging.Level),
			IncludeExecutionData: pulumi.BoolPtr(spec.Logging.IncludeExecutionData),
		}

		if spec.Logging.LogDestination.GetValue() != "" {
			logDest := spec.Logging.LogDestination.GetValue()
			// AWS requires the log group ARN to end with ":*".
			if !strings.HasSuffix(logDest, ":*") {
				logDest = logDest + ":*"
			}
			logArgs.LogDestination = pulumi.StringPtr(logDest)
		}

		args.LoggingConfiguration = logArgs
	}

	// -------------------------------------------------------------------
	// Encryption configuration
	// -------------------------------------------------------------------

	if spec.Encryption != nil && spec.Encryption.KmsKeyId.GetValue() != "" {
		encArgs := &sfn.StateMachineEncryptionConfigurationArgs{
			Type:     pulumi.StringPtr("CUSTOMER_MANAGED_KMS_KEY"),
			KmsKeyId: pulumi.StringPtr(spec.Encryption.KmsKeyId.GetValue()),
		}
		if spec.Encryption.KmsDataKeyReusePeriodSeconds != 0 {
			encArgs.KmsDataKeyReusePeriodSeconds = pulumi.IntPtr(int(spec.Encryption.KmsDataKeyReusePeriodSeconds))
		}
		args.EncryptionConfiguration = encArgs
	}

	// -------------------------------------------------------------------
	// Create state machine
	// -------------------------------------------------------------------

	sm, err := sfn.NewStateMachine(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Step Functions state machine")
	}

	// Export outputs matching AwsStepFunctionStackOutputs.
	ctx.Export(OpStateMachineArn, sm.Arn)
	ctx.Export(OpStateMachineName, sm.Name)

	return nil
}
