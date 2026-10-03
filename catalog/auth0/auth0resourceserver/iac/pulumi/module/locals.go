package module

import (
	auth0resourceserverv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0resourceserver/v1alpha1"
)

// Locals holds the values the module computes from the IaC input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	Auth0ResourceServer *auth0resourceserverv1alpha1.Auth0ResourceServer

	// Core configuration
	ResourceName string
	Identifier   string
	Name         string

	// Token settings
	SigningAlg          string
	AllowOfflineAccess  *bool
	TokenLifetime       int32
	TokenLifetimeForWeb int32

	// Access control settings
	SkipConsentForVerifiableFirstPartyClients *bool
	EnforcePolicies                           *bool
	TokenDialect                              string

	// Scopes
	Scopes []*auth0resourceserverv1alpha1.Auth0ResourceServerScope

	// The settings below are unmanaged when unset: a nil pointer, nil block or
	// empty list is never sent, so the API keeps whatever Auth0 holds.
	AllowOnlineAccess                      *bool
	AllowOnlineAccessWithEphemeralSessions *bool
	ConsentPolicy                          *string
	TokenLifetimeForAnonymousAccessTokens  *int
	VerificationLocation                   *string
	SigningSecret                          *string

	AccessToken              *auth0resourceserverv1alpha1.Auth0ResourceServerAccessToken
	AuthorizationDetails     []*auth0resourceserverv1alpha1.Auth0ResourceServerAuthorizationDetail
	AuthorizationPolicy      *auth0resourceserverv1alpha1.Auth0ResourceServerAuthorizationPolicy
	ProofOfPossession        *auth0resourceserverv1alpha1.Auth0ResourceServerProofOfPossession
	SubjectTypeAuthorization *auth0resourceserverv1alpha1.Auth0ResourceServerSubjectTypeAuthorization
	TokenEncryption          *auth0resourceserverv1alpha1.Auth0ResourceServerTokenEncryption

	// DefaultGrants are the grants every third-party application gets on the
	// API, one per subject type (the key the grants are declared under).
	DefaultGrants []*auth0resourceserverv1alpha1.Auth0ResourceServerThirdPartyClientDefaultGrant
}

// initializeLocals creates and populates the Locals struct from IaC input
func initializeLocals(iacInput *auth0resourceserverv1alpha1.Auth0ResourceServerIacInput) *Locals {
	locals := &Locals{}

	// Store the target resource
	locals.Auth0ResourceServer = iacInput.Target

	spec := iacInput.Target.Spec
	metadata := iacInput.Target.Metadata

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

	// Unmanaged when unset. The proto's optional fields carry presence, so a
	// declared false, zero or empty value is sent as declared.
	locals.AllowOnlineAccess = spec.AllowOnlineAccess
	locals.AllowOnlineAccessWithEphemeralSessions = spec.AllowOnlineAccessWithEphemeralSessions
	locals.ConsentPolicy = spec.ConsentPolicy
	if spec.TokenLifetimeForAnonymousAccessTokens != nil {
		lifetime := int(spec.GetTokenLifetimeForAnonymousAccessTokens())
		locals.TokenLifetimeForAnonymousAccessTokens = &lifetime
	}
	locals.VerificationLocation = spec.VerificationLocation
	locals.SigningSecret = spec.SigningSecret

	locals.AccessToken = spec.AccessToken
	locals.AuthorizationDetails = spec.AuthorizationDetails
	locals.AuthorizationPolicy = spec.AuthorizationPolicy
	locals.ProofOfPossession = spec.ProofOfPossession
	locals.SubjectTypeAuthorization = spec.SubjectTypeAuthorization
	locals.TokenEncryption = spec.TokenEncryption

	locals.DefaultGrants = spec.ThirdPartyClientDefaultGrants

	return locals
}
