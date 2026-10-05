# AWS EventBridge API Destination

Deploys an EventBridge connection and API destination — the authenticated HTTPS endpoint that EventBridge rules, pipes, and schedules invoke directly, with no credential-holding Lambda in between. The connection is the auth trust anchor (api-key, basic, or OAuth client credentials, with the credential values stored in a Secrets Manager secret AWS creates and owns); the destination is the invocable endpoint with its HTTP method and rate cap. The two arms deploy together in one instance, or split: one connection-owning instance can serve many destination-only instances that bind to it by ARN.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EventBridge Connection** — created only when the `connection` arm is configured. Carries the auth mode (exactly one of `apiKey`, `basic`, `oauth`), optional static parameters (headers, query string, body fields) added to every invocation, optional private endpoints through VPC Lattice, and optional customer-managed KMS encryption of its secret material
- **EventBridge API Destination** — created only when the `destination` arm is configured. The HTTPS endpoint (with `*` path wildcards), the HTTP method, and the invocations-per-second cap, bound to this instance's connection or to an external one by ARN
- **AWS-owned Secrets Manager secret** — derived, not module-created: AWS creates and owns the `events!connection/...` secret that holds the connection's credential values. Its ARN surfaces as the `connection_secret_arn` output; no AWS read API ever returns the values

Neither resource is taggable at AWS — the usual tag conventions deliberately do not apply here.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with EventBridge and Secrets Manager permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Org secrets for every credential value** — the api key, basic password, OAuth client secret, and every connection HTTP parameter value are sensitive fields: the backend accepts only `$secret/<slug>` references there, never plaintext. Create the org secrets before applying the manifest.

### AWS Account

- The downstream API's credentials in hand — they enter the spec as secret references and land in the AWS-owned Secrets Manager secret.
- For private endpoints (`privateInvocationEndpoint` / `privateAuthorizationEndpoint`), a VPC Lattice resource configuration fronting the target — its ARN goes in the spec; AWS creates the resource association itself.
- For customer-managed encryption (`kmsKeyIdentifier`), a KMS key whose policy allows Secrets Manager decryption scoped to the AWS-created secret — condition on `kms:EncryptionContext:SecretARN` matching `arn:aws:secretsmanager:*:*:secret:events!connection/*`.

## Deploy

### Console

Open the deployment store, find **AWS EventBridge API Destination**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the two arms: the connection (auth mode, credentials as org-secret references, invocation parameters) and the destination (endpoint, method, rate cap). Start from the **API-Key Webhook** preset in the [Presets](#presets) tab for the workhorse shape, or the **OAuth SaaS Integration** preset when the downstream issues tokens via client credentials.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeApiDestination
metadata:
  name: partner-webhook
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  connection:
    name: partner-api
    description: Partner ingestion API (api-key auth)
    apiKey:
      key: x-api-key
      value: $secret/partner-api-key
  destination:
    name: partner-webhook
    description: Partner order-events webhook
    invocationEndpoint: https://api.example.com/events
    httpMethod: POST
    invocationRateLimitPerSecond: 50
```

```shell
planton apply -f eventbridge-api-destination.yaml
```

This creates an api-key connection (the key lands in the AWS-owned secret, never in state) and a POST destination capped at 50 invocations per second, ready to be targeted by a rule, pipe, or schedule. An Infra Job tracks the provisioning in real time.

### InfraChart

When several destinations share one connection in a chart, keep the connection in one owning instance and bind destination-only instances to it via ValueFromRef:

```yaml
spec:
  region: us-east-1
  destination:
    name: billing-webhook
    connectionArn:
      valueFrom:
        kind: AwsEventBridgeApiDestination
        name: partner-connection
        fieldPath: status.outputs.connection_arn
    invocationEndpoint: https://api.example.com/billing
    httpMethod: POST
```

The InfraPipeline resolves the dependency graph, deploys the connection-owning instance first, then binds each destination to its connection ARN.

## Key Configuration

These are the most important decisions when configuring an API destination. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One instance or a split** — Both arms in one instance is the default for a single integration. When many endpoints share one downstream's credentials, split: a connection-only instance owns the trust anchor, and destination-only instances reference its `connection_arn` output. The spec enforces exactly one connection source per destination — the owned arm or an external ARN, never both.

**Auth mode is exactly one of three** — `apiKey`, `basic`, or `oauth`; the modules derive AWS's authorization type from whichever block is set, so the two can never disagree. OAuth requires `oauthHttpParameters` with at least one entry — most token servers need `grant_type=client_credentials` in the token request's body, and AWS masks every connection HTTP parameter on read, so even that value is secret-typed.

**Names are identities** — Renaming the connection or the destination replaces the resource. A replaced connection re-walks authorization; a replaced destination gets a new ARN, orphaning every rule, pipe, and schedule that targeted the old one. Name them once, deliberately.

**Credentials live in AWS's secret, not yours** — No AWS read API returns the credential values, so drift in a credential is invisible and the manifest's `$secret` references are the source of truth. Rotation is an in-place update: change the secret's value and re-apply — AWS re-authorizes the connection without replacing it. A connection stuck DEAUTHORIZED after an update means the downstream rejected the new credentials; the connection's StateReason says why.

**The rate limit queues, not drops** — Invocations beyond `invocationRateLimitPerSecond` (AWS default: 300) queue in EventBridge for up to 24 hours, then dead-letter (if the invoking rule has a DLQ) or drop. A too-low limit shows up as delay first, loss second — size it to the downstream API's real ceiling.

**The endpoint is not validated at create** — AWS accepts any HTTPS URL: a typo'd `invocationEndpoint` deploys green and fails at first invocation. Test with a real event before wiring production rules. Path parameters use `*` wildcards filled by the event target's path parameter values.

**Budget for the auth state machine** — Connection creates and credential updates walk CREATING/AUTHORIZING → AUTHORIZED. Usually seconds to minutes, but the modules budget up to 20 minutes — first deploys against slow OAuth servers are not stuck, just authorizing.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEventBridgeApiDestination** | `destination.connectionArn` | `status.outputs.connection_arn` |
| **AwsKmsKey** | `connection.kmsKeyIdentifier` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `api_destination_arn` | The owned destination's ARN; empty when the instance has no destination arm | The target ARN for EventBridge rules, pipes, and schedules |
| `connection_arn` | The owned connection's ARN; empty when the instance has no connection arm | Destination-only instances bind to it via `connectionArn` |

`connection_secret_arn` is also exported — the AWS-owned Secrets Manager secret holding the credential values. It exists for KMS key-policy scoping and audit trails rather than composition; nothing downstream consumes it as an input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-integration webhook** — Both arms in one instance: an api-key connection and a POST destination with a rate cap sized to the partner's ceiling. The simplest shape, and the right one until a second endpoint needs the same credentials. Start from the **API-Key Webhook** preset.

**OAuth against a SaaS** — EventBridge fetches the token itself from the authorization endpoint and invokes under it; no token-refresh code to own. The trade is setup precision: the token request's `grant_type` and `scope` ride `oauthHttpParameters` as secret references, and a wrong body shape surfaces only as a DEAUTHORIZED connection. Start from the **OAuth SaaS Integration** preset.

**Shared connection, many destinations** — One connection-only instance owns the downstream's credentials; each integration deploys a destination-only instance bound by `connectionArn`. Credential rotation touches one manifest instead of many, at the cost of a shared blast radius — a DEAUTHORIZED connection stops every destination on it.

## Works With

- [**AWS EventBridge Rule**](/infra-catalog/aws-event-bridge-rule) — routes matched bus events to the destination's `api_destination_arn`
- [**AWS EventBridge Pipe**](/infra-catalog/aws-event-bridge-pipe) — streams source events (SQS, Kinesis, DynamoDB) into the destination point-to-point
- [**AWS EventBridge Scheduler**](/infra-catalog/aws-event-bridge-scheduler) — invokes the destination on a cron or rate schedule via its universal-target arm
- [**AWS EventBridge Bus**](/infra-catalog/aws-event-bridge-bus) — the custom bus whose rules commonly front the destination
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for the connection's secret material via `kmsKeyIdentifier`
