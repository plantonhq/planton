---
title: "KMS Key Ring"
description: "Deploy GCP Cloud KMS key rings using OpenMCF"
---

# GcpKmsKeyRing

Provision and manage Cloud KMS key rings -- organizational containers for
cryptographic keys in GCP.

## Overview

A key ring is a permanent grouping of cryptographic keys in Cloud KMS. Key rings
belong to a GCP project and reside in a specific location (region, multi-region,
or "global"). Once created, a key ring cannot be deleted.

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: my-key-ring
spec:
  projectId:
    value: "my-gcp-project"
  keyRingName: prod-encryption
  location: us-central1
```

## Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | Yes | GCP project ID |
| `keyRingName` | string | Yes | Key ring name (1-63 chars) |
| `location` | string | Yes | Region, multi-region, or "global" |

## Outputs

| Output | Description |
|--------|-------------|
| `key_ring_id` | Fully qualified path for key references |
| `key_ring_name` | Short name of the key ring |

## Important

**Key rings cannot be deleted from GCP.** On destroy, the resource is removed
from state but remains permanently in GCP.

## Related

- [GcpKmsKey](/docs/catalog/gcp/kms-key) -- Cryptographic keys (child resource)
