---
title: "Application Credential"
description: "Application Credential deployment documentation"
icon: "package"
order: 100
componentName: "openstackapplicationcredential"
---

# OpenStack Application Credential

Deploys an OpenStack Identity (Keystone) application credential, providing a scoped authentication token that allows applications to authenticate to OpenStack without using a user's password. This is an immutable resource -- any change to the spec destroys and recreates the credential, generating a new secret.

## What Gets Created

When you deploy an OpenStackApplicationCredential resource, Planton provisions:

- **Identity Application Credential** — an `openstack_identity_application_credential_v3` resource scoped to the project that is active during creation, with optional role restrictions and fine-grained API access rules

## Prerequisites

- **OpenStack credentials** configured via environment variables or Planton provider config
- **Keystone v3 API** available in the target OpenStack deployment
- **Appropriate roles** assigned to the authenticating user on the target project (the credential inherits from these roles)

## Quick Start

Create a file `app-credential.yaml`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackApplicationCredential
metadata:
  name: my-app-cred
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: dev.OpenstackApplicationCredential.my-app-cred
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackapplicationcredential/v1/iac/pulumi/module
spec: {}
```

Deploy:

```shell
planton apply -f app-credential.yaml
```

This creates an application credential named `my-app-cred` with an auto-generated secret, inheriting all roles of the creating user on the current project. The secret is available in `status.outputs.secret` after deployment and cannot be retrieved again from the OpenStack API.

## Configuration Reference

### Required Fields

All spec fields are optional. The credential name is derived from `metadata.name`.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the application credential. |
| `unrestricted` | `bool` | `false` | When `true`, the credential can create additional application credentials or trusts. This is a security risk and should be used with caution. |
| `secret` | `string` | auto-generated | User-provided secret for the credential. If omitted, OpenStack generates a random secret. Sensitive. |
| `roles` | `string[]` | all user roles | List of role names to scope the credential. The credential can only perform actions allowed by these roles. If omitted, inherits all roles of the creating user on the current project. |
| `accessRules` | `AccessRule[]` | — | Fine-grained API access restrictions. When set, the credential can only call the specified APIs. See Access Rule fields below. |
| `expiresAt` | `string` | — | Expiration timestamp in RFC 3339 format (e.g., `2027-01-01T00:00:00Z`). After this time, the credential becomes invalid. If omitted, the credential does not expire. |
| `region` | `string` | provider default | Overrides the region from the provider config for this credential. |

**Access Rule fields** (each element in `accessRules`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `path` | `string` | yes | URL path pattern for the API endpoint. Supports wildcards (e.g., `/v2.1/servers/*`). |
| `method` | `string` | yes | HTTP method allowed for this rule. Must be one of: `POST`, `GET`, `HEAD`, `PATCH`, `PUT`, `DELETE`. |
| `service` | `string` | yes | OpenStack service type (e.g., `identity`, `compute`, `block-storage`, `image`, `network`). |

## Examples

### Basic Application Credential

A credential with default settings, suitable for development automation:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackApplicationCredential
metadata:
  name: dev-automation
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: dev.OpenstackApplicationCredential.dev-automation
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackapplicationcredential/v1/iac/pulumi/module
spec:
  description: CI/CD automation credential for development
```

### Role-Scoped Credential with Expiration

A credential restricted to specific roles with a defined lifetime, suitable for temporary access:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackApplicationCredential
metadata:
  name: temp-reader
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: staging.OpenstackApplicationCredential.temp-reader
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackapplicationcredential/v1/iac/pulumi/module
spec:
  description: Temporary read-only credential for audit tooling
  roles:
    - reader
    - load-balancer_member
  expiresAt: "2027-06-01T00:00:00Z"
```

### Fine-Grained Access Rules

A credential locked down to specific API operations, allowing only compute server listing and identity project listing:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackApplicationCredential
metadata:
  name: monitoring-agent
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: prod.OpenstackApplicationCredential.monitoring-agent
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackapplicationcredential/v1/iac/pulumi/module
spec:
  description: Monitoring agent with read-only access to compute and identity APIs
  roles:
    - reader
  accessRules:
    - service: compute
      method: GET
      path: /v2.1/servers/*
    - service: identity
      method: GET
      path: /v3/projects
    - service: compute
      method: GET
      path: /v2.1/flavors/*
  expiresAt: "2027-12-31T23:59:59Z"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | UUID of the application credential in Keystone |
| `name` | `string` | Name of the credential, derived from `metadata.name` |
| `secret` | `string` | The credential secret (sensitive). Generated once at creation time and cannot be retrieved again from the OpenStack API. |
| `projectId` | `string` | UUID of the project this credential is scoped to, computed from the authentication scope used during creation |
| `region` | `string` | OpenStack region where the credential was created |

## Related Components

- [OpenStackProject](/docs/catalog/openstack/project) — the project scope that the application credential is bound to
- [OpenStackRoleAssignment](/docs/catalog/openstack/role-assignment) — manages role assignments that determine what actions the credential can perform
