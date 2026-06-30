# Auth0 Resource Server Pulumi Module

This Pulumi module deploys an Auth0 Resource Server (API) using Go.

## Overview

The module creates and configures an Auth0 Resource Server with:
- API identifier (audience) configuration
- Token settings (lifetime, signing algorithm, dialect)
- Scope/permission definitions
- RBAC enablement

## Project Structure

```
.
├── main.go           # Pulumi entry point
├── Pulumi.yaml       # Stack configuration
├── Makefile          # Convenience targets
└── module/
    ├── main.go       # Module orchestration
    ├── locals.go     # Local variable computation
    ├── resourceserver.go  # Resource server creation
    └── outputs.go    # Stack exports
```

## Prerequisites

1. **Pulumi CLI**: Install from https://www.pulumi.com/docs/install/
2. **Go 1.21+**: Required for building the module
3. **Auth0 Tenant**: An Auth0 tenant with M2M credentials

## Configuration

### Environment Variables

```bash
export AUTH0_DOMAIN="your-tenant.auth0.com"
export AUTH0_CLIENT_ID="your-m2m-client-id"
export AUTH0_CLIENT_SECRET="your-m2m-client-secret"
```

### Stack Input

The module expects a `stack_input.yaml` file with the following structure:

```yaml
target:
  api_version: auth0.planton.dev/v1
  kind: Auth0ResourceServer
  metadata:
    name: my-api
  spec:
    identifier: https://api.example.com/
    name: My API
    signing_alg: RS256
    token_lifetime: 86400
    scopes:
      - name: read:data
        description: Read data
provider_config:
  domain: ${AUTH0_DOMAIN}
  client_id: ${AUTH0_CLIENT_ID}
  client_secret: ${AUTH0_CLIENT_SECRET}
```

## Usage

### Initialize Stack

```bash
pulumi stack init dev
```

### Configure Stack Input

```bash
pulumi config set --path 'target.spec.identifier' 'https://api.example.com/'
```

Or use a stack input file with the Planton CLI.

### Deploy

```bash
make up
# or
pulumi up --yes
```

### Preview Changes

```bash
make preview
# or
pulumi preview
```

### Destroy

```bash
make destroy
# or
pulumi destroy --yes
```

## Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `id` | Auth0's internal resource server ID |
| `identifier` | The API audience |
| `name` | Display name |
| `signing_alg` | Token signing algorithm |
| `signing_secret` | HS256 signing secret (if applicable) |
| `token_lifetime` | Token validity in seconds |
| `token_lifetime_for_web` | Web token validity |
| `allow_offline_access` | Refresh token setting |
| `skip_consent_for_verifiable_first_party_clients` | Consent skip setting |
| `enforce_policies` | RBAC enabled |
| `token_dialect` | Token format |
| `is_system` | System resource server flag |
| `client_id` | Associated client ID |

## Resources Created

- `auth0:index/resourceServer:ResourceServer` - The API configuration
- `auth0:index/resourceServerScopes:ResourceServerScopes` - API permissions (if scopes defined)

## Error Handling

Common issues:

1. **Authentication failed**: Verify AUTH0_DOMAIN, AUTH0_CLIENT_ID, and AUTH0_CLIENT_SECRET
2. **Insufficient permissions**: Ensure M2M app has resource server management scopes
3. **Identifier exists**: Resource server identifiers must be unique per tenant

## Testing

Use the test manifest in `../hack/manifest.yaml`:

```bash
cd ../hack
planton pulumi up --manifest manifest.yaml
```

## Related

- [Auth0 APIs Documentation](https://auth0.com/docs/get-started/apis)
- [Pulumi Auth0 Provider](https://www.pulumi.com/registry/packages/auth0/)
