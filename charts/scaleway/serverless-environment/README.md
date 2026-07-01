# Scaleway Serverless Environment

The **Scaleway Serverless Environment** InfraChart deploys a serverless container on
Scaleway Cloud with optional private networking, container registry, and custom DNS
domain.

It supports **conditional resource generation via Jinja templates** -- for example, you
can choose whether to create a VPC for database connectivity, a private container
registry for your images, and a DNS zone for custom domain routing.

Chart resources and configuration parameters are defined in the [`templates`](templates)
directory and documented in [`values.yaml`](values.yaml).

---

## Architecture

The chart deploys a single serverless container with optional supporting infrastructure:

```
ScalewayVpc (optional)
  └── ScalewayPrivateNetwork (optional)
        └── ScalewayServerlessContainer ← ScalewayContainerRegistry (optional)
              └── ScalewayDnsRecord (optional)
                    ↑
              ScalewayDnsZone (optional)
```

When `create_network` is enabled, the container joins a Private Network for secure
access to databases and other VPC resources. When `create_container_registry` is
enabled, the container image is pulled from a dedicated Scaleway registry. When
`create_dns_zone` is enabled, a CNAME record is created pointing the custom domain
to the container's native Scaleway endpoint.

---

## Included Cloud Resources (conditional)

| Resource | Always created | Controlled by boolean flag |
|----------|----------------|----------------------------|
| **Scaleway VPC** | No | `create_network` |
| **Scaleway Private Network** | No | `create_network` |
| **Scaleway Container Registry** | No | `create_container_registry` |
| **Scaleway Serverless Container** | Yes | -- |
| **Scaleway DNS Zone** | No | `create_dns_zone` |
| **Scaleway DNS Record** (CNAME) | No | `create_dns_zone` |

### How the `create_network` flag works

* `create_network: true` -- Creates a VPC and Private Network. The container joins the
  Private Network so it can reach databases (RDB, Redis, MongoDB) and other resources on
  the same network without traversing the public internet.
* `create_network: false` (default) -- The container runs with public-only networking.
  Suitable for standalone APIs and services that do not need VPC connectivity.

### How the `create_container_registry` flag works

* `create_container_registry: true` -- Creates a private Scaleway Container Registry.
  The container's image endpoint is automatically wired to the registry via `valueFrom`.
* `create_container_registry: false` (default) -- Uses the `registry_endpoint` parameter
  as a plain value. Supports any OCI registry (Docker Hub, GHCR, ECR, etc.).

---

## Chart Input Values

| Parameter | Description | Example / Options | Required / Default |
|-----------|-------------|-------------------|-------------------|
| **create_network** | Create VPC and Private Network | `true` / `false` | Default `false` |
| **region** | Scaleway region | `fr-par`, `nl-ams`, `pl-waw` | Default `fr-par` |
| **create_container_registry** | Create container registry | `true` / `false` | Default `false` |
| **registry_name** | Registry namespace name | `my-registry` | Required if `create_container_registry=true` |
| **registry_endpoint** | External registry endpoint | `docker.io/library` | Used when `create_container_registry=false` |
| **image_name** | Container image name | `nginx`, `my-app` | Required |
| **image_tag** | Container image tag | `latest`, `v1.0` | Required |
| **container_name** | Serverless container name | `my-container` | Required |
| **port** | Container listening port | `8080` | Default `8080` |
| **privacy** | Endpoint privacy | `public`, `private` | Default `public` |
| **http_option** | HTTP behavior | `enabled`, `redirected` | Default `redirected` |
| **memory_limit_mb** | Memory limit per instance (MB) | `128`, `256`, `512`, `1024` | Default `256` |
| **cpu_limit** | CPU limit per instance (milliCPU) | `0` (auto), `140`, `280` | Default `0` |
| **min_scale** | Minimum instances (0 = scale-to-zero) | `0`, `1` | Default `0` |
| **max_scale** | Maximum instances | `5` | Default `5` |
| **create_dns_zone** | Create DNS zone with CNAME | `true` / `false` | Default `false` |
| **domain_name** | Parent domain for DNS | `example.com` | Required if `create_dns_zone=true` |
| **subdomain** | Subdomain prefix for DNS zone | `api` | Optional |
| **dns_record_name** | DNS record name relative to zone | `app` | Default `app` |

---

## Customization & Management

* Toggle `create_network` to add VPC connectivity for containers that need to access
  databases on a Private Network (e.g., from a database-stack deployment).
* Toggle `create_container_registry` based on whether you push images to Scaleway or use
  an external registry.
* Use `privacy: private` to require authentication tokens for the container endpoint.
* Use `http_option: redirected` to automatically redirect HTTP requests to HTTPS.
* Set `min_scale: 1` to eliminate cold starts at the cost of continuous billing.
* Set `min_scale: 0` for scale-to-zero -- the container stops when idle and incurs no
  compute charges.
* Resource references (`valueFrom` vs `value`) are automatically wired by the templates --
  no manual edits needed.

---

## Important Notes

* **Scaleway Serverless Containers are regional** -- ensure the `region` matches your
  container registry region for fastest image pulls.
* **Scale-to-zero** (`min_scale: 0`) means the first request after idle incurs a cold
  start latency (typically 1-3 seconds for small images).
* **DNS delegation** -- when using `create_dns_zone`, configure the nameservers (from the
  DNS zone outputs) at your domain registrar before the CNAME record will resolve.
* **Container registry is private by default** -- when `create_container_registry` is
  enabled, the registry is created with `isPublic: false` for security.
* **Memory determines pricing** -- Scaleway bills serverless containers based on memory
  allocation and execution duration. Choose the smallest memory limit that satisfies your
  workload requirements.

---

(c) 2025 Planton. All rights reserved.
