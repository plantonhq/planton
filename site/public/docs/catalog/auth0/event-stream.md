---
title: "Event Stream"
description: "Event Stream deployment documentation"
icon: "package"
order: 100
componentName: "auth0eventstream"
---

# Auth0 Event Stream

Deploys an Auth0 Event Stream that delivers real-time Auth0 events to an external destination. Supports AWS EventBridge for serverless event processing and HTTPS webhooks for custom endpoint delivery, with configurable event type subscriptions.

## What Gets Created

When you deploy an Auth0EventStream resource, OpenMCF provisions:

- **Auth0 Event Stream** — an `auth0_event_stream` resource configured with the specified destination type, event subscriptions, and destination-specific settings (EventBridge or webhook)

For EventBridge destinations, Auth0 creates a partner event source in the target AWS account that must be associated with an EventBridge event bus to begin receiving events.

## Prerequisites

- **Auth0 credentials** configured via environment variables (`AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET`) or OpenMCF provider config
- **An Auth0 tenant** with Event Streams enabled
- **An AWS account** (if using EventBridge) with permissions to accept partner event sources
- **A publicly accessible HTTPS endpoint** (if using webhooks) that can respond to POST requests within 10 seconds

## Quick Start

Create a file `auth0-event-stream.yaml`:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0EventStream
metadata:
  name: login-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.Auth0EventStream.login-events
spec:
  destinationType: webhook
  subscriptions:
    - authentication.success
    - authentication.failure
  webhookConfiguration:
    webhookEndpoint: "https://api.example.com/webhooks/auth0"
    webhookAuthorization:
      method: bearer
      token: "your-secret-token"
```

Deploy:

```shell
openmcf apply -f auth0-event-stream.yaml
```

This creates an event stream that delivers authentication success and failure events to the specified webhook endpoint using bearer token authorization.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `destinationType` | `string` | Destination where events are delivered. Determines which configuration block is required. | Must be one of: `eventbridge`, `webhook` |
| `subscriptions` | `string[]` | Event types this stream subscribes to. Only matching events are delivered. | At least one entry required |

### Optional Fields

#### EventBridge Configuration (`eventbridgeConfiguration`)

Required when `destinationType` is `eventbridge`. EventBridge configurations cannot be updated after creation; any change forces resource recreation.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `eventbridgeConfiguration.awsAccountId` | `string` | 12-digit AWS account ID where events are delivered. Auth0 creates a partner event source in this account. | Must match `^[0-9]{12}$` |
| `eventbridgeConfiguration.awsRegion` | `string` | AWS region for the EventBridge event bus. | Non-empty string |

#### Webhook Configuration (`webhookConfiguration`)

Required when `destinationType` is `webhook`. Webhook configurations can be updated after creation.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `webhookConfiguration.webhookEndpoint` | `string` | HTTPS URL that receives event payloads via POST. Must be publicly accessible. | Must match `^https://.+` |
| `webhookConfiguration.webhookAuthorization.method` | `string` | Authorization method for the webhook endpoint. | Must be one of: `basic`, `bearer` |
| `webhookConfiguration.webhookAuthorization.username` | `string` | Username for Basic authentication. Required when `method` is `basic`. | Required if `method` is `basic` |
| `webhookConfiguration.webhookAuthorization.password` | `string` | Password for Basic authentication. Stored securely, never returned by the API. Required when `method` is `basic`. | Required if `method` is `basic` |
| `webhookConfiguration.webhookAuthorization.token` | `string` | Bearer token for token-based authentication. Stored securely, never returned by the API. Required when `method` is `bearer`. | Required if `method` is `bearer` |

## Examples

### EventBridge — Security Monitoring

Stream authentication events to AWS EventBridge for processing by Lambda functions or a SIEM integration:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0EventStream
metadata:
  name: security-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0EventStream.security-events
spec:
  destinationType: eventbridge
  subscriptions:
    - authentication.success
    - authentication.failure
    - user.created
    - user.updated
  eventbridgeConfiguration:
    awsAccountId: "123456789012"
    awsRegion: us-east-1
```

After deployment, associate the partner event source (available in `status.outputs.awsPartnerEventSource`) with an EventBridge event bus in the target AWS account, then create rules to route events to downstream targets.

### Webhook — Bearer Token Authorization

Deliver user lifecycle events to an HTTPS endpoint, authenticated with a bearer token:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0EventStream
metadata:
  name: user-lifecycle
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.Auth0EventStream.user-lifecycle
spec:
  destinationType: webhook
  subscriptions:
    - user.created
    - user.updated
    - user.deleted
  webhookConfiguration:
    webhookEndpoint: "https://api.example.com/webhooks/auth0/users"
    webhookAuthorization:
      method: bearer
      token: "dGhpcyBpcyBhIHNlY3VyZSB0b2tlbg=="
```

Generate a secure token with:

```shell
openssl rand -base64 32
```

### Webhook — Basic Authentication

Deliver API authorization events to an internal endpoint using HTTP Basic authentication:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0EventStream
metadata:
  name: api-audit
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0EventStream.api-audit
spec:
  destinationType: webhook
  subscriptions:
    - api.authorization.success
    - api.authorization.failure
    - management.client.created
    - management.connection.updated
  webhookConfiguration:
    webhookEndpoint: "https://audit.internal.example.com/auth0"
    webhookAuthorization:
      method: basic
      username: auth0-webhook
      password: "s3cureP@ssword!"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Unique Auth0 identifier for the event stream. Format: `est_XXXXXXXXXXXXXXXX` |
| `name` | `string` | Name of the event stream, derived from `metadata.name` |
| `status` | `string` | Current status of the stream: `active`, `suspended`, or `disabled` |
| `destinationType` | `string` | Destination type: `eventbridge` or `webhook` |
| `createdAt` | `string` | ISO 8601 timestamp when the stream was created |
| `updatedAt` | `string` | ISO 8601 timestamp when the stream was last updated |
| `subscriptions` | `string[]` | Event types this stream is subscribed to |
| `awsPartnerEventSource` | `string` | AWS partner event source name. Only populated for EventBridge destinations. Format: `aws.partner/auth0.com/<tenant-id>/<stream-name>` |

## Related Components

- [Auth0Client](/docs/catalog/auth0/client) — applications that generate the authentication and user events consumed by this stream
- [Auth0ResourceServer](/docs/catalog/auth0/resource-server) — APIs whose authorization events can be streamed via `api.authorization.*` subscriptions
- [Auth0Connection](/docs/catalog/auth0/connection) — authentication connections whose events can be monitored through this stream
