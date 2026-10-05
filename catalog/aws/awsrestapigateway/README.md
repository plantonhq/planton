<p align="center">
  <img src="logo.svg" alt="AWS REST API Gateway" width="80"/>
</p>

# AWS REST API Gateway

Create and manage [Amazon API Gateway REST APIs](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-rest-api.html)
(API Gateway v1) — the full-featured protocol surface: mapping templates,
JSON Schema models, request validation, per-method caching and throttling,
WAF-capable stages, and EDGE / REGIONAL / PRIVATE endpoints.

HTTP APIs (the leaner, cheaper alternative) are the
[AwsHttpApiGateway](../awshttpapigateway) kind.

## What Gets Created

- **A REST API** whose definition is exactly one of typed `routes`
  (the modules derive the resource tree from the paths) or an `openapi`
  document AWS imports.
- **An explicit deployment** whose trigger hashes the full definition,
  so every spec change redeploys automatically, plus **one stage**
  (Planton resources are already environment-scoped).
- **API-scoped satellites**: named authorizers, models, request
  validators, gateway responses, a resource policy, documentation, and
  an optional generated client certificate.

Custom domains are [AwsRestApiDomain](../awsrestapidomain). VPC links
are [AwsRestApiVpcLink](../awsrestapivpclink). Usage plans and API keys
are [AwsRestApiUsagePlan](../awsrestapiusageplan).

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
