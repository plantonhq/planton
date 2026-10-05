# AWS Organization Account

Creates a member account of an AWS Organization, born in the right OU with the right contacts on file. One resource carries the account itself (name, root email, the pre-created bootstrap role, billing-info access), its placement in the OU tree, the billing/operations/security alternate contacts, the primary contact information, and opt-in region enablement. The delete contract is explicit: by default destroy removes the account from the organization and it lives on standalone; `closeOnDeletion: true` closes it instead, with AWS's ~90-day suspension window.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Member Account** -- the account itself (12-digit account ID), created into the organization under the parent named by `parentId` (or the organization root when unset), with the bootstrap role AWS pre-creates for the management account to assume
- **Alternate Contacts** -- one per category set under `alternateContacts` (billing, operations, security); AWS routes that category's communications to the contact, and removing an arm deletes it
- **Primary Contact** -- created only when `primaryContact` is set; the postal address and phone AWS keeps on file for the account
- **Region Enablements** -- one per `regions` entry, enabling (or explicitly disabling) an opt-in region for the account

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module whose credentials belong to the organization's MANAGEMENT account, with Organizations and Account Management permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An organization must already exist -- this resource creates accounts INTO it, from the management account.
- The contact and region arms require trusted access for AWS Account Management on the organization: `account.amazonaws.com` in the AWS Organization's `awsServiceAccessPrincipals`.
- The root email must not belong to any existing AWS account -- emails are unique across all of AWS, and closed accounts hold theirs through the ~90-day closure window.

## Deploy

### Console

Open the deployment store, find **AWS Organization Account**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: account name and root email, OU placement, the delete contract, contacts, and regions. Start from the **Workload Account** preset in the [Presets](#presets) tab for the standard production member account.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganizationAccount
metadata:
  name: workloads-prod
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  accountName: Workloads Prod
  email: aws+workloads-prod@acme-corp.com
  parentId:
    valueFrom:
      kind: AwsOrganizationalUnit
      name: workloads
      fieldPath: status.outputs.ou_id
  iamUserAccessToBilling: DENY
  alternateContacts:
    billing:
      name: Finance Team
      title: Billing Contact
      emailAddress: finance@acme-corp.com
      phoneNumber: "+1 555 0100"
    security:
      name: Security Team
      title: Security Contact
      emailAddress: security@acme-corp.com
      phoneNumber: "+1 555 0101"
```

```shell
planton apply -f aws-organization-account.yaml
```

This creates a member account inside the referenced workloads OU with root-only billing access and billing/security contacts on file from day one; account creation runs asynchronously and the module polls it to completion. An Infra Job tracks the provisioning in real time.

### InfraChart

When the account deploys alongside its OU in one chart, wire the parent via ValueFromRef:

```yaml
spec:
  region: us-east-1
  accountName: Workloads Prod
  email: aws+workloads-prod@acme-corp.com
  parentId:
    valueFrom:
      kind: AwsOrganizationalUnit
      name: workloads
      fieldPath: status.outputs.ou_id
```

The InfraPipeline resolves the dependency graph, creates the OU first, then creates the account inside it.

## Key Configuration

These are the most important decisions when configuring a member account. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The root email is forever** -- `email` is immutable (a change forces replacement, which is a full account lifecycle event) and unique across ALL of AWS; a closed account holds its email through the ~90-day closure window. Use plus-addressed per-account aliases (`aws+workloads-prod@example.com`) so a recreated account never fights its predecessor for the address.

**Choose the delete contract deliberately** -- `closeOnDeletion: false` (the default) removes the account from the organization on destroy: the account survives standalone, and AWS requires it to carry its own billing information for the removal to succeed. `closeOnDeletion: true` CLOSES the account: ~90 days of PENDING_CLOSURE suspension before permanent deletion, quota-limited to roughly 10% of your accounts per rolling 30 days, and unsupported for GovCloud. Neither path is a clean delete -- treat member accounts as long-lived, and reserve closure for true throwaways.

**OU placement governs the account from birth** -- `parentId` references an AWS Organizational Unit (or takes a literal `r-...`/`ou-...` ID, or stays unset for the root). Guardrail policies attached to that OU apply the moment the account exists. Changing `parentId` later MOVES the account in place -- the standard mechanism for promoting an account between lifecycle OUs.

**`roleName` is write-once** -- it names the bootstrap IAM role AWS pre-creates in the new account (default `OrganizationAccountAccessRole`). AWS exposes no API to read it back, so both engines deliberately ignore later changes; a genuine role change is an account rebuild. Decide before the first deploy.

**Billing access is immutable** -- `iamUserAccessToBilling` (unset = AWS's default ALLOW) cannot be changed without replacing the account. Production estates usually want `DENY` so only the root user sees billing; decide it up front.

**Contact arms have no unset semantics** -- the contact APIs are eventually-consistent Puts. The provider polls until writes are visible (a slow-but-green apply is normal), primary-contact delete is a no-op at AWS (the last-written contact stays on file), and clearing an optional primary-contact leaf leaves that leaf's last value in place. Overwrite deliberately, never by omission.

**Region enablement is slow and sticky** -- enabling or disabling an opt-in region takes up to ~60 minutes each way, and removing a `regions` entry does NOT opt the region back out (the provider's destroy is a no-op). When a region must actually close, keep the entry and set `enabled: false`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsOrganizationalUnit** | `parentId` | `status.outputs.ou_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `account_id` | The member account's 12-digit AWS account ID | Delegated-administrator registrations on the AWS Organization, account-targeted AWS Organization Policy attachments, and cross-account trust policies |

The remaining outputs -- `arn`, `state` (ACTIVE / SUSPENDED / PENDING_CLOSURE), and `govcloud_id` (set only when `createGovcloud` was used) -- are record values for audit and lifecycle visibility, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Workload account** -- a long-lived production account placed in a workloads OU (its guardrails apply from birth), `iamUserAccessToBilling: DENY`, and billing/security alternate contacts routed to the right teams instead of the root inbox. The default delete contract (removal, not closure) means destroying the resource never destroys a production account's data. Start from the **Workload Account** preset.

**Sandbox account** -- a disposable account whose destroy actually closes it: `closeOnDeletion: true`, placed in a sandbox OU with a spend-limiting SCP attached there. Respect the closure quota -- churn sandboxes deliberately, not hourly. Start from the **Sandbox Account** preset.

**Account promotion between OUs** -- because a `parentId` change moves the account in place, staged estates (sandbox OU, then workloads OU) promote accounts by editing one field -- the attached guardrails switch with the move, no account rebuild involved.

## Works With

- [**AWS Organization**](/infra-catalog/aws-organization) -- the organization this account is created into; the contact and region arms need `account.amazonaws.com` in its trusted-access list
- [**AWS Organizational Unit**](/infra-catalog/aws-organizational-unit) -- the OU the account is placed in, wired via the `parentId` reference
- [**AWS Organization Policy**](/infra-catalog/aws-organization-policy) -- guardrails attached to this account directly or inherited from its OU
