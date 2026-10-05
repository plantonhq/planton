# GCP Global Address

Reserves a static IP address or CIDR range at global scope in Google Cloud. External addresses provide stable public IPs for HTTP(S) load balancers, Cloud CDN, and global forwarding rules. Internal addresses reserve private IP ranges for VPC peering (used by Cloud SQL, Memorystore, AlloyDB, Filestore) or Private Service Connect endpoints. Every field except labels is ForceNew — any configuration change releases and re-reserves the address.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Global Address** -- a `compute.GlobalAddress` resource in the specified GCP project, configured as either an external public IP or an internal private IP range depending on the `addressType` setting
- **Compute Engine API enablement** -- `compute.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the address reservation will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A VPC network** (only for internal addresses) for the IP range allocation. Provide the network self-link directly or reference a GcpVpcNetwork Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **GCP Global Address**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **External Static IP** preset in the [Presets](#presets) tab to pre-populate a public IP reservation.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGlobalAddress
metadata:
  name: lb-ip
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  addressName: prod-lb-ip
  addressType: EXTERNAL
  ipVersion: IPV4
```

```shell
planton apply -f global-address.yaml
```

This reserves a public IPv4 address at global scope. GCP automatically assigns an available IP. No VPC network or prefix length is needed for external addresses. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the global address to infrastructure deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: production-vpc
      fieldPath: status.outputs.network_self_link
```

The InfraPipeline resolves the dependency graph, deploys the project and VPC first, then provisions the global address with the resolved values.

## Key Configuration

These are the most important decisions when configuring a global address. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Address type** -- Set `addressType` to `EXTERNAL` (default) for a public IP, or `INTERNAL` for a private IP range within a VPC. External addresses are used with load balancers and CDN; internal addresses are used for VPC peering and Private Service Connect.

**Purpose** -- Required when `addressType` is `INTERNAL`. Set to `VPC_PEERING` to reserve a CIDR range for managed services (Cloud SQL, Memorystore, AlloyDB, Filestore private networking). Set to `PRIVATE_SERVICE_CONNECT` for PSC endpoints. Leave empty for external addresses.

**Prefix length** -- Required when `purpose` is `VPC_PEERING`. Determines the size of the reserved CIDR range (e.g., `20` for a /20 with 4,096 IPs). Use `/20` for typical multi-service environments, `/16` for large-scale deployments, or `/24` when IP space is constrained.

**IP version** -- Defaults to `IPV4`. Set to `IPV6` when the target load balancer or forwarding rule requires an IPv6 address. Only applicable for external addresses.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional) | `network` | `status.outputs.network_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `address` | Reserved IP address or start of reserved CIDR range | DNS A records, load balancer frontend IP, peering connection range |
| `self_link` | Self-link URL of the global address resource | Forwarding rules, load balancer configurations |
| `name` | Name of the global address resource in GCP | GcpServiceNetworkingConnection `reservedPeeringRanges` — the private-services-access composition key |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**External static IP** -- Reserves a global public IPv4 address for HTTP(S) load balancers, Cloud CDN, or global forwarding rules. GCP assigns an available IP automatically, providing a stable address that persists across resource recreation. Start from the **External Static IP** preset.

**Internal VPC peering range** -- Reserves a /20 private CIDR range for VPC peering with managed services. Required for Cloud SQL, Memorystore, AlloyDB, and Filestore private networking via Private Services Access. Start from the **Internal VPC Peering Range** preset.

**Private Service Connect endpoint** -- Reserves a single internal IP for a Private Service Connect endpoint, enabling private connectivity to Google APIs or third-party services without traffic leaving the VPC. Start from the **Private Service Connect Endpoint Address** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the address reservation is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for internal address IP allocation
- [**GCP Forwarding Rule**](/infra-catalog/gcp-global-forwarding-rule) -- consumes the reserved external IP as a load-balancer frontend
- [**GCP Service Networking Connection**](/infra-catalog/gcp-service-networking-connection) -- consumes VPC_PEERING ranges by name for private services access