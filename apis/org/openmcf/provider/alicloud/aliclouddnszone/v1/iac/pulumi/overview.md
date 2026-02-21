# Overview

The AlicloudDnsZone Pulumi module creates a single Alibaba Cloud DNS domain from an OpenMCF manifest. The module is intentionally minimal -- three files, one cloud resource -- because the DNS domain is a composable building block. DNS records within the domain are managed by the separate AlicloudDnsRecord component.

## Module Architecture

```
iac/pulumi/
├── main.go              Entry point: loads stack input, calls module.Resources()
└── module/
    ├── locals.go         Transforms stack input into computed values (tag merging)
    ├── main.go           Controller: creates the Alicloud provider and the DNS domain resource
    └── outputs.go        Defines output constant names (domain_id, domain_name, etc.)
```

**Entry point** (`iac/pulumi/main.go`): Deserializes the Pulumi config into `AlicloudDnsZoneStackInput` via `stackinput.LoadStackInput()`, then delegates to `module.Resources()`.

**Controller** (`module/main.go`): Initializes locals, creates the Alicloud provider scoped to `spec.region`, provisions the DNS domain via `dns.NewAlidnsDomain`, and exports five outputs.

**Locals** (`module/locals.go`): Builds the merged tag map. System tags (`resource`, `resource_name`, `resource_kind`) are set first. Metadata fields (`resource_id`, `organization`, `environment`) are added conditionally. User-defined `spec.tags` are merged last, so user values override system tags on key conflict.

**Outputs** (`module/outputs.go`): String constants for the five export names, ensuring consistency between the Pulumi exports and `stack_outputs.proto`.

## Data Flow

```
AlicloudDnsZoneStackInput
  │
  ├─ target.Metadata  ──► initializeLocals() ──► Locals.Tags (merged map)
  ├─ target.Spec.Tags ─┘
  │
  └─ target.Spec ──► Resources()
                        │
                        ├─ alicloud.NewProvider (region)
                        │
                        └─ dns.NewAlidnsDomain
                              │
                              ├─► Export: domain_id
                              ├─► Export: domain_name
                              ├─► Export: dns_servers
                              ├─► Export: group_name
                              └─► Export: puny_code
```

## Design Decisions

**Single resource scope**: The DNS domain component creates only the domain registration in Alidns. DNS records are a separate component (AlicloudDnsRecord) because they have independent lifecycles and many-to-one cardinality.

**`optionalString()` helper**: Empty proto strings are converted to `nil` before passing to the Pulumi SDK. This prevents Alibaba Cloud API errors that occur when empty strings are sent for optional fields like `group_id`, `remark`, and `resource_group_id`.

**Tag merge order**: System tags are written first, then user tags overwrite. This gives users the final say on all tag values while ensuring that every domain gets baseline metadata tags for governance and tracking.

**Global service, regional provider**: Alidns is a global service -- domain registration is not scoped to a specific region. However, the Alibaba Cloud provider requires a region for API endpoint resolution, so `spec.region` is included for provider initialization consistency.

## Customization

| Goal | File to Change |
|------|---------------|
| Add spec fields to the domain resource | `module/main.go` -- add args to `dns.AlidnsDomainArgs` |
| Change tag logic or add new system tags | `module/locals.go` -- modify `initializeLocals()` |
| Add new stack outputs | `module/outputs.go` (constant) + `module/main.go` (export call) |
