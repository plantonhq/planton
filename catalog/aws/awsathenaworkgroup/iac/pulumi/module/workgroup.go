package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/athena"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// workgroup creates the Athena workgroup and exports its stack outputs.
//
// One provider resource carries the whole surface; everything interesting
// lives inside the single Configuration input. Two provider behaviors shape
// the wiring below:
//
//   - Optional nested blocks are presence-driven: the provider treats an
//     absent block as the feature's disabled state, so each optional spec
//     message maps to a nil-guarded args struct -- never an always-sent
//     struct with zero values, which would pin AWS defaults and create
//     phantom diffs.
//   - Where AWS requires an explicit `enabled` flag INSIDE a block
//     (managed results, the three logging arms), the block's presence in the
//     spec asserts it: enabled is hardcoded true because "present but
//     disabled" and "absent" are the same AWS state.
func workgroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	// -------------------------------------------------------------------
	// Configuration block
	// -------------------------------------------------------------------

	config := &athena.WorkgroupConfigurationArgs{}

	// 0 means "no limit" in the spec; the provider expresses no-limit by
	// omitting the argument.
	if spec.BytesScannedCutoffPerQuery > 0 {
		config.BytesScannedCutoffPerQuery = pulumi.IntPtr(int(spec.BytesScannedCutoffPerQuery))
	}

	// Tri-state dials (proto `optional bool`, spec default true): unset falls
	// through to the provider default (also true), so an omitted dial and an
	// explicit true deploy identically while explicit false is representable.
	if spec.EnforceWorkgroupConfiguration != nil {
		config.EnforceWorkgroupConfiguration = pulumi.BoolPtr(*spec.EnforceWorkgroupConfiguration)
	}
	if spec.PublishCloudwatchMetricsEnabled != nil {
		config.PublishCloudwatchMetricsEnabled = pulumi.BoolPtr(*spec.PublishCloudwatchMetricsEnabled)
	}

	// Always sent (matching the Terraform module): the provider drops an
	// explicit false before it reaches AWS, so this is purely a send-style
	// convergence.
	config.RequesterPaysEnabled = pulumi.BoolPtr(spec.RequesterPaysEnabled)

	// Compliance guardrail: query results are written with at least SSE_S3
	// even when individual queries specify no encryption. Always sent: the
	// provider attribute is Optional+Computed, so omitting the value would
	// keep whatever AWS last reported -- an explicit false is the ONLY way
	// to turn the guardrail back off once enabled (the Terraform module
	// sends it unconditionally for the same reason).
	config.EnableMinimumEncryptionConfiguration = pulumi.BoolPtr(spec.EnableMinimumEncryptionConfiguration)

	if spec.SelectedEngineVersion != "" {
		config.EngineVersion = &athena.WorkgroupConfigurationEngineVersionArgs{
			SelectedEngineVersion: pulumi.StringPtr(spec.SelectedEngineVersion),
		}
	}

	// Assumed for Spark workloads and Identity Center-enabled workgroups;
	// plain SQL workgroups leave it unset.
	if spec.ExecutionRole.GetValue() != "" {
		config.ExecutionRole = pulumi.StringPtr(spec.ExecutionRole.GetValue())
	}

	// KMS encryption for Spark notebook content and session data (SQL query
	// results are covered by the result blocks below, not this).
	if spec.CustomerContentEncryptionKmsKey.GetValue() != "" {
		config.CustomerContentEncryptionConfiguration = &athena.WorkgroupConfigurationCustomerContentEncryptionConfigurationArgs{
			KmsKey: pulumi.StringPtr(spec.CustomerContentEncryptionKmsKey.GetValue()),
		}
	}

	// -------------------------------------------------------------------
	// Result storage: customer-managed S3 XOR AWS-managed storage
	// (the spec CEL mirrors the provider's own plan-time exclusivity rule).
	// -------------------------------------------------------------------

	if rc := spec.ResultConfiguration; rc != nil {
		resultConfig := &athena.WorkgroupConfigurationResultConfigurationArgs{}

		if rc.OutputLocation != "" {
			resultConfig.OutputLocation = pulumi.StringPtr(rc.OutputLocation)
		}
		if rc.ExpectedBucketOwner != "" {
			resultConfig.ExpectedBucketOwner = pulumi.StringPtr(rc.ExpectedBucketOwner)
		}

		if rc.EncryptionOption != "" {
			encConfig := &athena.WorkgroupConfigurationResultConfigurationEncryptionConfigurationArgs{
				EncryptionOption: pulumi.StringPtr(rc.EncryptionOption),
			}
			if rc.KmsKeyArn.GetValue() != "" {
				encConfig.KmsKeyArn = pulumi.StringPtr(rc.KmsKeyArn.GetValue())
			}
			resultConfig.EncryptionConfiguration = encConfig
		}

		if rc.S3AclOption != "" {
			resultConfig.AclConfiguration = &athena.WorkgroupConfigurationResultConfigurationAclConfigurationArgs{
				S3AclOption: pulumi.String(rc.S3AclOption),
			}
		}

		config.ResultConfiguration = resultConfig
	}

	// AWS-managed result storage: no bucket to own, 24-hour retention,
	// results retrievable through Athena APIs only. Each of these blocks
	// carries its own enable switch (on by default once the block is
	// declared, filled by the platform before this runs); the switch is
	// forwarded as the provider's own `enabled`, which treats a disabled
	// block and an absent one as the same state.
	if mqr := spec.ManagedQueryResults; mqr != nil {
		managedArgs := &athena.WorkgroupConfigurationManagedQueryResultsConfigurationArgs{
			Enabled: pulumi.BoolPtr(mqr.GetEnabled()),
		}
		if mqr.KmsKey.GetValue() != "" {
			managedArgs.EncryptionConfiguration = &athena.WorkgroupConfigurationManagedQueryResultsConfigurationEncryptionConfigurationArgs{
				KmsKey: pulumi.StringPtr(mqr.KmsKey.GetValue()),
			}
		}
		config.ManagedQueryResultsConfiguration = managedArgs
	}

	// -------------------------------------------------------------------
	// Identity integration
	// -------------------------------------------------------------------

	// IAM Identity Center trusted identity propagation. Create-time settings:
	// changing them replaces the workgroup.
	if ic := spec.IdentityCenter; ic != nil {
		icArgs := &athena.WorkgroupConfigurationIdentityCenterConfigurationArgs{
			EnableIdentityCenter: pulumi.BoolPtr(ic.EnableIdentityCenter),
		}
		if ic.IdentityCenterInstanceArn != "" {
			icArgs.IdentityCenterInstanceArn = pulumi.StringPtr(ic.IdentityCenterInstanceArn)
		}
		config.IdentityCenterConfiguration = icArgs
	}

	// S3 Access Grants credentials for the result location, scoped to the
	// propagated user identity.
	if ag := spec.S3AccessGrants; ag != nil {
		config.QueryResultsS3AccessGrantsConfiguration = &athena.WorkgroupConfigurationQueryResultsS3AccessGrantsConfigurationArgs{
			EnableS3AccessGrants:  pulumi.Bool(ag.EnableS3AccessGrants),
			AuthenticationType:    pulumi.String(ag.AuthenticationType),
			CreateUserLevelPrefix: pulumi.BoolPtr(ag.CreateUserLevelPrefix),
		}
	}

	// -------------------------------------------------------------------
	// Monitoring: the three log-delivery arms are independent and
	// combinable; each is enabled by presence with its required `enabled`
	// flag asserted.
	// -------------------------------------------------------------------

	if mon := spec.Monitoring; mon != nil {
		monArgs := &athena.WorkgroupConfigurationMonitoringConfigurationArgs{}

		if cw := mon.CloudWatchLogging; cw != nil {
			cwArgs := &athena.WorkgroupConfigurationMonitoringConfigurationCloudWatchLoggingConfigurationArgs{
				Enabled: pulumi.Bool(cw.GetEnabled()),
			}
			if cw.LogGroup != "" {
				cwArgs.LogGroup = pulumi.StringPtr(cw.LogGroup)
			}
			if cw.LogStreamNamePrefix != "" {
				cwArgs.LogStreamNamePrefix = pulumi.StringPtr(cw.LogStreamNamePrefix)
			}
			// AWS models log selection as worker type -> log streams
			// (e.g. SPARK_DRIVER -> [STDOUT, STDERR]); the spec's repeated
			// entries map one-to-one onto the provider's logTypes list.
			if len(cw.LogTypes) > 0 {
				logTypes := athena.WorkgroupConfigurationMonitoringConfigurationCloudWatchLoggingConfigurationLogTypeArray{}
				for _, entry := range cw.LogTypes {
					logTypes = append(logTypes, &athena.WorkgroupConfigurationMonitoringConfigurationCloudWatchLoggingConfigurationLogTypeArgs{
						Key:    pulumi.String(entry.Key),
						Values: pulumi.ToStringArray(entry.Values),
					})
				}
				cwArgs.LogTypes = logTypes
			}
			monArgs.CloudWatchLoggingConfiguration = cwArgs
		}

		if ml := mon.ManagedLogging; ml != nil {
			mlArgs := &athena.WorkgroupConfigurationMonitoringConfigurationManagedLoggingConfigurationArgs{
				Enabled: pulumi.Bool(ml.GetEnabled()),
			}
			if ml.KmsKey.GetValue() != "" {
				mlArgs.KmsKey = pulumi.StringPtr(ml.KmsKey.GetValue())
			}
			monArgs.ManagedLoggingConfiguration = mlArgs
		}

		if s3l := mon.S3Logging; s3l != nil {
			s3Args := &athena.WorkgroupConfigurationMonitoringConfigurationS3LoggingConfigurationArgs{
				Enabled: pulumi.Bool(s3l.GetEnabled()),
			}
			if s3l.LogLocation != "" {
				s3Args.LogLocation = pulumi.StringPtr(s3l.LogLocation)
			}
			if s3l.KmsKey.GetValue() != "" {
				s3Args.KmsKey = pulumi.StringPtr(s3l.KmsKey.GetValue())
			}
			monArgs.S3LoggingConfiguration = s3Args
		}

		config.MonitoringConfiguration = monArgs
	}

	// -------------------------------------------------------------------
	// Workgroup resource
	// -------------------------------------------------------------------

	args := &athena.WorkgroupArgs{
		// The cloud name is set explicitly from metadata.name (never Pulumi
		// auto-naming) so both engines create the identical workgroup.
		Name:          pulumi.StringPtr(locals.WorkgroupName),
		Configuration: config,
		// DISABLED rejects new query submissions but keeps configuration,
		// history, and saved queries -- the pause switch, not a teardown.
		State: pulumi.StringPtr(locals.WorkgroupState),
		// Allows destroy to proceed even when the workgroup still holds
		// named queries or prepared statements.
		ForceDestroy: pulumi.BoolPtr(spec.ForceDestroy),
		Tags:         pulumi.ToStringMap(locals.AwsTags),
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	wg, err := athena.NewWorkgroup(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Athena workgroup")
	}

	// -------------------------------------------------------------------
	// Stack outputs (contract: AwsAthenaWorkgroupStackOutputs)
	// -------------------------------------------------------------------

	ctx.Export(OpWorkgroupArn, wg.Arn)
	ctx.Export(OpWorkgroupName, wg.Name)

	// effective_engine_version lives at configuration.engine_version.
	// The chained Elem accessors are ApplyT-free: a nil block yields the
	// string zero value, matching the Terraform module's try(..., "").
	ctx.Export(OpEffectiveEngineVersion, wg.Configuration.EngineVersion().EffectiveEngineVersion().Elem())

	return nil
}
