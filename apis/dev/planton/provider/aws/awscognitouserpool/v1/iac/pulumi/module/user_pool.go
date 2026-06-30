package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func userPool(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*cognito.UserPool, error) {
	spec := locals.Spec

	args := &cognito.UserPoolArgs{
		Name: pulumi.String(locals.Target.Metadata.Name),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// ---------------------------------------------------------------------------
	// Identity model
	// ---------------------------------------------------------------------------

	if len(spec.UsernameAttributes) > 0 {
		args.UsernameAttributes = pulumi.ToStringArray(spec.UsernameAttributes)
	}

	if len(spec.AliasAttributes) > 0 {
		args.AliasAttributes = pulumi.ToStringArray(spec.AliasAttributes)
	}

	args.UsernameConfiguration = &cognito.UserPoolUsernameConfigurationArgs{
		CaseSensitive: pulumi.Bool(spec.UsernameCaseSensitive),
	}

	// ---------------------------------------------------------------------------
	// Password policy
	// ---------------------------------------------------------------------------

	if spec.PasswordPolicy != nil {
		pp := spec.PasswordPolicy
		args.PasswordPolicy = &cognito.UserPoolPasswordPolicyArgs{
			RequireLowercase: pulumi.BoolPtr(pp.RequireLowercase),
			RequireUppercase: pulumi.BoolPtr(pp.RequireUppercase),
			RequireNumbers:   pulumi.BoolPtr(pp.RequireNumbers),
			RequireSymbols:   pulumi.BoolPtr(pp.RequireSymbols),
		}
		if pp.MinimumLength > 0 {
			args.PasswordPolicy.(*cognito.UserPoolPasswordPolicyArgs).MinimumLength = pulumi.IntPtr(int(pp.MinimumLength))
		}
		if pp.TemporaryPasswordValidityDays > 0 {
			args.PasswordPolicy.(*cognito.UserPoolPasswordPolicyArgs).TemporaryPasswordValidityDays = pulumi.IntPtr(int(pp.TemporaryPasswordValidityDays))
		}
	}

	// ---------------------------------------------------------------------------
	// MFA
	// ---------------------------------------------------------------------------

	if spec.MfaConfiguration != "" {
		args.MfaConfiguration = pulumi.StringPtr(spec.MfaConfiguration)
	}

	if spec.SoftwareTokenMfaEnabled {
		args.SoftwareTokenMfaConfiguration = &cognito.UserPoolSoftwareTokenMfaConfigurationArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	// ---------------------------------------------------------------------------
	// Auto-verified attributes
	// ---------------------------------------------------------------------------

	if len(spec.AutoVerifiedAttributes) > 0 {
		args.AutoVerifiedAttributes = pulumi.ToStringArray(spec.AutoVerifiedAttributes)
	}

	// ---------------------------------------------------------------------------
	// Account recovery
	// ---------------------------------------------------------------------------

	if len(spec.AccountRecoveryMechanisms) > 0 {
		var mechanisms cognito.UserPoolAccountRecoverySettingRecoveryMechanismArray
		for _, m := range spec.AccountRecoveryMechanisms {
			mechanisms = append(mechanisms, &cognito.UserPoolAccountRecoverySettingRecoveryMechanismArgs{
				Name:     pulumi.String(m.Name),
				Priority: pulumi.Int(int(m.Priority)),
			})
		}
		args.AccountRecoverySetting = &cognito.UserPoolAccountRecoverySettingArgs{
			RecoveryMechanisms: mechanisms,
		}
	}

	// ---------------------------------------------------------------------------
	// Email configuration
	// ---------------------------------------------------------------------------

	if spec.EmailConfiguration != nil {
		ec := spec.EmailConfiguration
		emailArgs := &cognito.UserPoolEmailConfigurationArgs{}

		if ec.EmailSendingAccount != "" {
			emailArgs.EmailSendingAccount = pulumi.StringPtr(ec.EmailSendingAccount)
		}
		if ec.SourceArn.GetValue() != "" {
			emailArgs.SourceArn = pulumi.StringPtr(ec.SourceArn.GetValue())
		}
		if ec.FromEmailAddress != "" {
			emailArgs.FromEmailAddress = pulumi.StringPtr(ec.FromEmailAddress)
		}
		if ec.ReplyToEmailAddress != "" {
			emailArgs.ReplyToEmailAddress = pulumi.StringPtr(ec.ReplyToEmailAddress)
		}
		if ec.ConfigurationSet != "" {
			emailArgs.ConfigurationSet = pulumi.StringPtr(ec.ConfigurationSet)
		}

		args.EmailConfiguration = emailArgs
	}

	// ---------------------------------------------------------------------------
	// Admin and deletion protection
	// ---------------------------------------------------------------------------

	if spec.AllowAdminCreateUserOnly {
		args.AdminCreateUserConfig = &cognito.UserPoolAdminCreateUserConfigArgs{
			AllowAdminCreateUserOnly: pulumi.BoolPtr(true),
		}
	}

	if spec.DeletionProtection {
		args.DeletionProtection = pulumi.StringPtr("ACTIVE")
	}

	// ---------------------------------------------------------------------------
	// Custom attributes (schema)
	// ---------------------------------------------------------------------------

	if len(spec.CustomAttributes) > 0 {
		var schemas cognito.UserPoolSchemaArray
		for _, attr := range spec.CustomAttributes {
			schemaArgs := &cognito.UserPoolSchemaArgs{
				Name:                   pulumi.String(attr.Name),
				AttributeDataType:      pulumi.String(attr.AttributeDataType),
				Mutable:                pulumi.BoolPtr(attr.Mutable),
				DeveloperOnlyAttribute: pulumi.BoolPtr(false),
			}

			if attr.Required {
				schemaArgs.Required = pulumi.BoolPtr(true)
			}

			if attr.AttributeDataType == "String" && (attr.StringMinLength != "" || attr.StringMaxLength != "") {
				strConstraints := &cognito.UserPoolSchemaStringAttributeConstraintsArgs{}
				if attr.StringMinLength != "" {
					strConstraints.MinLength = pulumi.StringPtr(attr.StringMinLength)
				}
				if attr.StringMaxLength != "" {
					strConstraints.MaxLength = pulumi.StringPtr(attr.StringMaxLength)
				}
				schemaArgs.StringAttributeConstraints = strConstraints
			}

			if attr.AttributeDataType == "Number" && (attr.NumberMinValue != "" || attr.NumberMaxValue != "") {
				numConstraints := &cognito.UserPoolSchemaNumberAttributeConstraintsArgs{}
				if attr.NumberMinValue != "" {
					numConstraints.MinValue = pulumi.StringPtr(attr.NumberMinValue)
				}
				if attr.NumberMaxValue != "" {
					numConstraints.MaxValue = pulumi.StringPtr(attr.NumberMaxValue)
				}
				schemaArgs.NumberAttributeConstraints = numConstraints
			}

			schemas = append(schemas, schemaArgs)
		}
		args.Schemas = schemas
	}

	// ---------------------------------------------------------------------------
	// Lambda triggers
	// ---------------------------------------------------------------------------

	if spec.LambdaConfig != nil {
		lc := spec.LambdaConfig
		lambdaArgs := &cognito.UserPoolLambdaConfigArgs{}

		if lc.PreSignUp.GetValue() != "" {
			lambdaArgs.PreSignUp = pulumi.StringPtr(lc.PreSignUp.GetValue())
		}
		if lc.PreAuthentication.GetValue() != "" {
			lambdaArgs.PreAuthentication = pulumi.StringPtr(lc.PreAuthentication.GetValue())
		}
		if lc.PostAuthentication.GetValue() != "" {
			lambdaArgs.PostAuthentication = pulumi.StringPtr(lc.PostAuthentication.GetValue())
		}
		if lc.PostConfirmation.GetValue() != "" {
			lambdaArgs.PostConfirmation = pulumi.StringPtr(lc.PostConfirmation.GetValue())
		}
		if lc.PreTokenGeneration.GetValue() != "" {
			lambdaArgs.PreTokenGeneration = pulumi.StringPtr(lc.PreTokenGeneration.GetValue())
		}
		if lc.CustomMessage.GetValue() != "" {
			lambdaArgs.CustomMessage = pulumi.StringPtr(lc.CustomMessage.GetValue())
		}
		if lc.UserMigration.GetValue() != "" {
			lambdaArgs.UserMigration = pulumi.StringPtr(lc.UserMigration.GetValue())
		}
		if lc.DefineAuthChallenge.GetValue() != "" {
			lambdaArgs.DefineAuthChallenge = pulumi.StringPtr(lc.DefineAuthChallenge.GetValue())
		}
		if lc.CreateAuthChallenge.GetValue() != "" {
			lambdaArgs.CreateAuthChallenge = pulumi.StringPtr(lc.CreateAuthChallenge.GetValue())
		}
		if lc.VerifyAuthChallengeResponse.GetValue() != "" {
			lambdaArgs.VerifyAuthChallengeResponse = pulumi.StringPtr(lc.VerifyAuthChallengeResponse.GetValue())
		}

		args.LambdaConfig = lambdaArgs
	}

	// ---------------------------------------------------------------------------
	// Create user pool
	// ---------------------------------------------------------------------------

	created, err := cognito.NewUserPool(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Cognito user pool")
	}

	// ---------------------------------------------------------------------------
	// Exports
	// ---------------------------------------------------------------------------

	ctx.Export(OpUserPoolId, created.ID())
	ctx.Export(OpUserPoolArn, created.Arn)

	// The OIDC endpoint is: https://cognito-idp.{region}.amazonaws.com/{pool_id}
	ctx.Export(OpUserPoolEndpoint, created.Endpoint)

	return created, nil
}

// formatRegionPoolDomain builds the full Cognito-hosted domain URL.
func formatRegionPoolDomain(region, domain string) string {
	return fmt.Sprintf("https://%s.auth.%s.amazoncognito.com", domain, region)
}
