# GCP PSC Service Attachment

Publishes a service through Private Service Connect -- the producer half. A service attachment sits in front of your internal load balancer (the regional forwarding rule of an internal passthrough Network Load Balancer or an internal Application Load Balancer) and lets consumers in other VPC networks -- other projects, other organizations -- reach it through a PSC endpoint in their own network, over private IPs, with no VPC peering and no overlapping-range constraints. You decide who may connect and how many endpoints each consumer may open.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Service Attachment** -- a `compute.ServiceAttachment` in the chosen region, pointing at the target forwarding rule, translating consumer traffic into the PSC NAT subnets, and enforcing the connection preference with its accept and reject lists, reconciliation, PROXY protocol, Cloud DNS domain, and propagated-connection limit

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the producer project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Networks

- **An internal load balancer to publish** -- a regional `GcpGlobalForwardingRule` (scheme `INTERNAL` or `INTERNAL_MANAGED`) in the same region.
- **PSC NAT subnets** -- one or more `GcpSubnetwork` with `purpose: PRIVATE_SERVICE_CONNECT` in the producer VPC and region.

## Deploy

### Console

Open the deployment store, find **GCP PSC Service Attachment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Publish an Internal Load Balancer (Manual Acceptance)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPscServiceAttachment
metadata:
  name: orders-db-psc
  org: acme-corp
  env: prod
spec:
  region: us-central1
  targetService:
    valueFrom:
      kind: GcpGlobalForwardingRule
      name: orders-db-ilb
      fieldPath: status.outputs.self_link
  natSubnets:
    - valueFrom:
        kind: GcpSubnetwork
        name: orders-psc-nat
        fieldPath: status.outputs.subnetwork_self_link
  connectionPreference: ACCEPT_MANUAL
  consumerAcceptLists:
    - projectId:
        value: consumer-analytics-prod
      connectionLimit: 10
  deletionPolicy: PREVENT
```

```shell
planton apply -f psc-service-attachment.yaml
```

This publishes the orders database's internal load balancer to one consumer project with up to ten endpoints. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the attachment to the load balancer and NAT subnet deployed in the same InfraPipeline; the consumer's PSC endpoint (a regional `GcpGlobalForwardingRule` with an empty scheme) references this attachment's `self_link` as its `target`.

## Key Configuration

These are the most important decisions when configuring a service attachment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Who may connect** -- `connectionPreference: ACCEPT_MANUAL` admits only `consumerAcceptLists` (each entry a `GcpProject` reference, a `GcpVpcNetwork` reference, or an endpoint URL, with a `connectionLimit`); `ACCEPT_AUTOMATIC` admits everyone but `consumerRejectLists`. `reconcileConnections: true` makes list edits apply to already-connected endpoints.

**NAT capacity** -- every connected endpoint consumes addresses from `natSubnets`; add subnets to grow capacity.

**PROXY protocol** -- `enableProxyProtocol: true` gives backends the consumer's original address, but only backends that speak the protocol can read it.

**Destroy semantics** -- `deletionPolicy: PREVENT` guards a service other teams depend on; deleting the attachment cuts every consumer off.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpGlobalForwardingRule** | `targetService` | `status.outputs.self_link` |
| **GcpSubnetwork** | `natSubnets[]` | `status.outputs.subnetwork_self_link` |
| **GcpProject** | `consumerAcceptLists[].projectId`, `consumerRejectLists[]` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `consumerAcceptLists[].network` | `status.outputs.network_self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | The attachment's self link | A consumer `GcpGlobalForwardingRule`'s `target` (PSC endpoint) |
| `attachment_name` | Name in GCP | Audit, dashboards |
| `region` | Region | Placing consumer endpoints |
| `fingerprint` | Server-computed fingerprint | Out-of-band updates |
| `connected_endpoints_count` | Connected consumer endpoints at provisioning time | Capacity review |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Publish an internal load balancer with manual acceptance** -- The production posture: named consumer projects with connection limits and a reject list. Start from the **Publish an Internal Load Balancer (Manual Acceptance)** preset.

**Publish to any consumer** -- `ACCEPT_AUTOMATIC` for an internal platform service every project in the organization may reach. Start from the **Publish to Any Consumer (Automatic Acceptance)** preset.

## Works With

- [**GCP Global Forwarding Rule**](/cloud-catalog/gcp-global-forwarding-rule) -- the producer's internal load balancer (regional arm, `INTERNAL` scheme) this attachment publishes; and the consumer's PSC endpoint (regional arm, empty scheme) that targets it
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- the PSC NAT subnets
- [**GCP Backend Service**](/cloud-catalog/gcp-backend-service) -- the regional backend service behind the producer's forwarding rule
