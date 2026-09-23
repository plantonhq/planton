# Private GPU

## Use Case

A GPU notebook machine for sensitive data: on your private subnetwork with no internet path, users' own credentials blocked, Secure Boot on, disks encrypted under your key, and a 30-minute idle shutdown.

## When to Use

- Model development on regulated or confidential data
- Teams whose security policy forbids public egress from analysis machines

## What This Creates

- A template on `n1-standard-8` with one T4 GPU, a 200 GB SSD, a private subnetwork (Private Google Access required), no internet access, EUC disabled, Secure Boot, a `notebooks` network tag, CMEK, a 30-minute idle shutdown, and `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `machineSpec.acceleratorType` | `NVIDIA_TESLA_T4` | `NVIDIA_L4` or `NVIDIA_TESLA_A100` for larger models (with a matching machine type). |
| `networkSpec` | private | The VPC and subnetwork your security team approved. |
| `idleTimeout` | `1800s` | GPU time is expensive; keep it short. |
