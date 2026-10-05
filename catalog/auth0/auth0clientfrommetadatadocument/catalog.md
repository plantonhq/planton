# Auth0 Client From Metadata Document

Registers an application in an Auth0 tenant from the Client ID Metadata Document it hosts at an https URL, the way an MCP client onboards itself. Auth0 fetches the document and takes the application's name, redirect URIs, logo and keys from it; the spec sets what the tenant decides over them -- token lifetimes, refresh-token rotation, origins, the redirect policy and metadata. The application is a strict third-party application with no client secret, and its Management API id is ready for client grants and domain-level connections.

## What Gets Created

When you deploy this Infra Component, the IaC module creates:

- **An application registered from its metadata document** in the tenant your Auth0 connection's credential belongs to -- Auth0 fetches the document at `externalClientId`, validates it, and registers the application as a strict third-party client; the module then applies only the settings the spec declares
- **The document's validation result** -- reported in the outputs on every read, with the warnings and violations Auth0 found in the document

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Auth0 Tenant Settings** -- with Client ID Metadata Document registration turned on; Auth0 refuses the registration otherwise.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `create:clients`, `read:clients`, `update:clients` and `delete:clients` on the tenant's Management API.
- **A metadata document served at `externalClientId`** whose own `client_id` is exactly that URL, reachable from the public internet without redirects.
- **No active Rules on the tenant** -- sign-in for these applications fails while Rules are active; move them to Actions first.
- **The Enterprise plan** (only for a confidential client whose document declares `private_key_jwt`).

## Deploy

### Console

Open the deployment store, find **Auth0 Client From Metadata Document**, and click **Deploy**. Start from the **MCP Client From Its Document** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ClientFromMetadataDocument
metadata:
  name: mcp-client
  org: acme-corp
  env: prod
spec:
  externalClientId: https://mcp-client.acme.com/.well-known/oauth-client-metadata
  externalClientIdVersion: 1
  description: MCP client for Acme's internal tools
```

```shell
planton apply -f auth0-client-from-metadata-document.yaml
```

This registers the application from the document Acme's MCP client serves, with a description of your own over the document's. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an Auth0 Client From Metadata Document. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The URL is the identity** -- `externalClientId` is the application's `client_id` in every sign-in flow, and Auth0 keys the registration by it. Applying a URL the tenant already registered takes that registration over (and destroy then deletes it); changing the URL registers a new application with a new `client_id`, so every grant that named the old one is applied again.

**Pull document changes on purpose** -- Auth0 reads the document once. When its owner changes redirect URIs, keys or the name, raise `externalClientIdVersion` and the next apply fetches it again; the settings the spec declares are applied over the fresh document right after, so they always win.

**Declare what the provider cannot leave alone** -- `description`, `allowedOrigins`, `webOrigins`, `clientMetadata`, `requireProofOfPossession`, `skipNonVerifiableCallbackUriConfirmationPrompt`, `organizationDiscoveryMethods`, `defaultOrganization` and `tokenQuota` are reset (or proposed for clearing) when Auth0 holds a value the spec does not declare. When adopting an existing registration, or when the document carries a description, declare the live values.

**Refresh tokens must expire** -- `refreshToken` needs `rotationType` and `expirationType: expiring`, and an idle lifetime never longer than the absolute one. Rotation is what makes a copied refresh token self-revealing, which is why public MCP clients should rotate.

**Keep redirect protection** -- `redirectionPolicy: open_redirect_protection` (Auth0's default here) shows Auth0's own error page when sign-in fails instead of redirecting to a callback someone else controls. Choose `allow_always` only for a client whose callbacks you trust.

**Sender-constrained tokens** -- `requireProofOfPossession: true` binds every token to the client's key (DPoP or mutual TLS); the APIs it calls must accept sender-constrained tokens, and mutual TLS needs the Enterprise plan with the Highly Regulated Identity add-on.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. `defaultOrganization.organizationId` takes an Auth0 Organization's id as a string, and the tenant is the one the Auth0 connection's credential belongs to.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `client_id` | The application's Management API id (`tpc_...`) | A client grant for an API; an Auth0 Connection's `enabledClients` |
| `external_client_id` | The document's URL, the `client_id` sent in sign-in flows | An API's allow list of callers; audit correlation |
| `callbacks` | The redirect URIs Auth0 took from the document | Reviewing where people are sent after sign-in |
| `validation_warnings` | What Auth0 ignored in the document | Telling the document's owner what to fix |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Register an MCP client** -- The document's URL, a description of your own and redirect protection, with a version to fetch the document again later. Start from the **MCP Client From Its Document** preset.

**Long-running agent sessions** -- The refresh-token grant with rotation, a 15-day idle and 30-day absolute lifetime, and a few seconds of retry overlap. Start from the **Rotating Refresh Tokens** preset.

**Give the client something to do** -- Promote the connections it may sign people in through to the domain level, and grant it each API it calls: a third-party application reaches no API without a client grant, and never the Management API.

## Works With

- [**Auth0 Tenant Settings**](/infra-catalog/auth0-tenant-settings) -- turns on Client ID Metadata Document registration, which this kind needs.
- [**Auth0 Connection**](/infra-catalog/auth0-connection) -- a domain-level connection (`isDomainConnection: true`) is the only way third-party applications sign people in.
- [**Auth0 Resource Server**](/infra-catalog/auth0-resource-server) -- the API the client is granted access to, and where sender-constrained tokens are required.
- [**Auth0 Client**](/infra-catalog/auth0-client) -- for a first-party application whose settings you own entirely, with no metadata document.
