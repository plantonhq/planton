# AwsEventBridgeApiDestination

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsEventBridgeApiDestinationSpec defines the authenticated HTTP(S)
endpoint EventBridge rules, pipes, and schedules invoke, as two
independently deployable arms:

The CONNECTION arm is the auth trust anchor: api-key, basic, or
OAuth client credentials. AWS stores the credential values in a
Secrets Manager secret it creates and owns (the
connection_secret_arn output); DescribeConnection NEVER returns
them. One connection serves many destinations - a shared connection
lives in ONE owning instance and other instances reference it by
ARN.

The DESTINATION arm is the invocable endpoint: HTTPS URL + method +
rate limit, bound to a connection - the one this instance owns, or
one that exists elsewhere by ARN.

## Example

```yaml
# Canonical AwsEventBridgeApiDestination example (hack/dev manifest and
# refgen Example source): an api-key connection fronting a partner
# webhook, with a static header on every invocation.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeApiDestination
metadata:
  name: partner-webhook
  id: partner-webhook
  org: test-org
  env: dev
spec:
  region: us-west-2
  connection:
    name: partner-api
    description: Partner ingestion API (api-key auth)
    apiKey:
      key: x-api-key
      value: replace-with-real-key
    invocationHttpParameters:
      header:
        - key: x-planton-source
          value: eventbridge
  destination:
    name: partner-webhook
    description: Partner order-events webhook
    invocationEndpoint: https://api.example.com/events
    httpMethod: POST
    invocationRateLimitPerSecond: 10
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.connection` | `AwsEventBridgeConnection` |  |  |  |
| `spec.connection.name` | `string` | yes |  |  |
| `spec.connection.description` | `string` |  |  |  |
| `spec.connection.apiKey` | `AwsEventBridgeConnectionApiKeyAuth` |  |  |  |
| `spec.connection.apiKey.key` | `string` | yes |  |  |
| `spec.connection.apiKey.value` | `string` (sensitive) | yes |  |  |
| `spec.connection.basic` | `AwsEventBridgeConnectionBasicAuth` |  |  |  |
| `spec.connection.basic.username` | `string` | yes |  |  |
| `spec.connection.basic.password` | `string` (sensitive) | yes |  |  |
| `spec.connection.oauth` | `AwsEventBridgeConnectionOAuth` |  |  |  |
| `spec.connection.oauth.authorizationEndpoint` | `string` | yes |  |  |
| `spec.connection.oauth.httpMethod` | `string` |  |  |  |
| `spec.connection.oauth.clientId` | `string` | yes |  |  |
| `spec.connection.oauth.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters` | `AwsEventBridgeConnectionHttpParameters` | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters.body` | `[]AwsEventBridgeConnectionHttpParameter` |  |  |  |
| `spec.connection.oauth.oauthHttpParameters.body[].key` | `string` | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters.body[].value` | `string` (sensitive) | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters.body[].isValueSecret` | `bool` |  |  |  |
| `spec.connection.oauth.oauthHttpParameters.header` | `[]AwsEventBridgeConnectionHttpParameter` |  |  |  |
| `spec.connection.oauth.oauthHttpParameters.header[].key` | `string` | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters.header[].value` | `string` (sensitive) | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters.header[].isValueSecret` | `bool` |  |  |  |
| `spec.connection.oauth.oauthHttpParameters.queryString` | `[]AwsEventBridgeConnectionHttpParameter` |  |  |  |
| `spec.connection.oauth.oauthHttpParameters.queryString[].key` | `string` | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters.queryString[].value` | `string` (sensitive) | yes |  |  |
| `spec.connection.oauth.oauthHttpParameters.queryString[].isValueSecret` | `bool` |  |  |  |
| `spec.connection.invocationHttpParameters` | `AwsEventBridgeConnectionHttpParameters` |  |  |  |
| `spec.connection.invocationHttpParameters.body` | `[]AwsEventBridgeConnectionHttpParameter` |  |  |  |
| `spec.connection.invocationHttpParameters.body[].key` | `string` | yes |  |  |
| `spec.connection.invocationHttpParameters.body[].value` | `string` (sensitive) | yes |  |  |
| `spec.connection.invocationHttpParameters.body[].isValueSecret` | `bool` |  |  |  |
| `spec.connection.invocationHttpParameters.header` | `[]AwsEventBridgeConnectionHttpParameter` |  |  |  |
| `spec.connection.invocationHttpParameters.header[].key` | `string` | yes |  |  |
| `spec.connection.invocationHttpParameters.header[].value` | `string` (sensitive) | yes |  |  |
| `spec.connection.invocationHttpParameters.header[].isValueSecret` | `bool` |  |  |  |
| `spec.connection.invocationHttpParameters.queryString` | `[]AwsEventBridgeConnectionHttpParameter` |  |  |  |
| `spec.connection.invocationHttpParameters.queryString[].key` | `string` | yes |  |  |
| `spec.connection.invocationHttpParameters.queryString[].value` | `string` (sensitive) | yes |  |  |
| `spec.connection.invocationHttpParameters.queryString[].isValueSecret` | `bool` |  |  |  |
| `spec.connection.privateInvocationEndpoint` | `AwsEventBridgePrivateEndpoint` |  |  |  |
| `spec.connection.privateInvocationEndpoint.resourceConfigurationArn` | `string` | yes |  |  |
| `spec.connection.privateAuthorizationEndpoint` | `AwsEventBridgePrivateEndpoint` |  |  |  |
| `spec.connection.privateAuthorizationEndpoint.resourceConfigurationArn` | `string` | yes |  |  |
| `spec.connection.kmsKeyIdentifier` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.destination` | `AwsEventBridgeDestination` |  |  |  |
| `spec.destination.name` | `string` | yes |  |  |
| `spec.destination.description` | `string` |  |  |  |
| `spec.destination.connectionArn` | `string \| valueFrom` |  |  | AwsEventBridgeApiDestination (`status.outputs.connection_arn`) |
| `spec.destination.invocationEndpoint` | `string` | yes |  |  |
| `spec.destination.httpMethod` | `string` |  |  |  |
| `spec.destination.invocationRateLimitPerSecond` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the connection and destination live in. Example:
"us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.connection

`AwsEventBridgeConnection`

The owned connection - the auth trust anchor.

- rule: configure exactly one of api_key, basic, and oauth

### spec.connection.name

`string` · required

The connection's name in AWS (its identity - renaming replaces
it). Up to 64 characters: letters, digits, dot, dash, underscore.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_.-]+$"}}

### spec.connection.description

`string`

What this connection authenticates against (the SaaS or internal
API it fronts).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512"}}

### spec.connection.apiKey

`AwsEventBridgeConnectionApiKeyAuth`

API-key auth: the key is sent as an HTTP header
("{key}: {value}") on every invocation.

### spec.connection.apiKey.key

`string` · required

The header name the key is sent under. Example: "x-api-key".

- rule: {"string":{"minLen":"1","maxLen":"512"}}

### spec.connection.apiKey.value

`string` · required · sensitive

The API key itself. Stored by AWS in the connection's Secrets
Manager secret; never returned by any AWS read API.

- rule: {"string":{"minLen":"1","maxLen":"512"}}

### spec.connection.basic

`AwsEventBridgeConnectionBasicAuth`

HTTP basic auth.

### spec.connection.basic.username

`string` · required

The username.

- rule: {"string":{"minLen":"1","maxLen":"512"}}

### spec.connection.basic.password

`string` · required · sensitive

The password. Stored by AWS in the connection's Secrets Manager
secret; never returned by any AWS read API.

- rule: {"string":{"minLen":"1","maxLen":"512"}}

### spec.connection.oauth

`AwsEventBridgeConnectionOAuth`

OAuth client-credentials auth: EventBridge fetches a token from
the authorization endpoint before invoking.

### spec.connection.oauth.authorizationEndpoint

`string` · required

The token endpoint EventBridge calls to obtain the access token.
Example: "https://auth.example.com/oauth2/token".

- rule: {"string":{"minLen":"1","maxLen":"2048"}}

### spec.connection.oauth.httpMethod

`string`

The HTTP method for the token request.

- rule: {"string":{"in":["GET","POST","PUT"]}}

### spec.connection.oauth.clientId

`string` · required

The OAuth client id.

- rule: {"string":{"minLen":"1","maxLen":"512"}}

### spec.connection.oauth.clientSecret

`string` · required · sensitive

The OAuth client secret. Stored by AWS in the connection's Secrets
Manager secret; never returned by any AWS read API.

- rule: {"string":{"minLen":"1","maxLen":"512"}}

### spec.connection.oauth.oauthHttpParameters

`AwsEventBridgeConnectionHttpParameters` · required

Parameters added to the TOKEN request (grant_type in the body,
scope, audience...). AWS requires at least one entry here - most
OAuth servers need "grant_type=client_credentials" in the body.

- rule: {"required":true}
- rule: configure at least one of body, header, and query_string parameters

### spec.connection.oauth.oauthHttpParameters.body

`[]AwsEventBridgeConnectionHttpParameter`

Body fields (form/JSON fields on the request body).

### spec.connection.oauth.oauthHttpParameters.body[].key

`string` · required

The parameter name.

- rule: {"string":{"minLen":"1"}}

### spec.connection.oauth.oauthHttpParameters.body[].value

`string` · required · sensitive

The parameter value. The provider treats every value as sensitive
(AWS masks them all on read), so this field is secret-typed
regardless of is_value_secret.

- rule: {"string":{"minLen":"1"}}

### spec.connection.oauth.oauthHttpParameters.body[].isValueSecret

`bool`

Store this value in the connection's Secrets Manager secret (true)
rather than the connection configuration (false). Set true for
anything credential-like.

### spec.connection.oauth.oauthHttpParameters.header

`[]AwsEventBridgeConnectionHttpParameter`

HTTP headers.

### spec.connection.oauth.oauthHttpParameters.header[].key

`string` · required

The parameter name.

- rule: {"string":{"minLen":"1"}}

### spec.connection.oauth.oauthHttpParameters.header[].value

`string` · required · sensitive

The parameter value. The provider treats every value as sensitive
(AWS masks them all on read), so this field is secret-typed
regardless of is_value_secret.

- rule: {"string":{"minLen":"1"}}

### spec.connection.oauth.oauthHttpParameters.header[].isValueSecret

`bool`

Store this value in the connection's Secrets Manager secret (true)
rather than the connection configuration (false). Set true for
anything credential-like.

### spec.connection.oauth.oauthHttpParameters.queryString

`[]AwsEventBridgeConnectionHttpParameter`

Query-string parameters.

### spec.connection.oauth.oauthHttpParameters.queryString[].key

`string` · required

The parameter name.

- rule: {"string":{"minLen":"1"}}

### spec.connection.oauth.oauthHttpParameters.queryString[].value

`string` · required · sensitive

The parameter value. The provider treats every value as sensitive
(AWS masks them all on read), so this field is secret-typed
regardless of is_value_secret.

- rule: {"string":{"minLen":"1"}}

### spec.connection.oauth.oauthHttpParameters.queryString[].isValueSecret

`bool`

Store this value in the connection's Secrets Manager secret (true)
rather than the connection configuration (false). Set true for
anything credential-like.

### spec.connection.invocationHttpParameters

`AwsEventBridgeConnectionHttpParameters`

Extra static parameters (headers, query string, body fields) sent
with EVERY invocation through this connection - API version pins,
tenant ids, and the like.

- rule: configure at least one of body, header, and query_string parameters

### spec.connection.invocationHttpParameters.body

`[]AwsEventBridgeConnectionHttpParameter`

Body fields (form/JSON fields on the request body).

### spec.connection.invocationHttpParameters.body[].key

`string` · required

The parameter name.

- rule: {"string":{"minLen":"1"}}

### spec.connection.invocationHttpParameters.body[].value

`string` · required · sensitive

The parameter value. The provider treats every value as sensitive
(AWS masks them all on read), so this field is secret-typed
regardless of is_value_secret.

- rule: {"string":{"minLen":"1"}}

### spec.connection.invocationHttpParameters.body[].isValueSecret

`bool`

Store this value in the connection's Secrets Manager secret (true)
rather than the connection configuration (false). Set true for
anything credential-like.

### spec.connection.invocationHttpParameters.header

`[]AwsEventBridgeConnectionHttpParameter`

HTTP headers.

### spec.connection.invocationHttpParameters.header[].key

`string` · required

The parameter name.

- rule: {"string":{"minLen":"1"}}

### spec.connection.invocationHttpParameters.header[].value

`string` · required · sensitive

The parameter value. The provider treats every value as sensitive
(AWS masks them all on read), so this field is secret-typed
regardless of is_value_secret.

- rule: {"string":{"minLen":"1"}}

### spec.connection.invocationHttpParameters.header[].isValueSecret

`bool`

Store this value in the connection's Secrets Manager secret (true)
rather than the connection configuration (false). Set true for
anything credential-like.

### spec.connection.invocationHttpParameters.queryString

`[]AwsEventBridgeConnectionHttpParameter`

Query-string parameters.

### spec.connection.invocationHttpParameters.queryString[].key

`string` · required

The parameter name.

- rule: {"string":{"minLen":"1"}}

### spec.connection.invocationHttpParameters.queryString[].value

`string` · required · sensitive

The parameter value. The provider treats every value as sensitive
(AWS masks them all on read), so this field is secret-typed
regardless of is_value_secret.

- rule: {"string":{"minLen":"1"}}

### spec.connection.invocationHttpParameters.queryString[].isValueSecret

`bool`

Store this value in the connection's Secrets Manager secret (true)
rather than the connection configuration (false). Set true for
anything credential-like.

### spec.connection.privateInvocationEndpoint

`AwsEventBridgePrivateEndpoint`

Invoke a PRIVATE endpoint (behind a VPC Lattice resource
configuration) instead of a public one. Unset for public
endpoints.

### spec.connection.privateInvocationEndpoint.resourceConfigurationArn

`string` · required

The VPC Lattice resource configuration's ARN. AWS creates the
resource association itself (the association ARN is observable in
AWS, not modeled here).

- rule: {"string":{"minLen":"1","pattern":"^arn:aws.*$"}}

### spec.connection.privateAuthorizationEndpoint

`AwsEventBridgePrivateEndpoint`

Reach a PRIVATE OAuth authorization endpoint through VPC Lattice.
Only meaningful with oauth. Unset for public authorization
endpoints.

### spec.connection.privateAuthorizationEndpoint.resourceConfigurationArn

`string` · required

The VPC Lattice resource configuration's ARN. AWS creates the
resource association itself (the association ARN is observable in
AWS, not modeled here).

- rule: {"string":{"minLen":"1","pattern":"^arn:aws.*$"}}

### spec.connection.kmsKeyIdentifier

`string | valueFrom`

Customer-managed KMS key that encrypts the connection's secret
material (unset uses AWS-owned keys). The key policy must allow
Secrets Manager decryption scoped to the AWS-created secret
(kms:EncryptionContext:SecretARN on
"arn:aws:secretsmanager:*:*:secret:events!connection/*").
Reference an AwsKmsKey key_arn output or pass a literal key
id/ARN/alias.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.destination

`AwsEventBridgeDestination`

The owned API destination - the invocable endpoint.

### spec.destination.name

`string` · required

The destination's name in AWS (its identity - renaming replaces
it). Up to 64 characters: letters, digits, dot, dash, underscore.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_.-]+$"}}

### spec.destination.description

`string`

What this destination invokes.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512"}}

### spec.destination.connectionArn

`string | valueFrom`

A connection that exists elsewhere, by ARN - another instance's
owned connection (reference its connection_arn output) or any
pre-existing one. Leave unset when this instance's connection arm
serves the destination.

- references: AwsEventBridgeApiDestination (`status.outputs.connection_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEventBridgeApiDestination, name: <that resource's name>, fieldPath: status.outputs.connection_arn}} -- a bare string does not parse

### spec.destination.invocationEndpoint

`string` · required

The HTTPS endpoint to invoke. Path parameters use "*" wildcards
filled by the event target's path parameter values. Example:
"https://api.example.com/orders/*/refund".

- rule: {"string":{"minLen":"1","pattern":"^https://.+$"}}

### spec.destination.httpMethod

`string`

The HTTP method for invocations.

- rule: {"string":{"in":["GET","POST","PUT","PATCH","DELETE","HEAD","OPTIONS"]}}

### spec.destination.invocationRateLimitPerSecond

`int32` · optional (explicit presence)

The maximum invocations per second (AWS default: 300). Invocations
beyond the limit queue in EventBridge - size it to what the
downstream API tolerates. Presence-typed so 1 is expressible.

- rule: {"int32":{"gte":1}}

## Validation Rules

- `spec.at_least_one_arm`: configure the connection arm, the destination arm, or both - an empty spec manages nothing
- `spec.destination_connection_exactly_one`: the destination uses this instance's connection arm OR an external connection_arn - configure exactly one of the two

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEventBridgeApiDestination, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.connection_arn` | `string` | The owned connection's ARN - what other instances' destinations and pipe/rule targets reference. Empty when the instance has no connection arm. |
| `status.outputs.connection_secret_arn` | `string` | The Secrets Manager secret AWS created to hold the connection's credential values (the "events!connection/..." secret). AWS owns its lifecycle. Empty when the instance has no connection arm. |
| `status.outputs.api_destination_arn` | `string` | The owned API destination's ARN - what EventBridge rule targets, pipes, and schedules invoke. Empty when the instance has no destination arm. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.connection.kmsKeyIdentifier` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.destination.connectionArn` | AwsEventBridgeApiDestination | `status.outputs.connection_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEventBridgeApiDestination | `spec.destination.connectionArn` | `status.outputs.connection_arn` |
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.api_destination_arn` |

## See Also

- [Overview](../README.md)
