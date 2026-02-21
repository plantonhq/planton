# OciLogGroup Examples

## Custom Log for Application Ingestion

A log group with a single custom log for application-level log ingestion:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLogGroup
metadata:
  name: app-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciLogGroup.app-logs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  logs:
    - displayName: "application"
      logType: custom
      retentionDuration: 60
```

## VCN Flow Logs

Service logs auto-collecting VCN flow data from a subnet:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLogGroup
metadata:
  name: network-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciLogGroup.network-logs
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  description: "VCN flow logs for network traffic analysis"
  logs:
    - displayName: "vcn-flow-log"
      logType: service
      retentionDuration: 90
      configuration:
        service: "flowlogs"
        resource:
          valueFrom:
            kind: OciSubnet
            name: private-subnet
            fieldPath: status.outputs.subnetId
        category: "all"
```

## Mixed Service and Custom Logs

A log group with Object Storage write logs and a custom application log:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLogGroup
metadata:
  name: platform-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciLogGroup.platform-logs
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  description: "Platform observability logs"
  logs:
    - displayName: "bucket-writes"
      logType: service
      retentionDuration: 180
      configuration:
        service: "objectstorage"
        resource:
          valueFrom:
            kind: OciObjectStorageBucket
            name: data-bucket
            fieldPath: status.outputs.bucketId
        category: "write"
    - displayName: "app-audit"
      logType: custom
      retentionDuration: 180
```

## API Gateway Access Logs

Service logs collecting access patterns from an API Gateway:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLogGroup
metadata:
  name: api-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciLogGroup.api-logs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  logs:
    - displayName: "gateway-access"
      logType: service
      configuration:
        service: "apigateway"
        resource:
          value: "ocid1.apigateway.oc1..example"
        category: "access"
```

## Common Operations

### Add a new log to the group

Append a new entry to the `logs` list and re-apply. Existing logs are not affected.

### Increase retention

Update `retentionDuration` on the relevant log entry and re-apply. Retention changes do not force recreation.

### Disable a log

Set `isEnabled: false` on the log entry and re-apply. The log stops collecting data but retains existing entries until the retention period expires.

### Change log type

Changing `logType` from `custom` to `service` (or vice versa) forces log recreation. The old log is destroyed and a new one is created.

## Best Practices

1. **Group related logs together** — use one log group per application or service domain for organizational clarity.
2. **Use descriptive display names** — log display names are used as IaC resource keys and appear in the OCI Console.
3. **Set retention based on compliance needs** — 30 days for development, 90-180 days for production audit trails.
4. **Use `valueFrom` for service log resources** — references OCI resources by output, maintaining dependency ordering.
5. **Prefer service logs over custom ingestion** — service logs are zero-code, auto-collected, and maintained by OCI. Use custom logs only for application-level data.
