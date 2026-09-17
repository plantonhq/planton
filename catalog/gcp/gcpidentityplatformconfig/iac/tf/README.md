# GcpIdentityPlatformConfig - Terraform Module

This Terraform module provisions a project's Identity Platform configuration (`google_identity_platform_config`) plus one composed resource per identity provider the spec lists (`google_identity_platform_default_supported_idp_config`, `google_identity_platform_oauth_idp_config`, `google_identity_platform_inbound_saml_config`). It is the Terraform-side implementation of the Planton `GcpIdentityPlatformConfig` resource kind and has feature parity with the Pulumi module.

## Overview

The config resource is a ONE-WAY project singleton: the first apply permanently initializes Identity Platform on the project (billing required), and destroy abandons the configuration in place — GCP has no de-initialize. Every setting stays freely updatable after initialization. The composed IdP configs carry the spec's `deletion_policy`; the config resource itself is undeletable and carries none (provider truth). The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpidentityplatformconfig/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpIdentityPlatformConfig spec | — |

The `spec` object includes: `sign_in` (email/phone/anonymous arms with explicit-send `enabled` flags), `authorized_domains`, `mfa` (state, PHONE_SMS, TOTP provider configs), `blocking_functions` (beforeCreate/beforeSignIn triggers → Cloud Function URIs), `sign_up_quota` (all three fields together), `sms_region_config` (exactly one of allow_by_default/allowlist_only), `client_permissions`, `request_logging_enabled`, `multi_tenant` (allow_tenants gates the tenant kind), `autodelete_anonymous_users`, the three IdP lists (`default_supported_idps`, `oauth_idp_configs` with `oidc.` names, `inbound_saml_configs` with `saml.` names), `project_id` (empty falls back to the provider default project), and `deletion_policy` (DELETE/PREVENT/ABANDON — composed IdP configs only).

## Outputs

| Name | Description |
|------|-------------|
| `config_name` | `projects/{project}/config` |
| `api_key` | The auto-provisioned client SDK API key (sensitive) |
| `firebase_subdomain` | The project's default hosted sign-in domain |
