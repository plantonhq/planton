package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func userPool(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*cognito.UserPool, error) {
	spec := locals.Spec

	// The pool's cloud name is metadata.name -- the cross-engine naming basis.
	args := &cognito.UserPoolArgs{
		Name: pulumi.String(locals.Target.Metadata.Name),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// ---------------------------------------------------------------------------
	// Identity model (all ForceNew -- changing any of these replaces the pool
	// and destroys every user in it)
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
	// Pool-level posture
	// ---------------------------------------------------------------------------

	// AWS expresses deletion protection as an ACTIVE/INACTIVE string; the spec
	// keeps it an honest boolean and the module translates. Always sent (both
	// engines state-pin it) so a true->false edit deactivates protection
	// instead of silently no-opping.
	if spec.DeletionProtection {
		args.DeletionProtection = pulumi.StringPtr("ACTIVE")
	} else {
		args.DeletionProtection = pulumi.StringPtr("INACTIVE")
	}

	// Omitted tier means AWS's default (ESSENTIALS); only forward an explicit
	// choice so manifests that predate tiers keep deploying unchanged.
	if spec.UserPoolTier != "" {
		args.UserPoolTier = pulumi.StringPtr(spec.UserPoolTier)
	}

	// ---------------------------------------------------------------------------
	// Password and sign-in policy
	// ---------------------------------------------------------------------------

	// The zero-gates below are faithful, not lossy: AWS itself treats a
	// submitted 0 as null for temporary_password_validity_days (applying its
	// 7-day default), and 0 is AWS's own default posture for
	// password_history_size (history off) -- so 0 and absent are the same
	// policy at the API for all three numerics.
	if spec.PasswordPolicy != nil {
		pp := spec.PasswordPolicy
		passwordPolicy := &cognito.UserPoolPasswordPolicyArgs{
			RequireLowercase: pulumi.BoolPtr(pp.RequireLowercase),
			RequireUppercase: pulumi.BoolPtr(pp.RequireUppercase),
			RequireNumbers:   pulumi.BoolPtr(pp.RequireNumbers),
			RequireSymbols:   pulumi.BoolPtr(pp.RequireSymbols),
		}
		if pp.MinimumLength > 0 {
			passwordPolicy.MinimumLength = pulumi.IntPtr(int(pp.MinimumLength))
		}
		if pp.PasswordHistorySize > 0 {
			passwordPolicy.PasswordHistorySize = pulumi.IntPtr(int(pp.PasswordHistorySize))
		}
		if pp.TemporaryPasswordValidityDays > 0 {
			passwordPolicy.TemporaryPasswordValidityDays = pulumi.IntPtr(int(pp.TemporaryPasswordValidityDays))
		}
		args.PasswordPolicy = passwordPolicy
	}

	// The passwordless dial: listing first factors switches the pool to
	// choice-based sign-in for clients that enable ALLOW_USER_AUTH.
	if len(spec.AllowedFirstAuthFactors) > 0 {
		args.SignInPolicy = &cognito.UserPoolSignInPolicyArgs{
			AllowedFirstAuthFactors: pulumi.ToStringArray(spec.AllowedFirstAuthFactors),
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

	if spec.EmailMfa != nil && spec.EmailMfa.GetEnabled() {
		emailMfa := &cognito.UserPoolEmailMfaConfigurationArgs{}
		if spec.EmailMfa.Message != "" {
			emailMfa.Message = pulumi.StringPtr(spec.EmailMfa.Message)
		}
		if spec.EmailMfa.Subject != "" {
			emailMfa.Subject = pulumi.StringPtr(spec.EmailMfa.Subject)
		}
		args.EmailMfaConfiguration = emailMfa
	}

	if spec.WebAuthn != nil {
		webAuthn := &cognito.UserPoolWebAuthnConfigurationArgs{}
		if spec.WebAuthn.RelyingPartyId != "" {
			webAuthn.RelyingPartyId = pulumi.StringPtr(spec.WebAuthn.RelyingPartyId)
		}
		if spec.WebAuthn.UserVerification != "" {
			webAuthn.UserVerification = pulumi.StringPtr(spec.WebAuthn.UserVerification)
		}
		args.WebAuthnConfiguration = webAuthn
	}

	// ---------------------------------------------------------------------------
	// SMS delivery (Cognito assumes the referenced role to publish through SNS)
	// ---------------------------------------------------------------------------

	if spec.SmsConfiguration != nil {
		smsConfiguration := &cognito.UserPoolSmsConfigurationArgs{
			ExternalId:   pulumi.String(spec.SmsConfiguration.ExternalId),
			SnsCallerArn: pulumi.String(spec.SmsConfiguration.SnsCallerArn.GetValue()),
		}
		if spec.SmsConfiguration.SnsRegion != "" {
			smsConfiguration.SnsRegion = pulumi.StringPtr(spec.SmsConfiguration.SnsRegion)
		}
		args.SmsConfiguration = smsConfiguration
	}

	if spec.SmsAuthenticationMessage != "" {
		args.SmsAuthenticationMessage = pulumi.StringPtr(spec.SmsAuthenticationMessage)
	}

	// ---------------------------------------------------------------------------
	// Verification and recovery
	// ---------------------------------------------------------------------------

	if len(spec.AutoVerifiedAttributes) > 0 {
		args.AutoVerifiedAttributes = pulumi.ToStringArray(spec.AutoVerifiedAttributes)
	}

	// Keep the previous value active until the new one verifies -- without
	// this, an unverified typo in an email update can lock a user out.
	if len(spec.AttributesRequireVerificationBeforeUpdate) > 0 {
		args.UserAttributeUpdateSettings = &cognito.UserPoolUserAttributeUpdateSettingsArgs{
			AttributesRequireVerificationBeforeUpdates: pulumi.ToStringArray(spec.AttributesRequireVerificationBeforeUpdate),
		}
	}

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
		if ec.ConfigurationSet.GetValue() != "" {
			emailArgs.ConfigurationSet = pulumi.StringPtr(ec.ConfigurationSet.GetValue())
		}

		args.EmailConfiguration = emailArgs
	}

	// The modern verification_message_template block is the single spelling
	// this module forwards -- the provider's legacy top-level message/subject
	// fields conflict with it and are deliberately not modeled.
	if spec.VerificationMessageTemplate != nil {
		vt := spec.VerificationMessageTemplate
		template := &cognito.UserPoolVerificationMessageTemplateArgs{}
		if vt.DefaultEmailOption != "" {
			template.DefaultEmailOption = pulumi.StringPtr(vt.DefaultEmailOption)
		}
		if vt.EmailMessage != "" {
			template.EmailMessage = pulumi.StringPtr(vt.EmailMessage)
		}
		if vt.EmailSubject != "" {
			template.EmailSubject = pulumi.StringPtr(vt.EmailSubject)
		}
		if vt.EmailMessageByLink != "" {
			template.EmailMessageByLink = pulumi.StringPtr(vt.EmailMessageByLink)
		}
		if vt.EmailSubjectByLink != "" {
			template.EmailSubjectByLink = pulumi.StringPtr(vt.EmailSubjectByLink)
		}
		if vt.SmsMessage != "" {
			template.SmsMessage = pulumi.StringPtr(vt.SmsMessage)
		}
		args.VerificationMessageTemplate = template
	}

	// ---------------------------------------------------------------------------
	// Admin create user (self-registration gate + invitation templates)
	// ---------------------------------------------------------------------------

	if spec.AllowAdminCreateUserOnly || spec.InviteMessageTemplate != nil {
		// The flag is always sent inside the block (both engines state-pin
		// it): a template-only configuration explicitly keeps
		// self-registration open rather than leaving the choice unstated.
		adminCreateUser := &cognito.UserPoolAdminCreateUserConfigArgs{
			AllowAdminCreateUserOnly: pulumi.BoolPtr(spec.AllowAdminCreateUserOnly),
		}
		if spec.InviteMessageTemplate != nil {
			it := spec.InviteMessageTemplate
			invite := &cognito.UserPoolAdminCreateUserConfigInviteMessageTemplateArgs{}
			if it.EmailMessage != "" {
				invite.EmailMessage = pulumi.StringPtr(it.EmailMessage)
			}
			if it.EmailSubject != "" {
				invite.EmailSubject = pulumi.StringPtr(it.EmailSubject)
			}
			if it.SmsMessage != "" {
				invite.SmsMessage = pulumi.StringPtr(it.SmsMessage)
			}
			adminCreateUser.InviteMessageTemplate = invite
		}
		args.AdminCreateUserConfig = adminCreateUser
	}

	// ---------------------------------------------------------------------------
	// Remembered devices
	// ---------------------------------------------------------------------------

	if spec.DeviceConfiguration != nil {
		args.DeviceConfiguration = &cognito.UserPoolDeviceConfigurationArgs{
			ChallengeRequiredOnNewDevice:     pulumi.BoolPtr(spec.DeviceConfiguration.GetChallengeRequiredOnNewDevice()),
			DeviceOnlyRememberedOnUserPrompt: pulumi.BoolPtr(spec.DeviceConfiguration.GetDeviceOnlyRememberedOnUserPrompt()),
		}
	}

	// ---------------------------------------------------------------------------
	// Custom attributes (schema is append-only in AWS)
	// ---------------------------------------------------------------------------

	if len(spec.CustomAttributes) > 0 {
		var schemas cognito.UserPoolSchemaArray
		for _, attr := range spec.CustomAttributes {
			// All three flags are always sent (both engines state-pin them):
			// they are fixed at the moment the attribute is added, so an
			// unstated flag would freeze AWS's default invisibly.
			schemaArgs := &cognito.UserPoolSchemaArgs{
				Name:                   pulumi.String(attr.Name),
				AttributeDataType:      pulumi.String(attr.AttributeDataType),
				Mutable:                pulumi.BoolPtr(attr.Mutable),
				Required:               pulumi.BoolPtr(attr.Required),
				DeveloperOnlyAttribute: pulumi.BoolPtr(attr.DeveloperOnlyAttribute),
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
		// The plain field pins the V1_0 event; the config block selects the
		// version explicitly. The spec's CEL keeps them mutually exclusive.
		if lc.PreTokenGeneration.GetValue() != "" {
			lambdaArgs.PreTokenGeneration = pulumi.StringPtr(lc.PreTokenGeneration.GetValue())
		}
		if lc.PreTokenGenerationConfig != nil {
			lambdaArgs.PreTokenGenerationConfig = &cognito.UserPoolLambdaConfigPreTokenGenerationConfigArgs{
				LambdaArn:     pulumi.String(lc.PreTokenGenerationConfig.LambdaArn.GetValue()),
				LambdaVersion: pulumi.String(lc.PreTokenGenerationConfig.LambdaVersion),
			}
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
		// Custom senders deliver the message themselves; Cognito encrypts the
		// code with the KMS key before handing it to the function, which is
		// why the spec couples the key to the senders.
		if lc.CustomEmailSender != nil {
			lambdaArgs.CustomEmailSender = &cognito.UserPoolLambdaConfigCustomEmailSenderArgs{
				LambdaArn:     pulumi.String(lc.CustomEmailSender.LambdaArn.GetValue()),
				LambdaVersion: pulumi.String(lc.CustomEmailSender.LambdaVersion),
			}
		}
		if lc.CustomSmsSender != nil {
			lambdaArgs.CustomSmsSender = &cognito.UserPoolLambdaConfigCustomSmsSenderArgs{
				LambdaArn:     pulumi.String(lc.CustomSmsSender.LambdaArn.GetValue()),
				LambdaVersion: pulumi.String(lc.CustomSmsSender.LambdaVersion),
			}
		}
		if lc.KmsKeyId.GetValue() != "" {
			lambdaArgs.KmsKeyId = pulumi.StringPtr(lc.KmsKeyId.GetValue())
		}

		args.LambdaConfig = lambdaArgs
	}

	// ---------------------------------------------------------------------------
	// Threat protection
	// ---------------------------------------------------------------------------

	if spec.UserPoolAddOns != nil {
		addOns := &cognito.UserPoolUserPoolAddOnsArgs{
			AdvancedSecurityMode: pulumi.String(spec.UserPoolAddOns.AdvancedSecurityMode),
		}
		if spec.UserPoolAddOns.CustomAuthMode != "" {
			addOns.AdvancedSecurityAdditionalFlows = &cognito.UserPoolUserPoolAddOnsAdvancedSecurityAdditionalFlowsArgs{
				CustomAuthMode: pulumi.StringPtr(spec.UserPoolAddOns.CustomAuthMode),
			}
		}
		args.UserPoolAddOns = addOns
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

	// AWS reports the endpoint WITHOUT a scheme
	// ("cognito-idp.{region}.amazonaws.com/{pool_id}"); the issuer output adds
	// it so JWT authorizers can consume the value directly.
	ctx.Export(OpUserPoolEndpoint, created.Endpoint)
	ctx.Export(OpIssuer, pulumi.Sprintf("https://%s", created.Endpoint))

	return created, nil
}
