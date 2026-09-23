# Private Reserved v6e

## Use Case

Production training on a Trillium (v6e) slice drawn from your reservation, on a private subnetwork with no external IPs, running as its own service account, reading a shared dataset disk.

## When to Use

- Long training runs on capacity you have reserved
- Security policies that forbid public IPs and the default service account

## What This Creates

- A `v6e-8` slice in `us-east5-b` from a reservation, on the `tpu` subnetwork of `ml-vpc` (Private Google Access required) without external IPs, as the `tpu-trainer` account, with the `training-dataset` disk attached read-only, Secure Boot, a `tpu` firewall tag, and `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `zone` / `acceleratorType` | `us-east5-b` / `v6e-8` | Where and what your reservation covers. |
| `dataDisks` | a read-only dataset | Add a `READ_WRITE` disk for checkpoints. |
| `schedulingConfig.reserved` | `true` | Remove to use on-demand capacity. |
