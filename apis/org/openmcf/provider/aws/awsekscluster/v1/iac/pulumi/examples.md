# Create using CLI

Create a YAML file using the examples shown below. After the YAML file is created, use the following command to apply:

# Set credentials
To configure AWS credentials for your Pulumi project, you can use one of the following methods:

- Method 1: Using Pulumi Configuration
```shell
pulumi config set aws:region us-east-1
pulumi config set aws:accessKeyId <access-key>
pulumi config set aws:secretAccessKey <secret-key>
```
- Method 2: Using AWS Configuration Files
  Alternatively, you can configure the AWS region and credentials in the AWS configuration files.
```shell
# ~/.aws/config
[default]
region = us-east-1

# ~/.aws/credentials
[default]
aws_access_key_id = ***************
aws_secret_access_key = ***************
```


```shell
planton apply -f <yaml-path>
```

# Basic Example

This basic example creates an AWS AWS EKS Cluster with default settings.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: my-basic-aws-eks-cluster
spec:
  region: us-west-2
  workersManagementMode: MANAGED
```

# Example with Existing Network

This example creates an AWS EKS Cluster in an existing VPC.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: my-aws-eks-cluster-with-vpc
spec:
  region: us-east-1
  vpcId: "vpc-0123456789abcdef0"
  securityGroups: 
    - "sg-01234455jkasd"
  subnets:
    - "subnet-1234"
    - "subnet-4567"
    - "subnet-2024"
  roleArn: "arn:aws:iam::123456789012:role/EKSClusterRole"
  nodeRoleArn: "arn:aws:iam::123456789012:role/EKSNodeRole"
  instanceType: "t3.medium"
  desiredSize: 2 # Number of worker nodes
  maxSize: 10 # Maximum number of worker nodes
  minSize: 2 # Minimum number of worker nodes

```

# Example with Environment Variables

This example uses environment variables to parameterize the AWS EKS Cluster configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: my-env-aws-eks-cluster
spec:
  region: ${AWS_REGION}
  workersManagementMode: ${WORKERS_MANAGEMENT_MODE}
```

In this example, replace the placeholders like `${AWS_CREDENTIAL_ID}` with your actual environment variable names or values.

# Example with Environment Secrets

The below example assumes that the secrets are managed by Planton Cloud's [AWS Secrets Manager](https://buf.build/openmcf/apis/docs/main:ai.planton.code2cloud.v1.aws.awssecretsmanager) deployment module.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: my-secret-aws-eks-cluster
spec:
  region: us-west-2
  workersManagementMode: MANAGED
  someSecretConfig: ${awssm-my-org-prod-aws-secrets.secret-key}
```

In this example:

- **someSecretConfig** is a placeholder for any configuration value you want to retrieve from AWS Secrets Manager.
- The value before the dot (`awssm-my-org-prod-aws-secrets`) is the ID of the AWS Secrets Manager resource on Planton Cloud.
- The value after the dot (`secret-key`) is the name of the secret within that resource.

---

These examples illustrate how to define an AWS EKS Cluster using the `AwsEksCluster` API resource. Since the `spec` is minimal, we have provided a few examples to demonstrate how to specify different configurations.

Please ensure that you replace placeholder values like `vpc-0123456789abcdef0`, environment variable names, and secret references with your actual configuration details.
