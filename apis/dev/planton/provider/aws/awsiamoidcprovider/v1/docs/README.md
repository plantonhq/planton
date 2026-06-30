# AWS IAM OIDC Providers: Keyless Federation from First Principles

## Introduction

For two decades, the default way to give something access to AWS was a pair of secrets: an access key ID and a secret access key. Those keys end up in CI configuration, in `.env` files, in Kubernetes secrets, in screenshots, in shell history. They are long-lived, they are copyable, and they leak. Most "AWS breach" headlines trace back to a static credential that escaped.

The IAM OpenID Connect (OIDC) identity provider is AWS's answer to this problem. Instead of storing a secret, a workload proves *who it is* using a short-lived, cryptographically signed token minted by an identity provider it already trusts -- its EKS cluster, GitHub Actions, GitLab CI -- and AWS exchanges that token for temporary credentials via STS `AssumeRoleWithWebIdentity`. No secret is stored, the token expires in minutes, and the trust is scoped to an exact subject.

This document explains what an IAM OIDC provider actually is, the federation it enables, the deployment approaches available, and why Planton models it as a first-class, composable resource rather than burying it inside the EKS cluster.

## What an OIDC Provider Is (and Is Not)

An IAM OIDC provider is a small, declarative trust record in your AWS account. It says: "I trust tokens issued by this issuer URL, for these audiences, validated against this CA." That is all it is. It has three meaningful inputs:

- **url** — the issuer (`iss`) the tokens come from, e.g. `https://oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED` or `https://token.actions.githubusercontent.com`. AWS permits exactly one OIDC provider per unique URL per account.
- **client_id_list** — the audiences (`aud`) the tokens must carry. For EKS IRSA and GitHub Actions this is typically `sts.amazonaws.com`.
- **thumbprint_list** — optional SHA-1 fingerprints of the issuer's root CA certificate. For issuers fronted by a well-known public CA, AWS validates TLS against its own trusted store and you omit this entirely; you only supply thumbprints for private or non-standard CAs.

Critically, an OIDC provider **grants no permissions**. It is a trust anchor, not an authorization. Access is always granted by a separate IAM role whose trust policy names the provider as a `Federated` principal and constrains the token's claims. This separation -- trust here, authorization there -- is what makes the pattern safe and composable.

## The Federation Flow

When a workload wants AWS credentials via OIDC:

1. **Token mint**: The workload obtains a signed JWT from its issuer. For an EKS pod, the kubelet projects a ServiceAccount token whose `iss` is the cluster's OIDC issuer and whose `sub` identifies the namespace/ServiceAccount. For GitHub Actions, the runner requests a token whose `sub` encodes the repo, branch, and environment.
2. **STS exchange**: The workload calls `sts:AssumeRoleWithWebIdentity`, presenting the JWT and a target role ARN.
3. **Issuer validation**: STS checks that an IAM OIDC provider exists for the token's `iss`, validates the signature against the issuer's published JWKS, and checks the `aud` against the provider's `client_id_list`.
4. **Trust validation**: STS evaluates the target role's trust policy -- the `Federated` principal must be this provider's ARN, and any `<provider-url>:sub` / `<provider-url>:aud` conditions must match the token's claims.
5. **Credential issuance**: STS returns temporary credentials scoped to the role, valid for minutes to a few hours.

The two halves -- "register the issuer" (this component) and "trust the subject" (an `AwsIamRole`) -- are independent resources that compose. That is the whole point.

## The Two Flagship Patterns

### EKS IRSA (IAM Roles for Service Accounts)

When you create an EKS cluster, AWS provisions an OIDC issuer endpoint for it -- but it does **not** register an IAM OIDC provider for that issuer. Without that registration, IRSA does not work: STS has no record of the cluster's issuer and rejects every web-identity exchange. Registering the provider is the explicit, required step that turns a cluster into an IRSA-capable cluster.

Once registered, you annotate a Kubernetes ServiceAccount with an IAM role ARN, the role's trust policy pins `<issuer>:sub` to that ServiceAccount, and every pod using it receives scoped, auto-rotating AWS credentials. No node-wide instance profile, no shared secret, no over-broad access.

### CI/CD Federation (GitHub Actions, GitLab, and friends)

CI pipelines are the classic home of leaked AWS keys. With OIDC, you register the CI platform's issuer once, then create a deploy role whose trust policy restricts `<issuer>:sub` to the exact repository, branch, or environment allowed to deploy. The pipeline assumes the role with a freshly minted token on every run. There is no AWS secret in the CI configuration to steal.

## The Maturity Spectrum: How Teams Grant AWS Access

### Level 0: Long-Lived Access Keys

**What it is**: An IAM user with an access key, pasted into CI variables or a Kubernetes secret.

**Why it fails**: The secret is long-lived and copyable. It outlives the person who created it, it is rarely rotated, and it grants the same access from anywhere on earth. Detection of misuse is hard because the key *is* the identity.

**Verdict**: The pattern OIDC providers exist to eliminate. Acceptable only where no OIDC issuer is available.

### Level 1: Node Instance Profiles (for EKS)

**What it is**: Granting permissions to the EC2 node role so every pod on the node inherits them.

**Why it falls short**: All pods on a node share one identity. A low-trust sidecar gets the same access as your most sensitive workload. Least privilege is impossible at the pod level.

**Verdict**: A blunt instrument. IRSA exists precisely to replace it with per-ServiceAccount identity.

### Level 2: Manual OIDC Provider via Console/CLI

**What it is**: Clicking "Add provider" in the IAM console, or a one-off `aws iam create-open-id-connect-provider` call.

**What it improves**: It is the right primitive -- keyless, scoped, short-lived.

**Where it falls short**: It is unversioned and undocumented. Which issuers does this account trust, and why? Was the audience pinned? Did someone fat-finger the thumbprint? In a multi-account, multi-cluster estate, manual providers drift into an unauditable mess.

**Verdict**: Correct mechanism, wrong delivery. Production needs this in code.

### Level 3: Infrastructure as Code (Terraform / Pulumi / CloudFormation)

**What it is**: Declaring the OIDC provider as a versioned resource, reviewed in a pull request, applied by automation.

**Why it works**: The set of trusted issuers becomes auditable and reproducible. Drift is detectable. The provider can be composed with the cluster that produced its issuer URL and the roles that consume its ARN.

**Verdict**: The production baseline. The remaining question is how well the abstraction composes -- which is where Planton focuses.

## Design Decisions in This Component

### `url` is a reference, not just a string

The single most error-prone step in setting up IRSA is copying the cluster's OIDC issuer URL by hand into the provider definition. Planton models `url` as a `StringValueOrRef` that defaults to an `AwsEksCluster`'s `status.outputs.oidc_issuer_url`. You reference the cluster; the issuer URL flows through at deploy time. The cluster, the provider, and the role are three real, independently ownable nodes in the resource graph -- not one opaque bundle. This is the same composition pattern `AwsEksNodeGroup` uses to reference its cluster and node role.

### `provider_arn` is the primary output

The whole reason to create a provider is so a role can trust it. `provider_arn` is exported as the primary, semantically singular output, ready to be referenced as the `Federated` principal in an `AwsIamRole` trust policy. `provider_url` is exported alongside it for building the `:sub` / `:aud` trust conditions.

### Thumbprints are optional, and omission is the default

Since 2021, AWS secures the OIDC TLS connection using its own trusted CA store for issuers backed by well-known CAs, and derives the thumbprint automatically. Forcing operators to compute and paste a SHA-1 fingerprint for EKS or GitHub is needless friction and a common source of breakage. This component leaves `thumbprint_list` empty by default; both the Pulumi and Terraform modules omit the field entirely when no thumbprints are supplied (Terraform normalizes the empty list to `null` so the attribute stays Computed), letting AWS derive it. You supply thumbprints only for private CAs.

### No secrets to manage

An OIDC provider's inputs -- an issuer URL, a list of audiences, public CA fingerprints -- are all public identifiers. There is nothing secret in this resource, which is the entire value proposition: it replaces a stored secret with a verifiable, short-lived token. Accordingly the spec carries no sensitive fields.

## Planton's Dual-Engine Implementation

Planton ships both a Pulumi (Go) module and an OpenTofu/Terraform module for this component, at full behavioral parity. For an identical input they create the same `aws_iam_openid_connect_provider`, with the same name basis, tags, audience list, and thumbprint-omission behavior, and they emit the same `provider_arn` / `provider_url` outputs. You choose the engine that fits your workflow; the result is identical.

## Conclusion

The IAM OIDC provider is a small resource with an outsized security payoff: it is the hinge that lets you delete long-lived AWS keys from your clusters and pipelines and replace them with short-lived, scoped, verifiable identity. Modeled well -- as a composable node that references the cluster that produced its issuer and is referenced by the roles that trust it -- it makes keyless access the easy default rather than an expert-only configuration.

**Register the issuer once, pin the audience and subject, and let short-lived tokens do the rest.**
