# AWS Managed Prefix List

Deploys a customer-managed prefix list — a named, versioned set of CIDR blocks that security-group rules, NACL rules, and route tables reference as a single `pl-` id. Define "our offices" or "partner ranges" once, reference the one id everywhere, then update the list once when the network changes instead of hunting rules across accounts. The list's name in AWS is `metadata.name`, entries are managed as the complete declarative set, and the address family is fixed for life.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Managed Prefix List** — the list with its address family (IPv4 or IPv6), its capacity (`maxEntries`), and its described CIDR entries managed in-line as the complete set: an entry removed from the manifest is removed at AWS, and this kind deliberately never uses the standalone entry resource that would fight the in-line form

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with EC2 VPC permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing — prefix lists are standalone objects with no prerequisite resources.

## Deploy

### Console

Open the deployment store, find **AWS Managed Prefix List**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: address family, capacity, and the described CIDR entries. Start from the **Office Networks** preset in the [Presets](#presets) tab for the canonical shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsManagedPrefixList
metadata:
  name: office-cidrs
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  addressFamily: IPv4
  maxEntries: 5
  entries:
    - cidr: 203.0.113.0/24
      description: hq office
    - cidr: 198.51.100.0/24
      description: branch office
    - cidr: 192.0.2.0/24
      description: vpn egress
```

```shell
planton apply -f aws-managed-prefix-list.yaml
```

This creates an IPv4 list holding three described office ranges with headroom for two more. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a prefix list. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**maxEntries is a quota decision, not a guess** — a security-group rule referencing this list consumes `maxEntries` slots of the rules-per-group quota (default 60) regardless of how many entries actually exist: a 20-capacity list in one rule spends a third of the group's quota while holding two entries. Size capacity to real growth, not round numbers. Resizes apply in place, and AWS orders capacity increases before entry changes (and decreases after), so a resize never transiently strands entries.

**The family is forever** — `addressFamily` replaces the list on change, which breaks every rule and route referencing the old `pl-` id. There is no in-place IPv4-to-IPv6 story; run dual-stack as two lists.

**Entries are the complete set** — the manifest is the single owner: an entry dropped here is removed at AWS on the next apply. Resist adding out-of-band entries through other tooling — every modification carries the list's current version and AWS rejects stale writers (`PrefixListVersionMismatch`); the single-owner in-line model avoids that race entirely.

**Descriptions are the audit trail** — an entry's `description` is the only record of WHY a CIDR is trusted, shown wherever the list is inspected. Keep it current. Description edits look expensive in plans because they are: AWS has no update-description call, so the provider removes and re-adds the entry across two round trips — expected plan noise, not drift.

**Delete consumers first** — AWS refuses to delete a list while any rule or route still references its `pl-` id, and a deleted list drains through `delete-in-progress`. Deleting this component after its consumers is the clean order; charts get it right by dependency.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — it is a leaf: its entries are literal CIDR blocks, and other resources reference it rather than the other way around.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `prefix_list_id` | The list's `pl-` id | Security-group rules (`prefixListIds`), NACL rules, and route tables — AWS resolves the current entries at enforcement time |
| `prefix_list_arn` | The list's ARN | RAM resource shares and IAM policies when sharing the list across accounts |

`owner_id` and `version` are also present — the owning account and AWS's per-change version counter, audit echoes rather than composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Office and VPN allowlist** — every office and VPN egress range in one described list, referenced from bastion and dashboard security groups. When an office moves, edit one entry and every rule follows. Start from the **Office Networks** preset.

**Partner ranges with generous headroom** — third-party CIDRs change on the partner's schedule, not yours, so capacity is sized generously up front: onboarding a partner becomes an entry edit, never a resize rippling quota costs through every referencing security-group rule. Descriptions name the partner. Start from the **Partner Ranges** preset.

**Shared network vocabulary across accounts** — one list owned centrally and shared via RAM (`prefix_list_arn`), so "corporate networks" means the same CIDRs in every account's security groups. Trades a central change gate for organization-wide consistency.

## Works With

- [**AWS Security Group**](/infra-catalog/aws-security-group) — rules take the `pl-` id in their `prefixListIds`, inheriting the list's current entries at enforcement time
- [**AWS VPC**](/infra-catalog/aws-vpc) — the network whose route tables and NACLs can target the same `pl-` id
