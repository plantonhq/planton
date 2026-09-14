# AwsSesAccountSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsSesAccountSettingsSpec defines SES (SESv2) account-level
settings for one AWS region: the account suppression list and the
Virtual Deliverability Manager (VDM) posture.

This is a SETTINGS SINGLETON: AWS keeps exactly one SES account
object per account+region, and this component manages its
account-wide attributes. Deploy at most one instance per region;
two instances targeting the same region fight over the same
account object. metadata.name never reaches AWS - it is
Planton-side identity only. These settings are deliberately NOT
fields on AwsSesConfigurationSet or AwsSesEmailIdentity - multiple
sets/identities would fight over the one account object.

Destroy semantics DIFFER per arm (both taught on the arms below):
suppression settings PERSIST after destroy; the VDM arm is reset to
disabled.

## Example

```yaml
# Canonical AwsSesAccountSettings example (hack/dev manifest and
# refgen Example source): the recommended reputation posture -- both
# suppression reasons on, VDM enabled with engagement metrics.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesAccountSettings
metadata:
  name: ses-account-us-west-2
  id: ses-account-us-west-2
  org: test-org
  env: dev
spec:
  region: us-west-2
  suppression:
    reasons:
      - BOUNCE
      - COMPLAINT
  vdm:
    enabled: true
    engagementMetrics: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.suppression` | `AwsSesAccountSettingsSuppression` |  |  |  |
| `spec.suppression.reasons` | `[]string` |  |  |  |
| `spec.suppression.enabled` | `bool` |  | `true` |  |
| `spec.vdm` | `AwsSesAccountSettingsVdm` |  |  |  |
| `spec.vdm.enabled` | `bool` | yes |  |  |
| `spec.vdm.engagementMetrics` | `bool` |  |  |  |
| `spec.vdm.optimizedSharedDelivery` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region whose SES account settings this instance manages.
The region IS the resource identity - one instance per region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.suppression

`AwsSesAccountSettingsSuppression`

The account-level suppression list configuration. Omit the arm to
leave the account's suppression settings untouched; set it with
`enabled: false` to explicitly turn account-level auto-suppression
OFF.

- rule: suppression is on but names no events - list BOUNCE and/or COMPLAINT in reasons, or set enabled: false to turn account-level auto-suppression off
- rule: suppression is off but still names events - drop reasons, or set enabled: true to suppress on them

### spec.suppression.reasons

`[]string`

Which events auto-suppress a recipient:
  - "BOUNCE": hard bounces add the address to the list.
  - "COMPLAINT": spam complaints add the address to the list.
Both together is the recommended reputation posture.

Applying this arm overwrites whatever was previously set - and
the setting PERSISTS after this component is destroyed (SES has
no delete for it; the last-applied reasons stay in effect). To
stop suppressing, apply `enabled: false` before destroying.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["BOUNCE","COMPLAINT"]}}}}

### spec.suppression.enabled

`bool` · optional (explicit presence)

Whether account-level auto-suppression is on. Unset means on:
declaring the arm with events has always meant suppressing on them,
and `enabled: false` is the explicit OFF (SES then holds an empty
reason list).

- default: `true`

### spec.vdm

`AwsSesAccountSettingsVdm`

The Virtual Deliverability Manager (VDM) posture - SES's
deliverability analytics suite. Omit the arm to leave the
account's VDM state untouched.

### spec.vdm.enabled

`bool` · required · optional (explicit presence)

Master switch for VDM on the account. Required inside vdm: declaring
the arm takes VDM under management, and the switch says which way.
Destroying this component resets VDM to disabled (unlike the
suppression arm, this one IS reverted on destroy). VDM carries its own
AWS pricing - enabling it is a billing decision, not just a feature
flag.

- rule: {"required":true}

### spec.vdm.engagementMetrics

`bool` · optional (explicit presence)

Track open/click engagement metrics in the VDM dashboard. Unset
leaves AWS's default for the account; only meaningful while
enabled is true.

### spec.vdm.optimizedSharedDelivery

`bool` · optional (explicit presence)

Let Guardian optimize shared-delivery behavior (SES adjusts
sending patterns to protect deliverability). Unset leaves AWS's
default for the account; only meaningful while enabled is true.

## Validation Rules

- `spec.at_least_one_arm`: configure at least one of suppression / vdm - an instance managing neither is dead configuration

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSesAccountSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.account_id` | `string` | The 12-digit AWS account ID the settings belong to (also the provider's import ID for the suppression singleton). |

## See Also

- [Overview](../README.md)
