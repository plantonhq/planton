---
title: "Function"
description: "Function deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceanfunction"
---

# DigitalOcean Function

Deploys a serverless function on DigitalOcean App Platform with GitHub-based source deployment, configurable runtime and memory, environment variable and secret management, and optional scheduled execution via cron. Currently supported with the Pulumi provisioner only; Terraform support is not available. Under the hood, the component provisions a `digitalocean.App` resource with an `AppSpecFunction` definition, so the function runs inside App Platform rather than as a standalone DigitalOcean Function.

## What Gets Created

When you deploy a DigitalOceanFunction resource, Planton provisions:

- **App Platform Application** -- a `digitalocean.App` resource containing a single function component, deployed to the specified region
- **Function Definition** -- an `AppSpecFunction` within the app spec, configured with the chosen runtime, memory, timeout, entrypoint, and source directory
- **GitHub Integration** -- when `githubSource` is provided, connects the function to a GitHub repository and branch, with optional automatic redeployment on push
- **Environment Variables** -- plain and secret environment variables are translated into `AppSpecFunctionEnv` entries; secrets are stored in App Platform's encrypted secret store
- **HTTP Endpoint** -- when `isWeb` is true (the default), the function is exposed as a public HTTPS endpoint; the URL is available in stack outputs

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or Planton provider config
- **A GitHub repository** containing the function source code and a valid `project.yml` (required when using `githubSource`)
- **DigitalOcean App Platform access** enabled on the account (App Platform is available in all regions that support it)

## Quick Start

Create a file `do-function.yaml`:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFunction
metadata:
  name: my-function
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanFunction.my-function
spec:
  functionName: my-function
  region: nyc3
  runtime: nodejs_20
  sourceDirectory: "/functions/api-handler"
  githubSource:
    repo: "my-org/my-functions"
    branch: "main"
    deployOnPush: true
```

Deploy:

```shell
planton apply -f do-function.yaml
```

This creates a Node.js 20 function deployed via App Platform in the NYC3 region, with automatic redeployment on push to the `main` branch.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `functionName` | `string` | Name of the function. Must be unique within the project. | Min 1, max 64 characters |
| `region` | `enum` | DigitalOcean region for the function. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |
| `runtime` | `enum` | Runtime environment. Valid values: `nodejs_18`, `nodejs_20`, `python_39`, `python_310`, `python_311`, `go_120`, `go_121`, `php_82`. | Required |
| `sourceDirectory` | `string` | Path within the repository containing the function code and `project.yml`. | Min 1 character |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `githubSource.repo` | `string` | -- | GitHub repository in `owner/repo` format. |
| `githubSource.branch` | `string` | -- | Git branch to deploy from (e.g., `main`). Max 255 characters. |
| `githubSource.deployOnPush` | `bool` | `true` | Enables automatic redeployment when changes are pushed to the branch. |
| `memoryMb` | `uint32` | `256` | Memory allocated to the function in megabytes. Valid values: `128`, `256`, `512`, `1024`, `2048`. |
| `timeoutMs` | `uint32` | `3000` | Maximum execution time in milliseconds. Max: `300000` (5 minutes). |
| `environmentVariables` | `map<string, string>` | `{}` | Non-secret environment variables passed to the function. |
| `secretEnvironmentVariables` | `map<string, string>` | `{}` | Encrypted environment variables stored in App Platform's secret store. |
| `entrypoint` | `string` | -- | Function or script entrypoint name (e.g., `main` for Go, `handler` for Node.js). |
| `cronSchedule` | `string` | -- | Cron expression for scheduled execution (e.g., `0 * * * *` for hourly). When set, the function is not exposed as an HTTP endpoint. |
| `isWeb` | `bool` | `true` | Exposes the function as an HTTP endpoint. Set to `false` for background or scheduled functions. |

## Examples

### Node.js API Handler with Secrets

A function that connects to a database using secret credentials and runs with increased memory:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFunction
metadata:
  name: api-handler
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanFunction.api-handler
spec:
  functionName: api-handler
  region: fra1
  runtime: nodejs_20
  sourceDirectory: "/functions/api"
  memoryMb: 512
  timeoutMs: 10000
  githubSource:
    repo: "my-org/backend-functions"
    branch: "production"
    deployOnPush: true
  environmentVariables:
    LOG_LEVEL: "info"
    NODE_ENV: "production"
  secretEnvironmentVariables:
    DATABASE_URL: "postgresql://user:pass@db-host:5432/mydb"
    API_SECRET_KEY: "sk-live-abc123"
```

### Scheduled Python Cleanup Job

A Python function that runs on a cron schedule to perform periodic maintenance, with no HTTP endpoint:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFunction
metadata:
  name: nightly-cleanup
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanFunction.nightly-cleanup
spec:
  functionName: nightly-cleanup
  region: nyc3
  runtime: python_311
  sourceDirectory: "/functions/cleanup"
  entrypoint: "main"
  memoryMb: 1024
  timeoutMs: 300000
  cronSchedule: "0 2 * * *"
  isWeb: false
  githubSource:
    repo: "my-org/ops-functions"
    branch: "main"
    deployOnPush: false
  secretEnvironmentVariables:
    DATABASE_URL: "postgresql://admin:secret@db:5432/app"
```

### Go Webhook Processor with High Memory

A Go function that processes incoming webhooks with maximum memory allocation:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFunction
metadata:
  name: webhook-processor
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanFunction.webhook-processor
spec:
  functionName: webhook-processor
  region: sfo3
  runtime: go_121
  sourceDirectory: "/functions/webhook"
  entrypoint: "main"
  memoryMb: 2048
  timeoutMs: 30000
  isWeb: true
  githubSource:
    repo: "my-org/event-handlers"
    branch: "main"
    deployOnPush: true
  environmentVariables:
    WEBHOOK_PATH: "/api/v1/webhook"
    MAX_PAYLOAD_SIZE: "10485760"
  secretEnvironmentVariables:
    WEBHOOK_SECRET: "whsec_abc123"
    REDIS_URL: "rediss://default:secret@cache:6379"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `function_id` | `string` | Unique identifier of the deployed App Platform application |
| `https_endpoint` | `string` | Public HTTPS URL for invoking the function (populated when `isWeb` is true) |

## Related Components

- [DigitalOceanVpc](/docs/catalog/digitalocean/vpc) -- provides VPC networking for secure function-to-database connectivity
- [DigitalOceanDnsRecord](/docs/catalog/digitalocean/dns-record) -- creates custom DNS records to point a domain at the function endpoint
- [DigitalOceanDnsZone](/docs/catalog/digitalocean/dns-zone) -- manages the DNS zone for custom domain routing
- [DigitalOceanFirewall](/docs/catalog/digitalocean/firewall) -- controls network access rules for resources the function connects to
