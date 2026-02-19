# HetznerCloudDnsZone Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudDnsZoneStackInput (proto)
        ├── target: HetznerCloudDnsZone
        │     ├── metadata.name → label computation
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── domain_name → zone name
        │           ├── mode → primary or secondary
        │           ├── ttl → zone default TTL
        │           ├── delete_protection → zone protection flag
        │           ├── primary_nameservers[] → zone transfer config (secondary only)
        │           │     ├── address, port
        │           │     └── tsig_algorithm, tsig_key
        │           └── record_sets[] → one hcloud.ZoneRrset per entry
        │                 ├── name, type → rrset identity
        │                 ├── ttl → per-rrset TTL override
        │                 └── records[] → values (StringValueOrRef) + comments
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudDnsZoneStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `zone()` to create the zone, record sets, and export outputs

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/zone.go**: The core resource creation file. Contains three functions:

   **`zone()`** — creates the `hcloud.Zone` with domain name, mode, labels, TTL, delete protection, and (for secondary mode) primary nameservers. After creating the zone, calls `createRecordSets()` for primary-mode zones. Exports `zone_id` and `nameservers`.

   **`createRecordSets()`** — iterates over `spec.RecordSets` and creates one `hcloud.ZoneRrset` per entry. Each rrset is named `rrset-{sanitized_name}-{lowercase_type}` (e.g., `rrset-at-a`, `rrset-www-cname`). Records are built from `StringValueOrRef.GetValue()` with optional comments.

   **`sanitizeDnsName()`** — converts DNS record names into Pulumi-safe resource name components:
   - `@` → `at` (zone apex)
   - `*` → `wildcard`
   - `.`, `/`, `:` → `-`

5. **module/outputs.go**: Two constants matching `stack_outputs.proto` field names:
   - `OpZoneId` = `"zone_id"`
   - `OpNameservers` = `"nameservers"`

## Resource Graph

```
                        spec
                         │
            ┌────────────┴────────────────────┐
            │                                 │
    zone config                      record_sets[]
    (domain, mode, ttl,              (primary mode only)
     delete_protection,                    │
     primary_nameservers)                  │
            │                    ┌─────────┼──────────┐
            │                    │         │          │
    hcloud.Zone              ZoneRrset  ZoneRrset  ZoneRrset ...
    ("zone")                 ("rrset-   ("rrset-   ("rrset-
            │                 at-a")    www-cname") at-mx")
            │                    │         │          │
            │                    └─────────┴──────────┘
            │                              │
            │                    Each depends on zone.ID
            │
    Outputs:
    ├── zone_id       ← zone.ID()
    └── nameservers   ← zone.AuthoritativeNameservers.Assigneds()
```

## Key Design Points

- **Single resource file for zone + rrsets**: Both zone creation and record set creation live in `zone.go` because they are a single logical concern. The zone is the parent; rrsets are its children. Splitting into `zone.go` and `rrsets.go` would add indirection without benefit.

- **`sanitizeDnsName` for Pulumi resource naming**: DNS names contain characters (`@`, `*`, `.`) that are not valid in Pulumi resource identifiers. The `sanitizeDnsName` function converts these to safe alternatives. The mapping is deterministic, so resource names remain stable across deployments (no unnecessary replacements).

- **StringValueOrRef resolution**: Record values use `rec.Value.GetValue()` to extract the literal string. When used in an infra chart with `valueFrom`, the OpenMCF runtime resolves the reference before the Pulumi module executes — the module always receives a resolved string. The module does not contain any cross-stack reference logic.

- **CG02 sub-resource keying**: Record sets are named by `{sanitized_name}-{lowercase_type}`, which matches the CG02 keying pattern. This ensures that adding a new record set does not cause existing rrsets to be renamed or replaced. The key is stable as long as the record set's name and type do not change.

- **Mode-conditional primary nameservers**: The zone function only sets `PrimaryNameservers` on the `ZoneArgs` when the spec has a non-empty `primary_nameservers` slice. For primary-mode zones (which should never have primary nameservers per the proto validation), this field is simply not set — the provider default of no primary nameservers applies.

- **Zone ID type conversion**: `hcloud.Zone.ID()` returns a `pulumi.IDOutput`. The `createRecordSets` function receives it as a `pulumi.StringOutput` via `.ToStringOutput()`. This conversion is necessary because `ZoneRrsetArgs.Zone` expects a `pulumi.StringPtrInput`, not a `pulumi.IDOutput`.

- **Label merge strategy**: Same CG01 pattern as all other Hetzner Cloud components. Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) always take precedence over user-specified labels. Labels are applied to the zone resource only — individual rrset resources do not carry labels (the provider does not surface rrset labels in a way that's useful for organization).

- **No ID conversion**: Unlike components that reference external resources by numeric ID (e.g., HetznerCloudSnapshot converts `serverId` from string to int), this module has no input ID conversion. The zone ID is an output, and record values are strings. The only input that references another resource is `StringValueOrRef`, which is resolved externally.
