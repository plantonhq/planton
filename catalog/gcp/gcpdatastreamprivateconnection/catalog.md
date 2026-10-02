# GCP Datastream Private Connection

Lets Datastream reach databases with no public address -- VMs, on-premises servers over VPN or Interconnect, private instances behind a proxy -- through one shared link into your VPC: a peering through a free /29, or a Private Service Connect interface.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `datastream.googleapis.com` on the project (never disabled on destroy)
- **Private connection** -- a `datastream_private_connection` with VPC peering or a PSC interface

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Datastream admin permissions (`roles/datastream.admin`) on the project, plus network admin on the VPC for peering. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpVpcNetwork`** -- the network Datastream peers with, or a network attachment for the PSC interface.

## Deploy

### Console

Open the deployment store, find **GCP Datastream Private Connection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **VPC Peering** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamPrivateConnection
metadata:
  name: data-vpc
  org: acme-corp
  env: prod
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

This peers Datastream's network with the `data` VPC. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpVpcNetwork` from `vpcPeeringConfig.vpc`; connection profiles in the same chart reference this connection from `privateConnection`.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Peering or PSC interface** -- peering needs a free /29 in the VPC; a PSC interface needs a network attachment and no reserved range.

**The /29** -- it must overlap no subnet, peering, or route in the VPC, and it can never change.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `vpcPeeringConfig.vpc` | `status.outputs.network_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The connection's full resource name | `GcpDatastreamConnectionProfile.privateConnection` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**VPC peering** -- Datastream peered with the VPC its sources are reachable from. Start from the **VPC Peering** preset.

**PSC interface** -- Datastream connected through a network attachment. Start from the **PSC Interface** preset.

## Works With

- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the peered network
- [**GCP Datastream Connection Profile**](/cloud-catalog/gcp-datastream-connection-profile) -- the profiles that use it
- [**GCP Datastream Stream**](/cloud-catalog/gcp-datastream-stream) -- the streams they feed
