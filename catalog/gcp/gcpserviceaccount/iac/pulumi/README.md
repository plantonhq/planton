# GcpServiceAccount - Pulumi Module

This Pulumi (Go) module provisions a Google Cloud service account (`serviceaccount.Account`) with an optional user-managed key (`serviceaccount.Key`) and additive project/organization role grants (`projects.IAMMember` / `organizations.IAMMember`). It is the Pulumi-side implementation of the Planton `GcpServiceAccount` resource kind and has feature parity with the Terraform module.

## Overview

`display_name` falls back to `metadata.name` so every account is human-identifiable in the console. An empty `project_id` leaves the provider to resolve its own project from the ambient configuration (`GOOGLE_PROJECT` / `GOOGLE_CLOUD_PROJECT`). Role grants use the ADDITIVE member resources: each grant manages exactly one (role, member) pair and never clobbers other members' bindings on the same role.

**Keys are opt-in and shaped, not just toggled**: the `user_managed_key` message's presence creates a key; its fields choose the generate flow (algorithm, private/public key formats — the private key exports once as the secret `key_base64` output) or the upload flow (`public_key_data` — GCP never sees a private key and `key_base64` stays empty). `keepers` is the idiomatic rotation trigger: changing any value replaces the key. Keyless patterns (Workload Identity, impersonation, federation) remain the recommended default.

**Sharp edges**: `account_id` and `project` are immutable — changing either replaces the account and invalidates every IAM binding and workload identity referencing the old email. `deletion_policy: PREVENT` fails destroys — a guard rail for identities whose deletion would break bindings fleet-wide. `create_ignore_already_exists` adopts an existing account of the same email instead of failing (idempotent bootstrap). IAM conditions on the role lists are deliberately not modeled — conditioned, first-class grants belong to the `GcpProjectIamMember` kind, which references this account's `member` output.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `main.go` — Pulumi entrypoint (loads the IaC input, calls the module)
- `module/main.go` — provider setup + orchestration
- `module/locals.go` — iac-input container
- `module/service_account.go` — the account + the optional key
- `module/iam.go` — additive project/organization role grants
- `module/outputs.go` — output key constants

## Outputs

| Output | Description |
|--------|-------------|
| `email` | Service account email — the identity handle workloads attach by |
| `member` | IAM member string (`serviceAccount:<email>`) — feed directly into IAM grants |
| `unique_id` | Stable numeric ID, never reused across delete/recreate |
| `name` | Fully-qualified resource name (`projects/.../serviceAccounts/<email>`) |
| `key_base64` | Base64 private key (generate flow only; secret) |
