package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createEventStream creates an Auth0 event stream based on the configuration
func createEventStream(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.EventStream, error) {
	// Build subscriptions array
	subscriptionsArray := pulumi.StringArray{}
	for _, subscription := range locals.Subscriptions {
		subscriptionsArray = append(subscriptionsArray, pulumi.String(subscription))
	}

	// Build event stream arguments
	eventStreamArgs := &auth0.EventStreamArgs{
		Name:            pulumi.String(locals.StreamName),
		DestinationType: pulumi.String(locals.DestinationType),
		Subscriptions:   subscriptionsArray,
	}

	// Configure destination based on type
	switch locals.DestinationType {
	case "eventbridge":
		if locals.EventBridgeConfiguration != nil {
			eventStreamArgs.EventbridgeConfiguration = &auth0.EventStreamEventbridgeConfigurationArgs{
				AwsAccountId: pulumi.String(locals.EventBridgeConfiguration.AwsAccountId),
				AwsRegion:    pulumi.String(locals.EventBridgeConfiguration.AwsRegion),
			}
		}

	case "webhook":
		if locals.WebhookConfiguration != nil {
			webhookConfig := &auth0.EventStreamWebhookConfigurationArgs{
				WebhookEndpoint: pulumi.String(locals.WebhookConfiguration.WebhookEndpoint),
			}

			// Configure authorization
			if locals.WebhookConfiguration.WebhookAuthorization != nil {
				auth := locals.WebhookConfiguration.WebhookAuthorization
				webhookAuthArgs := &auth0.EventStreamWebhookConfigurationWebhookAuthorizationArgs{
					Method: pulumi.String(auth.Method),
				}

				switch auth.Method {
				case "basic":
					if auth.Username != "" {
						webhookAuthArgs.Username = pulumi.String(auth.Username)
					}
					if auth.Password != "" {
						webhookAuthArgs.Password = pulumi.String(auth.Password)
					}
				case "bearer":
					if auth.Token != "" {
						webhookAuthArgs.Token = pulumi.String(auth.Token)
					}
				}

				webhookConfig.WebhookAuthorization = webhookAuthArgs
			}

			eventStreamArgs.WebhookConfiguration = webhookConfig
		}
	}

	// Create the event stream resource
	eventStream, err := auth0.NewEventStream(ctx, locals.StreamName, eventStreamArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 event stream %s", locals.StreamName)
	}

	return eventStream, nil
}
