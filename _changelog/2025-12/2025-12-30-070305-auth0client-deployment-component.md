# Auth0Client Deployment Component

**Date**: December 30, 2025
**Type**: Feature
**Components**: Provider Framework, API Definitions, Pulumi IaC Module, Terraform IaC Module, Proto Validations

## Summary

Implemented `Auth0Client` as the second deployment component for the Auth0 provider. This component enables declarative management of Auth0 Applications (OAuth 2.0 clients) including Single Page Apps, Native mobile apps, Regular Web apps, and Machine-to-Machine services through both Pulumi and Terraform IaC modules.

## Problem Statement / Motivation

Following the Auth0Connection component, users needed a way to manage Auth0 Applications programmatically. Auth0 Applications represent OAuth 2.0 clients that authenticate users and request API access.

### Pain Points

- No way to manage Auth0 Applications through Planton
- Manual dashboard configuration lacks version control and reproducibility
- Complex OAuth settings (callbacks, grant types, tokens) require expertise
- Mobile app configuration (iOS/Android) is error-prone without proper structure
- Inconsistent application settings across environments

## Solution / What's New

Created a complete Auth0Client deployment component with full IaC support for all four Auth0 application types.

### Registry Allocation

Added to Auth0 provider block (2100–2299):

```protobuf
Auth0Client = 2101 [(kind_meta) = {
  provider: auth0
  version: v1
  id_prefix: "a0cli"
}];
```

### Supported Application Types

| Type | Use Case | OAuth Flow |
|------|----------|------------|
| `native` | Mobile/Desktop apps | Authorization Code + PKCE |
| `spa` | Single Page Applications | Authorization Code + PKCE |
| `regular_web` | Server-side web apps | Authorization Code |
| `non_interactive` | M2M/Backend services | Client Credentials |

## Implementation Details

### Proto API (4 files)

**spec.proto** - Comprehensive application specification:
- Application type with validation
- OAuth configuration (callbacks, logout URLs, grant types, origins)
- JWT configuration (lifetime, algorithm, custom scopes)
- Refresh token settings (rotation, expiration, idle timeout)
- Mobile configuration (iOS Team ID, Android package)
- Native social login (Apple, Facebook)
- Organization support (multi-tenant apps)
- OIDC backchannel logout

**stack_outputs.proto** - Deployment outputs:
- Client ID (public OAuth identifier)
- Application name and type
- Signing keys for token verification

**api.proto** - KRM envelope:
- apiVersion: `auth0.planton.dev/v1`
- kind: `Auth0Client`

**stack_input.proto** - IaC module inputs

### Pulumi Module (Go)

```
iac/pulumi/
├── main.go           # Entry point
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Computed values from spec
│   ├── client.go     # Client creation with all options
│   └── outputs.go    # Stack output exports
└── Makefile, Pulumi.yaml, debug.sh
```

### Terraform Module (HCL)

```
iac/tf/
├── provider.tf   # Auth0 provider configuration
├── variables.tf  # Input variables mirroring spec.proto
├── locals.tf     # Computed local values
├── main.tf       # Client resource with dynamic blocks
└── outputs.tf    # Output definitions
```

### Validation Tests (28 cases)

spec_test.go covers:
- Valid configurations for all 4 application types
- Required field validation (application_type, metadata, spec)
- Enum validation (JWT alg, rotation_type, organization_usage)
- Range validation (JWT lifetime: 0-2592000 seconds)
- Description length validation (max 140 chars)

## Benefits

### For Users
- **Declarative Application Management**: Configure OAuth apps via YAML
- **All App Types Supported**: Native, SPA, Web, and M2M
- **Mobile-Ready**: iOS and Android configuration built-in
- **Token Control**: JWT lifetime, refresh token rotation settings
- **Organization Support**: Multi-tenant B2B SaaS scenarios

### For Developers
- **Pattern Consistency**: Follows established Auth0Connection patterns
- **Full Documentation**: README, examples.md, research docs
- **Dual IaC Support**: Both Pulumi and Terraform with feature parity
- **Comprehensive Testing**: 28 validation test cases

## Impact

### Direct
- Users can now manage Auth0 Applications through Planton
- CLI supports Auth0Client manifests with `--auth0-provider-config` flag
- Auth0 provider now has two deployment components

### Registry
- Auth0 range: 2100–2299 (198 slots remaining)
- Second Auth0 component registered (2101)

### Future Work Enabled
- Auth0Role component (RBAC)
- Auth0Action component (Auth0 Actions)
- Auth0ResourceServer component (APIs)
- Auth0Organization component (multi-tenant)

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

**Total**: ~35 files, ~3500 lines of code

## Usage Examples

### Single Page Application

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Client
metadata:
  name: my-spa-app
spec:
  application_type: spa
  description: My React SPA
  callbacks:
    - https://myapp.com/callback
  web_origins:
    - https://myapp.com
  grant_types:
    - authorization_code
    - refresh_token
  refresh_token:
    rotation_type: rotating
    expiration_type: expiring
```

### Machine-to-Machine Application

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Client
metadata:
  name: backend-api-client
spec:
  application_type: non_interactive
  description: Backend service
  grant_types:
    - client_credentials
  jwt_configuration:
    lifetime_in_seconds: 86400
```

### Native Mobile App

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Client
metadata:
  name: mobile-app
spec:
  application_type: native
  callbacks:
    - com.example.myapp://callback
  mobile:
    ios:
      team_id: ABCDE12345
      app_bundle_identifier: com.example.myapp
    android:
      app_package_name: com.example.myapp
  native_social_login:
    apple:
      enabled: true
```

## Related Work

- Builds on Auth0 provider integration (2025-12-30-054629-auth0-provider-integration.md)
- Companion to Auth0Connection component (2025-12-30-063818-auth0connection-deployment-component.md)
- Follows patterns from other SaaS providers (Confluent, Atlas, Snowflake)

---

**Status**: ✅ Production Ready
**Build**: CLI compiles, 28/28 tests pass, Terraform validates


