# Scaleway Serverless Function

## Overview

The **ScalewayServerlessFunction** resource kind provides a declarative interface for deploying and managing serverless functions on Scaleway. This is a **composite** resource that bundles a function namespace, the function itself, and optional cron triggers into a single declarable unit.

This resource defines the function *infrastructure* (runtime, memory, scaling, networking, environment). Code can be deployed alongside via zip upload, or separately through the Scaleway CLI or CI/CD pipeline.

## Key Features

- **Composite resource** -- Automatically creates and manages the function namespace, function, and scheduled triggers as a single lifecycle unit
- **Kubernetes-style environment variables** -- Variables and secrets defined as ordered name-value lists (not maps), preserving sort order and enabling future `valueFrom` extension
- **VPC connectivity** -- Optional Private Network attachment for secure access to databases, Redis, and other VPC resources
- **Auto-scaling** -- Configurable min/max scale with scale-to-zero support (no compute charges when idle)
- **Scheduled triggers** -- Inline cron triggers for recurring function invocations (hourly cleanups, nightly backups, periodic syncs)
- **Zip-based deployment** -- Optional code deployment via zip upload for fully IaC-managed functions
- **Dual IaC backend** -- Deploy using either Pulumi (Go) or Terraform with identical specifications

## Scaleway Terraform Resource Mapping

| Planton Kind | Terraform Resources | Relationship |
|---|---|---|
| ScalewayServerlessFunction | `scaleway_function_namespace` + `scaleway_function` + `scaleway_function_cron` | 1:1:N (composite) |

## Architecture

```
ScalewayServerlessFunction
├── scaleway_function_namespace (1x, auto-created)
│   └── Groups the function, holds region/project scope
├── scaleway_function (1x)
│   └── Runtime, handler, memory, scaling, env vars, networking
└── scaleway_function_cron (0..Nx, optional)
    └── Scheduled triggers with cron expressions and JSON args
```

## Spec Fields

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `region` | string | Yes | -- | Scaleway region (e.g., "fr-par", "nl-ams", "pl-waw") |
| `runtime` | string | Yes | -- | Language runtime (e.g., "node20", "python311", "go122") |
| `handler` | string | Yes | -- | Function entrypoint, runtime-dependent |
| `privacy` | enum | Yes | -- | `public` (no auth) or `private` (token required) |
| `description` | string | No | "" | Human-readable description |
| `memory_limit_mb` | uint32 | No | 256 | Memory allocation in MB (also affects CPU) |
| `min_scale` | uint32 | No | 0 | Minimum always-running instances (0 = scale-to-zero) |
| `max_scale` | uint32 | No | 20 | Maximum concurrent instances |
| `timeout_seconds` | uint32 | No | 300 | Max execution time per invocation (seconds) |
| `http_option` | enum | No | enabled | `enabled` (HTTP+HTTPS) or `redirected` (HTTP->HTTPS) |
| `env` | message | No | -- | Environment variables and secrets (see below) |
| `private_network_id` | StringValueOrRef | No | -- | Private Network for VPC connectivity |
| `sandbox` | string | No | "" | Execution environment (e.g., "v1", "v2") |
| `zip_file` | string | No | "" | Path to zip with function source code |
| `zip_hash` | string | No | "" | Hash of zip for change detection |
| `cron_triggers` | repeated | No | [] | Scheduled cron triggers |

### Environment Variables

Environment variables use Kubernetes-style repeated name-value messages grouped in an `env` message:

```yaml
env:
  variables:
    - name: NODE_ENV
      value: production
    - name: LOG_LEVEL
      value: info
  secrets:
    - name: DATABASE_URL
      value: postgresql://private-db:5432/mydb
    - name: API_KEY
      value: sk_live_xxxxxxxxxxxx
```

- **`variables`**: Non-secret, visible in Scaleway console and logs
- **`secrets`**: Encrypted at rest, masked in the console

### Cron Trigger Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | No | Human-readable trigger name |
| `schedule` | string | Yes | UNIX cron expression |
| `args` | string | Yes | JSON string passed to function event object |

## Stack Outputs

| Output | Description |
|---|---|
| `function_id` | Scaleway function UUID |
| `namespace_id` | Function namespace UUID |
| `domain_name` | Native Scaleway invoke domain (for custom domain CNAME) |

## Dependencies

**Upstream:**
- **ScalewayPrivateNetwork** -- `private_network_id` references `status.outputs.private_network_id` for VPC connectivity

**Downstream:**
- **ScalewayDnsRecord** -- Create CNAME records pointing to `domain_name` for custom domains

## Supported Runtimes

Scaleway supports five languages. Runtimes are specified as strings (not enums) to accommodate frequent additions:

| Language | Example Runtimes |
|---|---|
| Node.js | `node20`, `node22` |
| Python | `python39`, `python310`, `python311`, `python312`, `python313` |
| Go | `go122`, `go123`, `go124` |
| Rust | `rust165` |
| PHP | `php82` |

See [Scaleway Functions Runtimes](https://www.scaleway.com/en/docs/serverless-functions/reference-content/functions-runtimes/) for the latest list.

## Code Deployment Model

This resource supports two deployment models:

### 1. IaC-Managed Deployment (zip upload)

Provide `zip_file` and `zip_hash` in the spec. The IaC module uploads the archive and triggers deployment automatically.

```yaml
spec:
  zip_file: "./dist/function.zip"
  zip_hash: "sha256:abc123..."
```

### 2. Separate Deployment (CLI / CI/CD)

Omit `zip_file`. The IaC module creates the function infrastructure; code is deployed separately:

```bash
scw function deploy --name my-function --runtime node20 --handler handler.handler
```

This separation follows the standard pattern where IaC manages infrastructure and CI/CD manages code deployment.

## Important Constraints

### Namespace Lifecycle
The namespace is an implementation detail -- users interact with the function as a single resource. One namespace is created per ScalewayServerlessFunction for clean isolation.

### Name Immutability
Changing the function `name` (from `metadata.name`) or namespace name triggers resource recreation. Plan accordingly.

### Secret Lifecycle
Secret environment variables are ignored in Terraform's change detection (`lifecycle.ignore_changes`) to prevent unnecessary updates when secrets are managed externally. Changes to secrets via the spec will still be applied on initial creation and explicit forced updates.

### No Built-in Git Integration
Unlike DigitalOcean's App Platform, Scaleway serverless functions do not have built-in GitHub/GitLab integration for automatic deployments. Use external CI/CD pipelines for git-triggered deployments.

### No Built-in Custom Domains
Custom domain binding is handled by creating ScalewayDnsRecord CNAME records pointing to the function's `domain_name` output. The `scaleway_function_domain` Terraform resource exists but is not bundled in this composite to avoid overlapping with the DNS tier.

## Scaleway Documentation

- [Scaleway Serverless Functions](https://www.scaleway.com/en/docs/serverless/functions/)
- [Functions Runtimes](https://www.scaleway.com/en/docs/serverless-functions/reference-content/functions-runtimes/)
- [CRON Schedules](https://www.scaleway.com/en/docs/serverless/functions/reference-content/cron-schedules/)
- [Terraform: scaleway_function](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/function)
- [Terraform: scaleway_function_namespace](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/function_namespace)
- [Terraform: scaleway_function_cron](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/function_cron)
- [Pulumi: scaleway.functions.Function](https://www.pulumi.com/registry/packages/scaleway/api-docs/functions/function/)
