package auth0promptcustomtextv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAuth0PromptCustomText(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0PromptCustomText Suite")
}

func customText(spec *Auth0PromptCustomTextSpec) *Auth0PromptCustomText {
	return &Auth0PromptCustomText{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0PromptCustomText",
		Metadata:   &shared.CatalogObjectMetadata{Name: "login-en"},
		Spec:       spec,
	}
}

// loginWords is one screen with one text, the smallest words a spec declares.
func loginWords() map[string]*Auth0PromptScreenText {
	return map[string]*Auth0PromptScreenText{
		"login": {Texts: map[string]string{"title": "Welcome back"}},
	}
}

// prompts are the prompts Auth0 stores custom text for.
var prompts = []string{
	"login", "login-id", "login-password", "login-passwordless", "login-email-verification",
	"signup", "signup-id", "signup-password",
	"phone-identifier-enrollment", "phone-identifier-challenge", "email-identifier-challenge",
	"reset-password", "custom-form", "consent", "customized-consent", "logout",
	"mfa-push", "mfa-otp", "mfa-voice", "mfa-phone", "mfa-webauthn", "mfa-sms", "mfa-email", "mfa-recovery-code", "mfa",
	"status", "device-flow", "email-verification", "email-otp-challenge",
	"organizations", "invitation", "common", "passkeys", "captcha", "brute-force-protection", "confirmation",
}

var _ = ginkgo.Describe("Auth0PromptCustomText Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts one screen's words in English", func() {
			err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
				Prompt:   "login",
				Language: "en",
				Screens:  loginWords(),
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts several screens of one prompt, with Universal Login's variables", func() {
			err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
				Prompt:   "reset-password",
				Language: "fr-CA",
				Screens: map[string]*Auth0PromptScreenText{
					"reset-password-request": {Texts: map[string]string{"title": "Mot de passe oublié?"}},
					"reset-password-success": {Texts: map[string]string{"buttonText": "Retour à ${clientName}"}},
				},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts every prompt Auth0 stores custom text for", func() {
			for _, p := range prompts {
				err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
					Prompt:   p,
					Language: "en",
					Screens:  loginWords(),
				}))
				gomega.Expect(err).To(gomega.BeNil(), "prompt %s", p)
			}
		})

		ginkgo.It("accepts the language codes Auth0 names", func() {
			for _, language := range []string{"en", "fr-CA", "pt-BR", "zh-TW", "es-419", "cnr", "zgh"} {
				err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
					Prompt:   "login",
					Language: language,
					Screens:  loginWords(),
				}))
				gomega.Expect(err).To(gomega.BeNil(), "language %s", language)
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing prompt", func() {
			err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
				Language: "en",
				Screens:  loginWords(),
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a prompt Auth0 has no custom text for", func() {
			err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
				Prompt:   "sign-in",
				Language: "en",
				Screens:  loginWords(),
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing language", func() {
			err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
				Prompt:  "login",
				Screens: loginWords(),
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a language that is not a code as Auth0 names it", func() {
			for _, language := range []string{"English", "en_US", "fr-ca", "EN", "en-"} {
				err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
					Prompt:   "login",
					Language: language,
					Screens:  loginWords(),
				}))
				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("language is a language code as Auth0 names it")), "language %s", language)
			}
		})

		ginkgo.It("rejects a spec with no screens", func() {
			err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
				Prompt:   "login",
				Language: "en",
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a screen with no texts", func() {
			err := protovalidate.Validate(customText(&Auth0PromptCustomTextSpec{
				Prompt:   "login",
				Language: "en",
				Screens:  map[string]*Auth0PromptScreenText{"login": {}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
