# CI/CD Pipeline Service Account

This preset creates a GCP service account with a JSON key for CI/CD pipelines (GitHub Actions, GitLab CI, Jenkins). It has permissions to push container images, deploy to GKE, and deploy Cloud Run services.

## When to Use

- CI/CD pipelines that need to authenticate to GCP from external systems
- Automated deployments to GKE clusters and Cloud Run services
- Build pipelines that push container images to Artifact Registry

## Key Configuration Choices

- **Key generated** (`createKey: true`) -- produces a JSON key for use in CI/CD secret stores
- **Artifact Registry writer** -- push container images and other artifacts
- **Container developer** -- deploy workloads to GKE clusters
- **Cloud Run developer** -- deploy and update Cloud Run services
- **Service Account User** -- required to deploy as other service accounts (e.g., Cloud Run service identity)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-sa-id>` | Service account ID (6-30 chars, lowercase) | Choose a descriptive ID (e.g., `ci-cd-pipeline`) |
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |

## Related Presets

- **01-workload-identity** -- Use for runtime service accounts that don't need exported keys
