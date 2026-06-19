# EKS IRSA OIDC Provider

This preset registers an EKS cluster's OIDC issuer as a trusted IAM identity provider, enabling IAM Roles for Service Accounts (IRSA). It references the cluster directly, so the issuer URL is resolved from the cluster's outputs at deploy time -- no copy-pasting the issuer endpoint.

## When to Use

- Any EKS cluster where pods should receive scoped, auto-rotating AWS credentials via ServiceAccounts
- The required first step before creating IRSA roles (an EKS cluster provisions an issuer but does not register the IAM provider)

## Key Configuration Choices

- **Cluster reference** (`url.valueFrom` -> `AwsEksCluster.status.outputs.oidc_issuer_url`) -- composes the provider onto the cluster as a first-class node
- **Audience** (`sts.amazonaws.com`) -- the audience EKS IRSA tokens carry
- **No thumbprints** -- EKS issuers are backed by a well-known CA, so AWS derives the thumbprint automatically

## Placeholders to Replace

- `<aws-region>` -- the AWS region used to configure the provider
- `<eks-cluster-name>` -- the `metadata.name` of the `AwsEksCluster` resource to reference

## Related Presets

- **02-github-actions** -- Use instead for keyless GitHub Actions deployments
- **03-generic-issuer** -- Use instead for an issuer whose CA is not publicly trusted
