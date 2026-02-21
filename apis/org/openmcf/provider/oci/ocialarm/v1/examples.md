# OciAlarm Examples

## CPU Utilization Alarm

A basic alarm monitoring compute instance CPU utilization with a 5-minute pending duration:

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

## Multi-Threshold Tiered Alerting

An alarm with warning at 70% and critical at 90% using overrides:

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
      body: "CPU above 70% — investigate before reaching critical"
      pendingDuration: "PT10M"
```

## Database Storage with Repeat Notifications

An alarm for Autonomous Database storage with hourly re-notifications:

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

## Per-Dimension VCN Traffic Alarm

An alarm with per-resource notifications across sub-compartments:

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
    value: "ocid1.tenancy.oc1..example"
  namespace: "oci_vcn"
  query: "VnicEgressDropsSecurityList[5m].sum() > 100"
  severity: error
  destinations:
    - "ocid1.onstopic.oc1..example"
  isEnabled: true
  isNotificationsPerMetricDimensionEnabled: true
  metricCompartmentIdInSubtree: true
```

## Common Operations

### Enable or disable an alarm

Set `isEnabled` to `true` or `false` and re-apply. Disabled alarms do not evaluate metrics or send notifications.

### Adjust the threshold

Modify the MQL `query` expression (e.g., change `> 80` to `> 90`) and re-apply. Query changes are updatable without recreation.

### Change notification destinations

Update the `destinations` list and re-apply. Add or remove ONS topic or Streaming stream OCIDs.

### Add a tiered override

Add an entry to the `overrides` list with a unique `ruleName` and re-apply. Overrides are evaluated in list order before the base rule.

### Reduce notification noise

Set `repeatNotificationDuration` to a longer interval (e.g., `PT4H`) to reduce re-notification frequency while the alarm remains FIRING.

## Best Practices

1. **Set `isEnabled: true` explicitly** — proto3 defaults booleans to false, so alarms start disabled unless you set this.
2. **Use meaningful `pendingDuration` values** — `PT1M` (minimum) fires quickly but may cause false positives. `PT5M` to `PT15M` is typical for production.
3. **Use overrides for tiered alerting** — avoids creating multiple alarms for the same metric at different thresholds.
4. **Include dynamic variables in notification body** — `{{resourceId}}` and `{{timestamp}}` help on-call engineers identify the affected resource quickly.
5. **Use `valueFrom` references** for `compartmentId` and `metricCompartmentId` — avoids hardcoding OCIDs.
6. **Choose `messageFormat` based on destination** — `pretty_json` and `ons_optimized` only work with ONS topics, not Streaming.
