package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/athena"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func workgroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	// -------------------------------------------------------------------
	// Build configuration block
	// -------------------------------------------------------------------

	config := &athena.WorkgroupConfigurationArgs{}

	// Cost controls
	if spec.BytesScannedCutoffPerQuery > 0 {
		config.BytesScannedCutoffPerQuery = pulumi.IntPtr(int(spec.BytesScannedCutoffPerQuery))
	}

	// Governance booleans
	if spec.EnforceWorkgroupConfiguration != nil {
		config.EnforceWorkgroupConfiguration = pulumi.BoolPtr(*spec.EnforceWorkgroupConfiguration)
	}

	if spec.PublishCloudwatchMetricsEnabled != nil {
		config.PublishCloudwatchMetricsEnabled = pulumi.BoolPtr(*spec.PublishCloudwatchMetricsEnabled)
	}

	if spec.RequesterPaysEnabled {
		config.RequesterPaysEnabled = pulumi.BoolPtr(true)
	}

	// DEFERRED: enable_minimum_encryption_configuration is defined in the spec
	// but is not available in the pinned Pulumi AWS SDK v7.3.0. The spec
	// retains the field for forward compatibility; it will be wired when the
	// SDK dependency is upgraded.

	// Engine version
	if spec.SelectedEngineVersion != "" {
		config.EngineVersion = &athena.WorkgroupConfigurationEngineVersionArgs{
			SelectedEngineVersion: pulumi.StringPtr(spec.SelectedEngineVersion),
		}
	}

	// Execution role (for Spark workgroups)
	if spec.ExecutionRole.GetValue() != "" {
		config.ExecutionRole = pulumi.StringPtr(spec.ExecutionRole.GetValue())
	}

	// -------------------------------------------------------------------
	// Result configuration
	// -------------------------------------------------------------------

	if spec.ResultConfiguration != nil {
		rc := spec.ResultConfiguration

		resultConfig := &athena.WorkgroupConfigurationResultConfigurationArgs{}

		if rc.OutputLocation != "" {
			resultConfig.OutputLocation = pulumi.StringPtr(rc.OutputLocation)
		}

		if rc.ExpectedBucketOwner != "" {
			resultConfig.ExpectedBucketOwner = pulumi.StringPtr(rc.ExpectedBucketOwner)
		}

		// Encryption
		if rc.EncryptionOption != "" {
			encConfig := &athena.WorkgroupConfigurationResultConfigurationEncryptionConfigurationArgs{
				EncryptionOption: pulumi.StringPtr(rc.EncryptionOption),
			}
			if rc.KmsKeyArn.GetValue() != "" {
				encConfig.KmsKeyArn = pulumi.StringPtr(rc.KmsKeyArn.GetValue())
			}
			resultConfig.EncryptionConfiguration = encConfig
		}

		// ACL
		if rc.S3AclOption != "" {
			resultConfig.AclConfiguration = &athena.WorkgroupConfigurationResultConfigurationAclConfigurationArgs{
				S3AclOption: pulumi.String(rc.S3AclOption),
			}
		}

		config.ResultConfiguration = resultConfig
	}

	// -------------------------------------------------------------------
	// Create workgroup
	// -------------------------------------------------------------------

	wg, err := athena.NewWorkgroup(ctx, locals.Target.Metadata.Name, &athena.WorkgroupArgs{
		Name:          pulumi.StringPtr(locals.WorkgroupName),
		Configuration: config,
		ForceDestroy:  pulumi.BoolPtr(spec.ForceDestroy),
		Tags:          pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Athena workgroup")
	}

	// -------------------------------------------------------------------
	// Export outputs matching AwsAthenaWorkgroupStackOutputs
	// -------------------------------------------------------------------

	ctx.Export(OpWorkgroupArn, wg.Arn)
	ctx.Export(OpWorkgroupName, wg.Name)

	// effective_engine_version is nested: configuration.engine_version.effective_engine_version
	ctx.Export(OpEffectiveEngineVersion, wg.Configuration.ApplyT(func(c *athena.WorkgroupConfiguration) string {
		if c != nil && c.EngineVersion != nil && c.EngineVersion.EffectiveEngineVersion != nil {
			return *c.EngineVersion.EffectiveEngineVersion
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}
