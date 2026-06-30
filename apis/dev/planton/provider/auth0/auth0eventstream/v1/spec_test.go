package auth0eventstreamv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAuth0EventStream(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0EventStream Suite")
}

var _ = ginkgo.Describe("Auth0EventStream Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("auth0_event_stream with eventbridge destination", func() {
			var input *Auth0EventStream

			ginkgo.BeforeEach(func() {
				input = &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "security-events",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions: []string{
							"user.created",
							"user.updated",
							"authentication.success",
							"authentication.failure",
						},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "123456789012",
							AwsRegion:    "us-east-1",
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_event_stream with eventbridge in different regions", func() {
			ginkgo.It("should not return a validation error for eu-west-1", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "eu-events",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions: []string{
							"user.created",
						},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "987654321098",
							AwsRegion:    "eu-west-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for ap-southeast-1", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "apac-events",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions: []string{
							"authentication.success",
						},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "111222333444",
							AwsRegion:    "ap-southeast-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_event_stream with webhook destination using bearer token", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "user-events-webhook",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions: []string{
							"user.created",
							"user.updated",
							"user.deleted",
						},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://api.example.com/webhooks/auth0",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "bearer",
								Token:  "secret-token-12345",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_event_stream with webhook destination using basic auth", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "basic-auth-webhook",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions: []string{
							"authentication.success",
							"authentication.failure",
						},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://siem.example.com/auth0/events",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method:   "basic",
								Username: "webhook-user",
								Password: "super-secret-password",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_event_stream with many subscriptions", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "all-user-events",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions: []string{
							"user.created",
							"user.updated",
							"user.deleted",
							"user.blocked",
							"user.unblocked",
							"authentication.success",
							"authentication.failure",
							"api.authorization.success",
							"api.authorization.failure",
						},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "123456789012",
							AwsRegion:    "us-west-2",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_event_stream with single subscription", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "login-only",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions: []string{
							"authentication.success",
						},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://analytics.example.com/logins",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "bearer",
								Token:  "analytics-token",
							},
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
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata:   nil,
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions:   []string{"user.created"},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "123456789012",
							AwsRegion:    "us-east-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "bearer",
								Token:  "test-token",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "bearer",
								Token:  "test-token",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required destination_type", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "",
						Subscriptions:   []string{"user.created"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid destination_type value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "kafka",
						Subscriptions:   []string{"user.created"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("empty subscriptions list", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions:   []string{},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "123456789012",
							AwsRegion:    "us-east-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing aws_account_id in eventbridge config", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions:   []string{"user.created"},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "",
							AwsRegion:    "us-east-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid aws_account_id format (not 12 digits)", func() {
			ginkgo.It("should return a validation error for short account id", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions:   []string{"user.created"},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "12345",
							AwsRegion:    "us-east-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for account id with letters", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions:   []string{"user.created"},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "12345678901a",
							AwsRegion:    "us-east-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing aws_region in eventbridge config", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "eventbridge",
						Subscriptions:   []string{"user.created"},
						EventbridgeConfiguration: &Auth0EventBridgeConfiguration{
							AwsAccountId: "123456789012",
							AwsRegion:    "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing webhook_endpoint in webhook config", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "bearer",
								Token:  "test-token",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("non-HTTPS webhook_endpoint", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "http://insecure.example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "bearer",
								Token:  "test-token",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing webhook_authorization in webhook config", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint:      "https://example.com/webhook",
							WebhookAuthorization: nil,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing authorization method", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "",
								Token:  "test-token",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid authorization method", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "oauth2",
								Token:  "test-token",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("bearer method without token", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method: "bearer",
								Token:  "", // missing token
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("basic method without username", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method:   "basic",
								Username: "", // missing username
								Password: "secret-password",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("basic method without password", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0EventStream{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0EventStream",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-stream",
					},
					Spec: &Auth0EventStreamSpec{
						DestinationType: "webhook",
						Subscriptions:   []string{"user.created"},
						WebhookConfiguration: &Auth0WebhookConfiguration{
							WebhookEndpoint: "https://example.com/webhook",
							WebhookAuthorization: &Auth0WebhookAuthorization{
								Method:   "basic",
								Username: "webhook-user",
								Password: "", // missing password
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
