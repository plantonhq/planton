# Hetzner Cloud DNS Zone — Research Documentation

## Introduction

DNS is the address book of the internet, and a DNS zone is the authoritative source of truth for a single domain. Every service that users reach by name — web servers, mail servers, APIs — depends on records in a DNS zone resolving correctly. Managing DNS through infrastructure-as-code is not optional for production systems: a mistyped A record or a forgotten MX entry can silently break email delivery or take a website offline.

Hetzner Cloud added first-party DNS hosting in 2024, giving customers authoritative nameservers integrated into the same API and console they use for servers, networks, and load balancers. The `HetznerCloudDnsZone` component brings this capability into OpenMCF as a single declarative resource that provisions a zone and its record sets together, with cross-component references that tie DNS records to the IP addresses of other Hetzner Cloud resources.

The component supports two operating modes. In **primary** mode, Hetzner Cloud is the authoritative nameserver and records are managed directly through the OpenMCF manifest. In **secondary** mode, Hetzner Cloud acts as a slave, synchronizing records from an external primary nameserver via zone transfer (AXFR/IXFR), optionally authenticated with TSIG.

## DNS in the Hetzner Cloud Ecosystem

### Zone Modes: Primary vs Secondary

Hetzner Cloud DNS zones operate in one of two modes, selected at creation time and immutable thereafter (changing mode forces replacement):

**Primary mode** — Hetzner Cloud is the authoritative nameserver. You manage records directly through the API, CLI, console, or IaC tools. Hetzner Cloud assigns three authoritative nameservers (e.g., `helium.ns.hetzner.de`, `hydrogen.ns.hetzner.com`, `oxygen.ns.hetzner.com`) that you configure at your domain registrar as the NS records for the domain.

**Secondary mode** — Hetzner Cloud acts as a secondary (slave) nameserver, pulling records from one or more external primary nameservers via zone transfer. You cannot manage records directly in secondary mode — they are read-only copies of whatever the primary serves. This is useful for DNS redundancy: your primary DNS is hosted elsewhere (e.g., BIND, PowerDNS, another cloud provider), and Hetzner Cloud provides geographically distributed secondary resolution.

The distinction matters for the component's field validation:
- Primary mode: `recordSets` is valid, `primaryNameservers` is forbidden.
- Secondary mode: `primaryNameservers` is required (at least one), `recordSets` is forbidden.

These constraints are enforced at the proto level via CEL expressions, so invalid manifests fail validation before any cloud API call is made.

### Record Types

Hetzner Cloud DNS supports standard record types through the `hcloud_zone_rrset` resource:

| Type | Purpose | Value Format | Example |
|------|---------|--------------|---------|
| A | IPv4 address | Dotted-decimal | `93.184.216.34` |
| AAAA | IPv6 address | Colon-hex | `2606:2800:220:1:248:1893:25c8:1946` |
| CNAME | Canonical name alias | FQDN with trailing dot | `example.com.` |
| MX | Mail exchange | Priority + FQDN | `10 mail.example.com.` |
| TXT | Arbitrary text | Quoted string | `"v=spf1 include:_spf.google.com ~all"` |
| NS | Nameserver delegation | FQDN with trailing dot | `ns1.example.com.` |
| SRV | Service locator | Priority weight port target | `10 60 5060 sip.example.com.` |
| CAA | Certificate authority authorization | Flag tag value | `0 issue "letsencrypt.org"` |
| PTR | Reverse DNS | FQDN with trailing dot | `host.example.com.` |
| TLSA | TLS authentication | Usage selector type cert | `3 1 1 abc123...` |
| DS | Delegation signer (DNSSEC) | Key-tag algorithm type digest | `12345 8 2 abc123...` |

Record values follow DNS standards. Common formatting pitfalls:
- **MX records** include the priority as part of the value string: `"10 mail.example.com."` — not separate fields.
- **TXT records** must be enclosed in double quotes within the YAML value: `"\"v=spf1 ...\""`
- **CNAME and MX targets** require a trailing dot to indicate a fully qualified domain name.
- **TXT records longer than 255 characters** must be split into multiple quoted strings per RFC 4408: `"\"first 255 chars\"\"next 255 chars\""`.

### Record Sets vs Individual Records

The Hetzner Cloud provider offers two resource types for managing DNS records:

- **`hcloud_zone_rrset`** — manages all records for a (name, type) pair as a single resource. Adding, removing, or changing values within the set is an in-place update. This is the recommended approach and what OpenMCF uses.
- **`hcloud_zone_record`** — manages individual records. Each record is a separate resource. Adding a second A record for the same name requires a second `hcloud_zone_record` resource. Cannot manage TTL (TTL is set at the RRSet level). Only use this when RRSet management is not possible (the provider docs explicitly recommend RRSet).

OpenMCF uses `hcloud_zone_rrset` exclusively because:
1. It matches the DNS data model — a record set is the natural unit for a (name, type) pair.
2. It allows in-place updates — adding a value to an existing record set does not require destroying and recreating.
3. It controls TTL — the TTL is set per record set, which is how DNS actually works.
4. It reduces resource count — one record set with 3 A records is one resource, not three.

### TSIG Authentication

Transaction Signature (TSIG) is a mechanism for authenticating DNS messages, primarily zone transfers between primary and secondary nameservers. When a secondary zone is configured with TSIG:

1. The secondary (Hetzner Cloud) sends a zone transfer request (AXFR or IXFR) to the primary nameserver.
2. The request includes a TSIG signature computed from a shared secret key using a specified algorithm (e.g., HMAC-SHA256).
3. The primary verifies the signature before responding.
4. The response is also signed, and the secondary verifies it.

TSIG prevents unauthorized zone transfers, which could expose all DNS records to an attacker. Without TSIG, zone transfers rely on IP-based ACLs alone — sufficient in some environments, but TSIG adds a cryptographic authentication layer.

The OpenMCF spec exposes `tsigAlgorithm` and `tsigKey` on each primary nameserver entry. Both fields are optional: omit them if the primary does not require TSIG. The `tsigKey` is a shared secret and should be treated as sensitive, though the current proto does not mark it with a sensitive annotation (the provider transmits it over TLS to the Hetzner Cloud API).

### Zone Lifecycle

Key lifecycle characteristics:

- **Creation**: Zone is created with its mode and domain name. In primary mode, Hetzner Cloud assigns authoritative nameservers. In secondary mode, Hetzner Cloud initiates zone transfer from the configured primary nameservers.
- **NS delegation**: After creation, you must configure the assigned nameservers at your domain registrar. Until NS delegation is complete, the zone exists but is not authoritative for the domain on the public internet.
- **Immutable fields**: `domainName` and `mode` force replacement if changed. All other fields (TTL, records, labels, delete protection, primary nameservers) can be updated in place.
- **Delete protection**: When enabled, the Hetzner Cloud API rejects deletion requests. The IaC module must first disable protection, then delete the zone. Both Pulumi and Terraform handle this automatically.
- **Record propagation**: Changes to record sets take effect immediately in the Hetzner Cloud nameservers. External DNS resolvers see the changes after the TTL of the previous record expires.

### Pricing

Hetzner Cloud DNS is free. There is no per-zone or per-query charge. This makes it an attractive option for projects already on Hetzner Cloud infrastructure — no separate DNS provider subscription needed.

This contrasts with:
- AWS Route53: $0.50/hosted zone/month + $0.40 per million queries
- Cloudflare: Free tier available, paid plans for advanced features
- Google Cloud DNS: $0.20/zone/month + $0.40 per million queries

For Hetzner Cloud users, the cost advantage is clear: zero marginal cost for DNS, regardless of zone count or query volume.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

#### Creating a Primary Zone

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **DNS** in the left sidebar
3. Click **Add Zone**
4. Enter the domain name (e.g., `example.com`)
5. Select **Primary** mode
6. Set the default TTL (or accept the 3600-second default)
7. Click **Create Zone**
8. The console shows the assigned nameservers — configure these at your registrar
9. Navigate to the zone and click **Add Record** for each record set

#### Creating a Secondary Zone

1. Navigate to **DNS** > **Add Zone**
2. Enter the domain name
3. Select **Secondary** mode
4. Enter the primary nameserver address(es), optionally with TSIG credentials
5. Click **Create Zone**
6. Hetzner Cloud initiates a zone transfer; records appear once the transfer completes

#### Managing Records

The console provides a table view of all record sets in a zone. Each record set can be edited inline: change values, add/remove records within a set, adjust TTL. The console automatically groups records by name and type.

**Pros:**
- Visual overview of all zones and records
- Immediate feedback on record creation
- Built-in validation for record value formats
- No tooling required

**Cons:**
- No version control for DNS configurations
- No way to reproduce a zone across environments
- No audit trail beyond the Hetzner Cloud API log
- Manual record entry is error-prone for complex zones (SPF, DKIM, DMARC)
- No dependency tracking between DNS records and the infrastructure they point to
- Tedious for zones with many record sets

**Verdict:** Acceptable for personal domains or initial exploration. Not suitable for production zones where reproducibility, drift detection, and cross-resource references matter.

### Level 1: CLI (`hcloud`)

#### Zone Management

```bash
# Create a primary zone
hcloud zone create --name example.com --mode primary --ttl 3600

# Create a secondary zone with TSIG
hcloud zone create --name example.com --mode secondary \
  --primary-nameserver "address=10.0.0.1,port=53,tsig_algorithm=hmac-sha256,tsig_key=secret123"

# List zones
hcloud zone list

# Describe a zone (shows nameservers, TTL, mode)
hcloud zone describe example.com

# Update TTL
hcloud zone update --ttl 1800 example.com

# Enable delete protection
hcloud zone update --enable-delete-protection example.com

# Delete a zone
hcloud zone delete example.com
```

#### Record Set Management

```bash
# Add an A record set
hcloud zone rrset create --zone example.com --name "@" --type A --ttl 300 \
  --record "value=93.184.216.34,comment=web-1" \
  --record "value=93.184.216.35,comment=web-2"

# Add a CNAME
hcloud zone rrset create --zone example.com --name www --type CNAME \
  --record "value=example.com."

# Add MX records
hcloud zone rrset create --zone example.com --name "@" --type MX \
  --record "value=10 mail.example.com." \
  --record "value=20 backup.example.com."

# Update a record set (replaces all records)
hcloud zone rrset update --zone example.com --name "@" --type A \
  --record "value=93.184.216.36"

# Delete a record set
hcloud zone rrset delete --zone example.com --name www --type CNAME

# List all record sets in a zone
hcloud zone rrset list --zone example.com
```

**Key CLI behaviors:**
- Zone creation returns immediately. For secondary zones, the initial zone transfer happens asynchronously.
- RRSet create/update replaces all records in the set. There is no "add a record to an existing set" — you provide the complete desired state.
- Zones can be referenced by name or numeric ID.
- The CLI does not support bulk record import from a zone file.

**Pros:**
- Scriptable for CI/CD pipelines
- Fast for one-off operations
- Direct access to all zone and record set features

**Cons:**
- No state tracking — cannot detect if someone changed a record via the console
- No dependency management between DNS records and infrastructure
- Scripts accumulate complexity for zones with many records
- No rollback mechanism — mistakes are immediately live
- TSIG keys appear in command history (security concern)

**Verdict:** Good for quick operations, debugging, and verification. Not suitable as the primary management method for production DNS.

### Level 2: IaC (Terraform)

#### Primary Zone with Records

```hcl
resource "hcloud_zone" "example" {
  name              = "example.com"
  mode              = "primary"
  ttl               = 3600
  delete_protection = true

  labels = {
    env = "production"
  }
}

resource "hcloud_zone_rrset" "apex_a" {
  zone = hcloud_zone.example.id
  name = "@"
  type = "A"
  ttl  = 300

  records = [
    { value = "93.184.216.34", comment = "web-1" },
    { value = "93.184.216.35", comment = "web-2" },
  ]
}

resource "hcloud_zone_rrset" "www_cname" {
  zone = hcloud_zone.example.id
  name = "www"
  type = "CNAME"

  records = [
    { value = "example.com." },
  ]
}

resource "hcloud_zone_rrset" "mx" {
  zone = hcloud_zone.example.id
  name = "@"
  type = "MX"

  records = [
    { value = "10 mail.example.com." },
    { value = "20 backup.example.com." },
  ]
}
```

#### Secondary Zone with TSIG

```hcl
resource "hcloud_zone" "secondary" {
  name = "example.com"
  mode = "secondary"

  primary_nameservers = [
    {
      address        = "10.0.0.1"
      port           = 53
      tsig_algorithm = "hmac-sha256"
      tsig_key       = var.tsig_secret
    }
  ]
}
```

#### Key Terraform behaviors

- `hcloud_zone` requires `name` (the domain) and `mode`. The `mode` and `name` are ForceNew.
- `hcloud_zone_rrset` requires `zone`, `name`, `type`, and at least one `records` entry.
- The `records` block is a set, not a list — order does not matter for plan stability.
- TTL at the zone level is a default; TTL on the rrset overrides it. If rrset TTL is not specified, the zone TTL applies.
- `primary_nameservers` is a list of objects, forbidden for primary zones and required for secondary zones.
- The `authoritative_nameservers.assigned` computed attribute returns the list of Hetzner-assigned nameservers.
- Delete protection must be disabled before the zone can be destroyed. Terraform handles this automatically during `destroy`.

**Pros:**
- State tracking and drift detection for all zones and records
- Dependency graph between zones and record sets
- Plan preview shows exact changes before applying
- TSIG keys can be injected via variables (no command-history exposure)
- Reproducible across environments via `.tfvars` files

**Cons:**
- Each record set is a separate `hcloud_zone_rrset` resource block, creating verbose configurations for zones with many records
- No built-in zone-file import — each record set must be declared individually
- The for_each pattern for dynamic record sets requires careful key management to avoid unnecessary replacements
- No cross-stack references to other Hetzner resources without manual output wiring

**Verdict:** Production-ready for DNS management. The verbosity of one resource block per record set is the main ergonomic cost, which OpenMCF's nested record set array solves.

### Level 3: IaC (Pulumi)

#### Primary Zone with Records

```go
z, err := hcloud.NewZone(ctx, "example", &hcloud.ZoneArgs{
    Name:             pulumi.StringPtr("example.com"),
    Mode:             pulumi.String("primary"),
    Ttl:              pulumi.IntPtr(3600),
    DeleteProtection: pulumi.BoolPtr(true),
    Labels: pulumi.StringMap{
        "env": pulumi.String("production"),
    },
})
if err != nil {
    return err
}

_, err = hcloud.NewZoneRrset(ctx, "apex-a", &hcloud.ZoneRrsetArgs{
    Zone: z.ID().ToStringOutput(),
    Name: pulumi.StringPtr("@"),
    Type: pulumi.String("A"),
    Ttl:  pulumi.IntPtr(300),
    Records: hcloud.ZoneRrsetRecordArray{
        hcloud.ZoneRrsetRecordArgs{
            Value:   pulumi.String("93.184.216.34"),
            Comment: pulumi.StringPtr("web-1"),
        },
        hcloud.ZoneRrsetRecordArgs{
            Value:   pulumi.String("93.184.216.35"),
            Comment: pulumi.StringPtr("web-2"),
        },
    },
}, pulumi.Provider(provider))
```

#### Secondary Zone with TSIG

```go
_, err := hcloud.NewZone(ctx, "secondary", &hcloud.ZoneArgs{
    Name: pulumi.StringPtr("example.com"),
    Mode: pulumi.String("secondary"),
    PrimaryNameservers: hcloud.ZonePrimaryNameserverArray{
        hcloud.ZonePrimaryNameserverArgs{
            Address:       pulumi.String("10.0.0.1"),
            Port:          pulumi.IntPtr(53),
            TsigAlgorithm: pulumi.StringPtr("hmac-sha256"),
            TsigKey:       pulumi.StringPtr(cfg.RequireSecret("tsig-key")),
        },
    },
}, pulumi.Provider(provider))
```

**Key Pulumi behaviors:**
- `hcloud.Zone` and `hcloud.ZoneRrset` mirror the Terraform resources exactly (bridged provider).
- Zone ID is `pulumi.IDOutput` — must be converted via `.ToStringOutput()` for use in `ZoneRrsetArgs.Zone`.
- `pulumi.StringPtr` is used for optional string fields (Name on ZoneRrset), `pulumi.String` for required strings.
- Secrets can use `cfg.RequireSecret` for config-based secrets or `pulumi.ToSecret` for inline values.
- The Go type system catches field name typos and type mismatches at compile time.

**Pros:**
- Same benefits as Terraform: state, drift detection, plan preview
- Go type safety catches errors at compile time
- Loops and conditionals for dynamic record generation are native Go, not HCL
- Built-in secret handling for TSIG keys

**Cons:**
- More verbose than Terraform for simple cases (Go struct initialization)
- Requires Go toolchain
- Same per-resource-call pattern as Terraform for each record set

**Verdict:** Production-ready, with the ergonomic advantage of Go for complex dynamic record generation. The per-call verbosity is the same trade-off as Terraform.

## Comparative Analysis

| Aspect | Console | CLI | Terraform | Pulumi | OpenMCF |
|--------|---------|-----|-----------|--------|---------|
| **Reproducible** | No | Partial | Yes | Yes | Yes |
| **State tracked** | No | No | Yes | Yes | Yes |
| **Drift detection** | No | No | Yes | Yes | Yes |
| **Cross-resource refs** | No | No | Manual outputs | Manual outputs | `valueFrom` |
| **Record grouping** | Automatic | Per-command | Per-resource block | Per-resource call | Nested array |
| **Zone + records in one** | No (separate steps) | No | No (separate resources) | No (separate resources) | Yes (single manifest) |
| **TSIG handling** | Form field | CLI flag | Variable (sensitive) | Config secret | Proto field |
| **Validation** | API-level | API-level | Plan-time | Compile-time | Proto-level (pre-deploy) |
| **Version control** | No | Scripts | HCL files | Go source | YAML manifests |

## The OpenMCF Approach

### Why Bundle Zone and Record Sets

Terraform and Pulumi treat zones and record sets as separate resources. OpenMCF bundles them into a single component because:

1. **A zone without records is incomplete.** Creating a zone is only the first step — the zone is useless until records are added. Bundling ensures the zone and its records are deployed atomically.

2. **Records belong to the zone.** DNS record sets do not exist independently — they are always scoped to a zone. The one-to-many relationship is naturally expressed as a nested array in the spec.

3. **Fewer manifests.** A domain with 10 record types needs one manifest, not eleven. This reduces the operational surface area.

4. **Atomic updates.** Adding a record set to an existing zone is a single manifest update, not a new resource file plus a dependency wire.

The trade-off is that very large zones (hundreds of record sets) produce large manifests. In practice, most domains need 5-15 record sets, which fits comfortably in a single YAML file.

### The 80/20 Field Scoping

**Included fields (the 80%):**

| Field | Why Included |
|-------|-------------|
| `domainName` | The zone's domain — the most fundamental configuration. |
| `mode` | Primary vs secondary determines the entire zone behavior. |
| `ttl` | Default TTL affects all records. Important for caching strategy. |
| `recordSets` | The records themselves — the primary reason the zone exists. |
| `primaryNameservers` | Required for secondary mode. Without this, secondary zones cannot function. |
| `deleteProtection` | Prevents accidental deletion of a zone with live DNS traffic. |
| `recordSets[].name` | Record name (e.g., "@", "www", "mail") — required per DNS standard. |
| `recordSets[].type` | Record type (A, AAAA, MX, etc.) — required per DNS standard. |
| `recordSets[].ttl` | Per-record-set TTL override — needed for records with different caching needs (e.g., low TTL for failover). |
| `recordSets[].records[].value` | The record value itself, with `StringValueOrRef` for cross-component references. |
| `recordSets[].records[].comment` | Per-record comments — useful for documenting purpose ("primary web server", "SPF record"). |

**Excluded fields (derived or unnecessary):**

| Field | Why Excluded |
|-------|-------------|
| Zone `name` (Hetzner name) | Set to `domainName`. The Hetzner zone "name" is the domain, not a display name. |
| `labels` | Derived from metadata following CG01. Standard labels take precedence. |
| RRSet `labels` | Not exposed — zone-level labels are sufficient for organization. |
| RRSet `changeProtection` | Not exposed — delete protection at the zone level provides sufficient safety. |
| `registrar` | Read-only computed attribute. Not configurable. |

### StringValueOrRef for Record Values

The most distinctive feature of this component is that record values use `StringValueOrRef` instead of plain strings. This enables infra-chart composability: a DNS A record can reference the IPv4 address output of a `HetznerCloudServer` or `HetznerCloudFloatingIp` instead of hardcoding an IP address.

```yaml
recordSets:
  - name: "@"
    type: A
    records:
      - value:
          valueFrom:
            kind: HetznerCloudServer
            name: web-1
            fieldPath: status.outputs.ipv4_address
```

When the server's IP changes (e.g., after a replacement), the DNS record updates automatically on the next deployment. Without `valueFrom`, the operator must manually copy the new IP into the DNS manifest — a common source of configuration drift.

For literal values, `StringValueOrRef` accepts a plain string:

```yaml
records:
  - value: "93.184.216.34"
```

This dual nature means the same field works for both static records (MX, TXT, CAA) and dynamic records (A, AAAA pointing to infrastructure).

### Mode-Dependent Validation

The spec uses CEL expressions to enforce mode-dependent constraints at the proto validation level:

1. Primary mode forbids `primaryNameservers` — these are meaningless when Hetzner Cloud is the primary.
2. Secondary mode requires `primaryNameservers` — Hetzner Cloud needs at least one server to transfer from.
3. Secondary mode forbids `recordSets` — records come from the primary and cannot be managed locally.

These rules catch configuration errors before any cloud API call:

```
# Invalid: primary zone with primaryNameservers
# → "primary_nameservers must be empty when mode is primary"

# Invalid: secondary zone without primaryNameservers
# → "primary_nameservers is required when mode is secondary"

# Invalid: secondary zone with recordSets
# → "record_sets must be empty when mode is secondary"
```

This is a significant improvement over Terraform and Pulumi, where these constraints are only enforced by the Hetzner Cloud API at apply time.

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module follows the standard OpenMCF pattern:

```
main.go           → loads HetznerCloudDnsZoneStackInput, calls module.Resources
module/main.go    → initializes locals, sets up provider, calls zone()
module/locals.go  → extracts config, computes labels (CG01)
module/zone.go    → creates hcloud.Zone + hcloud.ZoneRrset per record set
module/outputs.go → output name constants (zone_id, nameservers)
```

**`zone()` function** creates the `hcloud.Zone` with:
- `Name` set to `spec.DomainName`
- `Mode` set to `spec.Mode.String()` (proto enum to string conversion)
- `Labels` from locals (CG01 merge)
- `DeleteProtection` from spec
- `Ttl` from spec (if set)
- `PrimaryNameservers` from spec (if non-empty, for secondary mode)

Then calls `createRecordSets()` which iterates over `spec.RecordSets` and creates one `hcloud.NewZoneRrset` per record set.

**Resource naming** uses `sanitizeDnsName()` to convert DNS names into Pulumi-safe resource identifiers:
- `@` → `at` (apex records)
- `*` → `wildcard`
- `.`, `/`, `:` → `-`

The resulting Pulumi resource name is `rrset-{sanitized_name}-{lowercase_type}`, e.g., `rrset-at-a`, `rrset-www-cname`, `rrset-wildcard-a`.

**StringValueOrRef resolution** uses `rec.Value.GetValue()` to extract the literal string. When used in an infra chart with `valueFrom`, the OpenMCF runtime resolves the reference before the Pulumi module sees it — the module always receives a resolved string.

**Outputs** are exported directly from the zone resource:
- `zone_id` → `createdZone.ID()`
- `nameservers` → `createdZone.AuthoritativeNameservers.Assigneds()`

### Terraform Module Architecture

The Terraform module uses the same two-resource pattern:

```
provider.tf    → hetznercloud/hcloud ~> 1.60
variables.tf   → metadata, spec, hcloud_token
locals.tf      → standard_labels (CG01), record_sets map
main.tf        → hcloud_zone.this + hcloud_zone_rrset.this (for_each)
outputs.tf     → zone_id, nameservers
```

**Record set keying** uses `for_each` with keys computed as `"${rs.name}-${lower(rs.type)}"`:

```hcl
local.record_sets = {
  for rs in (var.spec.record_sets != null ? var.spec.record_sets : []) :
  "${rs.name}-${lower(rs.type)}" => rs
}
```

This produces keys like `@-a`, `www-cname`, `@-mx` — matching the CG02 sub-resource keying pattern.

**Conditional primary nameservers** uses a ternary to transform the list only when present:

```hcl
primary_nameservers = var.spec.primary_nameservers != null ? [
  for ns in var.spec.primary_nameservers : { ... }
] : null
```

**Outputs** read from the zone resource:
- `zone_id` → `hcloud_zone.this.id`
- `nameservers` → `hcloud_zone.this.authoritative_nameservers.assigned`

### Resource Dependency Graph

```
HetznerCloudDnsZoneStackInput
         │
         ├── provider_config → hcloud Provider
         │
         └── target.spec
               │
               ├── domain_name, mode, ttl, delete_protection
               │        │
               │        └── hcloud_zone
               │               │
               │               ├── zone_id (output)
               │               └── nameservers (output)
               │
               ├── primary_nameservers (secondary mode only)
               │        │
               │        └── hcloud_zone.primary_nameservers
               │
               └── record_sets[]
                        │
                        └── hcloud_zone_rrset (one per record set)
                               │
                               └── depends_on: hcloud_zone (via zone ID)
```

## Production Best Practices

### NS Delegation Checklist

After creating a primary zone, the zone is not active on the public internet until you configure NS delegation at your domain registrar:

1. **Create the zone** — note the assigned nameservers from the `nameservers` output.
2. **Log in to your registrar** (Namecheap, GoDaddy, Gandi, etc.).
3. **Set custom nameservers** for the domain to the Hetzner-assigned values (typically three: `helium.ns.hetzner.de`, `hydrogen.ns.hetzner.com`, `oxygen.ns.hetzner.com`).
4. **Wait for propagation** — NS record changes at the registrar can take minutes to 48 hours to propagate globally.
5. **Verify** — use `dig NS example.com` to confirm the Hetzner nameservers are active.

Until NS delegation completes, the zone's records exist in Hetzner Cloud but are not served to the public internet. Existing DNS resolution continues via the previous nameservers.

### TTL Strategy

TTL values control how long DNS resolvers cache records. The right TTL depends on the record's purpose:

| Record Type | Recommended TTL | Rationale |
|-------------|----------------|-----------|
| A/AAAA for stable servers | 3600 (1 hour) | Reduces query volume. Acceptable propagation delay for stable IPs. |
| A/AAAA for failover IPs | 60-300 (1-5 min) | Fast propagation when IP changes during failover. |
| CNAME | 3600 | Aliases rarely change. |
| MX | 3600-86400 | Mail servers are stable. Higher TTL reduces DNS load. |
| TXT (SPF, DKIM, DMARC) | 3600 | Rarely changes, but not so high that updates take days. |
| NS | 86400 (24 hours) | Nameserver delegation is very stable. |
| CAA | 3600 | Certificate policies rarely change. |

The zone-level `ttl` sets the default. Override it per record set for records with different caching needs. A common pattern: zone TTL of 3600, with a low TTL of 60-300 on A/AAAA records that might change during maintenance or failover.

### Record Value Formatting

Common mistakes with DNS record values:

**MX records** — the priority is part of the value string, not a separate field:
```yaml
# Correct
- value: "10 mail.example.com."

# Wrong — no priority
- value: "mail.example.com."
```

**TXT records** — values must be enclosed in quotes within the YAML string:
```yaml
# Correct — escaped quotes for the TXT value
- value: "\"v=spf1 include:_spf.google.com ~all\""

# Wrong — unquoted TXT value
- value: "v=spf1 include:_spf.google.com ~all"
```

**CNAME/MX targets** — must end with a trailing dot to indicate an FQDN:
```yaml
# Correct
- value: "example.com."

# Wrong — interpreted as relative to the zone
- value: "example.com"
```

**CAA records** — flag, tag, and value with the value in quotes:
```yaml
# Correct
- value: "0 issue \"letsencrypt.org\""
```

### TSIG Key Security

For secondary zones with TSIG authentication:

- **Generate strong keys** — use `tsig-keygen -a hmac-sha256 example-key` on your BIND server or equivalent tools. Keys should be base64-encoded random data, not human-readable passwords.
- **Rotate keys periodically** — TSIG keys are shared secrets. Rotate them on a schedule (e.g., quarterly) and update both the primary and secondary configurations.
- **Limit scope** — each TSIG key should be specific to one zone or one zone-transfer relationship. Do not reuse keys across zones.
- **Avoid committing keys** — use environment variables, sealed secrets, or a secrets manager to inject TSIG keys into manifests. Do not store them in version control.

### Delete Protection

Enable delete protection (`deleteProtection: true`) for production zones. This prevents accidental deletion via the API, CLI, or IaC tools. The IaC modules handle the protection lifecycle automatically:

- On destroy, protection is disabled first, then the zone is deleted.
- On update with `deleteProtection: false`, protection is removed.

Without delete protection, a `terraform destroy` or `pulumi destroy` silently removes a zone with all its records — taking the domain offline instantly.

### Monitoring and Health

Hetzner Cloud DNS does not provide built-in monitoring. Complement it with:

- **External DNS monitoring** (Pingdom, UptimeRobot, or similar) — verify that your domain resolves correctly from multiple geographic locations.
- **Certificate monitoring** — if using `HetznerCloudCertificate` with managed (Let's Encrypt) certificates, the ACME HTTP-01 challenge depends on DNS being correct. Monitor certificate expiration as an indirect DNS health signal.
- **TTL-aware alerting** — when changing records, allow at least one full TTL cycle before expecting changes to be visible globally.

### Cross-Component Reference Patterns

The `StringValueOrRef` mechanism enables common DNS integration patterns:

**Server A record:**
```yaml
recordSets:
  - name: "@"
    type: A
    records:
      - value:
          valueFrom:
            kind: HetznerCloudServer
            name: web-1
            fieldPath: status.outputs.ipv4_address
```

**Floating IP for failover:**
```yaml
recordSets:
  - name: "@"
    type: A
    records:
      - value:
          valueFrom:
            kind: HetznerCloudFloatingIp
            name: failover-ip
            fieldPath: status.outputs.ip_address
```

**Load balancer:**
```yaml
recordSets:
  - name: "@"
    type: A
    records:
      - value:
          valueFrom:
            kind: HetznerCloudLoadBalancer
            name: web-lb
            fieldPath: status.outputs.ipv4_address
```

These references create dependency edges in the deployment graph — the DNS zone waits for the referenced resource to be provisioned before creating the record with the resolved IP address.

## References

- [Hetzner Cloud DNS Documentation](https://docs.hetzner.cloud/#dns)
- [Hetzner Cloud DNS Console](https://dns.hetzner.com)
- [Terraform hcloud_zone Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/zone)
- [Terraform hcloud_zone_rrset Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/zone_rrset)
- [Pulumi hcloud.Zone Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/zone/)
- [Pulumi hcloud.ZoneRrset Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/zonerrset/)
- [RFC 1035 — Domain Names: Implementation and Specification](https://datatracker.ietf.org/doc/html/rfc1035)
- [RFC 2845 — Secret Key Transaction Authentication for DNS (TSIG)](https://datatracker.ietf.org/doc/html/rfc2845)
- [RFC 8659 — DNS Certification Authority Authorization (CAA)](https://datatracker.ietf.org/doc/html/rfc8659)
