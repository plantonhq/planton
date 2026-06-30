package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants matching stack_outputs.proto fields
const (
	OpId                    = "id"
	OpName                  = "name"
	OpStrategy              = "strategy"
	OpIsEnabled             = "is_enabled"
	OpProvisioningTicketURL = "provisioning_ticket_url"
	OpCallbackURL           = "callback_url"
	OpMetadataURL           = "metadata_url"
	OpEntityId              = "entity_id"
	OpEnabledClientIds      = "enabled_client_ids"
	OpRealms                = "realms"
)

// exportOutputs exports connection information to Pulumi stack outputs
func exportOutputs(ctx *pulumi.Context, connection *auth0.Connection, locals *Locals) error {
	// Export required outputs matching stack_outputs.proto
	ctx.Export(OpId, connection.ID())
	ctx.Export(OpName, connection.Name)
	ctx.Export(OpStrategy, connection.Strategy)

	// The connection is considered enabled if it has enabled clients configured
	ctx.Export(OpIsEnabled, pulumi.Bool(len(locals.EnabledClients) > 0))

	// Export enabled clients from the locals (they're managed by ConnectionClients resource)
	clientsArray := pulumi.StringArray{}
	for _, client := range locals.EnabledClients {
		clientsArray = append(clientsArray, pulumi.String(client))
	}
	ctx.Export(OpEnabledClientIds, clientsArray)

	// Export realms
	ctx.Export(OpRealms, connection.Realms)

	// For SAML connections, we can derive metadata and entity information
	// Note: Auth0 connection resource doesn't directly expose these,
	// they would need to be constructed based on the tenant domain
	// For now, we export empty strings as placeholders
	ctx.Export(OpProvisioningTicketURL, pulumi.String(""))
	ctx.Export(OpCallbackURL, pulumi.String(""))
	ctx.Export(OpMetadataURL, pulumi.String(""))
	ctx.Export(OpEntityId, pulumi.String(""))

	return nil
}
