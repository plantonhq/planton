# Multi-Container Sidecar

This preset creates a multi-container OCI Container Instance with an application container and a log-forwarder sidecar sharing an emptydir volume, plus a configfile volume for injecting configuration at deploy time. This demonstrates Container Instance's key differentiator over simpler container services: multiple containers sharing a network namespace and volumes in a pod-like construct. The sidecar pattern is a well-established container architecture where a primary application container is paired with auxiliary containers that handle cross-cutting concerns like log shipping, metrics collection, or reverse proxying.

## When to Use

- Application containers that need a sidecar for log forwarding (Fluent Bit, Fluentd, Vector) to ship structured logs to a centralized logging service
- Services that require a reverse proxy sidecar (Envoy, nginx) for TLS termination, rate limiting, or protocol translation in front of the application
- Workloads that need configuration files injected at deploy time without baking them into the container image
- Any deployment where you want to separate application logic from infrastructure concerns (observability, networking, configuration) into distinct containers

## Key Configuration Choices

- **Two containers with explicit resource limits** (`containers[0].resourceConfig`, `containers[1].resourceConfig`) -- The app container gets 3 GB / 3 vCPUs and the log-forwarder gets 0.5 GB / 0.5 vCPUs, totaling 3.5 GB / 3.5 vCPUs out of the instance's 4 GB / 4 vCPUs envelope. The remaining headroom (0.5 GB / 0.5 vCPUs) is available for the container runtime and OS overhead. Resource limits are critical in multi-container instances because without them, a misbehaving sidecar could starve the primary application. In single-container instances (presets 01 and 02), resource limits are unnecessary since the container inherits the full instance envelope.
- **Shared emptydir volume for logs** (`volumes[0]: shared-logs`, `volumeType: emptydir`) -- The app container writes logs to `/var/log/app` and the log-forwarder reads from the same path. The emptydir volume is backed by ephemeral disk storage (`EPHEMERAL_STORAGE`), which survives container restarts but is deleted when the instance is terminated. Use `MEMORY` backing instead for a tmpfs-backed volume when log data is small and you want faster I/O at the cost of consuming instance memory.
- **Log-forwarder mounts logs read-only** (`containers[1].volumeMounts[0].isReadOnly: true`) -- The sidecar can read log files but cannot modify or delete them. This prevents a bug or misconfiguration in the forwarder from corrupting application logs. The app container mounts the same volume read-write so it can create and rotate log files normally.
- **Configfile volume for deploy-time configuration** (`volumes[1]: app-config`, `volumeType: configfile`) -- Injects a `config.yaml` file into the container at `/etc/app/config.yaml` from base64-encoded data provided in the manifest. This decouples configuration from the container image, allowing the same image to run with different configurations across environments. The app container mounts this volume read-only to prevent the application from accidentally overwriting its own config. Add more entries to the `configs` list to inject multiple files into the same volume.
- **Environment variable for log directory** (`containers[0].environmentVariables.LOG_DIR: /var/log/app`) -- Tells the application where to write logs via environment variable rather than hardcoding the path in the image. This makes the container image reusable across deployments that may mount the log volume at different paths.
- **2 OCPU / 4 GB instance shape** (`shapeConfig`) -- Sized for a typical app + lightweight sidecar combination. The 1:2 OCPU-to-memory ratio matches preset 01. If your sidecar is heavier (e.g., Envoy with complex routing rules), increase `memoryInGbs` and adjust the per-container `resourceConfig` limits accordingly.
- **No public/private IP decision** -- The VNIC does not set `isPublicIpAssigned`, so it inherits the subnet's default public IP assignment setting. This keeps the preset flexible for both public and private subnet deployments. Set it explicitly if your deployment requires a specific network posture.
- **No security context** -- Intentionally omitted to keep the preset focused on the multi-container and volume patterns. For production deployments, combine this preset with the security context settings from preset 02 (non-root user, read-only rootfs, dropped capabilities) on both containers.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the container instance will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain name (e.g., `Uocm:PHX-AD-1`) | `oci iam availability-domain list` or OCI Console > Compute > Instances > Create Instance |
| `<app-container-image-url>` | Container image URL for the primary application (e.g., `ghcr.io/org/myapp:v1.0`) | Your container registry |
| `<log-forwarder-image-url>` | Container image URL for the log forwarder sidecar (e.g., `docker.io/fluent/fluent-bit:latest`, `docker.io/timberio/vector:latest-alpine`) | Docker Hub, GitHub Container Registry, or your organization's registry |
| `<base64-encoded-config>` | Base64-encoded contents of the configuration file to inject (e.g., output of `base64 < config.yaml`) | Encode your application's config file with `base64` or `openssl base64 -in config.yaml` |
| `<subnet-ocid>` | OCID of the subnet for the container instance's VNIC | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |

## Related Presets

- **01-web-service** -- Use instead for simple single-container deployments without sidecars or shared volumes
- **02-private-hardened** -- Use instead for single-container production services with security hardening; combine the security context from preset 02 with this preset's multi-container pattern for a hardened sidecar deployment
