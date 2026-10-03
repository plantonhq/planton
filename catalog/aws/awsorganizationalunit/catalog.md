# AWS Organizational Unit

Creates an organizational unit in an AWS Organization's account tree -- the container member accounts are grouped into and guardrail policies attach to. The parent is a reference: the organization's root for a first-level OU (the default wiring), another OU for nesting, or a literal `r-...`/`ou-...` ID for pre-existing trees. The parent is immutable -- AWS moves accounts between OUs, never OUs themselves -- so the tree shape is a decision to settle before populating it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Organizational Unit** -- one OU (AWS `ou-...` ID) under the parent named by `parentId`, carrying the display name from `ouName`. Creation retries through the organization's finalization window, so an OU deployed immediately after its organization lands cleanly.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module whose credentials belong to the organization's MANAGEMENT account, with Organizations permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An organization must already exist -- an OU cannot exist outside one, and `parentId` is required. The AWS Organization resource's `root_id` output is the default parent wiring.

## Deploy

### Console

Open the deployment store, find **AWS Organizational Unit**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: the OU's display name and its parent. Start from the **Workloads OU** preset in the [Presets](#presets) tab for the standard first-level container.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganizationalUnit
metadata:
  name: workloads
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  ouName: Workloads
  parentId:
    valueFrom:
      kind: AwsOrganization
      name: acme-organization
      fieldPath: status.outputs.root_id
```

```shell
planton apply -f aws-organizational-unit.yaml
```

This creates a first-level Workloads OU directly under the organization root, with the parent resolved from the referenced AWS Organization -- no hand-copied IDs. An Infra Job tracks the provisioning in real time.

### InfraChart

When the OU deploys alongside its organization in one chart, wire the parent via ValueFromRef:

```yaml
spec:
  region: us-east-1
  ouName: Workloads
  parentId:
    valueFrom:
      kind: AwsOrganization
      name: acme-organization
      fieldPath: status.outputs.root_id
```

The InfraPipeline resolves the dependency graph, creates the organization first, then hangs the OU off its root.

## Key Configuration

These are the most important decisions when configuring an organizational unit. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The parent is immutable -- design the tree before populating it** -- AWS has no API to move an OU: changing `parentId` forces replacement, and a populated OU cannot be deleted. Accounts move freely between OUs; OUs do not. Settle the tree shape (Workloads, Security, Sandbox, and their subdivisions) before accounts and policies pile in.

**Three parent wirings** -- a first-level OU references an AWS Organization (the default resolves its `root_id`); a nested OU references its parent OU's `ou_id`; a literal `parentId: {value: "ou-..."}` carries trees built outside Planton. Validation enforces the provider's own `r-...`/`ou-...` pattern on literals.

**Keep the tree shallow** -- AWS allows five OU levels under the root, but policy-inheritance reasoning gets hard long before the quota does. Prefer attaching guardrails high (root or first-level OUs) with exceptions low, over encoding every distinction as another level.

**`ouName` is display, not identity** -- spaces and arbitrary characters are legal ("Core Services"), which is why it is an explicit field rather than `metadata.name`, and renames apply in place. The stable identity is the `ou_id` output.

**Deletion requires the OU to be empty** -- move member accounts out and remove child OUs and policy attachments first; AWS refuses otherwise. A destroy failure here is AWS's ordering contract, not a module defect.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsOrganization** | `parentId` (first-level OU) | `status.outputs.root_id` |
| **AwsOrganizationalUnit** | `parentId` (nested OU) | `status.outputs.ou_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ou_id` | The OU's AWS-generated ID (`ou-...`) | `parentId` of AWS Organization Account and nested AWS Organizational Unit resources; target of AWS Organization Policy attachments |

The `arn` output is a record value for audit and import, not a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**First-level function OUs** -- Workloads, Security, Sandbox, Infrastructure: one OU per estate function directly under the root, each the attachment point for that function's guardrails. Production and staging accounts live under Workloads; a spend-limiting SCP lives on Sandbox. Start from the **Workloads OU** preset.

**Per-team nesting** -- a second-level OU under a Workloads-style container when a team needs its own guardrail scope: an SCP attached here governs only that team's accounts. The trade is depth -- every level added is another layer of inheritance to reason about. Start from the **Nested Team OU** preset.

**Adopting an existing tree** -- literal `parentId` values (`ou-...`) let new Planton-managed OUs and accounts hang off a tree built outside Planton, so migration proceeds branch by branch instead of all at once.

## Works With

- [**AWS Organization**](/infra-catalog/aws-organization) -- the organization whose root first-level OUs hang off, wired via the `parentId` reference
- [**AWS Organization Account**](/infra-catalog/aws-organization-account) -- member accounts placed in this OU (`parentId` → this OU's `ou_id` output)
- [**AWS Organization Policy**](/infra-catalog/aws-organization-policy) -- guardrails attached to this OU, inherited by everything beneath it
