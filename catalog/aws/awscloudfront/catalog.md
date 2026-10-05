# AWS CloudFront

Deploys an Amazon CloudFront distribution — the global CDN front door that terminates TLS at the edge, caches responses close to viewers, and routes requests to one or more origins (S3 buckets, load balancers, API endpoints, or anything HTTP-addressable). The model mirrors CloudFront's own composition: origins declare WHERE content comes from, origin groups compose failover pairs, cache behaviors declare HOW requests are matched and cached, and the viewer certificate plus aliases put the distribution on your own domain. One regional quirk governs custom domains: the viewer certificate -- and any CLOUDFRONT-scope WAF web ACL -- must live in `us-east-1`, regardless of where the origins are.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudFront Distribution** -- the global distribution; identity is the AWS-generated ID (metadata.name drives the Name tag)
- **Origins** -- one per `origins[]` entry, each with its type arm: an S3 REST origin (optionally with a module-created Origin Access Control so the bucket stays fully private), a custom/HTTP origin (load balancers, API endpoints, S3 *website* endpoints), or a provisioned VPC origin reaching private resources
- **Origin Groups** -- primary/failover pairs created only for `originGroups[]` entries; behaviors targeting a group's ID get automatic per-request failover
- **Cache Behaviors** -- the required default behavior plus one path-matched behavior per `orderedCacheBehaviors[]` entry, each on the modern cache-policy generation or the legacy forwarded-values generation
- **Viewer Certificate wiring** -- the default *.cloudfront.net certificate, or your ACM/IAM certificate with SNI and a minimum TLS version when custom domains are configured
- **Continuous-Deployment Policy** -- created when `continuousDeployment` is set: the blue/green policy routing a weighted or header-selected traffic slice to a staging distribution, attached to this (primary) distribution
- **Custom Error Responses, Geo Restriction, Access Logging** -- created only when configured
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An origin** -- the content source CloudFront fetches from. For S3 use the REGIONAL bucket endpoint (the `bucket_regional_domain_name` output of AwsS3Bucket) with an Origin Access Control; for load balancers the LB DNS name; S3 *website* endpoints join as custom origins (they speak plain HTTP only).
- **An ACM certificate** (required for custom domains) -- MUST live in `us-east-1` regardless of where anything else lives. Provide the ARN directly or reference an AwsCertManagerCert Infra Component.
- **A WAF web ACL** (optional) -- CLOUDFRONT scope, which must also live in `us-east-1`. Reference an AwsWafWebAcl Infra Component or pass the ARN.
- **Bucket policy for OAC origins** -- allow the distribution's ARN on the `cloudfront.amazonaws.com` principal (an `AWS:SourceArn` condition) so the private bucket serves CloudFront alone.

## Deploy

### Console

Open the deployment store, find **AWS CloudFront**, and click **Deploy**. The creation wizard walks you through origins (with the private-bucket access control), the default and path-matched cache behaviors (with AWS managed-policy quick-picks), custom domains and the certificate, error pages, protections, and logging. Start from the **S3 Static Website with Origin Access Control** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudFront
metadata:
  name: app-cdn
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  origins:
    - originId: s3-site
      domainName: app-assets.s3.us-east-1.amazonaws.com
      s3Origin:
        createOriginAccessControl: true
  defaultCacheBehavior:
    targetOriginId: s3-site
    viewerProtocolPolicy: redirect-to-https
    compress: true
    # Managed-CachingOptimized — the AWS managed cache policy for static content.
    cachePolicyId: 658327ea-f89d-4fab-a63d-7e88639e58f6
  defaultRootObject: index.html
```

```shell
planton apply -f cloudfront.yaml
```

This creates a distribution serving a fully private S3 bucket through a module-created Origin Access Control, redirecting viewers to HTTPS, compressing at the edge, and caching with the AWS managed static-content policy. Deploys take 5-15 minutes — CloudFront pushes configuration to every edge location. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying with a custom domain, use ValueFromRef to wire the distribution to an ACM certificate (and optionally a WAF ACL) deployed in the same InfraPipeline:

```yaml
spec:
  aliases:
    - cdn.example.com
  viewerCertificate:
    acmCertificateArn:
      valueFrom:
        kind: AwsCertManagerCert
        name: cdn-cert
        fieldPath: status.outputs.cert_arn
  webAclArn:
    valueFrom:
      kind: AwsWafWebAcl
      name: cdn-waf
      fieldPath: status.outputs.web_acl_arn
```

The InfraPipeline resolves the dependency graph, deploys the certificate and WAF first, then provisions the distribution with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring a CloudFront distribution. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Origins and behaviors compose by ID** -- Each origin's `originId` is your stable handle; the default behavior and every path-matched behavior route to a declared origin (or origin group) by that ID, and validation proves every reference resolves at manifest time instead of at deploy time.

**The caching generation** -- The modern generation (`cachePolicyId` + `originRequestPolicyId`) is recommended: AWS ships managed policies covering most cases (`Managed-CachingOptimized` for static content, `Managed-CachingDisabled` plus `Managed-AllViewer` for APIs). The legacy generation (inline `forwardedValues` with per-behavior TTLs) is kept for existing configurations — the two are mutually exclusive per behavior.

**Path-matched routing** -- `orderedCacheBehaviors[]` are evaluated in order before the default; first match wins. The classic split routes `/api/*` to a load balancer with caching disabled while everything else serves from S3.

**Custom domains** -- Add `aliases` and set `viewerCertificate` with the ACM arm (the certificate must live in `us-east-1` and cover every alias). Without aliases the distribution serves on its generated `*.cloudfront.net` domain. Point Route53 alias records at the `domain_name` output — and with `isIpv6Enabled`, add the AAAA record too.

**Private origins via OAC** -- `s3Origin.createOriginAccessControl: true` provisions an Origin Access Control (SigV4 request signing) so the bucket stays fully private. An existing OAC of any type attaches at the origin level (`originAccessControlId`) — the shape for Lambda function URL, MediaPackage v2, and MediaStore origins too. A legacy Origin Access Identity is accepted on the S3 arm; attach-vs-create stays mutually exclusive.

**Blue/green rollouts** -- Deploy the candidate configuration as its own distribution with `staging: true`, then give the primary a `continuousDeployment` block: a weighted slice (up to 15%, with optional session stickiness) or an `aws-cf-cd-*` opt-in header routes production traffic to staging before you promote. An externally managed policy attaches via `continuousDeploymentPolicyId`.

**Mutual TLS, both directions** -- `viewerMtls` validates VIEWER client certificates against a CloudFront trust store (required/optional/passthrough); `customOrigin.mtlsClientCertificateArn` has CloudFront present an ACM client certificate to the ORIGIN so only CloudFront can reach the backend.

**Operational posture** -- `enabled: false` keeps a distribution deployed-but-dark for staging a configuration; `waitForDeployment` decides whether deploys block on edge propagation (5-15 minutes); `retainOnDelete` disables instead of deleting on destroy; `enableAdditionalMetrics` turns on the cache-hit-rate/origin-latency dashboard; `cacheTagHeaderName` enables tag-based invalidation (one invalidation purges every object carrying a label; write the header name lowercase — AWS stores it lowercased).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCertManagerCert** (optional) | `viewerCertificate.acmCertificateArn` | `status.outputs.cert_arn` |
| **AwsCertManagerCert** (optional) | `origins[].customOrigin.mtlsClientCertificateArn` | `status.outputs.cert_arn` |
| **AwsWafWebAcl** (optional) | `webAclArn` | `status.outputs.web_acl_arn` |
| **AwsCloudFront** (optional) | `continuousDeployment.stagingDistributionDnsNames[]` | `status.outputs.domain_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `distribution_id` | The distribution ID (e.g. E2ABCDEF123456) | Cache invalidations, monitoring subscriptions |
| `distribution_arn` | The distribution ARN | WAF associations, resource policies, OAC bucket policies |
| `domain_name` | The CloudFront domain (e.g. d123.cloudfront.net) | Route53 alias record target, application asset URL base |
| `hosted_zone_id` | The Route53 hosted zone ID for CloudFront aliases | Alias records without hardcoding CloudFront's global zone |
| `status` | The deployment status (`Deployed` / `InProgress`) | Gating downstream steps on edge propagation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**S3 static website** -- A private bucket behind a created Origin Access Control, redirect-to-https with edge compression and `Managed-CachingOptimized`, `index.html` as the root object, and the single-page-app error mapping (S3's 403-for-missing-object served as `/index.html` with a 200). Start from the **S3 Static Website with Origin Access Control** preset.

**Custom domain CDN** -- The static-site shape plus `aliases`, the ACM certificate reference, and dual-stack IPv6. Start from the **Custom Domain CDN with ACM Certificate** preset.

**Blue/green rollout** -- A primary distribution owning a continuous-deployment policy that canaries 10% of traffic (session-sticky) to a staging distribution. Start from the **Blue/Green Rollout with Continuous Deployment** preset.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides the regional bucket endpoint origins serve from
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) -- provides the us-east-1 certificate for custom domain HTTPS
- [**AWS Route 53 DNS Record**](/infra-catalog/aws-route53-dns-record) -- points custom domains at the distribution via alias records
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- provides Lambda@Edge versions for behavior-attached edge logic
