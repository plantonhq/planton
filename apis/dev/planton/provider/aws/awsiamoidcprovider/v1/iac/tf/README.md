# Terraform Module to Deploy AwsIamOidcProvider

This module provisions an AWS IAM OpenID Connect (OIDC) identity provider: the trust
anchor that lets an external OIDC issuer's tokens be exchanged for AWS credentials via
STS web-identity federation (EKS IRSA, GitHub Actions, GitLab CI, etc.).

It creates a single `aws_iam_openid_connect_provider` from the issuer `url`, the allowed
`client_id_list` (audiences), and an optional `thumbprint_list`. When no thumbprints are
supplied the value is normalized to `null` so AWS derives it from its trusted CA store --
keeping behavior identical to the Pulumi module.

Generated `variables.tf` reflects the proto schema for `AwsIamOidcProvider`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For a ready-to-edit example, see [`hack/manifest.yaml`](../hack/manifest.yaml).
