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

Before Planton can deploy infrastructure in your AWS account -- VPCs, ECS clusters, RDS databases, S3 buckets, or anything else -- it needs a secure way to authenticate with AWS on your behalf. This tutorial walks you through creating that connection using three different approaches, keyless first, so you can choose the one that fits your security requirements.

## What You Will Learn

- What a provider connection is and why Planton needs one
- How to create a keyless AWS Provider Connection in the web console, with no stored credential (recommended)
- How to store AWS credentials as organization-level secrets in Planton
- How to create an AWS Provider Connection with a stored access key
- How to create an AWS Provider Connection with a runner's own identity (maximum isolation)
- How to set a default provider connection so your deployments use it automatically

## Prerequisites

- [ ] An AWS account where you have permission to create IAM identity providers, roles and policies (or CloudFormation stacks), IAM users, or IAM roles for service accounts
- [ ] A Planton account with an organization already created
- [ ] The `planton` CLI installed and authenticated (run `planton auth login` if you have not done this yet) for Paths B and C

## Understanding Provider Connections

A provider connection tells Planton how to authenticate when deploying infrastructure in your cloud account, and what Planton may deploy there. Credentials, when a connection holds any, are stored as secrets in Planton's secrets manager and referenced by slug in the connection manifest -- the actual values never appear in the spec. For more on provider connections, see the [connections documentation](/docs/connections/cloud-providers).

Every connection records a **grant**: what Planton may deploy through it. **Everything Planton Can Deploy** covers every AWS kind in Planton's catalog; **Only What I Pick** covers the kinds you choose. The permissions come from the catalog, so the identity Planton uses holds exactly what those kinds need, never a broad policy like `AdministratorAccess`.

### Three Authentication Approaches for AWS

This tutorial covers three approaches. Choose the one that matches your needs:

| | Keyless | Stored Key | Runner Identity |
|---|---|---|---|
| **How it works** | An IAM role in your account trusts Planton's identity issuer; Planton exchanges a short-lived token for that role's credentials on each operation | You create an IAM user, store its access key as a Planton secret, and the connection references that secret | A self-hosted Planton Runner uses its own environment credentials (IRSA, instance profile, ECS task role) |
| **Security** | Nothing stored in Planton; credentials last at most an hour; the trust admits only this connection | Access key is long-lived; you manage rotation | Credentials never leave your infrastructure; nothing stored in Planton |
| **Setup time** | About five minutes (one script or one CloudFormation stack) | 5-10 minutes | 30+ minutes (requires deploying a runner) |
| **Best for** | Production and every account where you can create an IAM role | Quick evaluation, accounts where you cannot create an identity provider | Air-gapped environments, regulated industries, teams that already run a runner |

All three approaches result in the same outcome: a provider connection that Planton can use to deploy AWS infrastructure on your behalf.

## Path A: Keyless, in the Web Console

This approach stores nothing in Planton. You create an IAM role in your account that trusts Planton's identity issuer for this one connection, and Planton asks AWS for short-lived credentials whenever it works in your account. The console fills in every value for you; the connection is created only once AWS confirms the role trusts it.

### Step 1: Start the Connection

In the Planton web console, open **Connections** and click the **AWS** card under Infrastructure. **Name** the connection, for example `prod-aws`, then choose **Keyless** (listed first, and recommended).

### Step 2: Say Where It Deploys

On the **Account** step, enter your 12-digit **AWS Account ID** and the **Default Region**.

### Step 3: Choose What Planton May Deploy

On the **Permissions** step, keep **Everything Planton Can Deploy** (recommended), or choose **Only What I Pick** and tick the kinds you use, shelved by AWS's own service groups. **Organization Access** (AWS Organizations: its accounts and organizational units) and **Access Granting** (IAM users, roles, policies and keys across the account) stay off unless you turn them on. Click **Show the Policies** to read exactly what the role will hold.

### Step 4: Set Up the Trust in Your Account

The **Set Up** step shows a setup script with every value filled in. Run it with the AWS CLI signed in to the account you entered (AWS CloudShell works). It runs three steps, announcing each as it starts:

1. Registers Planton's issuer, `planton.ai`, as an IAM OIDC identity provider in your account, or keeps it when another Planton connection already did.
2. Creates this connection's own role, `planton-<your organization>-prod-aws`, whose trust policy accepts only the tokens Planton signs for this connection.
3. Writes the policies rendered from your grant onto the role.

If your changes go through CloudFormation, click **Prefer CloudFormation? Show the Template** instead, download the template, and in the CloudFormation console choose **Create stack**, **Upload a template file**. It creates exactly what the script does; run one of the two.

Then click **Check the Trust**. Planton asks AWS whether the role trusts this connection, and waits while AWS accepts the new role. Once AWS says yes, click **Next**.

If someone else administers the account, click **Create It Now, Finish the Trust Later** and hand them the script: the connection's page keeps the same setup and its check until the trust works.

### Step 5: Review, Create, and Read the Verdict

On **Review & Create**, click **Create**. The next screen runs the connection's first Verify and says what it can deploy, as **"N of M AWS Kinds Can Deploy"** counted against your grant. The connection stores only the role's ARN, `arn:aws:iam::123456789012:role/planton-<your organization>-prod-aws`; there is no key, no external ID, and nothing to rotate.

For what the script does in detail, running it again, and a second connection in the same account, see [Keyless Cloud Connections](/docs/connections/keyless-cloud-connections#aws).

## Path B: Stored Key

This approach stores a long-lived credential. You create an IAM user with programmatic access, store its credentials as Planton secrets, and create a connection that references them.

### Step 1: Create an IAM User in AWS

Sign in to the AWS Console and navigate to **IAM > Users > Create user**.

Create a user with programmatic access (access key, no console access). Attach the policies Planton's catalog renders for what you want it to deploy: in the console's AWS connection wizard, choose **Stored Key**, and on its **Permissions** step choose **Everything Planton Can Deploy** or **Only What I Pick** and click **Show the Policies**. They hold exactly what those kinds need; you do not need `AdministratorAccess`.

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
  grant:
    coverage: everything
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
- **`spec.grant`**: What Planton may deploy through the connection: `coverage: everything` for every kind Planton's catalog offers, or `coverage: chosen` with the `kinds` you pick. Every cloud connection records one, and the policies you attached in Step 1 should match it.

A few things to note about the field structure:

- Fields like `account_id` and `region` accept either a direct **`value`** (as shown above) or a **`variable`** reference that points to an org-level variable in config-manager. Using variables is useful when multiple connections share the same account ID or region.
- Fields like `access_key_id` and `secret_access_key` always use a **`secret`** reference that points to an org-level secret in config-manager. The actual credential value never appears in the YAML.
- **`auth_mode: inline`** is the default and can be omitted, but including it makes the intent explicit.

### Step 5: Apply the Manifest

```bash
planton apply -f aws-connection.yaml
```

Planton validates that the referenced secrets exist in your organization's config-manager and creates the connection. A manifest without `grant` is refused with a message that says how to add it.

### Step 6: Verify the Connection

Retrieve the connection to confirm it was created:

```bash
planton get AwsProviderConnection my-aws-connection
```

You should see the connection metadata and spec. The `access_key_id` and `secret_access_key` fields will display only the secret slug, not the actual credential values -- confirming that no sensitive data is stored in the connection resource itself.

## Path C: Runner Identity

This approach provides the highest security posture. Instead of storing any credentials in Planton -- even as secret references -- you deploy a Planton Runner in your own infrastructure and let it authenticate using the AWS SDK's default credential chain. The runner can use IAM Roles for Service Accounts (IRSA) on EKS, an EC2 instance profile, an ECS task role, or environment variables. No sensitive values are ever transmitted to or stored in the Planton control plane.

This approach requires a running Planton Runner. If you do not have one yet, the steps below walk you through enrollment.

### Step 1: Create a Runner Token

Create a runner token from the console's Runners area. The token is a named, revocable secret that authorizes runners to join your organization — it is never a runner's identity. A runner started with the token enrolls itself with the control plane on arrival and receives its own identity (a dedicated service account, an API key, and the mutual-TLS material for the runner-to-control-plane tunnel), stored locally on the machine where the runner runs.

### Step 2: Deploy the Runner with AWS IAM Bindings

Deploy the runner binary or container in your AWS environment. The runner's identity needs the policies Planton's catalog renders for what you want it to deploy: in the console's AWS connection wizard, choose **Runner Identity**, and on its **Permissions** step click **Show the Policies**.

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
  grant:
    coverage: everything
  auth_mode: runner
  runner: my-aws-runner
  account_id:
    value: "123456789012"
  region:
    value: us-west-2
```

The key differences from the stored-key manifest:

- **`auth_mode: runner`** tells Planton to delegate authentication entirely to the runner's environment.
- **`runner: my-aws-runner`** is the name the runner enrolled under in Step 3. This tells Planton which runner to route operations to.
- There are no `access_key_id` or `secret_access_key` fields. In runner mode, these fields are ignored even if provided.
- **`grant`** says what Planton may deploy through the connection, as in Path B.

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

To set a different default for a specific environment (for example, using the keyless connection for production):

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
