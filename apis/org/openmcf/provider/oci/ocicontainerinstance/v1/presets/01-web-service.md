# Web Service Container Instance

This preset creates a single-container OCI Container Instance running an HTTP service with a health check, a public IP for direct access, and an always-restart policy. It uses the CI.Standard.E4.Flex shape with minimal resources (1 OCPU, 2 GB memory), making it the standard starting point for web applications, REST APIs, and any container workload that serves HTTP traffic. This is the "30-second decision" configuration for Container Instance.

## When to Use

- Running a web application or REST API as a serverless container without managing compute infrastructure
- First-time Container Instance deployments where you need a publicly accessible endpoint
- Prototyping containerized services before graduating to OKE or a production-hardened configuration
- Stateless HTTP workloads that benefit from automatic restart on failure (container crashes, OOM kills)

## Key Configuration Choices

- **CI.Standard.E4.Flex with 1 OCPU / 2 GB** (`shape`, `shapeConfig`) -- AMD EPYC (Milan) flex shape, the same processor family as VM.Standard.E4.Flex compute instances. Container Instance shapes are always flex, so you pick exact OCPUs and memory. 1 OCPU (2 vCPUs) with 2 GB is sufficient for most web frameworks and lightweight services. Scale up by increasing both values; the E4 flex shape supports up to 64 OCPUs with 1-64 GB per OCPU. The 1:2 OCPU-to-memory ratio keeps costs low for CPU-bound web workloads; increase to 1:4 or 1:8 if your service is memory-intensive.
- **HTTP health check on /healthz:8080** (`containers[0].healthChecks`) -- OCI monitors the container and restarts it when the health check fails. The `/healthz` path and port 8080 are common conventions for containerized web services. Adjust the path and port to match your application's health endpoint. The 10-second initial delay gives the container time to start before the first probe; increase this if your application has a slow startup (e.g., JVM warm-up, large model loading).
- **Restart policy always** (`containerRestartPolicy: always`) -- The container is restarted regardless of exit code. This is the correct policy for long-running services that should never stop. Use `on_failure` instead for batch jobs that should stay stopped after successful completion, or `never` for one-shot tasks.
- **Public IP assigned** (`vnics[0].isPublicIpAssigned: true`) -- The container instance is directly reachable from the internet on all ports allowed by the subnet's security lists. The instance must be in a public subnet with an Internet Gateway route. Use preset 02-private-hardened instead if the service should only be reachable from within the VCN.
- **Single VNIC, no NSG** -- Kept minimal for simplicity. Network access is controlled by the subnet's default security list. For fine-grained ingress/egress rules, add `nsgIds` referencing an `OciNetworkSecurityGroup`.
- **No security context, volumes, or image pull secrets** -- Intentionally omitted to keep the preset focused on the simplest viable deployment. Add a `securityContext` for production hardening (see preset 02), `volumes` for shared storage (see preset 03), or `imagePullSecrets` if pulling from a private registry.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the container instance will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain name (e.g., `Uocm:PHX-AD-1`) | `oci iam availability-domain list` or OCI Console > Compute > Instances > Create Instance |
| `<container-image-url>` | Container image URL (e.g., `docker.io/library/nginx:latest`, `ghcr.io/org/app:v1.2`) | Your container registry; default registry is docker.io/library if not specified |
| `<public-subnet-ocid>` | OCID of a public subnet for the container instance's VNIC | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |

## Related Presets

- **02-private-hardened** -- Use instead for production backend services in a private subnet with security hardening, NSG association, and graceful shutdown controls
- **03-multi-container-sidecar** -- Use instead when you need multiple containers sharing volumes (e.g., app + log forwarder or app + reverse proxy)
