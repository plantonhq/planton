package auth0actionv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAuth0Action(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0Action Suite")
}

var _ = ginkgo.Describe("Auth0Action Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal post-login action with deploy", func() {
			var input *Auth0Action

			ginkgo.BeforeEach(func() {
				input = &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "enrich-token-claims",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code: `exports.onExecutePostLogin = async (event, api) => {
  api.idToken.setCustomClaim('https://myapp/roles', event.authorization?.roles || []);
};`,
						Deploy: true,
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("action with dependencies and secrets", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "slack-login-alert",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code: `exports.onExecutePostLogin = async (event, api) => {
  const axios = require('axios');
  await axios.post(event.secrets.SLACK_WEBHOOK, {
    text: 'User ' + event.user.email + ' logged in'
  });
};`,
						Runtime: "node22",
						Deploy:  true,
						Dependencies: []*Auth0ActionDependency{
							{Name: "axios", Version: "1.6.0"},
						},
						Secrets: []*Auth0ActionSecret{
							{Name: "SLACK_WEBHOOK", Value: "https://hooks.slack.com/services/T00/B00/xxx"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("action with trigger binding", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "validate-email-domain",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "pre-user-registration",
							Version: "v2",
						},
						Code: `exports.onExecutePreUserRegistration = async (event, api) => {
  const domain = event.user.email.split('@')[1];
  if (domain !== 'example.com') {
    api.access.deny('registration_denied', 'Domain not allowed');
  }
};`,
						Deploy: true,
						TriggerBinding: &Auth0ActionTriggerBinding{
							DisplayName: "Validate Email Domain",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("action with trigger binding using default display name", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "audit-m2m-exchange",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "credentials-exchange",
							Version: "v2",
						},
						Code: `exports.onExecuteCredentialsExchange = async (event, api) => {
  console.log('M2M exchange for client: ' + event.client.name);
};`,
						Deploy:         true,
						TriggerBinding: &Auth0ActionTriggerBinding{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("action with node18 runtime", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "legacy-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:    "exports.onExecutePostLogin = async (event, api) => {};",
						Runtime: "node18",
						Deploy:  true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("action without deploy (staging only)", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "draft-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: false,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("custom-token-exchange trigger", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "token-exchange-handler",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "custom-token-exchange",
							Version: "v1",
						},
						Code:   "exports.onExecuteCustomTokenExchange = async (event, api) => {};",
						Deploy: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("send-phone-message trigger", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-sms-provider",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "send-phone-message",
							Version: "v2",
						},
						Code: `exports.onExecuteSendPhoneMessage = async (event, api) => {
  const twilio = require('twilio');
  const client = twilio(event.secrets.TWILIO_SID, event.secrets.TWILIO_TOKEN);
  await client.messages.create({
    body: event.message_options.text,
    to: event.message_options.recipient,
    from: event.secrets.TWILIO_FROM,
  });
};`,
						Deploy: true,
						Dependencies: []*Auth0ActionDependency{
							{Name: "twilio", Version: "4.23.0"},
						},
						Secrets: []*Auth0ActionSecret{
							{Name: "TWILIO_SID", Value: "AC1234567890abcdef"},
							{Name: "TWILIO_TOKEN", Value: "auth-token-value"},
							{Name: "TWILIO_FROM", Value: "+15551234567"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata:   nil,
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required code", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "",
						Deploy: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required supported_trigger", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: nil,
						Code:             "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy:           true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid trigger id", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "invalid-trigger",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing trigger version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid runtime value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:    "exports.onExecutePostLogin = async (event, api) => {};",
						Runtime: "node12",
						Deploy:  true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("trigger_binding with deploy false", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: false,
						TriggerBinding: &Auth0ActionTriggerBinding{
							DisplayName: "Should Fail",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("dependency with empty name", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
						Dependencies: []*Auth0ActionDependency{
							{Name: "", Version: "1.0.0"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("dependency with empty version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
						Dependencies: []*Auth0ActionDependency{
							{Name: "axios", Version: ""},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("secret with empty name", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
						Secrets: []*Auth0ActionSecret{
							{Name: "", Value: "some-value"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("secret with empty value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Action{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Action",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-action",
					},
					Spec: &Auth0ActionSpec{
						SupportedTrigger: &Auth0ActionSupportedTrigger{
							Id:      "post-login",
							Version: "v3",
						},
						Code:   "exports.onExecutePostLogin = async (event, api) => {};",
						Deploy: true,
						Secrets: []*Auth0ActionSecret{
							{Name: "API_KEY", Value: ""},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
