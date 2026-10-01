package auth0promptscreenpartialsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAuth0PromptScreenPartials(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0PromptScreenPartials Suite")
}

func screenPartials(spec *Auth0PromptScreenPartialsSpec) *Auth0PromptScreenPartials {
	return &Auth0PromptScreenPartials{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0PromptScreenPartials",
		Metadata:   &shared.CloudResourceMetadata{Name: "signup-partials"},
		Spec:       spec,
	}
}

// terms is one screen with one fragment, the smallest partial a spec declares.
func terms(screen string) *Auth0PromptScreenPartial {
	return &Auth0PromptScreenPartial{
		ScreenName: screen,
		InsertionPoints: &Auth0PromptInsertionPoints{
			FormContentEnd: `<div class="terms">By signing up you accept the terms.</div>`,
		},
	}
}

// promptTypes are the prompts Auth0 accepts partials for.
var promptTypes = []string{
	"login-id", "login", "login-password", "signup", "signup-id", "signup-password",
	"login-passwordless", "customized-consent", "passkeys", "confirmation",
}

var _ = ginkgo.Describe("Auth0PromptScreenPartials Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts one screen with one fragment", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType:     "signup",
				ScreenPartials: []*Auth0PromptScreenPartial{terms("signup")},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts several screens of one prompt, each once", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType: "login-passwordless",
				ScreenPartials: []*Auth0PromptScreenPartial{
					terms("login-passwordless-sms-otp"),
					terms("login-passwordless-email-code"),
				},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts every insertion point on one screen", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType: "customized-consent",
				ScreenPartials: []*Auth0PromptScreenPartial{{
					ScreenName: "customized-consent",
					InsertionPoints: &Auth0PromptInsertionPoints{
						FormContent:           "<div>form</div>",
						FormContentStart:      "<div>start</div>",
						FormContentEnd:        "<div>end</div>",
						FormFooterStart:       "<div>footer start</div>",
						FormFooterEnd:         "<div>footer end</div>",
						SecondaryActionsStart: "<div>actions start</div>",
						SecondaryActionsEnd:   "<div>actions end</div>",
					},
				}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts each insertion point set alone", func() {
			for name, points := range map[string]*Auth0PromptInsertionPoints{
				"form_content":            {FormContent: "<div/>"},
				"form_content_start":      {FormContentStart: "<div/>"},
				"form_content_end":        {FormContentEnd: "<div/>"},
				"form_footer_start":       {FormFooterStart: "<div/>"},
				"form_footer_end":         {FormFooterEnd: "<div/>"},
				"secondary_actions_start": {SecondaryActionsStart: "<div/>"},
				"secondary_actions_end":   {SecondaryActionsEnd: "<div/>"},
			} {
				err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
					PromptType:     "login",
					ScreenPartials: []*Auth0PromptScreenPartial{{ScreenName: "login", InsertionPoints: points}},
				}))
				gomega.Expect(err).To(gomega.BeNil(), "insertion point %s", name)
			}
		})

		ginkgo.It("accepts every prompt Auth0 takes partials for", func() {
			for _, promptType := range promptTypes {
				err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
					PromptType:     promptType,
					ScreenPartials: []*Auth0PromptScreenPartial{terms(promptType)},
				}))
				gomega.Expect(err).To(gomega.BeNil(), "prompt %s", promptType)
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing prompt type", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				ScreenPartials: []*Auth0PromptScreenPartial{terms("signup")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a prompt Auth0 takes no partials for", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType:     "reset-password",
				ScreenPartials: []*Auth0PromptScreenPartial{terms("reset-password")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a spec with no screens", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{PromptType: "signup"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a screen listed twice", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType:     "signup",
				ScreenPartials: []*Auth0PromptScreenPartial{terms("signup"), terms("signup")},
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("each screen appears once in screen_partials")))
		})

		ginkgo.It("rejects a screen entry with no fragments", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType: "signup",
				ScreenPartials: []*Auth0PromptScreenPartial{{
					ScreenName:      "signup",
					InsertionPoints: &Auth0PromptInsertionPoints{},
				}},
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("set at least one insertion point")))
		})

		ginkgo.It("rejects a screen entry without insertion points", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType:     "signup",
				ScreenPartials: []*Auth0PromptScreenPartial{{ScreenName: "signup"}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a screen entry without a screen name", func() {
			err := protovalidate.Validate(screenPartials(&Auth0PromptScreenPartialsSpec{
				PromptType:     "signup",
				ScreenPartials: []*Auth0PromptScreenPartial{terms("")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
