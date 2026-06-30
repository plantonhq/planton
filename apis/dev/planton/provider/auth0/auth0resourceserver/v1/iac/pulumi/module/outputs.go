package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports stack outputs for the Auth0 Resource Server
func exportOutputs(ctx *pulumi.Context, resourceServer *auth0.ResourceServer, locals *Locals) error {
	// Export core identifiers
	ctx.Export("id", resourceServer.ID())
	ctx.Export("identifier", resourceServer.Identifier)
	ctx.Export("name", resourceServer.Name)

	// Export token settings
	ctx.Export("signing_alg", resourceServer.SigningAlg)
	ctx.Export("signing_secret", resourceServer.SigningSecret)
	ctx.Export("token_lifetime", resourceServer.TokenLifetime)
	ctx.Export("token_lifetime_for_web", resourceServer.TokenLifetimeForWeb)

	// Export access control settings
	ctx.Export("allow_offline_access", resourceServer.AllowOfflineAccess)
	ctx.Export("skip_consent_for_verifiable_first_party_clients", resourceServer.SkipConsentForVerifiableFirstPartyClients)
	ctx.Export("enforce_policies", resourceServer.EnforcePolicies)
	ctx.Export("token_dialect", resourceServer.TokenDialect)

	// Export system flags
	ctx.Export("is_system", resourceServer.IsSystem)
	ctx.Export("client_id", resourceServer.ClientId)

	return nil
}
