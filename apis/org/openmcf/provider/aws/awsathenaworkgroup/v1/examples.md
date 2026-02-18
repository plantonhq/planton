# AWS Athena Workgroup Examples

## 1. Minimal SQL Workgroup

A basic workgroup with query results stored in S3. All governance defaults apply
(configuration enforcement enabled, CloudWatch metrics enabled).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: analytics-team
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://my-athena-results/analytics-team/"
```

## 2. Cost-Controlled Workgroup

Workgroup with a 10 GB data scan limit per query to prevent runaway costs. SSE_S3
encryption is enforced via minimum encryption configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: data-science
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://data-science-results/queries/"
    encryptionOption: SSE_S3
  bytesScannedCutoffPerQuery: 10737418240
  enableMinimumEncryptionConfiguration: true
```

## 3. Production Workgroup with KMS Encryption

Production workgroup with SSE_KMS encryption using a customer-managed KMS key,
strict configuration enforcement, and a 50 GB scan limit.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: prod-analytics
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://prod-athena-results/queries/"
    encryptionOption: SSE_KMS
    kmsKeyArn: "arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"
  bytesScannedCutoffPerQuery: 53687091200
  enforceWorkgroupConfiguration: true
  publishCloudwatchMetricsEnabled: true
  enableMinimumEncryptionConfiguration: true
  selectedEngineVersion: "Athena engine version 3"
```

## 4. Production Workgroup with KMS Key Reference (valueFrom)

Same as above, but referencing a KMS key from another OpenMCF resource using
`valueFrom`. The platform resolves the ARN at deployment time and creates a
dependency edge.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: prod-analytics-ref
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://prod-athena-results/queries/"
    encryptionOption: SSE_KMS
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: analytics-encryption-key
        fieldPath: status.outputs.key_arn
  bytesScannedCutoffPerQuery: 53687091200
  enforceWorkgroupConfiguration: true
  enableMinimumEncryptionConfiguration: true
```

## 5. Cross-Account Workgroup

Workgroup writing results to a bucket owned by a different AWS account. The
bucket owner gets full control of result objects via the ACL setting.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: shared-analytics
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://central-data-lake/athena-results/"
    encryptionOption: SSE_S3
    expectedBucketOwner: "987654321098"
    s3AclOption: BUCKET_OWNER_FULL_CONTROL
  enforceWorkgroupConfiguration: true
```

## 6. Development Workgroup (Relaxed Governance)

Development workgroup with enforcement disabled so engineers can override settings
per query. No cost limits.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: dev-sandbox
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://dev-athena-results/sandbox/"
  enforceWorkgroupConfiguration: false
  forceDestroy: true
```

## 7. Spark Workgroup with Execution Role

Workgroup for Apache Spark on Athena, requiring an IAM execution role for PySpark
notebooks.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: spark-notebooks
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://spark-results/notebooks/"
    encryptionOption: SSE_S3
  executionRole:
    valueFrom:
      kind: AwsIamRole
      name: athena-spark-execution-role
      fieldPath: status.outputs.role_arn
  selectedEngineVersion: "PySpark engine version 3"
  forceDestroy: true
```
