# CloudflareWorker

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareWorkerSpec deploys a Cloudflare Worker script and everything that
hangs off it: its resource bindings, routing (workers.dev, custom domains,
routes), scheduled (cron) invocations, and runtime settings. Bindings are
grouped by type (the wrangler.toml authoring grain) and each cross-resource
binding accepts a literal value or a reference to the producing resource, so a
Worker composes as a real node in the resource graph.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWorker
metadata:
  name: test-worker
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  workerName: test-worker
  content: |
    export default {
      async fetch(request, env, ctx) {
        return new Response("ok");
      }
    };
  compatibilityDate: "2025-01-15"
  compatibilityFlags:
    - nodejs_compat
  vars:
    LOG_LEVEL: info
  workersDev:
    enabled: true
  observability:
    enabled: true
    headSamplingRate: 1
    # traces.propagationPolicy is offline-only here by design: setting it on
    # an account without the trace-propagation feature fails the deploy with
    # 403 code 100342 ("propagation_policy requires the trace propagation
    # feature to be enabled") -- measured live. The live scenario proves the
    # rest of the traces subtree.
    traces:
      enabled: true
      propagationPolicy: authenticated
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.workerName` | `string` | yes |  |  |
| `spec.compatibilityDate` | `string` |  |  |  |
| `spec.content` | `string` |  |  |  |
| `spec.r2Bundle` | `CloudflareWorkerScriptBundle` |  |  |  |
| `spec.r2Bundle.bucket` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.r2Bundle.path` | `string` | yes |  |  |
| `spec.mainModule` | `string` |  | `index.js` |  |
| `spec.compatibilityFlags` | `[]string` |  |  |  |
| `spec.vars` | `map<string, string>` |  |  |  |
| `spec.secrets` | `[]CloudflareWorkerSecretBinding` |  |  |  |
| `spec.secrets[].name` | `string` | yes |  |  |
| `spec.secrets[].value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.kvNamespaces` | `[]CloudflareWorkerKvBinding` |  |  |  |
| `spec.kvNamespaces[].name` | `string` | yes |  |  |
| `spec.kvNamespaces[].namespaceId` | `string \| valueFrom` | yes |  | CloudflareKvNamespace (`status.outputs.namespace_id`) |
| `spec.r2Buckets` | `[]CloudflareWorkerR2Binding` |  |  |  |
| `spec.r2Buckets[].name` | `string` | yes |  |  |
| `spec.r2Buckets[].bucketName` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.r2Buckets[].jurisdiction` | `string` |  |  |  |
| `spec.d1Databases` | `[]CloudflareWorkerD1Binding` |  |  |  |
| `spec.d1Databases[].name` | `string` | yes |  |  |
| `spec.d1Databases[].databaseId` | `string \| valueFrom` | yes |  | CloudflareD1Database (`status.outputs.database_id`) |
| `spec.hyperdriveConfigs` | `[]CloudflareWorkerHyperdriveBinding` |  |  |  |
| `spec.hyperdriveConfigs[].name` | `string` | yes |  |  |
| `spec.hyperdriveConfigs[].configId` | `string \| valueFrom` | yes |  | CloudflareHyperdriveConfig (`status.outputs.hyperdrive_id`) |
| `spec.services` | `[]CloudflareWorkerServiceBinding` |  |  |  |
| `spec.services[].name` | `string` | yes |  |  |
| `spec.services[].service` | `string \| valueFrom` | yes |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.services[].environment` | `string` |  |  |  |
| `spec.services[].entrypoint` | `string` |  |  |  |
| `spec.queues` | `[]CloudflareWorkerQueueBinding` |  |  |  |
| `spec.queues[].name` | `string` | yes |  |  |
| `spec.queues[].queueName` | `string \| valueFrom` | yes |  | CloudflareQueue (`status.outputs.queue_name`) |
| `spec.durableObjects` | `[]CloudflareWorkerDurableObjectBinding` |  |  |  |
| `spec.durableObjects[].name` | `string` | yes |  |  |
| `spec.durableObjects[].className` | `string` | yes |  |  |
| `spec.durableObjects[].scriptName` | `string \| valueFrom` |  |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.durableObjects[].environment` | `string` |  |  |  |
| `spec.durableObjects[].namespaceId` | `string` |  |  |  |
| `spec.durableObjects[].dispatchNamespace` | `string` |  |  |  |
| `spec.analyticsEngineDatasets` | `[]CloudflareWorkerAnalyticsEngineBinding` |  |  |  |
| `spec.analyticsEngineDatasets[].name` | `string` | yes |  |  |
| `spec.analyticsEngineDatasets[].dataset` | `string` | yes |  |  |
| `spec.vectorizeIndexes` | `[]CloudflareWorkerVectorizeBinding` |  |  |  |
| `spec.vectorizeIndexes[].name` | `string` | yes |  |  |
| `spec.vectorizeIndexes[].indexName` | `string` | yes |  |  |
| `spec.ai` | `[]CloudflareWorkerAiBinding` |  |  |  |
| `spec.ai[].name` | `string` | yes |  |  |
| `spec.versionMetadata` | `[]CloudflareWorkerVersionMetadataBinding` |  |  |  |
| `spec.versionMetadata[].name` | `string` | yes |  |  |
| `spec.workersDev` | `CloudflareWorkerWorkersDev` |  |  |  |
| `spec.workersDev.enabled` | `bool` |  |  |  |
| `spec.workersDev.previewsEnabled` | `bool` |  |  |  |
| `spec.customDomains` | `[]CloudflareWorkerCustomDomain` |  |  |  |
| `spec.customDomains[].hostname` | `string` | yes |  |  |
| `spec.customDomains[].zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.routes` | `[]CloudflareWorkerRoute` |  |  |  |
| `spec.routes[].zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.routes[].pattern` | `string` | yes |  |  |
| `spec.schedules` | `[]string` |  |  |  |
| `spec.observability` | `CloudflareWorkerObservability` |  |  |  |
| `spec.observability.enabled` | `bool` |  |  |  |
| `spec.observability.headSamplingRate` | `double` |  |  |  |
| `spec.observability.logs` | `CloudflareWorkerObservabilityLogs` |  |  |  |
| `spec.observability.logs.enabled` | `bool` |  |  |  |
| `spec.observability.logs.invocationLogs` | `bool` |  |  |  |
| `spec.observability.logs.destinations` | `[]string` |  |  |  |
| `spec.observability.logs.headSamplingRate` | `double` |  |  |  |
| `spec.observability.logs.persist` | `bool` |  |  |  |
| `spec.observability.traces` | `CloudflareWorkerObservabilityTraces` |  |  |  |
| `spec.observability.traces.destinations` | `[]string` |  |  |  |
| `spec.observability.traces.enabled` | `bool` |  |  |  |
| `spec.observability.traces.headSamplingRate` | `double` |  |  |  |
| `spec.observability.traces.persist` | `bool` |  |  |  |
| `spec.observability.traces.propagationPolicy` | `string` |  |  |  |
| `spec.placement` | `CloudflareWorkerPlacement` |  |  |  |
| `spec.placement.mode` | `string` |  |  |  |
| `spec.limits` | `CloudflareWorkerLimits` |  |  |  |
| `spec.limits.cpuMs` | `int64` |  |  |  |
| `spec.limits.subrequests` | `int64` |  |  |  |
| `spec.logpush` | `bool` |  |  |  |
| `spec.tailConsumers` | `[]CloudflareWorkerTailConsumer` |  |  |  |
| `spec.tailConsumers[].service` | `string \| valueFrom` | yes |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.tailConsumers[].environment` | `string` |  |  |  |
| `spec.tailConsumers[].namespace` | `string` |  |  |  |
| `spec.assets` | `CloudflareWorkerAssets` |  |  |  |
| `spec.assets.directory` | `string` | yes |  |  |
| `spec.assets.config` | `CloudflareWorkerAssetsConfig` |  |  |  |
| `spec.assets.config.htmlHandling` | `string` |  |  |  |
| `spec.assets.config.notFoundHandling` | `string` |  |  |  |
| `spec.assets.config.headers` | `string` |  |  |  |
| `spec.assets.config.redirects` | `string` |  |  |  |
| `spec.assets.config.runWorkerFirst` | `bool` |  |  |  |
| `spec.assets.config.runWorkerFirstRules` | `[]string` |  |  |  |
| `spec.assets.bindingName` | `string` |  |  |  |
| `spec.migrations` | `CloudflareWorkerMigrations` |  |  |  |
| `spec.migrations.deletedClasses` | `[]string` |  |  |  |
| `spec.migrations.newClasses` | `[]string` |  |  |  |
| `spec.migrations.newSqliteClasses` | `[]string` |  |  |  |
| `spec.migrations.newTag` | `string` |  |  |  |
| `spec.migrations.oldTag` | `string` |  |  |  |
| `spec.migrations.renamedClasses` | `[]CloudflareWorkerRenamedClass` |  |  |  |
| `spec.migrations.renamedClasses[].from` | `string` |  |  |  |
| `spec.migrations.renamedClasses[].to` | `string` |  |  |  |
| `spec.migrations.transferredClasses` | `[]CloudflareWorkerTransferredClass` |  |  |  |
| `spec.migrations.transferredClasses[].from` | `string` |  |  |  |
| `spec.migrations.transferredClasses[].fromScript` | `string \| valueFrom` |  |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.migrations.transferredClasses[].to` | `string` |  |  |  |
| `spec.migrations.steps` | `[]CloudflareWorkerMigrationStep` |  |  |  |
| `spec.migrations.steps[].deletedClasses` | `[]string` |  |  |  |
| `spec.migrations.steps[].newClasses` | `[]string` |  |  |  |
| `spec.migrations.steps[].newSqliteClasses` | `[]string` |  |  |  |
| `spec.migrations.steps[].renamedClasses` | `[]CloudflareWorkerRenamedClass` |  |  |  |
| `spec.migrations.steps[].renamedClasses[].from` | `string` |  |  |  |
| `spec.migrations.steps[].renamedClasses[].to` | `string` |  |  |  |
| `spec.migrations.steps[].transferredClasses` | `[]CloudflareWorkerTransferredClass` |  |  |  |
| `spec.migrations.steps[].transferredClasses[].from` | `string` |  |  |  |
| `spec.migrations.steps[].transferredClasses[].fromScript` | `string \| valueFrom` |  |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.migrations.steps[].transferredClasses[].to` | `string` |  |  |  |
| `spec.keepAssets` | `bool` |  |  |  |
| `spec.keepBindings` | `[]string` |  |  |  |
| `spec.usageModel` | `string` |  |  |  |
| `spec.cacheOptions` | `CloudflareWorkerCacheOptions` |  |  |  |
| `spec.cacheOptions.enabled` | `bool` |  |  |  |
| `spec.cacheOptions.crossVersionCache` | `bool` |  |  |  |
| `spec.exports` | `map<string, CloudflareWorkerExport>` |  |  |  |
| `spec.exports.*.type` | `string` | yes |  |  |
| `spec.exports.*.cache` | `CloudflareWorkerExportCache` |  |  |  |
| `spec.exports.*.cache.enabled` | `bool` |  |  |  |
| `spec.packageDependencies` | `[]CloudflareWorkerPackageDependency` |  |  |  |
| `spec.packageDependencies[].name` | `string` | yes |  |  |
| `spec.packageDependencies[].installedVersion` | `string` | yes |  |  |
| `spec.packageDependencies[].packageJsonVersion` | `string` | yes |  |  |
| `spec.annotations` | `CloudflareWorkerAnnotations` |  |  |  |
| `spec.annotations.workersMessage` | `string` |  |  |  |
| `spec.annotations.workersTag` | `string` |  |  |  |
| `spec.bodyPart` | `string` |  |  |  |
| `spec.contentType` | `string` |  |  |  |
| `spec.mtlsCertificates` | `[]CloudflareWorkerMtlsCertificateBinding` |  |  |  |
| `spec.mtlsCertificates[].name` | `string` | yes |  |  |
| `spec.mtlsCertificates[].certificateId` | `string` | yes |  |  |
| `spec.dispatchNamespaces` | `[]CloudflareWorkerDispatchNamespaceBinding` |  |  |  |
| `spec.dispatchNamespaces[].name` | `string` | yes |  |  |
| `spec.dispatchNamespaces[].namespace` | `string` | yes |  |  |
| `spec.dispatchNamespaces[].outbound` | `CloudflareWorkerDispatchOutbound` |  |  |  |
| `spec.dispatchNamespaces[].outbound.params` | `[]string` |  |  |  |
| `spec.dispatchNamespaces[].outbound.worker` | `CloudflareWorkerDispatchOutboundWorker` |  |  |  |
| `spec.dispatchNamespaces[].outbound.worker.service` | `string \| valueFrom` |  |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.dispatchNamespaces[].outbound.worker.environment` | `string` |  |  |  |
| `spec.rateLimits` | `[]CloudflareWorkerRateLimitBinding` |  |  |  |
| `spec.rateLimits[].name` | `string` | yes |  |  |
| `spec.rateLimits[].namespace` | `string` | yes |  |  |
| `spec.rateLimits[].simple` | `CloudflareWorkerRateLimitSimple` | yes |  |  |
| `spec.rateLimits[].simple.limit` | `double` | yes |  |  |
| `spec.rateLimits[].simple.period` | `int64` | yes |  |  |
| `spec.rateLimits[].simple.mitigationTimeout` | `int64` |  |  |  |
| `spec.sendEmail` | `[]CloudflareWorkerSendEmailBinding` |  |  |  |
| `spec.sendEmail[].name` | `string` | yes |  |  |
| `spec.sendEmail[].destinationAddress` | `string` |  |  |  |
| `spec.sendEmail[].allowedDestinationAddresses` | `[]string` |  |  |  |
| `spec.sendEmail[].allowedSenderAddresses` | `[]string` |  |  |  |
| `spec.secretsStoreSecrets` | `[]CloudflareWorkerSecretsStoreBinding` |  |  |  |
| `spec.secretsStoreSecrets[].name` | `string` | yes |  |  |
| `spec.secretsStoreSecrets[].storeId` | `string` | yes |  |  |
| `spec.secretsStoreSecrets[].secretName` | `string` | yes |  |  |
| `spec.secretKeys` | `[]CloudflareWorkerSecretKeyBinding` |  |  |  |
| `spec.secretKeys[].name` | `string` | yes |  |  |
| `spec.secretKeys[].algorithm` | `string` | yes |  |  |
| `spec.secretKeys[].format` | `string` | yes |  |  |
| `spec.secretKeys[].usages` | `[]string` |  |  |  |
| `spec.secretKeys[].keyBase64` | `string` (sensitive) |  |  |  |
| `spec.secretKeys[].keyJwk` | `string` (sensitive) |  |  |  |
| `spec.workflows` | `[]CloudflareWorkerWorkflowBinding` |  |  |  |
| `spec.workflows[].name` | `string` | yes |  |  |
| `spec.workflows[].workflowName` | `string` | yes |  |  |
| `spec.pipelines` | `[]CloudflareWorkerPipelineBinding` |  |  |  |
| `spec.pipelines[].name` | `string` | yes |  |  |
| `spec.pipelines[].pipeline` | `string` | yes |  |  |
| `spec.jsonBindings` | `[]CloudflareWorkerJsonBinding` |  |  |  |
| `spec.jsonBindings[].name` | `string` | yes |  |  |
| `spec.jsonBindings[].json` | `string` | yes |  |  |
| `spec.inheritBindings` | `[]CloudflareWorkerInheritBinding` |  |  |  |
| `spec.inheritBindings[].name` | `string` | yes |  |  |
| `spec.inheritBindings[].oldName` | `string` |  |  |  |
| `spec.inheritBindings[].versionId` | `string` |  |  |  |
| `spec.dataBlobs` | `[]CloudflareWorkerBlobBinding` |  |  |  |
| `spec.dataBlobs[].name` | `string` | yes |  |  |
| `spec.dataBlobs[].part` | `string` | yes |  |  |
| `spec.textBlobs` | `[]CloudflareWorkerBlobBinding` |  |  |  |
| `spec.textBlobs[].name` | `string` | yes |  |  |
| `spec.textBlobs[].part` | `string` | yes |  |  |
| `spec.browsers` | `[]CloudflareWorkerNamedBinding` |  |  |  |
| `spec.browsers[].name` | `string` | yes |  |  |
| `spec.aiSearch` | `[]CloudflareWorkerAiSearchBinding` |  |  |  |
| `spec.aiSearch[].name` | `string` | yes |  |  |
| `spec.aiSearch[].instanceName` | `string` | yes |  |  |
| `spec.aiSearch[].namespace` | `string` |  |  |  |
| `spec.aiSearch[].appId` | `string` |  |  |  |
| `spec.aiSearchNamespaces` | `[]CloudflareWorkerAiSearchNamespaceBinding` |  |  |  |
| `spec.aiSearchNamespaces[].name` | `string` | yes |  |  |
| `spec.aiSearchNamespaces[].namespace` | `string` | yes |  |  |
| `spec.images` | `[]CloudflareWorkerNamedBinding` |  |  |  |
| `spec.images[].name` | `string` | yes |  |  |
| `spec.media` | `[]CloudflareWorkerNamedBinding` |  |  |  |
| `spec.media[].name` | `string` | yes |  |  |
| `spec.wasmModules` | `[]CloudflareWorkerBlobBinding` |  |  |  |
| `spec.wasmModules[].name` | `string` | yes |  |  |
| `spec.wasmModules[].part` | `string` | yes |  |  |
| `spec.vpcServices` | `[]CloudflareWorkerVpcServiceBinding` |  |  |  |
| `spec.vpcServices[].name` | `string` | yes |  |  |
| `spec.vpcServices[].serviceId` | `string` | yes |  |  |
| `spec.vpcNetworks` | `[]CloudflareWorkerVpcNetworkBinding` |  |  |  |
| `spec.vpcNetworks[].name` | `string` | yes |  |  |
| `spec.vpcNetworks[].networkId` | `string` |  |  |  |
| `spec.vpcNetworks[].tunnelId` | `string \| valueFrom` |  |  | CloudflareZeroTrustTunnel (`status.outputs.tunnel_id`) |
| `spec.tailConsumerBindings` | `[]CloudflareWorkerTailConsumerBinding` |  |  |  |
| `spec.tailConsumerBindings[].name` | `string` | yes |  |  |
| `spec.tailConsumerBindings[].service` | `string \| valueFrom` | yes |  | CloudflareWorker (`status.outputs.script_name`) |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID in which to create the worker.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.workerName

`string` · required

The Worker script name, visible in the Cloudflare dashboard and used as the
target of service bindings and routes.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63"}}

### spec.compatibilityDate

`string`

Compatibility date (YYYY-MM-DD) pinning the Workers runtime behavior. When
unset, the module pins today's date at deploy.

- rule: {"string":{"pattern":"^([0-9]{4}-[0-9]{2}-[0-9]{2})?$"}}

### spec.content

`string`

Inline ES-module source. Convenient for small or generated scripts.

### spec.r2Bundle

`CloudflareWorkerScriptBundle`

A pre-built script bundle stored in an R2 bucket (the CI build-artifact
flow). The module fetches the object and deploys it.

### spec.r2Bundle.bucket

`string | valueFrom` · required

The R2 bucket that holds the bundle, or a reference to a CloudflareR2Bucket.
Widened from a plain string so a Worker can compose against a bucket Planton manages.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.r2Bundle.path

`string` · required

The object key (path) of the script bundle within the bucket.

- rule: {"required":true}

### spec.mainModule

`string`

The entrypoint module filename within the bundle (module-format workers).
Both engines default it to "index.js" when unset (unless body_part marks
service-worker syntax): an upload without a main module is treated by
Cloudflare as a legacy service-worker script and ES-module code then
fails deploy with "Uncaught SyntaxError: Unexpected token 'export'"
(measured live).

- default: `index.js`

### spec.compatibilityFlags

`[]string`

Runtime compatibility flags (e.g. "nodejs_compat"). See the Cloudflare
compatibility-flags docs for the set valid on a given compatibility_date.

### spec.vars

`map<string, string>`

Plain-text variable bindings: non-secret configuration exposed to the Worker
as env vars. Map key is the JS binding name, value is the literal string.

### spec.secrets

`[]CloudflareWorkerSecretBinding`

Secret bindings: sensitive values exposed to the Worker as env vars. Provide
each as a managed-secret reference; resolved just-in-time at deploy.

### spec.secrets[].name

`string` · required

The JS binding (variable) name the Worker reads the secret through.

- rule: {"required":true}

### spec.secrets[].value

`string | valueFrom` · required · sensitive

The secret value. Provide a managed-secret reference; the platform resolves
it just-in-time at deploy and never stores plaintext.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.kvNamespaces

`[]CloudflareWorkerKvBinding`

KV namespace bindings. Each binds a CloudflareKvNamespace (by id or
reference) to a JS variable the Worker uses for edge key-value access.

### spec.kvNamespaces[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.kvNamespaces[].namespaceId

`string | valueFrom` · required

The KV namespace ID, or a reference to a CloudflareKvNamespace resource.

- references: CloudflareKvNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareKvNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.r2Buckets

`[]CloudflareWorkerR2Binding`

R2 bucket bindings. Each binds a CloudflareR2Bucket (by name or reference)
for object storage access from the Worker.

### spec.r2Buckets[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.r2Buckets[].bucketName

`string | valueFrom` · required

The R2 bucket name, or a reference to a CloudflareR2Bucket resource.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.r2Buckets[].jurisdiction

`string`

Optional data-residency jurisdiction of the bucket. One of "eu", "fedramp",
or "fedramp-high"; leave empty for the default jurisdiction. (This binding
vocabulary is the provider's own and deliberately differs from the R2
BUCKET kind's jurisdiction set -- "default", "eu", "fedramp", "us".)

- rule: jurisdiction must be one of "eu", "fedramp", "fedramp-high"

### spec.d1Databases

`[]CloudflareWorkerD1Binding`

D1 database bindings. Each binds a CloudflareD1Database (by id or reference)
for serverless SQL access from the Worker.

### spec.d1Databases[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.d1Databases[].databaseId

`string | valueFrom` · required

The D1 database ID, or a reference to a CloudflareD1Database resource.

- references: CloudflareD1Database (`status.outputs.database_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareD1Database, name: <that resource's name>, fieldPath: status.outputs.database_id}} -- a bare string does not parse

### spec.hyperdriveConfigs

`[]CloudflareWorkerHyperdriveBinding`

Hyperdrive bindings. Each binds a CloudflareHyperdriveConfig (by id or
reference) for pooled, cached access to a regional SQL database.

### spec.hyperdriveConfigs[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.hyperdriveConfigs[].configId

`string | valueFrom` · required

The Hyperdrive config ID, or a reference to a CloudflareHyperdriveConfig resource.

- references: CloudflareHyperdriveConfig (`status.outputs.hyperdrive_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareHyperdriveConfig, name: <that resource's name>, fieldPath: status.outputs.hyperdrive_id}} -- a bare string does not parse

### spec.services

`[]CloudflareWorkerServiceBinding`

Service bindings: bind another Worker (by name or reference) for direct
worker-to-worker calls without going over the public internet.

### spec.services[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.services[].service

`string | valueFrom` · required

The target Worker's script name, or a reference to a CloudflareWorker resource.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.services[].environment

`string`

Optional environment of the target Worker to bind to.

### spec.services[].entrypoint

`string`

Optional named entrypoint (WorkerEntrypoint) to invoke on the target Worker.

### spec.queues

`[]CloudflareWorkerQueueBinding`

Queue producer bindings: let the Worker enqueue messages to a Cloudflare
Queue by name.

### spec.queues[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.queues[].queueName

`string | valueFrom` · required

The Cloudflare Queue name to produce to, or a reference to a CloudflareQueue resource.

- references: CloudflareQueue (`status.outputs.queue_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareQueue, name: <that resource's name>, fieldPath: status.outputs.queue_name}} -- a bare string does not parse

### spec.durableObjects

`[]CloudflareWorkerDurableObjectBinding`

Durable Object bindings: bind a Durable Object namespace (a class) for
strongly-consistent, stateful coordination.

### spec.durableObjects[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.durableObjects[].className

`string` · required

The Durable Object class name that implements the namespace.

- rule: {"required":true}

### spec.durableObjects[].scriptName

`string | valueFrom`

Script that defines the class, when it lives in a different Worker. Leave
empty when the class is defined in this Worker. Widened to a Worker FK so
a Durable Object defined in another Planton Worker composes as a graph edge.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.durableObjects[].environment

`string`

Optional environment of the defining script.

### spec.durableObjects[].namespaceId

`string`

Durable Object namespace identifier. Optional alternative to class_name
when binding an already-created namespace by id.

### spec.durableObjects[].dispatchNamespace

`string`

Dispatch namespace the Durable Object script belongs to (Workers for Platforms).

### spec.analyticsEngineDatasets

`[]CloudflareWorkerAnalyticsEngineBinding`

Analytics Engine bindings: bind a dataset the Worker writes analytics data points to.

### spec.analyticsEngineDatasets[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.analyticsEngineDatasets[].dataset

`string` · required

The Analytics Engine dataset the Worker writes data points to.

- rule: {"required":true}

### spec.vectorizeIndexes

`[]CloudflareWorkerVectorizeBinding`

Vectorize bindings: bind a vector-search index for AI/embedding workloads.

### spec.vectorizeIndexes[].name

`string` · required

The JS binding (variable) name.

- rule: {"required":true}

### spec.vectorizeIndexes[].indexName

`string` · required

The Vectorize index name.

- rule: {"required":true}

### spec.ai

`[]CloudflareWorkerAiBinding`

Workers AI bindings: bind the account's AI inference gateway. Each entry is
just the JS variable name the Worker calls models through.

### spec.ai[].name

`string` · required

The JS binding (variable) name the Worker calls AI models through.

- rule: {"required":true}

### spec.versionMetadata

`[]CloudflareWorkerVersionMetadataBinding`

Version-metadata bindings: expose the deployed version id/tag to the Worker
at runtime via the named binding.

### spec.versionMetadata[].name

`string` · required

The JS binding (variable) name exposing the version id/tag.

- rule: {"required":true}

### spec.workersDev

`CloudflareWorkerWorkersDev`

workers.dev subdomain exposure (e.g. <name>.<subdomain>.workers.dev).

### spec.workersDev.enabled

`bool`

Expose the Worker at <name>.<account-subdomain>.workers.dev.

### spec.workersDev.previewsEnabled

`bool`

Also enable per-version preview URLs (<version>-<name>.<subdomain>.workers.dev).

### spec.customDomains

`[]CloudflareWorkerCustomDomain`

Custom domains that route directly to this Worker (a managed hostname with
automatic TLS), distinct from pattern-based routes.

### spec.customDomains[].hostname

`string` · required

The fully qualified hostname to route to the Worker (e.g. "api.example.com").
Cloudflare provisions and manages TLS for it automatically.

- rule: {"required":true}

### spec.customDomains[].zoneId

`string | valueFrom`

The Cloudflare Zone ID hosting the hostname, or a reference to a
CloudflareDnsZone. Optional — Cloudflare can infer the zone from the hostname.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.routes

`[]CloudflareWorkerRoute`

Route patterns within a zone that map matching requests to this Worker.

### spec.routes[].zoneId

`string | valueFrom` · required

The Cloudflare Zone ID the route belongs to, or a reference to a CloudflareDnsZone.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.routes[].pattern

`string` · required

The route pattern (e.g. "api.example.com/webhooks/*").

- rule: {"required":true}

### spec.schedules

`[]string`

Cron schedules that invoke the Worker's scheduled handler. Each entry is a
cron expression (e.g. "0 * * * *").

### spec.observability

`CloudflareWorkerObservability`

Observability (Workers Logs) configuration.

### spec.observability.enabled

`bool`

Enable Workers Logs (structured logs visible in the dashboard / queryable).

### spec.observability.headSamplingRate

`double`

Fraction of requests sampled for logging, 0.0–1.0. Leave 0 for the default
(1.0 = sample all) when enabled.

- rule: head_sampling_rate must be between 0 and 1

### spec.observability.logs

`CloudflareWorkerObservabilityLogs`

Nested log export / persistence settings. Distinct from the top-level
enabled + head_sampling_rate pair — Cloudflare lets you turn logs on
independently of the overall observability switch.

### spec.observability.logs.enabled

`bool`

### spec.observability.logs.invocationLogs

`bool`

### spec.observability.logs.destinations

`[]string`

### spec.observability.logs.headSamplingRate

`double`

- rule: head_sampling_rate must be between 0 and 1

### spec.observability.logs.persist

`bool`

### spec.observability.traces

`CloudflareWorkerObservabilityTraces`

Nested trace export / persistence settings.

### spec.observability.traces.destinations

`[]string`

### spec.observability.traces.enabled

`bool`

### spec.observability.traces.headSamplingRate

`double`

- rule: head_sampling_rate must be between 0 and 1

### spec.observability.traces.persist

`bool`

### spec.observability.traces.propagationPolicy

`string`

How inbound traceparent/tracestate headers are handled: "authenticated"
(default) or "accept".
Requires the account's trace-propagation feature: without it Cloudflare
rejects the whole script upload with 403 code 100342 ("propagation_policy
requires the trace propagation feature to be enabled") -- measured live.
Leave unset unless the account carries the feature.

- rule: propagation_policy must be "authenticated" or "accept"

### spec.placement

`CloudflareWorkerPlacement`

Smart Placement: let Cloudflare run the Worker closer to backend services it
calls, instead of closest to the user.

### spec.placement.mode

`string`

Placement mode. "smart" lets Cloudflare run the Worker near the backends it
calls. "targeted" pins it (region/hostname/host are computed-only on the
provider and are not authorable here). Leave empty to keep the default
(run near the user).

- rule: placement mode must be "smart" or "targeted"

### spec.limits

`CloudflareWorkerLimits`

Per-invocation resource limits.

### spec.limits.cpuMs

`int64`

CPU time limit in milliseconds per invocation. Leave 0 for the plan default.

- rule: cpu_ms must be 0 (default) or positive

### spec.limits.subrequests

`int64`

Maximum number of subrequests the Worker may make per invocation. Leave 0
for the plan default.

- rule: subrequests must be 0 (default) or positive

### spec.logpush

`bool`

Enable Logpush for the Worker's logs (push to a configured Logpush job).

### spec.tailConsumers

`[]CloudflareWorkerTailConsumer`

Tail consumers: other Workers that receive this Worker's tail (trace) events.

### spec.tailConsumers[].service

`string | valueFrom` · required

The consuming Worker's service (script) name, or a reference to a
CloudflareWorker. Widened from a plain string so tail consumers compose
as graph edges.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.tailConsumers[].environment

`string`

Optional environment of the consuming Worker.

### spec.tailConsumers[].namespace

`string`

Optional dispatch namespace of the consuming Worker.

### spec.assets

`CloudflareWorkerAssets`

Static assets served by this Worker (Cloudflare Workers Static Assets). Point
it at a built site directory to host a static site or single-page app at the
edge; combine it with a script source above to run a full-stack app whose
Functions handle dynamic routes while everything else is served as a static
asset. When assets are set without a script source, the Worker is a pure
static site. This is Cloudflare's converged successor to Pages for the
"build-and-upload" hosting model.

### spec.assets.directory

`string` · required

Filesystem path to the directory of built assets to upload. Interpreted on
the deploy runner — point it at your build output (e.g. "dist" or "build").

- rule: {"required":true}

### spec.assets.config

`CloudflareWorkerAssetsConfig`

Behavior for serving the assets (header injection, redirects, HTML routing,
SPA handling). Optional; sensible Cloudflare defaults apply when omitted.

- rule: set run_worker_first (apply to all paths) or run_worker_first_rules (per-path), not both

### spec.assets.config.htmlHandling

`string`

Trailing-slash and rewrite behavior for HTML requests. One of
"auto-trailing-slash", "force-trailing-slash", "drop-trailing-slash", or
"none". Empty uses Cloudflare's default.

- rule: html_handling must be one of "auto-trailing-slash", "force-trailing-slash", "drop-trailing-slash", "none"

### spec.assets.config.notFoundHandling

`string`

Response when a request matches no static asset and no Worker script handles
it. One of "none", "404-page" (serve /404.html), or "single-page-application"
(serve /index.html — the SPA fallback). Empty uses Cloudflare's default.

- rule: not_found_handling must be one of "none", "404-page", "single-page-application"

### spec.assets.config.headers

`string`

The contents of a `_headers` file: rules that attach custom headers to asset
responses. See Cloudflare's _headers syntax.

### spec.assets.config.redirects

`string`

The contents of a `_redirects` file: redirect/rewrite rules applied ahead of
asset serving. See Cloudflare's _redirects syntax.

### spec.assets.config.runWorkerFirst

`bool`

Invoke the Worker script before attempting to serve any asset (applies to all
paths). Use this when the script must run on every request. Mutually
exclusive with run_worker_first_rules. When neither is set, Cloudflare serves
a matching asset first and falls back to the script.

### spec.assets.config.runWorkerFirstRules

`[]string`

Per-path control over whether the Worker runs before asset serving. Each
entry is a path rule; "/api/*" routes through the script, "!/api/docs/*"
(negative, higher precedence) excludes a subtree. Mutually exclusive with
run_worker_first.

### spec.assets.bindingName

`string`

When set, exposes the asset namespace to the Worker script as a binding of
this JS variable name (so code can call e.g. env.ASSETS.fetch(request)).
Leave empty for a pure static site that has no script. Only meaningful when
a script source is also provided.

### spec.migrations

`CloudflareWorkerMigrations`

Durable Object migrations applied on this upload. Cloudflare rejects a
second apply of the same tag, so this is a one-shot plan — not something
the live E2E harness can re-apply. Author it, plan it, prove it later by
hand if needed.

### spec.migrations.deletedClasses

`[]string`

### spec.migrations.newClasses

`[]string`

### spec.migrations.newSqliteClasses

`[]string`

### spec.migrations.newTag

`string`

### spec.migrations.oldTag

`string`

### spec.migrations.renamedClasses

`[]CloudflareWorkerRenamedClass`

### spec.migrations.renamedClasses[].from

`string`

### spec.migrations.renamedClasses[].to

`string`

### spec.migrations.transferredClasses

`[]CloudflareWorkerTransferredClass`

### spec.migrations.transferredClasses[].from

`string`

### spec.migrations.transferredClasses[].fromScript

`string | valueFrom`

Source Worker that currently owns the class. A CloudflareWorker FK so a
transfer composes as a graph edge.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.migrations.transferredClasses[].to

`string`

### spec.migrations.steps

`[]CloudflareWorkerMigrationStep`

### spec.migrations.steps[].deletedClasses

`[]string`

### spec.migrations.steps[].newClasses

`[]string`

### spec.migrations.steps[].newSqliteClasses

`[]string`

### spec.migrations.steps[].renamedClasses

`[]CloudflareWorkerRenamedClass`

### spec.migrations.steps[].renamedClasses[].from

`string`

### spec.migrations.steps[].renamedClasses[].to

`string`

### spec.migrations.steps[].transferredClasses

`[]CloudflareWorkerTransferredClass`

### spec.migrations.steps[].transferredClasses[].from

`string`

### spec.migrations.steps[].transferredClasses[].fromScript

`string | valueFrom`

Source Worker that currently owns the class. A CloudflareWorker FK so a
transfer composes as a graph edge.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.migrations.steps[].transferredClasses[].to

`string`

### spec.keepAssets

`bool`

Keep previously uploaded assets instead of sending a new bundle. An
explicit assets upload wins over this flag.

### spec.keepBindings

`[]string`

Binding types to keep from the previous upload (e.g. "secret_text") so a
redeploy does not have to resend every secret.

### spec.usageModel

`string`

Usage model for invocations: "standard", "bundled", or "unbound". Empty
lets Cloudflare default to "standard".

- rule: usage_model must be one of "standard", "bundled", "unbound"

### spec.cacheOptions

`CloudflareWorkerCacheOptions`

Global CacheW settings. Pulumi's Cloudflare SDK v6.17.0 has no matching
input — tofu honors this; Pulumi logs a PARITY-EXCEPTION and skips it.

### spec.cacheOptions.enabled

`bool`

### spec.cacheOptions.crossVersionCache

`bool`

### spec.exports

`map<string, CloudflareWorkerExport>`

Per-entrypoint export configuration. Map key is the export name. Pulumi
SDK v6.17.0 has no matching input — tofu honors this; Pulumi skips it.

### spec.exports.*.type

`string` · required

- rule: {"required":true}

### spec.exports.*.cache

`CloudflareWorkerExportCache`

### spec.exports.*.cache.enabled

`bool`

### spec.packageDependencies

`[]CloudflareWorkerPackageDependency`

npm packages recorded against this Worker build. Pulumi SDK v6.17.0 has
no matching input — tofu honors this; Pulumi skips it.

### spec.packageDependencies[].name

`string` · required

- rule: {"required":true}

### spec.packageDependencies[].installedVersion

`string` · required

- rule: {"required":true}

### spec.packageDependencies[].packageJsonVersion

`string` · required

- rule: {"required":true}

### spec.annotations

`CloudflareWorkerAnnotations`

Version annotations written on this upload (message + tag). The
server-set workers_triggered_by leaf is computed-only and omitted.

### spec.annotations.workersMessage

`string`

### spec.annotations.workersTag

`string`

### spec.bodyPart

`string`

Uploaded file that contains the service-worker-syntax script (the file
that adds a fetch listener). Mutually exclusive with main_module.

### spec.contentType

`string`

Content-Type of the Worker. Required when uploading a non-JavaScript
Worker (e.g. Python). Empty is JavaScript.

- rule: content_type must be one of the five Cloudflare Worker MIME types

### spec.mtlsCertificates

`[]CloudflareWorkerMtlsCertificateBinding`

mTLS certificate bindings. certificate_id is a literal — Planton has no
Cloudflare mTLS kind yet.

### spec.mtlsCertificates[].name

`string` · required

- rule: {"required":true}

### spec.mtlsCertificates[].certificateId

`string` · required

Literal certificate id — Planton has no mTLS kind yet.

- rule: {"required":true}

### spec.dispatchNamespaces

`[]CloudflareWorkerDispatchNamespaceBinding`

Dispatch-namespace bindings (Workers for Platforms). The outbound worker
service is a CloudflareWorker FK when set.

### spec.dispatchNamespaces[].name

`string` · required

- rule: {"required":true}

### spec.dispatchNamespaces[].namespace

`string` · required

- rule: {"required":true}

### spec.dispatchNamespaces[].outbound

`CloudflareWorkerDispatchOutbound`

### spec.dispatchNamespaces[].outbound.params

`[]string`

### spec.dispatchNamespaces[].outbound.worker

`CloudflareWorkerDispatchOutboundWorker`

### spec.dispatchNamespaces[].outbound.worker.service

`string | valueFrom`

Outbound Worker name, or a reference to a CloudflareWorker.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.dispatchNamespaces[].outbound.worker.environment

`string`

### spec.rateLimits

`[]CloudflareWorkerRateLimitBinding`

Rate-limit bindings (the Workers rate-limiting API).

### spec.rateLimits[].name

`string` · required

- rule: {"required":true}

### spec.rateLimits[].namespace

`string` · required

Rate-limit namespace id (a Cloudflare-side identifier, not a Planton kind).

- rule: {"required":true}

### spec.rateLimits[].simple

`CloudflareWorkerRateLimitSimple` · required

- rule: {"required":true}

### spec.rateLimits[].simple.limit

`double` · required

- rule: {"required":true}

### spec.rateLimits[].simple.period

`int64` · required

- rule: {"required":true}

### spec.rateLimits[].simple.mitigationTimeout

`int64`

### spec.sendEmail

`[]CloudflareWorkerSendEmailBinding`

send_email bindings: let the Worker send email through Email Routing.

### spec.sendEmail[].name

`string` · required

- rule: {"required":true}

### spec.sendEmail[].destinationAddress

`string`

### spec.sendEmail[].allowedDestinationAddresses

`[]string`

### spec.sendEmail[].allowedSenderAddresses

`[]string`

### spec.secretsStoreSecrets

`[]CloudflareWorkerSecretsStoreBinding`

Secrets Store bindings. store_id and secret_name are literals — Planton
has no Secrets Store kind yet.

### spec.secretsStoreSecrets[].name

`string` · required

- rule: {"required":true}

### spec.secretsStoreSecrets[].storeId

`string` · required

- rule: {"required":true}

### spec.secretsStoreSecrets[].secretName

`string` · required

- rule: {"required":true}

### spec.secretKeys

`[]CloudflareWorkerSecretKeyBinding`

SubtleCrypto secret-key bindings. Material fields are sensitive; the
wrapper list is exempt like `secrets`.

### spec.secretKeys[].name

`string` · required

- rule: {"required":true}

### spec.secretKeys[].algorithm

`string` · required

- rule: {"required":true}

### spec.secretKeys[].format

`string` · required

- rule: format must be one of "raw", "pkcs8", "spki", "jwk"
- rule: {"required":true}

### spec.secretKeys[].usages

`[]string`

### spec.secretKeys[].keyBase64

`string` · sensitive

### spec.secretKeys[].keyJwk

`string` · sensitive

### spec.workflows

`[]CloudflareWorkerWorkflowBinding`

Workflow bindings. workflow_name is a literal — Planton has no Workflows kind yet.

### spec.workflows[].name

`string` · required

- rule: {"required":true}

### spec.workflows[].workflowName

`string` · required

- rule: {"required":true}

### spec.pipelines

`[]CloudflareWorkerPipelineBinding`

Pipeline bindings. pipeline is a literal — Planton has no Pipelines kind yet.

### spec.pipelines[].name

`string` · required

- rule: {"required":true}

### spec.pipelines[].pipeline

`string` · required

- rule: {"required":true}

### spec.jsonBindings

`[]CloudflareWorkerJsonBinding`

JSON bindings: structured config exposed to the Worker as a parsed object.

### spec.jsonBindings[].name

`string` · required

- rule: {"required":true}

### spec.jsonBindings[].json

`string` · required

- rule: {"required":true}

### spec.inheritBindings

`[]CloudflareWorkerInheritBinding`

Inherit bindings: copy a binding from a previous version (optionally renamed).

### spec.inheritBindings[].name

`string` · required

- rule: {"required":true}

### spec.inheritBindings[].oldName

`string`

### spec.inheritBindings[].versionId

`string`

### spec.dataBlobs

`[]CloudflareWorkerBlobBinding`

data_blob bindings (service-worker syntax). `part` is the uploaded file name.

### spec.dataBlobs[].name

`string` · required

- rule: {"required":true}

### spec.dataBlobs[].part

`string` · required

- rule: {"required":true}

### spec.textBlobs

`[]CloudflareWorkerBlobBinding`

text_blob bindings (service-worker syntax). `part` is the uploaded file name.

### spec.textBlobs[].name

`string` · required

- rule: {"required":true}

### spec.textBlobs[].part

`string` · required

- rule: {"required":true}

### spec.browsers

`[]CloudflareWorkerNamedBinding`

Browser Rendering bindings. Each entry is just the JS variable name.

### spec.browsers[].name

`string` · required

- rule: {"required":true}

### spec.aiSearch

`[]CloudflareWorkerAiSearchBinding`

AI Search instance bindings. instance_name is a literal — Planton has no
AI Search kind yet. app_id is the Flagship leaf the provider hangs on the
same binding object.

### spec.aiSearch[].name

`string` · required

- rule: {"required":true}

### spec.aiSearch[].instanceName

`string` · required

- rule: {"required":true}

### spec.aiSearch[].namespace

`string`

Namespace the instance belongs to. Cloudflare defaults this to "default".

### spec.aiSearch[].appId

`string`

Flagship app id — the provider hangs this leaf on the same binding object.

### spec.aiSearchNamespaces

`[]CloudflareWorkerAiSearchNamespaceBinding`

AI Search namespace bindings.

### spec.aiSearchNamespaces[].name

`string` · required

- rule: {"required":true}

### spec.aiSearchNamespaces[].namespace

`string` · required

- rule: {"required":true}

### spec.images

`[]CloudflareWorkerNamedBinding`

Images bindings. Each entry is just the JS variable name.

### spec.images[].name

`string` · required

- rule: {"required":true}

### spec.media

`[]CloudflareWorkerNamedBinding`

Media bindings. Each entry is just the JS variable name.

### spec.media[].name

`string` · required

- rule: {"required":true}

### spec.wasmModules

`[]CloudflareWorkerBlobBinding`

wasm_module bindings (service-worker syntax). `part` is the uploaded file name.

### spec.wasmModules[].name

`string` · required

- rule: {"required":true}

### spec.wasmModules[].part

`string` · required

- rule: {"required":true}

### spec.vpcServices

`[]CloudflareWorkerVpcServiceBinding`

VPC service bindings. service_id is a literal — Planton has no VPC-service kind yet.

### spec.vpcServices[].name

`string` · required

- rule: {"required":true}

### spec.vpcServices[].serviceId

`string` · required

- rule: {"required":true}

### spec.vpcNetworks

`[]CloudflareWorkerVpcNetworkBinding`

VPC network bindings. network_id is a literal ("cf1:network"); tunnel_id
is a CloudflareZeroTrustTunnel FK. The two are mutually exclusive.

- rule: set network_id or tunnel_id, not both

### spec.vpcNetworks[].name

`string` · required

- rule: {"required":true}

### spec.vpcNetworks[].networkId

`string`

Literal network id. Only "cf1:network" is currently supported by Cloudflare.

### spec.vpcNetworks[].tunnelId

`string | valueFrom`

Cloudflare Tunnel to bind, or a reference to a CloudflareZeroTrustTunnel.

- references: CloudflareZeroTrustTunnel (`status.outputs.tunnel_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnel, name: <that resource's name>, fieldPath: status.outputs.tunnel_id}} -- a bare string does not parse

### spec.tailConsumerBindings

`[]CloudflareWorkerTailConsumerBinding`

tail_consumer *bindings* (type=tail_consumer on the script). Distinct from
the top-level tail_consumers list, which names Workers that consume this
Worker's logs. This list binds another Worker as a tail consumer resource
the script can call.

### spec.tailConsumerBindings[].name

`string` · required

- rule: {"required":true}

### spec.tailConsumerBindings[].service

`string | valueFrom` · required

Worker that is the tail-consumer resource, or a reference to a CloudflareWorker.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

## Validation Rules

- `spec.code_or_assets_required`: provide a script source (content or r2_bundle) and/or a static-asset directory (assets)
- `spec.body_part_xor_main_module`: set body_part (service-worker syntax) or main_module (ES-module syntax), not both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareWorker, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.script_id` | `string` | The Cloudflare-assigned identifier of the deployed Worker script (equals the script name in v5). |
| `status.outputs.script_name` | `string` | The Worker script name — the target a service binding references. |
| `status.outputs.custom_domain_hostnames` | `[]string` | The custom-domain hostnames attached to this Worker. |
| `status.outputs.route_patterns` | `[]string` | The route patterns mapped to this Worker. |
| `status.outputs.custom_domain_ids` | `map<string, string>` | Cloudflare-assigned custom-domain ids, keyed by hostname (the for_each key). Needed so import can address cloudflare_workers_custom_domain as {account_id}/{domain_id}. |
| `status.outputs.route_ids` | `map<string, string>` | Cloudflare-assigned route ids, keyed by the same index key the module uses for for_each (the list index as a string). Needed so import can address cloudflare_workers_route as {zone_id}/{route_id}. |
| `status.outputs.route_zone_ids` | `map<string, string>` | Zone id of each route, keyed the same way as route_ids. Import needs both halves of {zone_id}/{route_id}. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.r2Bundle.bucket` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.kvNamespaces[].namespaceId` | CloudflareKvNamespace | `status.outputs.namespace_id` |
| `spec.r2Buckets[].bucketName` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.d1Databases[].databaseId` | CloudflareD1Database | `status.outputs.database_id` |
| `spec.hyperdriveConfigs[].configId` | CloudflareHyperdriveConfig | `status.outputs.hyperdrive_id` |
| `spec.services[].service` | CloudflareWorker | `status.outputs.script_name` |
| `spec.queues[].queueName` | CloudflareQueue | `status.outputs.queue_name` |
| `spec.durableObjects[].scriptName` | CloudflareWorker | `status.outputs.script_name` |
| `spec.customDomains[].zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.routes[].zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.tailConsumers[].service` | CloudflareWorker | `status.outputs.script_name` |
| `spec.migrations.transferredClasses[].fromScript` | CloudflareWorker | `status.outputs.script_name` |
| `spec.migrations.steps[].transferredClasses[].fromScript` | CloudflareWorker | `status.outputs.script_name` |
| `spec.dispatchNamespaces[].outbound.worker.service` | CloudflareWorker | `status.outputs.script_name` |
| `spec.vpcNetworks[].tunnelId` | CloudflareZeroTrustTunnel | `status.outputs.tunnel_id` |
| `spec.tailConsumerBindings[].service` | CloudflareWorker | `status.outputs.script_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareEmailRoutingRule | `spec.actions[].worker` | `status.outputs.script_name` |
| CloudflareEmailRoutingZone | `spec.catchAll.actions[].worker` | `status.outputs.script_name` |
| CloudflarePagesProject | `spec.deploymentConfigs.preview.services[].service` | `status.outputs.script_name` |
| CloudflarePagesProject | `spec.deploymentConfigs.production.services[].service` | `status.outputs.script_name` |
| CloudflareQueue | `spec.consumer.scriptName` | `status.outputs.script_name` |
| CloudflareWorker | `spec.services[].service` | `status.outputs.script_name` |
| CloudflareWorker | `spec.durableObjects[].scriptName` | `status.outputs.script_name` |
| CloudflareWorker | `spec.tailConsumers[].service` | `status.outputs.script_name` |
| CloudflareWorker | `spec.migrations.transferredClasses[].fromScript` | `status.outputs.script_name` |
| CloudflareWorker | `spec.migrations.steps[].transferredClasses[].fromScript` | `status.outputs.script_name` |
| CloudflareWorker | `spec.dispatchNamespaces[].outbound.worker.service` | `status.outputs.script_name` |
| CloudflareWorker | `spec.tailConsumerBindings[].service` | `status.outputs.script_name` |
| CloudflareWorkflow | `spec.scriptName` | `status.outputs.script_name` |

## See Also

- [Overview](../README.md)
