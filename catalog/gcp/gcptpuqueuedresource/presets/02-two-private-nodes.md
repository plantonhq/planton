# Two Private Nodes

## Use Case

Request two TPU v5e nodes together on your private subnetwork without external IPs -- they come up as a pair once Google has capacity for both.

## When to Use

- Jobs that need two slices at once (data-parallel experiments, paired training and evaluation)
- Private networking policies

## What This Creates

- A queued request in `us-west4-a` for `worker-a` and `worker-b`, each a `v5litepod-8` on the `tpu` subnetwork of `ml-vpc` (Private Google Access required) without external IPs, with `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `nodeSpecs` | two nodes | Add or remove nodes (each bills once provisioned). |
| `networkConfig` | private | Your approved VPC and subnetwork. |
