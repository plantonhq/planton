---
title: "Preset: Private Service Connect Endpoint"
description: "**Rank**: 3"
type: "preset"
rank: "03"
presetSlug: "03-private-psc"
componentSlug: "vertex-ai-endpoint"
componentTitle: "Vertex AI Endpoint"
provider: "gcp"
icon: "package"
order: 3
---

# Preset: Private Service Connect Endpoint

**Rank**: 3

## Use Case

The strongest network isolation for Vertex AI endpoints. Uses Private Service Connect (PSC) instead of VPC peering, with an explicit project allowlist controlling which consumer projects can connect. Ideal for multi-tenant environments or cross-project model serving.

## What This Creates

- One Vertex AI Endpoint exposed via PSC service attachment
- Explicit project allowlist for access control
- Customer-managed encryption via Cloud KMS
- No VPC peering required

## Prerequisites

- Cloud KMS key ring and key in the same region
- Consumer projects that need access must be listed in `projectAllowlist`
- Consumer projects must create PSC forwarding rules to connect

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | Must match your KMS key region |
| `projectAllowlist` | placeholder | List the consumer project IDs that need access |
| `kmsKeyName` | placeholder | Your KMS key's fully qualified path |

## Note on Dedicated Endpoint

`dedicatedEndpointEnabled` is **not available** with Private Service Connect. PSC provides its own dedicated connectivity model.
