# GcpManagedKafkaAcl — Pulumi Implementation

This directory contains the Pulumi implementation for a Managed Kafka ACL from the Planton spec: one `gcp.managedkafka.Acl`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `acl` |
| `module/locals.go` | The bare cluster id |
| `module/acl.go` | Maps the entries; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`cluster`** -- the last path segment of the reference or literal.
- **`permission_type`, `host`** -- sent only when set; the provider defaults them to `ALLOW` and `*`.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
