package awscognitouserpoolv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
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

// minimalClient returns a minimal valid app client for reuse in tests.
func minimalClient(name string) *AwsCognitoUserPoolClient {
	return &AwsCognitoUserPoolClient{
		Name: name,
	}
}

var _ = ginkgo.Describe("AwsCognitoUserPoolSpec validations", func() {
	var spec *AwsCognitoUserPoolSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: email sign-in, one app client.
		spec = &AwsCognitoUserPoolSpec{
			UsernameAttributes: []string{"email"},
			Clients: []*AwsCognitoUserPoolClient{
				minimalClient("web-app"),
			},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec with email sign-in and one client", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts phone_number as username attribute", func() {
		spec.UsernameAttributes = []string{"phone_number"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts alias attributes instead of username attributes", func() {
		spec.UsernameAttributes = nil
		spec.AliasAttributes = []string{"email", "preferred_username"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts neither username nor alias attributes (username-based pool)", func() {
		spec.UsernameAttributes = nil
		spec.AliasAttributes = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a password policy", func() {
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{
			MinimumLength:                 12,
			RequireLowercase:              true,
			RequireUppercase:              true,
			RequireNumbers:                true,
			RequireSymbols:                true,
			TemporaryPasswordValidityDays: 7,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MFA OPTIONAL with software token", func() {
		spec.MfaConfiguration = "OPTIONAL"
		spec.SoftwareTokenMfaEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MFA ON with software token", func() {
		spec.MfaConfiguration = "ON"
		spec.SoftwareTokenMfaEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts auto_verified_attributes", func() {
		spec.AutoVerifiedAttributes = []string{"email"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts account recovery mechanisms", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "verified_email", Priority: 1},
			{Name: "verified_phone_number", Priority: 2},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts email configuration with DEVELOPER mode and SES", func() {
		spec.EmailConfiguration = &AwsCognitoUserPoolEmailConfig{
			EmailSendingAccount: "DEVELOPER",
			SourceArn:           strRef("arn:aws:ses:us-east-1:123456789012:identity/noreply@example.com"),
			FromEmailAddress:    "No Reply <noreply@example.com>",
			ReplyToEmailAddress: "support@example.com",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts email configuration with COGNITO_DEFAULT mode", func() {
		spec.EmailConfiguration = &AwsCognitoUserPoolEmailConfig{
			EmailSendingAccount: "COGNITO_DEFAULT",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts admin-only user creation with deletion protection", func() {
		spec.AllowAdminCreateUserOnly = true
		spec.DeletionProtection = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts custom attributes", func() {
		spec.CustomAttributes = []*AwsCognitoUserPoolSchemaAttribute{
			{
				Name:              "tenant_id",
				AttributeDataType: "String",
				Mutable:           true,
				StringMinLength:   "1",
				StringMaxLength:   "64",
			},
			{
				Name:              "employee_number",
				AttributeDataType: "Number",
				Mutable:           false,
				NumberMinValue:    "1",
				NumberMaxValue:    "999999",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Lambda triggers", func() {
		spec.LambdaConfig = &AwsCognitoUserPoolLambdaConfig{
			PreSignUp:          strRef("arn:aws:lambda:us-east-1:123456789012:function:pre-signup"),
			PostConfirmation:   strRef("arn:aws:lambda:us-east-1:123456789012:function:post-confirm"),
			PreTokenGeneration: strRef("arn:aws:lambda:us-east-1:123456789012:function:pre-token"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an OAuth-enabled client with callbacks", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{
				Name:                            "web-app",
				AllowedOauthFlowsUserPoolClient: true,
				AllowedOauthFlows:               []string{"code"},
				AllowedOauthScopes:              []string{"openid", "email", "profile"},
				CallbackUrls:                    []string{"https://app.example.com/callback"},
				LogoutUrls:                      []string{"https://app.example.com/logout"},
				DefaultRedirectUri:              "https://app.example.com/callback",
				SupportedIdentityProviders:      []string{"COGNITO"},
				ExplicitAuthFlows:               []string{"ALLOW_USER_SRP_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"},
				AccessTokenValidityMinutes:      60,
				IdTokenValidityMinutes:          60,
				RefreshTokenValidityDays:        30,
				EnableTokenRevocation:           true,
				PreventUserExistenceErrors:      "ENABLED",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multiple clients", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: "web-app"},
			{Name: "api-server", GenerateSecret: true},
			{Name: "mobile-app"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a Cognito-hosted domain", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{
			Domain: "myapp-auth",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a custom domain with certificate", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{
			Domain:         "auth.example.com",
			CertificateArn: strRef("arn:aws:acm:us-east-1:123456789012:certificate/abc-def-ghi"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready full configuration", func() {
		spec.UsernameAttributes = []string{"email"}
		spec.UsernameCaseSensitive = false
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{
			MinimumLength:                 12,
			RequireLowercase:              true,
			RequireUppercase:              true,
			RequireNumbers:                true,
			RequireSymbols:                true,
			TemporaryPasswordValidityDays: 3,
		}
		spec.MfaConfiguration = "OPTIONAL"
		spec.SoftwareTokenMfaEnabled = true
		spec.AutoVerifiedAttributes = []string{"email"}
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "verified_email", Priority: 1},
		}
		spec.EmailConfiguration = &AwsCognitoUserPoolEmailConfig{
			EmailSendingAccount: "DEVELOPER",
			SourceArn:           strRef("arn:aws:ses:us-east-1:123456789012:identity/noreply@example.com"),
			FromEmailAddress:    "No Reply <noreply@example.com>",
		}
		spec.DeletionProtection = true
		spec.AllowAdminCreateUserOnly = false
		spec.LambdaConfig = &AwsCognitoUserPoolLambdaConfig{
			PreSignUp: strRef("arn:aws:lambda:us-east-1:123456789012:function:pre-signup"),
		}
		spec.CustomAttributes = []*AwsCognitoUserPoolSchemaAttribute{
			{Name: "tenant_id", AttributeDataType: "String", Mutable: true, StringMaxLength: "64"},
		}
		spec.Clients = []*AwsCognitoUserPoolClient{
			{
				Name:                            "web-app",
				AllowedOauthFlowsUserPoolClient: true,
				AllowedOauthFlows:               []string{"code"},
				AllowedOauthScopes:              []string{"openid", "email", "profile"},
				CallbackUrls:                    []string{"https://app.example.com/callback"},
				LogoutUrls:                      []string{"https://app.example.com/logout"},
				ExplicitAuthFlows:               []string{"ALLOW_USER_SRP_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"},
				EnableTokenRevocation:           true,
				PreventUserExistenceErrors:      "ENABLED",
			},
			{
				Name:                            "api-server",
				GenerateSecret:                  true,
				AllowedOauthFlowsUserPoolClient: true,
				AllowedOauthFlows:               []string{"client_credentials"},
				AllowedOauthScopes:              []string{"api/read", "api/write"},
				ExplicitAuthFlows:               []string{"ALLOW_USER_SRP_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"},
			},
		}
		spec.Domain = &AwsCognitoUserPoolDomainConfig{
			Domain: "myapp-auth",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure scenarios
	// -------------------------------------------------------------------------

	ginkgo.It("rejects when both username_attributes and alias_attributes are set", func() {
		spec.UsernameAttributes = []string{"email"}
		spec.AliasAttributes = []string{"phone_number"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid username_attributes value", func() {
		spec.UsernameAttributes = []string{"email", "invalid_attr"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid alias_attributes value", func() {
		spec.UsernameAttributes = nil
		spec.AliasAttributes = []string{"invalid_attr"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid auto_verified_attributes value", func() {
		spec.AutoVerifiedAttributes = []string{"name"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid mfa_configuration value", func() {
		spec.MfaConfiguration = "INVALID"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects software_token_mfa when MFA is OFF", func() {
		spec.MfaConfiguration = "OFF"
		spec.SoftwareTokenMfaEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects software_token_mfa when MFA is empty (default OFF)", func() {
		spec.MfaConfiguration = ""
		spec.SoftwareTokenMfaEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects no clients (min_items = 1)", func() {
		spec.Clients = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects duplicate client names", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: "web-app"},
			{Name: "web-app"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects client with empty name", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: ""},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid allowed_oauth_flows value", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: "web-app", AllowedOauthFlows: []string{"invalid_flow"}},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid explicit_auth_flows value", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: "web-app", ExplicitAuthFlows: []string{"NOT_A_REAL_FLOW"}},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid prevent_user_existence_errors value", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: "web-app", PreventUserExistenceErrors: "INVALID"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects password minimum_length below 6", func() {
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{MinimumLength: 3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects password minimum_length above 99", func() {
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{MinimumLength: 100}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects temporary_password_validity_days above 365", func() {
		spec.PasswordPolicy = &AwsCognitoUserPoolPasswordPolicy{
			MinimumLength:                 8,
			TemporaryPasswordValidityDays: 400,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid account recovery mechanism name", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "invalid_method", Priority: 1},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects account recovery mechanism with priority 0", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "verified_email", Priority: 0},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects account recovery mechanism with priority 3", func() {
		spec.AccountRecoveryMechanisms = []*AwsCognitoUserPoolRecoveryMechanism{
			{Name: "verified_email", Priority: 3},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects DEVELOPER email without source_arn", func() {
		spec.EmailConfiguration = &AwsCognitoUserPoolEmailConfig{
			EmailSendingAccount: "DEVELOPER",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid email_sending_account value", func() {
		spec.EmailConfiguration = &AwsCognitoUserPoolEmailConfig{
			EmailSendingAccount: "INVALID",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid custom attribute data type", func() {
		spec.CustomAttributes = []*AwsCognitoUserPoolSchemaAttribute{
			{Name: "test", AttributeDataType: "InvalidType"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects custom attribute with empty name", func() {
		spec.CustomAttributes = []*AwsCognitoUserPoolSchemaAttribute{
			{Name: "", AttributeDataType: "String"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects custom attribute with name longer than 20 chars", func() {
		spec.CustomAttributes = []*AwsCognitoUserPoolSchemaAttribute{
			{Name: "this_name_is_way_too_long_for_cognito", AttributeDataType: "String"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects custom domain without certificate_arn", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{
			Domain: "auth.example.com",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects domain with empty domain string", func() {
		spec.Domain = &AwsCognitoUserPoolDomainConfig{
			Domain: "",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects access_token_validity_minutes above 1440", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: "web-app", AccessTokenValidityMinutes: 2000},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects refresh_token_validity_days above 3650", func() {
		spec.Clients = []*AwsCognitoUserPoolClient{
			{Name: "web-app", RefreshTokenValidityDays: 4000},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
