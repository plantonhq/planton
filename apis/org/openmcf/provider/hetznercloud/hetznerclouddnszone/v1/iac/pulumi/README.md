# HetznerCloudDnsZone Pulumi Module

Pulumi (Go) IaC module for provisioning Hetzner Cloud DNS zones with record sets. Supports primary mode (records managed via spec) and secondary mode (records synchronized from an external primary nameserver via zone transfer).

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── zone.go       # Zone creation, record set creation, DNS name sanitization
│   └── outputs.go    # Output name constants
├── Pulumi.yaml       # Pulumi project configuration
└── BUILD.bazel       # Bazel build configuration
```

## Resources Created

- `hcloud.Zone` (always) — the DNS zone with domain name, mode, TTL, labels, delete protection, and (for secondary mode) primary nameserver configuration.
- `hcloud.ZoneRrset` (0–N, primary mode only) — one per record set entry in the spec. Each rrset manages all DNS records for a unique (name, type) pair.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | Hetzner Cloud numeric ID of the created DNS zone |
| `nameservers` | Authoritative Hetzner nameservers assigned to the zone |

## Usage

```bash
# Build
bazel build //apis/org/openmcf/provider/hetznercloud/hetznerclouddnszone/v1/iac/pulumi:pulumi

# Test with local manifest
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)
pulumi up
```

## Debug

```bash
# Run locally against the hack manifest
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)
export HCLOUD_TOKEN="your-api-token"
pulumi up --stack dev
```

Or use the provided debug script:

```bash
export HCLOUD_TOKEN="your-api-token"
./debug.sh
```
