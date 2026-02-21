---
title: "Functions Application"
description: "Functions Application deployment documentation"
icon: "package"
order: 100
componentName: "ocifunctionsapplication"
---

# OCI Functions Application

Deploys an Oracle Cloud Infrastructure Functions application — the organizational container for serverless functions. Configures the shared execution environment including subnet placement, processor architecture (x86, ARM, or multi-arch), application-level environment variables, optional network security groups, image signature verification, and APM tracing.

## What Gets Created

When you deploy an OciFunctionsApplication resource, OpenMCF provisions:

- **Functions Application** — a `functions.Application` resource in the specified compartment and subnets with configurable processor shape, application config (environment variables), optional NSG bindings, optional image signature verification via KMS keys, and optional APM tracing integration.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the application will be created — either a literal value or a reference to an OciCompartment resource
- **At least one subnet OCID** — subnets where functions will execute, either as literal values or via `valueFrom` referencing OciSubnet resources
- **KMS key OCIDs** (for image verification only) — if enabling image signature verification
- **An APM domain OCID** (for tracing only) — if integrating with OCI Application Performance Monitoring

## Quick Start

Create a file `functions-app.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: my-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciFunctionsApplication.my-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetIds:
    - value: "ocid1.subnet.oc1..example"
```

Deploy:

```shell
openmcf apply -f functions-app.yaml
```

This creates a Functions application with GENERIC_X86 architecture in the specified subnet. The application OCID is exported as a stack output. Individual functions are deployed separately via `fn deploy` or CI/CD pipelines.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the application will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `subnetIds` | `StringValueOrRef[]` | OCIDs of the subnets where functions execute. Functions can reach resources accessible from these subnets. Immutable after creation. Can reference OciSubnet resources via `valueFrom`. | Min 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | metadata name | Display name for the application. Must be unique within the compartment. Immutable after creation. |
| `shape` | `enum` | `generic_x86` | Processor architecture. Values: `generic_x86` (Intel/AMD x86-64), `generic_arm` (Ampere A1), `generic_x86_arm` (multi-architecture). Immutable after creation. |
| `config` | `map<string, string>` | — | Application configuration passed as environment variables to all functions. Keys: ASCII letters, digits, underscores (cannot start with a digit). Max total size: 4 KB. |
| `networkSecurityGroupIds` | `StringValueOrRef[]` | — | OCIDs of network security groups applied to the application. Can reference OciSecurityGroup resources via `valueFrom`. |
| `syslogUrl` | `string` | — | Syslog URL for function logs (e.g., `"tcp://logserver.example.com:514"`). Must be reachable from the configured subnets. |
| `imagePolicyConfig` | `ImagePolicyConfig` | — | Image signature verification policy. See below. |
| `traceConfig` | `TraceConfig` | — | APM tracing configuration. See below. |

### ImagePolicyConfig

| Field | Type | Description |
|-------|------|-------------|
| `isPolicyEnabled` | `bool` | Whether image signature verification is enabled. |
| `keyDetails` | `ImagePolicyKeyDetail[]` | KMS keys used to verify image signatures. Required when `isPolicyEnabled` is `true`. |

### ImagePolicyKeyDetail

| Field | Type | Description |
|-------|------|-------------|
| `kmsKeyId` | `StringValueOrRef` | OCID of the KMS key for image signature verification. Can reference an OciKmsKey resource via `valueFrom`. |

### TraceConfig

| Field | Type | Description |
|-------|------|-------------|
| `isEnabled` | `bool` | Whether tracing is enabled. |
| `domainId` | `string` | OCID of the APM domain (collector) where trace events are sent. |

## Examples

### Minimal Application

An application with default x86 architecture in a single subnet:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: my-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciFunctionsApplication.my-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetIds:
    - value: "ocid1.subnet.oc1..example"
```

### ARM Architecture with Environment Config

An application running on Ampere A1 processors with shared environment variables and NSG binding:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: arm-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFunctionsApplication.arm-app
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  subnetIds:
    - valueFrom:
        kind: OciSubnet
        name: private-subnet
        fieldPath: status.outputs.subnetId
  shape: generic_arm
  config:
    LOG_LEVEL: "info"
    DB_ENDPOINT: "adb.us-ashburn-1.oraclecloud.com"
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: fn-nsg
        fieldPath: status.outputs.networkSecurityGroupId
```

### Image Signature Verification

An application with image signature verification — only images signed by the specified KMS key can be deployed:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: secure-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFunctionsApplication.secure-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetIds:
    - value: "ocid1.subnet.oc1..example"
  imagePolicyConfig:
    isPolicyEnabled: true
    keyDetails:
      - kmsKeyId:
          valueFrom:
            kind: OciKmsKey
            name: image-signing-key
            fieldPath: status.outputs.keyId
```

### Full-Featured with APM Tracing

An application with multi-architecture support, syslog forwarding, and APM distributed tracing:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: traced-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFunctionsApplication.traced-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetIds:
    - value: "ocid1.subnet.oc1..example-1"
    - value: "ocid1.subnet.oc1..example-2"
  shape: generic_x86_arm
  config:
    APP_ENV: "production"
  syslogUrl: "tcp://logserver.example.com:514"
  traceConfig:
    isEnabled: true
    domainId: "ocid1.apmdomain.oc1..example"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `application_id` | `string` | OCID of the functions application |

## Related Components

- [OciSubnet](/docs/catalog/oci/subnet) — provides the subnets referenced by `subnetIds` via `valueFrom`
- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — provides NSGs referenced by `networkSecurityGroupIds` via `valueFrom`
- [OciKmsKey](/docs/catalog/oci/kms-key) — provides signing keys for image verification via `valueFrom`
- [OciApiGateway](/docs/catalog/oci/api-gateway) — exposes functions via HTTP endpoints using the `oracle_functions` backend type
