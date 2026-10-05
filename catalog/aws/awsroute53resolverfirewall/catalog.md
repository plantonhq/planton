# AWS Route 53 Resolver DNS Firewall

Deploys a Route 53 Resolver DNS Firewall rule group — the filtering policy for DNS queries leaving your VPCs — with its domain lists, filtering rules, and VPC associations managed in-line. Rules match an owned domain list, an external or AWS-managed list by ID, or a DNS threat class (DGA, dictionary DGA, DNS tunneling), and answer with ALLOW, ALERT, or BLOCK — including a CNAME sinkhole for blocked lookups. Associating the group to a VPC is what turns the policy on there; no endpoints or agents are involved.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DNS Firewall Rule Group** — the container the rules and associations hang off; its name is fixed for life (the group's only update path is tags)
- **Firewall Domain Lists** — one per `domainLists` entry, owned by this group's lifecycle and keyed by name. Both engines store every domain as a trailing-dot FQDN, matching AWS's canonical form
- **Firewall Rules** — one per `rules` entry, bound to the group in ascending priority order. For OVERRIDE block responses the module pins the record type to CNAME, the only value AWS accepts
- **Rule Group VPC Associations** — one per `vpcAssociations` entry, putting the policy into effect for that VPC at the stated evaluation priority

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Route 53 Resolver permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- The VPCs to filter (`vpcAssociations[].vpcId`) — nothing else; DNS Firewall evaluates queries inside the Resolver, so there are no endpoints, ENIs, or agents to stand up.
- Only for rules matching an AWS-managed or external list: the list's ID (`rslvr-fdl-...`). Managed-list IDs are account- and region-specific — resolve the ID once with `aws route53resolver list-firewall-domain-lists` and pass it as `domainListId`.

## Deploy

### Console

Open the deployment store, find **AWS Route 53 Resolver DNS Firewall**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, the owned domain lists, the rules with their priorities and actions, and the VPC associations. Start from the **Blocklist with Sinkhole** preset or the **Advanced Threat Detection (Alert First)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53ResolverFirewall
metadata:
  name: egress-dns-policy
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  domainLists:
    - name: blocked-domains
      domains:
        - malware.example.
        - phishing.example.
  rules:
    - name: sinkhole-blocked
      priority: 100
      action: BLOCK
      domainListName: blocked-domains
      blockResponse: OVERRIDE
      blockOverrideDomain: sinkhole.example.com
      blockOverrideTtl: 300
  vpcAssociations:
    - name: app-vpc
      vpcId:
        valueFrom:
          kind: AwsVpc
          name: app-vpc
          fieldPath: status.outputs.vpc_id
      priority: 200
```

```shell
planton apply -f resolver-firewall.yaml
```

This creates a rule group with one owned blocklist whose matches answer with a CNAME to your sinkhole host instead of resolving, active in the app VPC. An Infra Job tracks the provisioning in real time.

### InfraChart

When the firewall deploys alongside the VPC it filters in one chart, wire the VPC reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  domainLists:
    - name: blocked-domains
      domains:
        - malware.example.
  rules:
    - name: block-bad-domains
      priority: 100
      action: BLOCK
      domainListName: blocked-domains
  vpcAssociations:
    - name: app-vpc
      vpcId:
        valueFrom:
          kind: AwsVpc
          name: app-vpc
          fieldPath: status.outputs.vpc_id
      priority: 200
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then puts the firewall policy into effect on it.

## Key Configuration

These are the most important decisions when configuring a DNS Firewall rule group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Priorities are the policy** — evaluation is ascending priority, first match wins: an ALLOW at 100 defeats a BLOCK at 200 for the same domain. Leave gaps (100, 200, 300) so rules can be inserted later without renumbering, and remember that a VPC carrying several rule-group associations evaluates the GROUPS in ascending association priority the same way.

**A rule's match source is fixed for life** — switching a rule between list-matched and threat-matched replaces the rule at the provider. Action, priority, and the block-response shape all update in place, so tune those freely; changing what a rule matches means accepting a replace.

**ALERT before BLOCK for threat rules** — Advanced protection (`dnsThreatProtection` with its paired `confidenceThreshold`) is heuristic: DGA and DNS-tunneling detection surprise even at HIGH confidence. Run new threat rules as ALERT alongside query logging until the hit pattern is understood, then flip the action to BLOCK — an in-place edit.

**Shape the block response deliberately** — NODATA (the default) says the domain exists but has no records, NXDOMAIN says it does not exist, and OVERRIDE substitutes a CNAME to your sinkhole host (`blockOverrideDomain` + `blockOverrideTtl`, both required together). Sinkholes turn silent drops into observable connections to a host you control — the difference between "it stopped" and "we saw who asked".

**Mutation protection blocks its own destroy** — an association with `mutationProtection: ENABLED` refuses deletion, including the declarative destroy of this very resource. Enable it only on associations whose lifecycle is deliberately manual, and disable it before decommissioning.

**Fail-open is a VPC decision, not a rule-group decision** — what happens when the firewall cannot evaluate a query (fail closed vs fail open) is configured per VPC with the VPC's other resolver settings, deliberately not on this kind. A rule group associated to ten VPCs must not fight over one VPC's toggle.

**Domain matching and canonical form** — a bare domain entry matches itself and all subdomains; a `*.` prefix matches subdomains only. AWS stores every firewall domain (list entries and the OVERRIDE record's value) as a trailing-dot FQDN — both engines append the dot when absent, so write whichever form you like and the plan stays clean.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcAssociations[].vpcId` | `status.outputs.vpc_id` |

Rules can also reference an external or AWS-managed domain list, but those travel as literal `rslvr-fdl-...` IDs in `domainListId`, not as typed references.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_group_id` | The rule group's id (`rslvr-frg-...`) | Addressing the group in AWS tooling; the provider's import ID |
| `rule_group_arn` | The rule group's ARN | IAM policies scoped to this group |
| `domain_list_ids` | AWS-generated domain list IDs (`rslvr-fdl-...`) keyed by list name | Another rule group's rules referencing an owned list as an external `domainListId` |

`share_status`, `association_ids`, and `rule_match_ids` are also exported, but they are operational echoes — the group's RAM sharing state and the AWS-generated association and per-rule match IDs — kept for audits and composite imports rather than composition.

## Common Patterns

**Blocklist with a sinkhole** — the starter policy: an owned blocklist whose matches answer with a CNAME to your sinkhole host instead of resolving, so blocked lookups become observable. Start from the **Blocklist with Sinkhole** preset.

**Threat detection, alert first** — DNS Firewall Advanced heuristics (algorithmically generated domains, DNS tunneling) in ALERT mode, run alongside query logging until the hit pattern is understood, then flipped to BLOCK in place. Start from the **Advanced Threat Detection (Alert First)** preset.

**One policy, many VPCs** — a single rule group associated to every VPC in the environment. Policy edits land everywhere at once, which is the point and the risk: an over-broad BLOCK rule takes effect fleet-wide, so stage rule changes through ALERT and watch the query logs before tightening.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) — the VPCs the policy filters, wired via `vpcAssociations[].vpcId`
- [**AWS Route 53 Resolver Query Logging**](/infra-catalog/aws-route53-resolver-query-log) — the only way to see what fired: query logs carry the firewall's rule verdicts, and the firewall itself logs nothing
- [**AWS Route 53 Resolver Endpoint**](/infra-catalog/aws-route53-resolver-endpoint) — hybrid-DNS forwarding for the same VPCs; the firewall evaluates queries before they forward
