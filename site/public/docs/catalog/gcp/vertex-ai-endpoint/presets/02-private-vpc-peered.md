---
title: "Preset: Private VPC-Peered Endpoint"
description: "**Rank**: 2"
type: "preset"
rank: "02"
presetSlug: "02-private-vpc-peered"
componentSlug: "vertex-ai-endpoint"
componentTitle: "Vertex AI Endpoint"
provider: "gcp"
icon: "package"
order: 2
---

# Preset: Private VPC-Peered Endpoint

**Rank**: 2

## Use Case

A production-grade endpoint with VPC peering for network isolation, CMEK encryption for data protection, and dedicated DNS for performance. Suitable for sensitive workloads in regulated environments (HIPAA, PCI, SOC 2).

## What This Creates

- One Vertex AI Endpoint peered to your VPC network
- Customer-managed encryption via Cloud KMS
- Dedicated DNS endpoint for isolated traffic
- Accessible only from within the peered VPC

## Prerequisites

- VPC with Private Services Access configured
- Cloud KMS key ring and key in the same region as the endpoint
- IAM permissions for the Vertex AI service agent on the KMS key

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | Must match your KMS key region |
| `network` | placeholder | Your VPC's fully qualified path (uses project number, not ID) |
| `kmsKeyName` | placeholder | Your KMS key's fully qualified path |
| `dedicatedEndpointEnabled` | `true` | Set `false` if dedicated DNS is not needed |
