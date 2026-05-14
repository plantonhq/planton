package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the stack outputs for the Auth0 client.
// Fields available directly on the resource are exported inline.
// Fields only available via the Management API read-back (client_secret,
// token_endpoint_auth_method) are retrieved via LookupClient.
func exportOutputs(ctx *pulumi.Context, client *auth0.Client, locals *Locals) error {
	ctx.Export("id", client.ID())
	ctx.Export("client_id", client.ClientId)
	ctx.Export("name", client.Name)
	ctx.Export("application_type", client.AppType)
	ctx.Export("signing_keys", client.SigningKeys)
	ctx.Export("allowed_clients", client.AllowedClients)

	// Look up the created client to access computed-only attributes
	// (client_secret, token_endpoint_auth_method) that the resource doesn't expose.
	lookupResult := auth0.LookupClientOutput(ctx, auth0.LookupClientOutputArgs{
		ClientId: client.ClientId,
	})
	ctx.Export("client_secret", lookupResult.ClientSecret())
	ctx.Export("token_endpoint_auth_method", lookupResult.TokenEndpointAuthMethod())

	return nil
}
