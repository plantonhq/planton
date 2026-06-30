package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createResourceServer creates an Auth0 Resource Server (API)
func createResourceServer(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.ResourceServer, error) {
	// Build resource server arguments
	resourceServerArgs := &auth0.ResourceServerArgs{
		Identifier: pulumi.String(locals.Identifier),
		Name:       pulumi.String(locals.Name),
	}

	// Add signing algorithm if specified
	if locals.SigningAlg != "" {
		resourceServerArgs.SigningAlg = pulumi.String(locals.SigningAlg)
	}

	// Token settings
	resourceServerArgs.AllowOfflineAccess = pulumi.Bool(locals.AllowOfflineAccess)

	if locals.TokenLifetime > 0 {
		resourceServerArgs.TokenLifetime = pulumi.Int(int(locals.TokenLifetime))
	}

	if locals.TokenLifetimeForWeb > 0 {
		resourceServerArgs.TokenLifetimeForWeb = pulumi.Int(int(locals.TokenLifetimeForWeb))
	}

	// Access control settings
	resourceServerArgs.SkipConsentForVerifiableFirstPartyClients = pulumi.Bool(locals.SkipConsentForVerifiableFirstPartyClients)
	resourceServerArgs.EnforcePolicies = pulumi.Bool(locals.EnforcePolicies)

	if locals.TokenDialect != "" {
		resourceServerArgs.TokenDialect = pulumi.String(locals.TokenDialect)
	}

	// Create the resource server
	resourceServer, err := auth0.NewResourceServer(ctx, locals.ResourceName, resourceServerArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 resource server %s", locals.ResourceName)
	}

	return resourceServer, nil
}

// createResourceServerScopes creates scopes for the resource server
func createResourceServerScopes(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider, resourceServer *auth0.ResourceServer) (*auth0.ResourceServerScopes, error) {
	if len(locals.Scopes) == 0 {
		return nil, nil
	}

	// Build scopes array
	scopeArray := auth0.ResourceServerScopesScopeArray{}
	for _, scope := range locals.Scopes {
		scopeArgs := &auth0.ResourceServerScopesScopeArgs{
			Name: pulumi.String(scope.Name),
		}
		if scope.Description != "" {
			scopeArgs.Description = pulumi.String(scope.Description)
		}
		scopeArray = append(scopeArray, scopeArgs)
	}

	// Create the resource server scopes
	resourceServerScopes, err := auth0.NewResourceServerScopes(ctx, locals.ResourceName+"-scopes", &auth0.ResourceServerScopesArgs{
		ResourceServerIdentifier: resourceServer.Identifier,
		Scopes:                   scopeArray,
	}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{resourceServer}))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scopes for resource server %s", locals.ResourceName)
	}

	return resourceServerScopes, nil
}
