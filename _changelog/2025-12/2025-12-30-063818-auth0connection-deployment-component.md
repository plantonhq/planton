# Auth0Connection Deployment Component

**Date**: December 30, 2025
**Type**: Feature
**Components**: Provider Framework, API Definitions, Pulumi IaC Module, Terraform IaC Module, Proto Validations

## Summary

Implemented `Auth0Connection` as the first deployment component for the newly added Auth0 provider. This component enables declarative management of Auth0 identity connections including database, social (Google, Facebook, GitHub), and enterprise SSO (SAML, OIDC, Azure AD) connections through both Pulumi and Terraform IaC modules.

## Problem Statement / Motivation

With Auth0 recently integrated as a cloud provider in Planton, there was no way to manage Auth0 resources. Users needed a deployment component to:

- Configure identity connections declaratively via YAML manifests
- Support multiple authentication strategies in a unified API
- Enable infrastructure-as-code workflows for identity management
- Integrate with the existing CLI credential flow for Auth0

### Pain Points

- No Auth0 deployment components existed in the provider range
- Auth0 connections have strategy-specific configurations that need proper abstraction
- Manual Auth0 dashboard configuration lacks version control and reproducibility
- Need consistent patterns matching other SaaS providers (Confluent, Atlas, Snowflake)

## Solution / What's New

Created a complete Auth0Connection deployment component following the established provider patterns with full IaC support.

### Registry Allocation

Allocated Auth0 provider block **2100–2299** (200 numbers) for future Auth0 components:

```protobuf
// 2100–2299: Auth0 resources
Auth0Connection = 2100 [(kind_meta) = {
  provider: auth0
  version: v1
  id_prefix: "a0conn"
}];
```

### Supported Connection Strategies

| Strategy | Type | Options Block |
|----------|------|---------------|
| `auth0` | Database | `database_options` |
| `google-oauth2` | Social | `social_options` |
| `facebook` | Social | `social_options` |
| `github` | Social | `social_options` |
| `linkedin` | Social | `social_options` |
| `twitter` | Social | `social_options` |
| `microsoft-account` | Social | `social_options` |
| `apple` | Social | `social_options` |
| `samlp` | Enterprise | `saml_options` |
| `oidc` | Enterprise | `oidc_options` |
| `waad` | Enterprise | `azure_ad_options` |

## Implementation Details

### Proto API (4 files)

**spec.proto** - Comprehensive connection specification:
- Strategy-specific option messages (Auth0DatabaseOptions, Auth0SocialOptions, Auth0SamlOptions, Auth0OidcOptions, Auth0AzureAdOptions)
- Validation rules using buf.validate (required fields, enum values, numeric ranges)
- Detailed documentation for every field

**stack_outputs.proto** - Deployment outputs:
- Connection ID, name, strategy
- Enabled status and client IDs
- Realms and metadata URLs (for enterprise connections)

**api.proto** - KRM envelope:
- apiVersion: `auth0.planton.dev/v1`
- kind: `Auth0Connection`
- Standard metadata, spec, status structure

**stack_input.proto** - IaC module inputs:
- Target Auth0Connection resource
- Auth0ProviderConfig credentials

### Pulumi Module (Go)

```
iac/pulumi/
├── main.go           # Entry point
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Computed values from spec
│   ├── connection.go # Connection creation with strategy routing
│   └── outputs.go    # Stack output exports
└── Makefile, Pulumi.yaml, debug.sh
```

Key implementation:
- Uses `github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0`
- Separate `auth0.ConnectionClients` resource for client enablement
- Strategy-specific option building with type-safe SDK calls

### Terraform Module (HCL)

```
iac/tf/
├── provider.tf   # Auth0 provider with credential variable
├── variables.tf  # Input variables mirroring spec.proto
├── locals.tf     # Computed local values
├── main.tf       # Connection + clients resources
└── outputs.tf    # Output definitions
```

Key implementation:
- Uses `auth0/auth0` provider v1.x
- Dynamic blocks for strategy-specific options
- `auth0_connection_clients` resource for client enablement

### Validation Tests (30 cases)

spec_test.go covers:
- Valid configurations for all 11 strategies
- Required field validation (strategy, metadata, spec)
- Enum validation (password_policy, protocol_binding, signature_algorithm)
- Range validation (password_history_size: 0-24, max_groups_to_retrieve: ≥0)
- Missing/invalid credential validation for each strategy type

## Benefits

### For Users
- **Declarative Identity Management**: Configure Auth0 connections via YAML
- **Multi-Strategy Support**: One component for database, social, and enterprise SSO
- **Credential Integration**: Uses existing `--auth0-provider-config` CLI flag
- **Production Defaults**: Secure-by-default options (brute force protection, password policies)

### For Developers
- **Pattern Consistency**: Follows established provider component patterns
- **Full Documentation**: README, examples.md, and comprehensive research docs
- **Dual IaC Support**: Both Pulumi and Terraform with feature parity
- **Extensible**: Easy to add new strategies following the option pattern

### For Platform Teams
- **Version Control**: Auth0 configuration as code
- **Environment Consistency**: Same manifests across dev/staging/production
- **Audit Trail**: Changes tracked in git history

## Impact

### Direct
- Users can now manage Auth0 connections through Planton
- CLI supports Auth0Connection manifests with `--auth0-provider-config` flag
- Auth0 provider now has its first deployment component

### Registry
- Auth0 range established: 2100–2299 (199 slots remaining for future components)
- First SaaS identity provider in the platform

### Future Work Enabled
- Auth0Client component (applications)
- Auth0Role component (RBAC)
- Auth0Rule component (extensibility)
- Auth0Action component (Auth0 Actions)

## Files Changed

| Category | Files |
|----------|-------|
| Registry | `cloud_resource_kind.proto` |
| Proto API | `spec.proto`, `api.proto`, `stack_input.proto`, `stack_outputs.proto` |
| Generated | `*.pb.go` (4 files) |
| Tests | `spec_test.go` |
| Pulumi | `main.go`, `module/*.go`, `Makefile`, `Pulumi.yaml`, `debug.sh` |
| Terraform | `provider.tf`, `variables.tf`, `locals.tf`, `main.tf`, `outputs.tf` |
| Docs | `README.md`, `examples.md`, `docs/README.md` |
| Supporting | `hack/manifest.yaml`, IaC READMEs, `overview.md` |

**Total**: ~30 files, ~3000 lines of code

## Usage Examples

### Database Connection
```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Connection
metadata:
  name: user-database
spec:
  strategy: auth0
  display_name: Email Sign Up
  enabled_clients: ["app-client-id"]
  database_options:
    password_policy: good
    brute_force_protection: true
```

### Google OAuth
```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Connection
metadata:
  name: google-login
spec:
  strategy: google-oauth2
  social_options:
    client_id: "google-client-id"
    client_secret: "google-secret"
    scopes: ["openid", "profile", "email"]
```

### Enterprise SAML SSO
```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Connection
metadata:
  name: okta-sso
spec:
  strategy: samlp
  is_domain_connection: true
  realms: ["company.com"]
  saml_options:
    sign_in_endpoint: "https://company.okta.com/sso"
    signing_cert: "-----BEGIN CERTIFICATE-----..."
```

## Related Work

- Builds on Auth0 provider integration ([2025-12-30-054629-auth0-provider-integration.md](./_changelog/2025-12/2025-12-30-054629-auth0-provider-integration.md))
- Follows patterns from MongodbAtlas, ConfluentKafka, SnowflakeDatabase components
- Uses Auth0 Pulumi SDK v3 and Auth0 Terraform provider v1

---

**Status**: ✅ Production Ready
**Build**: CLI compiles, 30/30 tests pass, Terraform validates

