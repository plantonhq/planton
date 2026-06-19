# AWS IAM OIDC Provider

Registers an OpenID Connect (OIDC) identity provider in AWS IAM. This is the trust anchor for keyless, web-identity federation: it lets an external issuer's short-lived tokens be exchanged for AWS credentials through STS `AssumeRoleWithWebIdentity`, so workloads and pipelines never hold long-lived AWS access keys. The component creates the provider from an issuer URL, a list of allowed client IDs (audiences), and optional CA thumbprints, then exports the provider ARN for IAM roles to trust.

## What Gets Created

When you deploy an AwsIamOidcProvider resource, OpenMCF provisions:

- **IAM OIDC Provider** — an `iam.OpenIdConnectProvider` (`aws_iam_openid_connect_provider`) registered under the issuer `url`, scoped to the supplied `clientIdList`, optionally pinned to `thumbprintList`

That single resource is the trust anchor. Access itself is granted by a separate `AwsIamRole` whose trust policy references this provider's ARN.

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An OIDC issuer URL** — for EKS this is the cluster's OIDC issuer; for CI it is the platform issuer (e.g. `https://token.actions.githubusercontent.com`)
- **The audience(s)** the issuer's tokens carry in the `aud` claim (commonly `sts.amazonaws.com`)

## Quick Start

Create a file `oidc-provider.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamOidcProvider
metadata:
  name: github-actions-oidc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsIamOidcProvider.github-actions-oidc
spec:
  region: us-east-1
  url:
    value: https://token.actions.githubusercontent.com
  clientIdList:
    - sts.amazonaws.com
```

Deploy:

```shell
openmcf apply -f oidc-provider.yaml
```

This registers GitHub Actions as a trusted OIDC issuer. Next, create an `AwsIamRole` whose trust policy references the exported `provider_arn`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region used to configure the provider (IAM is global). | Required, non-empty |
| `url` | `StringValueOrRef` | The OIDC issuer URL (`iss` claim). Inline value or a reference to an `AwsEksCluster`'s `status.outputs.oidc_issuer_url`. | Required |
| `clientIdList` | `string[]` | Allowed client IDs / audiences (`aud` claim). | At least 1, unique, each 1–255 chars |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `thumbprintList` | `string[]` | `[]` (AWS-derived) | SHA-1 thumbprints (40 hex chars each) of the issuer's root CA. Omit for well-known CAs; AWS derives them. Must be unique. |

## Examples

### EKS IRSA (referencing the cluster)

Wire the OIDC provider directly onto an EKS cluster so IRSA works without copying the issuer URL by hand:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamOidcProvider
metadata:
  name: eks-irsa
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsIamOidcProvider.eks-irsa
spec:
  region: us-west-2
  url:
    valueFrom:
      kind: AwsEksCluster
      name: my-eks-cluster
      fieldPath: status.outputs.oidc_issuer_url
  clientIdList:
    - sts.amazonaws.com
```

### Generic Issuer with an Explicit Thumbprint

For an issuer whose root CA is not publicly trusted, pin the thumbprint:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamOidcProvider
metadata:
  name: partner-oidc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsIamOidcProvider.partner-oidc
spec:
  region: eu-west-1
  url:
    value: https://oidc.partner.example.com
  clientIdList:
    - my-aws-integration
  thumbprintList:
    - 990f4193972f2becf12ddeda5237f9c952f20d9e
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `provider_arn` | `string` | ARN of the IAM OIDC provider (referenced as a `Federated` principal in IAM role trust policies) |
| `provider_url` | `string` | The issuer URL AWS stored, with the `https://` scheme stripped |

## Related Components

- [AwsIamRole](/docs/catalog/aws/awsiamrole) — the role whose trust policy references `provider_arn` to grant web-identity access
- [AwsEksCluster](/docs/catalog/aws/awsekscluster) — exports the `oidc_issuer_url` this provider consumes for IRSA
- [AwsIamUser](/docs/catalog/aws/awsiamuser) — the long-lived-credential alternative this component is designed to make unnecessary
