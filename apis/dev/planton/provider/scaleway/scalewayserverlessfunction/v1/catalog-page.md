# Scaleway Serverless Function

Deploys a Scaleway serverless function with its own dedicated namespace, configurable runtime, memory, scaling, networking, environment variables, and optional cron-based scheduled triggers. This composite resource provisions all infrastructure needed to run event-driven or HTTP-triggered workloads without managing servers.

## What Gets Created

When you deploy a ScalewayServerlessFunction resource, Planton provisions:

- **Function Namespace** — a single `scaleway_function_namespace` resource that acts as the grouping container for the function. One namespace is created per function for clean lifecycle management and isolation.
- **Serverless Function** — a single `scaleway_function` resource defining the runtime, handler, memory, scaling, privacy, environment variables, secrets, and optional code deployment.
- **Cron Triggers** — zero or more `scaleway_function_cron` resources, one per entry in `cronTriggers`, each invoking the function on the specified schedule with the given JSON arguments.

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **A supported runtime** string matching an available Scaleway serverless runtime (e.g., `"node20"`, `"python312"`, `"go124"`)
- **A Private Network** in the target region if using private connectivity (can be created via a ScalewayPrivateNetwork resource)

## Quick Start

Create a file `serverless-function.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessFunction
metadata:
  name: my-function
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayServerlessFunction.my-function
spec:
  region: fr-par
  runtime: python312
  handler: handler.handle
  privacy: public
```

Deploy:

```shell
planton apply -f serverless-function.yaml
```

This creates a public Python 3.12 serverless function in the `fr-par` region with 256 MB memory, scale-to-zero behavior, and a 5-minute timeout. Code can be deployed separately via the Scaleway CLI (`scw function deploy`) or CI/CD pipeline.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region where the function namespace and function are deployed (e.g., `"fr-par"`, `"nl-ams"`, `"pl-waw"`). Serverless functions are regional resources. | Required |
| `runtime` | `string` | Language runtime for the function (e.g., `"node20"`, `"python312"`, `"go124"`, `"rust165"`, `"php82"`). Plain string to avoid proto staleness as Scaleway adds new runtimes. | Required |
| `handler` | `string` | Function entrypoint, runtime-dependent. Examples: `"handler.handle"` (Python), `"handler.handler"` (Node.js), `"Handle"` (Go). | Required |
| `privacy` | `enum` | Authentication behavior for the function endpoint. `public` = no authentication required. `private` = requires a valid authentication token. | Required, must be `public` or `private` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the function. Applied to both the namespace and the function. |
| `memoryLimitMb` | `uint32` | `256` | Memory allocated to each function instance in megabytes. Higher memory also increases CPU allocation proportionally. Common values: 128, 256, 512, 1024, 2048. |
| `minScale` | `uint32` | `0` | Minimum number of always-running instances. `0` enables scale-to-zero (no compute charges when idle). `1+` keeps instances warm to eliminate cold starts. |
| `maxScale` | `uint32` | `20` | Maximum number of concurrent function instances. The function auto-scales based on workload but never exceeds this limit. |
| `timeoutSeconds` | `uint32` | `300` | Maximum execution time for a single invocation in seconds. Scaleway terminates the function if exceeded. |
| `httpOption` | `enum` | `enabled` | HTTP/HTTPS behavior. `enabled` = both HTTP and HTTPS accepted. `redirected` = HTTP requests automatically redirect to HTTPS. |
| `env` | `object` | — | Groups environment variables and secrets. See sub-fields below. |
| `env.variables` | `object[]` | `[]` | Non-secret environment variables. Each entry has `name` and `value` fields. Visible in the Scaleway console and logs. |
| `env.secrets` | `object[]` | `[]` | Encrypted environment variables. Each entry has `name` and `value` fields. Stored encrypted at rest and masked in the console. Use for database URLs, API keys, and tokens. |
| `privateNetworkId` | `StringValueOrRef` | — | Connects the function to a Scaleway Private Network for VPC-internal communication. When connected, the function can reach databases, Redis clusters, and other services without traversing the public internet. Can reference a ScalewayPrivateNetwork resource via `valueFrom`. |
| `sandbox` | `string` | `""` | Execution environment for the function. Common values: `"v1"` (standard), `"v2"` (enhanced security). Leave empty for the platform default. |
| `zipFile` | `string` | `""` | Local path to a zip archive containing the function source code. When provided, the module uploads the archive and triggers deployment. Leave empty when deploying code separately. |
| `zipHash` | `string` | `""` | Hash of the zip archive for change detection. When the hash changes, the module re-uploads and redeploys. Only meaningful when `zipFile` is also set. |
| `cronTriggers` | `object[]` | `[]` | Scheduled triggers that invoke the function on a cron schedule. See sub-fields below. |
| `cronTriggers[].name` | `string` | auto-generated | Human-readable identifier for the trigger. Makes triggers easier to identify in the console and logs. |
| `cronTriggers[].schedule` | `string` | — | UNIX cron expression (e.g., `"0 * * * *"` for hourly, `"0 2 * * *"` for daily at 2 AM). Required per trigger. |
| `cronTriggers[].args` | `string` | — | JSON string passed to the function's event object on each invocation. Must be valid JSON (e.g., `"{}"`). Required per trigger. |

## Examples

### Public HTTP Function

A minimal public HTTP function suitable for webhooks, API endpoints, or development:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessFunction
metadata:
  name: webhook-handler
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayServerlessFunction.webhook-handler
spec:
  region: fr-par
  runtime: node20
  handler: handler.handler
  privacy: public
  memoryLimitMb: 128
  timeoutSeconds: 30
  env:
    variables:
      - name: LOG_LEVEL
        value: debug
      - name: APP_ENV
        value: development
```

### Private Function with Secrets and Private Network

A production function with token-based authentication, encrypted secrets, HTTPS enforcement, and Private Network connectivity to reach backend databases:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessFunction
metadata:
  name: order-processor
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayServerlessFunction.order-processor
spec:
  region: fr-par
  runtime: python312
  handler: handler.handle
  privacy: private
  memoryLimitMb: 512
  minScale: 1
  maxScale: 50
  timeoutSeconds: 60
  httpOption: redirected
  env:
    variables:
      - name: APP_ENV
        value: production
      - name: LOG_LEVEL
        value: warn
    secrets:
      - name: DATABASE_URL
        value: postgres://user:pass@db.internal:5432/orders
      - name: STRIPE_API_KEY
        value: sk_live_xxxxxxxxxxxx
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
```

### Scheduled Function with Cron Triggers and Zip Deployment

A Go function deployed via zip archive with multiple cron triggers for recurring background tasks:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessFunction
metadata:
  name: nightly-jobs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayServerlessFunction.nightly-jobs
spec:
  region: nl-ams
  runtime: go124
  handler: Handle
  privacy: private
  memoryLimitMb: 1024
  minScale: 0
  maxScale: 5
  timeoutSeconds: 600
  sandbox: v2
  zipFile: ./dist/nightly-jobs.zip
  zipHash: sha256:a1b2c3d4e5f6...
  env:
    variables:
      - name: BATCH_SIZE
        value: "500"
    secrets:
      - name: DATABASE_URL
        value: postgres://user:pass@db.internal:5432/analytics
  cronTriggers:
    - name: cleanup-stale-sessions
      schedule: "0 2 * * *"
      args: '{"type": "stale_sessions", "max_age_hours": 24}'
    - name: generate-daily-report
      schedule: "0 6 * * *"
      args: '{"report": "daily_summary", "format": "csv"}'
    - name: health-check
      schedule: "*/15 * * * *"
      args: '{}'
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `function_id` | `string` | Scaleway-assigned UUID of the deployed serverless function. Used for API operations, Scaleway CLI commands, and resource import. |
| `namespace_id` | `string` | UUID of the function namespace. Useful for managing additional resources that reference the namespace (external cron triggers, tokens, or additional functions). |
| `domain_name` | `string` | Native Scaleway HTTPS domain for invoking the function (e.g., `"myfunc-abc123.functions.fnc.fr-par.scw.cloud"`). Downstream ScalewayDnsRecord resources can create CNAME records pointing to this domain for custom domain routing. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/scalewayprivatenetwork) — provides private connectivity between the function and backend services such as databases and Redis clusters
- [ScalewayRedisCluster](/docs/catalog/scaleway/scalewayrediscluster) — deploys managed Redis clusters reachable from the function over a Private Network
- [ScalewayRdbInstance](/docs/catalog/scaleway/scalewayrdbinstance) — deploys managed PostgreSQL or MySQL databases that the function can connect to
- [ScalewayObjectBucket](/docs/catalog/scaleway/scalewayobjectbucket) — provides S3-compatible object storage for function input/output data
- [ScalewayContainerRegistry](/docs/catalog/scaleway/scalewaycontainerregistry) — hosts container images if migrating from serverless functions to containerized workloads
