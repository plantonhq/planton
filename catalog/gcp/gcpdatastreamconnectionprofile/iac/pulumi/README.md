# GcpDatastreamConnectionProfile — Pulumi Implementation

This directory contains the Pulumi implementation for a Datastream connection profile from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.datastream.ConnectionProfile`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `connectionProfile` |
| `module/locals.go` | The defaulted id and display name, the merged labels |
| `module/connection_profile.go` | Enables the API; maps the one profile type and the connectivity option; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`connection_profile_id`, `display_name`** -- defaulting to `metadata.name`.
- **`bigquery_profile`, `srv_connection_format`** -- the spec's bools emit Google's empty marker blocks.
- **`private_connection`** -- the lifted leaf; the `private_connectivity` block is sent only when it is set.
- **Ports** -- zero is omitted, so the provider's engine default applies.
- **`ssl_config`** -- declared (even empty) means sent.
- **Passwords, keys, certificates** -- marked secret (`ToSecret`); sensitive in the provider.
- **Empty strings, empty maps, false flags** -- omitted.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
