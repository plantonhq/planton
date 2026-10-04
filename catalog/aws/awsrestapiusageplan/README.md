<p align="center">
  <img src="logo.svg" alt="AWS REST API Usage Plan" width="80"/>
</p>

# AWS REST API Usage Plan

Create and manage [API Gateway usage plans](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-api-usage-plans.html)
and the API keys they admit — quota, throttle, and consumer metering
for REST APIs.

A plan spans APIs and stages, which is why it is its own kind
rather than a field on [AwsRestApiGateway](../awsrestapigateway).
Routes opt in with `api_key_required`; requests then need a valid key
on the `X-Api-Key` header (or from the authorizer, per the API's
`api_key_source`).

API keys identify consumers for metering — they are not an
authentication mechanism. Pair them with IAM, Cognito, or Lambda
authorization on the routes.

## What Gets Created

- **A usage plan** covering one or more REST API stages, with optional
  quota (per day / week / month) and throttle ceilings.
- **API keys** created and attached to the plan. Key *values* are
  secrets and are not exported — read them from the AWS console when
  distributing to consumers.

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
