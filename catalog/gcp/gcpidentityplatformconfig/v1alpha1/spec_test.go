package gcpidentityplatformconfigv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpIdentityPlatformConfigSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpIdentityPlatformConfigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpIdentityPlatformConfig {
		return &GcpIdentityPlatformConfig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpIdentityPlatformConfig",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-idp-config",
			},
			Spec: &GcpIdentityPlatformConfigSpec{
				SignIn: &GcpIdentityPlatformConfigSignIn{
					Email: &GcpIdentityPlatformConfigSignInEmail{
						Enabled:          proto.Bool(true),
						PasswordRequired: true,
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal email/password config", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a sign-in arm switched off and refuse one that does not say on or off", func() {
		target := minimal()
		target.Spec.SignIn = &GcpIdentityPlatformConfigSignIn{
			Anonymous: &GcpIdentityPlatformConfigSignInAnonymous{Enabled: proto.Bool(false)},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.SignIn.Anonymous = &GcpIdentityPlatformConfigSignInAnonymous{}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("anonymous.enabled"))
	})

	ginkgo.It("should accept an entirely empty spec (initialize-only)", func() {
		target := minimal()
		target.Spec = &GcpIdentityPlatformConfigSpec{}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal and authorized domains", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		target.Spec.AuthorizedDomains = []string{"app.example.com"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every sign-in arm plus anonymous autodelete", func() {
		target := minimal()
		target.Spec.SignIn.PhoneNumber = &GcpIdentityPlatformConfigSignInPhone{
			Enabled:          proto.Bool(true),
			TestPhoneNumbers: map[string]string{"+15555550100": "123456"},
		}
		target.Spec.SignIn.Anonymous = &GcpIdentityPlatformConfigSignInAnonymous{Enabled: proto.Bool(true)}
		target.Spec.SignIn.AllowDuplicateEmails = false
		target.Spec.AutodeleteAnonymousUsers = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a full MFA policy", func() {
		target := minimal()
		target.Spec.Mfa = &GcpIdentityPlatformConfigMfa{
			State:            "ENABLED",
			EnabledProviders: []string{"PHONE_SMS"},
			ProviderConfigs: []*GcpIdentityPlatformConfigMfaProviderConfig{{
				State: "MANDATORY",
				TotpProviderConfig: &GcpIdentityPlatformConfigMfaTotp{
					AdjacentIntervals: 5,
				},
			}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept blocking functions with a trigger and forwarding", func() {
		target := minimal()
		target.Spec.BlockingFunctions = &GcpIdentityPlatformConfigBlockingFunctions{
			Triggers: []*GcpIdentityPlatformConfigBlockingFunctionTrigger{{
				EventType:   "beforeSignIn",
				FunctionUri: litRef("https://us-central1-p.cloudfunctions.net/check"),
			}},
			ForwardInboundCredentials: &GcpIdentityPlatformConfigForwardInboundCredentials{
				IdToken: true,
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a complete sign-up quota", func() {
		target := minimal()
		target.Spec.SignUpQuota = &GcpIdentityPlatformConfigSignUpQuota{
			Quota:         100,
			QuotaDuration: "86400s",
			StartTime:     "2026-09-01T00:00:00Z",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each SMS region policy arm alone", func() {
		target := minimal()
		target.Spec.SmsRegionConfig = &GcpIdentityPlatformConfigSmsRegionConfig{
			AllowByDefault: &GcpIdentityPlatformConfigSmsAllowByDefault{
				DisallowedRegions: []string{"KP"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.SmsRegionConfig = &GcpIdentityPlatformConfigSmsRegionConfig{
			AllowlistOnly: &GcpIdentityPlatformConfigSmsAllowlistOnly{
				AllowedRegions: []string{"US", "DE"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a default supported IdP with credentials", func() {
		target := minimal()
		target.Spec.DefaultSupportedIdps = []*GcpIdentityPlatformConfigDefaultSupportedIdp{{
			IdpId:        "google.com",
			ClientId:     "1234-abc.apps.googleusercontent.com",
			ClientSecret: "console-issued-secret",
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an OIDC provider with the oidc. name prefix", func() {
		target := minimal()
		target.Spec.OauthIdpConfigs = []*GcpIdentityPlatformConfigOauthIdp{{
			Name:         "oidc.corp-sso",
			Issuer:       "https://accounts.example.com",
			ClientId:     "corp-client",
			ClientSecret: "corp-secret",
			ResponseType: &GcpIdentityPlatformConfigOauthResponseType{Code: true},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a SAML provider with the saml. name prefix", func() {
		target := minimal()
		target.Spec.InboundSamlConfigs = []*GcpIdentityPlatformConfigInboundSaml{{
			Name:        "saml.okta-prod",
			DisplayName: "Okta",
			IdpConfig: &GcpIdentityPlatformConfigSamlIdpConfig{
				IdpEntityId: "http://www.okta.com/exk123",
				SsoUrl:      "https://corp.okta.com/app/sso/saml",
				IdpCertificates: []*GcpIdentityPlatformConfigSamlCertificate{{
					X509Certificate: "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
				}},
			},
			SpConfig: &GcpIdentityPlatformConfigSamlSpConfig{
				CallbackUri: "https://my-project.firebaseapp.com/__/auth/handler",
				SpEntityId:  "my-project-sp",
			},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject an invalid mfa state", func() {
		target := minimal()
		target.Spec.Mfa = &GcpIdentityPlatformConfigMfa{State: "ON"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "MANDATORY")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an unsupported mfa provider", func() {
		target := minimal()
		target.Spec.Mfa = &GcpIdentityPlatformConfigMfa{
			EnabledProviders: []string{"TOTP"},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an out-of-range totp adjacent_intervals", func() {
		target := minimal()
		target.Spec.Mfa = &GcpIdentityPlatformConfigMfa{
			ProviderConfigs: []*GcpIdentityPlatformConfigMfaProviderConfig{{
				TotpProviderConfig: &GcpIdentityPlatformConfigMfaTotp{AdjacentIntervals: 11},
			}},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid blocking-function event type", func() {
		target := minimal()
		target.Spec.BlockingFunctions = &GcpIdentityPlatformConfigBlockingFunctions{
			Triggers: []*GcpIdentityPlatformConfigBlockingFunctionTrigger{{
				EventType:   "afterSignIn",
				FunctionUri: litRef("https://example.com/fn"),
			}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "beforeCreate")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a trigger without a function_uri", func() {
		target := minimal()
		target.Spec.BlockingFunctions = &GcpIdentityPlatformConfigBlockingFunctions{
			Triggers: []*GcpIdentityPlatformConfigBlockingFunctionTrigger{{
				EventType: "beforeCreate",
			}},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject blocking_functions without any trigger", func() {
		target := minimal()
		target.Spec.BlockingFunctions = &GcpIdentityPlatformConfigBlockingFunctions{}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a partial sign-up quota", func() {
		target := minimal()
		target.Spec.SignUpQuota = &GcpIdentityPlatformConfigSignUpQuota{Quota: 100}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "one unit")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a quota above 1000", func() {
		target := minimal()
		target.Spec.SignUpQuota = &GcpIdentityPlatformConfigSignUpQuota{
			Quota:         2000,
			QuotaDuration: "86400s",
			StartTime:     "2026-09-01T00:00:00Z",
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject both SMS region arms together", func() {
		target := minimal()
		target.Spec.SmsRegionConfig = &GcpIdentityPlatformConfigSmsRegionConfig{
			AllowByDefault: &GcpIdentityPlatformConfigSmsAllowByDefault{},
			AllowlistOnly:  &GcpIdentityPlatformConfigSmsAllowlistOnly{},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an SMS region policy with neither arm", func() {
		target := minimal()
		target.Spec.SmsRegionConfig = &GcpIdentityPlatformConfigSmsRegionConfig{}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown default-supported IdP", func() {
		target := minimal()
		target.Spec.DefaultSupportedIdps = []*GcpIdentityPlatformConfigDefaultSupportedIdp{{
			IdpId:        "myidp.example.com",
			ClientId:     "c",
			ClientSecret: "s",
		}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "google.com")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a default-supported IdP without a client secret", func() {
		target := minimal()
		target.Spec.DefaultSupportedIdps = []*GcpIdentityPlatformConfigDefaultSupportedIdp{{
			IdpId:    "google.com",
			ClientId: "c",
		}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an OIDC name without the oidc. prefix", func() {
		target := minimal()
		target.Spec.OauthIdpConfigs = []*GcpIdentityPlatformConfigOauthIdp{{
			Name:     "corp-sso",
			Issuer:   "https://accounts.example.com",
			ClientId: "corp-client",
		}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "oidc.")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a malformed SAML name", func() {
		for _, name := range []string{"okta", "saml.Okta", "saml.x", "saml.okta-"} {
			target := minimal()
			target.Spec.InboundSamlConfigs = []*GcpIdentityPlatformConfigInboundSaml{{
				Name:        name,
				DisplayName: "Okta",
				IdpConfig: &GcpIdentityPlatformConfigSamlIdpConfig{
					IdpEntityId: "e",
					SsoUrl:      "https://sso.example.com",
				},
			}}
			gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("should reject a SAML config without idp_config", func() {
		target := minimal()
		target.Spec.InboundSamlConfigs = []*GcpIdentityPlatformConfigInboundSaml{{
			Name:        "saml.okta-prod",
			DisplayName: "Okta",
		}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a non-https SAML callback_uri", func() {
		target := minimal()
		target.Spec.InboundSamlConfigs = []*GcpIdentityPlatformConfigInboundSaml{{
			Name:        "saml.okta-prod",
			DisplayName: "Okta",
			IdpConfig: &GcpIdentityPlatformConfigSamlIdpConfig{
				IdpEntityId: "e",
				SsoUrl:      "https://sso.example.com",
			},
			SpConfig: &GcpIdentityPlatformConfigSamlSpConfig{
				CallbackUri: "http://insecure.example.com/handler",
			},
		}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "https://")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "ABANDON")).To(gomega.BeTrue())
	})
})
