---
title: "Container Instance"
description: "Container Instance deployment documentation"
icon: "package"
order: 100
componentName: "ocicontainerinstance"
---

# OCI Container Instance

Deploys an Oracle Cloud Infrastructure Container Instance — OCI's serverless container service for running one or more containers in a pod-like construct without managing compute infrastructure. Containers within an instance share networking (VNICs) and volumes, communicate over localhost, and support HTTP/TCP health checks, Linux security contexts, and image pull secrets for private registries. Shapes are always flex (CI.Standard.E4.Flex, CI.Standard.E3.Flex), with configurable OCPU and memory allocation.

## What Gets Created

When you deploy an OciContainerInstance resource, OpenMCF provisions:

- **Container Instance** — an `oci_container_instances_container_instance` resource in the specified compartment and availability domain. The instance runs one or more containers sharing the same network namespace and volume mounts, using the specified flex shape and OCPU/memory allocation. Each container can have independent health checks, resource limits, security contexts, and volume mounts. Standard OpenMCF freeform tags are applied for resource tracking.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the container instance will be created — literal value or reference to an OciCompartment resource
- **An availability domain name** (e.g., `Uocm:PHX-AD-1`) — run `oci iam availability-domain list` to see domains in your region
- **A subnet OCID** for the instance's VNIC — literal value or reference to an OciSubnet resource
- **A container image URL** accessible from the subnet (e.g., `docker.io/library/nginx:latest`). For private registries, configure `imagePullSecrets`

## Quick Start

Create a file `container-instance.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: web-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciContainerInstance.web-server
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
  containers:
    - imageUrl: "docker.io/library/nginx:latest"
  vnics:
    - subnetId:
        value: "ocid1.subnet.oc1.iad.example"
```

Deploy:

```shell
openmcf apply -f container-instance.yaml
```

This creates a single-container instance running nginx on a 1-OCPU E4 Flex shape. OCI assigns memory based on the OCPU count (minimum 1 GB per OCPU). The container uses the default restart policy (ALWAYS) and inherits DNS settings from the subnet's DHCP options. The container instance ID and container IDs are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the container instance will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `availabilityDomain` | `string` | Availability domain where the container instance runs. Example: `Uocm:PHX-AD-1`. | Minimum 1 character |
| `shape` | `string` | Compute shape for the container instance. Container Instance shapes are always flex. Example: `CI.Standard.E4.Flex`, `CI.Standard.E3.Flex`. | Minimum 1 character |
| `shapeConfig` | `ShapeConfig` | CPU and memory allocation for the entire container instance. Individual containers can set resource limits within this envelope. See [shapeConfig fields](#shapeconfig-fields). | Required |
| `containers` | `Container[]` | Containers to run on this instance. Multiple containers share the same network namespace and can communicate over localhost. See [container fields](#container-fields). | Minimum 1 item |
| `vnics` | `Vnic[]` | Virtual network interface cards providing network connectivity. Each VNIC is attached to a subnet. All containers share the instance's VNICs. See [vnic fields](#vnic-fields). | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name for the container instance shown in the OCI Console. |
| `containerRestartPolicy` | `enum` | `always` | Restart policy applied to all containers. Values: `always`, `never`, `on_failure`. |
| `faultDomain` | `string` | Auto-selected | Fault domain within the availability domain. Example: `FAULT-DOMAIN-1`. When omitted, OCI selects a fault domain automatically. |
| `gracefulShutdownTimeoutInSeconds` | `int64` | — | Seconds to wait for containers to gracefully terminate before forcefully stopping them. Applies when the instance is stopped or deleted. |
| `dnsConfig` | `DnsConfig` | Subnet DHCP | DNS resolver configuration for containers. When omitted, containers inherit DNS settings from the subnet's DHCP options. See [dnsConfig fields](#dnsconfig-fields). |
| `imagePullSecrets` | `ImagePullSecret[]` | — | Credentials for pulling container images from private registries. Supports basic authentication and OCI Vault-based credentials. See [imagePullSecret fields](#imagepullsecret-fields). |
| `volumes` | `Volume[]` | — | Volumes accessible to containers via volume mounts. A container instance supports up to 32 volumes. See [volume fields](#volume-fields). |

### shapeConfig Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ocpus` | `float` | Number of OCPUs allocated to the container instance. Example: `1.0`, `2.0`, `4.0`. | Greater than 0 |
| `memoryInGbs` | `float` | Memory in gigabytes. When omitted, OCI assigns a default based on the OCPU count (minimum 1 GB per OCPU). | Optional |

### container Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `imageUrl` | `string` | Container image URL. Example: `docker.io/library/nginx:latest`, `ghcr.io/org/app:v1.2`. Default registry is `docker.io/library` if not specified. | Minimum 1 character |
| `displayName` | `string` | Human-readable name for the container. | Optional |
| `command` | `string[]` | Overrides the image's ENTRYPOINT. Each element is a separate argument. | Optional |
| `arguments` | `string[]` | Arguments passed to the ENTRYPOINT process. Total size of all arguments combined must be <= 64 KB. | Optional |
| `environmentVariables` | `map<string, string>` | Environment variables injected into the container. Total size of all names + values combined must be <= 64 KB. | Optional |
| `workingDirectory` | `string` | Working directory for the container's entrypoint process. | Optional |
| `isResourcePrincipalDisabled` | `bool` | When true, disables OCI resource principal access for this container. Resource principal (v2.2) is enabled by default. | Optional |
| `resourceConfig` | `ContainerResourceConfig` | CPU and memory limits for this container within the instance-level envelope. When omitted, the container can use all resources available to the instance. See [containerResourceConfig fields](#containerresourceconfig-fields). | Optional |
| `healthChecks` | `HealthCheck[]` | Health checks for monitoring container readiness. Supports HTTP and TCP probe types. See [healthCheck fields](#healthcheck-fields). | Optional |
| `securityContext` | `SecurityContext` | Linux security settings for the container process. See [securityContext fields](#securitycontext-fields). | Optional |
| `volumeMounts` | `VolumeMount[]` | Volumes to mount into this container's filesystem. See [volumeMount fields](#volumemount-fields). | Optional |

### containerResourceConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `memoryLimitInGbs` | `float` | Maximum memory in gigabytes the container can consume. When omitted, the container can use all available instance memory. |
| `vcpusLimit` | `float` | Maximum logical CPUs the container can consume. 1 OCPU = 2 logical CPUs. Values can be fractional (e.g., `0.5`). When omitted, the container can use all available instance CPUs. |

### healthCheck Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `healthCheckType` | `enum` | Protocol for the health check. Values: `http`, `tcp`. | Required (cannot be unspecified) |
| `port` | `int32` | Port to probe. | Greater than 0 |
| `name` | `string` | Name for the health check, unique within the container instance. | Optional |
| `path` | `string` | URL path for HTTP health checks. Required when `healthCheckType` is `http`. Example: `/healthz`. | Optional |
| `failureAction` | `enum` | Action to take when the health check fails. Values: `kill`, `none`. Default: `kill`. | Optional |
| `failureThreshold` | `int32` | Consecutive failures required to consider the container unhealthy. | Optional |
| `successThreshold` | `int32` | Consecutive successes required to consider the container healthy again. | Optional |
| `initialDelayInSeconds` | `int32` | Seconds to wait after container start before running the first check. | Optional |
| `intervalInSeconds` | `int32` | Seconds between consecutive health checks. | Optional |
| `timeoutInSeconds` | `int32` | Seconds to wait for a health check response before considering it failed. | Optional |
| `headers` | `HealthCheckHeader[]` | Custom HTTP headers sent with HTTP health checks. Each header has `name` and `value` string fields. | Optional |

### securityContext Fields

| Field | Type | Description |
|-------|------|-------------|
| `isNonRootUserCheckEnabled` | `bool` | When true, validates at runtime that the container does not run as UID 0. Fails the container start if the image runs as root. |
| `isRootFileSystemReadonly` | `bool` | When true, the container's root filesystem is mounted read-only. |
| `runAsUser` | `int32` | User ID (UID) for the container's entrypoint process. Defaults to the UID specified in the container image. |
| `runAsGroup` | `int32` | Group ID (GID) for the container's entrypoint process. When specified, `runAsUser` should also be provided. |
| `capabilities` | `Capabilities` | Linux capabilities to add or drop from the container process. See [capabilities fields](#capabilities-fields). |

### capabilities Fields

| Field | Type | Description |
|-------|------|-------------|
| `addCapabilities` | `string[]` | Capabilities to add to the container process. Example: `["NET_ADMIN", "SYS_TIME"]`. |
| `dropCapabilities` | `string[]` | Capabilities to drop from the container process. Example: `["ALL"]`. |

### volumeMount Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `mountPath` | `string` | Path inside the container where the volume is mounted. Example: `/data`, `/etc/config`. | Minimum 1 character |
| `volumeName` | `string` | Name of the volume to mount. Must match a volume defined in the instance-level `volumes` list. | Minimum 1 character |
| `isReadOnly` | `bool` | When true, the volume is mounted read-only. | Optional |
| `partition` | `int32` | If the volume has partitions, the partition number to mount. | Optional |
| `subPath` | `string` | Sub-path within the volume to mount instead of the volume root. | Optional |

### vnic Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `subnetId` | `StringValueOrRef` | OCID of the subnet in which to create the VNIC. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `displayName` | `string` | Human-readable name for the VNIC. | Optional |
| `hostnameLabel` | `string` | Hostname label for the VNIC's primary private IP in subnet DNS. | Optional |
| `isPublicIpAssigned` | `bool` | Whether to assign a public IP to the VNIC. When omitted, uses the subnet's default public IP assignment setting. | Optional |
| `nsgIds` | `StringValueOrRef[]` | OCIDs of network security groups to add this VNIC to. Can reference OciSecurityGroup resources via `valueFrom`. | Optional |
| `privateIp` | `string` | Static private IP address within the subnet's CIDR. When omitted, OCI assigns one automatically. | Optional |
| `skipSourceDestCheck` | `bool` | When true, disables source/destination checking on the VNIC. Required for NAT instances or virtual routers. | Optional |

### dnsConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `nameservers` | `string[]` | IP addresses of DNS name servers (IPv4 or IPv6). When omitted, uses nameservers from the subnet's DHCP options. |
| `options` | `string[]` | Resolver options in resolv.conf format. Example: `["ndots:5", "edns0"]`. |
| `searches` | `string[]` | Search domains for unqualified hostname lookups. When omitted, uses searches from the subnet's DHCP options. |

### imagePullSecret Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `registryEndpoint` | `string` | Registry endpoint URL. Example: `ghcr.io`, `docker.io`, `us-ashburn-1.ocir.io`. | Minimum 1 character |
| `secretType` | `enum` | Authentication method for the registry. Values: `basic`, `vault`. | Required (cannot be unspecified) |
| `username` | `string` | Username for basic authentication. Required when `secretType` is `basic`. Must be base64-encoded. | Optional |
| `password` | `string` | Password for basic authentication. Required when `secretType` is `basic`. Must be base64-encoded. | Optional |
| `secretId` | `StringValueOrRef` | OCID of an OCI Vault secret containing registry credentials. Required when `secretType` is `vault`. | Optional |

### volume Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name for this volume within the container instance. Containers reference this name in their `volumeMounts`. | Minimum 1 character |
| `volumeType` | `enum` | Storage backing for the volume. Values: `emptydir`, `configfile`. | Required (cannot be unspecified) |
| `backingStore` | `string` | Backing store for emptydir volumes. Options: `EPHEMERAL_STORAGE` (disk-backed) or `MEMORY` (tmpfs). Only applicable when `volumeType` is `emptydir`. | Optional |
| `configs` | `VolumeConfig[]` | Config file entries for configfile volumes. Each entry becomes a file in the volume. Only applicable when `volumeType` is `configfile`. See [volumeConfig fields](#volumeconfig-fields). | Optional |

### volumeConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `data` | `string` | Base64-encoded contents of the file. Decoded to plain text at mount time. |
| `fileName` | `string` | Name of the file within the volume. Must be unique across the volume. |
| `path` | `string` | Relative path within the volume mount directory. When omitted, the file is placed at the volume mount root. |

## Examples

### Minimal Single-Container Instance

A single nginx container — the simplest path to running a container on OCI:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: web-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciContainerInstance.web-server
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
  containers:
    - imageUrl: "docker.io/library/nginx:latest"
  vnics:
    - subnetId:
        value: "ocid1.subnet.oc1.iad.example"
```

### Multi-Container Instance with Shared Volume

A web application container with a Fluent Bit sidecar collecting logs via a shared emptydir volume — the pod-like multi-container pattern:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: web-with-logging
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciContainerInstance.web-with-logging
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 2
    memoryInGbs: 8
  containerRestartPolicy: always
  containers:
    - imageUrl: "ghcr.io/acme/web-app:v2.1"
      displayName: "web-app"
      resourceConfig:
        memoryLimitInGbs: 6
        vcpusLimit: 3
      environmentVariables:
        LOG_DIR: "/var/log/app"
        PORT: "8080"
      healthChecks:
        - healthCheckType: http
          port: 8080
          path: "/healthz"
          initialDelayInSeconds: 10
          intervalInSeconds: 15
          timeoutInSeconds: 5
          failureThreshold: 3
      volumeMounts:
        - mountPath: "/var/log/app"
          volumeName: "shared-logs"
    - imageUrl: "docker.io/fluent/fluent-bit:latest"
      displayName: "log-collector"
      resourceConfig:
        memoryLimitInGbs: 1
        vcpusLimit: 0.5
      volumeMounts:
        - mountPath: "/var/log/app"
          volumeName: "shared-logs"
          isReadOnly: true
  vnics:
    - subnetId:
        valueFrom:
          kind: OciSubnet
          name: app-subnet
          fieldPath: status.outputs.subnetId
      nsgIds:
        - valueFrom:
            kind: OciSecurityGroup
            name: app-nsg
            fieldPath: status.outputs.networkSecurityGroupId
  volumes:
    - name: "shared-logs"
      volumeType: emptydir
      backingStore: "EPHEMERAL_STORAGE"
```

### Hardened Production Instance with Private Registry

A production container with a read-only root filesystem, non-root user enforcement, dropped capabilities, health checks, config file injection, and OCI Vault-based image pull credentials:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: secure-api
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciContainerInstance.secure-api
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:PHX-AD-1"
  faultDomain: "FAULT-DOMAIN-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 4
    memoryInGbs: 16
  gracefulShutdownTimeoutInSeconds: 30
  containers:
    - imageUrl: "us-ashburn-1.ocir.io/acme-tenancy/api-service:v3.0"
      displayName: "api-service"
      command:
        - "/app/server"
      arguments:
        - "--config=/etc/app/config.json"
        - "--port=8443"
      environmentVariables:
        APP_ENV: "production"
        LOG_LEVEL: "info"
      resourceConfig:
        memoryLimitInGbs: 14
        vcpusLimit: 7
      securityContext:
        isNonRootUserCheckEnabled: true
        isRootFileSystemReadonly: true
        runAsUser: 1000
        runAsGroup: 1000
        capabilities:
          dropCapabilities:
            - "ALL"
          addCapabilities:
            - "NET_BIND_SERVICE"
      healthChecks:
        - healthCheckType: http
          port: 8443
          path: "/healthz"
          name: "readiness"
          initialDelayInSeconds: 15
          intervalInSeconds: 10
          timeoutInSeconds: 5
          failureThreshold: 3
          successThreshold: 1
        - healthCheckType: tcp
          port: 8443
          name: "liveness"
          intervalInSeconds: 30
          failureThreshold: 5
          failureAction: kill
      volumeMounts:
        - mountPath: "/etc/app"
          volumeName: "app-config"
          isReadOnly: true
        - mountPath: "/tmp"
          volumeName: "tmp-dir"
  vnics:
    - subnetId:
        valueFrom:
          kind: OciSubnet
          name: private-subnet
          fieldPath: status.outputs.subnetId
      isPublicIpAssigned: false
      nsgIds:
        - valueFrom:
            kind: OciSecurityGroup
            name: api-nsg
            fieldPath: status.outputs.networkSecurityGroupId
  imagePullSecrets:
    - registryEndpoint: "us-ashburn-1.ocir.io"
      secretType: vault
      secretId:
        value: "ocid1.vaultsecret.oc1.iad.example"
  dnsConfig:
    searches:
      - "internal.acme.com"
      - "svc.acme.com"
  volumes:
    - name: "app-config"
      volumeType: configfile
      configs:
        - fileName: "config.json"
          data: "eyJkYXRhYmFzZSI6eyJob3N0IjoiZGIuaW50ZXJuYWwifX0="
    - name: "tmp-dir"
      volumeType: emptydir
      backingStore: "MEMORY"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `container_instance_id` | `string` | OCID of the container instance. |
| `container_ids` | `string` | Comma-separated OCIDs of the individual containers within the instance. Useful for operational tasks (viewing logs, exec into container). |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/subnet) — provides subnets for VNIC attachment (`vnics[].subnetId`) via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — manages network security rules for instance VNICs (`vnics[].nsgIds`) via `valueFrom`
