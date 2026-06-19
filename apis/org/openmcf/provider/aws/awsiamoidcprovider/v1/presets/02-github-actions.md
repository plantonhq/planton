# GitHub Actions OIDC Provider

This preset registers GitHub Actions as a trusted OIDC identity provider, enabling keyless deployments from CI. Workflows assume an IAM deploy role with a short-lived token on every run, removing static AWS access keys from GitHub secrets entirely.

## When to Use

- Any GitHub Actions workflow that deploys to or reads from AWS
- Replacing long-lived `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` repository secrets with federation

## Key Configuration Choices

- **GitHub issuer** (`https://token.actions.githubusercontent.com`) -- the same issuer for all GitHub-hosted repositories
- **Audience** (`sts.amazonaws.com`) -- the audience AWS expects for GitHub Actions tokens
- **No thumbprints** -- GitHub's issuer is backed by a well-known CA, so AWS derives the thumbprint automatically

## Placeholders to Replace

- `<aws-region>` -- the AWS region used to configure the provider

## Next Step

Create an `AwsIamRole` whose trust policy references this provider's `provider_arn` as a `Federated` principal, and restrict `token.actions.githubusercontent.com:sub` to the exact repo/branch/environment allowed to deploy (e.g. `repo:my-org/my-repo:ref:refs/heads/main`).

## Related Presets

- **01-eks-irsa** -- Use instead to enable IRSA on an EKS cluster
- **03-generic-issuer** -- Use instead for an issuer whose CA is not publicly trusted
