# Basic JupyterLab Domain

A minimal SageMaker Domain for getting started with JupyterLab in your VPC.

## When to Use

- Development and exploration environments
- Small teams getting started with SageMaker Studio
- Quick setup for ML experimentation

## Configuration Highlights

- **Clean teardown**: `homeEfsRetentionPolicy: Delete` removes the domain's
  auto-created EFS home file system with the domain (AWS's default Retain
  leaves it behind, still billing)

- **Auth mode**: IAM (simplest setup, no SSO required)
- **Network**: PublicInternetOnly (default, notebooks can access internet)
- **Encryption**: AWS-managed default keys for EFS
- **IDE**: JupyterLab available with default settings
- **Storage**: Default EFS home directories

## Cost Estimate

No domain-level charges. Costs accrue when users launch JupyterLab instances:
- Instance compute (e.g. `ml.t3.medium`): billed hourly while the instance runs — the dominant line if instances are left running
- EFS home directories: billed per GB-month

The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awssagemakerdomain.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Add `kmsKeyId` for custom EFS encryption
- Set `appNetworkAccessType: VpcOnly` for production security
- Add `jupyterLabAppSettings.idleSettings` to control costs via auto-shutdown
- Add `securityGroupIds` for network isolation
