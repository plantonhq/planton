# Auth0EventStream

Auth0EventStream is a Planton deployment component that manages Auth0 Event Streams. Event Streams enable real-time delivery of Auth0 events to external systems like AWS EventBridge or custom webhook endpoints.

## Overview

Auth0 Event Streams provide a way to subscribe to tenant-level events and deliver them in real-time to external systems. This is essential for:

- **Security Monitoring**: Stream authentication events to SIEM systems
- **User Analytics**: Track user lifecycle events for analytics platforms
- **Workflow Automation**: Trigger actions when users sign up or authenticate
- **Compliance Auditing**: Maintain audit logs of all authentication activity

## Supported Destinations

| Destination | Description | Update Policy |
|-------------|-------------|---------------|
| `eventbridge` | AWS EventBridge for serverless event processing | **Immutable** - Recreate to change |
| `webhook` | Custom HTTPS endpoint for event delivery | **Mutable** - Can update after creation |

## Usage

### EventBridge Destination

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0EventStream
metadata:
  name: security-events
  org: my-organization
  env: production
spec:
  destination_type: eventbridge
  subscriptions:
    - user.created
    - user.updated
    - authentication.success
    - authentication.failure
  eventbridge_configuration:
    aws_account_id: "123456789012"
    aws_region: us-east-1
```

### Webhook Destination with Bearer Token

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0EventStream
metadata:
  name: user-events-webhook
  org: my-organization
  env: production
spec:
  destination_type: webhook
  subscriptions:
    - user.created
    - user.updated
    - user.deleted
  webhook_configuration:
    webhook_endpoint: "https://api.example.com/webhooks/auth0"
    webhook_authorization:
      method: bearer
      token: "your-secret-token"
```

### Webhook Destination with Basic Auth

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0EventStream
metadata:
  name: siem-events
  org: my-organization
  env: production
spec:
  destination_type: webhook
  subscriptions:
    - authentication.success
    - authentication.failure
    - api.authorization.failure
  webhook_configuration:
    webhook_endpoint: "https://siem.example.com/auth0/events"
    webhook_authorization:
      method: basic
      username: webhook-user
      password: super-secret-password
```

## Event Types

Common event types you can subscribe to:

### User Events
- `user.created` - A new user was created
- `user.updated` - A user profile was updated
- `user.deleted` - A user was deleted
- `user.blocked` - A user was blocked
- `user.unblocked` - A user was unblocked

### Authentication Events
- `authentication.success` - Successful authentication
- `authentication.failure` - Failed authentication attempt

### API Authorization Events
- `api.authorization.success` - Successful API authorization
- `api.authorization.failure` - Failed API authorization

### Management API Events
- `management.client.created` - An application was created
- `management.connection.updated` - A connection was updated

For a complete list, see [Auth0 Event Types Documentation](https://auth0.com/docs/customize/log-streams/event-types).

## Deployment

Deploy using the Planton CLI:

```bash
# Create Auth0 credentials file
cat > auth0-creds.yaml << EOF
domain: your-tenant.auth0.com
clientId: your-client-id
clientSecret: your-client-secret
EOF

# Deploy the event stream
planton apply --manifest eventstream.yaml \
  --auth0-provider-config auth0-creds.yaml
```

## Stack Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `id` | Unique identifier of the event stream (est_XXXX) |
| `name` | Name of the event stream |
| `status` | Current status (active, suspended, disabled) |
| `destination_type` | Destination type (eventbridge or webhook) |
| `subscriptions` | List of subscribed event types |
| `created_at` | ISO 8601 timestamp of creation |
| `updated_at` | ISO 8601 timestamp of last update |
| `aws_partner_event_source` | AWS partner event source (EventBridge only) |

## AWS EventBridge Setup

When using EventBridge as the destination:

1. Auth0 creates a partner event source in your AWS account
2. The partner event source name is available in `aws_partner_event_source` output
3. Associate the event source with an EventBridge event bus in AWS
4. Create rules on the event bus to route events to your targets (Lambda, SQS, etc.)

```bash
# After deployment, get the partner event source
aws events describe-partner-event-source \
  --name "aws.partner/auth0.com/tenant-id/security-events"

# Create an event bus from the partner event source
aws events create-partner-event-source-connection \
  --name "auth0-security-events" \
  --partner-event-source "aws.partner/auth0.com/tenant-id/security-events"
```

## Related Components

- [Auth0Connection](../auth0connection/v1/README.md) - Manage identity provider connections
- [Auth0Client](../auth0client/v1/README.md) - Manage Auth0 applications

## References

- [Auth0 Event Streams Documentation](https://auth0.com/docs/customize/log-streams)
- [Auth0 Event Types Reference](https://auth0.com/docs/customize/log-streams/event-types)
- [Terraform auth0_event_stream Resource](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/event_stream)
- [Pulumi Auth0 EventStream Resource](https://www.pulumi.com/registry/packages/auth0/api-docs/eventstream/)

