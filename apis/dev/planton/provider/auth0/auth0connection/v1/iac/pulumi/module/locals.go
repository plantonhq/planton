package module

import (
	auth0connectionv1 "github.com/plantonhq/planton/apis/dev/planton/provider/auth0/auth0connection/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values for the Auth0 Connection deployment
type Locals struct {
	Auth0Connection *auth0connectionv1.Auth0Connection

	// Core connection configuration
	ConnectionName     string
	Strategy           string
	DisplayName        string
	IsDomainConnection bool
	ShowAsButton       bool

	// Client configuration
	EnabledClients []string
	Realms         []string
	Metadata       map[string]string

	// Database options (when strategy is auth0)
	DatabaseOptions *auth0connectionv1.Auth0DatabaseOptions

	// Social options (for social strategies)
	SocialOptions *auth0connectionv1.Auth0SocialOptions

	// SAML options (when strategy is samlp)
	SamlOptions *auth0connectionv1.Auth0SamlOptions

	// OIDC options (when strategy is oidc)
	OidcOptions *auth0connectionv1.Auth0OidcOptions

	// Azure AD options (when strategy is waad)
	AzureAdOptions *auth0connectionv1.Auth0AzureAdOptions
}

// initializeLocals creates and populates the Locals struct from stack input
func initializeLocals(ctx *pulumi.Context, stackInput *auth0connectionv1.Auth0ConnectionStackInput) *Locals {
	locals := &Locals{}

	// Store the target resource
	locals.Auth0Connection = stackInput.Target

	spec := stackInput.Target.Spec
	metadata := stackInput.Target.Metadata

	// Core configuration
	locals.ConnectionName = metadata.Name
	locals.Strategy = spec.Strategy

	// Display name - default to metadata name if not specified
	if spec.DisplayName != "" {
		locals.DisplayName = spec.DisplayName
	} else {
		locals.DisplayName = metadata.Name
	}

	locals.IsDomainConnection = spec.IsDomainConnection
	locals.ShowAsButton = spec.ShowAsButton

	// Client and realm configuration - extract values from StringValueOrRef
	for _, client := range spec.EnabledClients {
		if client != nil && client.GetValue() != "" {
			locals.EnabledClients = append(locals.EnabledClients, client.GetValue())
		}
	}
	locals.Realms = spec.Realms
	locals.Metadata = spec.Metadata

	// Strategy-specific options
	locals.DatabaseOptions = spec.DatabaseOptions
	locals.SocialOptions = spec.SocialOptions
	locals.SamlOptions = spec.SamlOptions
	locals.OidcOptions = spec.OidcOptions
	locals.AzureAdOptions = spec.AzureAdOptions

	return locals
}
