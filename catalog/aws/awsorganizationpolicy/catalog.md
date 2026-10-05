# AWS Organization Policy

Creates an AWS Organizations policy -- a service control policy or any of its twelve sibling types -- together with its attachments to the organization root, OUs, and member accounts. The document is structured JSON in the type's own syntax, attachments fold into the same resource as pure edges, and effects are inheritance-scoped: attached to the root it governs everything, attached to an OU it governs that subtree, attached to an account, just that account. The policy's type must already be enabled on the organization before any attachment succeeds.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Policy** -- the policy itself (AWS `p-...` ID) with its name, type (SCP by default), structured document, and description; name, content, and description update in place, and the provider suppresses JSON-equivalent content diffs
- **Policy Attachments** -- one per `attachments` entry, each binding the policy to one root, OU, or member account; both leaves are immutable, so changing a target detaches and re-attaches

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module whose credentials belong to the organization's MANAGEMENT account, with Organizations permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An organization with ALL features must exist, with this policy's type in its `enabledPolicyTypes` -- enable `SERVICE_CONTROL_POLICY` there before shipping SCPs.
- AWS-managed policies (like FullAWSAccess) cannot be adopted by this resource -- the provider refuses to import them.

## Deploy

### Console

Open the deployment store, find **AWS Organization Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: the policy name and type, the document, and its attachment targets. Start from the **Deny-Leave SCP** preset in the [Presets](#presets) tab for the foundational guardrail.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganizationPolicy
metadata:
  name: deny-leave-org
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  policyName: Deny Leave Organization
  type: SERVICE_CONTROL_POLICY
  description: No member account can remove itself from the organization
  content:
    Version: "2012-10-17"
    Statement:
      - Sid: DenyLeaveOrganization
        Effect: Deny
        Action: organizations:LeaveOrganization
        Resource: "*"
  attachments:
    - targetId:
        valueFrom:
          kind: AwsOrganization
          name: acme-organization
          fieldPath: status.outputs.root_id
```

```shell
planton apply -f aws-organization-policy.yaml
```

This creates an SCP denying `organizations:LeaveOrganization` and attaches it at the organization root, so no member account can walk out of governance (SCPs never bind the management account, which can still remove accounts deliberately). An Infra Job tracks the provisioning in real time.

### InfraChart

When the policy deploys alongside the tree it governs, wire attachment targets via ValueFromRef:

```yaml
spec:
  region: us-east-1
  policyName: Workloads Guardrails
  content:
    Version: "2012-10-17"
    Statement:
      - Sid: DenyLeaveOrganization
        Effect: Deny
        Action: organizations:LeaveOrganization
        Resource: "*"
  attachments:
    - targetId:
        valueFrom:
          kind: AwsOrganizationalUnit
          name: workloads
          fieldPath: status.outputs.ou_id
```

The InfraPipeline resolves the dependency graph, creates the OU first, then attaches the policy to it.

## Key Configuration

These are the most important decisions when configuring an organization policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The type gate fails at attach time, not create time** -- a policy of a type missing from the organization's `enabledPolicyTypes` creates fine, and then every attachment errors. The fix is the AWS Organization resource's `enabledPolicyTypes`, not a retry. The type itself (`type`, default `SERVICE_CONTROL_POLICY`) is immutable -- changing it replaces the policy.

**SCPs cap, they never grant** -- an SCP is an outer boundary on what IAM inside the subtree can allow; no SCP ever grants a permission. AWS's own FullAWSAccess policy stays attached unless deliberately replaced, which is what keeps a new SCP from locking anyone out by mere existence. Test new guardrails against an empty OU or a sandbox account before attaching high.

**Attachment height is blast radius** -- a wrong deny at the root locks the entire estate out of an API (management account excluded -- SCPs never bind it). Prefer first-level OU attachments, reserving the root for universal invariants like the deny-leave guardrail. Each attachment target is immutable, so moving a guardrail is a detach-and-attach, which the module performs when you edit the entry.

**Content changes propagate in seconds, everywhere** -- document updates apply in place with no staged rollout: every account under every attachment sees the new policy almost immediately. Pair risky edits with a narrow attachment first (one OU, one sandbox account), then widen once proven.

**Destroy means detach and delete** -- the provider's skip-destroy escape hatches are deliberately not modeled. Destroying this resource detaches every target and deletes the policy; the guardrail is gone for the whole subtree in one operation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsOrganizationalUnit** | `attachments[].targetId` (OU scope, the default wiring) | `status.outputs.ou_id` |
| **AwsOrganization** | `attachments[].targetId` (root scope) | `status.outputs.root_id` |
| **AwsOrganizationAccount** | `attachments[].targetId` (single-account scope) | `status.outputs.account_id` |

### What This Kind Provides

`status.outputs` carries the policy's identity -- `policy_id` (`p-...`) and `arn`. These are record values for audit and import rather than composition inputs: attachments fold into this resource itself, so no downstream Infra Component consumes a policy ID by reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Deny-leave foundation** -- the first SCP of every governed organization: deny `organizations:LeaveOrganization` at the root, so no member account's root user can exit governance with one API call. It forbids one administrative action and cannot break workloads -- the safest possible first guardrail. Common companions in the same document: deny `account:CloseAccount` and deny disabling CloudTrail. Start from the **Deny-Leave SCP** preset.

**Region guardrail** -- deny everything outside an approved region list (`aws:RequestedRegion`), with a `NotAction` exemption for the global services (IAM, Organizations, Route53, CloudFront, STS, Support) that would otherwise break under the fence. Attach it to one workloads OU first; widen to the root only after a soak, because a wrong region fence at the root is the widest possible blast radius. Start from the **Region Guardrail SCP** preset.

**Beyond SCPs** -- the same kind carries tag policies (enforce tagging standards), backup policies, declarative EC2 policies, and the rest of the thirteen types: one document in the type's own syntax, the same attachment mechanics, gated by the same `enabledPolicyTypes` list on the organization.

## Works With

- [**AWS Organization**](/infra-catalog/aws-organization) -- enables this policy's type via `enabledPolicyTypes`; its `root_id` is the target for root-scoped attachments
- [**AWS Organizational Unit**](/infra-catalog/aws-organizational-unit) -- the usual attachment target; everything beneath the OU inherits the policy
- [**AWS Organization Account**](/infra-catalog/aws-organization-account) -- the target for single-account exceptions and sandbox soak tests
