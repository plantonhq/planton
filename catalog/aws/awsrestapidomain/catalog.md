# AWS REST API Domain

Deploys an API Gateway custom domain for REST APIs — callers hit `https://api.example.com/orders` instead of the execute-api endpoint, TLS terminates on your certificate, and base-path mappings fan the hostname's paths out across APIs and stages. A domain outlives any one API and maps many, which is why it is its own component rather than a field on the REST API. The bundle covers the domain, its base-path mappings, and — for PRIVATE domains — the VPC-endpoint access associations; DNS stays outside, composed through the alias-target outputs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Custom Domain** — the hostname bound to its TLS certificate (ACM, or directly uploaded material on the legacy path), with the chosen endpoint type (REGIONAL, EDGE, or PRIVATE), security policy, optional mutual TLS truststore, and — on PRIVATE domains — the resource policy
- **Base-Path Mapping** — one per `basePathMappings` entry, routing a path segment under the domain to a REST API and stage; the empty base path maps the domain root
- **Access Association** — one per `accessAssociations` entry (PRIVATE domains only), granting an interface VPC endpoint access to the private hostname

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with API Gateway domain permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An ACM certificate covering the hostname, wired as `certificateArn` — in this region for REGIONAL and PRIVATE domains, in `us-east-1` for EDGE domains (the CloudFront region)
- The REST APIs and stages the mappings will point at
- Interface VPC endpoints for `com.amazonaws.{region}.execute-api` (only for PRIVATE domains)
- An S3 truststore bundle of CA certificates (only for `mutualTls`)

## Deploy

### Console

Open the deployment store, find **AWS REST API Domain**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the hostname and certificate, and the base-path mappings. Start from the **Regional Mapped Domain** preset in the [Presets](#presets) tab for the default shape, or the **Edge Custom Domain** preset when the domain should ride CloudFront's global edge.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiDomain
metadata:
  name: api-domain
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  domainName: api.acme-corp.com
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: api-acme-cert
      fieldPath: status.outputs.cert_arn
  endpointConfiguration:
    type: REGIONAL
  basePathMappings:
    - basePath: orders
      restApiId:
        valueFrom:
          kind: AwsRestApiGateway
          name: orders-api
          fieldPath: status.outputs.rest_api_id
      stageName:
        value: prod
```

```shell
planton apply -f rest-api-domain.yaml
```

This creates a REGIONAL custom domain on your ACM certificate with `https://api.acme-corp.com/orders` routed to the orders API's `prod` stage. An Infra Job tracks the provisioning in real time.

### InfraChart

When the domain deploys alongside the API it fronts in one chart, wire the API reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  domainName: api.acme-corp.com
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: api-acme-cert
      fieldPath: status.outputs.cert_arn
  basePathMappings:
    - basePath: orders
      restApiId:
        valueFrom:
          kind: AwsRestApiGateway
          name: orders-api
          fieldPath: status.outputs.rest_api_id
      stageName:
        valueFrom:
          kind: AwsRestApiGateway
          name: orders-api
          fieldPath: status.outputs.stage_name
```

The InfraPipeline resolves the dependency graph, deploys the API first, then maps the domain onto its stage.

## Key Configuration

These are the most important decisions when configuring a custom domain. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Never rotate the hostname in place** — `domainName` is immutable in AWS; changing it replaces the domain and every mapping on it. Stand up the new name, move DNS, then destroy the old one.

**REGIONAL is the default for new work** — EDGE still exists for global CloudFront distribution, but it requires the certificate in `us-east-1` and a longer create. PRIVATE domains are reachable only through VPC endpoints and need both `accessAssociations` and a resource `policy` — without the associations, the hostname exists and nobody inside the VPC can call it.

**Keep the root mapping** — an empty `basePath` (the `(none)` output key, AWS's own empty-path sentinel) is how `https://api.example.com/` itself reaches a stage. Forgetting it 404s the bare hostname while every mapped path works, which is a confusing way to find out.

**Pin the stage on every mapping** — with `stageName` omitted, AWS selects the stage from the request path's next segment (stage-in-path behavior), which exposes every stage of the API under the domain. Setting it is the usual and safer choice.

**Version the mTLS truststore** — `mutualTls.truststoreUri` points at an S3 bundle of CA certificates clients must chain to; an unpinned truststore changes under the domain, and a truststore mistake locks every caller out at the TLS layer. Always set `truststoreVersion`.

**endpointAccessMode pairs with modern security policies** — AWS supports STRICT host-header enforcement only with the `SecurityPolicy_*` family; the manifest validation enforces the pairing so a legacy `TLS_1_2` domain never carries a dead access-mode setting.

**Prefer ACM over uploaded certificates** — `uploadedCertificate` is the legacy path for material that cannot be imported into ACM; AWS's own documentation is ambiguous about which endpoint types accept it. ACM certificates renew themselves; uploaded ones become your rotation problem.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCertManagerCert** | `certificateArn`, `ownershipVerificationCertificateArn` | `status.outputs.cert_arn` |
| **AwsRestApiGateway** | `basePathMappings[].restApiId` | `status.outputs.rest_api_id` |
| **AwsRestApiGateway** | `basePathMappings[].stageName` | `status.outputs.stage_name` |
| **AwsVpcEndpoint** | `accessAssociations[].vpcEndpointId` | `status.outputs.vpc_endpoint_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `regional_domain_name` / `regional_zone_id` | The regional target hostname and its hosted zone ID | Route 53 alias records for REGIONAL and PRIVATE domains |
| `cloudfront_domain_name` / `cloudfront_zone_id` | The CloudFront target hostname and its fixed zone ID | Route 53 alias records for EDGE domains |
| `domain_name_arn` | The domain's ARN | IAM policies and access-association wiring |

`domain_name`, `domain_name_id`, `base_path_mapping_ids`, and `access_association_arns` are also exported; they echo configuration and identify child resources for import and operational tooling rather than feeding composition.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One hostname, many APIs** — a REGIONAL domain with base-path mappings fanning `/orders`, `/billing`, and the root out to separate REST APIs' stages. Teams ship APIs independently while callers see one host. Start from the **Regional Mapped Domain** preset.

**Global edge domain** — an EDGE domain riding CloudFront with the certificate in `us-east-1`, for geographically-spread callers who should hit the nearest edge. Start from the **Edge Custom Domain** preset.

**Private internal API host** — a PRIVATE domain with dualstack addressing, access associations granting your VPC endpoints entry, and a resource policy naming the allowed principals. Pairs with a PRIVATE REST API so neither the domain nor the API is reachable from the internet.

## Works With

- [**AWS REST API Gateway**](/infra-catalog/aws-rest-api-gateway) — the APIs and stages the base-path mappings route to
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) — the ACM certificate TLS terminates on
- [**AWS Route 53 DNS Record**](/infra-catalog/aws-route53-dns-record) — the alias record pointing the hostname at the regional or CloudFront target
- [**AWS VPC Endpoint**](/infra-catalog/aws-vpc-endpoint) — the interface endpoints PRIVATE domains admit through access associations
