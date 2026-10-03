# GcpServiceAccount - Terraform Module

This Terraform/OpenTofu module provisions a Google Cloud service account (`google_service_account`) with an optional user-managed key (`google_service_account_key`) and additive project/organization role grants (`google_project_iam_member` / `google_organization_iam_member`). It is the Terraform-side implementation of the Planton `GcpServiceAccount` resource kind and has feature parity with the Pulumi module.

## Overview

`display_name` falls back to `metadata.name` so every account is human-identifiable in the console. An empty `project_id` is normalized to `null` so the provider resolves its own project from the ambient configuration (`GOOGLE_PROJECT` / `GOOGLE_CLOUD_PROJECT`). Role grants use the ADDITIVE member resources: each grant manages exactly one (role, member) pair and never clobbers other members' bindings on the same role.

**Keys are opt-in and shaped, not just toggled**: the `user_managed_key` object's presence creates a key; its fields choose the generate flow (algorithm, private/public key formats — the private key exports once as the sensitive `key_base64` output) or the upload flow (`public_key_data` — GCP never sees a private key and `key_base64` stays empty). `keepers` is the idiomatic rotation trigger: changing any value replaces the key. Keyless patterns (Workload Identity, impersonation, federation) remain the recommended default.

**Sharp edges**: `account_id` and `project` are ForceNew — changing either replaces the account and invalidates every IAM binding and workload identity referencing the old email. `deletion_policy: PREVENT` fails destroys — a guard rail for identities whose deletion would break bindings fleet-wide. `create_ignore_already_exists` adopts an existing account of the same email instead of failing (idempotent bootstrap). IAM conditions on the role lists are deliberately not modeled — conditioned, first-class grants belong to the `GcpProjectIamMember` kind, which references this account's `member` output.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `provider.tf` — google provider pin
- `variables.tf` — the converter-contract `metadata`/`spec` variables
- `locals.tf` — project fallback, display-name fallback, role-list filtering
- `main.tf` — the account, the optional key, and the additive role grants
- `outputs.tf` — `email`, `member`, `unique_id`, `name`, `key_base64`

## Outputs

| Output | Description |
|--------|-------------|
| `email` | Service account email — the identity handle workloads attach by |
| `member` | IAM member string (`serviceAccount:<email>`) — feed directly into IAM grants |
| `unique_id` | Stable numeric ID, never reused across delete/recreate |
| `name` | Fully-qualified resource name (`projects/.../serviceAccounts/<email>`) |
| `key_base64` | Base64 private key (generate flow only; sensitive) |
