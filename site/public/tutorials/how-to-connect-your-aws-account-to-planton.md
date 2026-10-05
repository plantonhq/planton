---
title: "How to Connect Your AWS Account to Planton"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "aws"
  - "connect"
  - "provider-connection"
  - "getting-started"
category: "connect"
excerpt: "Grant Planton secure access to deploy and manage resources in your AWS account."
---

# How to Connect Your AWS Account to Planton

Before Planton can deploy infrastructure in your AWS account -- VPCs, ECS clusters, RDS databases, S3 buckets, or anything else -- it needs a secure way to authenticate with AWS on your behalf. This tutorial walks you through creating that connection using three different approaches, so you can choose the one that fits your security requirements.

## What You Will Learn

- What a provider connection is and why Planton needs one
- How to store AWS credentials as organization-level secrets in Planton
- How to create an AWS Provider Connection with inline credentials (quickest path)
- How to create an AWS Provider Connection with cross-account trust via CloudFormation (recommended for production)
- How to create an AWS Provider Connection with runner-delegated authentication (maximum security)
- How to set a default provider connection so your deployments use it automatically

## Prerequisites

- [ ] An AWS account where you have permission to create IAM users, CloudFormation stacks, or IAM roles for service accounts
- [ ] A Planton account with an organization already created
- [ ] The `planton` CLI installed and authenticated (run `planton auth login` if you have not done this yet)

## Understanding Provider Connections

A provider connection tells Planton how to authenticate when deploying infrastructure in your cloud account. Credentials are stored as secrets in Planton's secrets manager and referenced by slug in the connection manifest -- the actual values never appear in the spec. For more on provider connections, see the [connections documentation](/docs/connections/cloud-providers).

### Three Authentication Approaches for AWS

This tutorial covers three approaches. Choose the one that matches your needs:

| | Inline Credentials | Cross-Account Trust | Runner-Delegated |
|---|---|---|---|
| **How it works** | You create an IAM user, store its access key as a Planton secret, and the connection references that secret | A CloudFormation stack creates an IAM role with a trust policy; Planton assumes the role via STS for temporary credentials | A self-hosted Planton Runner uses its own environment credentials (IRSA, instance profile, ECS task role) |
| **Security** | Access key is long-lived; you manage rotation | No long-lived credentials; temporary tokens auto-rotate via STS | Credentials never leave your infrastructure; nothing stored in Planton |
| **Setup time** | 5-10 minutes | 10-15 minutes | 30+ minutes (requires deploying a runner) |
| **Best for** | Getting started, development environments | Production, compliance-sensitive environments | Maximum security, air-gapped environments, regulated industries |

All three approaches result in the same outcome: a provider connection that Planton can use to deploy AWS infrastructure on your behalf.

## Path A: Inline Credentials

This is the quickest way to connect your AWS account. You create an IAM user with programmatic access, store its credentials as Planton secrets, and create a connection that references them.

### Step 1: Create an IAM User in AWS

Sign in to the AWS Console and navigate to **IAM > Users > Create user**.

Create a user with programmatic access (access key, no console access). Attach policies that grant the permissions Planton needs for the AWS services you plan to use. For a broad starting point, `AdministratorAccess` works, though you should scope this down for production use.

After creating the user, note the **Access Key ID** and **Secret Access Key**. You will need both in the next steps.

### Step 2: Store the Access Key ID as an Org-Level Secret

Use the `planton secret set` command to store the access key ID. The name you choose here (in this case, `aws-access-key-id`) is the slug you will reference in the connection manifest.

```bash
planton secret set aws-access-key-id 'AKIAIOSFODNN7EXAMPLE'
```

### Step 3: Store the Secret Access Key as an Org-Level Secret

Store the secret access key the same way. Use a distinct, descriptive name.

```bash
planton secret set aws-secret-access-key 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
```

After these two commands, your AWS credentials are stored securely in Planton's config-manager. The connection you create next will reference them by slug -- the actual key values never appear in the connection spec.

### Step 4: Write the Connection Manifest

Create a file named `aws-connection.yaml` with the following content:

```yaml
apiVersion: connect.planton.ai/v1
kind: AwsProviderConnection
metadata:
  name: my-aws-connection
  org: your-org
spec:
  auth_mode: inline
  account_id:
    value: "123456789012"
  region:
    value: us-west-2
  access_key_id:
    secret: aws-access-key-id
  secret_access_key:
    secret: aws-secret-access-key
```

Replace these values with your own:

- **`metadata.name`**: A human-readable name for this connection. Planton generates a URL-safe slug from it.
- **`metadata.org`**: Your Planton organization slug.
- **`spec.account_id.value`**: Your 12-digit AWS account ID (find it in the AWS Console under your account dropdown).
- **`spec.region.value`**: The AWS region you want as the default for this connection.
- **`spec.access_key_id.secret`**: The slug of the secret you created in Step 2.
- **`spec.secret_access_key.secret`**: The slug of the secret you created in Step 3.

A few things to note about the field structure:

- Fields like `account_id` and `region` accept either a direct **`value`** (as shown above) or a **`variable`** reference that points to an org-level variable in config-manager. Using variables is useful when multiple connections share the same account ID or region.
- Fields like `access_key_id` and `secret_access_key` always use a **`secret`** reference that points to an org-level secret in config-manager. The actual credential value never appears in the YAML.
- **`auth_mode: inline`** is the default and can be omitted, but including it makes the intent explicit.

### Step 5: Apply the Manifest

```bash
planton apply -f aws-connection.yaml
```

Planton validates that the referenced secrets exist in your organization's config-manager and creates the connection.

### Step 6: Verify the Connection

Retrieve the connection to confirm it was created:

```bash
planton get AwsProviderConnection my-aws-connection
```

You should see the connection metadata and spec. The `access_key_id` and `secret_access_key` fields will display only the secret slug, not the actual credential values -- confirming that no sensitive data is stored in the connection resource itself.

## Path B: Cross-Account Trust via CloudFormation

This approach creates an IAM role in your AWS account with a trust policy that allows Planton to assume it via STS `AssumeRole`. No long-lived credentials are stored anywhere -- Planton obtains temporary security tokens that automatically expire and rotate.

The `planton connect aws` command automates the entire setup through a CloudFormation Quick Create flow.

### Step 1: Run the Connect Command

```bash
planton connect aws \
  --name prod-aws \
  --org your-org \
  --account-id 123456789012
```

This command generates a CloudFormation Quick Create URL and opens it in your browser. It also generates a unique external ID (stored automatically as a config-manager secret) that is embedded in the IAM role's trust policy for confused-deputy prevention.

To scope the IAM role's permissions to specific AWS services, use the `--capabilities` flag:

```bash
planton connect aws \
  --name prod-aws \
  --org your-org \
  --account-id 123456789012 \
  --capabilities eks,vpc_networking,s3_storage,ecr_registry
```

Available capabilities: `eks`, `ecs`, `vpc_networking`, `s3_storage`, `rds_database`, `lambda_compute`, `route53_dns`, `cloudwatch_monitoring`, `ecr_registry`, `kms_encryption`. Each capability maps to a conditional IAM policy in the CloudFormation template, so the role only has permissions for the services you select.

### Step 2: Review and Create the CloudFormation Stack

Your browser opens to the AWS CloudFormation console with the stack parameters pre-filled. Review the IAM role that will be created, then click **Create stack**.

The CloudFormation stack creates:

- An IAM role with a trust policy scoped to Planton's runner identity
- An external ID condition on the trust policy for additional security
- Conditional IAM policies based on the capabilities you selected
- A Lambda function that notifies Planton when the stack completes

### Step 3: Wait for Completion

The CLI polls for the CloudFormation callback. When the Lambda function in the stack notifies Planton that the stack is complete, the CLI reports success and displays the connection details.

The resulting connection has:

- **`iam_role_arn`**: The ARN of the IAM role, stored as a plain string on the connection spec (role ARNs are not sensitive).
- **`external_id`**: A secret reference to the config-manager secret that was auto-generated during initiation.
- **`capabilities`**: The list of AWS service capabilities you selected.

### Step 4: Verify the Connection

```bash
planton get AwsProviderConnection prod-aws
```

The connection is ready to use. Planton's runner will assume this IAM role via STS whenever it needs to perform operations in your AWS account.

## Path C: Runner-Delegated Authentication

This approach provides the highest security posture. Instead of storing any credentials in Planton -- even as secret references -- you deploy a Planton Runner in your own infrastructure and let it authenticate using the AWS SDK's default credential chain. The runner can use IAM Roles for Service Accounts (IRSA) on EKS, an EC2 instance profile, an ECS task role, or environment variables. No sensitive values are ever transmitted to or stored in the Planton control plane.

This approach requires a running Planton Runner. If you do not have one yet, the steps below walk you through enrollment.

### Step 1: Create a Runner Token

Create a runner token from the console's Runners area. The token is a named, revocable secret that authorizes runners to join your organization — it is never a runner's identity. A runner started with the token enrolls itself with the control plane on arrival and receives its own identity (a dedicated service account, an API key, and the mutual-TLS material for the runner-to-control-plane tunnel), stored locally on the machine where the runner runs.

### Step 2: Deploy the Runner with AWS IAM Bindings

Deploy the runner binary or container in your AWS environment. The runner needs IAM permissions for the AWS services you plan to manage through Planton.

How you bind IAM permissions depends on where the runner runs:

- **EKS**: Create an IAM role and bind it to the runner's Kubernetes service account via IRSA (IAM Roles for Service Accounts). This is the most common pattern.
- **EC2**: Attach an instance profile with the required IAM policies to the EC2 instance running the runner.
- **ECS**: Assign a task IAM role with the required policies to the ECS task definition running the runner.

The runner uses the AWS SDK's default credential chain, so any method that makes credentials available to a standard AWS SDK client will work.

### Step 3: Start the Runner

Start the runner with your runner token and a name (for example, `my-aws-runner`). On first start it enrolls itself — the registration appears in your Runners list at that moment — then establishes an outbound mTLS tunnel to the Planton control plane. It is now ready to receive and execute operations.

### Step 4: Write the Connection Manifest

Create a file named `aws-runner-connection.yaml`:

```yaml
apiVersion: connect.planton.ai/v1
kind: AwsProviderConnection
metadata:
  name: aws-runner
  org: your-org
spec:
  auth_mode: runner
  runner: my-aws-runner
  account_id:
    value: "123456789012"
  region:
    value: us-west-2
```

The key differences from the inline manifest:

- **`auth_mode: runner`** tells Planton to delegate authentication entirely to the runner's environment.
- **`runner: my-aws-runner`** is the name the runner enrolled under in Step 3. This tells Planton which runner to route operations to.
- There are no `access_key_id` or `secret_access_key` fields. In runner mode, these fields are ignored even if provided.

### Step 5: Apply the Manifest

```bash
planton apply -f aws-runner-connection.yaml
```

### Step 6: Verify the Connection

```bash
planton get AwsProviderConnection aws-runner
```

The connection spec will show `auth_mode: runner` and the runner slug. No secret references appear because the runner handles all authentication locally.

## Setting the Default AWS Connection

Once you have a connection, set it as the default for your organization so that AWS deployments use it automatically without needing to specify a connection in every manifest.

```bash
planton connection default set --provider aws --connection my-aws-connection
```

To set a different default for a specific environment (for example, using the cross-account trust connection for production):

```bash
planton connection default set --provider aws --connection prod-aws --env production
```

When a deployment needs an AWS connection, Planton resolves it in this order:

1. **Explicit connection** specified in the resource manifest (highest priority)
2. **Environment-level default** for the target environment
3. **Organization-level default** (lowest priority)

If none of these exist, the deployment fails with a clear error.

Verify your defaults are set correctly:

```bash
planton connection default list
```

## What to Do Next

Your AWS account is now connected to Planton. From here, you can:

- **Deploy AWS infrastructure** through the Infra Catalog -- ECS clusters, RDS databases, VPCs, and more
- **Connect additional AWS accounts** by repeating this process (organizations commonly have separate connections for production and non-production accounts)
- **Scope connection access** by creating provider connection authorizations that control which connections are allowed in which environments

When more tutorials become available, look for guides on deploying a production-ready AWS ECS environment and deploying your first backend service with zero-config CI/CD.
