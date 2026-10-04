# AwsRestApiUsagePlan

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsRestApiUsagePlanSpec defines the desired configuration for an AWS
API Gateway usage plan with its API keys.

Usage plans meter and throttle API consumers: each plan covers one
or more REST API stages, sets quota (requests per day/week/month)
and throttle ceilings, and admits callers through API keys attached
to the plan. Routes opt in with api_key_required on the
AwsRestApiGateway component; requests then need a valid key on the
X-Api-Key header (or from the authorizer, per the API's
api_key_source).

A plan spans APIs and stages - which is why it is its own component
rather than part of AwsRestApiGateway. The component bundles the
plan, its stage coverage, and the keys it admits (each key is
created AND attached to the plan).

API keys identify consumers for metering - they are NOT an
authentication mechanism; pair them with IAM, Cognito, or Lambda
authorization on the routes.

## Example

```yaml
# Canonical AwsRestApiUsagePlan example (hack/dev manifest and refgen
# Example source): a plan covering one REST API stage with a quota, a
# throttle, a per-method throttle, and one API key. Literal ids stand
# in for composed references so the offline `tofu plan` renders every
# arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiUsagePlan
metadata:
  name: orders-free-tier
  id: orders-free-tier
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Free tier -- 1000 requests/day
  apiStages:
    - restApiId:
        value: abcdef1234
      stageName:
        value: prod
      methodThrottles:
        - path: /search/GET
          burstLimit: 10
          rateLimit: 5
  quota:
    limit: 1000
    period: DAY
    offset: 0
  throttle:
    burstLimit: 100
    rateLimit: 50
  apiKeys:
    - name: acme-mobile
      description: acme-corp mobile app
      enabled: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.productCode` | `string` |  |  |  |
| `spec.apiStages` | `[]AwsRestApiUsagePlanApiStage` |  |  |  |
| `spec.apiStages[].restApiId` | `string \| valueFrom` | yes |  | AwsRestApiGateway (`status.outputs.rest_api_id`) |
| `spec.apiStages[].stageName` | `string \| valueFrom` | yes |  | AwsRestApiGateway (`status.outputs.stage_name`) |
| `spec.apiStages[].methodThrottles` | `[]AwsRestApiUsagePlanMethodThrottle` |  |  |  |
| `spec.apiStages[].methodThrottles[].path` | `string` | yes |  |  |
| `spec.apiStages[].methodThrottles[].burstLimit` | `int32` |  |  |  |
| `spec.apiStages[].methodThrottles[].rateLimit` | `double` |  |  |  |
| `spec.quota` | `AwsRestApiUsagePlanQuota` |  |  |  |
| `spec.quota.limit` | `int32` |  |  |  |
| `spec.quota.period` | `string` |  |  |  |
| `spec.quota.offset` | `int32` |  |  |  |
| `spec.throttle` | `AwsRestApiUsagePlanThrottle` |  |  |  |
| `spec.throttle.burstLimit` | `int32` |  |  |  |
| `spec.throttle.rateLimit` | `double` |  |  |  |
| `spec.apiKeys` | `[]AwsRestApiUsagePlanApiKey` |  |  |  |
| `spec.apiKeys[].name` | `string` | yes |  |  |
| `spec.apiKeys[].description` | `string` |  |  |  |
| `spec.apiKeys[].enabled` | `bool` |  |  |  |
| `spec.apiKeys[].customerId` | `string` |  |  |  |
| `spec.apiKeys[].value` | `string` (sensitive) | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the usage plan will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

What this plan meters (e.g. "Free tier: 1000 requests/day").

- rule: {"string":{"maxLen":"1024"}}

### spec.productCode

`string`

AWS Marketplace product code - links the plan to a Marketplace
SaaS listing so subscribers are metered through it.

### spec.apiStages

`[]AwsRestApiUsagePlanApiStage`

The REST API stages this plan covers. A key admitted by this plan
may call exactly these stages.

- rule: method_throttles entries must have unique path values

### spec.apiStages[].restApiId

`string | valueFrom` · required

The REST API whose stage this plan covers.

- references: AwsRestApiGateway (`status.outputs.rest_api_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRestApiGateway, name: <that resource's name>, fieldPath: status.outputs.rest_api_id}} -- a bare string does not parse

### spec.apiStages[].stageName

`string | valueFrom` · required

The stage name (must be deployed before the plan attaches -
referencing the AwsRestApiGateway stage_name output wires the
ordering).

- references: AwsRestApiGateway (`status.outputs.stage_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRestApiGateway, name: <that resource's name>, fieldPath: status.outputs.stage_name}} -- a bare string does not parse

### spec.apiStages[].methodThrottles

`[]AwsRestApiUsagePlanMethodThrottle`

Per-method throttles within this stage, addressed by
"{resource-path}/{METHOD}" (e.g. "/search/GET").

### spec.apiStages[].methodThrottles[].path

`string` · required

The method to throttle: "{resource-path}/{METHOD}" (e.g.
"/search/GET", "/orders/{id}/DELETE").

- rule: {"string":{"minLen":"1"}}

### spec.apiStages[].methodThrottles[].burstLimit

`int32`

Token-bucket burst for this method.

- rule: {"int32":{"gte":0}}

### spec.apiStages[].methodThrottles[].rateLimit

`double`

Steady-state request rate (requests/second) for this method.

- rule: {"double":{"gte":0}}

### spec.quota

`AwsRestApiUsagePlanQuota`

Total request quota over a period. Omitted = unmetered (throttle
only).

- rule: quota offset must be 0 for DAY, 0-6 for WEEK, and 0-27 for MONTH

### spec.quota.limit

`int32`

Requests allowed per period.

- rule: {"int32":{"gte":1}}

### spec.quota.period

`string`

The metering period.

- rule: {"string":{"in":["DAY","WEEK","MONTH"]}}

### spec.quota.offset

`int32`

Requests already counted against the FIRST period (a mid-period
migration knob). DAY takes 0; WEEK takes 0-6; MONTH takes 0-27.

- rule: {"int32":{"gte":0}}

### spec.throttle

`AwsRestApiUsagePlanThrottle`

Steady-state and burst throttle ceilings across the plan. Omitted =
the account-level limits apply.

- rule: set at least one of burst_limit or rate_limit (omit throttle entirely for account defaults)

### spec.throttle.burstLimit

`int32`

Token-bucket burst across the plan.

- rule: {"int32":{"gte":0}}

### spec.throttle.rateLimit

`double`

Steady-state request rate (requests/second) across the plan.

- rule: {"double":{"gte":0}}

### spec.apiKeys

`[]AwsRestApiUsagePlanApiKey`

API keys created and attached to this plan.

### spec.apiKeys[].name

`string` · required

Key name in AWS (unique account-wide). The for_each key on both
engines and the key in the `api_key_ids` output map.

- rule: {"string":{"minLen":"1"}}

### spec.apiKeys[].description

`string`

Who this key identifies (e.g. "acme-corp mobile app").

- rule: {"string":{"maxLen":"1024"}}

### spec.apiKeys[].enabled

`bool` · optional (explicit presence)

Whether the key is accepted. Omitted = enabled; set false to cut a
consumer off without deleting (and re-generating) the key.

### spec.apiKeys[].customerId

`string`

AWS Marketplace customer identifier for Marketplace-metered
consumers.

### spec.apiKeys[].value

`string` · required · sensitive

The key value (20-128 characters). Omitted = AWS generates one
(recommended). Set it only when a consumer's key must survive
re-creation - and treat it as the secret it is.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"20","maxLen":"128"}}

## Validation Rules

- `api_key_names_unique`: api_keys entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRestApiUsagePlan, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.usage_plan_id` | `string` | The usage plan ID. |
| `status.outputs.usage_plan_arn` | `string` | The usage plan ARN. |
| `status.outputs.api_key_ids` | `map<string, string>` | API key IDs keyed by each `api_keys` entry's name. |
| `status.outputs.api_key_arns` | `map<string, string>` | API key ARNs keyed by each `api_keys` entry's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.apiStages[].restApiId` | AwsRestApiGateway | `status.outputs.rest_api_id` |
| `spec.apiStages[].stageName` | AwsRestApiGateway | `status.outputs.stage_name` |

## See Also

- [Overview](../README.md)
