# Auth0Client API Grants Support

**Date**: January 6, 2026
**Type**: Feature
**Components**: Auth0 Provider, API Definitions, Pulumi Module, Terraform Module

## Summary

Added `api_grants` support to the Auth0Client deployment component, enabling Machine-to-Machine (M2M) applications to be fully configured with API access authorization without manual Auth0 Dashboard intervention. This closes a critical gap where M2M apps could authenticate but couldn't access any APIs without manual configuration.

## Problem Statement / Motivation

The Auth0Client component allowed creating M2M applications with OAuth grant types, but did not support authorizing clients for specific APIs. This meant that after deploying an M2M client via Planton, operators had to manually log into the Auth0 Dashboard to authorize the client and grant API scopes.

### Pain Points

- **Broken IaC workflow**: Manual steps defeated the purpose of infrastructure-as-code
- **Environment repetition**: Manual authorization needed for every environment (dev, staging, prod)
- **Human error risk**: Wrong scopes or forgotten authorization could cause production issues
- **Incomplete automation**: InfraCharts couldn't fully automate Auth0 tenant setup

### The Key Distinction

| Concept | What it does | Status |
|---------|--------------|--------|
| `grant_types` | Which OAuth flows the client can use (e.g., `client_credentials`) | Already supported |
| `api_grants` | Which APIs the client can access with what scopes | **Now supported** |

Without `api_grants`, an M2M application exists but can't actually call any APIs.

## Solution / What's New

Added a new `api_grants` field to `Auth0ClientSpec` that creates `auth0_client_grant` resources for each specified API.

### New Proto Message

```protobuf
message Auth0ClientApiGrant {
  string audience = 1 [(buf.validate.field).required = true];
  repeated string scopes = 2;
  bool allow_any_organization = 3;
  string organization_usage = 4;
}
```

### Usage Example

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Client
metadata:
  name: user-manager
spec:
  application_type: non_interactive
  grant_types:
    - client_credentials
  api_grants:
    - audience: "https://my-tenant.us.auth0.com/api/v2/"
      scopes:
        - read:users
        - read:user_idp_tokens
```

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/auth0/auth0client/v1/spec.proto`

- Added `Auth0ClientApiGrant` message with:
  - `audience` (required): API identifier (Resource Server identifier)
  - `scopes`: Permissions granted for the API
  - `allow_any_organization`: Organization access control flag
  - `organization_usage`: Organization mode (deny/allow/require)
- Added `api_grants` field (field 29) to `Auth0ClientSpec`

### Pulumi Module Changes

**Files**: `iac/pulumi/module/locals.go`, `client.go`, `main.go`

- Added `ApiGrants` field to `Locals` struct
- Created `createClientGrants()` function that:
  - Iterates over each API grant
  - Creates `auth0.ClientGrant` resources with proper dependencies
  - Handles organization settings when specified
- Called from main `Resources()` function after client creation

```go
func createClientGrants(ctx *pulumi.Context, locals *Locals, 
    provider *auth0.Provider, client *auth0.Client) error {
    for i, grant := range locals.ApiGrants {
        _, err := auth0.NewClientGrant(ctx, grantName, &auth0.ClientGrantArgs{
            ClientId: client.ClientId,
            Audience: pulumi.String(grant.Audience),
            Scopes:   scopesArray,
        }, pulumi.DependsOn([]pulumi.Resource{client}))
    }
}
```

### Terraform Module Changes

**Files**: `iac/tf/variables.tf`, `locals.tf`, `main.tf`

- Added `api_grants` to spec variable type definition
- Added `api_grants` local value
- Created `auth0_client_grant` resource with `for_each` loop:

```hcl
resource "auth0_client_grant" "api_grants" {
  for_each = {
    for idx, grant in local.api_grants : idx => grant
    if grant.audience != null && grant.audience != ""
  }
  
  client_id = auth0_client.this.id
  audience  = each.value.audience
  scopes    = coalesce(each.value.scopes, [])
}
```

## Benefits

### For Platform Engineers

- **Full automation**: Deploy M2M apps with complete API authorization in one manifest
- **Environment consistency**: Same configuration across dev/staging/prod
- **Version control**: API grants defined in Git, not in dashboard clicks

### For Operations

- **Reproducible deployments**: InfraCharts can now fully configure Auth0 tenants
- **Audit trail**: All API authorizations tracked in Git history
- **Reduced errors**: No manual steps means no forgotten authorizations

### For Development Velocity

- **Faster onboarding**: New environments get full Auth0 setup automatically
- **Self-service**: Teams can define their own API access needs in manifests

## Impact

### Files Changed

| Path | Change Type |
|------|-------------|
| `auth0client/v1/spec.proto` | Added message and field |
| `auth0client/v1/spec.pb.go` | Regenerated |
| `auth0client/v1/iac/pulumi/module/locals.go` | Added ApiGrants field |
| `auth0client/v1/iac/pulumi/module/client.go` | Added createClientGrants function |
| `auth0client/v1/iac/pulumi/module/main.go` | Called createClientGrants |
| `auth0client/v1/iac/tf/variables.tf` | Added api_grants variable |
| `auth0client/v1/iac/tf/locals.tf` | Added api_grants local |
| `auth0client/v1/iac/tf/main.tf` | Added client_grant resource |
| `auth0client/v1/docs/README.md` | Updated scope, added documentation |
| `auth0client/v1/examples.md` | Added M2M + API grants examples |
| `frontend/.../spec_pb.ts` | TypeScript stubs regenerated |

### Documentation Updates

- Research doc now includes "API Grants" as in-scope (removed from out-of-scope)
- Added comprehensive section explaining grant_types vs api_grants
- Added 4 new M2M examples with api_grants configurations
- Updated Auth0 credential requirements to include client_grant permissions

## Related Work

- Issue: `_issues/2026-01-06-061930.deployment-component.feat.auth0-client-api-grants.md`
- Related component: Auth0Connection (may benefit from similar patterns)
- InfraChart: `planton-auth0-tenant-stack` can now be fully automated

## Usage Examples

### Management API Access

```yaml
api_grants:
  - audience: "https://my-tenant.us.auth0.com/api/v2/"
    scopes:
      - read:users
      - read:user_idp_tokens
```

### Custom API Access

```yaml
api_grants:
  - audience: "https://api.example.com/"
    scopes:
      - read:resources
      - write:resources
```

### Multiple APIs

```yaml
api_grants:
  - audience: "https://my-tenant.us.auth0.com/api/v2/"
    scopes: [read:users, update:users]
  - audience: "https://api.example.com/"
    scopes: [admin:resources]
```

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation
**Validation**: Proto generation, component tests, and full build passed

