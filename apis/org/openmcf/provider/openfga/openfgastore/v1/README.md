# OpenFgaStore

Creates an OpenFGA store for managing fine-grained authorization.

## What is OpenFGA?

[OpenFGA](https://openfga.dev) is an open-source authorization system that implements Google's Zanzibar paper. It provides:

- **Fine-grained authorization**: Control access at the resource level
- **Relationship-based access control**: Define permissions based on relationships between users and resources
- **High performance**: Designed for low-latency authorization checks at scale

## What is a Store?

A **store** is the top-level container for authorization data in OpenFGA. Each store contains:

- **Authorization models**: Define the types (user, document, folder, etc.), relations (owner, viewer, parent), and permissions
- **Relationship tuples**: The actual authorization data (e.g., "user:alice is owner of document:budget")

Use separate stores for:
- Different environments (development, staging, production)
- Different applications or services
- Multi-tenant isolation

## Important: Terraform Only

> ⚠️ **OpenFGA only has a Terraform provider.** There is no Pulumi provider available.
> 
> You **must** use `--provisioner tofu` when deploying this component.

## Usage

### 1. Create OpenFGA Credentials

```yaml
# openfga-creds.yaml
apiUrl: http://localhost:8080
apiToken: your-api-token
```

Or for client credentials authentication:

```yaml
# openfga-creds.yaml
apiUrl: https://api.us1.fga.dev
clientId: your-client-id
clientSecret: your-client-secret
apiTokenIssuer: https://fga.us.auth0.com/oauth/token
apiAudience: https://api.us1.fga.dev/
```

### 2. Create Manifest

```yaml
# openfga-store.yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: production-authz
  org: my-company
  env: production
spec:
  name: production-authorization-store
```

### 3. Deploy

```bash
openmcf apply --manifest openfga-store.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Specification

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `spec.name` | string | Yes | Display name of the OpenFGA store. Immutable. |

## Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `id` | Unique identifier of the store (used in API calls) |
| `name` | Display name of the store |

## Examples

### Development Store

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: dev-store
  org: engineering
  env: development
spec:
  name: development-authz
```

### Production Store

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: prod-store
  org: engineering
  env: production
  labels:
    team: platform
    compliance: soc2
spec:
  name: production-authorization
```

## What's Next?

After creating a store, you typically need to:

1. **Create an authorization model**: Define types, relations, and permissions
2. **Write relationship tuples**: Add the actual authorization data
3. **Check permissions**: Query the system to check if a user has access

These will be available as separate deployment components:
- `OpenFgaAuthorizationModel` (coming soon)
- `OpenFgaTuple` (coming soon)

## References

- [OpenFGA Documentation](https://openfga.dev/docs)
- [OpenFGA Concepts: What is a Store](https://openfga.dev/docs/concepts#what-is-a-store)
- [Terraform Provider OpenFGA](https://registry.terraform.io/providers/openfga/openfga/latest/docs)
- [OpenFGA Store Resource](https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/store)
