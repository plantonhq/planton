# Auth0EventStream

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

Auth0EventStreamSpec defines the configuration for an Auth0 Event Stream.
Event Streams enable real-time delivery of Auth0 events to external systems,
supporting both AWS EventBridge and Webhook destinations.

This spec covers the 80/20 use case for configuring Auth0 event streams:
- EventBridge integration for AWS-native event processing
- Webhook integration for custom HTTP endpoints

Supported destination types:
- eventbridge: AWS EventBridge for serverless event processing
- webhook: HTTPS endpoint for custom event handling

https://auth0.com/docs/customize/log-streams
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/event_stream

## Example

```yaml
# Test manifest for Auth0EventStream
# This file is used for local testing and development

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EventStream
metadata:
  name: test-event-stream
  org: test-organization
  env: development
  labels:
    team: platform
    purpose: testing
spec:
  destination_type: webhook
  subscriptions:
    - user.created
    - user.updated
  webhook_configuration:
    webhook_endpoint: "https://api.planton.ai"
    webhook_authorization:
      method: bearer
      token: "test-token-for-development"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.destinationType` | `string` | yes |  |  |
| `spec.subscriptions` | `[]string` | yes |  |  |
| `spec.eventbridgeConfiguration` | `Auth0EventBridgeConfiguration` |  |  |  |
| `spec.eventbridgeConfiguration.awsAccountId` | `string` | yes |  |  |
| `spec.eventbridgeConfiguration.awsRegion` | `string` | yes |  |  |
| `spec.webhookConfiguration` | `Auth0WebhookConfiguration` |  |  |  |
| `spec.webhookConfiguration.webhookEndpoint` | `string` | yes |  |  |
| `spec.webhookConfiguration.webhookAuthorization` | `Auth0WebhookAuthorization` | yes |  |  |
| `spec.webhookConfiguration.webhookAuthorization.method` | `string` | yes |  |  |
| `spec.webhookConfiguration.webhookAuthorization.username` | `string` |  |  |  |
| `spec.webhookConfiguration.webhookAuthorization.password` | `string` (sensitive) |  |  |  |
| `spec.webhookConfiguration.webhookAuthorization.token` | `string` (sensitive) |  |  |  |

## Field Details

### spec.destinationType

`string` · required

destination_type specifies where events should be delivered.
This determines which configuration block is required.

- "eventbridge": Events are delivered to AWS EventBridge.
  Requires eventbridge_configuration to be set.
  EventBridge configurations cannot be updated after creation.

- "webhook": Events are delivered to an HTTPS endpoint.
  Requires webhook_configuration to be set.
  Webhook configurations can be updated after creation.

https://auth0.com/docs/customize/log-streams#event-stream-destinations

- rule: {"required":true,"string":{"in":["eventbridge","webhook"]}}

### spec.subscriptions

`[]string` · required

subscriptions is a list of event types this stream is subscribed to.
Only events matching these types will be delivered to the destination.
At least one subscription is required.

Common event types:
- User events: user.created, user.updated, user.deleted
- Authentication: authentication.success, authentication.failure
- API operations: api.authorization.success, api.authorization.failure
- Management API: management.client.created, management.connection.updated

For a complete list of event types, see:
https://auth0.com/docs/customize/log-streams/event-types

- rule: {"repeated":{"minItems":"1"}}

### spec.eventbridgeConfiguration

`Auth0EventBridgeConfiguration`

eventbridge_configuration contains settings for AWS EventBridge destination.
Only applicable when destination_type is "eventbridge".

EventBridge configurations CANNOT be updated after creation.
Any change will force the resource to be recreated.

### spec.eventbridgeConfiguration.awsAccountId

`string` · required

aws_account_id is the AWS account ID where events will be delivered.
This is the 12-digit AWS account number.
Example: "123456789012"

Auth0 will create a partner event source in this account.

- rule: {"required":true,"string":{"pattern":"^[0-9]{12}$"}}

### spec.eventbridgeConfiguration.awsRegion

`string` · required

aws_region is the AWS region for the EventBridge event bus.
Events will be delivered to this region.
Example: "us-east-1", "eu-west-1"

Common regions:
- us-east-1 (N. Virginia)
- us-west-2 (Oregon)
- eu-west-1 (Ireland)
- ap-southeast-1 (Singapore)

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.webhookConfiguration

`Auth0WebhookConfiguration`

webhook_configuration contains settings for webhook destination.
Only applicable when destination_type is "webhook".

Webhook configurations CAN be updated after creation,
including the endpoint and authorization settings.

### spec.webhookConfiguration.webhookEndpoint

`string` · required

webhook_endpoint is the HTTPS URL that will receive webhook events.
Must be a valid, publicly accessible HTTPS URL.
Auth0 will POST event payloads to this endpoint.

Requirements:
- Must use HTTPS (HTTP is not supported)
- Must be publicly accessible
- Should respond with 2xx status within 10 seconds

Example: "https://api.example.com/webhooks/auth0"

- rule: {"required":true,"string":{"pattern":"^https://.+"}}

### spec.webhookConfiguration.webhookAuthorization

`Auth0WebhookAuthorization` · required

webhook_authorization contains authentication settings for the webhook endpoint.
Auth0 will include these credentials when calling your endpoint.

- rule: {"required":true}
- rule: token is required when authorization method is 'bearer'. Generate one with: openssl rand -base64 32
- rule: username is required when authorization method is 'basic'
- rule: password is required when authorization method is 'basic'

### spec.webhookConfiguration.webhookAuthorization.method

`string` · required

method specifies the authorization method for the webhook.

- "basic": HTTP Basic Authentication using username and password.
  Auth0 sends: Authorization: Basic base64(username:password)

- "bearer": Bearer token authentication.
  Auth0 sends: Authorization: Bearer <token>

- rule: {"required":true,"string":{"in":["basic","bearer"]}}

### spec.webhookConfiguration.webhookAuthorization.username

`string`

username is the username for Basic authentication.
Required when method is "basic".
Ignored when method is "bearer".

### spec.webhookConfiguration.webhookAuthorization.password

`string` · sensitive

password is the password for Basic authentication.
Required when method is "basic".
This value is stored securely and never returned by the API.
Ignored when method is "bearer".

### spec.webhookConfiguration.webhookAuthorization.token

`string` · sensitive

token is the bearer token for token-based authentication.
Required when method is "bearer".
This value is stored securely and never returned by the API.
Ignored when method is "basic".

Generate a secure token using OpenSSL:
  openssl rand -base64 32   # Base64 encoded (recommended)
  openssl rand -hex 32      # Hexadecimal

The same token must be configured on your webhook server to validate
incoming requests via the Authorization: Bearer <token> header.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0EventStream, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the unique identifier of the Auth0 event stream. This is used internally by Auth0 to identify the event stream. Format: "est_XXXXXXXXXXXXXXXX" |
| `status.outputs.name` | `string` | name is the name of the event stream. Derived from metadata.name in the Auth0EventStream resource. |
| `status.outputs.status` | `string` | status is the current status of the event stream. Possible values: - "active": The stream is operational and delivering events - "suspended": The stream is temporarily paused - "disabled": The stream is disabled and not delivering events |
| `status.outputs.destination_type` | `string` | destination_type is the type of event stream destination. Either "eventbridge" or "webhook". |
| `status.outputs.created_at` | `string` | created_at is the ISO 8601 timestamp when the stream was created. Example: "2024-01-15T10:30:00.000Z" |
| `status.outputs.updated_at` | `string` | updated_at is the ISO 8601 timestamp when the stream was last updated. Example: "2024-01-16T14:45:30.000Z" |
| `status.outputs.subscriptions` | `[]string` | subscriptions is the list of event types this stream is subscribed to. Reflects the subscriptions configured in the spec. |
| `status.outputs.aws_partner_event_source` | `string` | aws_partner_event_source is the AWS partner event source name. Only populated when destination_type is "eventbridge". This is the event source that must be associated with an EventBridge event bus. Format: "aws.partner/auth0.com/<tenant-id>/<stream-name>" |

## See Also

- [Overview](../README.md)
