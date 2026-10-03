# AWS SES Email Identity

Deploys an Amazon SES (SESv2) email identity — the verified domain or email address an application is allowed to send mail FROM. The identity is the trust anchor of the SES graph: nothing sends through SES without one. A DOMAIN identity is the production shape — it verifies through DNS (the `dkim_tokens` output composes directly into [AwsRoute53DnsRecord](/infra-catalog/aws-route53-dns-record) CNAMEs), signs mail with DKIM, covers every address at the domain, and unlocks a custom MAIL FROM domain for DMARC-aligned SPF. An EMAIL-ADDRESS identity verifies through a confirmation link — quick for testing. The identity inherits its default sending rules from an [AwsSesConfigurationSet](/infra-catalog/aws-ses-configuration-set), and BYODKIM private keys stay in managed secrets — never in the manifest.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SES Email Identity** -- the verified domain or address (create-time immutable; a replacement re-verifies from scratch)
- **DKIM Configuration** -- Easy DKIM (AWS-managed keys, the default) or BYODKIM (your own key pair), on domain identities
- **Custom MAIL FROM Domain** -- the envelope-sender subdomain for strict DMARC SPF alignment, when configured
- **Configuration Set Attachment** -- the default rule group every message from this identity inherits, when referenced
- **Identity Authorization Policies** -- one AWS sub-resource per named cross-account sending grant
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Configuration set first** -- to attach default sending rules, deploy the [AwsSesConfigurationSet](/infra-catalog/aws-ses-configuration-set) before the identity and reference its `configuration_set_name` output.
- **Managed secret for BYODKIM** -- when bringing your own DKIM key, store the base64-encoded private key as an org secret; the spec carries a `$secret/<slug>` reference and the runner resolves it just-in-time at deploy.

### AWS Account

- **DNS control** -- a domain identity stays PENDING until its three DKIM CNAMEs are published; you need control of the domain's DNS (composable with [AwsRoute53DnsRecord](/infra-catalog/aws-route53-dns-record) when the zone lives in Route 53).
- **Mailbox access** -- an email-address identity sends a confirmation link to that mailbox; someone must click it.
- **Sandbox note** -- new SES accounts start sandboxed per region (verified recipients only); request production access per region.

## Deploy

### Console

Open the deployment store, find **AWS SES Email Identity**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Domain with Easy DKIM** preset in the [Presets](#presets) tab to pre-populate the production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesEmailIdentity
metadata:
  name: example-domain
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  emailIdentity: example.com
  mailFrom:
    mailFromDomain: mail.example.com
  configurationSet:
    valueFrom:
      kind: AwsSesConfigurationSet
      name: transactional-prod
      fieldPath: status.outputs.configuration_set_name
```

```shell
planton apply -f ses-email-identity.yaml
```

This verifies `example.com` with AWS-managed Easy DKIM (the default when no DKIM configuration is set), a DMARC-aligned MAIL FROM subdomain, and the transactional configuration set as its default rules. An Infra Job tracks the provisioning in real time.

### InfraChart

When the identity deploys alongside its configuration set in one chart, wire the set reference via ValueFromRef:

```yaml
spec:
  emailIdentity: example.com
  mailFrom:
    mailFromDomain: mail.example.com
  configurationSet:
    valueFrom:
      kind: AwsSesConfigurationSet
      name: transactional-prod
      fieldPath: status.outputs.configuration_set_name
```

The InfraPipeline resolves the dependency graph, deploys the configuration set first, then the identity — and the AwsRoute53DnsRecord resources that publish the identity's `dkim_tokens` outputs as CNAMEs deploy after it, completing verification.

## Key Configuration

These are the most important decisions when configuring an email identity. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domain or address** -- a domain (`example.com`) covers every address at the domain, carries its own DKIM, and unlocks the custom MAIL FROM; a single address (`sender@example.com`) covers only itself and inherits the domain's DKIM when the domain is also verified. The identity string is create-time immutable — changing it replaces the identity and sending stops until the replacement verifies.

**DKIM signing** -- leave it unset to accept AWS's Easy DKIM with a 2048-bit key (the right choice for almost everyone; AWS generates and rotates the key pair). Set an explicit key length only to force 1024 (legacy DNS TXT limits) or to trigger a planned rotation. BYODKIM hands you the whole key lifecycle — generation, DNS publication at `<selector>._domainkey.<domain>`, and rotation.

**Custom MAIL FROM** -- by default SES's own bounce domain is the envelope sender, which fails strict DMARC SPF alignment. A subdomain like `mail.example.com` (with its MX + SPF records) aligns the envelope with your domain. `REJECT_MESSAGE` is the strict posture: if the MX record breaks, nothing leaves unaligned.

**Authorization policies** -- named resource-policy grants for cross-account sending (`ses:SendEmail` with this identity as the source). Grant to specific principals, never `"AWS": "*"`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSesConfigurationSet** (optional) | `configurationSet` | `status.outputs.configuration_set_name` |

Without a configuration set reference, the identity is a leaf.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `identity_arn` | Amazon Resource Name of the identity | Authorization policies; IAM statements scoping sending |
| `email_identity` | The verified domain or address | Downstream automation composing DNS names |
| `identity_type` | "DOMAIN" or "EMAIL_ADDRESS" | Conditional automation |
| `verification_status` | "PENDING" until DNS/link verification, then "SUCCESS" | Deployment gating — a PENDING identity cannot send |
| `dkim_tokens` | Easy DKIM's three CNAME names | AwsRoute53DnsRecord CNAMEs that complete domain verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production domain** -- Easy DKIM, custom MAIL FROM, a configuration set attached. Start from the **Domain with Easy DKIM** preset.

**Domain with delivery rules** -- the same shape with the configuration set reference wired to a deployed set. Start from the **Domain with Configuration Set** preset.

## Works With

- [**AWS SES Configuration Set**](/infra-catalog/aws-ses-configuration-set) -- the default sending rules this identity inherits (references `configuration_set_name`)
- [**AWS Route 53 DNS Record**](/infra-catalog/aws-route53-dns-record) -- publishes the DKIM CNAMEs and the MAIL FROM MX/SPF records that complete verification
- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) -- the hosted zone those records live in
