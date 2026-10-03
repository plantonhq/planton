# AWS IAM Account Settings

Manages an AWS account's IAM-wide settings in one place: the console sign-in alias, the password policy for IAM users, and the STS global-endpoint token version. This is a settings singleton for a global service -- IAM keeps exactly one of each setting per account, so deploy at most one instance per account; two instances fight over the same account objects. Each arm carries its own destroy contract: the alias deletes, the password policy resets to AWS defaults, and the STS preference persists.

## What Gets Created

This component adopts and configures IAM settings objects that exist on every AWS account -- nothing new is created at AWS; the module writes the values of account-level singletons:

- **Sign-In Alias** -- managed only when `accountAlias` is set; the friendly console URL (`https://<alias>.signin.aws.amazon.com/console`). An account has exactly one alias, so applying this arm REPLACES whatever alias existed
- **Password Policy** -- managed only when `passwordPolicy` is set; the account's one IAM-user password policy, replaced whole on every apply
- **STS Preferences** -- managed only when `sts` is set; which token version the global STS endpoint issues

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with IAM permissions on the target account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing to pre-create -- the settings objects always exist; this component manages their values.
- Before applying the alias arm, check `aws iam list-account-aliases`: applying replaces the current alias, and every bookmarked sign-in URL changes with it.

## Deploy

### Console

Open the deployment store, find **AWS IAM Account Settings**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the three arms: alias, password policy, and STS. Start from the **Hardened Password Policy** preset in the [Presets](#presets) tab for the CIS-benchmark-shaped posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamAccountSettings
metadata:
  name: iam-account-settings
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  accountAlias: acme-corp-prod
  passwordPolicy:
    minimumPasswordLength: 14
    requireLowercaseCharacters: true
    requireNumbers: true
    requireSymbols: true
    requireUppercaseCharacters: true
    maxPasswordAge: 90
    passwordReusePrevention: 24
  sts:
    globalEndpointTokenVersion: v2Token
```

```shell
planton apply -f aws-iam-account-settings.yaml
```

This sets the `acme-corp-prod` sign-in alias, applies a 14-character four-character-class password policy with 90-day rotation and 24-password reuse memory, and switches the global STS endpoint to v2 tokens valid in every region. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring IAM account settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One instance per account, ever** -- the identity is the account itself; `metadata.name` never reaches AWS. A second instance targeting the same account does not error -- it silently fights the first over the same singleton objects, each apply overwriting the other's values. Pick one home for these settings per account.

**The alias trap** -- an account has exactly one alias, aliases are globally unique across ALL of AWS, and applying `accountAlias` replaces whatever alias the account already had -- everyone's bookmarked sign-in URL changes at that moment. Namespace aliases (`<company>-<env>`) so they never collide, and check the current alias before the first apply.

**The password policy has no partial update** -- AWS's update is a full replacement: an unset field means AWS's default (6-character minimum, no expiry, no reuse prevention), never "keep the current setting". When adopting an account that already has a policy, capture its full posture in the spec first, or the first apply silently drops every omitted setting to defaults. Existing passwords age against `maxPasswordAge` from their last change, so enforcement on active users begins immediately.

**`hardExpiry` is a lockout policy** -- with it, an expired password requires an administrator reset instead of a self-service change at sign-in. Pair it deliberately with `allowUsersToChangePassword` and a realistic `maxPasswordAge`, or expect lockout tickets the week rotation lands.

**v1Token vs v2Token** -- the global STS endpoint issues v1 tokens by default: smaller, but only valid in default-enabled regions. `v2Token` works in every region including opt-in ones -- the right posture for any account enabling opt-in regions. The setting affects the global endpoint only; regional STS endpoints always issue v2.

**Three arms, three destroy contracts** -- destroying the alias arm deletes the alias (sign-in reverts to the bare account ID); destroying the password-policy arm resets the policy to AWS defaults; destroying the STS arm is a no-op -- the last-applied token version persists, and reverting means applying the other version.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- it configures account-level singletons and takes no references to other Infra Components.

### What This Kind Provides

`status.outputs` echoes the applied state -- `account_id` (the singleton's identity), `account_alias`, and `expire_passwords` (derived by AWS from `maxPasswordAge`). These are audit and onboarding-documentation values, not composition inputs: no downstream Infra Component consumes IAM account settings by reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hardened password baseline** -- the CIS-benchmark-shaped posture (14+ characters, all four character classes, 90-day rotation, 24-password reuse memory) with self-service changes kept on so users fix their own expiry, plus v2 STS tokens. The security baseline auditors look for first in any account with IAM user console access. Start from the **Hardened Password Policy** preset.

**Readable sign-in URLs per account** -- an alias per account following a naming convention, so "which account is this" is obvious at the sign-in page across a multi-account estate. Start from the **Sign-In Alias** preset.

**Adopting an already-configured account** -- read the account's current alias and password policy first, write them into the spec verbatim, apply (a no-op), and only then evolve the posture. Skipping the capture step is how an adoption apply silently weakens a policy.

## Works With

- [**AWS IAM User**](/infra-catalog/aws-iam-user) -- the users whose console passwords the policy arm governs
- [**AWS Organization**](/infra-catalog/aws-organization) -- centralized root-access management is deliberately NOT in this kind; it is a management-account act modeled on the organization resource
