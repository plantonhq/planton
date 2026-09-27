---
title: "How to Connect a Container Registry to Planton"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "container-registry"
  - "connect"
  - "docker"
  - "ecr"
  - "artifact-registry"
  - "getting-started"
category: "connect"
excerpt: "Connect your container registry -- through a connection you already trust with nothing to paste, or with stored keys -- so Planton can push built images during CI, and give clusters the read-only login they pull with."
---

# How to Connect a Container Registry to Planton

When Planton builds your service from source code, the result is a container image that needs to be stored somewhere. Your organization brings its own registry -- Google Cloud Artifact Registry, AWS Elastic Container Registry, Azure Container Registry, JFrog Artifactory, or GitHub Container Registry -- and a Container Registry Connection describes how Planton signs in to push images during CI.

A registry connection signs in one of two ways, and this tutorial covers both. **The fastest path** names a connection you already trust -- the GitHub sign-in on your machine for GHCR, or the AWS, GCP, or Azure connection you already have -- and stores no key at all. **The stored-keys paths** reference a secret you create, for registries outside the connections you have.

## What You Will Learn

- The two ways a registry connection signs in, and when each fits
- How to connect a registry with nothing to paste, from the card on the Connections page or one CLI command
- How to store registry credentials as organization-level secrets and reference them, for the stored-keys paths
- How to prove a registry connection works the moment it exists

## Prerequisites

- [ ] A Planton organization -- on Planton Desktop, a hosted account, or a self-hosted instance
- [ ] The `planton` CLI installed and authenticated
- [ ] For the fastest path: a connection Planton already trusts that reaches your registry -- a GitHub connection made from the `gh` sign-in on your machine (with the `write:packages` scope) for GHCR, or an AWS, GCP, or Azure connection for ECR, Artifact Registry, or ACR
- [ ] For the stored-keys paths, one of the following:
  - **GCP Artifact Registry**: A GCP project with Artifact Registry enabled and a service account key with Artifact Registry Writer role
  - **AWS ECR**: An AWS account with an ECR repository and an IAM user with ECR push permissions
  - **GitHub Container Registry**: A GitHub account with a personal access token that has `write:packages` scope

## How Registry Connections Work

A container registry connection tells Planton where pipelines push the images they build and how to sign in. It never holds the result: on the trusted-connection arm it names another connection and the runner derives a short-lived registry token from it at push time; on the stored-keys arm it references a secret by slug. Pulling is the workload's own affair -- a Kubernetes workload declares its login on its own manifest, and the deploy fills that login from the registry connection when the connection holds a login a cluster can keep (a stored key or token, or GHCR's read-only pull token; never a token minted from a sign-in, and never anything for ECR, whose tokens last twelve hours). For more details, see the [container registry connections documentation](/docs/connections/container-registries).

### Supported Providers

| Provider | Trusted connection | Stored keys | Registry Hostname Pattern |
|---|---|---|---|
| GCP Artifact Registry | A GCP connection | Service account key | `<region>-docker.pkg.dev` |
| AWS ECR | An AWS connection | IAM access keys | `<account-id>.dkr.ecr.<region>.amazonaws.com` |
| Azure Container Registry | An Azure connection | Service principal | `<name>.azurecr.io` |
| GitHub Container Registry | A GitHub connection (the sign-in on your machine) | Personal access token | `ghcr.io` |
| JFrog Artifactory | -- | Access token | your Artifactory host |

## The Fastest Path: A Connection You Already Trust

If your organization already has a connection that reaches your registry, this is the whole tutorial.

### From the Connections Page

Open **Connections** and click **Container Registry**. The page opens on **You Already Trust These** -- one card per registry your connections reach:

- *GHCR as priya-dev · via GitHub connection github-account-priya-dev* -- **Use This Registry**
- *ECR us-east-1 for 123456789012 · via AWS connection aws-profile-acme-dev* -- **Use This Registry**
- *Artifact Registry in acme-dev · via GCP connection gcp-configuration-acme-dev -- needs a region and a repository* -- fill both on the card, then **Use This Registry**

Click **Use This Registry**. Planton writes a connection that stores nothing -- the registry's coordinates and the connection it trusts -- then proves it: the runner that would push derives the credential once and asks the registry to accept it. You land on the connection's page reading **Confirmed -- ghcr.io accepted the credential derived from github-account-priya-dev**, or **Saved, but Not Yet Working** with the exact sentence to act on (for a GitHub sign-in without `write:packages`, the `gh auth refresh -s write:packages` to run).

### From the CLI

```bash
# Which registries do my connections reach?
planton connect registry detect

# GHCR through the GitHub sign-in on this machine
planton connect registry detect --from github-account-priya-dev --yes

# ECR in a region through an AWS connection
planton connect registry detect --from aws-profile-acme-dev --region us-east-1 --yes
```

The verb writes the same connection the card writes and reads **Registry Connection Ready**. Name the connection in a service's build (`spec.build.registry`) to push there.

If no connection reaches your registry yet, connect GitHub or the cloud first -- its registries then appear here -- or continue with a stored-keys path below.

## The Stored-Keys Paths

Choose the path that matches your registry provider. Each stores the registry's credential as an organization secret and references it from the connection by slug.

### Path A: GCP Artifact Registry (stored keys)

#### Step 1: Create a GCP Service Account

In your GCP project, create a service account with the **Artifact Registry Writer** role. Builds push with it; because a stored key is a login a cluster can keep, a Kubernetes deploy also fills the workload's pull login from it, by reference.

```bash
gcloud iam service-accounts create planton-registry \
  --display-name="Planton Container Registry"

gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:planton-registry@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"
```

Download the service account key as a JSON file:

```bash
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=planton-registry@YOUR_PROJECT_ID.iam.gserviceaccount.com
```

#### Step 2: Store the Service Account Key as a Planton Secret

Store the JSON key file as one single-value, organization-level secret, read from standard input with `--string`:

```bash
planton secret set gcr-sa-key --string < ./sa-key.json
```

After storing the secret, delete the local key file:

```bash
rm sa-key.json
```

#### Step 3: Create the Container Registry Connection

Create a file named `container-registry.yaml`:

```yaml
apiVersion: connect.planton.ai/v1
kind: ContainerRegistryConnection
metadata:
  name: gcp-registry
  org: your-org
spec:
  provider: gcp_artifact_registry
  gcp_artifact_registry:
    gcp_project_id:
      value: "your-gcp-project-id"
    gcp_region:
      value: "us-central1"
    service_account_key:
      secret: gcr-sa-key
    gcp_artifact_registry_repo_name:
      value: "platform-images"
```

Replace the placeholder values:

- `metadata.name`: A descriptive name for this connection (becomes the slug you reference in Services)
- `metadata.org`: Your Planton organization slug
- `gcp_project_id.value`: Your GCP project ID
- `gcp_region.value`: The region where your Artifact Registry repository is located
- `service_account_key.secret`: The slug of the secret you created in Step 2
- `gcp_artifact_registry_repo_name.value`: The name of your Docker repository in Artifact Registry

#### Step 4: Apply the Connection

```bash
planton apply -f container-registry.yaml
```

### Path B: AWS Elastic Container Registry (stored keys)

#### Step 1: Create an IAM User with ECR Permissions

In your AWS account, create an IAM user with programmatic access and attach a policy that grants ECR permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchGetImage",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchCheckLayerAvailability",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": "*"
    }
  ]
}
```

Note the **Access Key ID** and **Secret Access Key** from the user creation.

#### Step 2: Store Credentials as Planton Secrets

```bash
planton secret set ecr-access-key '<access-key-id>'
planton secret set ecr-secret-key '<secret-access-key>'
```

#### Step 3: Create the Container Registry Connection

Create a file named `container-registry.yaml`:

```yaml
apiVersion: connect.planton.ai/v1
kind: ContainerRegistryConnection
metadata:
  name: aws-registry
  org: your-org
spec:
  provider: aws_elastic_container_registry
  aws_elastic_container_registry:
    account_id:
      value: "123456789012"
    access_key_id:
      secret: ecr-access-key
    secret_access_key:
      secret: ecr-secret-key
    region:
      value: "us-east-1"
```

Replace the placeholder values:

- `metadata.name`: A descriptive name for this connection
- `metadata.org`: Your Planton organization slug
- `account_id.value`: Your 12-digit AWS account ID
- `access_key_id.secret`: The slug of the access key secret
- `secret_access_key.secret`: The slug of the secret key secret
- `region.value`: The AWS region where your ECR repository is located

#### Step 4: Apply the Connection

```bash
planton apply -f container-registry.yaml
```

### Path C: GitHub Container Registry (stored keys)

#### Step 1: Create a Personal Access Token

> **Tip:** If the machine running Planton is signed in with `gh`, skip this path -- the fastest path above connects GHCR through that sign-in with nothing to paste.

In GitHub, navigate to **Settings > Developer settings > Personal access tokens** and create a token with:

- **`write:packages`** scope (includes `read:packages`)
- **`delete:packages`** scope (optional, if you want Planton to clean up images)

Note the token value. You will not be able to see it again after creation.

#### Step 2: Store the Token as a Planton Secret

```bash
planton secret set ghcr-pat '<personal-access-token>'
```

#### Step 3: Create the Container Registry Connection

Create a file named `container-registry.yaml`:

```yaml
apiVersion: connect.planton.ai/v1
kind: ContainerRegistryConnection
metadata:
  name: github-registry
  org: your-org
spec:
  provider: github_container_registry
  github_container_registry:
    github_username:
      value: "your-github-username"
    personal_access_token:
      secret: ghcr-pat
```

Replace the placeholder values:

- `metadata.name`: A descriptive name for this connection
- `metadata.org`: Your Planton organization slug
- `github_username.value`: Your GitHub username or organization name
- `personal_access_token.secret`: The slug of the secret you created in Step 2

#### Step 4: Apply the Connection

```bash
planton apply -f container-registry.yaml
```

#### Step 5: Give Clusters a Read-Only Pull Token

Whichever GHCR path you took, a Kubernetes cluster needs a login it can keep: a build signs in once, but a pod signs in every time it starts. A token minted from your GitHub sign-in lasts an hour and is never copied into a cluster, so on the trusted path a deploy fills no pull login until you add one -- and on the stored-keys path the cluster would otherwise hold your write-capable token.

Create a personal access token on a bot account with only the `read:packages` scope, store it as an organization secret (say `github-ghcr-pull-token`), and add it to the connection -- in the wizard's **Pull Token** step, the connection page's **Pull Token** section, or the record itself:

```yaml
spec:
  github_container_registry:
    github_connection: github-account-priya-dev   # or the stored PAT fields
    pull_token:
      username:
        value: acme-pull-bot
      token:
        secret: github-ghcr-pull-token
```

From then on a Kubernetes deploy writes that login onto the workload's `pod.imageRegistries` as a reference, and the run says so: *Pull login for ghcr.io filled from registry connection 'ghcr-account-priya-dev' (the connection's pull token)*. Clusters that pull from their own cloud's registry -- EKS from ECR, GKE from Artifact Registry, AKS from ACR -- need nothing of the kind.

## Verifying Your Connection

Every registry connection -- trusted or stored -- has **Verify This Registry** on its page. Press it and the runner that would push derives the credential right now and asks the registry to accept it. You read **Confirmed** with the registry host and the credential's source, or **Registry Not Working Yet** with the platform's own sentence naming the remedy for your provider. Nothing usable as a credential reaches your browser.

From the CLI, the connection's record shows the provider, the coordinates, and either the trusted connection or the secret slugs -- never a secret value:

```bash
planton get container-registry-connection ghcr-account-priya-dev
```

## What to Do Next

- **Connect your GitHub repositories** to Planton if you have not already: [How to Connect Your GitHub Account to Planton](/tutorials/how-to-connect-your-github-account-to-planton)
- **Deploy your first backend service** with zero-config CI/CD: [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd)
