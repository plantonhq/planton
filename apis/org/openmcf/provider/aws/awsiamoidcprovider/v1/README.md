# AwsIamOidcProvider

An AWS IAM OpenID Connect (OIDC) identity provider is the trust anchor that lets an external OIDC issuer's short-lived tokens be exchanged for AWS credentials via STS `AssumeRoleWithWebIdentity` -- with no long-lived access keys. It is the missing link that turns an EKS cluster into an IRSA-capable cluster and lets CI systems (GitHub Actions, GitLab) deploy to AWS without stored secrets.

## Spec fields (summary)
- region: AWS region used to configure the provider (IAM is global; this only selects the IAM/STS endpoint)
- url: The OIDC issuer URL (the `iss` claim). A `StringValueOrRef` that defaults to an `AwsEksCluster`'s `status.outputs.oidc_issuer_url`, so an EKS cluster and its IRSA trust anchor compose by reference
- client_id_list: The allowed client IDs / audiences (the `aud` claim); for EKS IRSA this is `sts.amazonaws.com`
- thumbprint_list: Optional SHA-1 thumbprints of the issuer's root CA; omit for well-known CAs and AWS derives them

## Stack outputs
- provider_arn: ARN of the OIDC provider; referenced as a `Federated` principal in IAM role trust policies
- provider_url: The issuer URL AWS stored, with the `https://` scheme stripped (used to build `<url>:sub` / `<url>:aud` trust conditions)

## How it works
This resource is orchestrated by the OpenMCF CLI as part of a stack-update. The CLI validates your manifest, generates stack inputs, and invokes IaC backends in this repo:
- Pulumi (Go modules under iac/pulumi)
- Terraform (modules under iac/tf)

The OIDC provider does not grant any access on its own. It establishes *trust* in an issuer. Access is granted by an `AwsIamRole` whose trust policy names this provider's ARN as a `Federated` principal and constrains the `sub`/`aud` claims.

## Composition: the federation triangle
```
AwsEksCluster (oidc_issuer_url)
      |
      v
AwsIamOidcProvider (provider_arn)
      |
      v
AwsIamRole (Federated trust)  ->  Kubernetes ServiceAccount (IRSA)
```

Point `spec.url` at an `AwsEksCluster` reference and the issuer URL flows through automatically; reference `provider_arn` from an `AwsIamRole` trust policy and the loop is closed -- all as first-class, independently ownable nodes.

## Common use cases
- **EKS IRSA**: Map a Kubernetes ServiceAccount to an IAM role so pods get scoped AWS credentials automatically
- **GitHub Actions federation**: Let workflows assume a deploy role via OIDC, removing static AWS keys from CI
- **GitLab CI / other CI**: Same keyless pattern for any OIDC-capable pipeline
- **Self-hosted / partner issuers**: Federate any standards-compliant OIDC issuer (supply a thumbprint if its CA is not publicly trusted)

## Security best practices
- **Constrain the audience**: Always set `client_id_list` (e.g. `sts.amazonaws.com`); never federate an issuer without pinning the `aud`
- **Constrain the subject in the role**: In the consuming `AwsIamRole`, restrict `<provider-url>:sub` to the exact ServiceAccount or CI repo/branch
- **Prefer well-known CAs**: Leave `thumbprint_list` empty so AWS validates TLS against its trusted store; supply thumbprints only for private CAs
- **One provider per issuer**: AWS allows a single OIDC provider per unique URL per account; share it across roles rather than duplicating

## References
- IAM OIDC identity providers: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_oidc.html
- EKS IRSA: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
- GitHub Actions OIDC with AWS: https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services
- AssumeRoleWithWebIdentity: https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRoleWithWebIdentity.html
- Obtaining the root CA thumbprint: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc_verify-thumbprint.html
- Research documentation: [docs/README.md](docs/README.md)
