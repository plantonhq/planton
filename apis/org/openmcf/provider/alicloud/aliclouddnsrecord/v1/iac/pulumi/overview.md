# Overview

The AlicloudDnsRecord Pulumi module creates a single Alibaba Cloud DNS record from an OpenMCF manifest. The module is intentionally minimal -- three files, one cloud resource -- because DNS records are atomic building blocks. Multiple records within a domain are managed as separate OpenMCF resources.

## Module Architecture

```
iac/pulumi/
├── main.go              Entry point: loads stack input, calls module.Resources()
└── module/
    ├── locals.go         Stores the target resource reference (no tag computation)
    ├── main.go           Controller: creates the Alicloud provider and the DNS record
    └── outputs.go        Defines output constant names (record_id)
```

**Entry point** (`iac/pulumi/main.go`): Deserializes the Pulumi config into `AlicloudDnsRecordStackInput` via `stackinput.LoadStackInput()`, then delegates to `module.Resources()`.

**Controller** (`module/main.go`): Initializes locals, creates the Alicloud provider scoped to `spec.region`, provisions the DNS record via `dns.NewAlidnsRecord`, and exports one output.

**Locals** (`module/locals.go`): Minimal -- stores a reference to the target resource. No tag computation because `alicloud_alidns_record` does not support tags.

**Outputs** (`module/outputs.go`): Single constant for the record_id export name, ensuring consistency between the Pulumi export and `stack_outputs.proto`.

## Data Flow

```
AlicloudDnsRecordStackInput
  │
  └─ target.Spec ──► Resources()
                        │
                        ├─ alicloud.NewProvider (region)
                        │
                        └─ dns.NewAlidnsRecord
                              │
                              └─► Export: record_id
```

## Design Decisions

**No tags**: Unlike AlicloudDnsZone, the DNS record resource does not support tags. The Locals struct and initializeLocals function are intentionally minimal.

**Conditional optional fields**: Optional spec fields (ttl, priority, line, status, remark) are only passed to the Pulumi SDK when non-zero/non-empty. This prevents sending default proto values (0 for int32, "" for string) to the API, which would override provider defaults.

**Resource naming**: The Pulumi resource name uses `{rr}.{domain_name}` (e.g., `www.example.com`) for human-readable identification in Pulumi state.

**Global service, regional provider**: Alidns is a global service -- records are not region-scoped. The `spec.region` field is used solely for provider initialization, consistent with AlicloudDnsZone.

## Customization

| Goal | File to Change |
|------|---------------|
| Add spec fields to the record resource | `module/main.go` -- add args to `dns.AlidnsRecordArgs` |
| Add new stack outputs | `module/outputs.go` (constant) + `module/main.go` (export call) |
