# Auth0EventStream Deployment Component

**Date**: January 6, 2026
**Type**: Feature
**Components**: Provider Framework, API Definitions, Pulumi IaC Module, Terraform IaC Module, Proto Validations

## Summary

Implemented `Auth0EventStream` as the third deployment component for the Auth0 provider. This component enables declarative management of Auth0 Event Streams, supporting both AWS EventBridge and custom webhook destinations for real-time event delivery.

## Problem Statement / Motivation

Auth0 Event Streams enable real-time delivery of tenant events to external systems. Without this component, users could not:

- Stream authentication events to SIEM systems for security monitoring
- Trigger workflows when users are created or updated
- Build real-time dashboards for login analytics
- Integrate Auth0 events with AWS Lambda for serverless processing

### Pain Points

- No way to manage Auth0 Event Streams through Planton
- Manual Auth0 dashboard configuration lacks version control
- Need consistent patterns matching other Auth0 components (Auth0Connection, Auth0Client)
- EventBridge and webhook configurations require different handling

## Solution / What's New

Created a complete Auth0EventStream deployment component following the established Auth0 provider patterns with full IaC support for both EventBridge and webhook destinations.

### Registry Allocation

Added to Auth0 provider block (2100–2299):

```protobuf
Auth0EventStream = 2102 [(kind_meta) = {
  provider: auth0
  version: v1
  id_prefix: "a0es"
}];
```

### Supported Destination Types

| Destination | Description | Update Policy |
|-------------|-------------|---------------|
| `eventbridge` | AWS EventBridge for serverless event processing | Immutable - Recreate to change |
| `webhook` | Custom HTTPS endpoint for event delivery | Mutable - Can update after creation |

### Key Features

- **Dual Destination Support**: EventBridge for AWS-native architectures, webhooks for custom integrations
- **Flexible Subscriptions**: Subscribe to any combination of Auth0 event types
- **Secure Authorization**: Bearer token and Basic authentication for webhooks
- **AWS Account/Region Configuration**: Full control over EventBridge destination
- **Comprehensive Validation**: AWS account ID format, HTTPS requirement, authorization method validation

## Implementation Details

### Proto API (4 files)

**spec.proto** - Complete event stream specification:
- `Auth0EventStreamSpec` - Core configuration with destination type and subscriptions
- `Auth0EventBridgeConfiguration` - AWS account ID and region with validation
- `Auth0WebhookConfiguration` - Endpoint URL and authorization settings
- `Auth0WebhookAuthorization` - Basic and bearer authentication support

**stack_outputs.proto** - Deployment outputs:
- Event stream ID, name, status
- Destination type and subscriptions
- Timestamps (created_at, updated_at)
- AWS partner event source (for EventBridge)

**api.proto** - KRM envelope:
- apiVersion: `auth0.planton.dev/v1`
- kind: `Auth0EventStream`

**stack_input.proto** - IaC module inputs

### Pulumi Module (Go)

```
iac/pulumi/
├── main.go           # Entry point
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Computed values from spec
│   ├── eventstream.go# Event stream creation with destination routing
│   └── outputs.go    # Stack output exports
└── Makefile, Pulumi.yaml, debug.sh, overview.md
```

### Terraform Module (HCL)

```
iac/tf/
├── provider.tf   # Auth0 provider configuration
├── variables.tf  # Input variables mirroring spec.proto
├── locals.tf     # Computed local values
├── main.tf       # Event stream resource with dynamic blocks
└── outputs.tf    # Output definitions
```

### Validation Tests (23 cases)

spec_test.go covers:
- Valid configurations for EventBridge and webhook destinations
- Multiple AWS regions (us-east-1, eu-west-1, ap-southeast-1)
- Bearer and Basic authentication for webhooks
- Required field validation (destination_type, subscriptions)
- AWS account ID format validation (12 digits)
- HTTPS requirement for webhook endpoints
- Authorization method validation (basic, bearer)

## Benefits

### For Security Teams
- **Real-time Monitoring**: Stream auth events to SIEM systems
- **Audit Compliance**: Track all authentication activity
- **Threat Detection**: Monitor failed login attempts

### For Platform Engineers
- **Declarative Configuration**: Event streams as code
- **AWS Integration**: Native EventBridge support
- **Flexible Webhooks**: Any HTTPS endpoint

### For DevOps
- **Environment Consistency**: Same manifests across dev/staging/prod
- **Version Control**: Git-tracked event stream configuration
- **Automation**: CLI-based deployment workflows

## Impact

### Direct
- Users can manage Auth0 Event Streams through Planton
- CLI supports Auth0EventStream manifests with `--auth0-provider-config` flag
- Auth0 provider now has three deployment components

### Registry
- Auth0 range: 2100–2299 (197 slots remaining)
- Third Auth0 component registered (2102)

### Future Work Enabled
- Auth0Role component (RBAC)
- Auth0Action component (Auth0 Actions)
- Auth0ResourceServer component (APIs)
- Auth0Organization component (multi-tenant)

## Files Created

| Category | Files |
|----------|-------|
| Registry | `cloud_resource_kind.proto` (updated) |
| Proto API | `spec.proto`, `api.proto`, `stack_input.proto`, `stack_outputs.proto` |
| Generated | `*.pb.go` (4 files) |
| Tests | `spec_test.go` |
| Pulumi | `main.go`, `module/*.go`, `Makefile`, `Pulumi.yaml`, `debug.sh`, `overview.md`, `README.md` |
| Terraform | `provider.tf`, `variables.tf`, `locals.tf`, `main.tf`, `outputs.tf`, `README.md` |
| Docs | `README.md`, `examples.md`, `docs/README.md` |
| Supporting | `hack/manifest.yaml` |

**Total**: ~30 files, ~2500 lines of code

## Usage Examples

### EventBridge Destination

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0EventStream
metadata:
  name: security-events
  org: my-organization
  env: production
spec:
  destination_type: eventbridge
  subscriptions:
    - user.created
    - user.updated
    - authentication.success
    - authentication.failure
  eventbridge_configuration:
    aws_account_id: "123456789012"
    aws_region: us-east-1
```

### Webhook Destination with Bearer Token

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0EventStream
metadata:
  name: user-events-webhook
  org: my-organization
  env: production
spec:
  destination_type: webhook
  subscriptions:
    - user.created
    - user.updated
    - user.deleted
  webhook_configuration:
    webhook_endpoint: "https://api.example.com/webhooks/auth0"
    webhook_authorization:
      method: bearer
      token: "your-secret-token"
```

### Webhook Destination with Basic Auth

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0EventStream
metadata:
  name: siem-events
spec:
  destination_type: webhook
  subscriptions:
    - authentication.failure
  webhook_configuration:
    webhook_endpoint: "https://siem.example.com/auth0/events"
    webhook_authorization:
      method: basic
      username: webhook-user
      password: super-secret-password
```

## Related Work

- Builds on Auth0 provider integration (2025-12-30-054629-auth0-provider-integration.md)
- Companion to Auth0Connection component (2025-12-30-063818-auth0connection-deployment-component.md)
- Companion to Auth0Client component (2025-12-30-070305-auth0client-deployment-component.md)
- Uses patterns from Pulumi Auth0 SDK v3 and Terraform Auth0 provider v1

---

**Status**: ✅ Production Ready
**Build**: CLI compiles, 23/23 tests pass

