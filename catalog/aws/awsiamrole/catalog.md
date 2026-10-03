# AWS IAM Role

Deploys an IAM role with a configurable trust policy, managed policy attachments, inline policies, and an optional permissions boundary. The trust side takes exactly one of two forms: a free-form JSON `trustPolicy` document, or a typed `oidcTrust` arm that composes the `sts:AssumeRoleWithWebIdentity` document from an IAM OIDC provider's outputs — the form that makes keyless workload identity (EKS IRSA, GitHub Actions) deployable in one run, since the provider's ARN only exists after the provider is created. Name and path are create-only; the trust policy, description, session duration, boundary, and policy attachments all update in place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **IAM Role** -- created with the trust policy (written directly, or composed from the `oidcTrust` arm), description, IAM path, and session-duration ceiling
- **Managed Policy Attachments** -- one attachment resource per entry in `managedPolicyArns`, keyed by the policy ARN itself so reordering the list is a no-op instead of a transient detach/re-attach on a live role
- **Inline Policies** -- one inline policy per entry in `inlinePolicies`, embedded directly on the role with the specified policy name and JSON document
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the role

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **IAM permissions** -- the credentials used by the Provider Connection must have `iam:CreateRole`, `iam:AttachRolePolicy`, `iam:PutRolePolicy`, and related IAM permissions.
- **Managed policy ARNs** -- any literal policy ARNs listed in `managedPolicyArns` must exist in the account before deployment; referenced AwsIamPolicy resources deploy first automatically.
- **IAM OIDC provider** (only for `oidcTrust`) -- the provider referenced by `oidcTrust.providerArn` must exist or deploy in the same run as an AwsIamOidcProvider.

## Deploy

### Console

Open the deployment store, find **AWS IAM Role**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Lambda Execution Role** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamRole
metadata:
  name: eks-service-role
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: "Service role for EKS cluster management"
  path: "/"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: eks.amazonaws.com
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/AmazonEKSClusterPolicy
```

```shell
planton apply -f iam-role.yaml
```

This creates an IAM role that EKS can assume, with the `AmazonEKSClusterPolicy` managed policy attached (literal ARNs take the `value:` form; references to an AwsIamPolicy take `valueFrom:`). An Infra Job tracks the provisioning in real time.

### InfraChart

When the role deploys alongside a customer-managed policy in one chart, wire the attachment via ValueFromRef:

```yaml
spec:
  region: us-west-2
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: lambda.amazonaws.com
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - valueFrom:
        kind: AwsIamPolicy
        name: orders-service-policy
        fieldPath: status.outputs.policy_arn
```

The InfraPipeline resolves the dependency graph, deploys the policy first, then creates the role with the attachment in place. The same wiring works for `permissionsBoundary` and for `oidcTrust.providerArn` / `oidcTrust.providerUrl` against an AwsIamOidcProvider.

## Key Configuration

These are the most important decisions when configuring an IAM role. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Trust: free-form or typed OIDC** -- Exactly one of `trustPolicy` or `oidcTrust` is required. Most roles write the JSON document directly: prefer exact principals and add conditions (`aws:SourceAccount`, `aws:SourceArn`, `sts:ExternalId`) to prevent confused-deputy access. Roles assumed through OIDC federation use `oidcTrust` instead -- the module composes the web-identity document from the provider's outputs, which is the only way to trust a provider created in the same run.

**Scoping OIDC subjects** -- A role trusting every token a provider mints is an open door; `oidcTrust` requires at least one subject. Exact subjects (`subjects`, e.g. `system:serviceaccount:<namespace>:<serviceaccount>` for EKS IRSA) and wildcard subjects (`wildcardSubjects`, e.g. `repo:my-org/my-repo:*` for GitHub Actions) render as separate statements deliberately: IAM ANDs condition operators within a statement, so mixing them would require a token to satisfy both at once.

**Managed vs. inline policies** -- Use `managedPolicyArns` for reusable, centrally maintained policies (both AWS-managed literals and AwsIamPolicy references). Use `inlinePolicies` for role-specific permissions that should not be shared -- an inline policy lives and dies with the role. Attachments reconcile individually: adding or removing an entry attaches or detaches without touching the role.

**IAM path** -- The `path` field defaults to `/`. Use IAM paths like `/service-roles/` to organize roles and enable path-based conditions (e.g. grant `iam:PassRole` only for `/service-roles/*`). The path is create-only: it is part of the role's ARN, so changing it replaces the role.

**Session duration** -- `maxSessionDuration` caps how long AssumeRole sessions on this role may last. Unset keeps AWS's 1-hour default (right for service roles); raise it up to 43200 seconds (12 hours) for long-running human SSO or CI sessions. Updates in place.

**Permissions boundary** -- `permissionsBoundary` optionally caps the role's maximum permissions: effective permissions are the INTERSECTION of the boundary and the role's attached policies. Reference an AwsIamPolicy's `policy_arn` output or pass a literal policy ARN -- the standard guardrail when teams mint their own roles.

**Deletion posture** -- `forceDetachPolicies` is off by default: deleting the role fails if out-of-band policy attachments exist, surfacing them instead of silently severing another owner's wiring. Turn it on for ephemeral or CI-owned roles where teardown must always succeed.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamPolicy** | `managedPolicyArns[]` | `status.outputs.policy_arn` |
| **AwsIamPolicy** | `permissionsBoundary` | `status.outputs.policy_arn` |
| **AwsIamOidcProvider** | `oidcTrust.providerArn` | `status.outputs.provider_arn` |
| **AwsIamOidcProvider** | `oidcTrust.providerUrl` | `status.outputs.provider_url` |

Literal values are also accepted on all four fields -- how AWS-managed policy ARNs and pre-existing OIDC providers attach.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `role_arn` | Amazon Resource Name of the IAM role | EKS cluster service role, Lambda execution role, ECS task role, Step Functions execution role |
| `role_name` | IAM role name (mirrors `metadata.name`) | An AwsIamInstanceProfile's `role` field; AWS CLI and console lookups |
| `role_id` | The role's stable unique ID (never reused, unlike names) | `aws:userid` policy conditions and audit trails that must survive role re-creation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Lambda execution role** -- A role with `lambda.amazonaws.com` trust and the `AWSLambdaBasicExecutionRole` managed policy for CloudWatch Logs access. The standard starting point for Lambda functions. Start from the **Lambda Execution Role** preset.

**ECS task execution role** -- A role with `ecs-tasks.amazonaws.com` trust and the `AmazonECSTaskExecutionRolePolicy` managed policy for pulling container images and writing logs. Start from the **ECS Task Execution Role** preset.

**EC2 SSM role** -- A role with `ec2.amazonaws.com` trust and the `AmazonSSMManagedInstanceCore` managed policy for Systems Manager Session Manager access. Eliminates the need for SSH bastion hosts; wrap it in an AwsIamInstanceProfile to deliver it to instances. Start from the **EC2 SSM and CloudWatch Role** preset.

**Keyless CI federation** -- An `oidcTrust` role against a GitHub Actions OIDC provider with a `wildcardSubjects` entry like `repo:my-org/my-repo:*`. No long-lived access keys in CI: the pipeline exchanges its OIDC token for temporary role credentials on every run.

## Works With

- [**AWS IAM Policy**](/infra-catalog/aws-iam-policy) -- customer-managed policies attached via `managedPolicyArns` or used as the `permissionsBoundary`
- [**AWS IAM OIDC Provider**](/infra-catalog/aws-iam-oidc-provider) -- the federated identity provider the `oidcTrust` arm references
- [**AWS IAM Instance Profile**](/infra-catalog/aws-iam-instance-profile) -- wraps this role for delivery to EC2 instances (the only service that needs the wrapper)
- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) -- consumes `role_arn` as its cluster service role
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- consumes `role_arn` as its execution role
- [**AWS ECS Service**](/infra-catalog/aws-ecs-service) -- consumes `role_arn` for task and execution roles
