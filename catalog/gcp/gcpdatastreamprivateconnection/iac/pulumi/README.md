# GcpDatastreamPrivateConnection — Pulumi Implementation

This directory contains the Pulumi implementation for a Datastream private connection from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.datastream.PrivateConnection`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `privateConnection` |
| `module/locals.go` | The defaulted id and display name, the merged labels |
| `module/private_connection.go` | Enables the API; maps the one connectivity block; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`private_connection_id`, `display_name`** -- defaulting to `metadata.name`.
- **Connectivity** -- exactly one of `vpc_peering_config` or `psc_interface_config` is sent.
- **Labels** -- the spec's labels merged under the platform attribution labels.
- **`deletion_policy`** -- sent only when set, leaving the provider's `FORCE` default otherwise.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
