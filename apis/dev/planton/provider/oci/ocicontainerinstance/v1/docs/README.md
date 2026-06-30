# OCI Container Instance: Design Rationale and Research

## Introduction

The OciContainerInstance component manages OCI's serverless container service — a pod-like construct where one or more containers share networking (VNICs) and volumes without requiring compute infrastructure management. The design challenge is exposing OCI Container Instances' full configuration surface (containers, volumes, health checks, security contexts, VNICs, image pull secrets, DNS) through a declarative KRM API that is both expressive enough for production use and approachable for a first deployment.

The spec surface (13 top-level fields, 12 nested messages, 5 enums) mirrors OCI's Container Instance API with minimal abstraction. This document explains the design decisions that shaped the component.

## Why This Is a Single-Resource Component

The OciContainerInstance component creates exactly one cloud resource: `oci_container_instances_container_instance`. Unlike some Planton components that bundle multiple tightly coupled resources (e.g., OciVcn bundles gateways, OciApplicationLoadBalancer bundles backend sets), the container instance stands alone because:

1. **Self-contained by design.** OCI Container Instances embed containers, VNICs, volumes, and health checks as nested configuration within a single API call. There are no separate child resources to create — everything is part of the `CreateContainerInstance` API.

2. **No auxiliary resources.** Container instances don't require security lists, route tables, or other infrastructure scaffolding. The VNIC attaches to an existing subnet, and NSGs are referenced by OCID.

3. **Matches the provider API.** Both the Terraform resource (`oci_container_instances_container_instance`) and the Pulumi resource (`containerengine.ContainerInstance`) are single resources with nested blocks. There is nothing to bundle.

## Why Containers Are a Repeated List

The `containers` field is `repeated Container` — a list, not a map keyed by name or image URL.

### Why Not a Map?

A map keyed by container name was considered. The list was chosen because:

1. **Names are optional.** The `displayName` field on Container is optional. A map would require every container to have a unique key, adding a mandatory field that the API doesn't require.

2. **Positional semantics exist.** OCI assigns container IDs based on the order of containers in the creation request. The first container is often treated as the "primary" container in monitoring and logging tools. A list preserves this ordering; a map does not.

3. **Matches the provider API.** Both the Terraform resource (`containers` is a list of blocks) and the Pulumi SDK use list semantics. Introducing a map would require a transformation layer.

4. **Simplest YAML authoring.** Users write `containers: [{ imageUrl: ... }, { imageUrl: ... }]`. A map would require a wrapper key for each container, making the common single-container case more verbose.

## Why VNICs Are Top-Level, Not Embedded in Containers

VNICs are a top-level repeated field on the spec, separate from the containers list. This differs from how Kubernetes pods define networking (implicitly shared via the pod's network namespace) but matches OCI's API model:

1. **VNICs are instance-level, not container-level.** All containers in the instance share all VNICs. A container cannot have its own private VNIC. Placing VNICs inside the Container message would imply per-container networking, which is not how the service works.

2. **Multiple VNICs are valid.** An instance can have multiple VNICs attached to different subnets (e.g., one public-facing, one private). This is an instance-level concern, not a container-level one.

3. **NSGs reference the VNIC, not the container.** Network security group associations are on VNICs. Containers inherit the combined security posture of all VNICs.

## Why Security Context Is Per-Container, Not Instance-Level

The `securityContext` field is on the Container message, not on the top-level OciContainerInstanceSpec. This allows different containers in the same instance to have different security postures:

1. **Different trust levels.** A sidecar from a trusted vendor (e.g., a metrics exporter) may need different capabilities than the primary application container. Per-container security contexts enable least-privilege for each container.

2. **Different UID/GID requirements.** The primary application may run as UID 1000, while an Envoy sidecar needs to run as UID 101 (its default). Instance-level UID/GID would force all containers to the same user.

3. **Matches the provider API.** OCI's Container Instance API accepts `security_context` per container, not per instance. The design preserves this granularity.

The only instance-level constraint is the `containerRestartPolicy`, which applies to all containers. This is intentional — OCI does not support per-container restart policies.

## Why Volumes Are Instance-Level but Mounts Are Per-Container

Volumes are defined at the instance level (`spec.volumes`) and mounted into individual containers via `container.volumeMounts`. This two-level design mirrors Kubernetes and is dictated by OCI's API:

1. **Shared storage model.** The primary use case for volumes is sharing data between containers (logs, config, temp files). Defining volumes at the instance level and mounting them per-container makes sharing explicit and intentional.

2. **Mount options differ per consumer.** The same volume may be mounted read-write in the producer container and read-only in the consumer container. Per-container `volumeMounts` with `isReadOnly` enable this.

3. **Deduplication.** Without instance-level volume definitions, two containers mounting the same emptydir would need to define the volume inline, risking inconsistent volume configurations.

4. **Volume limit is global.** OCI enforces a 32-volume limit per container instance. Instance-level definition makes this limit visible and manageable.

## Why Configfile Volumes Use Base64-Encoded Data

The `VolumeConfig.data` field accepts base64-encoded file contents. This is not an abstraction choice — it directly matches OCI's API:

1. **OCI API requirement.** The Container Instances API accepts `data` as a base64-encoded string. The IaC modules pass it through directly.

2. **Binary-safe.** Base64 encoding handles any file content — YAML, JSON, binary certificates, PEM keys — without escaping issues in proto/JSON/YAML serialization.

3. **Kubernetes precedent.** Kubernetes ConfigMaps and Secrets use base64 encoding for `binaryData`. The pattern is familiar to the target audience.

**Trade-off acknowledged:** Base64-encoded data is not human-readable in YAML manifests. For complex configurations, users should base64-encode externally (`cat config.yaml | base64`) and paste the result. This is the same workflow as Kubernetes Secret management.

## Why Resource Principal Control Is Per-Container

The `isResourcePrincipalDisabled` field is on the Container message, not the instance spec. This enables:

1. **Least-privilege per container.** A metrics exporter sidecar that only needs to push metrics to a third-party service does not need OCI API access. Disabling resource principal for that container reduces the blast radius of a container compromise.

2. **Matches the OCI API.** Resource principal enablement is per-container in OCI's API. The IaC modules pass the boolean directly.

The default is `false` (resource principal enabled), which is the common case — most containers running in OCI benefit from automatic authentication to OCI services.

## Why the State Lifecycle Field Is Omitted

OCI Container Instances support a `state` field with values ACTIVE and INACTIVE. The OciContainerInstance spec intentionally omits this field:

1. **Planton convention.** Resources are always deployed to their active state. To decommission a resource, delete it. This is consistent across all Planton components.

2. **INACTIVE is rarely useful.** An inactive container instance stops all containers but retains the resource and its configuration. This is an operational state transition, not a desired-state declaration. Users who need to temporarily stop an instance can use the OCI CLI (`oci container-instances container-instance stop`).

3. **Simpler lifecycle model.** The spec expresses "what should exist," not "what state should it be in." Removing the state field eliminates a class of reconciliation edge cases (e.g., what happens if the spec says ACTIVE but the instance is INACTIVE due to a manual stop?).

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Defined Tags** — OCI defined tags (namespace-scoped, schema-validated) require a tag namespace to be created first. Freeform tags (from Planton labels) cover the majority of tagging use cases. Defined tag support can be added when the tag namespace pattern is established across OCI components.

- **Init Containers** — OCI Container Instances do not currently support Kubernetes-style init containers (containers that run to completion before the main containers start). All containers start simultaneously. If OCI adds this capability, the spec can be extended with an `initContainers` field using the same Container message type.

- **Container Instance Restart via Spec** — The OCI API supports restart operations on running instances. This is an imperative operation, not a declarative state change, and is better handled via CLI or operational tooling rather than the spec.

- **GPU Shapes** — OCI Container Instances currently only support CI.Standard.E4.Flex and CI.Standard.E3.Flex shapes. GPU shapes for container instances may become available in the future. When they do, the `shape` field already accepts any string, so no spec change is needed.

- **Logging Integration** — OCI logging can be configured to capture container stdout/stderr. This is a compartment-level or log group-level configuration, not a container instance-level setting. It would be managed via OciLoggingLogGroup when that component is implemented.

## Research Notes

### Container Instance Shape Limits

OCI Container Instance flex shapes have the following limits:

| Shape | Min OCPUs | Max OCPUs | Min Memory | Max Memory | Memory per OCPU |
|-------|-----------|-----------|------------|------------|-----------------|
| CI.Standard.E4.Flex | 1 | 64 | 1 GB | 1024 GB | 1-64 GB per OCPU |
| CI.Standard.E3.Flex | 1 | 64 | 1 GB | 1024 GB | 1-64 GB per OCPU |

When `memoryInGbs` is omitted, OCI assigns the minimum memory for the requested OCPU count (1 GB per OCPU). For production workloads, always set `memoryInGbs` explicitly.

### VNIC and Network Limits

- A container instance can have multiple VNICs, each attached to a different subnet.
- All containers share all VNICs and the same network namespace.
- Containers communicate over `localhost` (127.0.0.1) — no inter-container networking configuration needed.
- Public IPs are assigned per VNIC, not per container.
- NSG associations are per VNIC. A container instance with two VNICs can have different security postures for each network interface.

### Volume Limits

- Maximum 32 volumes per container instance.
- Emptydir volumes are ephemeral — data is lost when the instance is deleted (not when a container restarts).
- Configfile volumes have a size limit per file imposed by the OCI API (files are passed inline in the creation request).
- The `MEMORY` backing store for emptydir volumes consumes instance memory. A 2 GB tmpfs volume on a 4 GB instance leaves only 2 GB for containers.

### Health Check Behavior

- Health checks run independently per container. A health check failure on one container does not affect other containers.
- The `kill` failure action terminates the container. Combined with `containerRestartPolicy: always`, this creates an automatic restart loop for unhealthy containers.
- The `none` failure action logs the failure without taking action. This is useful for monitoring during initial deployment.
- HTTP health checks follow redirects (3xx responses are treated as failures, not followed).
- Health check probes originate from within the instance's network namespace — they access `localhost` ports.

### Container Restart Behavior

- `always`: Container is restarted regardless of exit code. Suitable for long-running services.
- `on_failure`: Container is restarted only on non-zero exit code. Exit code 0 (success) does not trigger a restart.
- `never`: Container is never restarted. The instance remains running but the container stays in a terminated state.
- When all containers in a `never`-policy instance have exited, the instance transitions to INACTIVE state.
- Restart attempts use exponential backoff to prevent rapid restart loops for containers that crash immediately on startup.

### Resource Principal (v2.2)

OCI resource principal provides automatic authentication for containers to access OCI APIs without explicit credential management:

- Enabled by default for all containers.
- Supports OCI SDK calls to services like Object Storage, Autonomous Database, Vault, and others.
- The instance's compartment and tenancy determine the scope of resource principal access. IAM policies control which OCI resources the container can access.
- `isResourcePrincipalDisabled: true` removes the resource principal token endpoint from the container's metadata service. The container cannot authenticate to OCI APIs via resource principal.
