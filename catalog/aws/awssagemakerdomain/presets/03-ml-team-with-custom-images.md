# ML Team with Custom Images

A fully-featured SageMaker Domain for advanced ML teams that need custom Docker images,
GPU compute, Docker build capabilities, notebook sharing, and auto-cloned code repositories.

## When to Use

- ML platform teams with custom training frameworks
- Teams building custom Docker images for training and inference
- GPU-intensive workloads (computer vision, NLP, deep learning)
- Collaborative teams that need notebook output sharing

## Configuration Highlights

- **Auth mode**: SSO (enterprise identity management)
- **Network**: VpcOnly (secure, Docker pulls restricted to trusted accounts)
- **Encryption**: Customer-managed KMS for EFS and shared notebook outputs
- **Docker**: Enabled with trusted account restrictions
- **JupyterLab**: `ml.m5.large` default, 3-hour idle timeout, 2 auto-cloned repos
- **Custom images**: PyTorch GPU and TensorFlow custom images for JupyterLab
- **KernelGateway**: `ml.g4dn.xlarge` (GPU) with custom ML framework image
- **Sharing**: Notebook outputs persisted to S3 with KMS encryption
- **Storage**: 50 GB default / 500 GB max EBS per space

## Cost Estimate

Per-user compute is the main driver, tempered by the 3-hour idle timeout:
- JupyterLab `ml.m5.large`: billed hourly while running
- KernelGateway `ml.g4dn.xlarge` (GPU): billed hourly when active — the GPU instance is the cost cliff, many times the CPU rate
- EBS storage: billed per GB-month (50-500 GB per space)
- EFS home directories: billed per GB-month
- S3 shared notebook outputs: billed per GB-month, the cheapest storage tier here

The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awssagemakerdomain.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Adjust `idleTimeoutInMinutes` based on team workflow (shorter = lower cost)
- Add more `customImages` as new ML frameworks are onboarded
- Increase `maximumEbsVolumeSizeInGb` for large dataset workflows
- Add additional `codeRepositories` for project-specific repos
