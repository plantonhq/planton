# GCP Region Network Endpoint Group

Deploys a regional network endpoint group (`google_compute_region_network_endpoint_group`) — the bridge that lets a load balancer's backend service send traffic to something other than a group of Compute Engine VMs: a Cloud Run service, a Cloud Functions function, an App Engine service, a Private Service Connect endpoint, or an external origin.

## What Gets Created

A single regional NEG in the chosen region. What it points at is decided by `networkEndpointType`:

- **`SERVERLESS`** (default) — one of `cloudRun`, `cloudFunction`, or `appEngine`. This is how serverless workloads sit behind an external Application Load Balancer (custom domains, Cloud CDN, Cloud Armor, IAP in front of Cloud Run).
- **`PRIVATE_SERVICE_CONNECT`** — a PSC endpoint fronting a published producer service or a Google API (`pscTargetService`).
- **`INTERNET_IP_PORT` / `INTERNET_FQDN_PORT`** — an external origin reached over the internet.
- **`GCE_VM_IP_PORTMAP`** — PSC port mapping to VM IP:port targets.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- For a serverless NEG, the Cloud Run/Functions/App Engine workload should live in the **same region** (it need not exist yet — GCP resolves endpoints at serving time)

## Quick Start

Create a file `neg.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRegionNetworkEndpointGroup
metadata:
  name: web-neg
spec:
  projectId:
    value: my-gcp-project-123
  region: us-central1
  cloudRun:
    service:
      value: my-cloud-run-service
```

Deploy:

```shell
planton apply -f neg.yaml
```

This creates a serverless NEG in `us-central1` pointing at the Cloud Run service `my-cloud-run-service`; reference its `self_link` from a `GcpBackendService` backend to put Cloud Run behind a global load balancer.

## Configuration Reference

### Endpoint target (by `networkEndpointType`)

| Block | Endpoint type | Fields |
|-------|---------------|--------|
| `cloudRun` | SERVERLESS | `service` (ref to GcpCloudRun or name), `tag`, `urlMask` |
| `cloudFunction` | SERVERLESS | `function`, `urlMask` |
| `appEngine` | SERVERLESS | `service`, `version`, `urlMask` (all optional — empty routes to the default app) |
| `pscTargetService` + `pscData` | PRIVATE_SERVICE_CONNECT | target service URL + optional producer port |
| `network` / `subnetwork` | PSC / INTERNET / GCE_VM_IP_PORTMAP | VPC and subnet self-links |

For a SERVERLESS NEG set exactly one of `cloudRun` / `cloudFunction` / `appEngine`; each serverless block needs either its target (`service`/`function`) or a `urlMask`.

`deletionPolicy` controls what destroy does: `DELETE` (default) removes the NEG — GCP refuses while a backend service still references it, so create the replacement before destroying the original; `PREVENT` fails the destroy outright; `ABANDON` leaves the NEG in GCP unmanaged.

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. Immutable. |
| `networkEndpointGroupName` | `string` | `metadata.name` | Cloud-side name (RFC1035). Immutable. |
| `region` | `string` | — (required) | Region the NEG lives in. Immutable. |
| `networkEndpointType` | `string` | `SERVERLESS` | Endpoint type. Immutable. |
| `description` | `string` | `""` | What the NEG fronts. Immutable. |
| `network` | `StringValueOrRef` | — | VPC for PSC/INTERNET/PORTMAP NEGs. Can reference a GcpVpcNetwork. Immutable. |
| `subnetwork` | `StringValueOrRef` | — | Subnet for PSC/PORTMAP NEGs. Can reference a GcpSubnetwork. Immutable. |
| `pscTargetService` | `string` | — | Target service URL for PSC/INTERNET NEGs (required for PSC). Immutable. |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value a backend service references in `backends[].group` |
| `network_endpoint_group_name` | `string` | Name of the NEG in GCP |
| `network_endpoint_type` | `string` | The NEG's endpoint type |
| `region` | `string` | Region the NEG lives in |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Fully immutable**: every field is ForceNew — any change destroys and recreates the NEG. Because an in-use NEG cannot be deleted, recreating one that a backend service references should be done create-before-destroy (Planton handles the ordering when you change the referencing backend service).
- **Serverless target need not exist at create time**: GCP resolves serverless endpoints at serving time, so a NEG can name a Cloud Run service that is deployed afterward.
- **Region must match**: a serverless NEG must be in the same region as the workload it fronts; a backend service can combine serverless NEGs from multiple regions.
- **No health checks for serverless backends**: a backend service whose backends are all serverless NEGs needs no health check — the serverless platform manages health.

## Related Kinds

- [GcpBackendService](/docs/catalog/gcp/gcpbackendservice) — references this NEG in `backends[].group`
- [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) — the serverless workload a SERVERLESS NEG fronts
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network for PSC/INTERNET NEGs
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the NEG

## Additional Resources

- [Serverless NEGs overview](https://cloud.google.com/load-balancing/docs/negs/serverless-neg-concepts)
- [Internet NEGs overview](https://cloud.google.com/load-balancing/docs/negs/internet-neg-concepts)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
