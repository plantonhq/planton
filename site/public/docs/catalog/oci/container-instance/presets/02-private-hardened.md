---
title: "Private Hardened Container Instance"
description: "This preset creates a production-grade OCI Container Instance in a private subnet with no public IP, full Linux security context hardening, NSG-based network segmentation, and a graceful shutdown..."
type: "preset"
rank: "02"
presetSlug: "02-private-hardened"
componentSlug: "container-instance"
componentTitle: "Container Instance"
provider: "oci"
icon: "package"
order: 2
---

# Private Hardened Container Instance

This preset creates a production-grade OCI Container Instance in a private subnet with no public IP, full Linux security context hardening, NSG-based network segmentation, and a graceful shutdown window. The container runs as a non-root user with a read-only root filesystem and all Linux capabilities dropped. This is the standard pattern for internal microservices, queue consumers, background workers, and any containerized backend that sits behind a load balancer or communicates only within the VCN.

## When to Use

- Backend microservices behind an OCI Load Balancer or Network Load Balancer that should not be directly internet-accessible
- Queue consumers, event processors, and background workers that communicate only with internal services
- Any production container workload where security policy mandates non-root execution, read-only filesystems, and least-privilege capabilities
- Internal APIs and gRPC services that are accessed only from other containers or compute instances within the VCN

## Key Configuration Choices

- **Private subnet with no public IP** (`vnics[0].isPublicIpAssigned: false`) -- The container instance is not reachable from the internet. Outbound internet access (for pulling images, calling external APIs) is available via the VCN's NAT Gateway. Inbound traffic is limited to sources within the VCN or connected via VPN/FastConnect. Use preset 01-web-service instead if the container needs direct public access.
- **NSG association** (`vnics[0].nsgIds`) -- Associates the container instance's VNIC with a Network Security Group for stateful, fine-grained ingress and egress rules. This is more flexible than subnet-level security lists because NSGs can be shared across subnets and applied selectively. Use the `OciSecurityGroup` component to define the rules.
- **Non-root execution** (`securityContext.isNonRootUserCheckEnabled: true`, `runAsUser: 1000`, `runAsGroup: 1000`) -- The container process runs as UID/GID 1000 instead of root. OCI validates at runtime that the container does not run as UID 0 and fails the start if it does. This prevents container escape attacks that exploit root privileges. Your container image must support running as a non-root user; most production images already do.
- **Read-only root filesystem** (`securityContext.isRootFileSystemReadonly: true`) -- The container's root filesystem is mounted read-only, preventing the process from writing to system directories. This mitigates attacks that modify binaries or inject malicious files into the container layer. If your application needs to write temporary files, mount an emptydir volume at the write path (see preset 03 for volume examples).
- **All capabilities dropped** (`securityContext.capabilities.dropCapabilities: [ALL]`) -- Removes all Linux capabilities from the container process, enforcing the principle of least privilege. Most application containers do not need any capabilities. If your workload requires specific capabilities (e.g., `NET_BIND_SERVICE` to bind to ports below 1024), add them explicitly via `addCapabilities` while keeping the `ALL` drop.
- **Restart on failure only** (`containerRestartPolicy: on_failure`) -- The container is restarted only when it exits with a non-zero exit code. A clean exit (code 0) leaves the container stopped. This is the correct policy for services that may perform graceful shutdown and should not be restarted after intentional termination. Use `always` instead if the service should never stop under any circumstances.
- **30-second graceful shutdown** (`gracefulShutdownTimeoutInSeconds: 30`) -- When the container instance is stopped or deleted, OCI sends SIGTERM and waits 30 seconds for the container to shut down gracefully before sending SIGKILL. This gives the application time to drain in-flight requests, close database connections, and flush buffers. Increase this for services with long-lived connections or large write buffers; decrease for stateless services that can stop instantly.
- **2 OCPU / 8 GB memory** (`shapeConfig`) -- A production-appropriate starting point with a 1:4 OCPU-to-memory ratio. This provides more headroom than the minimal 1:2 ratio in preset 01, accommodating services that maintain in-memory caches, connection pools, or process larger payloads. Scale up as needed; the E4 flex shape supports up to 64 OCPUs.
- **Health check with kill action** (`healthChecks[0].failureAction: kill`) -- When the health check fails 3 consecutive times, the container is killed and restarted (subject to the restart policy). The 15-second initial delay accommodates services with moderate startup time. The 10-second check interval is tighter than preset 01's 15 seconds, reflecting the higher reliability expectations of a production service.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the container instance will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain name (e.g., `Uocm:PHX-AD-1`) | `oci iam availability-domain list` or OCI Console > Compute > Instances > Create Instance |
| `<container-image-url>` | Container image URL (e.g., `ghcr.io/org/app:v1.2`, `us-ashburn-1.ocir.io/namespace/repo:tag`) | Your container registry |
| `<private-subnet-ocid>` | OCID of a private subnet for the container instance's VNIC (no Internet Gateway route) | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<nsg-ocid>` | OCID of a Network Security Group to associate with the container instance's VNIC | OCI Console > Networking > Network Security Groups, or `OciSecurityGroup` status outputs |

## Related Presets

- **01-web-service** -- Use instead for publicly accessible web services where security hardening is not yet a priority
- **03-multi-container-sidecar** -- Use instead when you need multiple containers sharing volumes, and combine with the security context settings from this preset for a hardened multi-container deployment
