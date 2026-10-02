# GCP Datastream Private Connection

The network link Datastream uses to reach databases that have no public address: a VPC peering between Datastream's Google-managed network and your VPC through a free /29, or a Private Service Connect interface on a network attachment you own. Connection profiles in the same project and region use it through their `privateConnection` reference, so one link serves every source reachable from that network.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `datastream.googleapis.com` on the project (never disabled on destroy)
- **Private connection** -- a `datastream_private_connection` with VPC peering or a PSC interface

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Datastream admin permissions (`roles/datastream.admin`) on the project, plus network admin on the VPC for peering.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpVpcNetwork`** -- the network Datastream peers with (`vpcPeeringConfig.vpc`, its `network_id`), or a network attachment for the PSC interface.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamPrivateConnection
metadata:
  name: data-vpc
spec:
  location: us-central1
  vpcPeeringConfig:
    vpc:
      value: projects/my-gcp-project/global/networks/data
    subnet: 10.200.0.0/29
```

```shell
planton apply -f datastream-private-connection.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Datastream region, e.g. `us-central1`. Immutable. |
| one connectivity option | -- | Exactly one of `vpcPeeringConfig` or `pscInterfaceConfig`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `privateConnectionId` | `string` | `metadata.name` | Immutable. |
| `displayName` | `string` | `metadata.name` | Immutable. |
| `vpcPeeringConfig` | `object` | -- | `vpc` (`GcpVpcNetwork` reference) and a free /29 `subnet`. |
| `pscInterfaceConfig` | `object` | -- | `networkAttachment` in the same region. |
| `createWithoutValidation` | `bool` | `false` | Skip Google's create-time checks. Immutable. |
| `labels` | `map<string,string>` | none | Merged with the platform attribution labels. |
| `deletionPolicy` | `string` | `FORCE` | `FORCE`, `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one of `vpcPeeringConfig` and `pscInterfaceConfig`.
- `vpcPeeringConfig.subnet` is an IPv4 CIDR of exactly /29 with no host bits set.
- `pscInterfaceConfig.networkAttachment` is `projects/{project}/regions/{region}/networkAttachments/{name}`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/privateConnections/{id}` -- what profiles reference |
| `private_connection_id` | `string` | The private connection's id |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Peering is not transitive.** A Cloud SQL or AlloyDB private IP sits in a network Google peers with your VPC, so Datastream reaches it only through a proxy VM in your VPC or a PSC interface.
- **Almost everything is immutable.** A change replaces the connection, and every profile using it must be re-pointed.
- **Delete the profiles first.** Google refuses to delete a private connection a profile still uses; `FORCE` (the default) also removes the routes Datastream created on it.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVpcNetwork** -- the peered network
- **GcpDatastreamConnectionProfile** -- the profiles that use the connection
- **GcpDatastreamStream** -- the streams those profiles feed

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
