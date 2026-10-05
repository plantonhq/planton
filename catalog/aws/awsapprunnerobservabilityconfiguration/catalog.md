# AWS App Runner Observability Configuration

Deploys an App Runner observability configuration — the reusable tracing policy that [App Runner services](/infra-catalog/aws-app-runner-service) reference to enable distributed request tracing. It is deliberately its own resource: one configuration is shared by any number of services, and attaching the reference IS the tracing on-switch — there is no separate toggle to keep in sync per service. With tracing configured, each instance of a referencing service runs an OpenTelemetry collector sidecar that forwards request spans to AWS X-Ray. Treat the configuration as immutable once created: to change tracing posture, register a new configuration name and repoint services (see the configuration notes).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **App Runner Observability Configuration** -- a named, versioned tracing policy; the trace settings are fixed at creation (see the change-workflow note below)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Application instrumentation** -- tracing only produces spans when the application emits them via the AWS Distro for OpenTelemetry (ADOT) SDK. This configuration enables the collector; the instrumentation is the application's half.
- **X-Ray costs** -- X-Ray bills per trace recorded and retrieved; tune the SDK's sampling rules for high-traffic services.

## Deploy

### Console

Open the deployment store, find **AWS App Runner Observability Configuration**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **X-Ray Tracing** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerObservabilityConfiguration
metadata:
  name: xray-tracing
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  traceConfiguration:
    vendor: AWSXRAY
```

```shell
planton apply -f app-runner-observability.yaml
```

This registers an X-Ray tracing policy. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an observability configuration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Presence is the switch** -- `traceConfiguration` present means referencing services trace; absent means the configuration registers without tracing (a valid but inert state, occasionally useful as a placeholder a later revision fills in). There is no boolean.

**AWS X-Ray is the only vendor** -- App Runner supports `AWSXRAY` today; the collector sidecar forwards spans there, and the service map, per-hop latency, and error attribution appear in the region's X-Ray console.

**Two halves make tracing work** -- this configuration enables the collector; the application must emit spans via the ADOT SDK. An uninstrumented app traces nothing even with X-Ray attached.

**Changing tracing means a new configuration name** (upstream provider gap) -- the provider's update path is tags-only and does not replace on trace changes, so editing `traceConfiguration` on an existing configuration applies cleanly and silently changes NOTHING at AWS. To change tracing posture, create a new configuration (a new `metadata.name`) and repoint referencing services; with X-Ray as the only supported vendor, the block's practical states are present or absent at creation.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — it is a leaf resource: App Runner services reference its ARN, never the other way around.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `configuration_arn` | Revision-carrying ARN of the configuration | App Runner service `observabilityConfigurationArn` |
| `configuration_revision` | The revision number this deployment registered | Audit and rollout tracking |
| `latest` | Whether this revision is the configuration name's latest | Fleet verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Fleet-wide X-Ray** -- one `xray-tracing` configuration per region, referenced by every production service, so the whole fleet appears in one service map. Start from the **X-Ray Tracing** preset.

**Selective tracing** -- attach the reference only to the services worth the X-Ray spend; detaching the reference (or referencing nothing) turns tracing off per service.

## Works With

- [**AWS App Runner Service**](/infra-catalog/aws-app-runner-service) -- adopts this tracing policy via `observabilityConfigurationArn` (consumes `configuration_arn`)
