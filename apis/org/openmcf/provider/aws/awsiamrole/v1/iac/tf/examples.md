# AWS IAM Role Examples

Below are several examples demonstrating how to define an AWS IAM Role component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic IAM Role

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: basic-iam-role
spec:
  region: us-east-1
  description: "Basic IAM role for application access"
  path: "/"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "ec2.amazonaws.com"
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
```

This example creates a basic IAM role:
• Trust policy allowing EC2 instances to assume the role.
• Attached S3 read-only managed policy.
• Standard path and description.
• Suitable for EC2 instance profiles.

---

## IAM Role for Lambda Function

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: lambda-execution-role
spec:
  region: us-east-1
  description: "IAM role for Lambda function execution"
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
    customS3Access:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "s3:GetObject"
            - "s3:PutObject"
          Resource: "arn:aws:s3:::my-bucket/*"
```

This example creates a Lambda execution role:
• Trust policy for Lambda service.
• Basic execution and VPC access policies.
• Custom inline policy for S3 access.
• Service role path for organization.

---

## IAM Role for ECS Tasks

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: ecs-task-role
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
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
  inlinePolicies:
    applicationAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "dynamodb:GetItem"
            - "dynamodb:PutItem"
            - "dynamodb:Query"
            - "dynamodb:Scan"
          Resource: "arn:aws:dynamodb:*:*:table/my-table"
```

This example creates an ECS task role:
• Trust policy for ECS tasks.
• Task execution managed policy.
• Custom DynamoDB access policy.
• ECS-specific path organization.

---

## IAM Role for EKS Node Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: eks-node-role
spec:
  region: us-east-1
  description: "IAM role for EKS worker nodes"
  path: "/eks/"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "ec2.amazonaws.com"
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
    - "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
    - "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  inlinePolicies:
    nodeGroupAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "ec2:DescribeInstances"
            - "ec2:DescribeRegions"
          Resource: "*"
```

This example creates an EKS node group role:
• Trust policy for EC2 instances.
• Required EKS worker node policies.
• Custom inline policy for node access.
• EKS-specific path organization.

---

## IAM Role with Cross-Account Access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: cross-account-role
spec:
  region: us-east-1
  description: "IAM role for cross-account access"
  path: "/cross-account/"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          AWS: "arn:aws:iam::123456789012:root"
        Action: "sts:AssumeRole"
        Condition:
          StringEquals:
            "sts:ExternalId": "unique-external-id"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/ReadOnlyAccess"
  inlinePolicies:
    limitedAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "s3:ListBucket"
            - "s3:GetObject"
          Resource:
            - "arn:aws:s3:::shared-bucket"
            - "arn:aws:s3:::shared-bucket/*"
```

This example creates a cross-account role:
• Trust policy for specific AWS account.
• External ID condition for security.
• Read-only managed policy.
• Limited S3 access inline policy.

---

## IAM Role for Application Services

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: application-service-role
spec:
  region: us-east-1
  description: "IAM role for application services"
  path: "/applications/"
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "ecs-tasks.amazonaws.com"
        Action: "sts:AssumeRole"
      - Effect: "Allow"
        Principal:
          Service: "lambda.amazonaws.com"
        Action: "sts:AssumeRole"
  managedPolicyArns:
    - "arn:aws:iam::aws:policy/AmazonS3FullAccess"
    - "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess"
  inlinePolicies:
    applicationSpecific:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "secretsmanager:GetSecretValue"
            - "kms:Decrypt"
          Resource:
            - "arn:aws:secretsmanager:*:*:secret:app-*"
            - "arn:aws:kms:*:*:key/*"
```

This example creates a multi-service role:
• Trust policy for both ECS and Lambda.
• Full access to S3 and DynamoDB.
• Custom policy for secrets and KMS access.
• Application-specific path organization.

---

## IAM Role with Minimal Configuration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamRole
metadata:
  name: minimal-iam-role
spec:
  region: us-east-1
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Principal:
          Service: "ec2.amazonaws.com"
        Action: "sts:AssumeRole"
```

This example creates a minimal IAM role:
• Only required trust policy specified.
• Uses default path ("/").
• No managed or inline policies.
• Suitable for basic EC2 instance profiles.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the IAM role is active via the AWS console or by
using the AWS CLI:

```shell
aws iam get-role --role-name <your-role-name>
```

For detailed role information including attached policies:

```shell
aws iam list-attached-role-policies --role-name <your-role-name>
```

To list inline policies:

```shell
aws iam list-role-policies --role-name <your-role-name>
```

This will show the IAM role details including the trust policy, attached managed policies, and inline policies.

