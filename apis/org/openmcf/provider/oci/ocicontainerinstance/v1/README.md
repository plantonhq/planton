# Overview

The **OCI Container Instance API Resource** provides a consistent and standardized interface for deploying and managing Oracle Cloud Infrastructure Container Instances — OCI's serverless container service for running containers without provisioning or managing compute infrastructure. A container instance is a pod-like construct: one or more containers sharing the same network namespace (VNICs) and volumes, scheduled onto a flex compute shape in a single availability domain. This component wraps the `oci_container_instances_container_instance` API surface with the standard OpenMCF KRM pattern.

## Purpose

This API resource streamlines the deployment of OCI Container Instances by offering a unified interface that covers the full range of serverless container configurations — from a minimal single-container service to a multi-container sidecar topology with health checks, security hardening, config file injection, and private registry authentication. It enables users to:

- **Run Multi-Container Workloads Without Infrastructure Management**: Deploy one or more containers in a pod-like construct where containers share networking and volumes. No VM provisioning, OS patching, or cluster management required — OCI manages the underlying compute.
- **Size Compute with Flex Shapes**: Container Instance shapes are always flex (CI.Standard.E4.Flex, CI.Standard.E3.Flex), allowing independent OCPU and memory configuration via `shapeConfig`. Individual containers can set resource limits within the instance-level envelope via `resourceConfig`.
- **Implement Sidecar Patterns**: Run auxiliary containers alongside the primary workload — log collectors, metrics exporters, proxies — sharing volumes and communicating over localhost. The multi-container model mirrors Kubernetes pod semantics without requiring a cluster.
- **Monitor Container Health**: Configure HTTP and TCP health checks with customizable thresholds, intervals, and failure actions. OCI restarts unhealthy containers based on the health check results and the configured restart policy.
- **Harden Container Security**: Apply Linux security contexts per container — enforce non-root execution, read-only root filesystems, specific UID/GID, and Linux capability grants/revocations. Each container can have a different security posture within the same instance.
- **Inject Configuration Files**: Mount configfile volumes containing base64-encoded file data directly into containers. This eliminates the need for config maps, secrets managers, or baked-in configuration for simple use cases.
- **Pull from Private Registries**: Authenticate to private container registries using basic credentials or OCI Vault secrets. Supports OCI Container Registry (OCIR), Docker Hub, GitHub Container Registry, and any Docker-compatible registry.
- **Compose with Other OCI Resources**: Reference OciCompartment, OciSubnet, and OciNetworkSecurityGroup outputs via `StringValueOrRef` for declarative, cross-resource dependency chains.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Pod-Like Multi-Container Model**: Multiple containers per instance share the same network namespace (communicate over localhost) and can mount the same volumes. This enables sidecar, adapter, and ambassador patterns without a container orchestrator.
- **Flex Shape Compute**: CI.Standard.E4.Flex and CI.Standard.E3.Flex shapes with independent OCPU and memory configuration. Container-level resource limits (`resourceConfig`) subdivide the instance-level allocation.
- **Health Checks**: HTTP and TCP probes with configurable initial delay, interval, timeout, failure threshold, success threshold, and failure action (kill or none). HTTP checks support custom headers and URL paths.
- **Linux Security Context**: Per-container security settings — non-root user enforcement, read-only root filesystem, UID/GID override, and Linux capability management (add/drop).
- **Two Volume Types**: Emptydir volumes (disk-backed or tmpfs) for ephemeral shared storage. Configfile volumes for injecting configuration files as base64-encoded data.
- **Image Pull Secrets**: Basic authentication (base64-encoded username/password) and OCI Vault-based credentials for private registries.
- **DNS Configuration**: Override subnet DHCP DNS settings with custom nameservers, search domains, and resolver options.
- **Resource Principal Access**: OCI resource principal (v2.2) enabled by default for each container, providing automatic authentication to OCI APIs. Can be disabled per container.
- **Restart Policies**: Instance-level restart policy (always, never, on_failure) controlling container lifecycle after exits.
- **Graceful Shutdown**: Configurable shutdown timeout for orderly container termination during instance stop or delete operations.
- **Automatic Tagging**: Standard OpenMCF freeform tags applied to the container instance (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Infra-Chart Composability**: Exports 2 stack outputs (`containerInstanceId`, `containerIds`) for downstream `StringValueOrRef` references. Consumes `compartmentId` from OciCompartment and `subnetId` from OciSubnet.

## How OCI Container Instances Differ from Other Providers

Understanding these differences is essential when coming from AWS, Azure, or GCP:

| Aspect | OCI Container Instance | AWS Fargate (ECS) | Azure Container Instances | GCP Cloud Run |
|--------|----------------------|-------------------|--------------------------|---------------|
| **Container model** | Multi-container instance (pod-like, shared networking/volumes) | Task with multiple containers (shared networking/volumes) | Container group with multiple containers (shared networking/volumes) | Single container per service (revision-based) |
| **Compute sizing** | Flex shapes with independent OCPU/memory | vCPU/memory pairs from a fixed list | vCPU/memory from a fixed list | vCPU/memory from a fixed list |
| **Networking** | VNICs attached to subnets, containers share network namespace | ENIs attached to subnets, containers share network namespace | Deployed into VNet subnets or public IP | VPC connectors (optional), primarily public endpoints |
| **Health checks** | HTTP and TCP probes per container | ELB-based or ECS container health checks | Liveness and readiness probes per container | HTTP readiness probes (built-in) |
| **Volume types** | Emptydir (disk/tmpfs) and configfile | EFS, EBS, bind mounts | Azure Files, emptyDir, gitRepo, secret | In-memory volumes only |
| **Config injection** | Configfile volumes (base64-encoded files) | SSM Parameter Store, Secrets Manager env vars | Azure Files, secret volumes | Environment variables, Secret Manager |
| **Security context** | Per-container Linux capabilities, UID/GID, read-only rootfs | Per-container Linux capabilities via task definition | No per-container security context | No per-container security context |
| **Identity** | Resource principal (v2.2) per container, can be disabled | IAM task role | Managed identity | Service account |
| **Scaling** | Manual (one instance per resource) | Auto-scaling via ECS service | Manual or KEDA-based | Automatic (request-based) |
| **Restart policy** | Instance-level (always, never, on_failure) | Task-level via ECS service desired count | Container group-level (always, never, on_failure) | Automatic (managed by Cloud Run) |

Key distinctions for OCI newcomers:

- **Flex Shapes Are the Only Option.** Unlike Fargate or ACI where you pick from fixed vCPU/memory combinations, OCI Container Instances always use flex shapes. You set OCPUs and memory independently, and can subdivide the allocation across containers via `resourceConfig`.
- **Resource Principal Is On by Default.** Each container gets OCI resource principal (v2.2) access automatically — no IAM role attachment or managed identity configuration needed. Disable it per container with `isResourcePrincipalDisabled: true` if the container should not access OCI APIs.
- **Configfile Volumes Replace Secrets/ConfigMaps.** OCI Container Instances have a built-in mechanism for injecting configuration files as base64-encoded data. This eliminates the need for external config stores for simple use cases, though it means config data is stored in the resource definition.
- **No Auto-Scaling.** OCI Container Instances do not have built-in auto-scaling. Each OpenMCF resource creates exactly one container instance. For scale-out workloads, use OKE (OciContainerEngineCluster + OciContainerEngineNodePool) with horizontal pod autoscaling.
- **Per-Container Security Context.** OCI supports Linux security contexts at the container level, not the instance level. Different containers in the same instance can have different UID/GID settings, capability profiles, and filesystem access modes.

## Critical Constraints

- **Shapes Are Always Flex**: Only `CI.Standard.E4.Flex` and `CI.Standard.E3.Flex` are available. There are no fixed-shape or GPU options for Container Instances.
- **Single Availability Domain**: A container instance runs in exactly one AD and one fault domain. There is no multi-AD redundancy at the instance level — deploy multiple instances across ADs for high availability.
- **No Live Migration**: Container instances cannot be moved between availability domains or compartments after creation. Changing `compartmentId` or `availabilityDomain` forces recreation.
- **Volume Limit**: A container instance supports up to 32 volumes.
- **Environment Variable Size**: Total size of all environment variable names and values per container must be <= 64 KB.
- **Argument Size**: Total size of all arguments per container must be <= 64 KB.
- **No Persistent Storage**: Emptydir volumes are ephemeral — data is lost when the container instance is deleted. Configfile volumes are read-only injections. For persistent storage, use OCI Block Volumes or Object Storage accessed via OCI APIs from within the container.
- **Restart Policy Is Instance-Level**: The `containerRestartPolicy` applies to all containers in the instance. Individual containers cannot have different restart policies.
- **No Init Containers**: OCI Container Instances do not support Kubernetes-style init containers. All containers start simultaneously. Use entrypoint scripts for initialization logic.
- **State Lifecycle Omitted**: The `state` field (ACTIVE/INACTIVE) is intentionally omitted from the spec. OpenMCF resources are always deployed to their active state. Delete the resource to decommission the instance.

## Use Cases

- **Microservice Sidecar Pattern**: Run a primary application container alongside auxiliary containers (log shippers, metrics exporters, reverse proxies) sharing localhost networking and volumes. The pod-like model enables sidecar patterns without a Kubernetes cluster.
- **Batch and One-Off Jobs**: Run data processing, ETL, or migration tasks using `containerRestartPolicy: never` or `on_failure`. The instance terminates when the containers exit. Resource principal access simplifies authentication to OCI services (Object Storage, Autonomous Database).
- **Dev/Test Environments**: Spin up isolated application stacks with minimal configuration. A single manifest deploys an application with all its sidecar dependencies, tearing down cleanly when deleted.
- **Config-Driven Services**: Inject application configuration via configfile volumes instead of building configuration into container images. Update configuration by changing the volume data and redeploying.
- **Secure API Services**: Deploy hardened API containers with non-root execution, read-only root filesystems, dropped capabilities, and health check monitoring — meeting compliance requirements without cluster infrastructure overhead.
- **Private Registry Workloads**: Run containers from private OCI Container Registry (OCIR) or third-party registries using Vault-based credentials, avoiding plaintext credentials in manifests.

## Production Features

This resource provides complete support for production-grade OCI Container Instance deployments, including:

- **Multi-Container Composition**: Up to multiple containers per instance sharing networking and volumes. Container-level resource limits prevent a single container from consuming all instance resources.
- **Health Check Monitoring**: HTTP and TCP probes with configurable thresholds trigger container restart based on the instance's restart policy. Initial delay settings prevent false positives during application startup.
- **Security Hardening**: Non-root user enforcement, read-only root filesystem, UID/GID pinning, and Linux capability management. Each container can be independently hardened.
- **Private Registry Support**: Basic authentication and OCI Vault-based secrets for pulling from private registries. Vault-based secrets avoid embedding credentials in manifests.
- **Graceful Shutdown**: Configurable shutdown timeout ensures containers have time to drain connections and flush state before forceful termination.
- **DNS Customization**: Override subnet DNS settings for containers that need to resolve internal service names or use custom DNS infrastructure.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Infra-Chart Composability**: Designed to compose with OciCompartment (upstream dependency), OciSubnet, and OciNetworkSecurityGroup via `StringValueOrRef`.
