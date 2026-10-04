# AwsCloudTrailEventDataStore

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudTrailEventDataStoreSpec defines the desired configuration
for a CloudTrail Lake event data store - a queryable, immutable
store of AWS activity events, independent of any trail.

An event data store ingests events matched by its advanced event
selectors and keeps them queryable with CloudTrail Lake SQL for the
retention period. It needs no trail and no S3 bucket - Lake owns
its own storage and its own billing (per GB ingested plus, on the
fixed pricing mode, per GB retained). The store's name is
metadata.name (AWS requires 3-128 characters).

Destroying this component soft-deletes the store: AWS holds it in
PENDING_DELETION for 7 days (the name stays reserved) before the
purge. The delete is REFUSED while termination protection is on -
set termination_protection_enabled to false and apply before
destroying.

AVAILABILITY: AWS has closed CloudTrail Lake to NEW customers -
CreateEventDataStore fails with "CloudTrail Lake is no longer
accepting new customers" on any account that has never created an
event data store (live-verified 2026-08-25). Deploy only on an
account already using Lake; there is no known exception process.

## Example

```yaml
# Canonical AwsCloudTrailEventDataStore example (hack/dev manifest and
# refgen Example source): a KMS-encrypted, fixed-pricing CloudTrail
# Lake store scoped to management events with a 90-day queryable
# window. Literal ARNs stand in for composed references so the offline
# `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudTrailEventDataStore
metadata:
  name: security-lake-store
  id: security-lake-store
  org: test-org
  env: dev
spec:
  region: us-west-2
  billingMode: FIXED_RETENTION_PRICING
  kmsKeyId:
    value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
  retentionPeriodDays: 90
  advancedEventSelectors:
    - name: Management events
      fieldSelectors:
        - field: eventCategory
          equals: ["Management"]
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.billingMode` | `string` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.multiRegionEnabled` | `bool` |  |  |  |
| `spec.organizationEnabled` | `bool` |  |  |  |
| `spec.retentionPeriodDays` | `int32` |  |  |  |
| `spec.terminationProtectionEnabled` | `bool` |  |  |  |
| `spec.suspend` | `bool` |  |  |  |
| `spec.advancedEventSelectors` | `[]AwsCloudTrailEventDataStoreAdvancedEventSelector` |  |  |  |
| `spec.advancedEventSelectors[].name` | `string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors` | `[]AwsCloudTrailEventDataStoreFieldSelector` | yes |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].field` | `string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].equals` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].notEquals` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].startsWith` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].notStartsWith` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].endsWith` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].notEndsWith` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the event data store lives in. A multi-region
store still has exactly one home region - manage it from there.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.billingMode

`string`

The pricing mode. EXTENDABLE_RETENTION_PRICING (the AWS default
when unset) bills ingestion once and includes up to 10 years of
retention; FIXED_RETENTION_PRICING bills less on ingest but
charges for retention and caps it at 7 years (2555 days). Pick at
creation - AWS allows changing it later only from fixed to
extendable.

- rule: {"string":{"in":["","EXTENDABLE_RETENTION_PRICING","FIXED_RETENTION_PRICING"]}}

### spec.kmsKeyId

`string | valueFrom`

KMS key that encrypts the stored events (SSE-KMS instead of the
default AWS-owned key). The key policy must allow CloudTrail to
use it, and losing the key makes the store unreadable - AWS
recommends multi-region keys for multi-region stores. Changing
this after creation forces a replacement. Unset = AWS-owned key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.multiRegionEnabled

`bool` · optional (explicit presence)

Ingest events from ALL regions, not just the home region. Unset =
enabled (the AWS default) - a single-region store misses activity
everywhere else. Set false only for deliberately region-scoped
stores.

### spec.organizationEnabled

`bool`

Ingest events from EVERY account in the AWS Organization.
Requires running in the organization's management account (or its
delegated CloudTrail administrator) with all-features
organizations enabled.

### spec.retentionPeriodDays

`int32`

How many days events stay queryable, 7-2555. Unset = 2555 (the
AWS default, 7 years). On EXTENDABLE_RETENTION_PRICING the
effective ceiling AWS accepts is 3653 days via the console, but
the Terraform provider caps at 2555 - stay within it.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":2555,"gte":7}}

### spec.terminationProtectionEnabled

`bool` · optional (explicit presence)

Refuse deletion while enabled. Unset = enabled (the AWS default).
A destroy of this component FAILS until this is set to false and
applied - the two-step teardown is deliberate AWS behavior, not a
module defect.

### spec.suspend

`bool` · optional (explicit presence)

Pause ingestion without deleting the store (already-ingested
events stay queryable; retention keeps counting). Unset = active.
AWS never reports this back, so it is asserted on every apply and
invisible to imports.

### spec.advancedEventSelectors

`[]AwsCloudTrailEventDataStoreAdvancedEventSelector`

Which events to ingest: fine-grained field matching (eventName,
resources.ARN prefixes, eventCategory, ...) over management,
data, and network-activity events. Unset = AWS materializes a
default selector ingesting ALL management events - scope
deliberately, ingestion is what Lake bills. AWS requires every
selector to carry an eventCategory condition (a server-side rule
the provider does not pre-check).

### spec.advancedEventSelectors[].name

`string`

Display name for the selector (shown in the CloudTrail console).

- rule: {"string":{"maxLen":"1000"}}

### spec.advancedEventSelectors[].fieldSelectors

`[]AwsCloudTrailEventDataStoreFieldSelector` · required

The AND-set of field conditions. AWS requires at least a field
selector on "eventCategory".

- rule: {"repeated":{"minItems":"1"}}
- rule: set at least one of equals, not_equals, starts_with, not_starts_with, ends_with, not_ends_with

### spec.advancedEventSelectors[].fieldSelectors[].field

`string`

The event field to match.

- rule: {"string":{"in":["errorCode","eventCategory","eventName","eventSource","eventType","readOnly","resources.ARN","resources.type","sessionCredentialFromConsole","userIdentity.arn","vpcEndpointId"]}}

### spec.advancedEventSelectors[].fieldSelectors[].equals

`[]string`

Exact-match values (OR within the list).

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"2048"}}}}

### spec.advancedEventSelectors[].fieldSelectors[].notEquals

`[]string`

Exact-mismatch values.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"2048"}}}}

### spec.advancedEventSelectors[].fieldSelectors[].startsWith

`[]string`

Prefix-match values.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"2048"}}}}

### spec.advancedEventSelectors[].fieldSelectors[].notStartsWith

`[]string`

Prefix-mismatch values.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"2048"}}}}

### spec.advancedEventSelectors[].fieldSelectors[].endsWith

`[]string`

Suffix-match values.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"2048"}}}}

### spec.advancedEventSelectors[].fieldSelectors[].notEndsWith

`[]string`

Suffix-mismatch values.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"2048"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudTrailEventDataStore, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.event_data_store_arn` | `string` | The event data store's ARN (also the provider's import ID). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
