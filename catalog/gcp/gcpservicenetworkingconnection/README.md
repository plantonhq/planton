# GCP Service Networking Connection

Deploys a private services access connection (`google_service_networking_connection`) — the VPC peering between one of your networks and a service producer's network. For the default producer (`servicenetworking.googleapis.com`), this is the prerequisite that lets Cloud SQL, AlloyDB, Memorystore (PRIVATE_SERVICE_ACCESS mode), and Filestore hand out private IPs from ranges reserved inside your VPC.

## What Gets Created

One connection per (network, service) pair. GCP creates a VPC peering on the network (named after the service, e.g. `servicenetworking-googleapis-com`) and the producer carves service subnets out of the reserved `VPC_PEERING` address ranges you pass in.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A VPC network** — referenced via `network` (a `GcpVpcNetwork` resource or a literal self-link)
- **At least one reserved range** — a `GcpGlobalAddress` with `addressType: INTERNAL` and `purpose: VPC_PEERING`, referenced by name
- **IAM permissions** — `servicenetworking.services.addPeering` (e.g. `roles/servicenetworking.networksAdmin`)

## Quick Start

Create a file `psa-connection.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceNetworkingConnection
metadata:
  name: prod-vpc-psa
spec:
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: prod-vpc
      fieldPath: status.outputs.network_self_link
  reservedPeeringRanges:
    - valueFrom:
        kind: GcpGlobalAddress
        name: prod-vpc-psa-range
        fieldPath: status.outputs.name
```

Deploy:

```shell
planton apply -f psa-connection.yaml
```

Once the connection exists, managed services in this network can use private IP — e.g. a `GcpCloudSql` with `privateIpEnabled: true`.

## Configuration Reference

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | Project used for in-module API enablement. Set only when the network's project differs from the provider default. |
| `network` | `StringValueOrRef` | — (required) | VPC network to peer — name or self-link. Can reference a GcpVpcNetwork. Immutable. |
| `service` | `string` | `servicenetworking.googleapis.com` | The service producer to peer with. Immutable. |
| `reservedPeeringRanges` | `StringValueOrRef[]` | — (min 1) | Names of INTERNAL `VPC_PEERING` global address ranges. Mutable — append to grow capacity. |
| `updateOnCreationFail` | `bool` | `false` | Adopt a pre-existing connection for the same pair instead of failing. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE` removes the peering (destroy producer instances first), `PREVENT` fails the destroy, `ABANDON` leaves the peering serving unmanaged. |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `peering` | `string` | Name of the VPC peering created on the network (e.g. `servicenetworking-googleapis-com`) |
| `network` | `string` | The peered VPC network as the connection resolved it |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **One connection per pair**: GCP rejects a second connection for the same (network, service) pair — capacity grows by appending ranges to this resource, never by creating another connection.
- **Ranges by NAME**: `reservedPeeringRanges` carries global address names, not self-links or CIDRs — the API's own addressing for peering ranges.
- **Additive growth is safe**: changing the range list never disturbs service subnets the producer already provisioned.
- **Teardown ordering**: GCP refuses to delete the connection while the producer still holds subnets — destroy private-IP service instances (Cloud SQL, AlloyDB, Memorystore, ...) before this resource.
- **Deletion policy**: `deletionPolicy` controls what destroy does — `DELETE` (default) removes the peering, `PREVENT` fails the destroy to protect producer connectivity, and `ABANDON` leaves the peering serving unmanaged (the historical workaround for stuck teardowns).
- **No cloud-side name**: the connection is addressed by (network, service); `metadata.name` names only the Planton resource.

### Deliberately not modeled (recorded reasons)

- **`google_service_networking_peered_dns_domain`** — a separate provider resource that forwards a private DNS suffix over this peering; a real but second-order need (Tier-2 candidate on concrete pull).
- **`google_service_networking_vpc_service_controls`** — the VPC-SC enablement toggle for this connection; enterprise perimeter tooling that belongs with a broader VPC Service Controls story (Tier-2).

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network being peered
- [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) — reserves the `VPC_PEERING` ranges this connection hands to the producer
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — uses private IP through this connection
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project for API enablement

## Additional Resources

- [Private services access overview](https://cloud.google.com/vpc/docs/private-services-access)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
