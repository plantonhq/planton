# Application Secrets

This preset creates a set of common application secrets in Google Cloud Secret Manager. The secrets are created as empty shells -- secret values must be populated separately via the GCP console, `gcloud`, or your CI/CD pipeline.

## When to Use

- Bootstrapping secret infrastructure for a new application
- Applications that need database credentials, API keys, and authentication secrets
- Any workload that reads secrets at runtime via the Secret Manager API or External Secrets Operator

## Key Configuration Choices

- **4 common secret names** -- database password, API key, OAuth client secret, and JWT signing key
- **Empty shells** -- only the secret metadata is created; values must be added separately
- **Naming convention** -- lowercase with hyphens, matching Kubernetes secret naming patterns

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
