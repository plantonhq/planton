# External Secrets on EKS with AWS Secrets Manager

This preset deploys the External Secrets Operator (ESO) on an EKS cluster with AWS Secrets Manager as the backing store. Authentication uses IRSA (IAM Roles for Service Accounts); an IAM role is auto-created if not explicitly overridden.

## When to Use

- You run EKS and store secrets in AWS Secrets Manager
- You want Kubernetes Secrets to be automatically synced from AWS Secrets Manager

## Key Configuration Choices

- **IRSA authentication** -- ESO's service account assumes an IAM role via OIDC; no long-lived AWS credentials needed
- **Auto-created IAM role** -- if `irsaRoleArnOverride` is not set, the stack creates and binds the required IAM role automatically
- **Default poll interval** (10 seconds) -- how often ESO checks for secret changes

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-aws-region>` | AWS region containing the secrets (e.g., `us-east-1`) | AWS Console > Secrets Manager |

## Related Presets

- **01-gke-secrets-manager** -- Use on GKE with Google Cloud Secret Manager
- **03-aks-key-vault** -- Use on AKS with Azure Key Vault
