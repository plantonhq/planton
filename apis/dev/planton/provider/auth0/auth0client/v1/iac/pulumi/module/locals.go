package module

import (
	auth0clientv1 "github.com/plantonhq/planton/apis/dev/planton/provider/auth0/auth0client/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values for the Auth0 Client deployment
type Locals struct {
	Auth0Client *auth0clientv1.Auth0Client

	// Core client configuration
	ClientName      string
	ApplicationType string
	Description     string
	LogoUri         string

	// URLs configuration
	Callbacks         []string
	AllowedLogoutUrls []string
	WebOrigins        []string
	AllowedOrigins    []string

	// OAuth configuration
	GrantTypes     []string
	OidcConformant bool
	IsFirstParty   bool

	// Cross-origin settings
	CrossOriginAuthentication bool
	CrossOriginLoc            string

	// SSO settings
	Sso         bool
	SsoDisabled bool

	// Custom login page
	CustomLoginPage   string
	CustomLoginPageOn bool
	InitiateLoginUri  string

	// Organization settings
	OrganizationUsage           string
	OrganizationRequireBehavior string

	// JWT configuration
	JwtConfiguration *auth0clientv1.Auth0JwtConfiguration

	// Refresh token configuration
	RefreshToken *auth0clientv1.Auth0RefreshTokenConfiguration

	// Native social login
	NativeSocialLogin *auth0clientv1.Auth0NativeSocialLogin

	// Mobile configuration
	Mobile *auth0clientv1.Auth0MobileConfiguration

	// Client metadata
	ClientMetadata map[string]string
	ClientAliases  []string

	// Additional settings
	IsTokenEndpointIpHeaderTrusted bool
	OidcBackchannelLogout          *auth0clientv1.Auth0OidcBackchannelLogout
	EnabledConnections             []string

	// API Grants for authorizing API access (with resolved audience values)
	ApiGrants []*ApiGrant
}

// ApiGrant represents an API grant configuration with resolved audience value
type ApiGrant struct {
	Audience             string
	Scopes               []string
	AllowAnyOrganization bool
	OrganizationUsage    string
}

// initializeLocals creates and populates the Locals struct from stack input
func initializeLocals(ctx *pulumi.Context, stackInput *auth0clientv1.Auth0ClientStackInput) *Locals {
	locals := &Locals{}

	// Store the target resource
	locals.Auth0Client = stackInput.Target

	spec := stackInput.Target.Spec
	metadata := stackInput.Target.Metadata

	// Core configuration
	locals.ClientName = metadata.Name
	locals.ApplicationType = spec.ApplicationType
	locals.Description = spec.Description
	locals.LogoUri = spec.LogoUri

	// URLs configuration
	locals.Callbacks = spec.Callbacks
	locals.AllowedLogoutUrls = spec.AllowedLogoutUrls
	locals.WebOrigins = spec.WebOrigins
	locals.AllowedOrigins = spec.AllowedOrigins

	// OAuth configuration
	locals.GrantTypes = spec.GrantTypes
	locals.OidcConformant = spec.OidcConformant
	locals.IsFirstParty = spec.IsFirstParty

	// Cross-origin settings
	locals.CrossOriginAuthentication = spec.CrossOriginAuthentication
	locals.CrossOriginLoc = spec.CrossOriginLoc

	// SSO settings
	locals.Sso = spec.Sso
	locals.SsoDisabled = spec.SsoDisabled

	// Custom login page
	locals.CustomLoginPage = spec.CustomLoginPage
	locals.CustomLoginPageOn = spec.CustomLoginPageOn
	locals.InitiateLoginUri = spec.InitiateLoginUri

	// Organization settings
	locals.OrganizationUsage = spec.OrganizationUsage
	locals.OrganizationRequireBehavior = spec.OrganizationRequireBehavior

	// JWT configuration
	locals.JwtConfiguration = spec.JwtConfiguration

	// Refresh token configuration
	locals.RefreshToken = spec.RefreshToken

	// Native social login
	locals.NativeSocialLogin = spec.NativeSocialLogin

	// Mobile configuration
	locals.Mobile = spec.Mobile

	// Client metadata
	locals.ClientMetadata = spec.ClientMetadata
	locals.ClientAliases = spec.ClientAliases

	// Additional settings
	locals.IsTokenEndpointIpHeaderTrusted = spec.IsTokenEndpointIpHeaderTrusted
	locals.OidcBackchannelLogout = spec.OidcBackchannelLogout

	// Extract enabled_connections values from StringValueOrRef
	for _, conn := range spec.EnabledConnections {
		if conn != nil && conn.GetValue() != "" {
			locals.EnabledConnections = append(locals.EnabledConnections, conn.GetValue())
		}
	}

	// API Grants - extract audience values from StringValueOrRef
	for _, grant := range spec.ApiGrants {
		if grant != nil && grant.Audience != nil && grant.Audience.GetValue() != "" {
			locals.ApiGrants = append(locals.ApiGrants, &ApiGrant{
				Audience:             grant.Audience.GetValue(),
				Scopes:               grant.Scopes,
				AllowAnyOrganization: grant.AllowAnyOrganization,
				OrganizationUsage:    grant.OrganizationUsage,
			})
		}
	}

	return locals
}
