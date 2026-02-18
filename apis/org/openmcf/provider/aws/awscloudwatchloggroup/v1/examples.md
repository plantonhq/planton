# AwsCloudwatchLogGroup Examples

## 1. Minimal — 30-Day Retention

The most common use case: a log group with a sensible retention period.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
  org: acme
  env: dev
  id: app-logs-dev
spec:
  region: us-west-2
  retentionInDays: 30
```

**What this creates:** A STANDARD class log group with 30-day retention. Log events older than 30 days are automatically deleted.

---

## 2. Never-Expire Retention

For audit logs or compliance data that must be retained indefinitely:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: audit-trail
  org: acme
  env: prod
  id: audit-trail-prod
spec:
  region: us-west-2
```

**What this creates:** A STANDARD class log group with indefinite retention (retention_in_days defaults to 0). Log events are never automatically deleted.

---

## 3. KMS-Encrypted Production Log Group

Encrypt log data at rest with a customer-managed KMS key:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: prod-app-logs
  org: acme
  env: prod
  id: prod-app-logs-prod
spec:
  region: us-west-2
  retentionInDays: 90
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: log-encryption-key
      fieldPath: status.outputs.key_arn
```

**What this creates:** A STANDARD class log group with 90-day retention and KMS encryption. The KMS key is referenced from an AwsKmsKey resource deployed in the same environment.

---

## 4. Infrequent Access — High-Volume, Low-Query Logs

For VPC flow logs, CDN access logs, or other high-volume data accessed infrequently:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: vpc-flow-logs
  org: acme
  env: prod
  id: vpc-flow-logs-prod
spec:
  region: us-west-2
  retentionInDays: 365
  logGroupClass: INFREQUENT_ACCESS
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: infra-key
      fieldPath: status.outputs.key_arn
```

**What this creates:** An INFREQUENT_ACCESS class log group (~50% cheaper storage) with 1-year retention and KMS encryption. Note: INFREQUENT_ACCESS does not support metric filters or subscription filters.

---

## 5. Step Functions Logging Destination

Create a log group for Step Functions execution logging, then wire it via `valueFrom`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: sfn-execution-logs
  org: acme
  env: prod
  id: sfn-execution-logs-prod
spec:
  region: us-west-2
  retentionInDays: 30
---
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: order-processor
  org: acme
  env: prod
  id: order-processor-prod
spec:
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sfn-role
      fieldPath: status.outputs.role_arn
  definition:
    StartAt: ProcessOrder
    States:
      ProcessOrder:
        Type: Pass
        End: true
  logging:
    level: ERROR
    logDestination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: sfn-execution-logs
        fieldPath: status.outputs.log_group_arn
```

---

## 6. API Gateway Access Logging

Create a log group for HTTP API Gateway access logs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: api-access-logs
  org: acme
  env: prod
  id: api-access-logs-prod
spec:
  region: us-west-2
  retentionInDays: 60
---
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: public-api
  org: acme
  env: prod
  id: public-api-prod
spec:
  stage:
    accessLog:
      destinationArn:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: api-access-logs
          fieldPath: status.outputs.log_group_arn
      format: '{"requestId":"$context.requestId","ip":"$context.identity.sourceIp","method":"$context.httpMethod","status":"$context.status"}'
  routes:
    - name: default
      routeKey: "$default"
      integration:
        integrationType: AWS_PROXY
        integrationUri:
          valueFrom:
            kind: AwsLambda
            name: api-handler
            fieldPath: status.outputs.invoke_arn
```

---

## 7. Delivery Class — AWS Service Log Delivery

For logs from VPC Flow Logs, CloudTrail, or Route53 Resolver at the lowest cost:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: cloudtrail-delivery
  org: acme
  env: prod
  id: cloudtrail-delivery-prod
spec:
  region: us-west-2
  logGroupClass: DELIVERY
```

**What this creates:** A DELIVERY class log group optimized for AWS service log delivery. Retention is managed by AWS — do not set `retentionInDays`.
