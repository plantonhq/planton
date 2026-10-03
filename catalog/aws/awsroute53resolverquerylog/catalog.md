# AWS Route 53 Resolver Query Logging

Deploys a Route 53 Resolver query logging configuration — the pipeline that records every DNS query your VPCs make, including DNS Firewall verdicts — with the VPC associations that turn it on managed in-line. Logs land in a CloudWatch log group, an S3 bucket, or a Kinesis Data Firehose stream, chosen by destination ARN. This logs RESOLVER queries (everything a VPC asks, including what the resolver answers from cache or forwards on-prem) — a different surface from hosted-zone query logging, which records only what Route 53 answers for one public zone. Both the name and the destination are fixed for life at the provider, so the destination decision is the one to get right up front.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Resolver Query Log Configuration** — the logging pipeline pointing at your destination ARN. Immutable except tags: changing the name or destination replaces it (log data already written stays in the destination)
- **Query Log Config Associations** — one per `vpcIds` entry, turning logging on for that VPC. Associations are pure joins with no update path, and an association can flip to FAILED asynchronously after a clean apply if the resolver cannot write to the destination

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Route 53 Resolver permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- The destination, writable by the resolver (`destinationArn`): same-account CloudWatch log groups work with no extra setup; S3 buckets and Firehose delivery streams need their resolver-logging resource policies in place first. An association against an unwritable destination fails asynchronously minutes after the apply succeeds.
- The VPCs to log (`vpcIds`) — each VPC can associate to at most one query-log configuration per account.

## Deploy

### Console

Open the deployment store, find **AWS Route 53 Resolver Query Logging**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, the destination ARN, and the VPCs to log. Start from the **CloudWatch Logs Destination** preset or the **S3 Archive Destination** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53ResolverQueryLog
metadata:
  name: vpc-dns-queries
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  destinationArn:
    valueFrom:
      kind: AwsCloudwatchLogGroup
      name: dns-query-logs
      fieldPath: status.outputs.log_group_arn
  vpcIds:
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.vpc_id
```

```shell
planton apply -f resolver-query-log.yaml
```

This creates a logging configuration writing the app VPC's resolver queries to the referenced CloudWatch log group. An Infra Job tracks the provisioning in real time.

### InfraChart

When query logging deploys alongside its destination and VPC in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  destinationArn:
    valueFrom:
      kind: AwsCloudwatchLogGroup
      name: dns-query-logs
      fieldPath: status.outputs.log_group_arn
  vpcIds:
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys the log group and VPC first, then wires logging between them.

## Key Configuration

These are the most important decisions when configuring resolver query logging. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pick the destination by retention economics** — CloudWatch Logs for interactive queries and alarms, S3 for long-term archival queryable with Athena, Firehose for streaming into SIEM pipelines. The cost driver is raw query volume times the destination's storage class: a busy VPC generates serious DNS query traffic, and its logs in CloudWatch can out-cost the workloads' own logs. Compliance retention belongs in S3.

**The destination is a one-way door** — `destinationArn` (and the configuration's name) are fixed for life; changing either replaces the configuration. The replacement is safe for data — logs already written stay in the destination — but it costs continuity: a gap during the swap. Existing history is never migrated.

**A clean apply is not a working pipeline** — association is asynchronous: AWS accepts it, then tries to write. An unwritable destination (missing S3 bucket policy, missing Firehose permissions) flips the association to FAILED minutes later with an error code. When logs do not appear, check the association's status, not just its existence.

**The only volume dial is which VPCs associate** — there is no query filtering: an associated VPC logs every query. Associate the VPCs under investigation or in compliance scope, not the whole estate by reflex.

**One configuration per VPC** — a VPC can associate to at most one query-log configuration per account; a second configuration's association for the same VPC fails. Keep one configuration per destination strategy and associate it wide, rather than one per team.

**Resolver logging is not zone logging** — this records what your VPCs ASK. Hosted-zone query logging, configured on the Route 53 zone, records what Route 53 ANSWERS for one public zone. Auditing workload DNS behavior needs this component; auditing a public zone's traffic needs the other.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCloudwatchLogGroup** | `destinationArn` | `status.outputs.log_group_arn` |
| **AwsVpc** | `vpcIds` | `status.outputs.vpc_id` |

S3 bucket and Kinesis Data Firehose destinations travel as literal ARNs in `destinationArn`.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `query_log_config_id` | The configuration's id (`rqlc-...`) | Addressing the configuration in AWS tooling; the provider's import ID |
| `query_log_config_arn` | The configuration's ARN | IAM policies scoped to this configuration |

`share_status` and `association_ids` are also exported, but they are operational echoes — the RAM sharing state and the AWS-generated association IDs keyed by VPC — kept for audits and imports rather than composition.

## Common Patterns

**Interactive visibility in CloudWatch** — queries land in a log group for Logs Insights and alarms; same-account log groups need no extra permissions. The shape for active investigations and DNS Firewall tuning. Start from the **CloudWatch Logs Destination** preset.

**Compliance archive in S3** — long-term retention at S3 storage economics, queryable with Athena. The bucket needs the resolver-logging bucket policy before the association lands, or it fails asynchronously. Start from the **S3 Archive Destination** preset.

**Firewall tuning loop** — pair query logging with a DNS Firewall rule group running new rules as ALERT: the logs carry each query's firewall verdict, which is the evidence for flipping a rule to BLOCK. The firewall itself logs nothing, so without this component ALERT actions are invisible.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) — the VPCs whose resolver queries are logged, wired via `vpcIds`
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the interactive destination, wired via `destinationArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the archival destination, passed as a literal bucket ARN in `destinationArn`
- [**AWS Route 53 Resolver DNS Firewall**](/infra-catalog/aws-route53-resolver-firewall) — its rule verdicts appear in these logs; the tuning loop for ALERT-first policies
- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) — the other logging surface: hosted-zone query logging for what Route 53 answers publicly
