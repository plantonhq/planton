package module

import (
	"github.com/pkg/errors"
	auth0resourceserverv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0resourceserver/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createResourceServer creates an Auth0 Resource Server (API) -- the twin of
// iac/tf/main.tf's auth0_resource_server.
func createResourceServer(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.ResourceServer, error) {
	resourceServer, err := auth0.NewResourceServer(ctx, locals.ResourceName, resourceServerArgs(locals), pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 resource server %s", locals.ResourceName)
	}

	return resourceServer, nil
}

// resourceServerArgs builds the resource server's arguments from the locals.
// It is a pure function of the spec (locals_test.go pins it): every setting
// the spec leaves unset stays nil and is never sent, so an API adopted into
// this kind keeps what Auth0 holds for it.
func resourceServerArgs(locals *Locals) *auth0.ResourceServerArgs {
	args := &auth0.ResourceServerArgs{
		Identifier: pulumi.String(locals.Identifier),
		Name:       pulumi.String(locals.Name),
	}

	// Add signing algorithm if specified
	if locals.SigningAlg != "" {
		args.SigningAlg = pulumi.String(locals.SigningAlg)
	}

	// Token settings
	if locals.AllowOfflineAccess != nil {
		args.AllowOfflineAccess = pulumi.BoolPtr(*locals.AllowOfflineAccess)
	}

	if locals.TokenLifetime > 0 {
		args.TokenLifetime = pulumi.Int(int(locals.TokenLifetime))
	}

	if locals.TokenLifetimeForWeb > 0 {
		args.TokenLifetimeForWeb = pulumi.Int(int(locals.TokenLifetimeForWeb))
	}

	// Access control settings
	if locals.SkipConsentForVerifiableFirstPartyClients != nil {
		args.SkipConsentForVerifiableFirstPartyClients = pulumi.BoolPtr(*locals.SkipConsentForVerifiableFirstPartyClients)
	}
	if locals.EnforcePolicies != nil {
		args.EnforcePolicies = pulumi.BoolPtr(*locals.EnforcePolicies)
	}

	if locals.TokenDialect != "" {
		args.TokenDialect = pulumi.String(locals.TokenDialect)
	}

	// Unmanaged when unset: each is sent only as the spec declares it.
	args.AllowOnlineAccess = pulumi.BoolPtrFromPtr(locals.AllowOnlineAccess)
	args.AllowOnlineAccessWithEphemeralSessions = pulumi.BoolPtrFromPtr(locals.AllowOnlineAccessWithEphemeralSessions)
	args.ConsentPolicy = pulumi.StringPtrFromPtr(locals.ConsentPolicy)
	args.TokenLifetimeForAnonymousAccessTokens = pulumi.IntPtrFromPtr(locals.TokenLifetimeForAnonymousAccessTokens)
	args.VerificationLocation = pulumi.StringPtrFromPtr(locals.VerificationLocation)
	if locals.SigningSecret != nil {
		// A shared signing key: sent as a Pulumi secret, so state and previews
		// never show it.
		args.SigningSecret = pulumi.ToSecret(pulumi.String(*locals.SigningSecret)).(pulumi.StringOutput)
	}

	// Each block is assigned only when declared: an interface holding a typed
	// nil would be sent as an empty block.
	if at := accessTokenArgs(locals.AccessToken); at != nil {
		args.AccessToken = at
	}
	if len(locals.AuthorizationDetails) > 0 {
		args.AuthorizationDetails = authorizationDetailsArgs(locals.AuthorizationDetails)
	}
	if p := locals.AuthorizationPolicy; p != nil {
		args.AuthorizationPolicy = &auth0.ResourceServerAuthorizationPolicyArgs{
			PolicyId: pulumi.StringPtrFromPtr(p.PolicyId),
		}
	}
	if p := locals.ProofOfPossession; p != nil {
		args.ProofOfPossession = &auth0.ResourceServerProofOfPossessionArgs{
			Disable:     pulumi.BoolPtrFromPtr(p.Disable),
			Mechanism:   pulumi.StringPtrFromPtr(p.Mechanism),
			Required:    pulumi.BoolPtrFromPtr(p.Required),
			RequiredFor: pulumi.StringPtrFromPtr(p.RequiredFor),
		}
	}
	if sta := subjectTypeAuthorizationArgs(locals.SubjectTypeAuthorization); sta != nil {
		args.SubjectTypeAuthorization = sta
	}
	if te := tokenEncryptionArgs(locals.TokenEncryption); te != nil {
		args.TokenEncryption = te
	}

	return args
}

// accessTokenArgs maps the access_token block. A declared claims_mapping is
// sent with its whole claim list -- an empty list clears the claims.
func accessTokenArgs(at *auth0resourceserverv1alpha1.Auth0ResourceServerAccessToken) *auth0.ResourceServerAccessTokenArgs {
	if at == nil {
		return nil
	}
	args := &auth0.ResourceServerAccessTokenArgs{}
	if cm := at.ClaimsMapping; cm != nil {
		claims := auth0.ResourceServerAccessTokenClaimsMappingCustomClaimArray{}
		for _, claim := range cm.CustomClaims {
			claims = append(claims, &auth0.ResourceServerAccessTokenClaimsMappingCustomClaimArgs{
				Name:       pulumi.String(claim.Name),
				Expression: pulumi.String(claim.Expression),
			})
		}
		args.ClaimsMapping = &auth0.ResourceServerAccessTokenClaimsMappingArgs{CustomClaims: claims}
	}
	return args
}

// authorizationDetailsArgs maps the authorization_details types, or the single
// entry that disables them.
func authorizationDetailsArgs(details []*auth0resourceserverv1alpha1.Auth0ResourceServerAuthorizationDetail) auth0.ResourceServerAuthorizationDetailArray {
	array := auth0.ResourceServerAuthorizationDetailArray{}
	for _, detail := range details {
		array = append(array, &auth0.ResourceServerAuthorizationDetailArgs{
			Type:    pulumi.StringPtrFromPtr(detail.Type),
			Disable: pulumi.BoolPtrFromPtr(detail.Disable),
		})
	}
	return array
}

// subjectTypeAuthorizationArgs maps the access policy. Each subject's block is
// sent only with its policy: the provider keeps a block the spec leaves out,
// so a policy never declared is never touched.
func subjectTypeAuthorizationArgs(sta *auth0resourceserverv1alpha1.Auth0ResourceServerSubjectTypeAuthorization) *auth0.ResourceServerSubjectTypeAuthorizationArgs {
	if sta == nil {
		return nil
	}
	args := &auth0.ResourceServerSubjectTypeAuthorizationArgs{}
	if user := sta.GetUser(); user != nil && user.Policy != nil {
		args.User = &auth0.ResourceServerSubjectTypeAuthorizationUserArgs{Policy: pulumi.String(user.GetPolicy())}
	}
	if client := sta.GetClient(); client != nil && client.Policy != nil {
		args.Client = &auth0.ResourceServerSubjectTypeAuthorizationClientArgs{Policy: pulumi.String(client.GetPolicy())}
	}
	if anonymous := sta.GetAnonymousUser(); anonymous != nil && anonymous.Policy != nil {
		args.AnonymousUser = &auth0.ResourceServerSubjectTypeAuthorizationAnonymousUserArgs{Policy: pulumi.String(anonymous.GetPolicy())}
	}
	return args
}

// tokenEncryptionArgs maps token encryption: the format and the API's public
// key, or disable.
func tokenEncryptionArgs(te *auth0resourceserverv1alpha1.Auth0ResourceServerTokenEncryption) *auth0.ResourceServerTokenEncryptionArgs {
	if te == nil {
		return nil
	}
	args := &auth0.ResourceServerTokenEncryptionArgs{
		Disable: pulumi.BoolPtrFromPtr(te.Disable),
		Format:  pulumi.StringPtrFromPtr(te.Format),
	}
	if key := te.EncryptionKey; key != nil {
		args.EncryptionKey = &auth0.ResourceServerTokenEncryptionEncryptionKeyArgs{
			Algorithm: pulumi.String(key.Algorithm),
			Pem:       pulumi.String(key.Pem),
			Kid:       pulumi.StringPtrFromPtr(key.Kid),
			Name:      pulumi.StringPtrFromPtr(key.Name),
		}
	}
	return args
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
