# DigitalOcean App Platform Service

Deploys a containerized application on DigitalOcean App Platform as a web service, background worker, or one-off job. The component supports two deployment sources -- building from a git repository or running a pre-built container image from DigitalOcean Container Registry (DOCR). It handles instance sizing, autoscaling, environment variable injection, and optional custom domain configuration backed by a DNS zone reference.

## What Gets Created

When you deploy a DigitalOceanAppPlatformService resource, OpenMCF provisions:

- **App Platform Application** -- a `digitalocean_app` resource containing a single service component configured according to the chosen `serviceType`
- **Web Service** -- created when `serviceType` is `web_service`; receives external HTTP traffic, supports autoscaling with CPU-based metrics (80% threshold), and accepts `buildCommand`/`runCommand` overrides for git sources
- **Worker** -- created when `serviceType` is `worker`; runs as a background process without HTTP ingress, supports fixed instance counts and environment variables scoped to `RUN_TIME`
- **Job** -- created when `serviceType` is `job`; executes as a `PRE_DEPLOY` task, useful for database migrations or one-off scripts
- **Custom Domain** -- created only when `customDomain` is specified; adds a `PRIMARY` domain entry to the app spec and expects DNS to be managed by a referenced DigitalOceanDnsZone resource
- **Environment Variables** -- injected from the `env` map; scoped to `RUN_AND_BUILD_TIME` for web services and `RUN_TIME` for workers and jobs

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **A git repository** accessible to DigitalOcean App Platform (for git source deployments), or **a DigitalOcean Container Registry** with a pushed image (for image source deployments)
- **Exactly one source** must be provided per deployment: either `gitSource` or `imageSource`, never both

## Quick Start

Create a file `app-service.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanAppPlatformService
metadata:
  name: my-web-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanAppPlatformService.my-web-app
spec:
  serviceName: my-web-app
  region: nyc3
  serviceType: web_service
  gitSource:
    repoUrl: "https://github.com/example/my-app.git"
    branch: main
  instanceSizeSlug: basic-xxs
  instanceCount: 1
```

Deploy:

```shell
openmcf apply -f app-service.yaml
```

This creates a single-instance web service in the NYC3 region, built from the `main` branch of the specified git repository using the smallest available instance size.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `serviceName` | `string` | Name of the app in DigitalOcean. Must be unique per account. | Required; lowercase alphanumeric and hyphens; max 63 characters; pattern `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `region` | `enum` | DigitalOcean region for the app. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |
| `serviceType` | `enum` | Type of component to deploy. Valid values: `web_service`, `worker`, `job`. | Required |
| `instanceSizeSlug` | `enum` | Instance size plan. Valid values: `basic-xxs`, `basic-xs`, `basic-s`, `basic-m`, `basic-l`, `professional-xs`, `professional-s`, `professional-m`, `professional-l`, `professional-xl`. | Required; recommended default `basic-xxs` |
| `gitSource.repoUrl` | `string` | HTTPS or git URL of the source repository. | Required when using `gitSource` |
| `gitSource.branch` | `string` | Git branch to deploy from. | Required when using `gitSource` |
| `imageSource.registry` | `StringValueOrRef` | Reference to a DigitalOceanContainerRegistry resource. Resolves to the registry URL via `status.outputs.server_url`. | Required when using `imageSource` |
| `imageSource.repository` | `string` | Repository name within the registry (e.g., `myapp/backend`). | Required when using `imageSource` |
| `imageSource.tag` | `string` | Image tag to deploy (e.g., `latest` or `v1.0.0`). | Required when using `imageSource` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `instanceCount` | `uint32` | `1` | Number of instances to run. Ignored when autoscaling is enabled. |
| `enableAutoscale` | `bool` | `false` | Enables CPU-based autoscaling for the service. Only supported for `web_service` type. |
| `minInstanceCount` | `uint32` | -- | Minimum number of instances when autoscaling is enabled. Required if `enableAutoscale` is `true`. |
| `maxInstanceCount` | `uint32` | -- | Maximum number of instances when autoscaling is enabled. Required if `enableAutoscale` is `true`. |
| `env` | `map<string, string>` | `{}` | Environment variables injected into the runtime. Keys are variable names, values are their contents. |
| `customDomain` | `StringValueOrRef` | -- | Custom domain for the app. Can reference a DigitalOceanDnsZone resource via `valueFrom` (resolves to `spec.domain_name`). |
| `gitSource.buildCommand` | `string` | -- | Overrides the default build command (e.g., `npm run build`). Only applies to git source deployments. |
| `gitSource.runCommand` | `string` | -- | Overrides the default start command (e.g., `npm start`). Only applies to git source deployments. |

## Examples

### Web Service from Git with Custom Build Commands

A Node.js web application built from source with explicit build and run commands:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanAppPlatformService
metadata:
  name: node-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanAppPlatformService.node-api
spec:
  serviceName: node-api
  region: fra1
  serviceType: web_service
  gitSource:
    repoUrl: "https://github.com/example/node-api.git"
    branch: main
    buildCommand: "npm ci && npm run build"
    runCommand: "npm start"
  instanceSizeSlug: basic-s
  instanceCount: 2
  env:
    NODE_ENV: "production"
    PORT: "8080"
```

### Worker from Container Registry with Autoscaling Web Service

A background worker deployed from a pre-built container image stored in DigitalOcean Container Registry:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanAppPlatformService
metadata:
  name: queue-processor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanAppPlatformService.queue-processor
spec:
  serviceName: queue-processor
  region: sfo3
  serviceType: worker
  imageSource:
    registry:
      valueFrom:
        kind: DigitalOceanContainerRegistry
        name: prod-registry
        field: status.outputs.server_url
    repository: "myapp/queue-processor"
    tag: "v2.1.0"
  instanceSizeSlug: professional-s
  instanceCount: 3
  env:
    REDIS_URL: "redis://private-redis:6379"
    QUEUE_NAME: "tasks"
```

### Production Web Service with Autoscaling and Custom Domain

A fully configured production web service with autoscaling, environment variables, and a custom domain backed by a DNS zone reference:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanAppPlatformService
metadata:
  name: prod-frontend
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanAppPlatformService.prod-frontend
spec:
  serviceName: prod-frontend
  region: ams3
  serviceType: web_service
  imageSource:
    registry:
      valueFrom:
        kind: DigitalOceanContainerRegistry
        name: prod-registry
        field: status.outputs.server_url
    repository: "myapp/frontend"
    tag: "v3.0.1"
  instanceSizeSlug: professional-m
  enableAutoscale: true
  minInstanceCount: 2
  maxInstanceCount: 10
  env:
    API_BASE_URL: "https://api.example.com"
    LOG_LEVEL: "info"
  customDomain:
    valueFrom:
      kind: DigitalOceanDnsZone
      name: example-zone
      field: spec.domain_name
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `appId` | `string` | Unique identifier of the created App Platform application |
| `defaultHostname` | `string` | Default hostname assigned to the app (typically ending in `ondigitalocean.app`) |
| `liveUrl` | `string` | Publicly accessible URL of the deployed service, including protocol. Reflects the custom domain if one was configured. |

## Related Components

- [DigitalOceanDnsZone](/docs/catalog/digitalocean/digitaloceandnszone) -- manages the DNS zone referenced by `customDomain` for routing traffic to the app
- [DigitalOceanContainerRegistry](/docs/catalog/digitalocean/digitaloceancontainerregistry) -- hosts private container images used by `imageSource` deployments
- [DigitalOceanDnsRecord](/docs/catalog/digitalocean/digitaloceandnsrecord) -- creates individual DNS records within a zone, useful for additional subdomains pointing to the service
- [DigitalOceanFirewall](/docs/catalog/digitalocean/digitaloceanfirewall) -- controls network access rules for DigitalOcean resources
- [DigitalOceanVpc](/docs/catalog/digitalocean/digitaloceanvpc) -- provides private networking for other DigitalOcean resources that interact with the app
