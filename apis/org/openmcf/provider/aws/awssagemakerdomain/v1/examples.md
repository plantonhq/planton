# AwsSagemakerDomain Examples

Realistic deployment examples for the AwsSagemakerDomain resource. Each example is a complete manifest ready for customization.

---

## 1. Minimal IAM domain

The simplest domain setup. IAM authentication, public internet access, and a single execution role. Suitable for individual developers or small teams getting started with SageMaker Studio.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: dev-ml-sandbox
spec:
  authMode: IAM
  vpcId:
    value: vpc-0abc123def456789
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::111122223333:role/SageMakerExecutionRole
```

**Key points:**
- `IAM` auth is the simplest mode — users authenticate with AWS credentials (access keys, CLI profiles, or instance roles).
- `PublicInternetOnly` (default) allows notebooks to install pip packages and clone public repos directly.
- No KMS key specified — AWS uses the default `aws/elasticfilesystem` service key for EFS encryption.
- No idle timeout — instances run until manually stopped. Fine for development; not recommended for production cost control.

---

## 2. VPC-only with KMS encryption

Locked-down domain with VPC-only networking and customer-managed encryption. No public internet access from notebooks — all traffic stays within the VPC. Suitable for regulated environments.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: secure-ml-domain
spec:
  authMode: IAM
  vpcId:
    value: vpc-0secure987654321
  subnetIds:
    - value: subnet-secure-az1
    - value: subnet-secure-az2
    - value: subnet-secure-az3
  kmsKeyId:
    value: arn:aws:kms:us-east-1:111122223333:key/mrk-abc123def456
  appNetworkAccessType: VpcOnly
  domainSecurityGroupIds:
    - value: sg-0domain001
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::111122223333:role/SecureSageMakerRole
    securityGroupIds:
      - value: sg-0notebook001
      - value: sg-0notebook002
```

**Key points:**
- `VpcOnly` prevents direct internet access from notebooks. Requires a NAT gateway in the VPC for pip/conda installs.
- Customer-managed KMS key encrypts the EFS home directory volume — enables key rotation policies and audit trails.
- Domain-level security group (`domainSecurityGroupIds`) controls traffic for domain-scoped shared resources.
- User-level security groups (`securityGroupIds`) control traffic for individual notebook ENIs.
- Three subnets across AZs for high availability.

---

## 3. JupyterLab with idle timeout and code repositories

Developer-friendly domain with JupyterLab as the default IDE, automatic idle shutdown for cost control, and team repositories pre-cloned on startup.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: team-notebooks
spec:
  authMode: IAM
  vpcId:
    value: vpc-0team123456789
  subnetIds:
    - value: subnet-team-az1
    - value: subnet-team-az2
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::111122223333:role/TeamSageMakerRole
    defaultLandingUri: "studio::relative/JupyterLab"
    studioWebPortal: ENABLED
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.t3.medium
      lifecycleConfigArns:
        - arn:aws:sagemaker:us-east-1:111122223333:studio-lifecycle-config/install-team-packages
      codeRepositories:
        - repositoryUrl: "https://github.com/acme-corp/ml-notebooks.git"
        - repositoryUrl: "https://github.com/acme-corp/data-pipelines.git"
        - repositoryUrl: "https://github.com/acme-corp/model-registry-tools.git"
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 120
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
```

**Key points:**
- `defaultLandingUri: "studio::relative/JupyterLab"` opens JupyterLab directly on login instead of the Studio home page.
- `ml.t3.medium` (2 vCPU, 4 GB) is cost-effective for exploration and light development.
- Lifecycle config script (`install-team-packages`) runs on every JupyterLab startup — install shared packages, configure extensions, set env vars.
- Three code repos are automatically cloned into each user's workspace.
- Idle timeout at 120 minutes (2 hours) — instances stop automatically after inactivity, releasing compute.
- `minIdleTimeoutInMinutes: 60` prevents users from setting disruptively short timeouts.
- `maxIdleTimeoutInMinutes: 480` prevents users from effectively disabling idle shutdown.

---

## 4. KernelGateway with custom images

Domain configured with custom Docker images for specialized ML frameworks. Use when standard SageMaker kernels don't have the libraries or GPU drivers your team needs.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: custom-kernels-domain
spec:
  authMode: IAM
  vpcId:
    value: vpc-0kernels123456
  subnetIds:
    - value: subnet-kernels-az1
    - value: subnet-kernels-az2
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::111122223333:role/CustomKernelSageMakerRole
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
      customImages:
        - appImageConfigName: pytorch-custom-config
          imageName: pytorch-210-cuda118
          imageVersionNumber: 3
        - appImageConfigName: tensorflow-custom-config
          imageName: tensorflow-215-gpu
        - appImageConfigName: rapids-config
          imageName: rapids-2312-cuda12
    kernelGatewayAppSettings:
      defaultResourceSpec:
        instanceType: ml.g4dn.xlarge
      lifecycleConfigArns:
        - arn:aws:sagemaker:us-east-1:111122223333:studio-lifecycle-config/setup-gpu-monitoring
      customImages:
        - appImageConfigName: xgboost-gpu-config
          imageName: xgboost-gpu-optimized
          imageVersionNumber: 2
        - appImageConfigName: spark-nlp-config
          imageName: spark-nlp-440
```

**Key points:**
- JupyterLab `customImages` provide three custom kernels: PyTorch with CUDA 11.8, TensorFlow with GPU support, and RAPIDS for GPU-accelerated data science.
- KernelGateway `customImages` add XGBoost GPU-optimized and Spark NLP kernels for specialized workloads.
- `ml.g4dn.xlarge` (1 NVIDIA T4 GPU, 4 vCPU, 16 GB) as the default KernelGateway instance for GPU workloads.
- `imageVersionNumber` pins PyTorch to version 3 for reproducibility; TensorFlow uses `latest` by omitting the field.
- Each `appImageConfigName` must correspond to a pre-existing `SageMaker AppImageConfig` that defines the kernel specification.
- Lifecycle config on KernelGateway sets up GPU monitoring tools on startup.

---

## 5. Docker-enabled ML development domain

Domain with Docker access enabled for building custom training containers and inference images directly from notebooks. Requires VpcOnly mode with trusted account restrictions.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: docker-ml-dev
spec:
  authMode: IAM
  vpcId:
    value: vpc-0docker123456789
  subnetIds:
    - value: subnet-docker-az1
    - value: subnet-docker-az2
  appNetworkAccessType: VpcOnly
  dockerSettings:
    enableDockerAccess: ENABLED
    vpcOnlyTrustedAccounts:
      - "111122223333"
      - "444455556666"
      - "763104351884"
  domainSecurityGroupIds:
    - value: sg-0docker-domain
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::111122223333:role/DockerSageMakerRole
    securityGroupIds:
      - value: sg-0docker-user
    defaultLandingUri: "studio::relative/JupyterLab"
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.xlarge
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 180
    spaceStorageSettings:
      defaultEbsVolumeSizeInGb: 100
      maximumEbsVolumeSizeInGb: 500
```

**Key points:**
- `enableDockerAccess: ENABLED` allows `docker build`, `docker pull`, and `docker run` from notebooks and terminals.
- `appNetworkAccessType: VpcOnly` is required for Docker access — public internet mode does not support Docker.
- `vpcOnlyTrustedAccounts` restricts image pulls to three AWS accounts: the team's account, a shared services account, and the AWS Deep Learning Containers account (`763104351884`).
- `ml.m5.xlarge` (4 vCPU, 16 GB) provides enough memory for Docker image builds.
- `spaceStorageSettings` allocates 100 GB default EBS for Docker images and build artifacts, with up to 500 GB available on request.
- 3-hour idle timeout balances cost savings with long-running build workflows.

---

## 6. Production-ready with all settings (SSO, VpcOnly, KMS, idle, sharing, Docker)

Enterprise-grade domain with SSO authentication, VPC-only networking, customer-managed encryption, idle timeout, notebook sharing, Docker access, and cross-resource references. This is the recommended pattern for production ML platforms.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: prod-ml-platform
spec:
  authMode: SSO
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: ml-platform-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: ml-platform-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: ml-platform-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: ml-platform-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: ml-data-encryption-key
      fieldPath: status.outputs.key_arn
  appNetworkAccessType: VpcOnly
  domainSecurityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ml-domain-boundary-sg
        fieldPath: status.outputs.security_group_id
  dockerSettings:
    enableDockerAccess: ENABLED
    vpcOnlyTrustedAccounts:
      - "111122223333"
      - "444455556666"
      - "763104351884"
  defaultUserSettings:
    executionRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: sagemaker-execution-role
        fieldPath: status.outputs.role_arn
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: ml-notebook-sg
          fieldPath: status.outputs.security_group_id
      - valueFrom:
          kind: AwsSecurityGroup
          name: ml-training-sg
          fieldPath: status.outputs.security_group_id
    defaultLandingUri: "studio::relative/JupyterLab"
    studioWebPortal: ENABLED
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
        lifecycleConfigArn: arn:aws:sagemaker:us-east-1:111122223333:studio-lifecycle-config/prod-jupyter-setup
      lifecycleConfigArns:
        - arn:aws:sagemaker:us-east-1:111122223333:studio-lifecycle-config/prod-jupyter-setup
        - arn:aws:sagemaker:us-east-1:111122223333:studio-lifecycle-config/install-compliance-tools
      customImages:
        - appImageConfigName: approved-pytorch-config
          imageName: approved-pytorch-210
          imageVersionNumber: 5
        - appImageConfigName: approved-tensorflow-config
          imageName: approved-tensorflow-215
          imageVersionNumber: 3
      codeRepositories:
        - repositoryUrl: "https://github.com/acme-corp/ml-platform-notebooks.git"
        - repositoryUrl: "https://github.com/acme-corp/feature-engineering.git"
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 120
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
    kernelGatewayAppSettings:
      defaultResourceSpec:
        instanceType: ml.g4dn.xlarge
      customImages:
        - appImageConfigName: approved-gpu-training-config
          imageName: approved-gpu-training
          imageVersionNumber: 2
    sharingSettings:
      notebookOutputOption: Allowed
      s3OutputPath: "s3://ml-platform-artifacts/notebook-outputs/"
      s3KmsKeyId:
        valueFrom:
          kind: AwsKmsKey
          name: ml-data-encryption-key
          fieldPath: status.outputs.key_arn
    spaceStorageSettings:
      defaultEbsVolumeSizeInGb: 50
      maximumEbsVolumeSizeInGb: 500
```

**Key points:**
- `SSO` authentication integrates with AWS IAM Identity Center for enterprise single sign-on — users log in with corporate credentials.
- `VpcOnly` with customer-managed KMS satisfies compliance requirements (HIPAA, SOC2, PCI-DSS).
- `valueFrom` references wire all infrastructure dependencies (VPC, subnets, security groups, KMS keys, IAM roles) without hardcoded IDs — ideal for infrastructure chart composition.
- Two user-level security groups separate notebook traffic from training job traffic.
- Approved custom images with pinned version numbers ensure reproducibility and security vetting.
- Lifecycle configs install compliance tools and configure the environment on every JupyterLab startup.
- Notebook sharing to S3 (encrypted with the same KMS key) enables review workflows and audit trails.
- Docker access with trusted accounts limits image pulls to approved registries.
- Space storage settings cap EBS volumes at 500 GB to control storage costs.
- Idle timeout at 2 hours with guardrails (min 1 hour, max 8 hours) balances cost savings with usability.
