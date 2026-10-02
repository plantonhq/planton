# Single v5e Node

## Use Case

Get an 8-chip TPU v5e slice in a zone where direct creates keep failing for lack of capacity: queue the request and let Google bring the node up when it can.

## When to Use

- Scarce generations and zones
- Jobs that can start whenever capacity arrives

## What This Creates

- A queued request in `us-west4-a` for one `v5litepod-8` node named `v5e-worker` on the v5e runtime

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `zone` | `us-west4-a` | A zone where your project has quota. |
| `nodeSpecs[].node.acceleratorType` | `v5litepod-8` | Larger slices or another generation (with its runtime). |
