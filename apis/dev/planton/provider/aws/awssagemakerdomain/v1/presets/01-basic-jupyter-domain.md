# Preset: Basic JupyterLab Domain

A minimal SageMaker Domain for getting started with JupyterLab in your VPC.

## When to Use

- Development and exploration environments
- Small teams getting started with SageMaker Studio
- Quick setup for ML experimentation

## Configuration Highlights

- **Auth mode**: IAM (simplest setup, no SSO required)
- **Network**: PublicInternetOnly (default, notebooks can access internet)
- **Encryption**: AWS-managed default keys for EFS
- **IDE**: JupyterLab available with default settings
- **Storage**: Default EFS home directories

## Cost Estimate

No domain-level charges. Costs accrue when users launch JupyterLab instances:
- `ml.t3.medium`: ~$0.05/hr (~$37/month if running 24/7)
- EFS storage: $0.30/GB-month for home directories

## Customization

- Add `kmsKeyId` for custom EFS encryption
- Set `appNetworkAccessType: VpcOnly` for production security
- Add `jupyterLabAppSettings.idleSettings` to control costs via auto-shutdown
- Add `securityGroupIds` for network isolation
