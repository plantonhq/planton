# GCP Identity Platform Tenant - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying an Identity Platform tenant using Planton's `GcpIdentityPlatformTenant` API. The module is written in Go and creates `identityplatform.Tenant` — an isolated user pool with its own sign-in configuration — plus one composed resource per identity provider the spec lists: `identityplatform.TenantDefaultSupportedIdpConfig`, `identityplatform.TenantOauthIdpConfig`, and `identityplatform.TenantInboundSamlConfig`.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **Identity Platform initialized on the project with `multi_tenant.allow_tenants = true`** (the `GcpIdentityPlatformConfig` kind) — the tenant API rejects creation otherwise
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: `roles/identityplatform.admin` (or broader) on the target project

## Directory Structure

```
iac/pulumi/
├── main.go                        # Pulumi program entry point
├── Pulumi.yaml                    # Pulumi project configuration
├── README.md                      # This file
└── module/
    ├── main.go                    # Module coordinator
    ├── identity_platform_tenant.go # Tenant + composed IdP configs
    ├── locals.go                  # Resolved resource
    └── outputs.go                 # Output constants
```

## Quick Start

```bash
cd iac/pulumi
pulumi stack init dev
```

Provide a `iac-input.yaml`:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpIdentityPlatformTenant
  metadata:
    name: acme-corp
  spec:
    display_name: acme-corp
    allow_password_signup: true
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpIdentityPlatformTenantIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpIdentityPlatformTenant` spec (display name, sign-in switches, IdPs) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `tenant_id` | string | The server-generated tenant ID (what client SDKs set as tenantId) |
| `tenant_name` | string | `projects/{project}/tenants/{tenant_id}` |

## Behavior Notes

- **The tenant ID is server-generated** — `display_name` is the only naming input; the IdP configs receive the generated ID through the created resource, never from the spec.
- **Tenant-level API differences honored**: OIDC `display_name` and the SAML `sp_config` (both fields) are REQUIRED at tenant level, and tenant OIDC has no `response_type` selection.
- **One deletion_policy governs everything**: the tenant and every composed IdP config — deleting a tenant deletes ALL its users, unrecoverable.
- **IdP `enabled` flags are sent explicitly** so a `true -> false` transition reaches the API instead of being omitted (the silent-no-op class).
- **API enablement**: the module enables `identitytoolkit.googleapis.com` (with `disable_on_destroy=false`) — a no-op on projects the config kind already initialized.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
