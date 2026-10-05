# AWS IAM User

Deploys an IAM user with managed policy attachments, inline policies, an optional permissions boundary, and explicit access-key control. IAM users carry permanent credentials, so this kind targets the narrow cases temporary role credentials cannot cover -- external CI systems without OIDC federation, legacy tooling, break-glass access; prefer an IAM role wherever federation is possible. One active access key is created by default and exported as sensitive outputs, and `accessKeyStatus` flips it between Active and Inactive in place -- the rotation lever.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **IAM User** -- created with the specified username matching the pattern `[a-zA-Z0-9+=,.@_-]{1,64}`, at the optional IAM `path` (defaults to `/`), with an optional `permissionsBoundary` capping its maximum permissions
- **Managed Policy Attachments** -- one attachment per entry in `managedPolicyArns`, linking AWS-managed or customer-managed policies to the user. Each entry is a literal ARN or a reference to an AwsIamPolicy's `policy_arn` output.
- **Inline Policies** -- one inline policy per entry in `inlinePolicies`, embedded directly on the user with the specified policy name and JSON document
- **Access Key** -- created only when `disableAccessKeys` is `false` (default); generates an access key pair for programmatic AWS API access. The secret access key is base64-encoded in outputs. `accessKeyStatus` flips the key between `Active` (default) and `Inactive` in place -- the rotation lever: an Inactive key keeps its id and secret while AWS rejects requests signed with it, so you can prove nothing depends on it before deleting.
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the user

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **IAM permissions** -- the credentials used by the Provider Connection must have `iam:CreateUser`, `iam:AttachUserPolicy`, `iam:PutUserPolicy`, `iam:CreateAccessKey`, and related IAM permissions.
- **Managed policy ARNs** -- any AWS-managed or customer-managed policy ARNs listed as literals in `managedPolicyArns` must exist in the account before deployment. Policies managed in Planton are referenced instead (see the consumes table below) and resolve at deploy time.

## Deploy

### Console

Open the deployment store, find **AWS IAM User**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **CI/CD Pipeline User** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamUser
metadata:
  name: deployer
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  userName: github-actions-deployer
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser
    - value: arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
```

```shell
planton apply -f iam-user.yaml
```

This creates an IAM user with ECR push access and S3 read-only access, plus an active access key pair (literal ARNs take the `value:` form; references to an AwsIamPolicy take `valueFrom:`). No inline policies are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When the user deploys alongside a customer-managed policy in one chart, wire the attachment via ValueFromRef:

```yaml
spec:
  region: us-west-2
  userName: deploy-bot
  managedPolicyArns:
    - valueFrom:
        kind: AwsIamPolicy
        name: deploy-bot-policy
        fieldPath: status.outputs.policy_arn
```

The InfraPipeline resolves the dependency graph, deploys the policy first, then creates the user with the attachment in place. The same wiring works for `permissionsBoundary`.

## Key Configuration

These are the most important decisions when configuring an IAM user. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Managed vs. inline policies** -- Use `managedPolicyArns` for reusable, centrally maintained policies (both AWS-managed and customer-managed). Use `inlinePolicies` for user-specific permissions that should not be shared. Managed policies are easier to audit across users; inline policies are deleted when the user is deleted.

**Access key generation** -- By default, an access key pair is created for programmatic access. Set `disableAccessKeys: true` for identity-only users (console access, federation) or monitoring integrations that use external credential delivery. The secret access key is base64-encoded in outputs and should be stored securely.

**Username conventions** -- The `userName` field accepts 1-64 characters matching `[a-zA-Z0-9+=,.@_-]`. Use descriptive names that identify the purpose (e.g., `github-actions-deployer`, `datadog-readonly`) rather than generic names. Unlike most IAM identifiers, the name is renameable in place -- AWS keeps the same underlying user and rewrites its ARN.

**IAM path** -- `path` organizes users for wildcard policy matching: a permission scoped to `user/ci/*` covers every user under the `/ci/` path. Must begin and end with `/` (e.g. `/ci/`); defaults to `/`. Users, unlike roles and policies, can move paths in place.

**Permissions boundary** -- `permissionsBoundary` names a managed policy that caps what the user can EVER do; effective permissions are the intersection of the boundary and the attached policies. Especially valuable on users, whose access keys are long-lived. Reference an AwsIamPolicy's `policy_arn` output or pass a literal ARN.

**Deletion posture** -- `forceDestroy` (off by default) controls what happens when credentials created outside this resource exist at delete time -- console login profiles, extra access keys, MFA devices, SSH keys, signing certificates. Off, deletion fails loudly and surfaces them; on, teardown always succeeds. Turn it on for ephemeral or CI-owned users.

**Least privilege** -- Scope policies to the minimum permissions required. Prefer specific managed policies over broad ones like `AdministratorAccess`. Use inline policies for fine-grained, resource-level permissions (e.g., allowing ECS updates only on specific clusters).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamPolicy** (optional) | `managedPolicyArns[]` | `status.outputs.policy_arn` |
| **AwsIamPolicy** (optional) | `permissionsBoundary` | `status.outputs.policy_arn` |

Both fields also accept literal ARNs -- literals are how AWS-managed policies like `arn:aws:iam::aws:policy/ReadOnlyAccess` attach.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `user_arn` | Amazon Resource Name of the IAM user | Resource policies and cross-account access grants naming this user as principal |
| `user_id` | Stable unique ID of the IAM user (never reused, unlike names) | `aws:userid` policy conditions that must survive user re-creation |
| `access_key_id` | Access key ID (present when access keys are enabled) | CI/CD pipeline `AWS_ACCESS_KEY_ID` configuration |
| `secret_access_key` | Base64-encoded secret access key (sensitive) | CI/CD pipeline `AWS_SECRET_ACCESS_KEY` configuration |

`user_name`, `console_url`, and `access_key_status` are also exported: the name mirrors `metadata.name`, the console URL is the account sign-in page, and the key status echoes the rotation-lever position for verification against `ListAccessKeys`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CI/CD pipeline user** -- ECR power user, S3 read-only, and ECS deployment permissions with access keys enabled. The standard configuration for GitHub Actions, GitLab CI, or Jenkins pipelines that build container images and deploy to ECS. Start from the **CI/CD Pipeline User** preset.

**Read-only service user** -- Broad read-only access with access keys disabled. Suitable for third-party monitoring tools (Datadog, New Relic) or audit integrations that only need to observe resources without making changes. Start from the **Read-Only Service User** preset.

## Works With

- [**AWS IAM Policy**](/infra-catalog/aws-iam-policy) -- provides customer-managed policies for attachment via `managedPolicyArns` and the permissions boundary via `permissionsBoundary`