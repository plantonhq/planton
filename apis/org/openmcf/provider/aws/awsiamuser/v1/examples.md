# AwsIamUser Examples

## Minimal manifest: Basic IAM user with read-only access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: basic-iam-user
  org: my-org
spec:
  region: us-east-1
  userName: "basic-user"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/ReadOnlyAccess"
```

## CI/CD pipeline user with S3 and ECR access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: cicd-pipeline-user
  org: my-org
  tags:
    purpose: cicd
    team: devops
spec:
  region: us-east-1
  userName: "cicd-deployment-user"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/AmazonS3FullAccess"
    - "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"
  inlinePolicies:
    deploymentAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "ecs:DescribeServices"
            - "ecs:UpdateService"
            - "ecs:DescribeTasks"
            - "ecs:RunTask"
          Resource: "*"
        - Effect: "Allow"
          Action:
            - "cloudformation:DescribeStacks"
            - "cloudformation:UpdateStack"
          Resource: "arn:aws:cloudformation:*:*:stack/app-*"
```

## Application service account with DynamoDB and Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: app-service-account
  org: my-org
  tags:
    app: payment-processor
    env: prod
spec:
  region: us-east-1
  userName: "payment-app-user"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess"
    - "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
  inlinePolicies:
    applicationSpecific:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "secretsmanager:GetSecretValue"
          Resource: "arn:aws:secretsmanager:*:*:secret:app/payment-*"
        - Effect: "Allow"
          Action:
            - "kms:Decrypt"
          Resource: "arn:aws:kms:*:*:key/*"
          Condition:
            StringEquals:
              "kms:ViaService": "secretsmanager.*.amazonaws.com"
```

## User without access keys (identity-only)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: identity-only-user
  org: my-org
spec:
  region: us-east-1
  userName: "audit-user"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/ReadOnlyAccess"
  disableAccessKeys: true
```

## Development user with broad permissions

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: dev-user
  org: my-org
  tags:
    environment: development
    team: engineering
spec:
  region: us-west-2
  userName: "developer-sandbox-user"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/PowerUserAccess"
  inlinePolicies:
    developmentAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "ec2:*"
            - "rds:*"
            - "lambda:*"
            - "cloudformation:*"
          Resource: "*"
        - Effect: "Deny"
          Action:
            - "iam:*"
            - "organizations:*"
          Resource: "*"
```

## Monitoring and observability user

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: monitoring-user
  org: my-org
  tags:
    purpose: monitoring
spec:
  region: us-east-1
  userName: "datadog-integration-user"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/ReadOnlyAccess"
  inlinePolicies:
    cloudwatchAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "cloudwatch:GetMetricStatistics"
            - "cloudwatch:ListMetrics"
            - "logs:DescribeLogGroups"
            - "logs:DescribeLogStreams"
            - "logs:FilterLogEvents"
            - "tag:GetResources"
          Resource: "*"
```

## Backup operations user

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: backup-user
  org: my-org
  tags:
    purpose: backup
spec:
  region: us-west-2
  userName: "backup-service-user"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/AWSBackupFullAccess"
  inlinePolicies:
    s3BackupAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "s3:GetObject"
            - "s3:PutObject"
            - "s3:DeleteObject"
          Resource: "arn:aws:s3:::backup-bucket-*/*"
        - Effect: "Allow"
          Action:
            - "s3:ListBucket"
          Resource: "arn:aws:s3:::backup-bucket-*"
```

## Cross-account access user

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: cross-account-user
  org: my-org
spec:
  region: us-east-1
  userName: "cross-account-automation-user"
  inlinePolicies:
    assumeRoleAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "sts:AssumeRole"
          Resource:
            - "arn:aws:iam::987654321098:role/automation-role"
            - "arn:aws:iam::123456789012:role/deployment-role"
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./iam-user.yaml
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./iam-user.yaml --stack my-org/project/dev --module-dir apis/org/openmcf/provider/aws/awsiamuser/v1/iac/pulumi
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./iam-user.yaml --auto-approve
```

Get outputs (including sensitive access keys):
```bash
# Get user ARN
openmcf pulumi stack output user_arn --stack my-org/project/dev

# Get access key ID (not sensitive)
openmcf pulumi stack output access_key_id --stack my-org/project/dev

# Get secret access key (sensitive, base64-encoded)
openmcf pulumi stack output secret_access_key --stack my-org/project/dev --show-secrets
```

Decode secret access key:
```bash
# The secret is base64-encoded for safe transmission
SECRET=$(openmcf pulumi stack output secret_access_key --stack my-org/project/dev --show-secrets)
echo $SECRET | base64 -d
```

Store credentials securely:
```bash
# Recommended: Store immediately in AWS Secrets Manager
aws secretsmanager create-secret \
  --name "/cicd/aws-user-credentials" \
  --description "CI/CD user credentials" \
  --secret-string "{\"access_key_id\":\"AKIAIOSFODNN7EXAMPLE\",\"secret_access_key\":\"decoded-secret-here\"}"
```

Note: Provider credentials (AWS access key, secret, region) are supplied via stack input, not in the spec.

## Security best practices

1. **Immediately store access keys securely**: Never commit to Git or display in logs
2. **Rotate keys regularly**: Set up 90-day rotation schedule
3. **Use least privilege**: Grant only required permissions
4. **Monitor usage**: Enable CloudTrail and review Access Advisor
5. **Prefer alternatives**: Use IAM roles for AWS workloads, federation for humans
6. **Tag appropriately**: Use metadata tags for cost tracking and access control

## References

For more detailed examples and best practices, see:
- [Research documentation](docs/README.md) - Deep dive into IAM user deployment approaches
- [Terraform examples](iac/tf/examples.md) - Additional Terraform-specific examples
- [Pulumi examples](iac/pulumi/examples.md) - Pulumi-specific examples

