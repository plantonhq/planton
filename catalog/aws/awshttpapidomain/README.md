# AwsHttpApiDomain

Deploy and manage an API Gateway v2 custom domain name using Planton -- the production front door that binds an owned domain (e.g. `api.example.com`) to an ACM certificate and routes requests onto one or more HTTP APIs.

## Overview

A custom domain replaces the default `https://{api-id}.execute-api.{region}.amazonaws.com` endpoint with your own DNS name and TLS certificate. The domain routes requests two ways, selected by `routingMode`:

- **API mappings** (the default): static path-key publishing -- one API at the root, another under `/orders`, a third under `/billing`.
- **Routing rules**: dynamic routing that matches requests on base path or header values and invokes a REST API stage by priority order -- a service split, tenant pinning, or a gradual migration under one hostname. Two AWS restrictions shape this surface (both live-verified): routing rules invoke ONLY REST-protocol APIs (pass the REST API id literally -- HTTP/WebSocket targets are rejected), and HTTP-API mappings cannot coexist with rule modes, so a domain either maps HTTP APIs statically (`apiMappings`) or routes to REST APIs by rules.

The domain is deliberately its own resource rather than a field on the API:

- **A domain outlives any one API** -- APIs come and go; the domain, its certificate binding, and its DNS records persist.
- **Many APIs, one domain** -- mappings compose multiple APIs under distinct path keys.
- **Independent certificate lifecycle** -- the ACM certificate rotates without touching any API.

DNS is composed, not embedded: the domain exports `target_domain_name` and `hosted_zone_id`; create an alias record with `AwsRoute53DnsRecord` pointing your domain at those outputs.

## When to Use

- Give a production API a stable, branded URL with your own TLS certificate.
- Consolidate several microservice APIs under one hostname with path-key namespacing.
- Route by header or path with **routing rules** -- pin beta tenants to a canary API, split `/orders` traffic to its own service, or migrate an API gradually behind one hostname.
- Enforce **mutual TLS** for B2B / machine-to-machine APIs (clients must present certificates chaining to your CA truststore).

## Prerequisites

- An ACM certificate **in the same region** covering the domain name (exact or wildcard match) -- compose with `AwsCertManagerCert`.
- An HTTP API to map -- compose with `AwsHttpApiGateway` (its `api_id` and `stage_name` outputs are the mapping's inputs).
- (Optional) A Route 53 zone to publish the alias record -- compose with `AwsRoute53Zone` + `AwsRoute53DnsRecord`.

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiDomain
metadata:
  name: api-example-com
  org: my-org
  env: prod
  id: api-example-com-prod
spec:
  region: us-east-1
  domainName: api.example.com
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: api-example-cert
      fieldPath: status.outputs.cert_arn
  apiMappings:
    - apiId:
        valueFrom:
          kind: AwsHttpApiGateway
          name: orders-api
          fieldPath: status.outputs.api_id
      stage: $default
```

## Multi-API Example

```yaml
spec:
  region: us-east-1
  domainName: api.example.com
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: api-example-cert
      fieldPath: status.outputs.cert_arn
  apiMappings:
    # https://api.example.com/ -> the public API
    - apiId:
        valueFrom:
          kind: AwsHttpApiGateway
          name: public-api
          fieldPath: status.outputs.api_id
      stage: $default
    # https://api.example.com/orders/... -> the orders service
    - apiId:
        valueFrom:
          kind: AwsHttpApiGateway
          name: orders-api
          fieldPath: status.outputs.api_id
      stage: $default
      apiMappingKey: orders
```

## Rule-Routed Example

Rules first (ascending priority), mappings as the fallback. Conditions on
one rule are ANDed; separate rules are alternatives.

```yaml
spec:
  region: us-east-1
  domainName: api.example.com
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: api-example-cert
      fieldPath: status.outputs.cert_arn
  # A rule-routed domain carries no apiMappings (AWS rejects HTTP-API
  # mappings on rule-mode domains), and rules invoke ONLY REST-protocol
  # APIs -- pass each REST API's id literally. The rule set must cover
  # every path family the domain serves (unmatched requests receive a 404).
  routingMode: ROUTING_RULE_ONLY
  routingRules:
    # Beta tenants reach the canary REST API regardless of path.
    - priority: 10
      conditions:
        - header:
            name: x-tenant-id
            valueGlob: beta-*
      apiId:
        value: f6a7b8c9d0
      stage: prod
    # /orders/... reaches the orders REST API, base path stripped.
    - priority: 20
      conditions:
        - basePaths:
            - orders
      apiId:
        value: a1b2c3d4e5
      stage: prod
      stripBasePath: true
```

## Spec Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | AWS region. Must match the certificate's and the mapped APIs' region. |
| `domainName` | string | Yes | The fully qualified custom domain (lowercase; wildcard `*.example.com` allowed). Immutable -- changing it replaces the domain. |
| `certificateArn` | StringValueOrRef | Yes | ACM certificate covering the domain, in the same region. |
| `ownershipVerificationCertificateArn` | StringValueOrRef | No | AWS-issued public certificate proving domain ownership -- required by AWS when `certificateArn` is Private-CA-issued, or when mTLS uses an ACM-imported certificate. |
| `ipAddressType` | string | No | `ipv4` or `dualstack` (AWS default). |
| `mutualTls.truststoreUri` | string | Conditional | S3 URI of the PEM CA bundle (e.g. `s3://bucket/truststore.pem`). Required when the block is present. |
| `mutualTls.truststoreVersion` | string | No | S3 object version pin -- makes CA rotation an explicit change. |
| `apiMappings[].apiId` | StringValueOrRef | Yes | The API to map (references `AwsHttpApiGateway.status.outputs.api_id`). Mutually exclusive with rule-mode routing. |
| `apiMappings[].stage` | string | Yes | The API stage to serve (typically `$default`). |
| `apiMappings[].apiMappingKey` | string | No | Path key (no slashes). Empty serves the API at the domain root. Keys must be unique across mappings. |
| `routingMode` | string | No | `API_MAPPING_ONLY` (AWS default), `ROUTING_RULE_ONLY`, or `ROUTING_RULE_THEN_API_MAPPING`. Must agree with `routingRules` (rules require a rule-honoring mode, and vice versa). |
| `routingRules[].priority` | int | Yes | Evaluation order, 1-1,000,000, lower first. Unique across the domain's rules. |
| `routingRules[].conditions[]` | list | Yes | Each condition sets exactly one of `basePaths` (candidate first path segments, case-sensitive) or `header` (`name` ≤40 chars, `valueGlob` ≤128 chars, `prefix-*`/`*-suffix`/`*infix*` globs). Conditions on a rule are ANDed. |
| `routingRules[].apiId` | StringValueOrRef | Yes | The REST API matching requests route to (literal id -- routing rules support only REST-protocol APIs). |
| `routingRules[].stage` | string | Yes | The target REST API stage (e.g. `prod`). |
| `routingRules[].stripBasePath` | bool | No | Strip the matched base path before forwarding (`/orders/list` reaches the API as `/list`). |

## Outputs

| Output | Description |
|--------|-------------|
| `domain_name` | The custom domain name (the domain's join key). |
| `domain_name_arn` | ARN of the domain name resource. |
| `target_domain_name` | The API Gateway regional domain to alias/CNAME to (e.g. `d-abc123.execute-api.us-east-1.amazonaws.com`). |
| `hosted_zone_id` | Route 53 hosted zone ID of the regional endpoint -- the alias target zone. |

## Publishing DNS

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53DnsRecord
metadata:
  name: api-alias
spec:
  # ... zone reference ...
  recordType: A
  alias:
    dnsName:
      valueFrom:
        kind: AwsHttpApiDomain
        name: api-example-com
        fieldPath: status.outputs.target_domain_name
    hostedZoneId:
      valueFrom:
        kind: AwsHttpApiDomain
        name: api-example-com
        fieldPath: status.outputs.hosted_zone_id
```

## Security Notes

- When mTLS is enabled, also set `disableExecuteApiEndpoint: true` on every mapped API -- otherwise callers can bypass mTLS via the default execute-api endpoint.
- API Gateway v2 domains accept only the REGIONAL endpoint type and TLS 1.2 security policy; the modules set both (they are not spec fields because AWS accepts no other value).

## Deliberately Omitted

- **Edge-optimized endpoint type**: a REST API (v1) feature; v2 domains are REGIONAL only.
- **Per-kind tags**: identity tags derive from metadata.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
