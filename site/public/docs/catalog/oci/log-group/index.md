---
title: "Log Group"
description: "Log Group deployment documentation"
icon: "package"
order: 100
componentName: "ociloggroup"
---

# OCI Log Group

Deploys an Oracle Cloud Infrastructure Log Group with bundled logs. The log group is the organizational container for OCI Logging service logs. Logs can be either service logs (auto-collected from OCI services like Object Storage, API Gateway, Functions) or custom logs (ingested via the Logging Ingestion API).

## What Gets Created

When you deploy an OciLogGroup resource, OpenMCF provisions:

- **Log Group** — a `logging.LogGroup` resource in the specified compartment with an optional description.
- **Logs** — one `logging.Log` per entry in the `logs` list. Each log is created within the group with a specified type (custom or service), optional enable flag, optional retention duration, and optional service log source configuration. Logs depend on the group for creation ordering.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the log group will be created — either a literal value or a reference to an OciCompartment resource
- **Source resource OCIDs** (for service logs) — the OCID of the OCI resource emitting logs (e.g., a VCN for flow logs, a bucket for Object Storage logs)

## Quick Start

Create a file `log-group.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLogGroup
metadata:
  name: my-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciLogGroup.my-logs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  logs:
    - displayName: "app-log"
      logType: custom
```

Deploy:

```shell
openmcf apply -f log-group.yaml
```

This creates a log group with one custom log that accepts entries via the Logging Ingestion API. The log group OCID is exported as a stack output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the log group will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Description for the log group. |
| `logs` | `Log[]` | — | Logs within this group. Each log is identified by its `displayName`, which must be unique within the group. |

### Log

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | — | Display name for the log. Used as the IaC resource key. Must be unique within the group. Required. |
| `logType` | `enum` | — | Type of log: `custom` (entries pushed via Ingestion API) or `service` (auto-collected from OCI services). Required. |
| `isEnabled` | `bool` | `true` | Whether the log is enabled. |
| `retentionDuration` | `int32` | `30` | Retention period in days. Must be a 30-day increment: 30, 60, 90, 120, 150, or 180. |
| `configuration` | `ServiceLogConfiguration` | — | Source configuration for service logs. Required when `logType` is `service`; ignored for custom logs. |

### ServiceLogConfiguration

| Field | Type | Description |
|-------|------|-------------|
| `service` | `string` | OCI service generating the log. Examples: `"objectstorage"`, `"flowlogs"`, `"apigateway"`, `"loadbalancer"`, `"functionsInvoke"`. Required. |
| `resource` | `StringValueOrRef` | OCID of the resource emitting logs (e.g., a VCN, bucket, API gateway). Can reference any OpenMCF component via `valueFrom`. Required. |
| `category` | `string` | Log category within the service. Examples: `"write"`, `"read"`, `"all"` (Object Storage); `"access"` (API Gateway); `"invoke"` (Functions). Required. |
| `parameters` | `map<string, string>` | Additional parameters for the log source. Pass-through to OCI. |
| `compartmentId` | `StringValueOrRef` | Optional compartment override for source resource lookup. When omitted, the log group's compartment is used. Can reference an OciCompartment via `valueFrom`. |

## Examples

### Custom Log for Application Ingestion

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

### VCN Flow Logs

A log group collecting VCN flow logs from a subnet:

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

### Mixed Service and Custom Logs

A log group with both Object Storage write logs and a custom application log:

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

### API Gateway Access Logs

A log group collecting access logs from an API Gateway deployment:

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

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `log_group_id` | `string` | OCID of the log group |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciVcn](/docs/catalog/oci/vcn) — VCNs and subnets are common sources for flow logs
- [OciObjectStorageBucket](/docs/catalog/oci/object-storage-bucket) — buckets are common sources for Object Storage service logs
- [OciApiGateway](/docs/catalog/oci/api-gateway) — API gateways emit access and execution logs
- [OciFunctionsApplication](/docs/catalog/oci/functions-application) — functions emit invocation logs
