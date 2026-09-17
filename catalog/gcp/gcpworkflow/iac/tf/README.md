# GCP Workflow - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying a Cloud Workflows workflow using Planton's `GcpWorkflow` API. The module creates `google_workflows_workflow` — a serverless orchestrator executing YAML/JSON-defined steps. Every source / env-var / service-account change deploys a NEW revision.

## Prerequisites

1. **OpenTofu** (or Terraform >= 1.5)
2. **GCP Project** with the Workflows API enabled (the module enables it if needed)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Module Files

| File | Purpose |
|---|---|
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpWorkflow`), extended with the spec-default for `deletion_protection` |
| `locals.tf` | Project/region fallback + name derivation + label merge |
| `main.tf` | API enablement + the workflow resource |
| `outputs.tf` | Stack outputs |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `workflow_name` | `name` | Defaults to metadata.name; ForceNew |
| `region` | `region` | `null` when empty — the provider's default region applies; ForceNew |
| `description` | `description` | ≤1000 characters |
| `labels` | `labels` | Spec labels merged with platform attribution labels (platform wins) |
| `source_contents` | `source_contents` | REQUIRED (API truth); ≤128KB; every change mints a revision |
| `service_account` | `service_account` | References a GcpServiceAccount's `email` output |
| `crypto_key` | `crypto_key_name` | CMEK; references a GcpKmsKey's `key_id` output |
| `call_log_level` | `call_log_level` | Provider ValidateEnum values |
| `execution_history_level` | `execution_history_level` | Provider ValidateEnum values |
| `user_env_vars` | `user_env_vars` | ≤20 entries; keys must not start GOOGLE/WORKFLOWS |
| `resource_manager_tags` | `tags` | ForceNew — a tag change REPLACES the workflow |
| `deletion_protection` | `deletion_protection` | Sent EXPLICITLY on every apply — a true→false transition must reach the engine |
| `project_id` | `project` | `null` when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | `null` when empty (provider default DELETE) |

`name_prefix` is deliberately not modeled: versioned-name rotation is the
catalog pattern, and without hardcoded create-before-destroy semantics the
argument is a dead lever.

## Outputs

| Output | Description |
|---|---|
| `workflow_id` | Full resource name — the value Eventarc destinations consume |
| `workflow_name` | The short workflow name |
| `revision_id` | The deployed revision |
| `state` | Workflow state (`ACTIVE` when executable) |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
