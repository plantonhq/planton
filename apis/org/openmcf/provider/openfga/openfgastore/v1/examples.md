# OpenFgaStore Examples

## Basic Store

Create a simple authorization store:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: my-store
  org: my-company
  env: development
spec:
  name: my-authorization-store
```

## Environment-Specific Stores

### Development

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: dev-authz
  org: engineering
  env: development
  labels:
    team: platform
spec:
  name: development-authorization
```

### Staging

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: staging-authz
  org: engineering
  env: staging
  labels:
    team: platform
spec:
  name: staging-authorization
```

### Production

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: prod-authz
  org: engineering
  env: production
  labels:
    team: platform
    compliance: soc2
    criticality: high
spec:
  name: production-authorization
```

## Multi-Tenant Stores

For SaaS applications with tenant isolation:

```yaml
# Tenant A
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: tenant-a-authz
  org: my-saas
  env: production
  labels:
    tenant: tenant-a
spec:
  name: tenant-a-authorization

---
# Tenant B
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: tenant-b-authz
  org: my-saas
  env: production
  labels:
    tenant: tenant-b
spec:
  name: tenant-b-authorization
```

## Application-Specific Stores

For microservices with separate authorization domains:

```yaml
# Document Service Authorization
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: docs-authz
  org: my-company
  env: production
  labels:
    service: document-service
spec:
  name: document-service-authorization

---
# Project Service Authorization
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: projects-authz
  org: my-company
  env: production
  labels:
    service: project-service
spec:
  name: project-service-authorization
```

## Deployment Commands

Deploy with local OpenFGA server:

```bash
# Create credentials file for local server
cat > openfga-local.yaml << EOF
apiUrl: http://localhost:8080
EOF

# Deploy (no auth needed for local server)
openmcf apply --manifest store.yaml \
  --openfga-provider-config openfga-local.yaml \
  --provisioner tofu
```

Deploy with Okta FGA (cloud-hosted):

```bash
# Create credentials file for Okta FGA
cat > openfga-cloud.yaml << EOF
apiUrl: https://api.us1.fga.dev
clientId: your-client-id
clientSecret: your-client-secret
apiTokenIssuer: https://fga.us.auth0.com/oauth/token
apiAudience: https://api.us1.fga.dev/
EOF

# Deploy
openmcf apply --manifest store.yaml \
  --openfga-provider-config openfga-cloud.yaml \
  --provisioner tofu
```

## Important Notes

1. **Store names are immutable**: Changing `spec.name` will replace the store, deleting all data
2. **Use Terraform/Tofu**: The `--provisioner tofu` flag is **required** (no Pulumi provider exists)
3. **Plan before apply**: Use `openmcf plan` to preview changes before applying
4. **Backup data**: Before deleting a store, export any important authorization models and tuples
