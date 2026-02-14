---
title: "Zero Trust Access Application"
description: "Zero Trust Access Application deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarezerotrustaccessapplication"
---

# Cloudflare Zero Trust Access Application

Deploys a Cloudflare Zero Trust Access Application that protects a hostname behind identity-aware access controls. The component creates a self-hosted Access Application for a given DNS zone and attaches an Access Policy with configurable email allowlists, Google Workspace group restrictions, session duration, and optional multi-factor authentication enforcement.

## What Gets Created

When you deploy a CloudflareZeroTrustAccessApplication resource, OpenMCF provisions:

- **Access Application** — a `cloudflare_access_application` of type `self_hosted`, bound to the specified DNS zone and hostname, with an optional custom session duration
- **Access Policy** — a `cloudflare_access_policy` attached to the application with an `allow` or `deny` decision, email-based include rules, optional Google Workspace group includes, and an optional MFA requirement

## Prerequisites

- **Cloudflare credentials** configured via environment variables or OpenMCF provider config
- **A Cloudflare DNS zone** already provisioned (the `zoneId` is required); you can obtain this from a CloudflareDnsZone resource's `status.outputs.zone_id`
- **DNS configured** — the hostname you intend to protect must resolve within the specified zone
- **Cloudflare Zero Trust subscription** — Access Applications require a Zero Trust plan on the Cloudflare account

## Quick Start

Create a file `access-app.yaml`:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: my-access-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareZeroTrustAccessApplication.my-access-app
spec:
  applicationName: My Internal App
  zoneId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  hostname: app.example.com
  allowedEmails:
    - alice@example.com
    - bob@example.com
```

Deploy:

```shell
openmcf apply -f access-app.yaml
```

This creates a Zero Trust Access Application protecting `app.example.com` with an allow policy that grants access only to `alice@example.com` and `bob@example.com`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `applicationName` | `string` | Display name of the Zero Trust Access Application shown in the Cloudflare dashboard. | Minimum 1 character |
| `zoneId` | `string` | Cloudflare DNS zone ID for the domain. Can reference a CloudflareDnsZone resource via `status.outputs.zone_id`. | Required, non-empty |
| `hostname` | `string` | Fully qualified domain name to protect (e.g., `app.example.com`). | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `policyType` | `enum` | `ALLOW` | Access policy decision. `ALLOW` permits matching identities; `BLOCK` denies them. |
| `allowedEmails` | `string[]` | `[]` | Email addresses permitted to access the application. Each email is added as a separate include rule on the policy. Applicable when `policyType` is `ALLOW`. |
| `sessionDurationMinutes` | `int32` | `1440` | Duration of each authenticated session in minutes. Default is 1440 (24 hours). When set, the value is passed to Cloudflare in `{N}m` format. |
| `requireMfa` | `bool` | `false` | When `true`, adds an MFA requirement to the access policy. Users must complete a second authentication factor to gain access. |
| `allowedGoogleGroups` | `string[]` | `[]` | Google Workspace group IDs to include in the access policy. Each group is added as a separate include rule. |

## Examples

### Email-Only Allow Policy

Restrict access to a staging application to a small list of team members:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: staging-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.CloudflareZeroTrustAccessApplication.staging-app
spec:
  applicationName: Staging Dashboard
  zoneId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  hostname: staging.example.com
  allowedEmails:
    - dev1@example.com
    - dev2@example.com
    - qa@example.com
```

### MFA-Enforced Production Application

Protect a production admin panel with mandatory multi-factor authentication and a shorter session window:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: admin-panel
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareZeroTrustAccessApplication.admin-panel
spec:
  applicationName: Admin Panel
  zoneId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  hostname: admin.example.com
  allowedEmails:
    - ops-lead@example.com
    - security@example.com
  sessionDurationMinutes: 480
  requireMfa: true
```

### Google Workspace Group Access with Block Policy

Grant access to a Google Workspace group while blocking specific email addresses:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: internal-wiki
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareZeroTrustAccessApplication.internal-wiki
spec:
  applicationName: Internal Wiki
  zoneId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  hostname: wiki.example.com
  allowedGoogleGroups:
    - a1b2c3d4-e5f6-7890-abcd-ef1234567890
  allowedEmails:
    - contractor@partner.com
  sessionDurationMinutes: 720
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `application_id` | `string` | The unique Cloudflare ID of the Access Application |
| `public_hostname` | `string` | The hostname protected by this Access Application (echoes the input `hostname`) |
| `policy_id` | `string` | The Cloudflare ID of the Access Policy attached to this application |

## Related Components

- [CloudflareDnsZone](/docs/catalog/cloudflare/cloudflarednszone) — provisions the DNS zone whose `zone_id` is referenced by this component
- [CloudflareDnsRecord](/docs/catalog/cloudflare/cloudflarednsrecord) — manages DNS records that route traffic to the hostname protected by this Access Application
- [CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker) — can serve content behind the Access Application for authenticated users
