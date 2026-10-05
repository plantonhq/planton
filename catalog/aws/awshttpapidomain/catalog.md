# AWS HTTP API Domain

Deploys an API Gateway v2 custom domain name — the production front door for HTTP APIs. It binds an owned domain (e.g. `api.example.com`) to an ACM certificate ([AWS ACM Certificate](/infra-catalog/aws-cert-manager-cert)) and routes requests onto one or more APIs — statically through path-key API mappings, dynamically through priority-ordered routing rules that match on base path or header, or both. The domain is deliberately its own resource: it outlives any one API, many APIs compose onto it under distinct path keys ("orders", "billing"), and the certificate rotates on its own lifecycle. DNS is composed, not embedded — the domain exports `target_domain_name` and `hosted_zone_id`, and an [AWS Route 53 DNS Record](/infra-catalog/aws-route53-dns-record) alias points your name at them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API Gateway v2 Custom Domain** -- the REGIONAL endpoint with the TLS_1_2 security policy (v2 domains support only these, so the modules set both)
- **Certificate Binding** -- TLS termination with the referenced ACM certificate (same region, covering the domain name), plus the ownership-verification certificate when Private-CA or imported-cert setups require it
- **Mutual TLS Configuration** -- the S3-hosted truststore requiring client certificates, when configured
- **API Mappings** -- one binding per row, each serving an API's stage at the root or under a path key
- **Routing Rules** -- one rule per row, matching requests on base path or header globs and invoking an API stage in priority order
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Certificate first** -- deploy the [AWS ACM Certificate](/infra-catalog/aws-cert-manager-cert) (in the same region, covering the domain) before the domain and reference its `cert_arn` output.
- **APIs to map** -- the [AWS HTTP API Gateway](/infra-catalog/aws-http-api-gateway) resources whose `api_id` outputs the mappings reference.

### AWS Account

- **A validated certificate** -- an unvalidated ACM certificate cannot bind; DNS-validated certificates in the same environment are the composed path.
- **For mTLS** -- a versioned S3 bucket (same region) holding the PEM truststore bundle, and `disableExecuteApiEndpoint` set on the mapped APIs so callers cannot bypass the client-certificate check.

## Deploy

### Console

Open the deployment store, find **AWS HTTP API Domain**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single API Domain** preset in the [Presets](#presets) tab to pre-populate the common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiDomain
metadata:
  name: api-example-com
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  domainName: api.example.com
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: api-example-com-cert
      fieldPath: status.outputs.cert_arn
  apiMappings:
    - apiId:
        valueFrom:
          kind: AwsHttpApiGateway
          name: orders-api
          fieldPath: status.outputs.api_id
      stage: $default
```

```shell
planton apply -f http-api-domain.yaml
```

This binds `api.example.com` to its certificate and serves the orders API at the domain root. An Infra Job tracks the provisioning in real time.

### InfraChart

When the domain deploys alongside its certificate and APIs in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  domainName: api.example.com
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: api-example-com-cert
      fieldPath: status.outputs.cert_arn
  apiMappings:
    - apiId:
        valueFrom:
          kind: AwsHttpApiGateway
          name: orders-api
          fieldPath: status.outputs.api_id
      stage: $default
```

The InfraPipeline resolves the dependency graph, deploys the certificate and APIs first, then binds the domain to them.

## Key Configuration

These are the most important decisions when configuring a custom domain. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The domain name is the one-way door** -- create-time immutable; changing it replaces the domain, and the DNS alias must re-point at the new target. A wildcard (`*.example.com`) matches all first-level subdomains but requires a certificate covering the wildcard.

**The certificate contract** -- issued (or imported) in the SAME region as the domain, covering the domain name exactly or by wildcard. Unlike CloudFront-backed REST domains, the certificate lives in the domain's own region — not us-east-1. Swapping certificates rotates in place.

**API mappings compose the domain** -- an empty path key serves an API at the root; distinct keys serve others under `https://<domain>/<key>/`. Keys are unique (only one root), single-segment (no slashes), and edit in place as APIs come and go.

**Routing rules route dynamically -- to REST APIs only** -- switch `routingMode` to `ROUTING_RULE_ONLY`, then declare rules that match on base paths or header globs and invoke a REST API stage (pass the REST API id literally: API Gateway rejects HTTP/WebSocket targets in routing rules, live-verified). Priorities (1-1,000,000, unique, lowest first) order evaluation; conditions on one rule are ANDed; `stripBasePath` forwards `/orders/list` to the API as `/list`. Rules and `apiMappings` are mutually exclusive: AWS rejects HTTP-API mappings on rule-mode domains (live-verified), so the spec rejects the combination at validate time -- along with rules without a rule-honoring mode, and a rule mode without rules. (`ROUTING_RULE_THEN_API_MAPPING` remains legal for domains whose fallback mappings are REST APIs managed outside this resource.)

**Mutual TLS is half a lock by itself** -- the S3 truststore requires client certificates at the domain edge, but each mapped API's default execute-api endpoint bypasses it. Disable that endpoint on the mapped APIs whenever mTLS is on.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCertManagerCert** | `certificateArn` | `status.outputs.cert_arn` |
| **AwsCertManagerCert** (optional) | `ownershipVerificationCertificateArn` | `status.outputs.cert_arn` |
| **AwsHttpApiGateway** | `apiMappings[].apiId` | `status.outputs.api_id` |

A routing rule's `apiId` takes a REST API id as a literal value — routing rules target REST APIs only, which live outside this catalog's HTTP API kind.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `domain_name_arn` | Amazon Resource Name of the domain | IAM policies and tag-based governance |
| `target_domain_name` | The API Gateway-managed regional domain to target from DNS | AwsRoute53DnsRecord alias target |
| `hosted_zone_id` | The Route 53 zone of the API Gateway endpoint | AwsRoute53DnsRecord alias target zone |

`domain_name` is also echoed back for automation that joins on the hostname.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single API front door** -- one API at the domain root; the everyday production shape. Start from the **Single API Domain** preset.

**Composed multi-API domain with mTLS** -- several teams' APIs under path keys, client certificates required at the edge. Start from the **Multi-API mTLS Domain** preset.

**Rule-routed domain** -- path-based service splits and header-based tenant pinning under one hostname, with the mappings as fallback. Start from the **Rule-Routed Domain** preset.

## Works With

- [**AWS HTTP API Gateway**](/infra-catalog/aws-http-api-gateway) -- the APIs mapped onto this domain (references `api_id`)
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) -- TLS termination (references `cert_arn`)
- [**AWS Route 53 DNS Record**](/infra-catalog/aws-route53-dns-record) -- the alias that routes the domain to API Gateway (consumes `target_domain_name` + `hosted_zone_id`)
- [**AWS HTTP API VPC Link**](/infra-catalog/aws-http-api-vpc-link) -- lets the mapped APIs reach private backends behind this front door
