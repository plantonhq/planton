# AWS VPC Peering

Deploys one side of a VPC peering connection — private networking between two VPCs over AWS's backbone, same account or across accounts and regions, with no gateway, VPN, or single point of failure. The spec is a request-XOR-accept union: the request arm creates the peering from its VPC (auto-accepting when both VPCs share the account and region), and the accept arm adopts a pending connection by its `pcx-` id and accepts it — cross-account and cross-region topologies deploy one instance of each. Peering is non-transitive by design: each VPC pair peers explicitly.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one of two arms:

- **VPC Peering Connection** (request arm) — the peering from the requester VPC toward the peer VPC, with the owner account and region for cross-boundary peers, optional same-account auto-acceptance, and the DNS-resolution options managed in-line
- **Peering Connection Accepter** (accept arm) — adopts a pending connection by `vpcPeeringConnectionId` and accepts it, with the accepter side's DNS-resolution option. Its destroy is a no-op at AWS: it abandons management without deleting the peering — only the requester side deletes

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with EC2 VPC permissions. Cross-account topologies need one connection per side — the accept arm deploys with the peer account's credentials. Map connections as environment defaults, or specify them explicitly when creating the Infra Components.

### AWS Account

- Two VPCs with NON-OVERLAPPING CIDR blocks — AWS rejects overlapping peers outright, and CIDRs are immutable.
- Only for DNS resolution across the peering: `enableDnsHostnames` on both participating VPCs.

## Deploy

### Console

Open the deployment store, find **AWS VPC Peering**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, then either the request arm (requester and peer VPCs, cross-boundary identifiers, auto-acceptance, DNS options) or the accept arm (the pending connection's id). Start from the **Same-Account Peering** preset or the **Cross-Account Accepter** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpcPeering
metadata:
  name: app-to-data
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  request:
    vpcId:
      valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.vpc_id
    peerVpcId:
      valueFrom:
        kind: AwsVpc
        name: data-vpc
        fieldPath: status.outputs.vpc_id
    autoAccept: true
    requesterAllowRemoteVpcDnsResolution: true
    accepterAllowRemoteVpcDnsResolution: true
```

```shell
planton apply -f vpc-peering.yaml
```

This creates a same-account, same-region peering between the app and data VPCs, auto-accepted to ACTIVE in one deploy, with private DNS resolvable both ways. An Infra Job tracks the provisioning in real time.

### InfraChart

When the peering deploys alongside its VPCs in one chart, wire both VPC references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  request:
    vpcId:
      valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.vpc_id
    peerVpcId:
      valueFrom:
        kind: AwsVpc
        name: data-vpc
        fieldPath: status.outputs.vpc_id
    autoAccept: true
```

The InfraPipeline resolves the dependency graph, deploys both VPCs first, then peers them.

## Key Configuration

These are the most important decisions when configuring a VPC peering. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One arm per instance, two instances per cross-boundary peering** — `request` and `accept` are mutually exclusive, and auto-accept works same-account, same-region only: a cross-account or cross-region request sits `pending-acceptance` until a second instance of this kind, running the accept arm with the peer side's credentials, accepts it by the `pcx-` id. Every identity field on the request arm (`vpcId`, `peerVpcId`, `peerOwnerId`, `peerRegion`) is fixed for life.

**Peering is plumbing, not connectivity** — an ACTIVE peering moves zero packets until both sides add routes toward the peer CIDR in their route tables and open security groups. "Peering is up but nothing connects" is almost always a missing route, not a peering problem.

**The duplicate-connection trap** — requesting a peering for an already-peered VPC pair does not error: AWS returns the EXISTING connection's id. Declare each pair exactly once; a second instance for the same pair silently co-manages the first one's connection, and its destroy deletes it out from under the original.

**Deletes belong to the requester** — the accept arm's destroy abandons management while the peering stays ACTIVE and the peer keeps routing; only the request arm's destroy deletes the connection. Decommission cross-account peerings from the requester side, and treat the accept arm's removal as a management handoff, not a teardown.

**DNS options need an ACTIVE connection** — the DNS-resolution flags are a modification of an accepted peering, and AWS rejects them on a pending cross-account request. Deploy the requester without the accepter-side option, let the peer accept, then set each side's option from the instance that owns it. From the request arm, `accepterAllowRemoteVpcDnsResolution` is only settable on an auto-accepted (same-account) connection — validation enforces this.

**Plan address space before you need the peering** — overlapping CIDRs are rejected and CIDRs are immutable, so the fix for a collision is rebuilding a VPC. And because peering is non-transitive, ten fully meshed VPCs mean 45 connections with 90 route-table sides — past a handful of VPCs, a Transit Gateway is the saner hub.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `request.vpcId` | `status.outputs.vpc_id` |
| **AwsVpc** | `request.peerVpcId` | `status.outputs.vpc_id` |
| **AwsVpcPeering** | `accept.vpcPeeringConnectionId` | `status.outputs.peering_connection_id` |

Cross-account peers pass the literal peer `vpc-...` id (with `peerOwnerId`) instead of a reference, and cross-account accepters paste the `pcx-...` id shared by the requesting account.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `peering_connection_id` | The connection's id (`pcx-...`) | Route-table entries toward the peer CIDR; the accept-arm instance of a request→accept chain; the provider's import ID |
| `accept_status` | `active` or `pending-acceptance` after this side's deploy | Gating follow-on steps on whether the peer side still has to accept |

## Common Patterns

**Same-account pair** — both VPCs in one account and region, auto-accepted, private DNS resolvable both ways in a single deploy. The workhorse for app-tier-to-data-tier separation. Start from the **Same-Account Peering** preset.

**Cross-account handshake** — the requester instance creates the peering toward the partner's VPC (`peerOwnerId` plus the literal peer VPC id) and shares the resulting `pcx-` id; the partner deploys the accepter instance with their credentials to accept it and set their DNS option. Start from the **Cross-Account Accepter** preset for the accepting side.

**A few VPCs, then a hub** — peering scales socially, not technically: each new VPC must peer with every VPC it talks to, and every pair carries its own routes. Use peering while the pairs stay countable on one hand; move to a Transit Gateway when the mesh, not the traffic, becomes the work.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) — the two VPCs being peered, wired via `request.vpcId` and `request.peerVpcId`
- [**AWS Security Group**](/infra-catalog/aws-security-group) — must allow the peer CIDR's traffic before anything connects across the peering
- [**AWS Transit Gateway**](/infra-catalog/aws-transit-gateway) — the hub-and-spoke alternative once the peering mesh outgrows a handful of VPCs
