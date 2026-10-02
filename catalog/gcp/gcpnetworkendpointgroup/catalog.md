# GCP Network Endpoint Group

Builds a network endpoint group (NEG) -- a named set of IP:port endpoints a backend service points at instead of an instance group. Set `zone` for a zonal group inside a VPC (VMs by instance, on-premises or other-cloud addresses reached over VPN or Interconnect, or internet endpoints for a regional external Application Load Balancer); leave it empty for a global internet group fronting an origin outside Google Cloud. The manifest's endpoint list is the group's whole membership. Serverless, Private Service Connect, and regional internet groups live in `GcpRegionNetworkEndpointGroup`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions exactly one of:

- **Zonal group** (`zone` set) -- a `compute.NetworkEndpointGroup` in that zone on your network, plus its membership written as one set through Google's bulk endpoint operation when `endpoints` is non-empty
- **Global internet group** (`zone` empty) -- a `compute.GlobalNetworkEndpointGroup`, plus one `compute.GlobalNetworkEndpoint` per entry in `endpoints`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Networks

- **Zonal groups need a network** -- a `GcpVpcNetwork` (and optionally a `GcpSubnetwork`); VM endpoints reference `GcpComputeInstance` resources in the same zone.
- **Global groups have no VPC** -- internet endpoints only.

## Deploy

### Console

Open the deployment store, find **GCP Network Endpoint Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Zonal VM Group for a Regional Load Balancer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpNetworkEndpointGroup
metadata:
  name: web-neg
  org: acme-corp
  env: prod
spec:
  zone: us-central1-a
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_self_link
  networkEndpointType: GCE_VM_IP_PORT
  defaultPort: 8080
  endpoints:
    - instance:
        valueFrom:
          kind: GcpComputeInstance
          name: web-1
          fieldPath: status.outputs.instance_name
    - instance:
        valueFrom:
          kind: GcpComputeInstance
          name: web-2
          fieldPath: status.outputs.instance_name
      port: 8081
```

```shell
planton apply -f network-endpoint-group.yaml
```

This creates a zonal group of two VM endpoints on port 8080 (the second on 8081) that a regional backend service can point at. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the group to the network, subnet, and instances deployed in the same InfraPipeline; a `GcpBackendService` then names this group's `self_link` in `backends[].group`.

## Key Configuration

These are the most important decisions when configuring a network endpoint group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- `zone` set for a zonal group (VM, hybrid, or regional internet endpoints); empty for a global internet group. Immutable.

**Endpoint type** -- zonal: `GCE_VM_IP_PORT` (default; Application and proxy load balancers), `GCE_VM_IP` (passthrough Network Load Balancers; no ports), `NON_GCP_PRIVATE_IP_PORT` (hybrid), `INTERNET_IP_PORT` / `INTERNET_FQDN_PORT` (regional external ALB origins), `GCE_VM_IP_DEDICATED_BACKEND`; global: `INTERNET_IP_PORT` or `INTERNET_FQDN_PORT`. Immutable.

**Membership** -- `endpoints` is the whole list, written as one set and changed in place; leave it empty when another controller owns membership. Each endpoint carries the fields its type needs (`instance`, `ipAddress`, `fqdn`, `port`), and every mismatch is rejected before deploy.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** | `subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpComputeInstance** | `endpoints[].instance` | `status.outputs.instance_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | The group's self link | `GcpBackendService.backends[].group` |
| `neg_name` | Name in GCP | Audit, dashboards |
| `neg_id` | Google's numeric id (zonal; empty globally) | Out-of-band tooling |
| `zone` | Zone of a zonal group; empty for a global one | Telling the scope from the outputs alone |
| `size` | Endpoints declared | Capacity review |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Zonal VM group for a regional load balancer** -- `GCE_VM_IP_PORT` VMs behind a regional or global Application Load Balancer. Start from the **Zonal VM Group for a Regional Load Balancer** preset.

**Hybrid connectivity group** -- on-premises addresses reached over Cloud VPN or Interconnect as load-balancer backends. Start from the **Hybrid Connectivity Group** preset.

**Global internet origin** -- an origin outside Google Cloud behind a global external Application Load Balancer. Start from the **Global Internet Origin** preset.

## Works With

- [**GCP Backend Service**](/cloud-catalog/gcp-backend-service) -- consumes the group's `self_link` as a backend
- [**GCP Region Network Endpoint Group**](/cloud-catalog/gcp-region-network-endpoint-group) -- serverless, PSC, and regional internet groups
- [**GCP Compute Instance**](/cloud-catalog/gcp-compute-instance) -- the VMs a zonal group names
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network), [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- the zonal group's network placement
