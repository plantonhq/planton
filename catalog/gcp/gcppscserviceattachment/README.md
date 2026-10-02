# GCP PSC Service Attachment

Publishes a service through Private Service Connect — the PRODUCER half. A service attachment sits in front of your internal load balancer (the regional forwarding rule of an internal passthrough Network Load Balancer or an internal Application Load Balancer) and lets consumers in other VPC networks — other projects, other organizations — reach it through a PSC endpoint in their own network, over private IPs, with no VPC peering and no overlapping-range constraints. Consumer traffic is translated into dedicated NAT subnets in your VPC; you decide who may connect.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Service attachment** -- the `compute_service_attachment` in your region, pointing at `targetService` (your internal load balancer's forwarding rule), translating consumer traffic into `natSubnets`, and enforcing `connectionPreference` with its accept and reject lists

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the producer project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Networks

- **An internal load balancer to publish** -- a regional `GcpGlobalForwardingRule` with `loadBalancingScheme: INTERNAL` (passthrough) or `INTERNAL_MANAGED` (internal ALB) in the same region; reference its `self_link` output.
- **One or more PSC NAT subnets** -- `GcpSubnetwork` with `purpose: PRIVATE_SERVICE_CONNECT` in the producer VPC and region, sized for the consumer endpoint count (each connected endpoint consumes NAT addresses).
- **The consumers** -- each one creates a regional `GcpGlobalForwardingRule` with an empty `loadBalancingScheme` whose `target` is this attachment's `self_link`.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPscServiceAttachment
metadata:
  name: orders-db-psc
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
```

```shell
planton apply -f psc-service-attachment.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | `string` | Region of the attachment, the target forwarding rule, and the NAT subnets. Immutable. |
| `targetService` | `StringValueOrRef` | The producer's internal load balancer: a regional `GcpGlobalForwardingRule` reference or its self link. |
| `natSubnets` | `[]StringValueOrRef` | PSC NAT subnets (`purpose: PRIVATE_SERVICE_CONNECT`) as `GcpSubnetwork` references or self links. At least one; compared as a set. |
| `connectionPreference` | `string` | `ACCEPT_AUTOMATIC` (anyone connects) or `ACCEPT_MANUAL` (only the accept list). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The producer project. Immutable. |
| `attachmentName` | `string` | `metadata.name` | Name in GCP. Immutable. |
| `description` | `string` | — | Free text. |
| `consumerAcceptLists` | `[]object` | — | `ACCEPT_MANUAL` only: each entry names exactly one of `projectId` (`GcpProject` reference), `network` (`GcpVpcNetwork` reference), or `endpointUrl`, with a `connectionLimit` ≥ 1. |
| `consumerRejectLists` | `[]StringValueOrRef` | — | Projects refused even under `ACCEPT_AUTOMATIC`. |
| `reconcileConnections` | `bool` | Google's default (`false`) | `true` moves existing ACCEPTED endpoints to REJECTED when their project lands on the reject list (and vice versa); `false` touches PENDING endpoints only. Sent only when set. |
| `enableProxyProtocol` | `bool` | `false` | PROXY protocol on connections, so backends see the consumer's original address. Backends must speak it. |
| `domainNames` | `[]string` | — | At most one Cloud DNS domain for connected endpoints, with a trailing dot (`p.mycompany.com.`). Immutable. |
| `propagatedConnectionLimit` | `int32` | Google's default (250) | Consumer spokes a connected endpoint may be propagated to through Network Connectivity Center. An explicit `0` is sent as 0. |
| `showNatIps` | `bool` | `false` | Show NAT IPs of connected endpoints (Google's API currently ignores the flag). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- **`consumerAcceptLists`** is rejected unless `connectionPreference` is `ACCEPT_MANUAL`.
- Each accept-list consumer names **exactly one** of `projectId`, `network`, `endpointUrl`.
- **`attachmentName`** matches `^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`; **`domainNames`** entries end with a dot.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | The attachment's self link -- a consumer forwarding rule's `target` |
| `attachment_name` | `string` | Name in GCP |
| `region` | `string` | Region |
| `fingerprint` | `string` | Server-computed fingerprint |
| `connected_endpoints_count` | `string` | Consumer endpoints connected at provisioning time |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Start with `ACCEPT_MANUAL`.** `ACCEPT_AUTOMATIC` lets any project in Google Cloud connect; only the reject list stands in the way.
- **Deleting the attachment cuts every consumer off.** `PREVENT` is the guard for a service other teams depend on.
- **The scheme of the consumer's forwarding rule is empty**, its `target` is this `self_link`, and it lives in the consumer's region; the attachment's region must match the producer load balancer's.
- **Cost**: the attachment itself is free; connected consumer endpoints bill forwarding-rule hours and PSC data processing on the consumer side.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpGlobalForwardingRule](/docs/catalog/gcp/gcpglobalforwardingrule) — the producer's internal load balancer this publishes, and the consumer's PSC endpoint that targets it
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the PSC NAT subnets
- [GcpBackendService](/docs/catalog/gcp/gcpbackendservice) — the regional backend service behind the producer's forwarding rule

## Additional Resources

- [Publish services using Private Service Connect](https://cloud.google.com/vpc/docs/about-vpc-hosted-services)
- [Manage access to published services](https://cloud.google.com/vpc/docs/manage-services-private-service-connect)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
