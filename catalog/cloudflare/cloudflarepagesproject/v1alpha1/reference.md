# CloudflarePagesProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflarePagesProjectSpec configures a Cloudflare Pages project: a managed
site host that builds and serves a static site or full-stack app (static
assets + Pages Functions) from Cloudflare's edge.

This models the durable *project* — its build configuration, optional git
connection, per-environment runtime configuration (bindings, env vars,
compatibility), and custom domains. The actual *deployments* (the built
versions of the site) are produced out-of-band: for a git-connected project
Cloudflare builds a new deployment on every push to the connected repository;
for a direct-upload project a deployment is pushed with `wrangler pages
deploy`. The Cloudflare Terraform/Pulumi providers expose no deployment
resource, so this component never creates a deployment — it manages the
project that deployments land in.

Prerequisite for git-connected projects: the Cloudflare account must already
be authorized with the git provider (the GitHub App install / GitLab OAuth
connection), which is a one-time, browser-driven step in the Cloudflare
dashboard. The provider manages the `source` configuration (which repo,
branch, and build settings), not the underlying authorization.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflarePagesProject
metadata:
  name: test-pages-project
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: test-pages-project
  productionBranch: main
  # Direct-upload project (no git source): deployments are pushed with
  # `wrangler pages deploy`. For a git-connected project, add a `source` block
  # (requires a one-time git authorization in the Cloudflare dashboard).
  buildConfig:
    buildCommand: npm run build
    destinationDir: dist
  deploymentConfigs:
    production:
      compatibilityDate: "2025-01-15"
      compatibilityFlags:
        - nodejs_compat
      usageModel: standard
      vars:
        LOG_LEVEL: info
      kvNamespaces:
        - name: CONFIG
          namespaceId:
            value: 0f1e2d3c4b5a69788796a5b4c3d2e1f0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.productionBranch` | `string` | yes |  |  |
| `spec.buildConfig` | `CloudflarePagesBuildConfig` |  |  |  |
| `spec.buildConfig.buildCommand` | `string` |  |  |  |
| `spec.buildConfig.destinationDir` | `string` |  |  |  |
| `spec.buildConfig.rootDir` | `string` |  |  |  |
| `spec.buildConfig.buildCaching` | `bool` |  |  |  |
| `spec.buildConfig.webAnalyticsTag` | `string` |  |  |  |
| `spec.buildConfig.webAnalyticsToken` | `string` (sensitive) |  |  |  |
| `spec.source` | `CloudflarePagesSource` |  |  |  |
| `spec.source.type` | `string` | yes |  |  |
| `spec.source.config` | `CloudflarePagesSourceConfig` | yes |  |  |
| `spec.source.config.owner` | `string` |  |  |  |
| `spec.source.config.repoName` | `string` |  |  |  |
| `spec.source.config.productionBranch` | `string` |  |  |  |
| `spec.source.config.prCommentsEnabled` | `bool` |  |  |  |
| `spec.source.config.deploymentsEnabled` | `bool` |  |  |  |
| `spec.source.config.productionDeploymentsEnabled` | `bool` |  |  |  |
| `spec.source.config.previewDeploymentSetting` | `string` |  |  |  |
| `spec.source.config.previewBranchIncludes` | `[]string` |  |  |  |
| `spec.source.config.previewBranchExcludes` | `[]string` |  |  |  |
| `spec.source.config.pathIncludes` | `[]string` |  |  |  |
| `spec.source.config.pathExcludes` | `[]string` |  |  |  |
| `spec.deploymentConfigs` | `CloudflarePagesDeploymentConfigs` |  |  |  |
| `spec.deploymentConfigs.preview` | `CloudflarePagesDeploymentConfig` |  |  |  |
| `spec.deploymentConfigs.preview.compatibilityDate` | `string` |  |  |  |
| `spec.deploymentConfigs.preview.compatibilityFlags` | `[]string` |  |  |  |
| `spec.deploymentConfigs.preview.alwaysUseLatestCompatibilityDate` | `bool` |  |  |  |
| `spec.deploymentConfigs.preview.buildImageMajorVersion` | `int64` |  |  |  |
| `spec.deploymentConfigs.preview.failOpen` | `bool` |  |  |  |
| `spec.deploymentConfigs.preview.usageModel` | `string` |  |  |  |
| `spec.deploymentConfigs.preview.limits` | `CloudflarePagesLimits` |  |  |  |
| `spec.deploymentConfigs.preview.limits.cpuMs` | `int64` |  |  |  |
| `spec.deploymentConfigs.preview.placement` | `CloudflarePagesPlacement` |  |  |  |
| `spec.deploymentConfigs.preview.placement.mode` | `string` |  |  |  |
| `spec.deploymentConfigs.preview.vars` | `map<string, string>` |  |  |  |
| `spec.deploymentConfigs.preview.secrets` | `[]CloudflarePagesSecret` |  |  |  |
| `spec.deploymentConfigs.preview.secrets[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.secrets[].value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.deploymentConfigs.preview.kvNamespaces` | `[]CloudflarePagesKvBinding` |  |  |  |
| `spec.deploymentConfigs.preview.kvNamespaces[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.kvNamespaces[].namespaceId` | `string \| valueFrom` | yes |  | CloudflareKvNamespace (`status.outputs.namespace_id`) |
| `spec.deploymentConfigs.preview.d1Databases` | `[]CloudflarePagesD1Binding` |  |  |  |
| `spec.deploymentConfigs.preview.d1Databases[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.d1Databases[].databaseId` | `string \| valueFrom` | yes |  | CloudflareD1Database (`status.outputs.database_id`) |
| `spec.deploymentConfigs.preview.r2Buckets` | `[]CloudflarePagesR2Binding` |  |  |  |
| `spec.deploymentConfigs.preview.r2Buckets[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.r2Buckets[].bucketName` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.deploymentConfigs.preview.r2Buckets[].jurisdiction` | `string` |  |  |  |
| `spec.deploymentConfigs.preview.queueProducers` | `[]CloudflarePagesQueueProducerBinding` |  |  |  |
| `spec.deploymentConfigs.preview.queueProducers[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.queueProducers[].queueName` | `string \| valueFrom` | yes |  | CloudflareQueue (`status.outputs.queue_name`) |
| `spec.deploymentConfigs.preview.hyperdriveBindings` | `[]CloudflarePagesHyperdriveBinding` |  |  |  |
| `spec.deploymentConfigs.preview.hyperdriveBindings[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.hyperdriveBindings[].configId` | `string \| valueFrom` | yes |  | CloudflareHyperdriveConfig (`status.outputs.hyperdrive_id`) |
| `spec.deploymentConfigs.preview.services` | `[]CloudflarePagesServiceBinding` |  |  |  |
| `spec.deploymentConfigs.preview.services[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.services[].service` | `string \| valueFrom` | yes |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.deploymentConfigs.preview.services[].entrypoint` | `string` |  |  |  |
| `spec.deploymentConfigs.preview.services[].environment` | `string` |  |  |  |
| `spec.deploymentConfigs.preview.durableObjectNamespaces` | `[]CloudflarePagesDurableObjectBinding` |  |  |  |
| `spec.deploymentConfigs.preview.durableObjectNamespaces[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.durableObjectNamespaces[].namespaceId` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.analyticsEngineDatasets` | `[]CloudflarePagesAnalyticsEngineBinding` |  |  |  |
| `spec.deploymentConfigs.preview.analyticsEngineDatasets[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.analyticsEngineDatasets[].dataset` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.vectorizeBindings` | `[]CloudflarePagesVectorizeBinding` |  |  |  |
| `spec.deploymentConfigs.preview.vectorizeBindings[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.vectorizeBindings[].indexName` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.aiBindings` | `[]CloudflarePagesAiBinding` |  |  |  |
| `spec.deploymentConfigs.preview.aiBindings[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.aiBindings[].projectId` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.mtlsCertificates` | `[]CloudflarePagesMtlsCertificateBinding` |  |  |  |
| `spec.deploymentConfigs.preview.mtlsCertificates[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.mtlsCertificates[].certificateId` | `string` | yes |  |  |
| `spec.deploymentConfigs.preview.browsers` | `[]CloudflarePagesBrowserBinding` |  |  |  |
| `spec.deploymentConfigs.preview.browsers[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production` | `CloudflarePagesDeploymentConfig` |  |  |  |
| `spec.deploymentConfigs.production.compatibilityDate` | `string` |  |  |  |
| `spec.deploymentConfigs.production.compatibilityFlags` | `[]string` |  |  |  |
| `spec.deploymentConfigs.production.alwaysUseLatestCompatibilityDate` | `bool` |  |  |  |
| `spec.deploymentConfigs.production.buildImageMajorVersion` | `int64` |  |  |  |
| `spec.deploymentConfigs.production.failOpen` | `bool` |  |  |  |
| `spec.deploymentConfigs.production.usageModel` | `string` |  |  |  |
| `spec.deploymentConfigs.production.limits` | `CloudflarePagesLimits` |  |  |  |
| `spec.deploymentConfigs.production.limits.cpuMs` | `int64` |  |  |  |
| `spec.deploymentConfigs.production.placement` | `CloudflarePagesPlacement` |  |  |  |
| `spec.deploymentConfigs.production.placement.mode` | `string` |  |  |  |
| `spec.deploymentConfigs.production.vars` | `map<string, string>` |  |  |  |
| `spec.deploymentConfigs.production.secrets` | `[]CloudflarePagesSecret` |  |  |  |
| `spec.deploymentConfigs.production.secrets[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.secrets[].value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.deploymentConfigs.production.kvNamespaces` | `[]CloudflarePagesKvBinding` |  |  |  |
| `spec.deploymentConfigs.production.kvNamespaces[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.kvNamespaces[].namespaceId` | `string \| valueFrom` | yes |  | CloudflareKvNamespace (`status.outputs.namespace_id`) |
| `spec.deploymentConfigs.production.d1Databases` | `[]CloudflarePagesD1Binding` |  |  |  |
| `spec.deploymentConfigs.production.d1Databases[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.d1Databases[].databaseId` | `string \| valueFrom` | yes |  | CloudflareD1Database (`status.outputs.database_id`) |
| `spec.deploymentConfigs.production.r2Buckets` | `[]CloudflarePagesR2Binding` |  |  |  |
| `spec.deploymentConfigs.production.r2Buckets[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.r2Buckets[].bucketName` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.deploymentConfigs.production.r2Buckets[].jurisdiction` | `string` |  |  |  |
| `spec.deploymentConfigs.production.queueProducers` | `[]CloudflarePagesQueueProducerBinding` |  |  |  |
| `spec.deploymentConfigs.production.queueProducers[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.queueProducers[].queueName` | `string \| valueFrom` | yes |  | CloudflareQueue (`status.outputs.queue_name`) |
| `spec.deploymentConfigs.production.hyperdriveBindings` | `[]CloudflarePagesHyperdriveBinding` |  |  |  |
| `spec.deploymentConfigs.production.hyperdriveBindings[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.hyperdriveBindings[].configId` | `string \| valueFrom` | yes |  | CloudflareHyperdriveConfig (`status.outputs.hyperdrive_id`) |
| `spec.deploymentConfigs.production.services` | `[]CloudflarePagesServiceBinding` |  |  |  |
| `spec.deploymentConfigs.production.services[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.services[].service` | `string \| valueFrom` | yes |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.deploymentConfigs.production.services[].entrypoint` | `string` |  |  |  |
| `spec.deploymentConfigs.production.services[].environment` | `string` |  |  |  |
| `spec.deploymentConfigs.production.durableObjectNamespaces` | `[]CloudflarePagesDurableObjectBinding` |  |  |  |
| `spec.deploymentConfigs.production.durableObjectNamespaces[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.durableObjectNamespaces[].namespaceId` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.analyticsEngineDatasets` | `[]CloudflarePagesAnalyticsEngineBinding` |  |  |  |
| `spec.deploymentConfigs.production.analyticsEngineDatasets[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.analyticsEngineDatasets[].dataset` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.vectorizeBindings` | `[]CloudflarePagesVectorizeBinding` |  |  |  |
| `spec.deploymentConfigs.production.vectorizeBindings[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.vectorizeBindings[].indexName` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.aiBindings` | `[]CloudflarePagesAiBinding` |  |  |  |
| `spec.deploymentConfigs.production.aiBindings[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.aiBindings[].projectId` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.mtlsCertificates` | `[]CloudflarePagesMtlsCertificateBinding` |  |  |  |
| `spec.deploymentConfigs.production.mtlsCertificates[].name` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.mtlsCertificates[].certificateId` | `string` | yes |  |  |
| `spec.deploymentConfigs.production.browsers` | `[]CloudflarePagesBrowserBinding` |  |  |  |
| `spec.deploymentConfigs.production.browsers[].name` | `string` | yes |  |  |
| `spec.domains` | `[]string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this project.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string` · required

The project name. Lowercase alphanumeric with hyphens; this is also the
project's `*.pages.dev` subdomain label. Immutable — changing it replaces
the project.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"58","pattern":"^[a-z0-9][a-z0-9-]*$"}}

### spec.productionBranch

`string` · required

The production branch name (e.g. "main"). For a git-connected project,
pushes to this branch produce production deployments; other branches produce
preview deployments. Required even for direct-upload projects.

- rule: {"required":true}

### spec.buildConfig

`CloudflarePagesBuildConfig`

Build configuration: how Cloudflare builds the site from source (used by
git-connected projects and by `wrangler pages deploy` builds). Omit for a
pre-built direct-upload project.

### spec.buildConfig.buildCommand

`string`

The build command Cloudflare runs (e.g. "npm run build").

### spec.buildConfig.destinationDir

`string`

The directory containing the build output to serve (e.g. "dist").

### spec.buildConfig.rootDir

`string`

The repository subdirectory to treat as the project root (for monorepos).

### spec.buildConfig.buildCaching

`bool`

Enable build caching to speed up successive builds.

### spec.buildConfig.webAnalyticsTag

`string`

Web Analytics tag to inject into served pages.

### spec.buildConfig.webAnalyticsToken

`string` · sensitive

Web Analytics token paired with the tag. Secret — provide a managed-secret
reference; resolved just-in-time at deploy.

### spec.source

`CloudflarePagesSource`

Git connection. When set, Cloudflare connects to the repository and builds a
new deployment on every push (Cloudflare is the CI). Omit for a direct-upload
project whose deployments are pushed with `wrangler pages deploy`.

### spec.source.type

`string` · required

The git provider hosting the repository.

- rule: source type must be one of "github", "gitlab"
- rule: {"required":true}

### spec.source.config

`CloudflarePagesSourceConfig` · required

Repository and branch-deployment settings.

- rule: {"required":true}

### spec.source.config.owner

`string`

The repository owner (user or organization).

### spec.source.config.repoName

`string`

The repository name.

### spec.source.config.productionBranch

`string`

The production branch; pushes here produce production deployments. Defaults to
the project's production_branch when unset.

### spec.source.config.prCommentsEnabled

`bool`

Post a deployment status comment on pull requests.

### spec.source.config.deploymentsEnabled

`bool`

Enable automatic deployments on push (both preview and production).

### spec.source.config.productionDeploymentsEnabled

`bool`

Enable automatic production deployments specifically.

### spec.source.config.previewDeploymentSetting

`string`

Which non-production branches get preview deployments. One of "all", "none",
or "custom" (use the include/exclude lists below). Empty uses Cloudflare's
default ("all").

- rule: preview_deployment_setting must be one of "all", "none", "custom"

### spec.source.config.previewBranchIncludes

`[]string`

With preview_deployment_setting = "custom", branches matched here get preview
deployments.

### spec.source.config.previewBranchExcludes

`[]string`

With preview_deployment_setting = "custom", branches matched here are excluded
from preview deployments.

### spec.source.config.pathIncludes

`[]string`

Only build when files under these paths change.

### spec.source.config.pathExcludes

`[]string`

Skip builds when only files under these paths change.

### spec.deploymentConfigs

`CloudflarePagesDeploymentConfigs`

Per-environment runtime configuration (bindings, env vars, compatibility,
limits) for preview and production deployments.

### spec.deploymentConfigs.preview

`CloudflarePagesDeploymentConfig`

Configuration applied to preview deployments.

### spec.deploymentConfigs.preview.compatibilityDate

`string`

Compatibility date (YYYY-MM-DD) pinning the Workers runtime behavior for
Functions.

- rule: {"string":{"pattern":"^([0-9]{4}-[0-9]{2}-[0-9]{2})?$"}}

### spec.deploymentConfigs.preview.compatibilityFlags

`[]string`

Runtime compatibility flags (e.g. "nodejs_compat").

### spec.deploymentConfigs.preview.alwaysUseLatestCompatibilityDate

`bool`

Always build Functions against the latest compatibility date instead of a
pinned one.

### spec.deploymentConfigs.preview.buildImageMajorVersion

`int64`

Major version of the Cloudflare build image to use. Leave 0 for the default.

- rule: build_image_major_version must be 0 (default) or positive

### spec.deploymentConfigs.preview.failOpen

`bool`

Continue serving (fail open) when a Function cannot be invoked, rather than
returning an error.

### spec.deploymentConfigs.preview.usageModel

`string`

Functions usage/pricing model. One of "standard", "bundled", or "unbound".
Empty uses the account default. ("bundled"/"unbound" are legacy models.)

- rule: usage_model must be one of "standard", "bundled", "unbound"

### spec.deploymentConfigs.preview.limits

`CloudflarePagesLimits`

Per-invocation limits for Functions.

### spec.deploymentConfigs.preview.limits.cpuMs

`int64`

CPU time limit in milliseconds per invocation. Leave 0 for the plan default.

- rule: cpu_ms must be 0 (default) or positive

### spec.deploymentConfigs.preview.placement

`CloudflarePagesPlacement`

Smart Placement configuration for Functions.

### spec.deploymentConfigs.preview.placement.mode

`string`

Placement mode. "smart" runs Functions near the backends they call. Empty
keeps the default (run near the user).

- rule: placement mode must be "smart"

### spec.deploymentConfigs.preview.vars

`map<string, string>`

Plain-text environment variables exposed to builds and Functions. Map key is
the variable name, value is the literal string. Put secret values in
`secrets` instead.

### spec.deploymentConfigs.preview.secrets

`[]CloudflarePagesSecret`

Secret environment variables exposed to Functions. Each value is secret —
provide a managed-secret reference; resolved just-in-time at deploy.

### spec.deploymentConfigs.preview.secrets[].name

`string` · required

The environment variable name the Function reads the secret through.

- rule: {"required":true}

### spec.deploymentConfigs.preview.secrets[].value

`string | valueFrom` · required · sensitive

The secret value. Provide a managed-secret reference; the platform resolves
it just-in-time at deploy and never stores plaintext.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.deploymentConfigs.preview.kvNamespaces

`[]CloudflarePagesKvBinding`

KV namespace bindings.

### spec.deploymentConfigs.preview.kvNamespaces[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.kvNamespaces[].namespaceId

`string | valueFrom` · required

The KV namespace ID, or a reference to a CloudflareKvNamespace resource.

- references: CloudflareKvNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareKvNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.deploymentConfigs.preview.d1Databases

`[]CloudflarePagesD1Binding`

D1 database bindings.

### spec.deploymentConfigs.preview.d1Databases[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.d1Databases[].databaseId

`string | valueFrom` · required

The D1 database ID, or a reference to a CloudflareD1Database resource.

- references: CloudflareD1Database (`status.outputs.database_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareD1Database, name: <that resource's name>, fieldPath: status.outputs.database_id}} -- a bare string does not parse

### spec.deploymentConfigs.preview.r2Buckets

`[]CloudflarePagesR2Binding`

R2 bucket bindings.

### spec.deploymentConfigs.preview.r2Buckets[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.r2Buckets[].bucketName

`string | valueFrom` · required

The R2 bucket name, or a reference to a CloudflareR2Bucket resource.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.deploymentConfigs.preview.r2Buckets[].jurisdiction

`string`

Optional data-residency jurisdiction of the bucket ("eu", "fedramp",
"fedramp-high"); leave empty for the default jurisdiction. (This binding
vocabulary is the provider's own and deliberately differs from the R2
BUCKET kind's jurisdiction set -- "default", "eu", "fedramp", "us".)

- rule: jurisdiction must be one of "eu", "fedramp", "fedramp-high"

### spec.deploymentConfigs.preview.queueProducers

`[]CloudflarePagesQueueProducerBinding`

Queue producer bindings (enqueue to a Cloudflare Queue).

### spec.deploymentConfigs.preview.queueProducers[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.queueProducers[].queueName

`string | valueFrom` · required

The Cloudflare Queue name, or a reference to a CloudflareQueue resource.

- references: CloudflareQueue (`status.outputs.queue_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareQueue, name: <that resource's name>, fieldPath: status.outputs.queue_name}} -- a bare string does not parse

### spec.deploymentConfigs.preview.hyperdriveBindings

`[]CloudflarePagesHyperdriveBinding`

Hyperdrive bindings (pooled access to a regional SQL database).

### spec.deploymentConfigs.preview.hyperdriveBindings[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.hyperdriveBindings[].configId

`string | valueFrom` · required

The Hyperdrive config ID, or a reference to a CloudflareHyperdriveConfig resource.

- references: CloudflareHyperdriveConfig (`status.outputs.hyperdrive_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareHyperdriveConfig, name: <that resource's name>, fieldPath: status.outputs.hyperdrive_id}} -- a bare string does not parse

### spec.deploymentConfigs.preview.services

`[]CloudflarePagesServiceBinding`

Service bindings to other Workers.

### spec.deploymentConfigs.preview.services[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.services[].service

`string | valueFrom` · required

The target Worker's script name, or a reference to a CloudflareWorker resource.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.deploymentConfigs.preview.services[].entrypoint

`string`

Optional named entrypoint (WorkerEntrypoint) on the target Worker.

### spec.deploymentConfigs.preview.services[].environment

`string`

Optional environment of the target Worker.

### spec.deploymentConfigs.preview.durableObjectNamespaces

`[]CloudflarePagesDurableObjectBinding`

Durable Object namespace bindings.

### spec.deploymentConfigs.preview.durableObjectNamespaces[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.durableObjectNamespaces[].namespaceId

`string` · required

The Durable Object namespace ID to bind.

- rule: {"required":true}

### spec.deploymentConfigs.preview.analyticsEngineDatasets

`[]CloudflarePagesAnalyticsEngineBinding`

Analytics Engine dataset bindings.

### spec.deploymentConfigs.preview.analyticsEngineDatasets[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.analyticsEngineDatasets[].dataset

`string` · required

The Analytics Engine dataset to write data points to.

- rule: {"required":true}

### spec.deploymentConfigs.preview.vectorizeBindings

`[]CloudflarePagesVectorizeBinding`

Vectorize index bindings.

### spec.deploymentConfigs.preview.vectorizeBindings[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.vectorizeBindings[].indexName

`string` · required

The Vectorize index name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.aiBindings

`[]CloudflarePagesAiBinding`

Workers AI / Constellation bindings.

### spec.deploymentConfigs.preview.aiBindings[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.aiBindings[].projectId

`string` · required

The Constellation project ID to bind.

- rule: {"required":true}

### spec.deploymentConfigs.preview.mtlsCertificates

`[]CloudflarePagesMtlsCertificateBinding`

mTLS certificate bindings.

### spec.deploymentConfigs.preview.mtlsCertificates[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.preview.mtlsCertificates[].certificateId

`string` · required

The mTLS certificate ID to bind.

- rule: {"required":true}

### spec.deploymentConfigs.preview.browsers

`[]CloudflarePagesBrowserBinding`

Browser Rendering bindings. Each entry is just the JS binding name.

### spec.deploymentConfigs.preview.browsers[].name

`string` · required

The JS binding (variable) name the Function calls Browser Rendering through.

- rule: {"required":true}

### spec.deploymentConfigs.production

`CloudflarePagesDeploymentConfig`

Configuration applied to production deployments.

### spec.deploymentConfigs.production.compatibilityDate

`string`

Compatibility date (YYYY-MM-DD) pinning the Workers runtime behavior for
Functions.

- rule: {"string":{"pattern":"^([0-9]{4}-[0-9]{2}-[0-9]{2})?$"}}

### spec.deploymentConfigs.production.compatibilityFlags

`[]string`

Runtime compatibility flags (e.g. "nodejs_compat").

### spec.deploymentConfigs.production.alwaysUseLatestCompatibilityDate

`bool`

Always build Functions against the latest compatibility date instead of a
pinned one.

### spec.deploymentConfigs.production.buildImageMajorVersion

`int64`

Major version of the Cloudflare build image to use. Leave 0 for the default.

- rule: build_image_major_version must be 0 (default) or positive

### spec.deploymentConfigs.production.failOpen

`bool`

Continue serving (fail open) when a Function cannot be invoked, rather than
returning an error.

### spec.deploymentConfigs.production.usageModel

`string`

Functions usage/pricing model. One of "standard", "bundled", or "unbound".
Empty uses the account default. ("bundled"/"unbound" are legacy models.)

- rule: usage_model must be one of "standard", "bundled", "unbound"

### spec.deploymentConfigs.production.limits

`CloudflarePagesLimits`

Per-invocation limits for Functions.

### spec.deploymentConfigs.production.limits.cpuMs

`int64`

CPU time limit in milliseconds per invocation. Leave 0 for the plan default.

- rule: cpu_ms must be 0 (default) or positive

### spec.deploymentConfigs.production.placement

`CloudflarePagesPlacement`

Smart Placement configuration for Functions.

### spec.deploymentConfigs.production.placement.mode

`string`

Placement mode. "smart" runs Functions near the backends they call. Empty
keeps the default (run near the user).

- rule: placement mode must be "smart"

### spec.deploymentConfigs.production.vars

`map<string, string>`

Plain-text environment variables exposed to builds and Functions. Map key is
the variable name, value is the literal string. Put secret values in
`secrets` instead.

### spec.deploymentConfigs.production.secrets

`[]CloudflarePagesSecret`

Secret environment variables exposed to Functions. Each value is secret —
provide a managed-secret reference; resolved just-in-time at deploy.

### spec.deploymentConfigs.production.secrets[].name

`string` · required

The environment variable name the Function reads the secret through.

- rule: {"required":true}

### spec.deploymentConfigs.production.secrets[].value

`string | valueFrom` · required · sensitive

The secret value. Provide a managed-secret reference; the platform resolves
it just-in-time at deploy and never stores plaintext.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.deploymentConfigs.production.kvNamespaces

`[]CloudflarePagesKvBinding`

KV namespace bindings.

### spec.deploymentConfigs.production.kvNamespaces[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.kvNamespaces[].namespaceId

`string | valueFrom` · required

The KV namespace ID, or a reference to a CloudflareKvNamespace resource.

- references: CloudflareKvNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareKvNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.deploymentConfigs.production.d1Databases

`[]CloudflarePagesD1Binding`

D1 database bindings.

### spec.deploymentConfigs.production.d1Databases[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.d1Databases[].databaseId

`string | valueFrom` · required

The D1 database ID, or a reference to a CloudflareD1Database resource.

- references: CloudflareD1Database (`status.outputs.database_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareD1Database, name: <that resource's name>, fieldPath: status.outputs.database_id}} -- a bare string does not parse

### spec.deploymentConfigs.production.r2Buckets

`[]CloudflarePagesR2Binding`

R2 bucket bindings.

### spec.deploymentConfigs.production.r2Buckets[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.r2Buckets[].bucketName

`string | valueFrom` · required

The R2 bucket name, or a reference to a CloudflareR2Bucket resource.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.deploymentConfigs.production.r2Buckets[].jurisdiction

`string`

Optional data-residency jurisdiction of the bucket ("eu", "fedramp",
"fedramp-high"); leave empty for the default jurisdiction. (This binding
vocabulary is the provider's own and deliberately differs from the R2
BUCKET kind's jurisdiction set -- "default", "eu", "fedramp", "us".)

- rule: jurisdiction must be one of "eu", "fedramp", "fedramp-high"

### spec.deploymentConfigs.production.queueProducers

`[]CloudflarePagesQueueProducerBinding`

Queue producer bindings (enqueue to a Cloudflare Queue).

### spec.deploymentConfigs.production.queueProducers[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.queueProducers[].queueName

`string | valueFrom` · required

The Cloudflare Queue name, or a reference to a CloudflareQueue resource.

- references: CloudflareQueue (`status.outputs.queue_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareQueue, name: <that resource's name>, fieldPath: status.outputs.queue_name}} -- a bare string does not parse

### spec.deploymentConfigs.production.hyperdriveBindings

`[]CloudflarePagesHyperdriveBinding`

Hyperdrive bindings (pooled access to a regional SQL database).

### spec.deploymentConfigs.production.hyperdriveBindings[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.hyperdriveBindings[].configId

`string | valueFrom` · required

The Hyperdrive config ID, or a reference to a CloudflareHyperdriveConfig resource.

- references: CloudflareHyperdriveConfig (`status.outputs.hyperdrive_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareHyperdriveConfig, name: <that resource's name>, fieldPath: status.outputs.hyperdrive_id}} -- a bare string does not parse

### spec.deploymentConfigs.production.services

`[]CloudflarePagesServiceBinding`

Service bindings to other Workers.

### spec.deploymentConfigs.production.services[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.services[].service

`string | valueFrom` · required

The target Worker's script name, or a reference to a CloudflareWorker resource.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.deploymentConfigs.production.services[].entrypoint

`string`

Optional named entrypoint (WorkerEntrypoint) on the target Worker.

### spec.deploymentConfigs.production.services[].environment

`string`

Optional environment of the target Worker.

### spec.deploymentConfigs.production.durableObjectNamespaces

`[]CloudflarePagesDurableObjectBinding`

Durable Object namespace bindings.

### spec.deploymentConfigs.production.durableObjectNamespaces[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.durableObjectNamespaces[].namespaceId

`string` · required

The Durable Object namespace ID to bind.

- rule: {"required":true}

### spec.deploymentConfigs.production.analyticsEngineDatasets

`[]CloudflarePagesAnalyticsEngineBinding`

Analytics Engine dataset bindings.

### spec.deploymentConfigs.production.analyticsEngineDatasets[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.analyticsEngineDatasets[].dataset

`string` · required

The Analytics Engine dataset to write data points to.

- rule: {"required":true}

### spec.deploymentConfigs.production.vectorizeBindings

`[]CloudflarePagesVectorizeBinding`

Vectorize index bindings.

### spec.deploymentConfigs.production.vectorizeBindings[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.vectorizeBindings[].indexName

`string` · required

The Vectorize index name.

- rule: {"required":true}

### spec.deploymentConfigs.production.aiBindings

`[]CloudflarePagesAiBinding`

Workers AI / Constellation bindings.

### spec.deploymentConfigs.production.aiBindings[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.aiBindings[].projectId

`string` · required

The Constellation project ID to bind.

- rule: {"required":true}

### spec.deploymentConfigs.production.mtlsCertificates

`[]CloudflarePagesMtlsCertificateBinding`

mTLS certificate bindings.

### spec.deploymentConfigs.production.mtlsCertificates[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.deploymentConfigs.production.mtlsCertificates[].certificateId

`string` · required

The mTLS certificate ID to bind.

- rule: {"required":true}

### spec.deploymentConfigs.production.browsers

`[]CloudflarePagesBrowserBinding`

Browser Rendering bindings. Each entry is just the JS binding name.

### spec.deploymentConfigs.production.browsers[].name

`string` · required

The JS binding (variable) name the Function calls Browser Rendering through.

- rule: {"required":true}

### spec.domains

`[]string`

Custom domains to attach to the project. Each must be a hostname in a zone on
this Cloudflare account; Cloudflare provisions and manages its certificate.
The project's `*.pages.dev` subdomain is always available regardless.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflarePagesProject, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.project_name` | `string` | The project name (echoed; downstream resources reference this value). |
| `status.outputs.subdomain` | `string` | The project's `*.pages.dev` subdomain (e.g. "my-site.pages.dev"). |
| `status.outputs.domains` | `[]string` | The custom domains attached to the project. |
| `status.outputs.created_on` | `string` | RFC3339 timestamp of when the project was created. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.deploymentConfigs.preview.kvNamespaces[].namespaceId` | CloudflareKvNamespace | `status.outputs.namespace_id` |
| `spec.deploymentConfigs.preview.d1Databases[].databaseId` | CloudflareD1Database | `status.outputs.database_id` |
| `spec.deploymentConfigs.preview.r2Buckets[].bucketName` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.deploymentConfigs.preview.queueProducers[].queueName` | CloudflareQueue | `status.outputs.queue_name` |
| `spec.deploymentConfigs.preview.hyperdriveBindings[].configId` | CloudflareHyperdriveConfig | `status.outputs.hyperdrive_id` |
| `spec.deploymentConfigs.preview.services[].service` | CloudflareWorker | `status.outputs.script_name` |
| `spec.deploymentConfigs.production.kvNamespaces[].namespaceId` | CloudflareKvNamespace | `status.outputs.namespace_id` |
| `spec.deploymentConfigs.production.d1Databases[].databaseId` | CloudflareD1Database | `status.outputs.database_id` |
| `spec.deploymentConfigs.production.r2Buckets[].bucketName` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.deploymentConfigs.production.queueProducers[].queueName` | CloudflareQueue | `status.outputs.queue_name` |
| `spec.deploymentConfigs.production.hyperdriveBindings[].configId` | CloudflareHyperdriveConfig | `status.outputs.hyperdrive_id` |
| `spec.deploymentConfigs.production.services[].service` | CloudflareWorker | `status.outputs.script_name` |

## See Also

- [Overview](../README.md)
