# AwsIamAccountSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsIamAccountSettingsSpec defines IAM's account-level settings.

This is a SETTINGS SINGLETON for a GLOBAL service: IAM keeps
exactly one of each setting per ACCOUNT (not per region), so deploy
at most one instance per account - two instances fight over the
same account objects. metadata.name never reaches AWS; the account
is the identity.

Destroy semantics DIFFER per arm (each taught on its arm):
  - account_alias: destroy truly DELETES the alias (sign-in URLs
    revert to the bare account ID);
  - password_policy: destroy RESETS the policy to AWS's defaults
    (6-character minimum, nothing else required);
  - sts: destroy is a NO-OP - the last-applied token version
    persists (reverting is an apply with the other version).

## Example

```yaml
# Canonical AwsIamAccountSettings example (hack/dev manifest and
# refgen Example source): a hardened password policy plus the v2 STS
# token version. The alias arm is deliberately omitted here - applying
# an alias REPLACES whatever alias the account already had, which
# changes the sign-in URL everyone uses (see the spec comment).
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamAccountSettings
metadata:
  name: iam-account-baseline
  id: iam-account-baseline
  org: test-org
  env: dev
spec:
  region: us-east-1
  passwordPolicy:
    minimumPasswordLength: 14
    requireLowercaseCharacters: true
    requireNumbers: true
    requireSymbols: true
    requireUppercaseCharacters: true
    allowUsersToChangePassword: true
    maxPasswordAge: 90
    passwordReusePrevention: 24
  sts:
    globalEndpointTokenVersion: v2Token
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.accountAlias` | `string` |  |  |  |
| `spec.passwordPolicy` | `AwsIamAccountSettingsPasswordPolicy` |  |  |  |
| `spec.passwordPolicy.minimumPasswordLength` | `int32` |  |  |  |
| `spec.passwordPolicy.requireLowercaseCharacters` | `bool` |  |  |  |
| `spec.passwordPolicy.requireNumbers` | `bool` |  |  |  |
| `spec.passwordPolicy.requireSymbols` | `bool` |  |  |  |
| `spec.passwordPolicy.requireUppercaseCharacters` | `bool` |  |  |  |
| `spec.passwordPolicy.allowUsersToChangePassword` | `bool` |  |  |  |
| `spec.passwordPolicy.maxPasswordAge` | `int32` |  |  |  |
| `spec.passwordPolicy.passwordReusePrevention` | `int32` |  |  |  |
| `spec.passwordPolicy.hardExpiry` | `bool` |  |  |  |
| `spec.sts` | `AwsIamAccountSettingsSts` |  |  |  |
| `spec.sts.globalEndpointTokenVersion` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the settings.
IAM is a global service - these settings are account-wide - but
every AWS API call is still made against a regional endpoint.
Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.accountAlias

`string`

The account's sign-in alias: makes the console sign-in URL
https://<alias>.signin.aws.amazon.com/console. An account has
exactly ONE alias, and aliases are GLOBALLY unique across all of
AWS - applying this arm REPLACES whatever alias the account
already had, which changes the sign-in URL everyone uses.
3-63 characters (the pattern encodes the length): lowercase
letters, digits, and single hyphens. Destroying this arm deletes
the alias (sign-in reverts to the bare account ID).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9a-z][0-9a-z-]{2,62}$"}}

### spec.passwordPolicy

`AwsIamAccountSettingsPasswordPolicy`

The account's password policy for IAM user console passwords.
AWS keeps ONE policy per account and treats every update as a
full replacement: an arm field left unset is AWS's default, not
"keep the current setting". Destroying this arm RESETS the
policy to AWS's defaults (6-character minimum, nothing else
required).

### spec.passwordPolicy.minimumPasswordLength

`int32` · optional (explicit presence)

Minimum password length, 6-128. Unset = AWS default (6).

- rule: {"int32":{"lte":128,"gte":6}}

### spec.passwordPolicy.requireLowercaseCharacters

`bool`

Require at least one lowercase letter. Unset/false = not
required (AWS default).

### spec.passwordPolicy.requireNumbers

`bool`

Require at least one digit. Unset/false = not required.

### spec.passwordPolicy.requireSymbols

`bool`

Require at least one non-alphanumeric symbol. Unset/false = not
required.

### spec.passwordPolicy.requireUppercaseCharacters

`bool`

Require at least one uppercase letter. Unset/false = not
required.

### spec.passwordPolicy.allowUsersToChangePassword

`bool` · optional (explicit presence)

Let users change their own passwords. Unset = AWS default
(true); explicit false forbids self-service changes.

### spec.passwordPolicy.maxPasswordAge

`int32` · optional (explicit presence)

Days before a password expires and must be rotated, 1-1095.
Unset = passwords never expire.

- rule: {"int32":{"lte":1095,"gte":1}}

### spec.passwordPolicy.passwordReusePrevention

`int32` · optional (explicit presence)

How many previous passwords a new one must differ from, 1-24.
Unset = reuse is not prevented.

- rule: {"int32":{"lte":24,"gte":1}}

### spec.passwordPolicy.hardExpiry

`bool`

When a password expires, require an administrator reset instead
of letting the user set a new one at sign-in. Unset/false = users
reset their own expired passwords. Only meaningful with
max_password_age.

### spec.sts

`AwsIamAccountSettingsSts`

The account's STS settings. Destroying this arm is a NO-OP - the
last-applied token version persists.

### spec.sts.globalEndpointTokenVersion

`string`

Which token version the GLOBAL STS endpoint
(sts.amazonaws.com) issues. v1Token (the AWS default) is smaller
but only valid in default-enabled regions; v2Token works in ALL
regions including opt-in ones - the recommended posture for
accounts using opt-in regions. Regional endpoints always issue
v2 regardless.

- rule: {"string":{"in":["v1Token","v2Token"]}}

## Validation Rules

- `spec.at_least_one_arm`: configure at least one of account_alias / password_policy / sts - an instance managing none is dead configuration
- `spec.alias_no_double_hyphen`: account_alias cannot contain consecutive hyphens
- `spec.alias_no_trailing_hyphen`: account_alias cannot end with a hyphen

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsIamAccountSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.account_id` | `string` | The account these settings belong to (the singleton's identity; also the STS preference's resource ID at the provider). |
| `status.outputs.account_alias` | `string` | The applied sign-in alias (empty when the arm is unset). The console sign-in URL is https://<alias>.signin.aws.amazon.com/console. |
| `status.outputs.expire_passwords` | `string` | Whether the applied password policy expires passwords (AWS derives it from max_password_age; empty when the arm is unset). |

## See Also

- [Overview](../README.md)
