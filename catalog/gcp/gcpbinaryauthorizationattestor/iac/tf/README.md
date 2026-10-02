# GcpBinaryAuthorizationAttestor — Terraform Implementation

This directory contains the Terraform implementation for a Binary Authorization attestor from the Planton spec: `google_project_service` for the Binary Authorization and Artifact Analysis APIs, an optional `google_container_analysis_note` with its `google_container_analysis_note_iam_member` grant, the `google_kms_crypto_key_version` reads for Cloud KMS keys, and one `google_binary_authorization_attestor`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | The project, the attestor and note names, the Cloud KMS key versions by key position |
| `main.tf` | `data.google_client_config` (empty project only), the APIs, the note, the key-version reads, the attestor, the note grant |
| `outputs.tf` | `attestor_id`, `attestor_name`, `note_reference`, `delegation_service_account_email` |

## Send Posture

- **`project`** -- the spec's project, or the provider's project from `google_client_config` (provider configuration, no API call) -- PARITY with the Pulumi module's `GetClientConfig`.
- **`name`** -- `attestor_name`, defaulting to `metadata.name`; the created note defaults to `{attestor_name}-note`.
- **`note_reference`** -- the created note's ID, or the spec's reference.
- **Cloud KMS keys** -- `public_key_pem` and `signature_algorithm` read from `data.google_kms_crypto_key_version`; `id` defaults to the data source's `//cloudkms.googleapis.com/v1/...` ID.
- **`deletion_policy`** -- sent only when set, to the attestor and the created note.
- **Note grant** -- `roles/containeranalysis.notes.occurrences.viewer` for `delegation_service_account_email`, created note only.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
