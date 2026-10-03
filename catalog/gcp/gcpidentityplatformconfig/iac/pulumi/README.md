# GCP Identity Platform Config - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a project's Identity Platform configuration using Planton's `GcpIdentityPlatformConfig` API. The module is written in Go and creates `identityplatform.Config` — the project-singleton sign-in configuration — plus one composed resource per identity provider the spec lists: `identityplatform.DefaultSupportedIdpConfig` (Google, Facebook, Apple, ...), `identityplatform.OauthIdpConfig` (custom OIDC), and `identityplatform.InboundSamlConfig` (enterprise SAML SSO).

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project with BILLING enabled** — Identity Platform initialization fails without billing
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
    ├── identity_platform_config.go # Config + composed IdP configs
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
  kind: GcpIdentityPlatformConfig
  metadata:
    name: app-auth
  spec:
    sign_in:
      email:
        enabled: true
        password_required: true
    authorized_domains:
      - app.example.com
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpIdentityPlatformConfigIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpIdentityPlatformConfig` spec (sign-in methods, MFA, IdPs, quotas) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `config_name` | string | `projects/{project}/config` |
| `api_key` | string | The auto-provisioned client SDK API key (not a secret; ships inside client apps) |
| `firebase_subdomain` | string | The project's default hosted sign-in domain |

## Behavior Notes

- **One-way project singleton**: the FIRST apply permanently initializes Identity Platform on the project (the provider's create is a bare `initializeAuth` call followed by an update); GCP has no de-initialize, so destroy ABANDONS the configuration in place. Every setting stays freely updatable.
- **No deletion_policy on the config resource** (provider truth — the resource is undeletable); the spec's `deletion_policy` governs only the composed IdP configs.
- **Sign-in `enabled` flags are sent explicitly** whenever their arm is present so a `true -> false` transition reaches the API instead of being omitted (the silent-no-op class).
- **IdP `client_secret`s are secrets**: the platform stores them as managed-secret references and resolves them just-in-time at deploy; they come from each provider's own developer console (consent-screen clients have no programmatic creation path).
- **API enablement**: the module enables `identitytoolkit.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
