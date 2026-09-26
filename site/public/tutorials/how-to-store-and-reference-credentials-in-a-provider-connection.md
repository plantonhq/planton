---
title: "How to Store and Reference Credentials in a Provider Connection"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "secrets"
  - "variables"
  - "provider-connection"
  - "credentials"
category: "platform"
excerpt: "Store your cloud credentials as Planton secrets, create a shared variable for your default region, and wire both into a provider connection manifest."
---

# How to Store and Reference Credentials in a Provider Connection

Planton provider connections never contain raw credentials. Instead, sensitive values like access keys are stored as **secrets** and non-sensitive shared values like regions are stored as **variables** -- and your connection manifest references them by slug. This tutorial walks you through the complete workflow: storing credentials, creating a variable, wiring both into an AWS provider connection manifest, applying it, and rotating a credential when it changes.

> **Note**: If you have already completed the [AWS connection tutorial](/tutorials/how-to-connect-your-aws-account-to-planton), you have seen secrets used in context. This tutorial focuses specifically on the secrets and variables workflow -- how to create, reference, verify, and rotate them.

## What You Will Learn

- How to store a credential as an org-level secret using the Planton CLI
- How to create a shared variable for non-sensitive configuration
- How the `secret` and `variable` reference fields work in YAML manifests
- How to rotate a secret and verify the new version is active

## Prerequisites

- [ ] A Planton account with an organization already created
- [ ] The `planton` CLI installed and authenticated (`planton auth login`)
- [ ] An AWS access key ID and secret access key (or any cloud credential you want to store)

## How Secrets and Variables Work

Planton's secrets manager stores sensitive values (credentials, tokens, private keys) encrypted at rest with automatic versioning. Variables store non-sensitive configuration values (regions, project IDs, account identifiers) that multiple resources can share. Both are referenced by slug in YAML manifests -- the platform resolves them at execution time. For more on secrets, variables, and available backends, see the [secrets documentation](/docs/secrets/managing-secrets) and [variables documentation](/docs/variables).

## Step 1: Store Your Credentials as Secrets

Use `planton secret set` to store each credential. The slug you choose becomes the reference you use in manifests.

Store the access key ID:

```bash
planton secret set aws-access-key-id value=AKIAIOSFODNN7EXAMPLE
```

Store the secret access key:

```bash
planton secret set aws-secret-access-key value=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

Each command creates a versioned, encrypted secret in your organization's secrets manager. The slug (`aws-access-key-id`, `aws-secret-access-key`) is what you will reference in manifests -- the actual values are never exposed outside the secrets manager.

Verify the secrets were created:

```bash
planton secret get aws-access-key-id
planton secret get aws-secret-access-key
```

The output shows the secret's record (slug, scope, backend, creation time) but not the value itself: reading a record never decrypts the secret. Secret values are decrypted only at execution time by the platform, or when you ask for one with `--reveal`.

## Step 2: Create a Shared Variable

Variables are for non-sensitive values that multiple resources can share. Instead of hardcoding `us-west-2` in every manifest, store it as a variable:

```bash
planton variable set default-aws-region us-west-2
```

Now any resource manifest can reference `default-aws-region` instead of repeating the region string. If your organization moves to a different default region, you update the variable once and every resource that references it picks up the new value on its next execution.

Verify the variable:

```bash
planton variable get default-aws-region
```

## Step 3: Write a Connection Manifest Using References

Create a file named `aws-connection.yaml`. This manifest demonstrates both reference types:

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
    variable: default-aws-region
  access_key_id:
    secret: aws-access-key-id
  secret_access_key:
    secret: aws-secret-access-key
```

Three reference patterns appear in this manifest:

- **`account_id.value`**: A literal string. The AWS account ID is not sensitive and does not change across environments, so a direct value is appropriate.
- **`region.variable`**: A variable reference. The platform resolves the slug `default-aws-region` to the value you stored in Step 2 (`us-west-2`). This is useful when multiple connections or resources share the same region.
- **`access_key_id.secret` / `secret_access_key.secret`**: Secret references. The platform resolves these slugs to the encrypted values you stored in Step 1. The actual credentials never appear in the YAML.

## Step 4: Apply and Verify

Apply the manifest:

```bash
planton apply -f aws-connection.yaml
```

Planton validates that the referenced secrets and variables exist in your organization's secrets manager, resolves them, and creates the connection.

Verify the connection:

```bash
planton get AwsProviderConnection my-aws-connection -o yaml
```

In the output, `access_key_id` and `secret_access_key` show only the secret slug -- not the credential values. The `region` field shows the variable slug. This confirms that the connection stores references, not raw values.

## Step 5: Rotate a Credential

When credentials are compromised or need periodic rotation, update the secret:

```bash
planton secret set aws-access-key-id value=AKIAI_NEW_KEY_EXAMPLE
```

This creates a new immutable version of the secret. The previous version is preserved in the version history. Every resource that references `aws-access-key-id` -- including the connection you created -- will use the new value on its next execution.

Check the version history:

```bash
planton secret versions aws-access-key-id
```

If you need to roll back to a previous version:

```bash
planton secret rollback aws-access-key-id --version 1
```

The connection manifest itself does not change during rotation. Because it references the secret by slug, the rotation is transparent to the resource.

## What to Do Next

- **Connect your AWS account** using the full tutorial with all three authentication paths: [How to Connect Your AWS Account to Planton](/tutorials/how-to-connect-your-aws-account-to-planton)
- **Connect your GCP project** using the same secret reference pattern with a service account key: [How to Connect Your GCP Project to Planton](/tutorials/how-to-connect-your-gcp-project-to-planton)
- **Explore secrets documentation** for additional input methods (file, environment variable, stdin), secret backends, and environment-scoped secrets: [Secrets documentation](/docs/secrets/managing-secrets)
- **Explore variables documentation** for variable groups and environment-scoped variables: [Variables documentation](/docs/variables)
