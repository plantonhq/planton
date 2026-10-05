# AWS ACM Certificate

Provisions an AWS Certificate Manager (ACM) certificate in any of ACM's three creation modes: Amazon-issued public certificates validated by DNS or email, imported bring-your-own certificate material, or private certificates issued by an ACM Private CA. For Amazon-issued certificates with a Route53 zone, the module creates the DNS validation records and optionally waits for issuance; without one, it exports the records for external DNS. Amazon-issued certificates renew automatically while their validation records stay in place; imported certificates never renew.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ACM Certificate** -- an Amazon-issued public certificate (the default), an imported certificate distributing your own material, or a private certificate issued by your ACM Private CA
- **Route53 DNS Validation Records** -- created only for DNS-validated certificates when `route53HostedZoneId` is set; one CNAME per domain, deduplicated when a wildcard and its apex share a record
- **Certificate Validation Wait** -- with managed Route53 records, the deployment waits for the certificate to reach ISSUED (disable via `waitForValidation: false`); without them, the certificate rests in PENDING_VALIDATION and exports the required records as the `domain_validation_records` output
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Route53 public hosted zone** (optional, DNS validation) -- when the zone is authoritative for every certificate domain, validation records are created and renewals stay fully automatic. Provide the zone ID directly or reference an AwsRoute53Zone Infra Component via ValueFromRef. When DNS lives elsewhere (Cloudflare, your registrar), skip the zone and create the exported records manually.
- **An ACM Private CA** (private mode only) -- the certificate authority ARN that signs private certificates. Clients must trust its root.
- **Certificate material** (imported mode only) -- the PEM certificate body and its unencrypted private key, provided as a secret reference.

## Deploy

### Console

Open the deployment store, find **AWS ACM Certificate**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single Domain DNS-Validated Certificate** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCertManagerCert
metadata:
  name: api-cert
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  primaryDomainName: api.example.com
  route53HostedZoneId:
    value: "Z123456ABCXYZ"
```

```shell
planton apply -f cert-manager-cert.yaml
```

This requests an Amazon-issued certificate for `api.example.com` with DNS validation, automatically creating the required CNAME record in the specified Route53 hosted zone and waiting for issuance. Certificates issued in `us-east-1` can be used with CloudFront distributions globally. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the certificate to a Route53 hosted zone deployed in the same InfraPipeline:

```yaml
spec:
  route53HostedZoneId:
    valueFrom:
      kind: AwsRoute53Zone
      name: example-zone
      fieldPath: status.outputs.zone_id
```

The InfraPipeline resolves the dependency graph, deploys the Route53 hosted zone first, then provisions the ACM certificate with the resolved zone ID. Downstream TLS consumers (listeners, CloudFront) reference this certificate's `cert_arn` output the same way.

## Key Configuration

These are the most important decisions when configuring an ACM certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Issuance mode** -- Setting `primaryDomainName` selects the Amazon-issued mode (or the private mode when `certificateAuthorityArn` is also set); setting `imported` brings your own material. Exactly one of the two drives the certificate. Imported certificates never auto-renew -- track `not_after` and re-import before expiry (re-imports keep the same ARN, so consumers are undisturbed).

**Primary domain** -- The main domain or a wildcard (e.g., `api.example.com` or `*.example.com`). Wildcards cover one label level only and never the bare apex -- add the apex as a SAN for full coverage.

**Region** -- ACM certificates are regional. Certificates for CloudFront must be created in `us-east-1`. For ALB, API Gateway, and other regional services, create the certificate in the same region as the service.

**Validation** -- DNS validation (the default) proves ownership with one CNAME per domain and renews automatically as long as the records stay in place. With `route53HostedZoneId` set, the records are managed for you; without it, they are exported as the `domain_validation_records` output for external DNS and the deployment does not wait. EMAIL validation requires manual approval at issuance and every renewal; HTTP validation serves a token from the domain for cases where DNS is untouchable -- prefer DNS.

**Key algorithm** -- Create-time immutable. `RSA_2048` (the default) works with every client; ECDSA curves (`EC_prime256v1` and up) give smaller, faster TLS handshakes when your clients support them.

**Certificate options** -- Certificate Transparency logging defaults to enabled (browsers increasingly require it); disable only to keep internal hostnames out of public logs. Key export lets non-AWS infrastructure serve the certificate, at an extra AWS charge.

**Private-CA early renewal** -- For ACM-PCA certificates, `earlyRenewalDuration` (e.g. `P90D`) starts ACM's managed renewal ahead of expiry while keeping the same ARN. Public certificates renew on ACM's own schedule; imported ones never renew.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsRoute53Zone** (optional) | `route53HostedZoneId` | `status.outputs.zone_id` |
| **AwsPrivateCa** (private mode only) | `certificateAuthorityArn` | `status.outputs.certificate_authority_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cert_arn` | ACM certificate ARN | ALB/NLB HTTPS listeners, CloudFront distributions, API Gateway custom domains, Cognito |
| `status` | Certificate status at deployment end (`ISSUED`, `PENDING_VALIDATION`, ...) | Deciding whether validation records still need creating |
| `domain_validation_records` | The DNS records proving ownership, one per domain | Creating validation CNAMEs in external DNS |
| `not_before` / `not_after` | The certificate's validity window (RFC3339) | Re-import scheduling for imported certificates |
| `certificate_type` | `AMAZON_ISSUED`, `IMPORTED`, or `PRIVATE` | Auditing and lifecycle tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single domain certificate** -- A certificate for one specific domain or subdomain (e.g., `api.example.com`) with managed DNS validation. The simplest and most common pattern for HTTPS on a single endpoint. Start from the **Single Domain DNS-Validated Certificate** preset.

**Wildcard certificate** -- A certificate covering all first-level subdomains (`*.example.com`) with the apex domain as a SAN. Suited for microservice architectures where each service has its own subdomain. Start from the **Wildcard Domain DNS-Validated Certificate** preset.

**External DNS** -- A certificate whose validation records live outside Route53. The deployment exports the records and finishes without waiting; create them in your DNS provider and ACM issues within minutes. Start from the **DNS-Validated Certificate with External DNS** preset.

## Works With

- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) -- provides the hosted zone for managed DNS validation records
- [**AWS Private Certificate Authority**](/infra-catalog/aws-private-ca) -- signs private certificates, wired via `certificateAuthorityArn`
- [**AWS ALB**](/infra-catalog/aws-alb) -- consumes the certificate on HTTPS listeners via `cert_arn`
- [**AWS CloudFront**](/infra-catalog/aws-cloud-front) -- consumes us-east-1 certificates for custom domain TLS
