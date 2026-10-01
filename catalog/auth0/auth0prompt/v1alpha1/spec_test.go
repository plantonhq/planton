package auth0promptv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestAuth0Prompt(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0Prompt Suite")
}

func prompt(spec *Auth0PromptSpec) *Auth0Prompt {
	return &Auth0Prompt{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0Prompt",
		Metadata:   &shared.CloudResourceMetadata{Name: "prompt"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("Auth0Prompt Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts Universal Login with identifier-first login", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{
				UniversalLoginExperience: "new",
				IdentifierFirst:          proto.Bool(true),
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts every setting", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{
				UniversalLoginExperience:    "new",
				IdentifierFirst:             proto.Bool(true),
				WebauthnPlatformFirstFactor: proto.Bool(true),
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the classic experience alone", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{UniversalLoginExperience: "classic"}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts identifier-first turned off, leaving the experience unmanaged", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{IdentifierFirst: proto.Bool(false)}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the device-biometrics first factor turned off alone", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{WebauthnPlatformFirstFactor: proto.Bool(false)}))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a spec that manages nothing", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("configure at least one of universal_login_experience, identifier_first or webauthn_platform_first_factor")))
		})

		ginkgo.It("refuses the device-biometrics first factor without identifier-first", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{WebauthnPlatformFirstFactor: proto.Bool(true)}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("webauthn_platform_first_factor needs identifier_first: true")))
		})

		ginkgo.It("refuses the device-biometrics first factor with identifier-first turned off", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{
				IdentifierFirst:             proto.Bool(false),
				WebauthnPlatformFirstFactor: proto.Bool(true),
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("webauthn_platform_first_factor needs identifier_first: true")))
		})

		ginkgo.It("refuses an unknown login experience", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{UniversalLoginExperience: "universal"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a login experience in the wrong case", func() {
			err := protovalidate.Validate(prompt(&Auth0PromptSpec{UniversalLoginExperience: "New"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
