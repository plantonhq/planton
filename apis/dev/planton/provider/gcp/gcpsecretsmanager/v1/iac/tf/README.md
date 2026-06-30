# Terraform Module for GCP Secrets Manager

This Terraform module deploys and manages secrets in Google Cloud Secret Manager based on the `GcpSecretsManager` specification.

## Overview

The module provisions secrets in GCP Secret Manager with:
- Automatic multi-region replication for high availability
- Placeholder secret versions (actual values set separately for security)
- Proper labeling based on resource metadata
- Environment-specific secret naming when applicable

## Usage

### Initialize Terraform Backend

Initialize the Terraform backend with remote state storage:

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=planton/gcp-stacks/test-gcp-secrets-manager.tfstate"
```

### Plan Changes

Preview the changes that will be applied:

```shell
planton tofu plan --manifest hack/manifest.yaml
```

### Apply Configuration

Deploy the GCP Secrets Manager resources:

```shell
planton tofu apply --manifest hack/manifest.yaml --auto-approve
```

### Destroy Resources

Remove all managed GCP Secrets Manager resources:

```shell
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

## What Gets Created

- **Secret Resources**: One secret container for each name in `spec.secretNames`
- **Secret Versions**: Initial placeholder versions for each secret
- **Labels**: Automatic resource labels based on metadata (org, env, resource name)
- **Replication**: Automatic multi-region replication configured by default

## Setting Actual Secret Values

After the secrets are created, set the actual values using the `gcloud` CLI:

```shell
# Set a secret value
echo -n "actual-secret-value" | gcloud secrets versions add SECRET_NAME \
  --project=PROJECT_ID \
  --data-file=-

# For environment-prefixed secrets
echo -n "actual-secret-value" | gcloud secrets versions add ENV-SECRET_NAME \
  --project=PROJECT_ID \
  --data-file=-
```

## Example Manifest

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSecretsManager
metadata:
  name: app-secrets
  env: prod
spec:
  projectId: my-gcp-project
  secretNames:
    - database-password
    - api-key
    - oauth-secret
```

## Outputs

The module exports a map of secret names to their actual GCP secret IDs (including environment prefix if applicable):

```
secret_id_map = {
  "database-password" = "prod-database-password"
  "api-key" = "prod-api-key"
  "oauth-secret" = "prod-oauth-secret"
}
```

## Security Considerations

1. **Never commit actual secret values** to the manifest or Git repository
2. **Use separate GCP projects** for production and non-production secrets
3. **Apply least privilege IAM** using per-secret access controls
4. **Enable audit logging** to track secret access
5. **Rotate secrets regularly** using versioning capabilities

## Additional Resources

- [GCP Secret Manager Documentation](https://cloud.google.com/secret-manager/docs)
- [Secret Manager Best Practices](https://cloud.google.com/secret-manager/docs/best-practices)
- [IAM Permissions for Secret Manager](https://cloud.google.com/secret-manager/docs/access-control)
