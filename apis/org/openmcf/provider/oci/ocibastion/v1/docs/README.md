# OciBastion — Design Notes

## Design Rationale

OciBastion provisions a single bastion resource. The component intentionally excludes sessions, which are ephemeral operational artifacts.

### Why not bundle sessions with the bastion?

Sessions have fundamentally different lifecycles from the bastion itself. A bastion is infrastructure — it exists for months or years. Sessions last minutes to hours and are created on-demand for specific maintenance tasks. Bundling them would force IaC to manage ephemeral resources, leading to constant drift and unnecessary re-applies.

### Why hardcode bastionType to STANDARD?

OCI's Bastion service documents only the `STANDARD` type for general use. Other types (`EPHEMERAL`, etc.) are not publicly documented or available in most tenancies. Exposing `bastionType` as a field would add complexity without benefit — users would always set it to `STANDARD`.

### Why map isDnsProxyEnabled to DnsProxyStatus?

The OCI Bastion API accepts a `DnsProxyStatus` string (`"ENABLED"` or `"DISABLED"`), not a boolean. The spec uses a boolean for ergonomics, and the Pulumi module maps `true` to `"ENABLED"` and `false` to `"DISABLED"`. When the field is omitted (nil), the OCI default applies (disabled).

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Exclude sessions | Clean separation of infrastructure and operations | Sessions must be managed via CLI or automation |
| Hardcode STANDARD type | Simpler spec; prevents invalid configurations | Cannot support future bastion types without a spec update |
| Boolean DNS proxy flag | Better YAML ergonomics than string enum | Mapping logic in the Pulumi module |
| No NSG support | Simpler spec (bastions use CIDR allow lists natively) | Cannot restrict by NSG; CIDR is the only access control |

## Resource Graph

```
OciBastion
└── oci_bastion_bastion (always, type=STANDARD)
    ├── client_cidr_block_allow_lists (0..N)
    ├── dns_proxy_status (if isDnsProxyEnabled is set)
    └── outputs: bastion_id, private_endpoint_ip_address
```

## Deferred from v1

- **oci_bastion_session** — ephemeral operational resource; managed via CLI or automation, not IaC.
- **phone_book_entry** — not applicable to STANDARD bastions.
- **static_jump_host_ip_addresses** — not applicable to STANDARD bastions.
- **security_attributes** — Oracle ZPR (Zero-Trust Packet Routing) attributes; very low adoption.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciBastion` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
