# AWS Organization

Creates the AWS Organization for the deploying account, making that account the management account of a multi-account estate. Every org-wide lever folds into this one resource: the feature level (ALL or consolidated billing only), trusted service access per AWS service, the policy types enabled on the root, delegated administrator registrations, the organization's single resource-based delegation policy, and centralized root-access management. Exactly one organization exists per account, and deleting this resource deletes the entire organization -- AWS refuses until every member account, OU, and policy is gone.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Organization** -- the organization itself (AWS `o-...` ID), created with the chosen feature set, trusted service principals, and enabled policy types; the deploying account becomes the management account
- **Delegated Administrator registrations** -- one per `delegatedAdministrators` entry, each naming a member account as the org-wide administrator for one AWS service
- **Resource Policy** -- created only when `resourcePolicy` is set; the organization's single resource-based delegation policy (AWS keeps exactly one per organization, so this arm is that singleton)
- **IAM Organizations Features** -- created only when `rootAccessManagement` is set; enables centralized root-credentials management and/or privileged root sessions across member accounts

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module whose credentials belong to the account you intend to become the management account, with permission to call the Organizations APIs (and IAM's organization features when `rootAccessManagement` is used). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- The deploying account must NOT already belong to an organization -- creating this resource performs CreateOrganization and makes it the management account.
- The advanced arms -- `awsServiceAccessPrincipals`, `enabledPolicyTypes`, `delegatedAdministrators`, `resourcePolicy` -- all require the ALL feature set; spec validation rejects them under a consolidated-billing organization before AWS ever would.
- `rootAccessManagement` additionally requires `iam.amazonaws.com` in `awsServiceAccessPrincipals` (also validation-enforced).

## Deploy

### Console

Open the deployment store, find **AWS Organization**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: feature set, trusted service principals, policy types, and the optional delegation arms. Start from the **All-Features Foundation** preset in the [Presets](#presets) tab for the standard multi-account starting point.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganization
metadata:
  name: acme-organization
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  featureSet: ALL
  awsServiceAccessPrincipals:
    - cloudtrail.amazonaws.com
    - config.amazonaws.com
    - account.amazonaws.com
  enabledPolicyTypes:
    - SERVICE_CONTROL_POLICY
    - TAG_POLICY
```

```shell
planton apply -f aws-organization.yaml
```

This creates an all-features organization with trusted access for CloudTrail, Config, and Account Management, and with SCP and tag-policy types enabled on the root -- attachments work immediately. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an organization. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Feature set is effectively a one-way door** -- Unset defaults to ALL, the level every advanced arm requires. Upgrading `featureSet` from `CONSOLIDATED_BILLING` to `ALL` is an in-place update (AWS's EnableAllFeatures). The downgrade REPLACES the resource -- delete-and-recreate of the ENTIRE organization, which AWS only permits once every member account, OU, and policy is gone. Treat a downgrade as an organization rebuild project, never a settings change.

**Trusted service access has exactly one home** -- `awsServiceAccessPrincipals` is where service access lives; the provider's own documentation warns that managing the same principal anywhere else produces a perpetual diff. Enable principals deliberately: trusted access lets the named service create service-linked roles in every member account. Removing an entry disables that service's access. If member-account contacts and regions will be managed centrally (the AWS Organization Account component's settings arms), `account.amazonaws.com` must be in this list.

**Policy types gate the whole policy family** -- A type must appear in `enabledPolicyTypes` before any AWS Organization Policy of that type can attach anywhere in the org. Enable `SERVICE_CONTROL_POLICY` before shipping SCPs. On update, disables are applied before enables, and each waits for AWS to confirm the type's state.

**Delegated administrators are immutable pairs** -- Each entry is `{accountId, servicePrincipal}` and both leaves are immutable: changing either deregisters and re-registers. The account must already be a member of the organization, so delegations usually land in a second apply after member accounts exist. The classic use is delegating GuardDuty or Config administration to a dedicated security account so daily governance work leaves the management account.

**The resource policy is a singleton** -- AWS keeps exactly ONE resource-based delegation policy per organization; `resourcePolicy` is that singleton (PutResourcePolicy upserts it, removing the arm deletes it). Use it to delegate specific organization-management reads or actions to member accounts -- for example letting a tooling account list OUs -- without handing out management-account credentials.

**Root-access management cuts both ways** -- `rootAccessManagement` with `RootCredentialsManagement` centrally removes long-lived root passwords and keys from member accounts (the lock-down posture); `RootSessions` allows short-lived privileged root sessions for the rare tasks that genuinely need root. Destroying the arm DISABLES every enabled feature, making member-account root credentials locally manageable again -- an org-wide security posture change hiding in a field removal.

**Destroy deletes the organization** -- There is no "detach" semantics: destroying this resource calls DeleteOrganization. AWS refuses while members, OUs, or policies exist, so teardown order is structural -- policies and OUs (their own components) first, member accounts removed, then this resource.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. It is the root of the multi-account estate -- everything else in the Organizations family references it, not the other way around.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `root_id` | The organization root's ID (`r-...`), the top of the OU tree | Parent for first-level AWS Organizational Unit resources and target for root-scoped AWS Organization Policy attachments |
| `organization_id` | The organization's AWS-generated ID (`o-...`) | `aws:PrincipalOrgID` conditions in IAM and resource policies that trust the whole org |
| `management_account_id` | The management account's 12-digit ID | Policy conditions and cross-account trust that single out the management account |

The remaining outputs -- `arn`, `management_account_arn`, `management_account_email`, and `resource_policy_id` (empty when no `resourcePolicy` arm is set) -- are record values for audit and import, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**All-features foundation** -- an ALL-features organization with trusted access for the governance services (CloudTrail, Config, Account Management) and the two most common policy types (SCP, tag policy) enabled on the root. The first deploy of any multi-account estate; OUs, member accounts, and policies all hang off it. Start from the **All-Features Foundation** preset.

**Consolidated billing only** -- a `CONSOLIDATED_BILLING` organization: shared invoices and volume discounts, deliberately no central governance. Right for estates migrating in from separate accounts before the governance decision is made; the upgrade to ALL is in-place when that day comes, but the reverse rebuilds the org. Start from the **Consolidated Billing Only** preset.

**Delegated security administration** -- once member accounts exist, add `delegatedAdministrators` entries handing GuardDuty, Config, or Security Hub administration to a dedicated security account. Daily governance work moves out of the management account, which stays reserved for organization management itself.

## Works With

- [**AWS Organizational Unit**](/infra-catalog/aws-organizational-unit) -- the OU tree hangs off this organization's `root_id`
- [**AWS Organization Account**](/infra-catalog/aws-organization-account) -- member accounts created into the organization's OUs
- [**AWS Organization Policy**](/infra-catalog/aws-organization-policy) -- SCPs and their siblings, gated by this resource's `enabledPolicyTypes`
- [**AWS Config Aggregator**](/infra-catalog/aws-config-aggregator) -- organization-wide Config aggregation, enabled by `config.amazonaws.com` trusted access and typically run from a delegated administrator account
