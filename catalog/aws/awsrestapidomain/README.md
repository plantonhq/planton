<p align="center">
  <img src="logo.svg" alt="AWS REST API Domain" width="80"/>
</p>

# AWS REST API Domain

Create and manage an [API Gateway custom domain](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html)
for REST APIs — your hostname, your certificate, base-path mappings that
fan paths out across APIs and stages.

A domain outlives any one API, which is why it is its own kind
rather than a field on [AwsRestApiGateway](../awsrestapigateway).

## What Gets Created

- **A custom domain name** bound to an ACM certificate (REGIONAL,
  EDGE, or PRIVATE endpoint types).
- **Base-path mappings** from this hostname's paths onto REST API
  stages.
- **Access associations** for PRIVATE domains (which VPC endpoints may
  call the hostname).

DNS is not modeled here: point an AwsRoute53DnsRecord alias at the
regional or CloudFront target (both are outputs). Rule-based
routing stays on [AwsHttpApiDomain](../awshttpapidomain); this
kind models the v1 `routing_mode` knob that arbitrates between
the two mechanisms.

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
