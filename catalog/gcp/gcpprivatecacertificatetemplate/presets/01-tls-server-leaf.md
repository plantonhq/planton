# TLS Server Leaf

## Use Case

The shape of every internal HTTPS or gRPC server certificate: a 30-day leaf that may only authenticate a server, with the requested subject and DNS names carried through.

## When to Use

- Service teams issuing their own server certificates from a shared pool
- Any pool whose baseline values leave key usage to templates

## What This Creates

- A template in `us-central1` that stamps CA:FALSE, digital signature and key encipherment, and server authentication on every certificate issued with it

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `maximumLifetime` | `2592000s` (30 days) | Shorter rotates faster; the pool's own cap still applies. |
| `identityConstraints` | passthrough on | Add a `celExpression` to pin the allowed DNS names. |
