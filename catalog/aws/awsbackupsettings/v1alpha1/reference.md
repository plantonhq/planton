# AwsBackupSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBackupSettingsSpec defines AWS Backup account and region
settings as two independent arms.

This is a SETTINGS SINGLETON: the region arm's identity is the
account+region pair (AWS keeps exactly one set of Backup
preferences per region), and the global arm's identity is the
ACCOUNT (one set of global settings account-wide). Deploy at most
one instance per region, and set the global arm in exactly ONE
instance across all regions - two instances carrying the same arm
fight over the same settings object. metadata.name never reaches
AWS - it is Planton-side identity only. These settings are
deliberately NOT fields on AwsBackupVault or AwsBackupPlan -
multiple vaults/plans would fight over the one settings object.

Destroy is a NO-OP on BOTH arms (the provider registers no delete
call): whatever was last applied stays in effect indefinitely. To
revert a setting, apply the desired value - never rely on destroy.

## Example

```yaml
# Canonical AwsBackupSettings example (hack/dev manifest and refgen
# Example source): both arms -- the region's opt-in posture and the
# account's cross-account switch.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupSettings
metadata:
  name: backup-settings
  id: backup-settings
  org: test-org
  env: dev
spec:
  region: us-west-2
  global:
    settings:
      isCrossAccountBackupEnabled: "true"
  regionSettings:
    resourceTypeOptInPreference:
      EBS: true
      EC2: true
      RDS: true
      S3: false
    resourceTypeManagementPreference:
      DynamoDB: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.global` | `AwsBackupSettingsGlobal` |  |  |  |
| `spec.global.settings` | `map<string, string>` | yes |  |  |
| `spec.regionSettings` | `AwsBackupSettingsRegion` |  |  |  |
| `spec.regionSettings.resourceTypeOptInPreference` | `map<string, bool>` | yes |  |  |
| `spec.regionSettings.resourceTypeManagementPreference` | `map<string, bool>` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region this instance manages Backup settings in. For the
region arm, the region IS the resource identity - one instance
per region. The global arm is account-wide regardless of region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.global

`AwsBackupSettingsGlobal`

The account-wide global settings arm. Omit to leave the account's
global settings untouched. Set it in exactly one instance
account-wide.

### spec.global.settings

`map<string, string>` · required

The global settings map. AWS's one documented key is
"isCrossAccountBackupEnabled" ("true"/"false" as strings) -
cross-account backup copies for organizations. AWS returns EVERY
supported key on read, so list every key you intend to manage or
the apply plans show perpetual differences.

Applying overwrites what was previously set, and the values
PERSIST after this component is destroyed (destroy is a no-op) -
to turn a setting off, apply it as "false" before destroying.

- rule: {"map":{"minPairs":"1"}}

### spec.regionSettings

`AwsBackupSettingsRegion`

The per-region preferences arm. Omit to leave the region's
preferences untouched.

### spec.regionSettings.resourceTypeOptInPreference

`map<string, bool>` · required

Which resource types AWS Backup protects in this region, keyed by
AWS's type names ("EBS", "EC2", "RDS", "S3", "DynamoDB", ...).
AWS returns EVERY supported type on read, so list every type
deliberately - a type missing here shows as a perpetual
difference in plans (the provider requires the full map).

Values persist after destroy (no-op delete) - to opt a type out,
apply it as false before destroying.

- rule: {"map":{"minPairs":"1"}}

### spec.regionSettings.resourceTypeManagementPreference

`map<string, bool>`

Which resource types AWS Backup MANAGES fully (e.g. lets Backup
own DynamoDB's advanced features). Once set at AWS, the
preference cannot be cleared back to unset - only flipped per
type. Omit to leave management preferences untouched.

## Validation Rules

- `spec.at_least_one_arm`: configure at least one of global / region_settings - an instance managing neither is dead configuration

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBackupSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.account_id` | `string` | The AWS account ID the global arm manages (the global settings resource's own identity at AWS). |
| `status.outputs.region` | `string` | The region the region arm manages (the region settings resource's own identity at AWS). |

## See Also

- [Overview](../README.md)
