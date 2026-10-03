# GCP IAM OAuth Client - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a workforce OAuth client using Planton's `GcpIamOauthClient` API. The module is written in Go and creates `iam.OauthClient` — the Workforce Identity Federation OAuth registration — plus one `iam.OauthClientCredential` per `spec.credentials` entry, whose secrets GCP generates server-side.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the IAM API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: the deploying principal's permissions are listed in [`../permissions.yaml`](../permissions.yaml)

## Directory Structure

```
iac/pulumi/
├── main.go                # Pulumi program entry point
├── Pulumi.yaml            # Pulumi project configuration
├── README.md              # This file
└── module/
    ├── main.go            # Module coordinator
    ├── oauth_client.go    # Client + credential creation
    ├── locals.go          # Resolved resource + derived values
    └── outputs.go         # Output constants
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
  kind: GcpIamOauthClient
  metadata:
    name: my-app-client
  spec:
    client_type: CONFIDENTIAL_CLIENT
    allowed_grant_types:
      - AUTHORIZATION_CODE_GRANT
    allowed_scopes:
      - https://www.googleapis.com/auth/cloud-platform
    allowed_redirect_uris:
      - value: https://app.example.com/callback
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpIamOauthClientIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpIamOauthClient` spec (grant types, scopes, redirect URIs, credentials) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `client_id` | string | The system-generated OAuth client ID applications present |
| `client_name` | string | `projects/{project}/locations/{location}/oauthClients/{id}` |
| `state` | string | The client's lifecycle state |
| `client_secret` | string | The FIRST credential's system-generated secret (secret-marked in state; empty when no credentials) |

## Behavior Notes

- **Workforce clients only** (scope honesty): consent-screen OAuth clients have NO programmatic creation path anywhere since Google shut the IAP OAuth Admin API — see the component README.
- **`credential.disabled` is sent explicitly**: GCP requires a credential to be DISABLED before deletion, so the `false -> true` transition is exactly the pre-removal step and must reach the API. Removing an enabled credential fails at the API — disable in one apply, remove in the next.
- **The credential secret is provider-computed** and arrives already secret-marked from the SDK; the export needs no extra wrapping.
- **One deletion_policy** governs the client and every credential.
- **API enablement**: the module enables `iam.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
