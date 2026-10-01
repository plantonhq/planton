package auth0emailtemplatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAuth0EmailTemplate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0EmailTemplate Suite")
}

func emailTemplate(spec *Auth0EmailTemplateSpec) *Auth0EmailTemplate {
	return &Auth0EmailTemplate{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0EmailTemplate",
		Metadata:   &shared.CloudResourceMetadata{Name: "verify-email"},
		Spec:       spec,
	}
}

// verifyEmail is the minimal valid spec: the four fields Auth0 requires.
func verifyEmail() *Auth0EmailTemplateSpec {
	return &Auth0EmailTemplateSpec{
		Template: "verify_email",
		From:     "Planton <no-reply@planton.ai>",
		Subject:  "Verify your email for Planton",
		Body:     `<html><body><a href="{{ url }}">Verify {{ user.email }}</a></body></html>`,
	}
}

func strPtr(v string) *string { return &v }
func int32Ptr(v int32) *int32 { return &v }
func boolPtr(v bool) *bool    { return &v }

var _ = ginkgo.Describe("Auth0EmailTemplate Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the minimal template", func() {
			err := protovalidate.Validate(emailTemplate(verifyEmail()))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts every email Auth0 lets a tenant customize", func() {
			for _, template := range []string{
				"verify_email",
				"verify_email_by_code",
				"reset_email",
				"reset_email_by_code",
				"welcome_email",
				"blocked_account",
				"stolen_credentials",
				"enrollment_email",
				"mfa_oob_code",
				"user_invitation",
				"change_password",
				"password_reset",
				"async_approval",
				"auth_email_by_code",
			} {
				spec := verifyEmail()
				spec.Template = template
				err := protovalidate.Validate(emailTemplate(spec))
				gomega.Expect(err).To(gomega.BeNil(), "template %q", template)
			}
		})

		ginkgo.It("accepts every setting", func() {
			spec := verifyEmail()
			spec.Syntax = strPtr("liquid")
			spec.ResultUrl = "https://planton.ai"
			spec.UrlLifetimeInSeconds = int32Ptr(86400)
			spec.Enabled = boolPtr(true)
			spec.IncludeEmailInRedirect = boolPtr(false)
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a disabled template", func() {
			spec := verifyEmail()
			spec.Enabled = boolPtr(false)
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing template", func() {
			spec := verifyEmail()
			spec.Template = ""
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an email Auth0 does not offer", func() {
			spec := verifyEmail()
			spec.Template = "goodbye_email"
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing sender", func() {
			spec := verifyEmail()
			spec.From = ""
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing subject", func() {
			spec := verifyEmail()
			spec.Subject = ""
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing body", func() {
			spec := verifyEmail()
			spec.Body = ""
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a template language Auth0 does not render", func() {
			spec := verifyEmail()
			spec.Syntax = strPtr("handlebars")
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a result URL that is not a URL", func() {
			spec := verifyEmail()
			spec.ResultUrl = "planton welcome page"
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("result_url is an http(s) URL or a Liquid template")))
		})

		ginkgo.It("accepts a result URL written as a Liquid template", func() {
			spec := verifyEmail()
			spec.ResultUrl = "{{ application.callback_domain }}/welcome"
			gomega.Expect(protovalidate.Validate(emailTemplate(spec))).To(gomega.BeNil())
		})

		ginkgo.It("rejects a link lifetime of zero", func() {
			spec := verifyEmail()
			spec.UrlLifetimeInSeconds = int32Ptr(0)
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a negative link lifetime", func() {
			spec := verifyEmail()
			spec.UrlLifetimeInSeconds = int32Ptr(-60)
			err := protovalidate.Validate(emailTemplate(spec))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
