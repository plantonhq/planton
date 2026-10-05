# AWS REST API VPC Link

Deploys an API Gateway v1 VPC link — the Network Load Balancer attachment that REST API integrations route through to reach private services inside a VPC without exposing them to the internet. One link is shared by many APIs and owns its own network attachment, which is why it is its own kind rather than a field on the REST API. The target balancer is create-time immutable: AWS accepts exactly one NLB per link and has no update for it, so a different NLB means a new link. HTTP APIs use a different link resource that attaches to subnets directly — that is the AWS HTTP API VPC Link, and the two are not interchangeable.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VPC Link** — one API Gateway v1 VPC link named after `metadata.name`, fronting exactly one Network Load Balancer, with the optional `description` and Planton's resource tags
- **Network attachment** — the AWS-managed attachment to that balancer, built behind the link during creation; provisioning waits for it to reach AVAILABLE (up to about twenty minutes) before integrations can reference the link

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with API Gateway VPC-link control-plane permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An internal Network Load Balancer for the link to front, referenced via `targetArn`. REST API VPC links accept an NLB only — not an ALB, not Cloud Map (those are targets of the HTTP API link, a different resource).

## Deploy

### Console

Open the deployment store, find **AWS REST API VPC Link**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the region and target NLB. Start from the **NLB VPC Link** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiVpcLink
metadata:
  name: orders-nlb-link
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: Orders service internal NLB
  targetArn:
    valueFrom:
      kind: AwsNlb
      name: orders-internal-nlb
      fieldPath: status.outputs.load_balancer_arn
```

```shell
planton apply -f rest-api-vpc-link.yaml
```

This creates a VPC link in us-west-2 fronting the referenced internal NLB, ready for REST API integrations to reference once the attachment is AVAILABLE. An Infra Job tracks the provisioning in real time.

### InfraChart

When the link deploys alongside its balancer in one chart, wire the NLB reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  description: Orders service internal NLB
  targetArn:
    valueFrom:
      kind: AwsNlb
      name: orders-internal-nlb
      fieldPath: status.outputs.load_balancer_arn
```

The InfraPipeline resolves the dependency graph, deploys the NLB first, then creates the link fronting it.

## Key Configuration

These are the most important decisions when configuring a REST API VPC link. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The target NLB is a one-way door** — AWS has no update for a link's balancer, so changing `targetArn` replaces the link and issues a new `vpc_link_id`. Every integration referencing the old ID breaks at that moment. Migrate deliberately: stand up a new link on the new NLB, repoint integrations to the new ID, then destroy the old link.

**One link per NLB, shared by many APIs** — the link is the network attachment, not an API setting. Creating a link per API multiplies identical attachments for no isolation gain. Name the link after the backend it reaches (`orders-nlb-link`), not after any single API, and share its `vpc_link_id` across every REST API that needs that backend.

**NLB only, and internal is the point** — the spec takes exactly one balancer, and it must be an NLB. A public NLB is accepted but defeats the purpose of the link; keep the balancer internal so the API Gateway is the only public surface. Workloads behind an ALB or Cloud Map need an HTTP API and its own link kind — pointing a REST integration at an AWS HTTP API VPC Link fails at apply.

**Provisioning is slow by design** — creation blocks until the network attachment reaches AVAILABLE, which can take up to about twenty minutes. Budget pipeline timeouts for it, and expect the same wait on replacement.

**Destroy last** — integrations still referencing the link fail when it disappears. Drain or repoint every REST API using the link before destroying it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsNlb** | `targetArn` | `status.outputs.load_balancer_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpc_link_id` | The VPC link ID | REST API Gateway integrations set it as `vpcLinkId` with `connectionType: VPC_LINK` to route through the link |

`vpc_link_arn` is also exported — it serves IAM policy statements and operational tooling rather than composition wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private backend for one service** — a link fronting the service's internal NLB, referenced by a REST API's `HTTP_PROXY` integrations with `connectionType: VPC_LINK`. The API is the only public surface; the service never leaves the VPC. Start from the **NLB VPC Link** preset.

**Shared environment link** — one link per NLB, shared by every REST API in the environment that reaches that backend. It trades a wider blast radius on the single link for zero link sprawl and one attachment to operate — the production shape. Start from the **Shared Backend Link** preset.

**NLB migration** — because the target is immutable, moving APIs to a new balancer is a three-step dance: create a second link on the new NLB, repoint each API's integrations to the new `vpc_link_id`, then destroy the old link once nothing references it. Both links coexist during the cutover, so the migration is incremental and reversible until the final destroy.

## Works With

- [**AWS NLB**](/infra-catalog/aws-nlb) — the internal balancer the link fronts, wired via `targetArn`
- [**AWS REST API Gateway**](/infra-catalog/aws-rest-api-gateway) — integrations route through the link by setting `connectionType: VPC_LINK` and this link's `vpc_link_id`
- [**AWS HTTP API VPC Link**](/infra-catalog/aws-http-api-vpc-link) — the API Gateway v2 counterpart that attaches to subnets (ALB, NLB, or Cloud Map); not interchangeable with this link
