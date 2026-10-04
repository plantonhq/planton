# AwsGuardDuty

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsGuardDutySpec defines this account+region's GuardDuty threat-
detection posture: the detector, its protection-plan features, the
finding filters and trusted/threat IP lists, findings export, and -
for organization administrators - member-account management.

This is a REGION SINGLETON: AWS allows exactly ONE detector per
account per region, and the detector has no name - its AWS-assigned
ID is the identity, so metadata.name never reaches AWS. Deploy at
most one instance per region; a second instance (or a detector
enabled by hand or by AWS Organizations auto-enable) makes creation
fail with "detector already exists".

The component serves one of two postures per account:
  - ADMIN side: set `organization` (and optionally `members`) from
    the organization's delegated GuardDuty administrator account.
  - MEMBER side: set `accept_invitation_from_account_id` to join an
    administrator by invitation (the legacy non-Organizations flow).

Destroying this component deletes the detector - a REAL delete that
removes all findings and disassociates members. Feature and
organization-configuration arms are patches onto the detector (AWS
has no delete for them); removing an arm from the spec reverts
nothing on its own, which both engines' modules document and the
GUIDE teaches.

## Example

```yaml
# Canonical AwsGuardDuty example (hack/dev manifest and refgen Example
# source): the region's threat-detection posture -- detector, protection
# plans, a noise filter, trusted/threat lists, and findings export.
# Literal values stand in for composed references so the offline `tofu
# plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsGuardDuty
metadata:
  name: guardduty-us-west-2
  id: guardduty-us-west-2
  org: test-org
  env: dev
spec:
  region: us-west-2
  findingPublishingFrequency: FIFTEEN_MINUTES
  features:
    - name: S3_DATA_EVENTS
    - name: RUNTIME_MONITORING
      additionalConfiguration:
        - name: EC2_AGENT_MANAGEMENT
          enabled: false
  filters:
    - name: archive-sandbox-low-severity
      description: Archive low-severity findings from the sandbox account
      action: ARCHIVE
      rank: 1
      criteria:
        - field: severity
          lessThan: "4"
        - field: accountId
          equals: ["123456789012"]
  ipSets:
    - name: trusted-office
      format: TXT
      location: https://s3.amazonaws.com/my-security-lists/trusted.txt
      activate: true
  threatIntelSets:
    - name: known-bad
      format: TXT
      location: https://s3.amazonaws.com/my-security-lists/bad.txt
      activate: true
  publishingDestination:
    bucketArn:
      value: arn:aws:s3:::my-findings-archive
    kmsKeyArn:
      value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.enable` | `bool` |  |  |  |
| `spec.findingPublishingFrequency` | `string` |  |  |  |
| `spec.features` | `[]AwsGuardDutyFeature` |  |  |  |
| `spec.features[].name` | `string` |  |  |  |
| `spec.features[].enabled` | `bool` |  |  |  |
| `spec.features[].additionalConfiguration` | `[]AwsGuardDutyFeatureAdditionalConfiguration` |  |  |  |
| `spec.features[].additionalConfiguration[].name` | `string` |  |  |  |
| `spec.features[].additionalConfiguration[].enabled` | `bool` |  |  |  |
| `spec.filters` | `[]AwsGuardDutyFilter` |  |  |  |
| `spec.filters[].name` | `string` | yes |  |  |
| `spec.filters[].description` | `string` |  |  |  |
| `spec.filters[].action` | `string` |  |  |  |
| `spec.filters[].rank` | `int32` |  |  |  |
| `spec.filters[].criteria` | `[]AwsGuardDutyFilterCriterion` | yes |  |  |
| `spec.filters[].criteria[].field` | `string` | yes |  |  |
| `spec.filters[].criteria[].equals` | `[]string` |  |  |  |
| `spec.filters[].criteria[].notEquals` | `[]string` |  |  |  |
| `spec.filters[].criteria[].matches` | `[]string` |  |  |  |
| `spec.filters[].criteria[].notMatches` | `[]string` |  |  |  |
| `spec.filters[].criteria[].greaterThan` | `string` |  |  |  |
| `spec.filters[].criteria[].greaterThanOrEqual` | `string` |  |  |  |
| `spec.filters[].criteria[].lessThan` | `string` |  |  |  |
| `spec.filters[].criteria[].lessThanOrEqual` | `string` |  |  |  |
| `spec.ipSets` | `[]AwsGuardDutyIpSet` |  |  |  |
| `spec.ipSets[].name` | `string` | yes |  |  |
| `spec.ipSets[].format` | `string` |  |  |  |
| `spec.ipSets[].location` | `string` | yes |  |  |
| `spec.ipSets[].activate` | `bool` |  |  |  |
| `spec.threatIntelSets` | `[]AwsGuardDutyThreatIntelSet` |  |  |  |
| `spec.threatIntelSets[].name` | `string` | yes |  |  |
| `spec.threatIntelSets[].format` | `string` |  |  |  |
| `spec.threatIntelSets[].location` | `string` | yes |  |  |
| `spec.threatIntelSets[].activate` | `bool` |  |  |  |
| `spec.publishingDestination` | `AwsGuardDutyPublishingDestination` |  |  |  |
| `spec.publishingDestination.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.publishingDestination.kmsKeyArn` | `string \| valueFrom` | yes |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.organization` | `AwsGuardDutyOrganization` |  |  |  |
| `spec.organization.adminAccountId` | `string` |  |  |  |
| `spec.organization.autoEnableOrganizationMembers` | `string` |  |  |  |
| `spec.organization.features` | `[]AwsGuardDutyOrganizationFeature` |  |  |  |
| `spec.organization.features[].name` | `string` |  |  |  |
| `spec.organization.features[].autoEnable` | `string` |  |  |  |
| `spec.organization.features[].additionalConfiguration` | `[]AwsGuardDutyOrganizationFeatureAdditionalConfiguration` |  |  |  |
| `spec.organization.features[].additionalConfiguration[].name` | `string` |  |  |  |
| `spec.organization.features[].additionalConfiguration[].autoEnable` | `string` |  |  |  |
| `spec.members` | `[]AwsGuardDutyMember` |  |  |  |
| `spec.members[].accountId` | `string` |  |  |  |
| `spec.members[].email` | `string` | yes |  |  |
| `spec.members[].invite` | `bool` |  |  |  |
| `spec.members[].invitationMessage` | `string` |  |  |  |
| `spec.members[].disableEmailNotification` | `bool` |  |  |  |
| `spec.members[].features` | `[]AwsGuardDutyMemberFeature` |  |  |  |
| `spec.members[].features[].name` | `string` |  |  |  |
| `spec.members[].features[].enabled` | `bool` |  |  |  |
| `spec.members[].features[].additionalConfiguration` | `[]AwsGuardDutyFeatureAdditionalConfiguration` |  |  |  |
| `spec.members[].features[].additionalConfiguration[].name` | `string` |  |  |  |
| `spec.members[].features[].additionalConfiguration[].enabled` | `bool` |  |  |  |
| `spec.acceptInvitationFromAccountId` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region whose GuardDuty posture this instance manages. The
region IS the resource identity - one instance per region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.enable

`bool` · optional (explicit presence)

Whether the detector is actively monitoring. Unset = enabled.
Set false to suspend monitoring (and billing) without losing the
detector, its findings, or its configuration.

### spec.findingPublishingFrequency

`string`

How often UPDATED findings re-export to EventBridge/S3 (new
findings always publish within minutes). Unset = SIX_HOURS (the
AWS default). Members inherit the administrator's value - AWS
overwrites a member-side setting on the next org sync.

- rule: {"string":{"in":["","FIFTEEN_MINUTES","ONE_HOUR","SIX_HOURS"]}}

### spec.features

`[]AwsGuardDutyFeature`

Protection plans on this detector (S3 data events, EKS audit
logs, runtime monitoring, ...). Features NOT listed here are left
exactly as AWS has them (the foundational data sources are always
on) - AWS has no "delete" for a feature, so both engines patch
only the listed ones.

- rule: additional_configuration entries must have unique names

### spec.features[].name

`string`

The protection plan.

- rule: {"string":{"in":["S3_DATA_EVENTS","EKS_AUDIT_LOGS","EBS_MALWARE_PROTECTION","RDS_LOGIN_EVENTS","LAMBDA_NETWORK_LOGS","EKS_RUNTIME_MONITORING","RUNTIME_MONITORING","AI_PROTECTION","AI_ANALYST"]}}

### spec.features[].enabled

`bool` · optional (explicit presence)

Whether the plan is on. Unset = enabled (listing a feature means
you want it); set false to explicitly turn a plan off.

### spec.features[].additionalConfiguration

`[]AwsGuardDutyFeatureAdditionalConfiguration`

Sub-toggles within the plan (agent management for RUNTIME_
MONITORING and friends). Declare only the sub-toggles you enable:
AWS materializes a feature's FULL sub-toggle family server-side
(undeclared members read back DISABLED), and the modules mirror
that contract by always sending the complete family - a declared
member carries your value, an undeclared member an explicit
DISABLED (live-verified 2026-08-25).

### spec.features[].additionalConfiguration[].name

`string`

The sub-toggle.

- rule: {"string":{"in":["EKS_ADDON_MANAGEMENT","ECS_FARGATE_AGENT_MANAGEMENT","EC2_AGENT_MANAGEMENT"]}}

### spec.features[].additionalConfiguration[].enabled

`bool` · optional (explicit presence)

Whether the sub-toggle is on. Unset = enabled.

### spec.filters

`[]AwsGuardDutyFilter`

Finding filters: match criteria that auto-archive (or simply
organize) findings - the noise-control surface.

- rule: criteria entries must have unique field values

### spec.filters[].name

`string` · required

Filter name (3-64 characters: letters, digits, '.', '_', '-').
The for_each key on both engines.

- rule: {"string":{"minLen":"3","maxLen":"64","pattern":"^[a-zA-Z0-9_.-]+$"}}

### spec.filters[].description

`string`

What this filter is for (shown in the console).

- rule: {"string":{"maxLen":"512"}}

### spec.filters[].action

`string`

NOOP keeps matching findings visible (organizational filter);
ARCHIVE suppresses them - the noise-control action.

- rule: {"string":{"in":["NOOP","ARCHIVE"]}}

### spec.filters[].rank

`int32`

Evaluation order across filters (lower = earlier). AWS accepts
1-100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.filters[].criteria

`[]AwsGuardDutyFilterCriterion` · required

The AND-set of match conditions.

- rule: {"repeated":{"minItems":"1"}}
- rule: set at least one condition (equals, not_equals, matches, not_matches, or a bound)

### spec.filters[].criteria[].field

`string` · required

The finding attribute to match (the GuardDuty filter-attribute
vocabulary, e.g. "severity", "type", "accountId",
"resource.s3BucketDetails.name").

- rule: {"string":{"minLen":"1"}}

### spec.filters[].criteria[].equals

`[]string`

Exact-match values (OR within the list).

### spec.filters[].criteria[].notEquals

`[]string`

Exact-mismatch values.

### spec.filters[].criteria[].matches

`[]string`

Pattern-match values (at most 5).

- rule: {"repeated":{"maxItems":"5"}}

### spec.filters[].criteria[].notMatches

`[]string`

Pattern-mismatch values (at most 5).

- rule: {"repeated":{"maxItems":"5"}}

### spec.filters[].criteria[].greaterThan

`string`

Numeric/date lower bound (exclusive). A number ("7") or an ISO
timestamp.

### spec.filters[].criteria[].greaterThanOrEqual

`string`

Numeric/date lower bound (inclusive).

### spec.filters[].criteria[].lessThan

`string`

Numeric/date upper bound (exclusive).

### spec.filters[].criteria[].lessThanOrEqual

`string`

Numeric/date upper bound (inclusive).

### spec.ipSets

`[]AwsGuardDutyIpSet`

Trusted IP lists: traffic from these CIDRs/IPs never generates
findings. AWS allows ONE ACTIVE trusted list per detector.

### spec.ipSets[].name

`string` · required

List name (the for_each key on both engines).

- rule: {"string":{"minLen":"1","maxLen":"300"}}

### spec.ipSets[].format

`string`

The list file's format.

- rule: {"string":{"in":["TXT","STIX","OTX_CSV","ALIEN_VAULT","PROOF_POINT","FIRE_EYE"]}}

### spec.ipSets[].location

`string` · required

S3 URI of the list file (e.g.
"https://s3.amazonaws.com/my-bucket/trusted.txt" or
"s3://my-bucket/trusted.txt"). GuardDuty must be able to read it.

- rule: {"string":{"minLen":"1"}}

### spec.ipSets[].activate

`bool`

Whether GuardDuty uses the list. Only one trusted list can be
active per detector - keep inactive spares false.

### spec.threatIntelSets

`[]AwsGuardDutyThreatIntelSet`

Threat intel lists: known-malicious IPs that always generate
findings when seen.

### spec.threatIntelSets[].name

`string` · required

List name (the for_each key on both engines).

- rule: {"string":{"minLen":"1","maxLen":"300"}}

### spec.threatIntelSets[].format

`string`

The list file's format.

- rule: {"string":{"in":["TXT","STIX","OTX_CSV","ALIEN_VAULT","PROOF_POINT","FIRE_EYE"]}}

### spec.threatIntelSets[].location

`string` · required

S3 URI of the list file. GuardDuty must be able to read it.

- rule: {"string":{"minLen":"1"}}

### spec.threatIntelSets[].activate

`bool`

Whether GuardDuty uses the list.

### spec.publishingDestination

`AwsGuardDutyPublishingDestination`

Export findings to an S3 bucket (long-term retention beyond
GuardDuty's 90 days). AWS allows one publishing destination per
detector.

### spec.publishingDestination.bucketArn

`string | valueFrom` · required

The destination bucket's ARN. GuardDuty validates the export
write at EXACTLY the ARN it is given: a literal value may append
a key prefix (arn:aws:s3:::bucket/prefix) and scope the bucket
policy's s3:PutObject grant to that prefix, but the reference
path composes the BARE bucket ARN, so the grant must then cover
the whole bucket - a narrower prefix grant fails
CreatePublishingDestination at create time (live-verified).

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.publishingDestination.kmsKeyArn

`string | valueFrom` · required

The KMS key GuardDuty encrypts exported findings with. REQUIRED
by AWS for findings export.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.organization

`AwsGuardDutyOrganization`

ADMIN side: organization-wide GuardDuty administration - run from
the delegated administrator account (or delegate it here first
via admin_account_id from the management account).

- rule: features entries must have unique names

### spec.organization.adminAccountId

`string`

Delegate this account as the organization's GuardDuty
administrator - an ACCOUNT-GLOBAL act performed from the
MANAGEMENT account (one delegation per organization). Unset when
the delegation already exists (the common case: it is done once,
centrally). Deregistered on destroy.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

### spec.organization.autoEnableOrganizationMembers

`string`

Whether NEW organization accounts (or ALL accounts, or NONE) get
GuardDuty automatically.

- rule: {"string":{"in":["NEW","ALL","NONE"]}}

### spec.organization.features

`[]AwsGuardDutyOrganizationFeature`

Organization-wide feature auto-enablement, patch-keyed by name
like detector features.

- rule: additional_configuration entries must have unique names

### spec.organization.features[].name

`string`

The protection plan (the organization vocabulary has no
AI_ANALYST).

- rule: {"string":{"in":["S3_DATA_EVENTS","EKS_AUDIT_LOGS","EBS_MALWARE_PROTECTION","RDS_LOGIN_EVENTS","LAMBDA_NETWORK_LOGS","EKS_RUNTIME_MONITORING","RUNTIME_MONITORING","AI_PROTECTION"]}}

### spec.organization.features[].autoEnable

`string`

Which member accounts get the plan: NEW members only, ALL
members, or NONE.

- rule: {"string":{"in":["NEW","ALL","NONE"]}}

### spec.organization.features[].additionalConfiguration

`[]AwsGuardDutyOrganizationFeatureAdditionalConfiguration`

Sub-toggles, same NEW/ALL/NONE vocabulary.

### spec.organization.features[].additionalConfiguration[].name

`string`

The sub-toggle.

- rule: {"string":{"in":["EKS_ADDON_MANAGEMENT","ECS_FARGATE_AGENT_MANAGEMENT","EC2_AGENT_MANAGEMENT"]}}

### spec.organization.features[].additionalConfiguration[].autoEnable

`string`

Which member accounts get the sub-toggle.

- rule: {"string":{"in":["NEW","ALL","NONE"]}}

### spec.members

`[]AwsGuardDutyMember`

ADMIN side: member accounts monitored by this detector. With
organization.auto_enable_organization_members, new org accounts
join automatically and this list is only for exceptions or
non-Organizations members.

- rule: features entries must have unique names

### spec.members[].accountId

`string`

The member's AWS account ID.

- rule: {"string":{"pattern":"^[0-9]{12}$"}}

### spec.members[].email

`string` · required

The member account's root email address (AWS requires it to
create the member record).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.members[].invite

`bool` · optional (explicit presence)

Send an invitation (the non-Organizations flow; Organizations
members associate without one). The only member field AWS can
update in place.

### spec.members[].invitationMessage

`string`

Text of the invitation email.

### spec.members[].disableEmailNotification

`bool`

Skip the invitation email (invite via the API relationship only).

### spec.members[].features

`[]AwsGuardDutyMemberFeature`

Per-member protection-plan overrides (the org feature vocabulary,
binary per member).

- rule: additional_configuration entries must have unique names

### spec.members[].features[].name

`string`

The protection plan (the organization vocabulary).

- rule: {"string":{"in":["S3_DATA_EVENTS","EKS_AUDIT_LOGS","EBS_MALWARE_PROTECTION","RDS_LOGIN_EVENTS","LAMBDA_NETWORK_LOGS","EKS_RUNTIME_MONITORING","RUNTIME_MONITORING","AI_PROTECTION"]}}

### spec.members[].features[].enabled

`bool` · optional (explicit presence)

Whether the plan is on for this member. Unset = enabled.

### spec.members[].features[].additionalConfiguration

`[]AwsGuardDutyFeatureAdditionalConfiguration`

Sub-toggles for this member's plan.

### spec.members[].features[].additionalConfiguration[].name

`string`

The sub-toggle.

- rule: {"string":{"in":["EKS_ADDON_MANAGEMENT","ECS_FARGATE_AGENT_MANAGEMENT","EC2_AGENT_MANAGEMENT"]}}

### spec.members[].features[].additionalConfiguration[].enabled

`bool` · optional (explicit presence)

Whether the sub-toggle is on. Unset = enabled.

### spec.acceptInvitationFromAccountId

`string`

MEMBER side: accept a pending GuardDuty invitation from this
administrator account (the legacy invitation flow; Organizations
members never need it). Mutually exclusive with the admin-side
arms.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

## Validation Rules

- `feature_names_unique`: features entries must have unique names
- `filter_names_unique`: filters entries must have unique names
- `ip_set_names_unique`: ip_sets entries must have unique names
- `threat_intel_set_names_unique`: threat_intel_sets entries must have unique names
- `member_accounts_unique`: members entries must have unique account_id values
- `member_xor_admin_posture`: accept_invitation_from_account_id is mutually exclusive with organization and members (an account is either a member or an administrator)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsGuardDuty, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.detector_id` | `string` | The detector's AWS-assigned ID (also the provider's import ID and the key every satellite composes its import ID from). |
| `status.outputs.detector_arn` | `string` | The detector's ARN. |
| `status.outputs.account_id` | `string` | The AWS account ID the detector belongs to. |
| `status.outputs.ip_set_ids` | `map<string, string>` | Trusted IP list IDs keyed exactly like spec.ip_sets entries (by name) - the import path for folded satellites. |
| `status.outputs.threat_intel_set_ids` | `map<string, string>` | Threat intel list IDs keyed exactly like spec.threat_intel_sets entries (by name). |
| `status.outputs.publishing_destination_id` | `string` | The findings-export destination ID (set only when spec.publishing_destination is configured). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.publishingDestination.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.publishingDestination.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
