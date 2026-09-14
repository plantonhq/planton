package awscognitouserpoolv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsCognitoUserPoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCognitoUserPoolSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsCognitoUserPoolSpec validations", func() {
	var spec *AwsCognitoUserPoolSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: region + email sign-in.
		spec = &AwsCognitoUserPoolSpec{
			Region:             "us-west-2",
			UsernameAttributes: []string{"email"},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec with email sign-in", func() {
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts alias attributes instead of username attributes", func() {
		spec.UsernameAttributes = nil
		spec.AliasAttributes = []string{"email", "preferred_username"}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts neither username nor alias attributes (username-based pool)", func() {
		spec.UsernameAttributes = nil
		spec.AliasAttributes = nil
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full password policy with history", func() {
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{
			MinimumLength:                 12,
			RequireLowercase:              true,
			RequireUppercase:              true,
			RequireNumbers:                true,
			RequireSymbols:                true,
			PasswordHistorySize:           5,
			TemporaryPasswordValidityDays: 7,
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a passwordless sign-in policy with email OTP and passkeys", func() {
		spec.AllowedFirstAuthFactors = []string{"PASSWORD", "EMAIL_OTP", "WEB_AUTHN"}
		spec.WebAuthn = &AwsCognitoUserPoolWebAuthnConfig{
			RelyingPartyId:   "auth.example.com",
			UserVerification: "required",
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts SMS_OTP when sms_configuration is set", func() {
		spec.AllowedFirstAuthFactors = []string{"PASSWORD", "SMS_OTP"}
		spec.SmsConfiguration = &AwsCognitoUserPoolSmsConfig{
			SnsCallerArn: strRef("arn:aws:iam::123456789012:role/cognito-sms"),
			ExternalId:   "pool-external-id",
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts email MFA alongside OPTIONAL mfa", func() {
		spec.MfaConfiguration = "OPTIONAL"
		spec.EmailMfa = &AwsCognitoUserPoolEmailMfaConfig{
			Message: "Your code is {####}",
			Subject: "Sign-in code",
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts email MFA declared and switched off while mfa_configuration is OFF", func() {
		spec.MfaConfiguration = "OFF"
		spec.EmailMfa = &AwsCognitoUserPoolEmailMfaConfig{Enabled: proto.Bool(false), Subject: "Sign-in code"}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts device configuration that states one dial and leaves the other unset", func() {
		spec.DeviceConfiguration = &AwsCognitoUserPoolDeviceConfig{ChallengeRequiredOnNewDevice: proto.Bool(true)}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		gomega.Expect(spec.DeviceConfiguration.DeviceOnlyRememberedOnUserPrompt).To(gomega.BeNil(), "the unset dial stays unset at the schema layer; the platform fills its declared default before the module reads it")
	})

	ginkgo.It("accepts a PLUS-tier pool with enforced threat protection and auth-event logging", func() {
		spec.UserPoolTier = "PLUS"
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{
			AdvancedSecurityMode: "ENFORCED",
			CustomAuthMode:       "AUDIT",
		}
		spec.LogConfigurations = []*AwsCognitoUserPoolLogConfiguration{
			{
				EventSource:           "userAuthEvents",
				LogLevel:              "INFO",
				CloudwatchLogGroupArn: strRef("arn:aws:logs:us-west-2:123456789012:log-group:/aws/cognito/pool"),
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts verification and invite message templates with placeholders", func() {
		spec.VerificationMessageTemplate = &AwsCognitoUserPoolVerificationMessageTemplate{
			DefaultEmailOption: "CONFIRM_WITH_LINK",
			EmailMessageByLink: "Click {##here##} to verify your address.",
			EmailSubjectByLink: "Verify your address",
			SmsMessage:         "Your verification code is {####}",
		}
		spec.InviteMessageTemplate = &AwsCognitoUserPoolInviteMessageTemplate{
			EmailMessage: "Welcome {username}, your temporary password is {####}",
			EmailSubject: "Your invitation",
			SmsMessage:   "{username}, your temporary password is {####}",
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a versioned pre-token-generation trigger with custom senders", func() {
		spec.LambdaConfig = &AwsCognitoUserPoolLambdaConfig{
			PreTokenGenerationConfig: &AwsCognitoUserPoolPreTokenGenerationConfig{
				LambdaArn:     strRef("arn:aws:lambda:us-west-2:123456789012:function:claims"),
				LambdaVersion: "V2_0",
			},
			CustomEmailSender: &AwsCognitoUserPoolCustomSenderConfig{
				LambdaArn:     strRef("arn:aws:lambda:us-west-2:123456789012:function:mailer"),
				LambdaVersion: "V1_0",
			},
			KmsKeyId: strRef("arn:aws:kms:us-west-2:123456789012:key/abcd-1234"),
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a custom domain with a certificate and managed login", func() {
		v := int32(2)
		spec.Domain = &AwsCognitoUserPoolDomainConfig{
			Domain:              "auth.example.com",
			CertificateArn:      strRef("arn:aws:acm:us-east-1:123456789012:certificate/abcd"),
			ManagedLoginVersion: &v,
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Required / range validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when region is empty", func() {
		spec.Region = ""
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when password minimum_length is below 6", func() {
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{MinimumLength: 4}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when password_history_size exceeds 24", func() {
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{MinimumLength: 8, PasswordHistorySize: 25}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when managed_login_version is out of range", func() {
		v := int32(3)
		spec.Domain = &AwsCognitoUserPoolDomainConfig{Domain: "myapp-auth", ManagedLoginVersion: &v}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: identity model
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both username and alias attributes are set", func() {
		spec.AliasAttributes = []string{"preferred_username"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid username attribute", func() {
		spec.UsernameAttributes = []string{"nickname"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid alias attribute", func() {
		spec.UsernameAttributes = nil
		spec.AliasAttributes = []string{"given_name"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: tier, factors, MFA
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid user_pool_tier", func() {
		spec.UserPoolTier = "PREMIUM"
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid first auth factor", func() {
		spec.AllowedFirstAuthFactors = []string{"MAGIC_LINK"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when SMS_OTP is allowed without sms_configuration", func() {
		spec.AllowedFirstAuthFactors = []string{"SMS_OTP"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid mfa_configuration", func() {
		spec.MfaConfiguration = "REQUIRED"
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when software token MFA is enabled with MFA off", func() {
		spec.SoftwareTokenMfaEnabled = true
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when email MFA is configured with MFA off", func() {
		spec.EmailMfa = &AwsCognitoUserPoolEmailMfaConfig{Subject: "Code"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the email MFA message lacks the code placeholder", func() {
		spec.MfaConfiguration = "OPTIONAL"
		spec.EmailMfa = &AwsCognitoUserPoolEmailMfaConfig{Message: "Your code is 123456"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid web_authn user_verification", func() {
		spec.WebAuthn = &AwsCognitoUserPoolWebAuthnConfig{UserVerification: "always"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when sms_authentication_message lacks the code placeholder", func() {
		spec.SmsAuthenticationMessage = "Your sign-in code has been sent"
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: verification and recovery
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid auto-verified attribute", func() {
		spec.AutoVerifiedAttributes = []string{"address"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid attributes_require_verification_before_update value", func() {
		spec.AttributesRequireVerificationBeforeUpdate = []string{"nickname"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid recovery mechanism name", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "security_questions", Priority: 1},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on a recovery priority outside 1-2", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "verified_email", Priority: 3},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when admin_only is combined with another recovery mechanism", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "verified_email", Priority: 1},
			{Name: "admin_only", Priority: 2},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts admin_only as the sole recovery mechanism", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "admin_only", Priority: 1},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("fails on a prefix domain containing a reserved word", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{Domain: "my-aws-auth"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a custom domain containing a reserved word", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{
			Domain: "auth.aws-tooling.example.com",
			CertificateArn: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm:us-east-1:123456789012:certificate/abc"},
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: email configuration
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid email_sending_account", func() {
		spec.EmailConfiguration = &AwsCognitoUserPoolEmailConfig{EmailSendingAccount: "SES"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when DEVELOPER email mode lacks source_arn", func() {
		spec.EmailConfiguration = &AwsCognitoUserPoolEmailConfig{EmailSendingAccount: "DEVELOPER"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: message templates
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid default_email_option", func() {
		spec.VerificationMessageTemplate = &AwsCognitoUserPoolVerificationMessageTemplate{
			DefaultEmailOption: "CONFIRM_WITH_OTP",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the code-based verification email lacks its placeholder", func() {
		spec.VerificationMessageTemplate = &AwsCognitoUserPoolVerificationMessageTemplate{
			EmailMessage: "Please verify your address",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the link-based verification email lacks its placeholder pair", func() {
		spec.VerificationMessageTemplate = &AwsCognitoUserPoolVerificationMessageTemplate{
			EmailMessageByLink: "Click here to verify",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the verification SMS lacks its placeholder", func() {
		spec.VerificationMessageTemplate = &AwsCognitoUserPoolVerificationMessageTemplate{
			SmsMessage: "Here is your code",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the invite email lacks the username placeholder", func() {
		spec.InviteMessageTemplate = &AwsCognitoUserPoolInviteMessageTemplate{
			EmailMessage: "Your temporary password is {####}",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the invite SMS lacks the password placeholder", func() {
		spec.InviteMessageTemplate = &AwsCognitoUserPoolInviteMessageTemplate{
			SmsMessage: "Welcome {username}",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: custom attributes
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid attribute_data_type", func() {
		spec.CustomAttributes = []*AwsCognitoUserPoolSchemaAttribute{
			{Name: "tenant_id", AttributeDataType: "Text"},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: lambda triggers
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both pre_token_generation spellings are set", func() {
		spec.LambdaConfig = &AwsCognitoUserPoolLambdaConfig{
			PreTokenGeneration: strRef("arn:aws:lambda:us-west-2:123456789012:function:claims"),
			PreTokenGenerationConfig: &AwsCognitoUserPoolPreTokenGenerationConfig{
				LambdaArn:     strRef("arn:aws:lambda:us-west-2:123456789012:function:claims"),
				LambdaVersion: "V2_0",
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a custom sender is set without the KMS key", func() {
		spec.LambdaConfig = &AwsCognitoUserPoolLambdaConfig{
			CustomSmsSender: &AwsCognitoUserPoolCustomSenderConfig{
				LambdaArn:     strRef("arn:aws:lambda:us-west-2:123456789012:function:sms"),
				LambdaVersion: "V1_0",
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid pre_token_generation_config version", func() {
		spec.LambdaConfig = &AwsCognitoUserPoolLambdaConfig{
			PreTokenGenerationConfig: &AwsCognitoUserPoolPreTokenGenerationConfig{
				LambdaArn:     strRef("arn:aws:lambda:us-west-2:123456789012:function:claims"),
				LambdaVersion: "V4_0",
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid custom sender version", func() {
		spec.LambdaConfig = &AwsCognitoUserPoolLambdaConfig{
			CustomEmailSender: &AwsCognitoUserPoolCustomSenderConfig{
				LambdaArn:     strRef("arn:aws:lambda:us-west-2:123456789012:function:mailer"),
				LambdaVersion: "V2_0",
			},
			KmsKeyId: strRef("arn:aws:kms:us-west-2:123456789012:key/abcd-1234"),
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: threat protection
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid advanced_security_mode", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "MONITOR"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when custom_auth_mode is set with advanced security off", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{
			AdvancedSecurityMode: "OFF",
			CustomAuthMode:       "ENFORCED",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: log delivery
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid log event_source", func() {
		spec.LogConfigurations = []*AwsCognitoUserPoolLogConfiguration{
			{EventSource: "signIn", LogLevel: "ERROR", CloudwatchLogGroupArn: strRef("arn:aws:logs:us-west-2:1:log-group:x")},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a log configuration has no destination", func() {
		spec.LogConfigurations = []*AwsCognitoUserPoolLogConfiguration{
			{EventSource: "userNotification", LogLevel: "ERROR"},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a log configuration has two destinations", func() {
		spec.LogConfigurations = []*AwsCognitoUserPoolLogConfiguration{
			{
				EventSource:           "userNotification",
				LogLevel:              "ERROR",
				CloudwatchLogGroupArn: strRef("arn:aws:logs:us-west-2:1:log-group:x"),
				S3BucketArn:           strRef("arn:aws:s3:::logs-bucket"),
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: log delivery value domains
	// -------------------------------------------------------------------------

	ginkgo.It("fails on an invalid log_level", func() {
		spec.LogConfigurations = []*AwsCognitoUserPoolLogConfiguration{
			{EventSource: "userNotification", LogLevel: "DEBUG", CloudwatchLogGroupArn: strRef("arn:aws:logs:us-west-2:1:log-group:x")},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid custom_auth_mode value", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{
			AdvancedSecurityMode: "ENFORCED",
			CustomAuthMode:       "OFF",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: domain
	// -------------------------------------------------------------------------

	ginkgo.It("fails when a custom domain lacks a certificate", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{Domain: "auth.example.com"}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a prefix domain without a certificate", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{Domain: "myapp-auth"}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// User groups
	// -------------------------------------------------------------------------

	ginkgo.It("accepts user groups with roles and precedence", func() {
		spec.UserGroups = []*AwsCognitoUserPoolUserGroup{
			{Name: "admins", Description: "Administrators", Precedence: 1, RoleArn: strRef("arn:aws:iam::123456789012:role/admins")},
			{Name: "readers"},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("fails on duplicate user group names", func() {
		spec.UserGroups = []*AwsCognitoUserPoolUserGroup{
			{Name: "admins"},
			{Name: "admins"},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an empty user group name", func() {
		spec.UserGroups = []*AwsCognitoUserPoolUserGroup{{Name: ""}}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Risk configuration
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a full risk configuration with threat protection active", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			AccountTakeover: &AwsCognitoUserPoolAccountTakeoverConfig{
				HighAction:   &AwsCognitoUserPoolAccountTakeoverAction{EventAction: "BLOCK", Notify: true},
				MediumAction: &AwsCognitoUserPoolAccountTakeoverAction{EventAction: "MFA_IF_CONFIGURED", Notify: true},
				NotifyConfiguration: &AwsCognitoUserPoolRiskNotifyConfig{
					SourceArn: strRef("arn:aws:ses:us-west-2:123456789012:identity/security@example.com"),
					BlockEmail: &AwsCognitoUserPoolRiskNotifyEmail{
						Subject:  "Sign-in attempt blocked",
						HtmlBody: "<p>We blocked a risky sign-in attempt on your account.</p>",
						TextBody: "We blocked a risky sign-in attempt on your account.",
					},
				},
			},
			CompromisedCredentials: &AwsCognitoUserPoolCompromisedCredentialsConfig{
				EventAction: "BLOCK",
				EventFilter: []string{"SIGN_IN", "SIGN_UP"},
			},
			RiskException: &AwsCognitoUserPoolRiskExceptionConfig{
				BlockedIpRanges: []string{"192.0.2.0/24"},
				SkippedIpRanges: []string{"2001:db8::/32"},
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("fails when risk configuration is set without threat protection", func() {
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			CompromisedCredentials: &AwsCognitoUserPoolCompromisedCredentialsConfig{EventAction: "BLOCK"},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an empty risk configuration", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "AUDIT"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an account takeover config with no actions", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			AccountTakeover: &AwsCognitoUserPoolAccountTakeoverConfig{},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on notify templates without a notifying action", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			AccountTakeover: &AwsCognitoUserPoolAccountTakeoverConfig{
				HighAction: &AwsCognitoUserPoolAccountTakeoverAction{EventAction: "BLOCK", Notify: false},
				NotifyConfiguration: &AwsCognitoUserPoolRiskNotifyConfig{
					SourceArn: strRef("arn:aws:ses:us-west-2:123456789012:identity/security@example.com"),
				},
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid account takeover event_action", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			AccountTakeover: &AwsCognitoUserPoolAccountTakeoverConfig{
				HighAction: &AwsCognitoUserPoolAccountTakeoverAction{EventAction: "DENY"},
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid compromised credentials event_action", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			CompromisedCredentials: &AwsCognitoUserPoolCompromisedCredentialsConfig{EventAction: "MFA_REQUIRED"},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid compromised credentials event_filter", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			CompromisedCredentials: &AwsCognitoUserPoolCompromisedCredentialsConfig{
				EventAction: "BLOCK",
				EventFilter: []string{"SIGN_OUT"},
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on a risk exception with neither list", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			RiskException: &AwsCognitoUserPoolRiskExceptionConfig{},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on a non-CIDR risk exception entry", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			RiskException: &AwsCognitoUserPoolRiskExceptionConfig{
				BlockedIpRanges: []string{"192.0.2.1"},
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on a notify email with a too-short body", func() {
		spec.UserPoolAddOns = &AwsCognitoUserPoolAddOns{AdvancedSecurityMode: "ENFORCED"}
		spec.RiskConfiguration = &AwsCognitoUserPoolRiskConfiguration{
			AccountTakeover: &AwsCognitoUserPoolAccountTakeoverConfig{
				HighAction: &AwsCognitoUserPoolAccountTakeoverAction{EventAction: "BLOCK", Notify: true},
				NotifyConfiguration: &AwsCognitoUserPoolRiskNotifyConfig{
					SourceArn:  strRef("arn:aws:ses:us-west-2:123456789012:identity/security@example.com"),
					BlockEmail: &AwsCognitoUserPoolRiskNotifyEmail{Subject: "Blocked", HtmlBody: "short", TextBody: "long enough body"},
				},
			},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})
})
