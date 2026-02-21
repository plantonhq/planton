# Pulumi Module Overview — AliCloudLogProject

## Module Architecture

The module lives in `module/` and consists of three files:

| File | Responsibility |
|------|---------------|
| `main.go` | Controller — creates provider, project, stores, and indexes |
| `locals.go` | Transformations — tag merging, optional-field default resolution |
| `outputs.go` | Output constants — `project_name`, `project_id`, `log_store_names` |

Entry point: `Resources(ctx *pulumi.Context, stackInput *AliCloudLogProjectStackInput) error`

## Control Flow

```
StackInput (target manifest + provider config)
  │
  ▼
initializeLocals()
  ├── merge standard tags (resource_name, resource_kind, organization, environment)
  └── merge user-provided spec.Tags (user tags win on conflict)
  │
  ▼
alicloud.NewProvider(region)
  │
  ▼
log.NewProject(projectName, description, resourceGroupId, tags)
  │
  ▼
for each spec.LogStores:
  ├── log.NewStore(name, retention, shards, autoSplit, maxSplit, appendMeta)
  │     parent: Project
  │
  └── if enableIndex == true:
        log.NewStoreIndex(fullText: caseSensitive=false, standard tokenizer)
              parent: Store
  │
  ▼
ctx.Export(project_name, project_id, log_store_names)
```

## Resource Hierarchy

Resources use `pulumi.Parent()` to form a dependency tree:

```
Provider
  └── log.Project
        ├── log.Store "projectname-storename"
        │     └── log.StoreIndex "projectname-storename-index"
        ├── log.Store "projectname-anothername"
        │     └── log.StoreIndex "projectname-anothername-index"
        └── ...
```

Parent chaining ensures:
- Stores are created after the project and destroyed before it
- Indexes are created after their store and destroyed before it
- The Pulumi state tree mirrors the SLS resource hierarchy

## Default Resolution

Optional proto fields use `optional` syntax with `(org.openmcf.shared.options.default)`
annotations. Because proto3 optional fields are nil when unset, `locals.go`
provides helper functions that return the documented default:

| Function | Default | Proto Annotation |
|----------|---------|-----------------|
| `logStoreRetentionDays(ls)` | 30 | `"30"` |
| `logStoreShardCount(ls)` | 2 | `"2"` |
| `logStoreAutoSplit(ls)` | true | `"true"` |
| `logStoreMaxSplitShardCount(ls)` | 64 | `"64"` |
| `logStoreEnableIndex(ls)` | true | `"true"` |
| `logStoreAppendMeta(ls)` | true | `"true"` |

This pattern keeps defaults in sync between the proto contract and the Go
implementation. If a default changes in the proto, the corresponding helper
must be updated.

## Design Decisions

**Bundled resources (DD07)**: Project, stores, and indexes are provisioned in a
single `Resources()` call rather than as separate OpenMCF components. A project
without stores is an empty shell; a store without an index is unsearchable. The
bundled approach ensures the triad is always complete.

**Single entry point**: No separate resource files (e.g., `project.go`,
`store.go`). The module is small enough that splitting would add navigation cost
without reducing complexity. If store or index logic grows (e.g., adding
field-level indexes), extract to dedicated files.

**`optionalString` helper**: Returns `nil` for empty strings, preventing the
Pulumi provider from sending empty string values to the SLS API (which would
override server-side defaults).

**Index tokenizer**: Hardcoded to the SLS standard tokenizer rather than exposed
as a spec field. The standard tokenizer covers structured log formats (JSON,
key=value) and is the correct default for >90% of use cases. Field-level index
customization is a 20% feature deferred from the current API.

## Customization Guide

| Goal | File to Modify | Notes |
|------|---------------|-------|
| Add a new log store field (e.g., `hotTtl`) | `locals.go` (add default helper), `main.go` (pass to `log.StoreArgs`) | Also update `spec.proto` and regenerate |
| Add Logtail config | New file `logtail.go` | Would need `AliCloudLogtailConfig` message in `spec.proto` |
| Change index tokenizer | `main.go`, `logStoreIndex()` function | Consider exposing as a spec field if multiple tokenizers are needed |
| Add field-level indexes | `main.go`, new `log.StoreIndex` with `FieldSearch` args | Requires `AliCloudLogStoreFieldIndex` message in `spec.proto` |
| Add project-level encryption | `main.go`, `log.ProjectArgs` | SLS supports CMK encryption; add `kmsKeyId` to spec |
