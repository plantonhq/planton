# Overview

The **GCP Cloud Run** deployment component provisions and manages HTTP services on Google Cloud Run (v2). It provides a declarative interface for deploying containerized applications in a fully managed serverless environment, covering container configuration, networking, database connectivity, custom domains, and security controls.

## Purpose

This component abstracts the complexity of Cloud Run v2 service provisioning into a single, validated manifest. It supports:

- **Container deployment** with configurable CPU, memory, scaling, environment variables, and Secret Manager references
- **VPC connectivity** via Direct VPC Egress for accessing private resources (databases, caches, internal APIs)
- **Cloud SQL connectivity** with native volume mount or Auth Proxy sidecar -- both IAM-authenticated and TLS-encrypted, no VPC required
- **Custom DNS** mapping with automatic SSL certificate provisioning
- **Security controls** including ingress restrictions, IAM authentication, execution environment selection, and deletion protection

## Key Features

| Feature | Description |
|---|---|
| Container configuration | Image, CPU (1/2/4 vCPU), memory (128-32768 MiB), port, replicas (min/max), env vars, secrets |
| Autoscaling | Scale-to-zero capable; min/max instance counts; per-instance concurrency limit |
| VPC Direct Egress | Cloud Run instances get a NIC in the customer subnet for private resource access |
| Cloud SQL native connection | GCP-managed volume mount at `/cloudsql/<connection_name>`; no sidecar or VPC needed |
| Cloud SQL Auth Proxy | Sidecar container providing TCP proxy at `localhost:<port>`; supports `--private-ip` via VPC |
| Custom DNS | Domain mapping with Cloud DNS verification and automatic SSL |
| Ingress control | Public, internal-only, or internal + load balancer |
| Secret Manager | Inject secrets directly from GCP Secret Manager as environment variables |
| Deletion protection | Prevent accidental deletion of production services |
| Cross-resource references | Foreign key references to GcpProject, GcpVpc, GcpSubnetwork, and GcpCloudSql resources |

## Cloud SQL Connectivity Options

Three approaches are available for connecting Cloud Run to Cloud SQL. The native connection is recommended for most use cases.

| Approach | VPC Required? | How It Connects | DATABASE_URL Format |
|---|---|---|---|
| **Native connection** (recommended) | No | GCP-managed volume mount | `?host=/cloudsql/<connection_name>` |
| Auth Proxy sidecar | No (unless `use_private_ip`) | TCP proxy at localhost | `@localhost:<port>` |
| Direct VPC Egress | Yes + PSA | Private IP via VPC peering | `@<private_ip>:<port>?sslmode=require` |

## Use Cases

- **Public APIs and web applications** -- scale-to-zero, public ingress, custom domains
- **Backend services with database access** -- Cloud SQL native connection, environment secrets
- **Internal microservices** -- VPC egress, internal-only ingress, IAM authentication

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudRun
metadata:
  name: my-api
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  container:
    image:
      repo: us-docker.pkg.dev/my-project/repo/app
      tag: v1.0.0
    port: 8080
    cpu: 1
    memory: 512
    replicas:
      min: 0
      max: 10
  allowUnauthenticated: true
```

## Further Reading

- [Examples](examples.md) -- configuration examples for all features
- [Research Documentation](docs/README.md) -- deployment landscape, best practices, design rationale
- [Presets](presets/) -- ready-to-use configurations for common patterns
