# Auth0 Provider Integration

**Date**: December 30, 2025
**Type**: Feature
**Components**: Provider Framework, CLI Flags, API Definitions, Backend Services, Frontend Credentials, Tofu Integration

## Summary

Added Auth0 as a new cloud provider to Planton, enabling users to manage Auth0 identity resources through the platform. This implementation spans all layers of the system—from protobuf definitions through CLI flags, stack input processing, Tofu/Terraform environment configuration, backend credential management, and frontend credential forms—maintaining full consistency with existing provider patterns.

## Problem Statement / Motivation

Planton needed to expand its provider ecosystem to include Auth0, a popular identity platform used for authentication and authorization. Without Auth0 support, users could not:

- Store and manage Auth0 credentials through the platform
- Use Auth0 credentials for infrastructure deployments
- Leverage the consistent provider credential workflow for Auth0 resources

### Pain Points

- No Auth0 provider in the `CloudResourceProvider` enum
- No credential storage or management for Auth0
- No CLI flags to pass Auth0 credentials during deployments
- No Tofu/Terraform environment variable configuration for Auth0
- No frontend UI for capturing Auth0 credentials

## Solution / What's New

Implemented comprehensive Auth0 provider support across all system layers, following the established patterns for existing providers (AWS, GCP, Azure, Confluent, Snowflake, Atlas, etc.).

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                       Proto Layer                                │
│  cloud_resource_provider.proto → auth0 = 21                     │
│  provider/auth0/provider.proto → Auth0ProviderConfig            │
│  credential/v1/api.proto → AUTH0 enum + oneof case              │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Layer                                 │
│  flag.go → Auth0ProviderConfig constant                         │
│  apply/plan/init/destroy/refresh.go → --auth0-provider-config   │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Stack Input Layer                             │
│  options.go → Auth0ProviderConfig field + builder functions     │
│  auth0_provider.go → Load/Add functions                         │
│  user_provider.go → createAuth0ProviderConfigFileFromProto      │
│  providers.go → AddAuth0ProviderConfig aggregation              │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Tofu/Terraform Layer                         │
│  auth0_provider.go → AUTH0_DOMAIN, AUTH0_CLIENT_ID,             │
│                      AUTH0_CLIENT_SECRET env vars                │
│  providers.go → AddAuth0ProviderConfigEnvVars                   │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Backend Layer                               │
│  credential.go → Auth0Credential model                          │
│  credential_repo.go → CreateAuth0, UpdateAuth0, converter       │
│  credential_service.go → Full CRUD operations                   │
│  credential_resolver.go → Auth0 credential resolution           │
│  stack_update_service.go → Auth0 config handling                │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Frontend Layer                               │
│  auth0.tsx → Auth0CredentialForm component                      │
│  types.ts → auth0 in CredentialFormData                         │
│  credential-drawer.tsx → Auth0 form integration                 │
│  utils.ts → Auth0 provider config                               │
└─────────────────────────────────────────────────────────────────┘
```

## Implementation Details

### 1. Proto Definitions

**CloudResourceProvider enum** (`cloud_resource_provider.proto`):
```protobuf
auth0 = 21 [(provider_meta) = {
  group: "auth0.planton.dev"
  display_name: "Auth0"
}];
```

**Auth0ProviderConfig message** (`provider/auth0/provider.proto`):
```protobuf
message Auth0ProviderConfig {
  string domain = 1 [(buf.validate.field).required = true];
  string client_id = 2 [(buf.validate.field).required = true];
  string client_secret = 3 [(buf.validate.field).required = true];
}
```

**Credential API** (`credential/v1/api.proto`):
- Added `AUTH0 = 4` to `CredentialProvider` enum
- Added `auth0 = 11` case to `CredentialProviderConfig` oneof

### 2. CLI Flag Registration

Added `--auth0-provider-config` flag to all deployment commands:

```go
// internal/cli/flag/flag.go
Auth0ProviderConfig Flag = "auth0-provider-config"

// cmd/planton/root/apply.go (and plan, init, destroy, refresh)
Apply.PersistentFlags().String(string(flag.Auth0ProviderConfig), "", 
    "path of the auth0-credential file")
```

### 3. Stack Input Processing

**StackInputProviderConfigOptions** now includes Auth0:
```go
type StackInputProviderConfigOptions struct {
    AtlasProviderConfig      string
    Auth0ProviderConfig      string  // New
    AwsProviderConfig        string
    // ... other providers
}
```

**New auth0_provider.go**:
- `AddAuth0ProviderConfig()` - Reads and adds Auth0 config to stack input
- `LoadAuth0ProviderConfig()` - Loads Auth0 config from input directory

### 4. Tofu/Terraform Environment Variables

**New auth0_provider.go** in tofumodule/providerconfig:
```go
func AddAuth0ProviderConfigEnvVars(stackInputContentMap map[string]interface{},
    providerConfigEnvVars map[string]string) (map[string]string, error) {
    // Sets AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET
}
```

### 5. Backend Credential Management

**Auth0Credential model**:
```go
type Auth0Credential struct {
    ID           primitive.ObjectID
    Name         string
    Domain       string
    ClientID     string
    ClientSecret string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Repository methods**: `CreateAuth0()`, `UpdateAuth0()`, `convertToAuth0Credential()`

**Service methods**: Full CRUD in `credential_service.go` with validation

**Credential resolver**: Auth0 case returns `Auth0ProviderConfig` from stored credentials

### 6. Frontend Integration

**Auth0CredentialForm component**:
```tsx
export function Auth0CredentialForm({ register, disabled }: Auth0CredentialFormProps) {
  return (
    <>
      <SimpleInput path="auth0.domain" name="Domain" ... />
      <SimpleInput path="auth0.clientId" name="Client ID" ... />
      <SimpleInput path="auth0.clientSecret" name="Client Secret" type="password" ... />
    </>
  );
}
```

**Provider config utility**:
```typescript
[Credential_CredentialProvider.AUTH0]: {
  label: 'Auth0',
  description: 'Link your Auth0 tenant to manage identity resources',
  icon: undefined,
}
```

## Files Changed

| Layer | Files |
|-------|-------|
| Proto | `cloud_resource_provider.proto`, `provider/auth0/provider.proto`, `credential/v1/api.proto` |
| CLI | `flag.go`, `apply.go`, `plan.go`, `init.go`, `destroy.go`, `refresh.go` |
| Stack Input | `options.go`, `auth0_provider.go` (new), `user_provider.go`, `providers.go` |
| Tofu | `auth0_provider.go` (new), `providers.go` |
| Backend | `credential.go`, `credential_repo.go`, `credential_service.go`, `credential_resolver.go`, `stack_update_service.go` |
| Frontend | `auth0.tsx` (new), `types.ts`, `index.ts`, `credential-drawer.tsx`, `utils.ts` |

**Total**: 25+ files across 6 system layers

## Benefits

### For Users
- **Credential Management**: Store Auth0 credentials securely through the web UI
- **CLI Integration**: Pass Auth0 credentials via `--auth0-provider-config` flag
- **Consistent Workflow**: Same credential flow as AWS, GCP, Azure, etc.

### For Developers
- **Pattern Consistency**: Auth0 implementation follows established provider patterns
- **Full Stack Coverage**: All layers updated consistently
- **Ready for Resources**: Foundation laid for Auth0Connection and other resources

### For Operations
- **Tofu/Terraform Support**: Environment variables automatically configured
- **Backend Resolution**: Credentials automatically resolved for deployments

## Impact

### Direct Impact
- Users can now manage Auth0 credentials through Planton
- CLI supports `--auth0-provider-config` flag on all deployment commands
- Backend API supports Auth0 credential CRUD operations
- Frontend displays Auth0 in credential provider dropdown

### Future Work Enabled
- Auth0Connection deployment component
- Auth0 resource types (applications, APIs, rules, etc.)
- Auth0 tenant management workflows

## Usage Examples

### CLI Usage
```bash
# Create Auth0 credential file
cat > auth0-creds.yaml << EOF
domain: your-tenant.auth0.com
clientId: your-client-id
clientSecret: your-client-secret
EOF

# Use with deployment
planton apply --manifest auth0-connection.yaml \
  --auth0-provider-config auth0-creds.yaml
```

### Web UI
1. Navigate to Credentials page
2. Click "Create Credential"
3. Select "Auth0" from provider dropdown
4. Enter Domain, Client ID, and Client Secret
5. Save credential

## Related Work

- This lays the foundation for Auth0Connection deployment component
- Follows the same patterns established by existing providers (Confluent, Atlas, Snowflake)
- Integrates with the credential resolution system for automated deployments

---

**Status**: ✅ Production Ready
**Build**: CLI, Backend, and Frontend all compile successfully

