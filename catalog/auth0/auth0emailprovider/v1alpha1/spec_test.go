package auth0emailproviderv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAuth0EmailProvider(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0EmailProvider Suite")
}

func emailProvider(spec *Auth0EmailProviderSpec) *Auth0EmailProvider {
	return &Auth0EmailProvider{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0EmailProvider",
		Metadata:   &shared.CatalogObjectMetadata{Name: "email-provider"},
		Spec:       spec,
	}
}

// resend is the smtp arm the Resend Over SMTP preset sends through.
func resend() *Auth0EmailProviderSmtp {
	return &Auth0EmailProviderSmtp{
		Host:     "smtp.resend.com",
		Port:     587,
		User:     "resend",
		Password: "$secret/resend-api-key",
	}
}

func withSmtp(smtp *Auth0EmailProviderSmtp) *Auth0EmailProviderSpec {
	return &Auth0EmailProviderSpec{
		DefaultFromAddress: "Acme <no-reply@acme.com>",
		Service:            &Auth0EmailProviderSpec_Smtp{Smtp: smtp},
	}
}

func boolPtr(v bool) *bool { return &v }

var _ = ginkgo.Describe("Auth0EmailProvider Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts Resend over SMTP", func() {
			err := protovalidate.Validate(emailProvider(withSmtp(resend())))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a disabled provider, keeping its configuration", func() {
			spec := withSmtp(resend())
			spec.Enabled = boolPtr(false)
			err := protovalidate.Validate(emailProvider(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts every allowed X-MC-ViewContentLink header value", func() {
			for _, value := range []string{"", "true", "false"} {
				smtp := resend()
				smtp.Headers = &Auth0EmailProviderSmtpHeaders{XMcViewContentLink: value, XSesConfigurationSet: "auth0"}
				err := protovalidate.Validate(emailProvider(withSmtp(smtp)))
				gomega.Expect(err).To(gomega.BeNil(), "x_mc_view_content_link %q", value)
			}
		})

		ginkgo.It("accepts Amazon SES with a configuration set", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Ses{Ses: &Auth0EmailProviderSes{
					AccessKeyId:          "$secret/ses-access-key-id",
					SecretAccessKey:      "$secret/ses-secret-access-key",
					Region:               "eu-west-1",
					ConfigurationSetName: "auth0-transactional",
				}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts SendGrid", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_Sendgrid{Sendgrid: &Auth0EmailProviderSendgrid{ApiKey: "$secret/sendgrid-api-key"}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts SparkPost EU", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_Sparkpost{Sparkpost: &Auth0EmailProviderSparkpost{ApiKey: "$secret/sparkpost-api-key", Region: "eu"}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Mailgun with its sending domain", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Mailgun{Mailgun: &Auth0EmailProviderMailgun{
					ApiKey: "$secret/mailgun-api-key",
					Domain: "mg.acme.com",
				}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Mandrill with its view-content link", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Mandrill{Mandrill: &Auth0EmailProviderMandrill{
					ApiKey:          "$secret/mandrill-api-key",
					ViewContentLink: boolPtr(true),
				}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Azure Communication Services", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_AzureCs{AzureCs: &Auth0EmailProviderAzureCs{ConnectionString: "$secret/acs-connection-string"}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Microsoft 365 with an app registration", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Ms365{Ms365: &Auth0EmailProviderMs365{
					TenantId:     "$secret/ms365-tenant-id",
					ClientId:     "$secret/ms365-client-id",
					ClientSecret: "$secret/ms365-client-secret",
				}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom provider with no settings", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_Custom{Custom: &Auth0EmailProviderCustom{}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing sender", func() {
			spec := withSmtp(resend())
			spec.DefaultFromAddress = ""
			err := protovalidate.Validate(emailProvider(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a provider with no service", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{DefaultFromAddress: "Acme <no-reply@acme.com>"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts an SMTP host given as an IP address", func() {
			smtp := resend()
			smtp.Host = "192.0.2.25"
			gomega.Expect(protovalidate.Validate(emailProvider(withSmtp(smtp)))).To(gomega.BeNil())
		})

		ginkgo.It("rejects an SMTP host that is not a host name", func() {
			smtp := resend()
			smtp.Host = "smtp://smtp.resend.com"
			err := protovalidate.Validate(emailProvider(withSmtp(smtp)))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing SMTP port", func() {
			smtp := resend()
			smtp.Port = 0
			err := protovalidate.Validate(emailProvider(withSmtp(smtp)))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an SMTP port above 65535", func() {
			smtp := resend()
			smtp.Port = 65536
			err := protovalidate.Validate(emailProvider(withSmtp(smtp)))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing SMTP user", func() {
			smtp := resend()
			smtp.User = ""
			err := protovalidate.Validate(emailProvider(withSmtp(smtp)))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing SMTP password", func() {
			smtp := resend()
			smtp.Password = ""
			err := protovalidate.Validate(emailProvider(withSmtp(smtp)))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an X-MC-ViewContentLink header that is not true or false", func() {
			smtp := resend()
			smtp.Headers = &Auth0EmailProviderSmtpHeaders{XMcViewContentLink: "yes"}
			err := protovalidate.Validate(emailProvider(withSmtp(smtp)))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects SES without a region", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Ses{Ses: &Auth0EmailProviderSes{
					AccessKeyId:     "$secret/ses-access-key-id",
					SecretAccessKey: "$secret/ses-secret-access-key",
				}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects SES without its secret access key", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Ses{Ses: &Auth0EmailProviderSes{
					AccessKeyId: "$secret/ses-access-key-id",
					Region:      "eu-west-1",
				}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects SendGrid without an API key", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_Sendgrid{Sendgrid: &Auth0EmailProviderSendgrid{}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects SparkPost without an API key", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_Sparkpost{Sparkpost: &Auth0EmailProviderSparkpost{Region: "eu"}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a Mailgun domain shorter than four characters", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Mailgun{Mailgun: &Auth0EmailProviderMailgun{
					ApiKey: "$secret/mailgun-api-key",
					Domain: "a.b",
				}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects Mandrill without an API key", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_Mandrill{Mandrill: &Auth0EmailProviderMandrill{ViewContentLink: boolPtr(true)}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects Azure Communication Services without a connection string", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service:            &Auth0EmailProviderSpec_AzureCs{AzureCs: &Auth0EmailProviderAzureCs{}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects Microsoft 365 without a client secret", func() {
			err := protovalidate.Validate(emailProvider(&Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Service: &Auth0EmailProviderSpec_Ms365{Ms365: &Auth0EmailProviderMs365{
					TenantId: "$secret/ms365-tenant-id",
					ClientId: "$secret/ms365-client-id",
				}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
