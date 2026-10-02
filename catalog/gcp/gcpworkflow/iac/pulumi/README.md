# GCP Workflow - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Workflows workflow using Planton's `GcpWorkflow` API. The module is written in Go and creates `workflows.Workflow` — a serverless orchestrator executing YAML/JSON-defined steps (HTTP calls, GCP connector calls, conditionals, retries). Every source / env-var / service-account change deploys a NEW revision.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Workflows API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── workflow.go            # Workflow creation
    ├── locals.go              # Resolved resource + derived values + label merge
    └── outputs.go             # Stack output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `workflow_name` | `name` | Defaults to metadata.name; ForceNew |
| `region` | `region` | Omitted when empty — the provider's default region applies; ForceNew |
| `description` | `description` | ≤1000 characters |
| `labels` | `labels` | Spec labels merged with platform attribution labels (platform wins) |
| `source_contents` | `source_contents` | REQUIRED (API truth); ≤128KB; every change mints a revision |
| `service_account` | `service_account` | References a GcpServiceAccount's `email` output |
| `crypto_key` | `crypto_key_name` | CMEK; references a GcpKmsKey's `key_id` output |
| `call_log_level` | `call_log_level` | Provider ValidateEnum values |
| `execution_history_level` | `execution_history_level` | Provider ValidateEnum values |
| `user_env_vars` | `user_env_vars` | ≤20 entries together with `secret_env_vars`; keys must not start GOOGLE/WORKFLOWS |
| `secret_env_vars` | `user_env_vars` (merged) | Each value stored in its own Secret Manager secret, accessor granted to the service account on it alone; the merged env var holds the version's resource name |
| `resource_manager_tags` | `tags` | ForceNew — a tag change REPLACES the workflow |
| `deletion_protection` | `deletion_protection` | Sent EXPLICITLY on every apply — a true→false transition must reach the engine |
| `project_id` | `project` | Omitted when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | Omitted when empty (provider default DELETE) |

`name_prefix` is deliberately not modeled: versioned-name rotation is the
catalog pattern, and without hardcoded create-before-destroy semantics the
argument is a dead lever.

The module also enables `workflows.googleapis.com` on the target project
(`disable_on_destroy` false — tearing down one workflow never disables
Workflows project-wide).

## Stack Outputs

| Output | Description |
|---|---|
| `workflow_id` | Full resource name (`projects/{p}/locations/{region}/workflows/{name}`) — the value Eventarc destinations consume |
| `workflow_name` | The short workflow name |
| `revision_id` | The deployed revision (compare across applies to confirm a deploy rolled) |
| `state` | Workflow state (`ACTIVE` when executable) |

## Local development

`stack-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
