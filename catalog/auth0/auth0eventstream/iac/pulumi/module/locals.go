package module

import (
	auth0eventstreamv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0eventstream/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values for the Auth0 Event Stream deployment
type Locals struct {
	Auth0EventStream *auth0eventstreamv1alpha1.Auth0EventStream

	// Core configuration
	StreamName      string
	DestinationType string
	Subscriptions   []string

	// EventBridge configuration
	EventBridgeConfiguration *auth0eventstreamv1alpha1.Auth0EventBridgeConfiguration

	// Webhook configuration
	WebhookConfiguration *auth0eventstreamv1alpha1.Auth0WebhookConfiguration
}

// initializeLocals creates and populates the Locals struct from IaC input
func initializeLocals(ctx *pulumi.Context, iacInput *auth0eventstreamv1alpha1.Auth0EventStreamIacInput) *Locals {
	locals := &Locals{}

	// Store the target resource
	locals.Auth0EventStream = iacInput.Target

	spec := iacInput.Target.Spec
	metadata := iacInput.Target.Metadata

	// Core configuration
	locals.StreamName = metadata.Name
	locals.DestinationType = spec.DestinationType
	locals.Subscriptions = spec.Subscriptions

	// EventBridge configuration
	locals.EventBridgeConfiguration = spec.EventbridgeConfiguration

	// Webhook configuration
	locals.WebhookConfiguration = spec.WebhookConfiguration

	return locals
}
