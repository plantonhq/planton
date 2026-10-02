# GKE Node Pools Load on Google Provider 8.4

**Date**: October 2, 2026
**Type**: Fix
**Components**: `GcpGkeNodePool`

## Summary

**Every GKE node pool can apply again.** The OpenTofu module declared `node_config.host_maintenance_policy`, a block the Google provider removed in 8.4. The module pins `hashicorp/google ~> 8.3`, so 8.4.0 and 8.5.0 both refuse to load the module — empty field or not — and every node-pool apply fails before plan. The block is gone. A pool that leaves `host_maintenance_interval` empty (GKE's default, and what Planton's build pool sends) validates and applies. A pool that sets it is refused with a sentence that names Pulumi as the engine that still applies the field.

## What this does not do

It does not restore `PERIODIC` host maintenance on OpenTofu. That returns when the GA provider ships the block again. Pulumi still applies the field.

## Verification

`tofu init -backend=false` then `tofu validate` on the module against the resolved 8.x provider. `go test` for the kind's spec suite and `pkg/providerparity`.
