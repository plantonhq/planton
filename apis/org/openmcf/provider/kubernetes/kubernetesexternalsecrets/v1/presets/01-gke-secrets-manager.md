# External Secrets on GKE with Google Cloud Secret Manager

This preset deploys the External Secrets Operator (ESO) on a GKE cluster with Google Cloud Secret Manager as the backing store. Authentication uses GKE Workload Identity via a Google service account.

## When to Use

- You run GKE and store secrets in Google Cloud Secret Manager
- You want Kubernetes Secrets to be automatically synced from Cloud Secret Manager
- GKE Workload Identity is enabled on your cluster

## Key Configuration Choices

- **GKE Workload Identity** -- ESO authenticates via a Google service account; no JSON key files in the cluster
- **Default poll interval** (10 seconds) -- how often ESO checks for secret changes; increase to reduce API costs
- **Resource defaults** -- standard operator resources from proto defaults

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project hosting the secrets and the GKE cluster | GCP Console > Project Settings |
| `<your-gsa-email@your-project.iam.gserviceaccount.com>` | Google service account with Secret Manager Secret Accessor role | GCP Console > IAM > Service Accounts |

## Related Presets

- **02-eks-secrets-manager** -- Use on EKS with AWS Secrets Manager
- **03-aks-key-vault** -- Use on AKS with Azure Key Vault
