# Alibaba Cloud SLS Log Project Deployment: From Console Clicks to Control Planes

## Introduction

Alibaba Cloud Simple Log Service (SLS) is the centralized logging backbone for virtually every workload running on Alibaba Cloud. An SLS project is the top-level container that groups related log stores — the individual storage and query units — into a single namespace scoped to a region. Without a project, there is no log ingestion; without log stores, there is no data; without indexes, the data is unsearchable. These three resources form an inseparable triad.

Despite this simplicity, provisioning SLS correctly in production is deceptively error-prone. Teams routinely create projects with no stores (a shell that collects nothing), stores with no indexes (data that cannot be queried), or stores with retention periods and shard counts that don't match workload volume — leading to either data loss or runaway costs. The problem isn't complexity; it's fragmentation. The console, CLI, and raw IaC tools all treat the project, store, and index as independent resources, forcing operators to wire them together manually every time.

This document examines the full deployment landscape for SLS projects — from manual console provisioning to control-plane-based automation — and explains how OpenMCF bundles the project-store-index triad into a single, validated API resource that eliminates the most common misconfigurations while remaining flexible enough for production use.

## Evolution of SLS Deployment

Alibaba Cloud SLS has evolved significantly since its launch. Early adoption relied entirely on the Alibaba Cloud console, where creating a project, adding stores, and enabling indexes were three separate workflows spread across multiple pages. The introduction of the `aliyun` CLI brought scriptability but not idempotency — teams wrote shell scripts that broke on re-runs. Terraform's `alicloud` provider (via the `alicloud_log_project`, `alicloud_log_store`, and `alicloud_log_store_index` resources) brought declarative state management but still required users to wire three separate resource blocks together. Pulumi's Go SDK (`pulumi-alicloud`) offers the same granularity with the added benefit of type-safe composition.

The pattern across this evolution is consistent: every tool treats the project, store, and index as independent resources. This is architecturally correct (they are separate API objects in the SLS API) but operationally burdensome. The most common "production-ready" SLS setup — a project with 2-5 log stores, each with a full-text index — requires 7-16 resource declarations in Terraform or equivalent Pulumi calls. OpenMCF's contribution is collapsing this into a single resource declaration.

## The SLS Deployment Landscape

### Level 0: Manual Provisioning via Alibaba Cloud Console

The Alibaba Cloud console provides a wizard-driven workflow for SLS:

1. Navigate to **Log Service** in the console
2. Click **Create Project**, select a region, enter a name
3. After project creation, click **Create Logstore**, configure retention and shards
4. After store creation, navigate to **Index/Search**, enable full-text indexing

**Common Mistakes**:

1. **Empty Projects**: Creating a project without immediately adding log stores. The project exists but collects nothing. Teams often forget to come back and add stores, especially when the project was created "for later."

2. **Missing Indexes**: Creating log stores but not enabling full-text indexing. Data flows into the store but is completely unsearchable — the SLS query interface returns zero results. This is the most frustrating mistake because the console shows the store as "healthy" while the data is effectively invisible.

3. **Incorrect Retention**: Accepting the default retention period without considering compliance or cost requirements. Storing 365 days of debug logs costs 12x what 30-day retention costs, while storing only 7 days of audit logs may violate compliance policies.

4. **Under-Provisioned Shards**: Starting with 1 shard for a high-throughput workload. Each shard supports approximately 5 MB/s write throughput. A Kubernetes cluster generating 50 MB/s of logs needs at least 10 shards, but the console defaults to 2.

5. **No Auto-Split**: Leaving auto-split disabled. Without auto-split, a traffic spike that exceeds shard capacity silently drops log data. The SLS API returns throttling errors, but unless the log shipper (Logtail or SDK) is configured to retry and alert, the data loss goes unnoticed.

**Verdict**: Acceptable for learning and ad-hoc exploration. Unacceptable for production environments where reproducibility, consistency, and data integrity matter.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun` CLI provides direct access to the SLS API:

```bash
# Create project
aliyun sls CreateProject --body '{
  "projectName": "my-app-logs",
  "description": "Application logging"
}'

# Create log store
aliyun sls CreateLogStore \
  --project my-app-logs \
  --body '{
    "logstoreName": "app-logs",
    "ttl": 30,
    "shardCount": 2,
    "autoSplit": true,
    "maxSplitShard": 64
  }'

# Create full-text index
aliyun sls CreateIndex \
  --project my-app-logs \
  --logstore app-logs \
  --body '{
    "line": {
      "caseSensitive": false,
      "token": [",", " ", "\"", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t"]
    }
  }'
```

**The Sequence Problem**: Three separate API calls must execute in order (project before store, store before index). Each call is imperative — there is no built-in idempotency. Running the script twice fails on the second run because the project already exists. Teams wrap each call in `if-not-exists` checks, tripling the script length and introducing race conditions.

**The State Problem**: The CLI has no state file. There is no way to know whether the current SLS configuration matches the desired configuration without querying the API and comparing field-by-field. Drift detection is manual.

**Verdict**: Suitable for one-off tasks, CI/CD pipeline steps where idempotency is handled externally, or debugging. Not suitable for managing SLS infrastructure at scale.

### Level 2: Infrastructure as Code (Terraform / OpenTofu)

Terraform's `alicloud` provider breaks SLS into three granular resources:

```hcl
resource "alicloud_log_project" "main" {
  project_name = "my-app-logs"
  description  = "Application logging"
}

resource "alicloud_log_store" "app_logs" {
  project_name        = alicloud_log_project.main.project_name
  logstore_name       = "app-logs"
  retention_period    = 30
  shard_count         = 2
  auto_split          = true
  max_split_shard_count = 64
  append_meta         = true
}

resource "alicloud_log_store_index" "app_logs" {
  project  = alicloud_log_project.main.project_name
  logstore = alicloud_log_store.app_logs.logstore_name

  full_text {
    case_sensitive  = false
    include_chinese = false
    token           = ", '\"=()[]{}?@&<>/:\n\t\r"
  }
}
```

**Strengths**:

- **Declarative**: Define the desired end state; Terraform calculates the diff
- **Stateful**: Tracks resource IDs and attributes in a state file; detects drift
- **Dependency Graph**: Automatically creates project before store, store before index
- **Idempotent**: Running `terraform apply` twice produces the same result

**Weaknesses**:

- **Verbose**: A project with 3 stores and indexes requires 7 resource blocks (1 project + 3 stores + 3 indexes). With 5 stores, that's 11 blocks. The `for_each` meta-argument helps but adds its own complexity.
- **No Bundling Semantics**: Terraform has no concept of "a project always comes with stores and indexes." Each resource is independent, so nothing prevents creating a project with no stores.
- **State Management Overhead**: The state file must be stored remotely (OSS bucket + TableStore for locking) in team environments. This is operational overhead unrelated to the SLS domain.

**The `for_each` Pattern**: Production Terraform modules use `for_each` to iterate over a list of log store configurations:

```hcl
variable "log_stores" {
  type = list(object({
    name            = string
    retention_days  = optional(number, 30)
    shard_count     = optional(number, 2)
    enable_index    = optional(bool, true)
  }))
}

resource "alicloud_log_store" "stores" {
  for_each         = { for s in var.log_stores : s.name => s }
  project_name     = alicloud_log_project.main.project_name
  logstore_name    = each.value.name
  retention_period = each.value.retention_days
  shard_count      = each.value.shard_count
}

resource "alicloud_log_store_index" "indexes" {
  for_each = { for s in var.log_stores : s.name => s if s.enable_index }
  project  = alicloud_log_project.main.project_name
  logstore = alicloud_log_store.stores[each.key].logstore_name

  full_text {
    case_sensitive = false
    token          = ", '\"=()[]{}?@&<>/:\n\t\r"
  }
}
```

This is exactly the pattern OpenMCF's Terraform module uses — but wrapped behind a validated API that ensures the `for_each` logic is correct and the index configuration is consistent.

**Verdict**: The modern standard for managing SLS infrastructure. Recommended for teams already using Terraform. The granularity is a trade-off: maximum flexibility at the cost of verbosity.

### Level 3: Infrastructure as Code (Pulumi)

Pulumi's Go SDK provides type-safe SLS resource creation:

```go
project, err := log.NewProject(ctx, "my-app-logs", &log.ProjectArgs{
    ProjectName: pulumi.String("my-app-logs"),
    Description: pulumi.String("Application logging"),
})

store, err := log.NewStore(ctx, "app-logs", &log.StoreArgs{
    ProjectName:        pulumi.String("my-app-logs"),
    LogstoreName:       pulumi.String("app-logs"),
    RetentionPeriod:    pulumi.Int(30),
    ShardCount:         pulumi.Int(2),
    AutoSplit:          pulumi.Bool(true),
    MaxSplitShardCount: pulumi.Int(64),
    AppendMeta:         pulumi.Bool(true),
}, pulumi.Parent(project))

_, err = log.NewStoreIndex(ctx, "app-logs-index", &log.StoreIndexArgs{
    Project:  pulumi.String("my-app-logs"),
    Logstore: pulumi.String("app-logs"),
    FullText: &log.StoreIndexFullTextArgs{
        CaseSensitive:  pulumi.Bool(false),
        IncludeChinese: pulumi.Bool(false),
        Token:          pulumi.String(`, '";=()[]{}?@&<>/:\n\t\r`),
    },
}, pulumi.Parent(store))
```

**Key Advantages Over Terraform**:

- **Type Safety**: Compile-time validation of field names and types. Misspelling `RetentionPeriod` is a build error, not a runtime surprise.
- **Parent Chaining**: `pulumi.Parent(project)` and `pulumi.Parent(store)` create explicit dependency hierarchies that are visible in the Pulumi state and UI.
- **Programmatic Composition**: Loops, conditionals, and functions are native Go — no HCL `for_each` or `count` workarounds.
- **Multi-Language**: Same logic can be expressed in TypeScript, Python, Java, or C# for teams that prefer those languages.

**Key Disadvantage**: Requires compiling Go code (or running a Node/Python runtime). Terraform's declarative HCL is simpler for teams that don't need programmatic composition.

**Verdict**: Preferred for teams using Go or TypeScript, especially when SLS provisioning is embedded in a larger orchestration workflow. The type safety eliminates an entire class of configuration errors.

### Level 4: Control Planes and Continuous Reconciliation

The most advanced deployment model treats SLS configuration as a continuously reconciled desired state:

- **Crossplane**: Extends the Kubernetes API with custom resources for Alibaba Cloud. An operator watches for `AlicloudLogProject` custom resources and provisions/reconciles the SLS infrastructure automatically.
- **Custom Operators**: Teams build Kubernetes operators that watch for application deployments and automatically create corresponding SLS projects and stores.

**OpenMCF Context**: OpenMCF's protobuf-defined API is designed for this model. The YAML manifest is a desired-state declaration that can be applied once (CLI mode) or continuously reconciled (control-plane mode). The `AlicloudLogProject` resource is a Kubernetes-native API object, not just a CLI input format.

**Verdict**: The future of infrastructure management. OpenMCF's API design anticipates this model even when used in CLI mode today.

## Comparative Analysis

| Method | Idempotent | State Tracked | Bundled | Validated | Drift Detection | Effort to Add 5 Stores |
|--------|-----------|--------------|---------|-----------|----------------|----------------------|
| Console | No | No | No | No | No | ~15 minutes clicking |
| CLI (`aliyun`) | No | No | No | No | No | ~50 lines of bash |
| Terraform | Yes | Yes | No | Partial | Yes | 11 resource blocks |
| Pulumi | Yes | Yes | No | Compile-time | Yes | ~60 lines of Go |
| OpenMCF | Yes | Yes | Yes | Proto-validated | Yes | 5 list items in YAML |

The key differentiator is the **Bundled** column. Every other method treats the project, store, and index as independent resources. OpenMCF is the only approach that bundles them into a single validated declaration, ensuring that stores always have indexes and that the entire triad is provisioned atomically.

## The OpenMCF Approach

### Design Philosophy: The DD07 Bundling Decision

The most important design decision for AlicloudLogProject is **DD07: composite bundling**. Instead of creating three separate OpenMCF resources (`AlicloudLogProject`, `AlicloudLogStore`, `AlicloudLogStoreIndex`), the component bundles all three into a single resource.

**Why bundle?**

1. **An empty project is useless**: An SLS project without log stores collects no data. There is no use case where a user creates a project and intentionally leaves it empty. Unbundling would create an API that encourages a broken state.

2. **A store without an index is a trap**: Data flows into an unindexed store normally — the SLS API accepts writes. But the query interface returns nothing. This silent failure is the most common SLS misconfiguration. Bundling the index with the store (controlled by the `enableIndex` flag, defaulting to `true`) eliminates this trap.

3. **The 80% use case is simple**: Most teams need "a project with N stores, each searchable." The bundled API serves this case with a flat list of store configurations. The 20% case (stores with custom field-level indexes, cross-store queries, machine learning jobs) is not addressed — and intentionally so.

### 80/20 Scoping: What's In and What's Out

**Included (the 80%)**:

- Project creation with region, name, description, resource group, and tags
- Multiple log stores per project with configurable retention, shards, and auto-split
- Full-text indexing per store with sensible defaults (case-insensitive, standard tokenization)
- Metadata enrichment per store (`appendMeta` adds receive time and client IP)

**Excluded (the 20%)**:

- **Logtail configuration**: How logs get into the store (agent configuration, machine groups, log collection configs) is a separate operational concern. Different teams use Logtail, Fluent Bit, or the SLS SDK. Bundling log collection with project creation would create an overly opinionated resource.
- **Field-level indexes**: Full-text indexing covers the majority of query patterns. Field-level indexes (for structured queries like `status:200 AND latency>500`) require schema knowledge that varies by application. This is a customization best handled after initial provisioning.
- **Dashboards and alerts**: Monitoring configuration belongs to a separate lifecycle. Dashboard YAML changes far more frequently than infrastructure topology.
- **Cross-region replication**: An advanced feature that involves two projects in different regions. This is better modeled as a relationship between two `AlicloudLogProject` resources, not as a field within one.
- **Log ETL (data transformation)**: ETL jobs that transform data between stores are complex pipelines with their own lifecycle. Bundling them would make the resource unwieldy.

### API Design Decisions

**`projectName` vs `name`**: The spec uses `projectName` (not `name`) because the SLS API requires a globally unique project name that is distinct from the OpenMCF `metadata.name`. The metadata name is the local resource identifier; the project name is the Alibaba Cloud resource identifier.

**`logStores` as a repeated message**: Each store is a structured message with its own fields rather than a simple string list. This allows per-store configuration (different retention for app-logs vs. audit-logs) while keeping the API flat and readable.

**Optional fields with proto defaults**: Every optional field uses the `(org.openmcf.shared.options.default)` annotation to document its default value. The IaC modules read these defaults from the proto definition, ensuring consistency between the API contract and the implementation.

**`enableIndex` defaults to `true`**: This is an opinionated default. In raw SLS, indexes are not created automatically. OpenMCF inverts this because the primary value of SLS is querying logs — storing unsearchable data is almost never the intent.

### Foreign Key References

AlicloudLogProject has no `StringValueOrRef` fields — all fields are direct values. This is correct because SLS projects are foundation resources with no upstream dependencies. Downstream resources (AckManagedCluster, FcFunction, SaeApplication) reference this project's outputs via their own `StringValueOrRef` fields.

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module consists of three files under `v1/iac/pulumi/module/`:

**`main.go`** — The controller. Entry point is `Resources(ctx, stackInput)`.
1. Initializes locals (tag computation, default resolution)
2. Creates the Alibaba Cloud provider with the specified region
3. Creates the `log.Project` resource
4. Iterates over `spec.LogStores`, creating a `log.Store` for each (parented to the project)
5. For each store with `enableIndex == true`, creates a `log.StoreIndex` (parented to the store)
6. Exports outputs: project name, project ID, log store names map

**`locals.go`** — Transformations and defaults.
- Computes the tag map by merging standard tags (`resource`, `resource_name`, `resource_kind`, `resource_id`, `organization`, `environment`) with user-provided `spec.Tags`
- Provides helper functions for resolving optional fields: `logStoreRetentionDays()`, `logStoreShardCount()`, `logStoreAutoSplit()`, `logStoreMaxSplitShardCount()`, `logStoreEnableIndex()`, `logStoreAppendMeta()` — each returns the proto value if set, otherwise the documented default

**`outputs.go`** — Output constant definitions.
- `OpProjectName = "project_name"`
- `OpProjectId = "project_id"`
- `OpLogStoreNames = "log_store_names"`

**Resource Hierarchy**:

```
Provider (alicloud, region-scoped)
  └── log.Project
        ├── log.Store ("project-name-store-name")
        │     └── log.StoreIndex ("project-name-store-name-index")
        ├── log.Store (...)
        │     └── log.StoreIndex (...)
        └── ...
```

The `pulumi.Parent()` chaining ensures that deleting the project cascades to stores and indexes, and that the Pulumi dependency graph reflects the actual SLS API constraints.

### Terraform Module Architecture

The Terraform module consists of five files under `v1/iac/tf/`:

**`main.tf`** — Resource definitions.
- `alicloud_log_project.main`: Single project resource
- `alicloud_log_store.stores`: Uses `for_each` over `local.log_stores_map`
- `alicloud_log_store_index.indexes`: Uses `for_each` over `local.log_stores_with_index` (filtered to stores where `enable_index == true`)

**`variables.tf`** — Input variables matching the proto schema.
- `metadata` object with `name`, `id`, `org`, `env`, `labels`, `tags`
- `spec` object mirroring `AlicloudLogProjectSpec` with all fields and defaults

**`locals.tf`** — Computed values.
- `log_stores_map`: Converts the list of stores to a map keyed by name (required for `for_each`)
- `log_stores_with_index`: Filters `log_stores_map` to entries where `enable_index == true`
- `final_tags`: Merges standard tags with user tags (same logic as Pulumi's `locals.go`)

**`outputs.tf`** — Three outputs matching the Pulumi module.
- `project_name`, `project_id`, `log_store_names` (map comprehension)

**`provider.tf`** — Alibaba Cloud provider configuration with region from `var.spec.region`.

The Terraform and Pulumi modules create identical resources with identical outputs, ensuring that switching between IaC engines produces the same infrastructure.

## Production Best Practices

### Retention Policy Design

SLS charges by data volume stored. Retention is the single largest cost lever.

| Use Case | Recommended Retention | Rationale |
|----------|----------------------|-----------|
| Debug / Development | 7 days | Short-lived; cost-sensitive |
| Application logs | 30-90 days | Covers most incident investigation windows |
| Audit logs | 365 days | Compliance requirements (SOC 2, ISO 27001) |
| Security logs | 180-365 days | Forensic investigation needs |
| Permanent archive | 3650 days | Regulatory hold (set to max) |

**Anti-Pattern**: Using a single retention period for all stores. Application debug logs and compliance audit logs have fundamentally different lifecycle requirements. The `logStores` repeated field in the spec exists specifically to allow per-store retention configuration.

### Shard Sizing and Auto-Split

Each shard supports approximately 5 MB/s write throughput and 10 MB/s read throughput.

**Sizing Formula**: `shardCount = ceil(peakWriteThroughputMB / 5)`

| Workload Size | Peak Write Rate | Recommended Shards | Auto-Split |
|--------------|-----------------|-------------------|------------|
| Small (dev/staging) | < 5 MB/s | 1-2 | Enabled |
| Medium (single app) | 5-25 MB/s | 2-5 | Enabled |
| Large (platform) | 25-100 MB/s | 5-20 | Enabled |
| Very Large (multi-tenant) | > 100 MB/s | 20+ | Enabled, maxSplitShardCount = 256 |

**Why Auto-Split Should Almost Always Be Enabled**: Without auto-split, a traffic spike that exceeds shard capacity results in `403 ShardWriteQuotaExceed` errors. The SLS SDK and Logtail handle retries, but sustained throttling causes data loss. Auto-split transparently adds shards to absorb the spike. The `maxSplitShardCount` cap prevents runaway shard creation (each shard consumes read quota and increases query fan-out).

**Anti-Pattern**: Setting `maxSplitShardCount` equal to `shardCount`. This effectively disables auto-split. Always set `maxSplitShardCount` to at least 4x the initial `shardCount`.

### Index Configuration

**Full-Text Index Defaults**: The OpenMCF module creates a full-text index with:
- Case-insensitive matching (`caseSensitive: false`)
- Chinese text support disabled (`includeChinese: false`)
- Standard tokenization: `, '";=()[]{}?@&<>/:\n\t\r`

These defaults cover the majority of structured log formats (JSON, key=value pairs, plain text). The standard tokenizer splits on punctuation and whitespace, enabling queries like `error`, `status:500`, and `request_id:abc123`.

**When Full-Text Isn't Enough**: Applications with deeply nested JSON logs or high-cardinality fields may benefit from field-level indexes. These are not supported by the current OpenMCF API (see 80/20 scoping above) and should be configured directly via the SLS console or Terraform after initial provisioning.

### Tag Strategy

The Pulumi and Terraform modules automatically apply standard tags:

| Tag Key | Source | Purpose |
|---------|--------|---------|
| `resource` | `"true"` | Identifies OpenMCF-managed resources |
| `resource_name` | `metadata.name` | Links back to the manifest |
| `resource_kind` | `"alicloudlogproject"` | Resource type for filtering |
| `resource_id` | `metadata.id` | Unique resource instance ID |
| `organization` | `metadata.org` | Organizational grouping |
| `environment` | `metadata.env` | Environment isolation |

User-provided tags from `spec.tags` are merged with these standard tags. User tags take precedence on conflict.

### Security Considerations

- **No credentials in the manifest**: Alibaba Cloud credentials are injected via environment variables (`ALIBABA_CLOUD_ACCESS_KEY_ID`, `ALIBABA_CLOUD_ACCESS_KEY_SECRET`) by the runner. The manifest `spec` never contains secrets.
- **Resource Group isolation**: The optional `resourceGroupId` field allows placing the SLS project in a specific Alibaba Cloud resource group, enabling RAM policy isolation (e.g., "team A can only access projects in resource group rg-team-a").
- **Project naming**: SLS project names are globally unique within Alibaba Cloud. Use a naming convention that includes the organization and environment (e.g., `myorg-prod-platform-logs`) to prevent collisions and enable audit trail.

### Cost Optimization

SLS pricing has three dimensions:

1. **Ingestion**: Charged per GB of raw log data written
2. **Storage**: Charged per GB-day of stored data (retention × volume)
3. **Index Storage**: Full-text indexes consume approximately 0.7x the raw data volume

**Key Optimization Levers**:

- **Retention differentiation**: Use short retention for high-volume, low-value logs (access logs, debug logs) and long retention for low-volume, high-value logs (audit logs, security events)
- **Selective indexing**: Set `enableIndex: false` for stores used purely for archival (compliance retention without query needs). This eliminates index storage costs entirely.
- **Shard management**: Over-provisioning shards doesn't increase ingestion or storage costs, but increases read fan-out during queries. Right-size initial shards and rely on auto-split for spikes.

### Cross-Region Considerations

SLS projects are region-scoped. A project in `cn-hangzhou` can only receive logs from Logtail agents or SDK clients configured to send to the `cn-hangzhou` endpoint.

**Multi-Region Strategy**: Deploy one `AlicloudLogProject` resource per region where workloads run. Use SLS's built-in cross-region replication (not managed by this OpenMCF component) if centralized querying is needed.

**Endpoint Selection**: Choose the region closest to the log source to minimize ingestion latency. For Kubernetes clusters, this means the SLS project should be in the same region as the ACK cluster.

## Common Anti-Patterns

| Anti-Pattern | Consequence | OpenMCF Mitigation |
|-------------|-------------|-------------------|
| Project with no stores | Empty shell, no data collection | `logStores` field encourages bundling at creation |
| Store with no index | Data ingested but unsearchable | `enableIndex` defaults to `true` |
| Single retention for all stores | Overpaying for debug logs or losing audit logs | Per-store `retentionDays` field |
| Auto-split disabled | Data loss during traffic spikes | `autoSplit` defaults to `true` |
| maxSplitShardCount = shardCount | Auto-split effectively disabled | `maxSplitShardCount` defaults to 64 |
| Hardcoded credentials in manifest | Security vulnerability, audit failure | Credentials via env vars, never in spec |
| Non-unique project names | Creation failure (globally unique) | Validated with min/max length constraints |

## Conclusion

SLS project deployment is a solved problem at every level of the tooling spectrum — from console to control plane. What OpenMCF adds is not a new deployment mechanism but a **bundled, validated abstraction** that eliminates the most common misconfigurations:

- Projects are created with stores (no empty shells)
- Stores are created with indexes (no unsearchable data)
- Defaults are production-ready (auto-split enabled, 30-day retention, 2 shards)
- Tags are standardized (organization, environment, resource kind)
- Credentials are externalized (never in the manifest)

The DD07 bundling decision is the architectural cornerstone: by treating project + stores + indexes as a single resource, OpenMCF makes the common case simple (one YAML resource for a complete logging setup) while leaving the advanced case possible (custom field-level indexes and ETL can be layered on top).

For teams adopting Alibaba Cloud, `AlicloudLogProject` is typically the first resource deployed — it provides the logging foundation that AckManagedCluster, FcFunction, and SaeApplication reference for operational visibility.

### References

- [Alibaba Cloud SLS Product Overview](https://www.alibabacloud.com/help/en/sls/product-overview/)
- [SLS Log Store Concepts](https://www.alibabacloud.com/help/en/sls/user-guide/logstore/)
- [SLS Indexing and Query](https://www.alibabacloud.com/help/en/sls/user-guide/enable-and-configure-the-indexing-feature-for-a-logstore/)
- [SLS Shard Management](https://www.alibabacloud.com/help/en/sls/user-guide/manage-shards/)
- [Terraform alicloud_log_project Resource](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/log_project)
- [Pulumi Alibaba Cloud Log Package](https://www.pulumi.com/registry/packages/alicloud/api-docs/log/)
