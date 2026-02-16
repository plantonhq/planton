# GCP KMS Key Ring Examples

This document provides comprehensive examples for creating GCP Cloud KMS key rings using OpenMCF. Each example includes the manifest YAML and explains the use case and key configuration choices.

## Table of Contents

- [Example 1: Regional Key Ring](#example-1-regional-key-ring)
- [Example 2: Global Key Ring](#example-2-global-key-ring)
- [Example 3: Multi-Region Key Ring](#example-3-multi-region-key-ring)
- [Example 4: Using Project Reference](#example-4-using-project-reference)
- [Presets](#presets)

---

## Example 1: Regional Key Ring

### Use Case

Create a key ring in a specific GCP region for workloads that have data residency requirements or need lowest-latency access to encryption keys. This is the most common pattern — co-locate the key ring with the resources it protects (e.g., Cloud SQL, BigQuery, GKE in the same region).

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: prod-us-central1-keys
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.prod-us-central1-keys
spec:
  projectId:
    value: my-prod-project-123
  keyRingName: prod-encryption
  location: us-central1
```

Deploy:

```shell
openmcf apply -f prod-key-ring.yaml
```

### Key Choices

- **Regional location** (`us-central1`) — keys are stored in a single region. Best for workloads where data must not leave a specific geography.
- **Descriptive name** — `prod-encryption` clearly identifies the key ring's purpose. Choose names carefully; they are permanent.

---

## Example 2: Global Key Ring

### Use Case

Create a key ring in the `global` location when your encryption keys need to be accessible from any GCP region without latency concerns. Good for symmetric encryption keys used by globally distributed workloads.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: global-shared-keys
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.global-shared-keys
spec:
  projectId:
    value: my-security-project-456
  keyRingName: global-shared-keys
  location: global
```

### Key Choices

- **Global location** — keys accessible from all regions. No data residency guarantees.
- **Dedicated security project** — common pattern is to keep encryption keys in a separate GCP project with restricted access, away from the workload projects.

---

## Example 3: Multi-Region Key Ring

### Use Case

Create a key ring in a multi-region location (e.g., `us`, `europe`, `asia`) for high availability across an entire continent while maintaining data residency within that geography.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: eu-compliance-keys
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.eu-compliance-keys
spec:
  projectId:
    value: my-eu-project-789
  keyRingName: eu-compliance-keys
  location: europe
```

### Key Choices

- **Multi-region `europe`** — keys replicated across European regions for high availability, while guaranteeing data stays within EU boundaries. Ideal for GDPR-compliant workloads.
- **Compliance naming** — name reflects the regulatory context for easy identification.

---

## Example 4: Using Project Reference

### Use Case

Reference a project ID from a GcpProject resource instead of hardcoding. This enables infra-chart composition where the project and key ring are provisioned together in a dependency-aware pipeline.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: data-encryption-keys
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.data-encryption-keys
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: data-project
      fieldPath: status.outputs.project_id
  keyRingName: data-encryption
  location: us-central1
```

### Key Choices

- **`valueFrom` reference** — the project ID is resolved at deployment time from the GcpProject resource's outputs. This creates a dependency edge: the key ring deploys after the project.
- **Infra-chart ready** — this pattern is exactly what infra chart templates use to wire resources together.

---

## Presets

OpenMCF provides ready-to-use presets for common key ring configurations. See the `presets/` directory:

| Preset | Location | Use Case |
|--------|----------|----------|
| `01-regional-key-ring` | Regional (`us-central1`) | Standard workloads with regional data residency |
| `02-global-key-ring` | `global` | Cross-region workloads without data residency constraints |
| `03-multi-region-key-ring` | Multi-region (`us`) | High availability within a continental geography |
