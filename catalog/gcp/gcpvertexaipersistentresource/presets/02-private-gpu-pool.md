# Private GPU Pool

## Use Case

Hold NVIDIA L4 GPUs for a team's training jobs, scaling from one to four machines with the work, peered into the team's VPC so jobs reach private data, with least-privilege job identities and disks under your key.

## When to Use

- GPU capacity is hard to get on demand in your region
- Training jobs must reach private services in a VPC
- Security rules require user-managed job identities and CMEK

## What This Creates

- A persistent resource `gpu-training-pool` in `us-central1` with one pool of `g2-standard-8` machines, one L4 each, autoscaling between one and four, on 200 GB SSD boot disks
- VPC Network Peering to the `GcpVpcNetwork` named `ml-vpc` (it needs private services access), pinned to the `vertex-ai-range` allocation
- Jobs required to run as a custom service account
- Disks encrypted under the `GcpKmsKey` named `vertex-ai-key`
- `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `resourcePools[].machineSpec` | `g2-standard-8` + one L4 | The GPU type and count your jobs need (fixed at creation). |
| `resourcePools[].autoscalingSpec` | 1-4 | The committed floor and the ceiling. |
| `network` | `ml-vpc` reference | Your peered network; jobs must use the same one. |
| `reservedIpRanges` | `vertex-ai-range` | The peering allocation the machines take addresses from. |
| `kmsKeyName` | `vertex-ai-key` reference | Your key; jobs must use the same one. |
