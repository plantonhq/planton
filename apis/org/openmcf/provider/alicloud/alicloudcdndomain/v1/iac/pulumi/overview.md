# AliCloudCdnDomain — Pulumi Module Overview

## Module Architecture

```
iac/pulumi/
├── Pulumi.yaml          # Project config (Go runtime)
├── main.go              # Entrypoint: loads stack input, calls module
├── module/
│   ├── locals.go        # Tag computation from metadata + user tags
│   ├── main.go          # Provider creation and CDN domain resource
│   └── outputs.go       # Output key constants
├── README.md            # CLI usage and debugging
├── examples.md          # Deployment examples
└── overview.md          # This file
```

## Key Components

### Entrypoint (`main.go`)

The entrypoint initializes the Pulumi runtime, loads the stack input from the
OpenMCF manifest via `stackinput.LoadStackInput`, and delegates resource
creation to `module.Resources()`. The stack input is deserialized into the
protobuf-generated `AliCloudCdnDomainStackInput` struct.

### Locals (`module/locals.go`)

The `initializeLocals` function computes a `Locals` struct containing:

- A reference to the deserialized `AliCloudCdnDomain` manifest.
- A merged tag map combining system tags (`resource`, `resource_name`,
  `resource_kind`, `resource_id`, `organization`, `environment`) with
  user-defined `spec.tags`. User tags override system tags on key collision.

### Resources (`module/main.go`)

The `Resources` function performs three actions:

1. Creates an explicit alicloud provider scoped to `spec.region`.
2. Builds a `cdn.DomainNewArgs` struct from the spec fields, conditionally
   including optional fields (`scope`, `checkUrl`, `resourceGroupId`,
   `certificateConfig`) only when they have non-zero values.
3. Creates a `cdn.DomainNew` resource and exports three outputs.

Helper functions:

- `buildSources` — converts the protobuf `AliCloudCdnDomainSource` slice to
  a `cdn.DomainNewSourceArray`, including port/priority/weight only when
  explicitly set (non-zero).
- `buildCertificateConfig` — converts the protobuf
  `AliCloudCdnDomainCertificateConfig` to a `cdn.DomainNewCertificateConfigPtrInput`,
  including each field only when non-empty.

### Outputs (`module/outputs.go`)

Three string constants define the output keys:

| Constant | Value | Description |
|----------|-------|-------------|
| `OpDomainName` | `domain_name` | The registered accelerated domain name |
| `OpCname` | `cname` | The CNAME assigned by CDN for DNS configuration |
| `OpStatus` | `status` | Current domain status (`online`, `offline`, etc.) |

## Resource Relationships

```
AliCloudCdnDomain manifest
  │
  ├─ alicloud.Provider (region from spec.region)
  │      │
  │      └─ cdn.DomainNew (the CDN domain resource)
  │              │
  │              ├─ Sources[] (1..N origin servers)
  │              ├─ CertificateConfig (optional HTTPS)
  │              └─ Tags (system + user tags)
  │
  └─ Stack Outputs
         ├─ domain_name
         ├─ cname
         └─ status
```

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| Explicit provider per stack | Avoids reliance on ambient `ALICLOUD_REGION`; ensures region matches the manifest. |
| Conditional optional fields | Zero-value fields are omitted from the args to let the CDN API apply its own defaults. |
| Non-zero check for port/priority/weight | Protobuf int32 defaults to 0; a 0-value means "not set" and defers to provider defaults. |
| Single resource | The CDN domain + sources + certificate are one API resource; no benefit to splitting. |
| Tag merge order | System tags first, then user tags, so user tags can override system-generated values. |
| Output key constants | Centralizes key strings to avoid typos across the module and tests. |

## Customization Guide

| Goal | How |
|------|-----|
| Add a new origin source | Add an entry to `spec.sources` in the manifest. |
| Enable HTTPS | Add a `certificateConfig` block with `certType` and relevant fields. |
| Change geographic scope | Set `spec.scope` to `domestic`, `overseas`, or `global`. |
| Assign to resource group | Set `spec.resourceGroupId` to the target resource group ID. |
| Add custom tags | Add entries to `spec.tags`; they merge with system tags. |
| Rotate TLS certificate | Update `certificateConfig.certId` (CAS) or `serverCertificate`/`privateKey` (upload). |
| Switch from IP to domain origin | Change `sources[].type` from `ipaddr` to `domain` and update `content`. |

## Next Steps

- [README](./README.md) — CLI commands and debugging.
- [Examples](./examples.md) — progressive deployment examples.
- [Research Document](../../docs/README.md) — full design rationale.
