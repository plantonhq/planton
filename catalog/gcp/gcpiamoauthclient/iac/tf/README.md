# GcpIamOauthClient - Terraform Module

This Terraform module provisions a workforce OAuth client (`google_iam_oauth_client`) plus one managed credential (`google_iam_oauth_client_credential`) per `spec.credentials` entry. It is the Terraform-side implementation of the Planton `GcpIamOauthClient` resource kind and has feature parity with the Pulumi module.

## Overview

The client is a Workforce Identity Federation OAuth registration — the only kind of OAuth client Google's APIs can create programmatically (consent-screen clients remain a console step; see the component README). Credential secrets are generated server-side by GCP; the first credential's secret is the `client_secret` output. GCP requires a credential to be DISABLED before it can be deleted, so `disabled` is sent explicitly on every apply. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpiamoauthclient/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpIamOauthClient spec | — |

The `spec` object includes: `oauth_client_id` (empty defaults to `metadata.name`), `location` (empty defaults to `global`), `client_type` (PUBLIC_CLIENT/CONFIDENTIAL_CLIENT), `allowed_grant_types` (AUTHORIZATION_CODE_GRANT/REFRESH_TOKEN_GRANT — the API's closed enum), `allowed_scopes`, `allowed_redirect_uris` (accepts references to other resources' URL outputs), `credentials` (credential_id/display_name/disabled), `display_name`, `description`, `disabled`, `project_id` (empty falls back to the provider default project), and `deletion_policy` (DELETE/PREVENT/ABANDON — one switch for the client and its credentials).

## Outputs

| Name | Description |
|------|-------------|
| `client_id` | The system-generated OAuth client ID applications present |
| `client_name` | `projects/{project}/locations/{location}/oauthClients/{id}` |
| `state` | The client's lifecycle state |
| `client_secret` | The FIRST credential's system-generated secret (sensitive; empty when no credentials) |
