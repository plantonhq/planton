# Auth0 Event Stream

Deploys an Auth0 Event Stream that delivers tenant events — authentication results, user lifecycle changes, API authorizations — to AWS EventBridge or an HTTPS webhook in near real time. Subscriptions scope the stream to exactly the event types you name, and the two destination types carry different lifecycle rules, so the destination choice is the decision that matters most.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Auth0 Event Stream** — a stream configured with the specified destination type, event subscriptions, and destination-specific settings

For EventBridge destinations, Auth0 creates a partner event source in the target AWS account — events flow only after you associate that source with an EventBridge event bus. For webhook destinations, Auth0 delivers event payloads as HTTPS POST requests to the configured endpoint with the specified authorization credentials.

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** — an active connection in the Connect module with Auth0 domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **Management API scopes** — the M2M application behind the Provider Connection needs `create:event_streams`, `read:event_streams`, `update:event_streams`, and `delete:event_streams`. The update scope is exercised only for webhook destinations — an EventBridge configuration change forces delete-and-recreate instead.
- **An AWS account able to accept partner event sources** (only for `eventbridge`) — you need the 12-digit account ID and the target region for `eventbridgeConfiguration`.
- **A publicly accessible HTTPS endpoint** (only for `webhook`) — it must answer POST requests with a 2xx within 10 seconds and accept either Basic or Bearer token authorization.

## Deploy

### Console

Open the deployment store, find **Auth0 Event Stream**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the destination type with its EventBridge or webhook settings, and the event subscription list.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EventStream
metadata:
  name: login-events
  org: acme-corp
  env: prod
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

```shell
planton apply -f auth0-event-stream.yaml
```

This creates a stream that POSTs authentication success and failure events to the webhook endpoint with bearer token authorization. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an Auth0 Event Stream. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destination type** — `destinationType` is the one-way door. EventBridge configurations are immutable after creation: any change to `awsAccountId` or `awsRegion` destroys and recreates the stream, generating a NEW partner event source name that must be re-associated with your event bus. Webhook configurations update in place — endpoint and authorization included. Choose `eventbridge` for AWS-native processing (Lambda, Step Functions, SIEM pipelines); choose `webhook` when the consumer is a custom HTTPS endpoint or you expect the destination to change.

**Event subscriptions** — The `subscriptions` array names the event types delivered; at least one is required. Common types: `authentication.success`, `authentication.failure`, `user.created`, `user.updated`, `api.authorization.success`. Auth0 events carry user PII — emails, IP addresses, user agents — so subscribe only to the types the destination system is entitled to handle, not everything available.

**Webhook authorization** — `webhookConfiguration.webhookAuthorization.method` sets how Auth0 authenticates to your endpoint: `bearer` with a generated token (`openssl rand -base64 32`) for most integrations, `basic` with username and password for systems that require it. Configure the same credential on your webhook server — an endpoint that skips validation will accept forged events from anyone who finds its URL.

**EventBridge placement** — `eventbridgeConfiguration.awsRegion` decides where the partner event source lands; put it in the region where your event processing runs. `awsAccountId` must be the 12-digit account number. Both values are locked in by the immutability rule above, so confirm them before the first apply.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The AWS account for EventBridge destinations is supplied as a plain account ID rather than a typed reference.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | Auth0 event stream identifier (`est_...`) | Management API operations on the stream |
| `status` | Stream state (`active`, `suspended`, `disabled`) | Health checks and delivery alerting |
| `aws_partner_event_source` | Partner event source name (EventBridge only) | Associating the source with an EventBridge event bus |

## Common Patterns

**Security monitoring on EventBridge** — Stream `authentication.success` and `authentication.failure` to EventBridge and route them with bus rules into a SIEM or a Lambda-based detector. EventBridge's fan-out means adding a new consumer never touches the Auth0 side.

**User lifecycle sync over webhook** — Subscribe to `user.created`, `user.updated`, and `user.deleted` and POST them to a provisioning endpoint that mirrors Auth0 users into internal systems. The webhook destination fits here because the receiving endpoint tends to evolve — it can be updated in place.

**Long-term log retention** — Auth0's built-in log retention is plan-limited (days, not months). A stream into external storage is the standing mechanism for keeping authentication history beyond that window without upgrading the Auth0 plan.

## Works With

- [**AWS EventBridge Bus**](/infra-catalog/aws-event-bridge-bus) — the event bus the partner event source is associated with; rules on the bus route Auth0 events to targets
- [**AWS Lambda**](/infra-catalog/aws-lambda) — the usual EventBridge rule target for processing Auth0 events serverlessly
