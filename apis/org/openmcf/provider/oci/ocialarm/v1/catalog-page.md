# OCI Alarm

Deploys an Oracle Cloud Infrastructure Monitoring Alarm — a rule that evaluates metrics via Monitoring Query Language (MQL) expressions and triggers notifications to ONS topics or Streaming endpoints when thresholds are breached. Supports multi-threshold evaluation via overrides, configurable severity levels, notification formatting, and per-metric-dimension alerting.

## What Gets Created

When you deploy an OciAlarm resource, OpenMCF provisions:

- **Monitoring Alarm** — a `monitoring.Alarm` resource in the specified compartment with an MQL query, severity, notification destinations, optional overrides for multi-threshold evaluation, configurable pending and evaluation slack durations, and optional notification formatting.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the alarm will be created — either a literal value or a reference to an OciCompartment resource
- **A metric compartment OCID** — the compartment containing the metric being evaluated (often the same compartment)
- **At least one notification destination** — OCID of an ONS Notification Topic or a Streaming stream

## Quick Start

Create a file `alarm.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAlarm
metadata:
  name: high-cpu
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciAlarm.high-cpu
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  metricCompartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "oci_computeagent"
  query: "CpuUtilization[5m].mean() > 80"
  severity: critical
  destinations:
    - "ocid1.onstopic.oc1..example"
  isEnabled: true
```

Deploy:

```shell
openmcf apply -f alarm.yaml
```

This creates an alarm that fires when average CPU utilization exceeds 80% over a 5-minute window, sending a notification to the specified ONS topic. The alarm OCID is exported as a stack output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the alarm will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `metricCompartmentId` | `StringValueOrRef` | OCID of the compartment containing the metric being evaluated. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `namespace` | `string` | Source service or application emitting the metric. Examples: `"oci_computeagent"`, `"oci_blockstore"`, `"oci_autonomous_database"`, `"oci_vcn"`, `"oci_queue"`. | Min length 1 |
| `query` | `string` | MQL expression to evaluate. Must specify metric, statistic, interval, and trigger rule. Example: `"CpuUtilization[5m].mean() > 80"`. | Min length 1 |
| `severity` | `enum` | Perceived severity when the alarm is FIRING. Values: `critical`, `error`, `warning`, `info`. | Must be explicitly set |
| `destinations` | `string[]` | OCIDs of notification destinations. Each OCID must reference an ONS topic or Streaming stream. | Min 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isEnabled` | `bool` | `false` | Whether the alarm evaluates metrics and sends notifications. |
| `body` | `string` | — | Notification body content. Supports dynamic variables: `{{severity}}`, `{{query}}`, `{{metricValues}}`, `{{resourceId}}`, `{{timestamp}}`. |
| `alarmSummary` | `string` | — | Custom alarm summary for API responses and notification bodies. Supports dynamic variables. |
| `notificationTitle` | `string` | — | Notification title (email subject line, Slack title). Supports dynamic variables. |
| `pendingDuration` | `string` | `PT1M` | Time the condition must persist before transitioning from OK to FIRING. ISO 8601 duration. Min: PT1M, Max: PT1H. |
| `evaluationSlackDuration` | `string` | `PT3M` | Slack period for metric ingestion before evaluation. ISO 8601 duration. Min: PT3M, Max: PT2H. |
| `repeatNotificationDuration` | `string` | — | Frequency for re-submitting notifications while FIRING. ISO 8601 duration. Min: PT1M, Max: P30D. When omitted, no re-submissions. |
| `messageFormat` | `enum` | `raw` | Notification format. Values: `raw` (default, works with all destinations), `pretty_json` (human-readable JSON, Notifications only), `ons_optimized` (compact for email, Notifications only). |
| `metricCompartmentIdInSubtree` | `bool` | — | When true, evaluates metrics from the compartment and all sub-compartments. Only valid when `metricCompartmentId` is a tenancy OCID. |
| `isNotificationsPerMetricDimensionEnabled` | `bool` | `false` | When true, sends separate notifications per metric stream dimension. |
| `resourceGroup` | `string` | — | Resource group to match when filtering metric data. |
| `notificationVersion` | `string` | — | Alarm notification version. Format: a number followed by `.X` (e.g., `"1.X"`). |
| `ruleName` | `string` | `"BASE"` | Identifier for the alarm's base evaluation rule. Must be unique across all rule names (including overrides). |
| `overrides` | `AlarmOverride[]` | — | Alternative evaluation parameters for multi-threshold alerting. Evaluated in order before the base rule. |

### AlarmOverride

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ruleName` | `string` | Unique identifier for this override. | Min length 1 |
| `query` | `string` | Override MQL query. When omitted, the base query is used. | — |
| `severity` | `enum` | Override severity. When unspecified, the base severity is used. | — |
| `body` | `string` | Override notification body. Supports dynamic variables. | — |
| `pendingDuration` | `string` | Override pending duration. ISO 8601 format. | — |

## Examples

### CPU Utilization Alarm

A basic alarm monitoring compute instance CPU utilization:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAlarm
metadata:
  name: high-cpu
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciAlarm.high-cpu
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  metricCompartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "oci_computeagent"
  query: "CpuUtilization[5m].mean() > 80"
  severity: critical
  destinations:
    - "ocid1.onstopic.oc1..example"
  isEnabled: true
  pendingDuration: "PT5M"
  body: "CPU utilization exceeded 80% on {{resourceId}} at {{timestamp}}"
```

### Multi-Threshold Alarm with Overrides

An alarm with warning at 70% and critical at 90%, using overrides for tiered alerting:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAlarm
metadata:
  name: tiered-cpu
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciAlarm.tiered-cpu
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  metricCompartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  namespace: "oci_computeagent"
  query: "CpuUtilization[5m].mean() > 90"
  severity: critical
  destinations:
    - "ocid1.onstopic.oc1..example"
  isEnabled: true
  pendingDuration: "PT5M"
  ruleName: "critical-rule"
  overrides:
    - ruleName: "warning-rule"
      query: "CpuUtilization[5m].mean() > 70"
      severity: warning
      body: "CPU utilization above 70% — investigate before it reaches critical"
      pendingDuration: "PT10M"
```

### Autonomous Database Storage Alarm

An alarm monitoring Autonomous Database storage utilization:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAlarm
metadata:
  name: adb-storage
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciAlarm.adb-storage
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  metricCompartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "oci_autonomous_database"
  query: "StorageUtilization[1h].max() > 85"
  severity: warning
  destinations:
    - "ocid1.onstopic.oc1..example"
  isEnabled: true
  pendingDuration: "PT15M"
  repeatNotificationDuration: "PT1H"
  notificationTitle: "ADB Storage Alert: {{severity}}"
```

### Per-Dimension VCN Traffic Alarm

An alarm that sends separate notifications per VCN resource when traffic spikes:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAlarm
metadata:
  name: vcn-traffic-spike
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciAlarm.vcn-traffic-spike
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  metricCompartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "oci_vcn"
  query: "VnicEgressDropsSecurityList[5m].sum() > 100"
  severity: error
  destinations:
    - "ocid1.onstopic.oc1..example"
  isEnabled: true
  isNotificationsPerMetricDimensionEnabled: true
  metricCompartmentIdInSubtree: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `alarm_id` | `string` | OCID of the alarm |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides compartments referenced by `compartmentId` and `metricCompartmentId` via `valueFrom`
- [OciQueue](/docs/catalog/oci/ociqueue) — queue metrics (depth, message age) are common alarm targets
- [OciAutonomousDatabase](/docs/catalog/oci/ociautonomousdatabase) — database metrics are common alarm targets
- [OciComputeInstance](/docs/catalog/oci/ocicomputeinstance) — compute metrics (CPU, memory, disk) are common alarm targets
