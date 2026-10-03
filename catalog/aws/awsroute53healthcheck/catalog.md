# AWS Route 53 Health Check

Deploys a Route 53 health check — the availability signal DNS records reference to keep unhealthy endpoints out of DNS answers. Health checks power failover routing (primary/secondary), health-aware weighted routing, and multivalue answers. The `checkType` selects one of four monitoring models: endpoint probing from Route 53's global checker fleet (HTTP, HTTPS, body-match variants, TCP), CALCULATED aggregation of child checks into service-level health, CLOUDWATCH_METRIC mirroring of an alarm (the way to health-check private resources), and RECOVERY_CONTROL mirroring of an Application Recovery Controller switch for DR runbooks.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Route 53 Health Check** -- the monitoring model chosen by `checkType` (create-time immutable), with its per-model surface: probe target, child set, mirrored alarm, or routing control
- **Probe Configuration** -- interval (10s/30s, create-time immutable; AWS defaults to 30), failure threshold (AWS defaults to 3), optional latency graphing (create-time immutable), SNI, and an optional checker-region subset (min 3). Probe tuning applies to ENDPOINT checks only -- the aggregation, alarm, and recovery models take none of it, enforced at authoring time
- **Behavior Dials** -- inversion and the administrative disable switch (the maintenance lever), both editable in place
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Model prerequisites** -- CALCULATED checks reference other [AWS Route 53 Health Check](/infra-catalog/aws-route53-health-check) resources (deploy the children first); CLOUDWATCH_METRIC checks name an [AWS CloudWatch Alarm](/infra-catalog/aws-cloudwatch-alarm) (or composite) by its `alarm_name` output.

### AWS Account

- **Endpoint checks probe from the public internet** -- the target must be reachable by Route 53's checker fleet (security groups / firewalls admitting the checker ranges). For private resources, use the CLOUDWATCH_METRIC model instead.
- **Recovery-control checks** need an Application Recovery Controller cluster, control panel, and routing control already configured.

## Deploy

### Console

Open the deployment store, find **AWS Route 53 Health Check**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the monitoring model you pick on the first step shapes the rest of the flow. Start from the **HTTPS Endpoint Health Check** preset in the [Presets](#presets) tab for the production web shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53HealthCheck
metadata:
  name: orders-api-health
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  checkType: HTTPS
  fqdn: api.example.com
  resourcePath: /healthz
  failureThreshold: 3
  measureLatency: true
```

```shell
planton apply -f health-check.yaml
```

This probes `https://api.example.com/healthz` from the global checker fleet — about 90 seconds from outage to DNS reaction at the default cadence. An Infra Job tracks the provisioning in real time.

### InfraChart

When a CALCULATED check and its children deploy in one chart, wire the child references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  checkType: CALCULATED
  childHealthChecks:
    - valueFrom:
        kind: AwsRoute53HealthCheck
        name: orders-api-health
        fieldPath: status.outputs.health_check_id
    - valueFrom:
        kind: AwsRoute53HealthCheck
        name: orders-web-health
        fieldPath: status.outputs.health_check_id
  childHealthThreshold: 1
```

The InfraPipeline resolves the dependency graph, deploys the child checks first, then creates the aggregate on their IDs.

## Key Configuration

These are the most important decisions when configuring a health check. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The monitoring model is the one-way door** -- `checkType`, the probe interval, latency graphing, and the routing-control ARN are all create-time immutable: changing any of them replaces the check (new ID), and every referencing DNS record must re-point. Decide the model first; everything else in the wizard derives from it.

**Detection speed is interval × threshold** -- the default 30s × 3 reacts in ~90 seconds; fast checks (10s) with threshold 2 cut that to ~20 seconds at higher cost. The threshold updates in place — tune it freely.

**Private resources use the CloudWatch model** -- the checker fleet cannot reach a private ALB or internal API. A CloudWatch alarm watches from inside (by metric or heartbeat), and the health check mirrors its state into DNS — also the way to gate DNS on application-level signals like checkout success rate.

**The disable switch is the maintenance lever** -- `disabled` stops evaluation and reports always-healthy so planned work never triggers failover; it flips in place. Re-enabling is the runbook's last step — a forgotten disabled check permanently defeats the failover it gates.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsRoute53HealthCheck** (children of a CALCULATED check) | `childHealthChecks` | `status.outputs.health_check_id` |

A CLOUDWATCH_METRIC check names its [AwsCloudwatchAlarm](/infra-catalog/aws-cloudwatch-alarm) (or [AwsCloudwatchCompositeAlarm](/infra-catalog/aws-cloudwatch-composite-alarm)) by plain alarm name and region — `cloudwatchAlarmName` is a string, not a typed reference.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `health_check_id` | The health check's UUID | An AwsRoute53DnsRecord's `health_check_id` (failover / weighted / multivalue routing); a CALCULATED parent's `child_health_checks` |
| `health_check_arn` | Amazon Resource Name of the check | IAM policies; the CloudWatch `HealthCheckStatus` metric dimension |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTPS endpoint probe** -- the production web default: probe a real health endpoint over TLS. Start from the **HTTPS Endpoint Health Check** preset.

**Alarm-backed private health** -- a CLOUDWATCH_METRIC check mirroring an alarm on a private service, gating DNS failover on internal telemetry. Start from the **CloudWatch Alarm Health Check** preset.

## Works With

- [**AWS Route 53 DNS Record**](/infra-catalog/aws-route53-dns-record) -- the records whose failover / weighted / multivalue routing gates on this check's `health_check_id`
- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) -- the hosted zone those records live in
- [**AWS CloudWatch Alarm**](/infra-catalog/aws-cloudwatch-alarm) -- the alarm a CLOUDWATCH_METRIC check mirrors
- [**AWS CloudWatch Composite Alarm**](/infra-catalog/aws-cloudwatch-composite-alarm) -- a whole-service verdict a CLOUDWATCH_METRIC check can mirror into DNS
