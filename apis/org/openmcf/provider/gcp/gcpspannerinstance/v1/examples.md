# GCP Spanner Instance Examples

This document provides YAML examples for deploying Cloud Spanner instances via OpenMCF. Each example includes a use-case description and the manifest.

---

## Example 1: Free Instance (Development/Testing)

**When to use:** Zero-cost development, prototyping, or CI/CD testing. Limited to ~10 GB storage and restricted throughput.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: dev-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpSpannerInstance.dev-spanner
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: dev-spanner
  config: regional-us-central1
  displayName: Dev Spanner Instance
  instanceType: FREE_INSTANCE
```

---

## Example 2: Regional Production with Nodes

**When to use:** Production workloads with predictable capacity needs. Single-region deployment for lower latency and cost.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: prod-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.prod-spanner
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-spanner
  config: regional-us-central1
  displayName: Production Spanner
  numNodes: 1
  edition: ENTERPRISE
  defaultBackupScheduleType: AUTOMATIC
```

---

## Example 3: Regional Production with Processing Units

**When to use:** Smaller production workloads where a full node is more capacity than needed. Processing units provide finer-grained sizing.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: prod-spanner-pu
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.prod-spanner-pu
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-spanner-pu
  config: regional-us-central1
  displayName: Production Spanner PU
  processingUnits: 500
  edition: STANDARD
  defaultBackupScheduleType: AUTOMATIC
```

---

## Example 4: Autoscaling Production

**When to use:** Variable workloads where traffic patterns are unpredictable. Spanner automatically adjusts capacity between min and max bounds based on CPU and storage utilization targets.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: prod-spanner-autoscale
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.prod-spanner-autoscale
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-spanner-as
  config: regional-us-central1
  displayName: Autoscaling Spanner
  edition: ENTERPRISE
  defaultBackupScheduleType: AUTOMATIC
  autoscalingConfig:
    autoscalingLimits:
      minNodes: 1
      maxNodes: 5
    autoscalingTargets:
      highPriorityCpuUtilizationPercent: 65
      storageUtilizationPercent: 80
```

---

## Example 5: Multi-Region with Enterprise Plus

**When to use:** Mission-critical globally distributed workloads requiring 99.999% availability SLA. Data is replicated across multiple geographic regions.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: global-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.global-spanner
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: global-spanner
  config: nam-eur-asia1
  displayName: Global Spanner Instance
  numNodes: 3
  edition: ENTERPRISE_PLUS
  defaultBackupScheduleType: AUTOMATIC
```

---

## Example 6: Full Production (Everything Together)

**When to use:** Maximum production configuration: autoscaling, Enterprise edition, automatic backups, force_destroy for controlled teardown.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: prod-spanner-full
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.prod-spanner-full
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-spanner-full
  config: regional-us-central1
  displayName: Full Production Spanner
  instanceType: PROVISIONED
  edition: ENTERPRISE
  defaultBackupScheduleType: AUTOMATIC
  forceDestroy: true
  autoscalingConfig:
    autoscalingLimits:
      minNodes: 1
      maxNodes: 10
    autoscalingTargets:
      highPriorityCpuUtilizationPercent: 65
      storageUtilizationPercent: 80
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).
