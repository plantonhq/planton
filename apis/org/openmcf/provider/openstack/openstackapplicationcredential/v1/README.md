# OpenStackApplicationCredential

An OpenStack Identity (Keystone) application credential for passwordless automation.

## Overview

Application credentials allow applications and CI/CD pipelines to authenticate to OpenStack without using a user's password. They are scoped to the project that was active during creation and can be restricted to specific roles and fine-grained API access patterns.

## When to Use

- **CI/CD pipelines**: Create scoped credentials for deployment automation
- **Application authentication**: Allow services to authenticate without user passwords
- **Least-privilege access**: Restrict credentials to specific API operations via access rules

## Important Notes

- **Immutable resource**: ALL fields are ForceNew -- any change destroys and recreates the credential
- **Secret is generated once**: The secret is returned only at creation time and cannot be retrieved later
- **Project-scoped**: Automatically scoped to the project in the authentication context
- **unrestricted defaults to false**: Safe default; set to true only if sub-credential creation is needed

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable description |
| `unrestricted` | bool | Allow sub-credential creation (default: false) |
| `secret` | string | User-provided secret (auto-generated if omitted) |
| `roles` | list | Role names to scope the credential |
| `access_rules` | list | Fine-grained API access restrictions |
| `expires_at` | string | Expiration timestamp (RFC3339) |
| `region` | string | Region override |

## Outputs

| Output | Description |
|--------|-------------|
| `id` | Credential UUID |
| `name` | Credential name (from metadata.name) |
| `secret` | The credential secret (SENSITIVE) |
| `project_id` | Project the credential is scoped to |
| `region` | OpenStack region |

## Terraform Resource

`openstack_identity_application_credential_v3`

## Pulumi Resource

`openstack.identity.ApplicationCredential`
