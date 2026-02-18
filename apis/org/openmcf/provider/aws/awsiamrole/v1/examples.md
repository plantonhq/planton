# AwsIamRole Examples

## Minimal manifest: Lambda execution role

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: lambda-execution-role
  org: my-org
spec:
  region: us-east-1
  description: "IAM role for Lambda function execution"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "lambda.amazonaws.com"
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
```

## ECS task role with inline policy

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: ecs-task-role
  org: my-org
  tags:
    app: web-api
    env: prod
spec:
  region: us-east-1
  description: "IAM role for ECS task execution"
  path: "/ecs/"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "ecs-tasks.amazonaws.com"
        Action: "sts:AssumeRole"
        Condition:
          StringEquals:
            "aws:SourceAccount": "123456789012"
          ArnLike:
            "aws:SourceArn": "arn:aws:ecs:us-east-1:123456789012:*"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
  inlinePolicies:
    applicationAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "dynamodb:GetItem"
            - "dynamodb:Query"
            - "dynamodb:Scan"
          Resource: "arn:aws:dynamodb:us-east-1:123456789012:table/app-data"
        - Effect: "Allow"
          Action:
            - "s3:GetObject"
          Resource: "arn:aws:s3:::my-config-bucket/*"
```

## EC2 instance role with SSM access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: ec2-instance-role
  org: my-org
spec:
  region: us-west-2
  description: "IAM role for EC2 instances with Systems Manager access"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "ec2.amazonaws.com"
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
    - "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
```

## Cross-account access role

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: cross-account-reader
  org: my-org
spec:
  region: us-east-1
  description: "Role for trusted external account to read S3 data"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          AWS: "arn:aws:iam::987654321098:root"
        Action: "sts:AssumeRole"
        Condition:
          StringEquals:
            "sts:ExternalId": "unique-external-id-12345"
  inlinePolicies:
    s3ReadAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "s3:GetObject"
            - "s3:ListBucket"
          Resource:
            - "arn:aws:s3:::shared-data-bucket"
            - "arn:aws:s3:::shared-data-bucket/*"
```

## Lambda with VPC and multiple services access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: lambda-vpc-role
  org: my-org
  tags:
    service: data-processor
    team: backend
spec:
  region: us-east-1
  description: "Lambda role with VPC access and multi-service permissions"
  path: "/service-roles/"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "lambda.amazonaws.com"
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
    - "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  inlinePolicies:
    customAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "s3:GetObject"
            - "s3:PutObject"
          Resource: "arn:aws:s3:::my-bucket/*"
        - Effect: "Allow"
          Action:
            - "sqs:SendMessage"
            - "sqs:ReceiveMessage"
            - "sqs:DeleteMessage"
          Resource: "arn:aws:sqs:us-east-1:123456789012:my-queue"
        - Effect: "Allow"
          Action:
            - "secretsmanager:GetSecretValue"
          Resource: "arn:aws:secretsmanager:us-east-1:123456789012:secret:app/config-*"
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./iam-role.yaml
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./iam-role.yaml --stack my-org/project/dev --module-dir apis/org/openmcf/provider/aws/awsiamrole/v1/iac/pulumi
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./iam-role.yaml --auto-approve
```

Get outputs:
```bash
openmcf pulumi stack output role_arn --stack my-org/project/dev
openmcf pulumi stack output role_name --stack my-org/project/dev
```

Note: Provider credentials (AWS access key, secret) are supplied via stack input, not in the spec.

## References

For more detailed examples and best practices, see:
- [Research documentation](docs/README.md) - Deep dive into IAM role deployment approaches
- [Terraform examples](iac/tf/examples.md) - Additional Terraform-specific examples
- [Pulumi examples](iac/pulumi/examples.md) - Pulumi-specific examples

