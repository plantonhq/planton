# AwsOrganization

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsOrganizationSpec defines the desired configuration for THE AWS
Organization of the deploying account: creating it makes the caller
the organization's MANAGEMENT account. One organization exists per
account; deleting this resource deletes the entire organization
(AWS requires every member account, OU, and policy to be removed
first - the delete fails otherwise, by design).

The organization-wide levers fold in here because none of them has
a life of its own: trusted service access (the provider's own docs
warn that managing it anywhere else fights this resource with a
perpetual diff), delegated administrator registrations, the org's
single resource-based delegation policy (AWS keeps exactly one per
organization - PutResourcePolicy is an upsert), and centralized
root-access management (IAM's organization features - a
management-account act that requires iam.amazonaws.com trusted
access, wired on this very spec).

Organizations is a GLOBAL service scoped to the management account;
AWS identifies the organization as "o-..." (the import ID).

## Example

```yaml
# Canonical AwsOrganization example (hack/dev manifest and refgen
# Example source): an all-features organization with trusted access
# for CloudTrail, Account Management, and IAM, SCP + tag policies
# enabled, one delegated administrator, centralized root-access
# management, and the org's resource-based delegation policy. Literal
# values stand in for a real member account so the offline `tofu plan`
# renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganization
metadata:
  name: acme-organization
  id: acme-organization
  org: test-org
  env: dev
spec:
  region: us-west-2
  featureSet: ALL
  awsServiceAccessPrincipals:
    - cloudtrail.amazonaws.com
    - account.amazonaws.com
    - iam.amazonaws.com
  enabledPolicyTypes:
    - SERVICE_CONTROL_POLICY
    - TAG_POLICY
  delegatedAdministrators:
    - accountId: "111111111111"
      servicePrincipal: config.amazonaws.com
  rootAccessManagement:
    enabledFeatures:
      - RootCredentialsManagement
      - RootSessions
  resourcePolicy:
    Version: "2012-10-17"
    Statement:
      - Sid: DelegateDescribe
        Effect: Allow
        Principal:
          AWS: "111111111111"
        Action:
          - organizations:DescribeOrganization
          - organizations:ListAccounts
        Resource: "*"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.featureSet` | `string` |  |  |  |
| `spec.awsServiceAccessPrincipals` | `[]string` |  |  |  |
| `spec.enabledPolicyTypes` | `[]string` |  |  |  |
| `spec.delegatedAdministrators` | `[]AwsOrganizationDelegatedAdministrator` |  |  |  |
| `spec.delegatedAdministrators[].accountId` | `string` |  |  |  |
| `spec.delegatedAdministrators[].servicePrincipal` | `string` | yes |  |  |
| `spec.resourcePolicy` | `object` |  |  |  |
| `spec.rootAccessManagement` | `AwsOrganizationRootAccessManagement` |  |  |  |
| `spec.rootAccessManagement.enabledFeatures` | `[]string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the
organization. Organizations is a global service - the same
organization is visible everywhere - but every AWS API call is
still made against a regional endpoint, so a region is required.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.featureSet

`string`

The organization's feature level. Unset = ALL (the provider
default, and the level every advanced arm below requires).
CONSOLIDATED_BILLING provides shared billing only. Upgrading
CONSOLIDATED_BILLING -> ALL is an in-place update (AWS's
EnableAllFeatures); downgrading ALL -> CONSOLIDATED_BILLING
REPLACES the resource - that is delete-and-recreate of the ENTIRE
organization, which AWS only permits once every member account,
OU, and policy is gone. Treat the downgrade as an organization
rebuild, never a setting change.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ALL","CONSOLIDATED_BILLING"]}}

### spec.awsServiceAccessPrincipals

`[]string`

AWS service principals granted trusted access to the organization
(e.g. "cloudtrail.amazonaws.com", "config.amazonaws.com",
"account.amazonaws.com"). Trusted access lets the named service
create its own service-linked roles in member accounts. This
field is the ONE home for service access: AWS recommends enabling
it through the owning service where possible, and the provider
warns that managing the same principal elsewhere produces a
perpetual diff. Removing an entry disables that service's access.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^([0-9a-z-]+\\.){1,4}(amazonaws\\.com(\\.cn)?|amazon\\.com)$"}}}}

### spec.enabledPolicyTypes

`[]string`

Policy types enabled on the organization's root. A type must be
enabled here before AwsOrganizationPolicy resources of that type
can attach anywhere. Enables and disables are applied on update
(disables first); each waits for AWS to confirm the type's state.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["SERVICE_CONTROL_POLICY","RESOURCE_CONTROL_POLICY","TAG_POLICY","BACKUP_POLICY","AISERVICES_OPT_OUT_POLICY","CHATBOT_POLICY","DECLARATIVE_POLICY_EC2","SECURITYHUB_POLICY","INSPECTOR_POLICY","UPGRADE_ROLLOUT_POLICY","BEDROCK_POLICY","S3_POLICY","NETWORK_SECURITY_DIRECTOR_POLICY"]}}}}

### spec.delegatedAdministrators

`[]AwsOrganizationDelegatedAdministrator`

Delegated administrator registrations: each entry names a member
account as the administrator for one AWS service across the
organization (e.g. delegating GuardDuty administration to a
security account). Both leaves are immutable - changing either
re-registers (deregister + register). The registration imports as
"{account_id}/{service_principal}".

### spec.delegatedAdministrators[].accountId

`string`

The member account receiving delegated administration (12-digit
account ID). The account must already be a member of the
organization. Immutable.

- rule: {"string":{"pattern":"^[0-9]{12}$"}}

### spec.delegatedAdministrators[].servicePrincipal

`string` · required

The service principal being delegated (e.g.
"config.amazonaws.com", "guardduty.amazonaws.com"). Immutable.

- rule: {"string":{"minLen":"1","maxLen":"128","pattern":"^([0-9a-z-]+\\.){1,4}(amazonaws\\.com(\\.cn)?|amazon\\.com)$"}}

### spec.resourcePolicy

`object`

The organization's resource-based delegation policy as free-form
JSON (an IAM-style document delegating organization management
actions to member accounts). AWS keeps exactly ONE resource
policy per organization - this arm IS that singleton
(PutResourcePolicy upserts it; removing the arm deletes it). AWS
identifies it as "rp-..." (the import ID).

### spec.rootAccessManagement

`AwsOrganizationRootAccessManagement`

Centralized root-access management: which IAM organization
features are enabled across member accounts. Requires
"iam.amazonaws.com" in aws_service_access_principals. Destroying
this arm DISABLES every enabled feature (member-account root
credentials become locally manageable again).

### spec.rootAccessManagement.enabledFeatures

`[]string` · required

The enabled features:
  - "RootCredentialsManagement": the management account (or a
    delegated admin) centrally deletes member-account root
    credentials - the lock-down posture that removes long-lived
    root passwords/keys from member accounts.
  - "RootSessions": privileged short-lived root SESSIONS on
    member accounts for the rare tasks that genuinely need root
    (deleting a mis-owned S3 bucket policy, ...).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["RootCredentialsManagement","RootSessions"]}}}}

## Validation Rules

- `spec.service_access_requires_all_features`: aws_service_access_principals requires feature_set ALL (consolidated-billing organizations cannot enable trusted service access)
- `spec.policy_types_require_all_features`: enabled_policy_types requires feature_set ALL (consolidated-billing organizations cannot enable policy types)
- `spec.delegated_admins_require_all_features`: delegated_administrators requires feature_set ALL (consolidated-billing organizations cannot delegate administration)
- `spec.resource_policy_requires_all_features`: resource_policy requires feature_set ALL (consolidated-billing organizations cannot carry a resource policy)
- `spec.delegated_administrators_unique`: delegated_administrators entries must have unique (account_id, service_principal) pairs
- `spec.root_access_requires_iam_trusted_access`: root_access_management requires 'iam.amazonaws.com' in aws_service_access_principals - IAM needs trusted access to manage root credentials across the organization

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsOrganization, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.organization_id` | `string` | The organization's AWS-generated ID ("o-..." - also the provider's import ID). |
| `status.outputs.arn` | `string` | The organization's ARN. |
| `status.outputs.management_account_id` | `string` | The management account's 12-digit account ID (the account this resource was deployed from). |
| `status.outputs.management_account_arn` | `string` | The management account's ARN. |
| `status.outputs.management_account_email` | `string` | The management account's email address. |
| `status.outputs.root_id` | `string` | The organization root's ID ("r-..."). The root is the top of the OU tree: first-level organizational units and root-scoped policy attachments reference it. |
| `status.outputs.resource_policy_id` | `string` | The folded resource policy's AWS-generated ID ("rp-..." - its import ID), empty when the spec carries no resource_policy arm. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsOrganizationalUnit | `spec.parentId` | `status.outputs.root_id` |

## See Also

- [Overview](../README.md)
