# AwsSsmPatchBaseline

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSsmPatchBaselineSpec defines the desired configuration for one
AWS Systems Manager patch baseline: which patches get auto-approved
for one operating system, the patch groups the baseline governs, and
optionally the account/region default-baseline designation for that
OS.

The baseline's name is metadata.name (3-128 characters of letters,
digits, underscores, hyphens, and periods). AWS identifies the
baseline as "pb-..." (the import ID). Patch groups fold in as
registrations keyed by the group name (each imports as
"{patch_group},{baseline_id}"), and the default designation folds in
as set_as_default_baseline - both hang off the baseline's ID and
have no life of their own.

## Example

```yaml
# Canonical AwsSsmPatchBaseline example (hack/dev manifest and refgen
# Example source): an Amazon Linux 2023 security baseline with both
# approval-arm styles across two rules, a global filter, explicit
# patch lists, an alternative repo source, governed patch groups, and
# the default designation - so the offline `tofu plan` renders every
# arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmPatchBaseline
metadata:
  name: al2023-security-baseline
  id: al2023-security-baseline
  org: test-org
  env: dev
spec:
  region: us-west-2
  operatingSystem: AMAZON_LINUX_2023
  description: AL2023 security baseline with a 7-day soak
  approvalRules:
    - patchFilters:
        - key: CLASSIFICATION
          values:
            - Security
        - key: SEVERITY
          values:
            - Critical
            - Important
      approveAfterDays: 7
      complianceLevel: CRITICAL
    - patchFilters:
        - key: CLASSIFICATION
          values:
            - Bugfix
      approveUntilDate: "2026-12-31"
      complianceLevel: LOW
      enableNonSecurity: true
  globalFilters:
    - key: PRODUCT
      values:
        - AmazonLinux2023
  approvedPatches:
    - kernel-6.1.134
  approvedPatchesComplianceLevel: HIGH
  rejectedPatches:
    - nginx-1.25.0
  rejectedPatchesAction: BLOCK
  sources:
    - name: internal-repo
      configuration: |
        [internal]
        name=Internal AL2023 mirror
        baseurl=https://repo.example.com/al2023
        enabled=1
      products:
        - AmazonLinux2023.3
  patchGroups:
    - prod-servers
    - staging-servers
  setAsDefaultBaseline: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.operatingSystem` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.approvalRules` | `[]AwsSsmPatchBaselineApprovalRule` |  |  |  |
| `spec.approvalRules[].patchFilters` | `[]AwsSsmPatchBaselinePatchFilter` | yes |  |  |
| `spec.approvalRules[].patchFilters[].key` | `string` |  |  |  |
| `spec.approvalRules[].patchFilters[].values` | `[]string` | yes |  |  |
| `spec.approvalRules[].approveAfterDays` | `int32` |  |  |  |
| `spec.approvalRules[].approveUntilDate` | `string` |  |  |  |
| `spec.approvalRules[].complianceLevel` | `string` |  |  |  |
| `spec.approvalRules[].enableNonSecurity` | `bool` |  |  |  |
| `spec.globalFilters` | `[]AwsSsmPatchBaselinePatchFilter` |  |  |  |
| `spec.globalFilters[].key` | `string` |  |  |  |
| `spec.globalFilters[].values` | `[]string` | yes |  |  |
| `spec.approvedPatches` | `[]string` |  |  |  |
| `spec.approvedPatchesComplianceLevel` | `string` |  |  |  |
| `spec.approvedPatchesEnableNonSecurity` | `bool` |  |  |  |
| `spec.rejectedPatches` | `[]string` |  |  |  |
| `spec.rejectedPatchesAction` | `string` |  |  |  |
| `spec.availableSecurityUpdatesComplianceStatus` | `string` |  |  |  |
| `spec.sources` | `[]AwsSsmPatchBaselineSource` |  |  |  |
| `spec.sources[].name` | `string` | yes |  |  |
| `spec.sources[].configuration` | `string` | yes |  |  |
| `spec.sources[].products` | `[]string` | yes |  |  |
| `spec.patchGroups` | `[]string` |  |  |  |
| `spec.setAsDefaultBaseline` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the baseline lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.operatingSystem

`string`

The operating system the baseline approves patches for. Unset =
WINDOWS (the provider default). Changing it forces replacement.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["WINDOWS","AMAZON_LINUX","AMAZON_LINUX_2","AMAZON_LINUX_2022","AMAZON_LINUX_2023","UBUNTU","REDHAT_ENTERPRISE_LINUX","SUSE","CENTOS","ORACLE_LINUX","DEBIAN","MACOS","RASPBIAN","ROCKY_LINUX","ALMA_LINUX"]}}

### spec.description

`string`

Human description of the baseline, up to 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.approvalRules

`[]AwsSsmPatchBaselineApprovalRule`

Auto-approval rules: each rule's filters select patches and its
approval arm (days-after-release or a fixed cutoff date) schedules
approval. On Debian and Ubuntu, approve_after_days is not
supported by AWS - rules there typically carry neither arm and
approve by filter alone.

- rule: approve_after_days and approve_until_date are mutually exclusive

### spec.approvalRules[].patchFilters

`[]AwsSsmPatchBaselinePatchFilter` · required

Filters selecting which patches this rule approves (1-10; keys
valid per the baseline's operating system).

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.approvalRules[].patchFilters[].key

`string`

The patch property to filter on. Which keys are valid depends on
the baseline's operating system (PRODUCT/CLASSIFICATION/
MSRC_SEVERITY for Windows; PRODUCT/CLASSIFICATION/SEVERITY for
most Linux; SECTION/PRIORITY for Debian/Ubuntu; ...).

- rule: {"string":{"in":["ARCH","ADVISORY_ID","BUGZILLA_ID","PATCH_SET","PRODUCT","PRODUCT_FAMILY","CLASSIFICATION","CVE_ID","EPOCH","MSRC_SEVERITY","NAME","PATCH_ID","SECTION","PRIORITY","REPOSITORY","RELEASE","SEVERITY","SECURITY","VERSION"]}}

### spec.approvalRules[].patchFilters[].values

`[]string` · required

The values to match ("*" matches everything the key can hold; up
to 20 values).

- rule: {"repeated":{"minItems":"1","maxItems":"20","items":{"string":{"minLen":"1","maxLen":"64"}}}}

### spec.approvalRules[].approveAfterDays

`int32` · optional (explicit presence)

Days after a patch's release before this rule approves it (0-360;
0 = approve immediately on release). Mutually exclusive with
approve_until_date. Not supported by AWS on Debian or Ubuntu.

- rule: {"int32":{"lte":360,"gte":0}}

### spec.approvalRules[].approveUntilDate

`string`

Approve every selected patch released on or before this date
(YYYY-MM-DD). Mutually exclusive with approve_after_days.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[12][0-9]{3}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$"}}

### spec.approvalRules[].complianceLevel

`string`

Compliance level reported when a node misses a patch this rule
approved. Unset = UNSPECIFIED.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","HIGH","MEDIUM","LOW","INFORMATIONAL","UNSPECIFIED"]}}

### spec.approvalRules[].enableNonSecurity

`bool`

Include non-security updates in what this rule approves (Linux
only).

### spec.globalFilters

`[]AwsSsmPatchBaselinePatchFilter`

Filters narrowing which patches the whole baseline considers,
applied on top of every approval rule (up to 4).

- rule: {"repeated":{"maxItems":"4"}}

### spec.globalFilters[].key

`string`

The patch property to filter on. Which keys are valid depends on
the baseline's operating system (PRODUCT/CLASSIFICATION/
MSRC_SEVERITY for Windows; PRODUCT/CLASSIFICATION/SEVERITY for
most Linux; SECTION/PRIORITY for Debian/Ubuntu; ...).

- rule: {"string":{"in":["ARCH","ADVISORY_ID","BUGZILLA_ID","PATCH_SET","PRODUCT","PRODUCT_FAMILY","CLASSIFICATION","CVE_ID","EPOCH","MSRC_SEVERITY","NAME","PATCH_ID","SECTION","PRIORITY","REPOSITORY","RELEASE","SEVERITY","SECURITY","VERSION"]}}

### spec.globalFilters[].values

`[]string` · required

The values to match ("*" matches everything the key can hold; up
to 20 values).

- rule: {"repeated":{"minItems":"1","maxItems":"20","items":{"string":{"minLen":"1","maxLen":"64"}}}}

### spec.approvedPatches

`[]string`

Explicitly approved patches (IDs, KBs, or package names - up to
50), independent of the approval rules.

- rule: {"repeated":{"maxItems":"50","items":{"string":{"minLen":"1","maxLen":"100"}}}}

### spec.approvedPatchesComplianceLevel

`string`

Compliance level reported for approved-patches entries when a
managed node misses one. Unset = UNSPECIFIED.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","HIGH","MEDIUM","LOW","INFORMATIONAL","UNSPECIFIED"]}}

### spec.approvedPatchesEnableNonSecurity

`bool`

Treat the approved_patches list as including non-security updates
(Linux only - Windows patches are always security-classified).

### spec.rejectedPatches

`[]string`

Explicitly rejected patches (up to 50).

- rule: {"repeated":{"maxItems":"50","items":{"string":{"minLen":"1","maxLen":"100"}}}}

### spec.rejectedPatchesAction

`string`

What happens when a rejected patch is a dependency of an approved
one: install it anyway (ALLOW_AS_DEPENDENCY) or block it and
report noncompliance (BLOCK). Unset = ALLOW_AS_DEPENDENCY (AWS's
default, echoed back on read).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ALLOW_AS_DEPENDENCY","BLOCK"]}}

### spec.availableSecurityUpdatesComplianceStatus

`string`

Compliance status reported for AVAILABLE security updates that are
not yet approved (Windows Server managed nodes only, per AWS).
Unset = AWS's default posture.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["COMPLIANT","NON_COMPLIANT"]}}

### spec.sources

`[]AwsSsmPatchBaselineSource`

Alternative patch source repositories for Linux instances (up to
20), each a yum/apt repo definition patching pulls from instead of
the OS defaults.

- rule: {"repeated":{"maxItems":"20"}}

### spec.sources[].name

`string` · required

The source's name (3-50 characters of letters, digits, hyphens,
underscores, and periods).

- rule: {"string":{"minLen":"3","maxLen":"50","pattern":"^[0-9A-Za-z_.-]{3,50}$"}}

### spec.sources[].configuration

`string` · required

The repo definition in the OS package manager's own syntax (a yum
.repo section or an apt source line).

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.sources[].products

`[]string` · required

The OS product versions the source applies to (up to 20, e.g.
"AmazonLinux2023.3" or "*").

- rule: {"repeated":{"minItems":"1","maxItems":"20","items":{"string":{"minLen":"1","maxLen":"128"}}}}

### spec.patchGroups

`[]string`

Patch groups this baseline governs: each entry registers the named
group (the "Patch Group" tag value on managed nodes) with this
baseline. A group can be registered with only ONE baseline per OS
account-wide - AWS state, not validation, is the referee.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1","maxLen":"256"}}}}

### spec.setAsDefaultBaseline

`bool`

Designate this baseline as the account/region DEFAULT for its
operating system: nodes whose patch group is registered with no
baseline patch against the default. One default exists per OS per
region - claiming it silently displaces whichever baseline held
the designation, at most one baseline per OS should set it
(documents cannot see each other, so AWS state - not validation -
is the referee), and un-setting the flag (or deleting this
resource) RESTORES AWS's own predefined default baseline for the
OS (the provider records and reverts to it).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSsmPatchBaseline, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.baseline_id` | `string` | The baseline's AWS-generated ID ("pb-..." - also the provider's import ID; each folded patch group imports as "{patch_group},{baseline_id}"). |
| `status.outputs.baseline_arn` | `string` | The baseline's ARN. |
| `status.outputs.operating_system` | `string` | The operating system the baseline governs (the provider echoes WINDOWS when the spec leaves it unset; also the default designation's import ID). |

## See Also

- [Overview](../README.md)
