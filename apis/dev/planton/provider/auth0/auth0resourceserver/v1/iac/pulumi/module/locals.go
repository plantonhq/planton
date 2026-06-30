package module

import (
	auth0resourceserverv1 "github.com/plantonhq/planton/apis/dev/planton/provider/auth0/auth0resourceserver/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values for the Auth0 Resource Server deployment
type Locals struct {
	Auth0ResourceServer *auth0resourceserverv1.Auth0ResourceServer

	// Core configuration
	ResourceName string
	Identifier   string
	Name         string

	// Token settings
	SigningAlg          string
	AllowOfflineAccess  bool
	TokenLifetime       int32
	TokenLifetimeForWeb int32

	// Access control settings
	SkipConsentForVerifiableFirstPartyClients bool
	EnforcePolicies                           bool
	TokenDialect                              string

	// Scopes
	Scopes []*auth0resourceserverv1.Auth0ResourceServerScope
}

// initializeLocals creates and populates the Locals struct from stack input
func initializeLocals(ctx *pulumi.Context, stackInput *auth0resourceserverv1.Auth0ResourceServerStackInput) *Locals {
	locals := &Locals{}

	// Store the target resource
	locals.Auth0ResourceServer = stackInput.Target

	spec := stackInput.Target.Spec
	metadata := stackInput.Target.Metadata

	// Core configuration
	locals.ResourceName = metadata.Name
	locals.Identifier = spec.Identifier

	// Use spec.Name if provided, otherwise use metadata.Name
	if spec.Name != "" {
		locals.Name = spec.Name
	} else {
		locals.Name = metadata.Name
	}

	// Token settings
	locals.SigningAlg = spec.SigningAlg
	locals.AllowOfflineAccess = spec.AllowOfflineAccess
	locals.TokenLifetime = spec.TokenLifetime
	locals.TokenLifetimeForWeb = spec.TokenLifetimeForWeb

	// Access control settings
	locals.SkipConsentForVerifiableFirstPartyClients = spec.SkipConsentForVerifiableFirstPartyClients
	locals.EnforcePolicies = spec.EnforcePolicies
	locals.TokenDialect = spec.TokenDialect

	// Scopes
	locals.Scopes = spec.Scopes

	return locals
}
