# AwsSagemakerDomain

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsSagemakerDomainSpec defines the desired state of an Amazon SageMaker Domain.
A SageMaker Domain is the top-level organizational unit for Amazon SageMaker Studio,
providing a shared workspace where data scientists and ML engineers access JupyterLab
notebooks, the Code Editor IDE, Canvas no-code ML, custom kernels, and ML tooling.
The domain provisions a dedicated EFS file system for user home directories, configures
VPC networking for notebook and training traffic, and establishes IAM execution roles
that govern what AWS resources users can access from their Studio sessions.

The domain carries two inheritance planes: default_user_settings is the baseline every
user profile inherits, and default_space_settings is the baseline every shared
(collaborative) space inherits. Domain-scoped administration knobs -- Docker access,
domain-level security groups, RStudio licensing, identity-propagation posture -- sit at
the top level of this spec.

## Example

```yaml
# Full-surface development manifest for the offline plan proof: every optional
# arm the module wires is exercised so `tofu plan` / `pulumi preview` cover the
# arms the live E2E lanes exclude (RStudio needs a Posit license, SSO needs an
# IAM Identity Center instance). SSO auth mode is used here because trusted
# identity propagation only rides SSO domains -- the live API rejects the
# setting outright on IAM-auth domains, and the spec CEL mirrors that.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerDomain
metadata:
  org: example-org
  env: dev
  name: hack-sagemaker-domain
  id: sgmkd-hack-sagemaker-domain
spec:
  region: us-west-2
  authMode: SSO
  vpcId:
    value: vpc-0123456789abcdef0
  subnetIds:
    - value: subnet-0123456789abcdef1
    - value: subnet-0123456789abcdef2
  kmsKeyId:
    value: arn:aws:kms:us-west-2:123456789012:key/mrk-hack
  appNetworkAccessType: VpcOnly
  tagPropagation: ENABLED
  homeEfsRetentionPolicy: Delete
  executionRoleIdentityConfig: USER_PROFILE_NAME
  trustedIdentityPropagationStatus: ENABLED
  appSecurityGroupManagement: Customer
  domainSecurityGroupIds:
    - value: sg-0123456789abcdef3
  dockerSettings:
    enableDockerAccess: ENABLED
    vpcOnlyTrustedAccounts:
      - "123456789012"
  rStudioServerProDomainSettings:
    domainExecutionRoleArn:
      value: arn:aws:iam::123456789012:role/hack-rstudio-domain-role
    rStudioConnectUrl: https://connect.example.com
    rStudioPackageManagerUrl: https://packages.example.com
    defaultResourceSpec:
      instanceType: ml.t3.medium
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::123456789012:role/hack-sagemaker-execution-role
    securityGroupIds:
      - value: sg-0123456789abcdef4
    defaultLandingUri: "studio::"
    studioWebPortal: ENABLED
    autoMountHomeEfs: Enabled
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.t3.medium
        sagemakerImageArn: arn:aws:sagemaker:us-west-2:123456789012:image/hack-image
        sagemakerImageVersionAlias: latest
      lifecycleConfigArns:
        - arn:aws:sagemaker:us-west-2:123456789012:studio-lifecycle-config/install-packages
      builtInLifecycleConfigArn: arn:aws:sagemaker:us-west-2:123456789012:studio-lifecycle-config/built-in
      customImages:
        - appImageConfigName: hack-pytorch-config
          imageName: hack-pytorch
          imageVersionNumber: 1
      codeRepositories:
        - repositoryUrl: https://github.com/example/ml-notebooks.git
      idleSettings:
        idleTimeoutInMinutes: 120
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
      emrSettings:
        assumableRoleArns:
          - value: arn:aws:iam::123456789012:role/hack-emr-connect-role
        executionRoleArns:
          - value: arn:aws:iam::123456789012:role/hack-emr-runtime-role
    jupyterServerAppSettings:
      defaultResourceSpec:
        instanceType: system
      codeRepositories:
        - repositoryUrl: https://github.com/example/classic-notebooks.git
    kernelGatewayAppSettings:
      defaultResourceSpec:
        instanceType: ml.g4dn.xlarge
      customImages:
        - appImageConfigName: hack-gpu-config
          imageName: hack-gpu-image
    codeEditorAppSettings:
      defaultResourceSpec:
        instanceType: ml.t3.large
      idleSettings:
        idleTimeoutInMinutes: 180
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
    tensorBoardAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
    rSessionAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
    rStudioServerProAppSettings:
      accessStatus: ENABLED
      userGroup: R_STUDIO_USER
    canvasAppSettings:
      directDeployStatus: DISABLED
      emrServerlessSettings:
        executionRoleArn:
          value: arn:aws:iam::123456789012:role/hack-canvas-emr-role
        status: ENABLED
      generativeAiBedrockRoleArn:
        value: arn:aws:iam::123456789012:role/hack-canvas-bedrock-role
      identityProviderOauthSettings:
        - dataSourceName: Snowflake
          secretArn: arn:aws:secretsmanager:us-west-2:123456789012:secret:hack-snowflake-oauth
          status: ENABLED
      kendraSettingsStatus: DISABLED
      modelRegisterSettings:
        status: ENABLED
      timeSeriesForecastingSettings:
        amazonForecastRoleArn:
          value: arn:aws:iam::123456789012:role/hack-canvas-forecast-role
        status: ENABLED
      workspaceSettings:
        s3ArtifactPath: s3://hack-canvas-workspace/artifacts/
        s3KmsKeyId:
          value: arn:aws:kms:us-west-2:123456789012:key/mrk-canvas
    sharingSettings:
      notebookOutputOption: Allowed
      s3OutputPath: s3://hack-team-bucket/notebook-outputs/
      s3KmsKeyId:
        value: arn:aws:kms:us-west-2:123456789012:key/mrk-share
    spaceStorageSettings:
      defaultEbsVolumeSizeInGb: 20
      maximumEbsVolumeSizeInGb: 200
    customFileSystemConfigs:
      - efsFileSystemConfig:
          fileSystemId:
            value: fs-0123456789abcdef5
          fileSystemPath: /shared/datasets
    customPosixUserConfig:
      uid: 10000
      gid: 1001
    studioWebPortalSettings:
      hiddenAppTypes:
        - JupyterServer
      hiddenInstanceTypes:
        - ml.p3.2xlarge
      hiddenMlTools:
        - DataWrangler
  defaultSpaceSettings:
    executionRoleArn:
      value: arn:aws:iam::123456789012:role/hack-space-execution-role
    securityGroupIds:
      - value: sg-0123456789abcdef6
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.t3.medium
      idleSettings:
        idleTimeoutInMinutes: 120
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
    jupyterServerAppSettings:
      defaultResourceSpec:
        instanceType: system
    kernelGatewayAppSettings:
      defaultResourceSpec:
        instanceType: ml.t3.medium
    spaceStorageSettings:
      defaultEbsVolumeSizeInGb: 10
      maximumEbsVolumeSizeInGb: 100
    customFileSystemConfigs:
      - efsFileSystemConfig:
          fileSystemId:
            value: fs-0123456789abcdef5
          fileSystemPath: /shared/spaces
    customPosixUserConfig:
      uid: 10001
      gid: 1002
  # The folded per-person plane: profiles inherit defaultUserSettings and
  # override per person. The first profile's idle settings carry
  # lifecycleManagement DISABLED -- guardrails published, enforcement off.
  userProfiles:
    - userProfileName: alice
      userSettings:
        executionRoleArn:
          value: arn:aws:iam::123456789012:role/hack-user-execution-role
        jupyterLabAppSettings:
          defaultResourceSpec:
            instanceType: ml.t3.medium
          idleSettings:
            lifecycleManagement: DISABLED
            idleTimeoutInMinutes: 120
            minIdleTimeoutInMinutes: 60
            maxIdleTimeoutInMinutes: 240
    - userProfileName: bob
  # The folded collaboration plane: named spaces (ownership and sharing
  # travel together -- AWS's RequiredWith contract).
  spaces:
    - spaceName: team-analytics
      displayName: Team Analytics
      ownershipSettings:
        ownerUserProfileName: alice
      spaceSharingSettings:
        sharingType: Shared
      spaceSettings:
        appType: JupyterLab
        jupyterLabAppSettings:
          defaultResourceSpec:
            instanceType: ml.t3.medium
          # No idle settings here: alice (the owner profile) resolves
          # lifecycleManagement DISABLED for JupyterLab, and AWS rejects
          # SpaceIdleSettings on a disabled plane.
        spaceStorageSettings:
          ebsVolumeSizeInGb: 25
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.authMode` | `string` | yes |  |  |
| `spec.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.appNetworkAccessType` | `string` |  | `PublicInternetOnly` |  |
| `spec.appSecurityGroupManagement` | `string` |  |  |  |
| `spec.tagPropagation` | `string` |  | `DISABLED` |  |
| `spec.homeEfsRetentionPolicy` | `string` |  | `Retain` |  |
| `spec.defaultUserSettings` | `AwsSagemakerDomainUserSettings` | yes |  |  |
| `spec.defaultUserSettings.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultUserSettings.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.defaultUserSettings.defaultLandingUri` | `string` |  |  |  |
| `spec.defaultUserSettings.studioWebPortal` | `string` |  | `ENABLED` |  |
| `spec.defaultUserSettings.autoMountHomeEfs` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings` | `AwsSagemakerDomainJupyterLabAppSettings` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.builtInLifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.idleSettings` | `AwsSagemakerDomainIdleSettings` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.lifecycleManagement` | `string` |  | `ENABLED` |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.minIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.maxIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.emrSettings` | `AwsSagemakerDomainEmrSettings` |  |  |  |
| `spec.defaultUserSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultUserSettings.jupyterLabAppSettings.emrSettings.executionRoleArns` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultUserSettings.jupyterServerAppSettings` | `AwsSagemakerDomainJupyterServerAppSettings` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.defaultUserSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings` | `AwsSagemakerDomainKernelGatewayAppSettings` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.defaultUserSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings` | `AwsSagemakerDomainCodeEditorAppSettings` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.builtInLifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.idleSettings` | `AwsSagemakerDomainIdleSettings` |  |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.idleSettings.lifecycleManagement` | `string` |  | `ENABLED` |  |
| `spec.defaultUserSettings.codeEditorAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.idleSettings.minIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultUserSettings.codeEditorAppSettings.idleSettings.maxIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultUserSettings.tensorBoardAppSettings` | `AwsSagemakerDomainTensorBoardAppSettings` |  |  |  |
| `spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings` | `AwsSagemakerDomainRSessionAppSettings` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.defaultUserSettings.rSessionAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.defaultUserSettings.rStudioServerProAppSettings` | `AwsSagemakerDomainRStudioServerProAppSettings` |  |  |  |
| `spec.defaultUserSettings.rStudioServerProAppSettings.accessStatus` | `string` |  |  |  |
| `spec.defaultUserSettings.rStudioServerProAppSettings.userGroup` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings` | `AwsSagemakerDomainCanvasAppSettings` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.directDeployStatus` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.emrServerlessSettings` | `AwsSagemakerDomainCanvasEmrServerlessSettings` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.emrServerlessSettings.executionRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultUserSettings.canvasAppSettings.emrServerlessSettings.status` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.generativeAiBedrockRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings` | `[]AwsSagemakerDomainCanvasIdentityProviderOauthSettings` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings[].dataSourceName` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings[].secretArn` | `string` | yes |  |  |
| `spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings[].status` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.kendraSettingsStatus` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.modelRegisterSettings` | `AwsSagemakerDomainCanvasModelRegisterSettings` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.modelRegisterSettings.crossAccountModelRegisterRoleArn` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.modelRegisterSettings.status` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings` | `AwsSagemakerDomainCanvasTimeSeriesForecastingSettings` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings.amazonForecastRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings.status` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.workspaceSettings` | `AwsSagemakerDomainCanvasWorkspaceSettings` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.workspaceSettings.s3ArtifactPath` | `string` |  |  |  |
| `spec.defaultUserSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.defaultUserSettings.sharingSettings` | `AwsSagemakerDomainSharingSettings` |  |  |  |
| `spec.defaultUserSettings.sharingSettings.notebookOutputOption` | `string` |  | `Disabled` |  |
| `spec.defaultUserSettings.sharingSettings.s3KmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.defaultUserSettings.sharingSettings.s3OutputPath` | `string` |  |  |  |
| `spec.defaultUserSettings.spaceStorageSettings` | `AwsSagemakerDomainSpaceStorageSettings` |  |  |  |
| `spec.defaultUserSettings.spaceStorageSettings.defaultEbsVolumeSizeInGb` | `int32` | yes |  |  |
| `spec.defaultUserSettings.spaceStorageSettings.maximumEbsVolumeSizeInGb` | `int32` | yes |  |  |
| `spec.defaultUserSettings.customFileSystemConfigs` | `[]AwsSagemakerDomainCustomFileSystemConfig` |  |  |  |
| `spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig` | `AwsSagemakerDomainEfsFileSystemConfig` | yes |  |  |
| `spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | `string \| valueFrom` | yes |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |
| `spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemPath` | `string` | yes |  |  |
| `spec.defaultUserSettings.customPosixUserConfig` | `AwsSagemakerDomainCustomPosixUserConfig` |  |  |  |
| `spec.defaultUserSettings.customPosixUserConfig.uid` | `int64` | yes |  |  |
| `spec.defaultUserSettings.customPosixUserConfig.gid` | `int64` | yes |  |  |
| `spec.defaultUserSettings.studioWebPortalSettings` | `AwsSagemakerDomainStudioWebPortalSettings` |  |  |  |
| `spec.defaultUserSettings.studioWebPortalSettings.hiddenAppTypes` | `[]string` |  |  |  |
| `spec.defaultUserSettings.studioWebPortalSettings.hiddenInstanceTypes` | `[]string` |  |  |  |
| `spec.defaultUserSettings.studioWebPortalSettings.hiddenMlTools` | `[]string` |  |  |  |
| `spec.defaultSpaceSettings` | `AwsSagemakerDomainDefaultSpaceSettings` |  |  |  |
| `spec.defaultSpaceSettings.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultSpaceSettings.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.defaultSpaceSettings.jupyterLabAppSettings` | `AwsSagemakerDomainJupyterLabAppSettings` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.builtInLifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings` | `AwsSagemakerDomainIdleSettings` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.lifecycleManagement` | `string` |  | `ENABLED` |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.minIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.maxIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings` | `AwsSagemakerDomainEmrSettings` |  |  |  |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings.executionRoleArns` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultSpaceSettings.jupyterServerAppSettings` | `AwsSagemakerDomainJupyterServerAppSettings` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.defaultSpaceSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings` | `AwsSagemakerDomainKernelGatewayAppSettings` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.defaultSpaceSettings.spaceStorageSettings` | `AwsSagemakerDomainSpaceStorageSettings` |  |  |  |
| `spec.defaultSpaceSettings.spaceStorageSettings.defaultEbsVolumeSizeInGb` | `int32` | yes |  |  |
| `spec.defaultSpaceSettings.spaceStorageSettings.maximumEbsVolumeSizeInGb` | `int32` | yes |  |  |
| `spec.defaultSpaceSettings.customFileSystemConfigs` | `[]AwsSagemakerDomainCustomFileSystemConfig` |  |  |  |
| `spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig` | `AwsSagemakerDomainEfsFileSystemConfig` | yes |  |  |
| `spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | `string \| valueFrom` | yes |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |
| `spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemPath` | `string` | yes |  |  |
| `spec.defaultSpaceSettings.customPosixUserConfig` | `AwsSagemakerDomainCustomPosixUserConfig` |  |  |  |
| `spec.defaultSpaceSettings.customPosixUserConfig.uid` | `int64` | yes |  |  |
| `spec.defaultSpaceSettings.customPosixUserConfig.gid` | `int64` | yes |  |  |
| `spec.domainSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.dockerSettings` | `AwsSagemakerDomainDockerSettings` |  |  |  |
| `spec.dockerSettings.enableDockerAccess` | `string` |  |  |  |
| `spec.dockerSettings.vpcOnlyTrustedAccounts` | `[]string` |  |  |  |
| `spec.executionRoleIdentityConfig` | `string` |  |  |  |
| `spec.rStudioServerProDomainSettings` | `AwsSagemakerDomainRStudioServerProDomainSettings` |  |  |  |
| `spec.rStudioServerProDomainSettings.domainExecutionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.rStudioServerProDomainSettings.rStudioConnectUrl` | `string` |  |  |  |
| `spec.rStudioServerProDomainSettings.rStudioPackageManagerUrl` | `string` |  |  |  |
| `spec.rStudioServerProDomainSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.rStudioServerProDomainSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.rStudioServerProDomainSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.rStudioServerProDomainSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.rStudioServerProDomainSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.rStudioServerProDomainSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.trustedIdentityPropagationStatus` | `string` |  |  |  |
| `spec.userProfiles` | `[]AwsSagemakerDomainUserProfile` |  |  |  |
| `spec.userProfiles[].userProfileName` | `string` | yes |  |  |
| `spec.userProfiles[].singleSignOnUserIdentifier` | `string` |  |  |  |
| `spec.userProfiles[].singleSignOnUserValue` | `string` |  |  |  |
| `spec.userProfiles[].userSettings` | `AwsSagemakerDomainUserSettings` |  |  |  |
| `spec.userProfiles[].userSettings.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.userProfiles[].userSettings.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.userProfiles[].userSettings.defaultLandingUri` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.studioWebPortal` | `string` |  | `ENABLED` |  |
| `spec.userProfiles[].userSettings.autoMountHomeEfs` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings` | `AwsSagemakerDomainJupyterLabAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.builtInLifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings` | `AwsSagemakerDomainIdleSettings` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.lifecycleManagement` | `string` |  | `ENABLED` |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.minIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.maxIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings` | `AwsSagemakerDomainEmrSettings` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings.executionRoleArns` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings` | `AwsSagemakerDomainJupyterServerAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.userProfiles[].userSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings` | `AwsSagemakerDomainKernelGatewayAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings` | `AwsSagemakerDomainCodeEditorAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.builtInLifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings` | `AwsSagemakerDomainIdleSettings` |  |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.lifecycleManagement` | `string` |  | `ENABLED` |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.minIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.maxIdleTimeoutInMinutes` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.tensorBoardAppSettings` | `AwsSagemakerDomainTensorBoardAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings` | `AwsSagemakerDomainRSessionAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.rSessionAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.userProfiles[].userSettings.rStudioServerProAppSettings` | `AwsSagemakerDomainRStudioServerProAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.rStudioServerProAppSettings.accessStatus` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.rStudioServerProAppSettings.userGroup` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings` | `AwsSagemakerDomainCanvasAppSettings` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.directDeployStatus` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.emrServerlessSettings` | `AwsSagemakerDomainCanvasEmrServerlessSettings` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.emrServerlessSettings.executionRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.userProfiles[].userSettings.canvasAppSettings.emrServerlessSettings.status` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.generativeAiBedrockRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings` | `[]AwsSagemakerDomainCanvasIdentityProviderOauthSettings` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings[].dataSourceName` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings[].secretArn` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings[].status` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.kendraSettingsStatus` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.modelRegisterSettings` | `AwsSagemakerDomainCanvasModelRegisterSettings` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.modelRegisterSettings.crossAccountModelRegisterRoleArn` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.modelRegisterSettings.status` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.timeSeriesForecastingSettings` | `AwsSagemakerDomainCanvasTimeSeriesForecastingSettings` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.timeSeriesForecastingSettings.amazonForecastRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.userProfiles[].userSettings.canvasAppSettings.timeSeriesForecastingSettings.status` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings` | `AwsSagemakerDomainCanvasWorkspaceSettings` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings.s3ArtifactPath` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.userProfiles[].userSettings.sharingSettings` | `AwsSagemakerDomainSharingSettings` |  |  |  |
| `spec.userProfiles[].userSettings.sharingSettings.notebookOutputOption` | `string` |  | `Disabled` |  |
| `spec.userProfiles[].userSettings.sharingSettings.s3KmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.userProfiles[].userSettings.sharingSettings.s3OutputPath` | `string` |  |  |  |
| `spec.userProfiles[].userSettings.spaceStorageSettings` | `AwsSagemakerDomainSpaceStorageSettings` |  |  |  |
| `spec.userProfiles[].userSettings.spaceStorageSettings.defaultEbsVolumeSizeInGb` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.spaceStorageSettings.maximumEbsVolumeSizeInGb` | `int32` | yes |  |  |
| `spec.userProfiles[].userSettings.customFileSystemConfigs` | `[]AwsSagemakerDomainCustomFileSystemConfig` |  |  |  |
| `spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig` | `AwsSagemakerDomainEfsFileSystemConfig` | yes |  |  |
| `spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | `string \| valueFrom` | yes |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |
| `spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemPath` | `string` | yes |  |  |
| `spec.userProfiles[].userSettings.customPosixUserConfig` | `AwsSagemakerDomainCustomPosixUserConfig` |  |  |  |
| `spec.userProfiles[].userSettings.customPosixUserConfig.uid` | `int64` | yes |  |  |
| `spec.userProfiles[].userSettings.customPosixUserConfig.gid` | `int64` | yes |  |  |
| `spec.userProfiles[].userSettings.studioWebPortalSettings` | `AwsSagemakerDomainStudioWebPortalSettings` |  |  |  |
| `spec.userProfiles[].userSettings.studioWebPortalSettings.hiddenAppTypes` | `[]string` |  |  |  |
| `spec.userProfiles[].userSettings.studioWebPortalSettings.hiddenInstanceTypes` | `[]string` |  |  |  |
| `spec.userProfiles[].userSettings.studioWebPortalSettings.hiddenMlTools` | `[]string` |  |  |  |
| `spec.spaces` | `[]AwsSagemakerDomainSpace` |  |  |  |
| `spec.spaces[].spaceName` | `string` | yes |  |  |
| `spec.spaces[].displayName` | `string` |  |  |  |
| `spec.spaces[].ownershipSettings` | `AwsSagemakerDomainSpaceOwnership` |  |  |  |
| `spec.spaces[].ownershipSettings.ownerUserProfileName` | `string` | yes |  |  |
| `spec.spaces[].spaceSharingSettings` | `AwsSagemakerDomainSpaceSharing` |  |  |  |
| `spec.spaces[].spaceSharingSettings.sharingType` | `string` | yes |  |  |
| `spec.spaces[].spaceSettings` | `AwsSagemakerDomainSpaceSettings` |  |  |  |
| `spec.spaces[].spaceSettings.appType` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings` | `AwsSagemakerDomainSpaceJupyterLabAppSettings` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` | yes |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.idleSettings` | `AwsSagemakerDomainSpaceIdleSettings` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings` | `AwsSagemakerDomainSpaceCodeEditorAppSettings` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` | yes |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.idleSettings` | `AwsSagemakerDomainSpaceIdleSettings` |  |  |  |
| `spec.spaces[].spaceSettings.codeEditorAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings` | `AwsSagemakerDomainJupyterServerAppSettings` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.codeRepositories` | `[]AwsSagemakerDomainCodeRepository` |  |  |  |
| `spec.spaces[].spaceSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl` | `string` | yes |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings` | `AwsSagemakerDomainKernelGatewayAppSettings` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec` | `AwsSagemakerDomainResourceSpec` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn` | `string` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.lifecycleConfigArns` | `[]string` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages` | `[]AwsSagemakerDomainCustomImage` |  |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages[].appImageConfigName` | `string` | yes |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages[].imageName` | `string` | yes |  |  |
| `spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber` | `int32` |  |  |  |
| `spec.spaces[].spaceSettings.customFileSystems` | `[]AwsSagemakerDomainSpaceCustomFileSystem` |  |  |  |
| `spec.spaces[].spaceSettings.customFileSystems[].fileSystemId` | `string \| valueFrom` | yes |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |
| `spec.spaces[].spaceSettings.spaceStorageSettings` | `AwsSagemakerDomainSpaceStorage` |  |  |  |
| `spec.spaces[].spaceSettings.spaceStorageSettings.ebsVolumeSizeInGb` | `int32` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.authMode

`string` · required

auth_mode determines how users authenticate to the SageMaker Domain.
"IAM": users authenticate with AWS IAM credentials. Suitable for single-account teams
  or programmatic access. Each user's IAM identity determines their permissions.
"SSO": users authenticate via AWS IAM Identity Center (formerly AWS SSO). Recommended
  for enterprise teams with centralized identity management, and required for
  trusted identity propagation.
ForceNew: changing auth_mode forces domain replacement.

- rule: {"required":true}

### spec.vpcId

`string | valueFrom` · required

vpc_id is the VPC in which the SageMaker Domain is created.
All domain network interfaces (for notebooks, training, and app traffic) are placed
in this VPC. The VPC must have DNS resolution and DNS hostnames enabled.
ForceNew: changing the VPC forces domain replacement.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom` · required

subnet_ids are the VPC subnets where SageMaker provisions elastic network interfaces
for notebook and training traffic. For high availability, provide subnets in at least
two Availability Zones. Maximum 16 subnets.
ForceNew: changing subnets forces domain replacement.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"16"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.kmsKeyId

`string | valueFrom`

kms_key_id is the KMS key used to encrypt the EFS volume attached to the domain.
Each SageMaker Domain creates a dedicated EFS file system for user home directories.
If omitted, AWS uses the default aws/elasticfilesystem service key.
ForceNew: changing the KMS key forces domain replacement.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.appNetworkAccessType

`string` · optional (explicit presence)

app_network_access_type controls whether notebook and training traffic can reach the internet.
"PublicInternetOnly" (default): ENIs have internet access via AWS-managed networking.
"VpcOnly": all traffic stays within the VPC; internet access requires a NAT gateway.
"VpcOnly" is recommended for production to prevent data exfiltration and satisfy
compliance requirements. Docker trusted accounts only work in VpcOnly mode.

- default: `PublicInternetOnly`

### spec.appSecurityGroupManagement

`string` · optional (explicit presence)

app_security_group_management selects who manages the security groups attached to the
ENIs that SageMaker creates for apps.
"Service": SageMaker creates and manages the security groups.
"Customer": you manage them (required when RStudio needs to reach a license endpoint
  through customer-controlled networking).
AWS only honors this setting when RStudio Server Pro is configured on the domain
(r_studio_server_pro_domain_settings), so the spec requires that pairing up front
rather than letting the value be silently ignored at deploy time.

### spec.tagPropagation

`string` · optional (explicit presence)

tag_propagation controls whether tags on the domain automatically propagate to the
SageMaker resources created within it (apps, spaces, user profiles).
"ENABLED": in-domain resources inherit the domain's tags -- recommended for cost
  allocation, because per-app compute charges then carry the domain's tags.
"DISABLED" (default, AWS's own): in-domain resources start untagged.

- default: `DISABLED`

### spec.homeEfsRetentionPolicy

`string` · optional (explicit presence)

home_efs_retention_policy decides what happens to the domain's auto-created EFS file
system (user home directories) when the domain is deleted.
"Retain" (default, AWS's own): the EFS file system survives domain deletion. Home
  directories are preserved, but the file system keeps accruing storage charges and
  must be deleted by hand when no longer needed.
"Delete": the EFS file system is deleted with the domain -- the right choice for
  ephemeral and test domains, and the only choice that leaves nothing behind.
ForceNew: the retention decision is fixed at create time.

- default: `Retain`

### spec.defaultUserSettings

`AwsSagemakerDomainUserSettings` · required

default_user_settings defines the default configuration inherited by all user profiles
in the domain. These settings control the execution environment, per-IDE app
configurations, security boundaries, and storage defaults. Individual user profiles
can override these defaults.

- rule: {"required":true}
- rule: studio_web_portal must be 'ENABLED' or 'DISABLED'
- rule: auto_mount_home_efs must be 'Enabled' or 'Disabled' at the domain level

### spec.defaultUserSettings.executionRoleArn

`string | valueFrom` · required

execution_role_arn is the IAM role assumed by SageMaker when running notebooks,
training jobs, and other ML workloads on behalf of users. This role determines what
AWS resources (S3 buckets, ECR repos, Secrets Manager, etc.) users can access from
their Studio sessions. The role must have a trust policy allowing
sagemaker.amazonaws.com to assume it.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultUserSettings.securityGroupIds

`[]string | valueFrom`

security_group_ids are user-level security groups controlling network access for
notebook instances and apps. These are attached to the ENIs created for each user's
apps and control inbound/outbound traffic at the user level. Maximum 5 security groups.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.defaultUserSettings.defaultLandingUri

`string`

default_landing_uri is the URI of the default app opened when a user accesses the domain.
Common values:
  "studio::relative/JupyterLab" - opens JupyterLab (recommended for most teams)
  "studio::relative/JupyterServer:" - opens classic Jupyter Server
  "studio::" - opens SageMaker Studio home
If omitted, AWS uses the platform default.

### spec.defaultUserSettings.studioWebPortal

`string` · optional (explicit presence)

studio_web_portal controls whether the SageMaker Studio web portal is accessible.
"ENABLED" (default): users can access the full Studio web interface.
"DISABLED": restricts access to programmatic-only usage (API/CLI).

- default: `ENABLED`

### spec.defaultUserSettings.autoMountHomeEfs

`string` · optional (explicit presence)

auto_mount_home_efs controls whether each user's home directory on the domain's EFS
file system is automatically mounted into their apps.
"Enabled": home directories mount automatically (the classic Studio experience).
"Disabled": apps start without the shared home directory -- pair with space storage
  or custom file systems when home directories are not wanted.
(AWS's third value, "DefaultAsDomain", is only valid on per-user profiles -- it means
"inherit this domain-level setting" and is rejected here at the domain level.)

### spec.defaultUserSettings.jupyterLabAppSettings

`AwsSagemakerDomainJupyterLabAppSettings`

jupyter_lab_app_settings configures JupyterLab, the primary IDE for SageMaker Studio.
JupyterLab provides a modern notebook and code editing experience with built-in Git
integration, terminal access, extension support, and collaborative editing.

### spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new JupyterLab apps. Users can override the instance type at app creation time.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultUserSettings.jupyterLabAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts that run when
JupyterLab apps start. Use lifecycle configs to install Python packages, configure
JupyterLab extensions, set environment variables, or mount additional storage.

### spec.defaultUserSettings.jupyterLabAppSettings.builtInLifecycleConfigArn

`string`

built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
configuration to run at app start, as opposed to the customer-authored scripts in
lifecycle_config_arns.

### spec.defaultUserSettings.jupyterLabAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as JupyterLab kernels.
Each image must be registered in SageMaker via an AppImageConfig that defines
the kernel specification and file system layout. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.defaultUserSettings.jupyterLabAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.jupyterLabAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.jupyterLabAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.defaultUserSettings.jupyterLabAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

code_repositories are Git repositories automatically cloned into JupyterLab on startup.
Provides immediate access to team code, shared notebooks, and documentation.
Maximum 10 repositories.

- rule: {"repeated":{"maxItems":"10"}}

### spec.defaultUserSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultUserSettings.jupyterLabAppSettings.idleSettings

`AwsSagemakerDomainIdleSettings`

idle_settings configures automatic shutdown of idle JupyterLab instances.
Critical for cost management: without idle timeout, instances run 24/7 at full
compute cost even when no user is interacting with them.

- rule: max_idle_timeout_in_minutes must be >= min_idle_timeout_in_minutes
- rule: lifecycle_management must be 'ENABLED' or 'DISABLED'

### spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.lifecycleManagement

`string` · optional (explicit presence)

lifecycle_management is the enforcement switch: "ENABLED" (the default when
this block is present) turns automatic idle shutdown on; "DISABLED" keeps
the block's timeout values as published guardrails users may adopt WITHOUT
forcing auto-shutdown on — the defined-but-disabled state. Both engines
send the explicit value, so flipping to "DISABLED" genuinely turns
enforcement off.

- default: `ENABLED`

### spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes

`int32` · required

idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
is automatically shut down. Range: 60-525600 (1 hour to 365 days).
A reasonable production default is 120 (2 hours).

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.minIdleTimeoutInMinutes

`int32` · required

min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
configure for their own apps. Prevents users from setting excessively short timeouts
that would cause disruptive shutdowns during brief pauses. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultUserSettings.jupyterLabAppSettings.idleSettings.maxIdleTimeoutInMinutes

`int32` · required

max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
configure. Prevents users from effectively disabling idle shutdown by setting
extremely long timeouts. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultUserSettings.jupyterLabAppSettings.emrSettings

`AwsSagemakerDomainEmrSettings`

emr_settings pre-authorizes the IAM roles JupyterLab uses to discover and connect
to Amazon EMR clusters for large-scale data processing directly from notebooks.

### spec.defaultUserSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns

`[]string | valueFrom`

assumable_role_arns are IAM roles the JupyterLab app can assume to CONNECT to EMR
clusters -- including clusters in other AWS accounts (cross-account access).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultUserSettings.jupyterLabAppSettings.emrSettings.executionRoleArns

`[]string | valueFrom`

execution_role_arns are IAM runtime roles available for EMR workloads submitted from
the notebook (the role the EMR job itself runs under).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultUserSettings.jupyterServerAppSettings

`AwsSagemakerDomainJupyterServerAppSettings`

jupyter_server_app_settings configures the classic Jupyter Server app (Studio
Classic). Teams still running the previous-generation Studio experience configure
its default resources and startup repositories here.

### spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default configuration for the Jupyter Server app.
Jupyter Server runs on a lightweight system-managed instance ("system" instance type).

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultUserSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultUserSettings.jupyterServerAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
Jupyter Server app.

### spec.defaultUserSettings.jupyterServerAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

code_repositories are Git repositories automatically cloned on startup.
Maximum 10 repositories.

- rule: {"repeated":{"maxItems":"10"}}

### spec.defaultUserSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultUserSettings.kernelGatewayAppSettings

`AwsSagemakerDomainKernelGatewayAppSettings`

kernel_gateway_app_settings configures KernelGateway apps that provide custom compute
kernels. Use KernelGateway to bring your own Docker images with specialized ML
frameworks, custom libraries, or GPU-optimized environments that go beyond the
standard SageMaker-provided kernels.

### spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new KernelGateway apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultUserSettings.kernelGatewayAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.

### spec.defaultUserSettings.kernelGatewayAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as KernelGateway kernels.
Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.defaultUserSettings.kernelGatewayAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.kernelGatewayAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.defaultUserSettings.codeEditorAppSettings

`AwsSagemakerDomainCodeEditorAppSettings`

code_editor_app_settings configures the Code Editor app -- SageMaker's VS Code
(Code-OSS) IDE. It carries the same resource-spec/lifecycle/custom-image/idle
controls as JupyterLab, for teams that prefer a full code editor over notebooks.

### spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new Code Editor apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultUserSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultUserSettings.codeEditorAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for Code Editor apps.

### spec.defaultUserSettings.codeEditorAppSettings.builtInLifecycleConfigArn

`string`

built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
configuration to run at app start.

### spec.defaultUserSettings.codeEditorAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available in Code Editor. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.defaultUserSettings.codeEditorAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.codeEditorAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.codeEditorAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.defaultUserSettings.codeEditorAppSettings.idleSettings

`AwsSagemakerDomainIdleSettings`

idle_settings configures automatic shutdown of idle Code Editor instances -- the
same cost-control dial JupyterLab carries.

- rule: max_idle_timeout_in_minutes must be >= min_idle_timeout_in_minutes
- rule: lifecycle_management must be 'ENABLED' or 'DISABLED'

### spec.defaultUserSettings.codeEditorAppSettings.idleSettings.lifecycleManagement

`string` · optional (explicit presence)

lifecycle_management is the enforcement switch: "ENABLED" (the default when
this block is present) turns automatic idle shutdown on; "DISABLED" keeps
the block's timeout values as published guardrails users may adopt WITHOUT
forcing auto-shutdown on — the defined-but-disabled state. Both engines
send the explicit value, so flipping to "DISABLED" genuinely turns
enforcement off.

- default: `ENABLED`

### spec.defaultUserSettings.codeEditorAppSettings.idleSettings.idleTimeoutInMinutes

`int32` · required

idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
is automatically shut down. Range: 60-525600 (1 hour to 365 days).
A reasonable production default is 120 (2 hours).

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultUserSettings.codeEditorAppSettings.idleSettings.minIdleTimeoutInMinutes

`int32` · required

min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
configure for their own apps. Prevents users from setting excessively short timeouts
that would cause disruptive shutdowns during brief pauses. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultUserSettings.codeEditorAppSettings.idleSettings.maxIdleTimeoutInMinutes

`int32` · required

max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
configure. Prevents users from effectively disabling idle shutdown by setting
extremely long timeouts. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultUserSettings.tensorBoardAppSettings

`AwsSagemakerDomainTensorBoardAppSettings`

tensor_board_app_settings configures the TensorBoard app used to visualize
training runs. Only the default resource spec is configurable.

### spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for the TensorBoard app.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultUserSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultUserSettings.rSessionAppSettings

`AwsSagemakerDomainRSessionAppSettings`

r_session_app_settings configures RSession apps (the R kernels backing RStudio
sessions). Requires RStudio to be enabled on the domain via
r_studio_server_pro_domain_settings.

### spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for RSession apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultUserSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultUserSettings.rSessionAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as RSession kernels. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.defaultUserSettings.rSessionAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.rSessionAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.defaultUserSettings.rSessionAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.defaultUserSettings.rStudioServerProAppSettings

`AwsSagemakerDomainRStudioServerProAppSettings`

r_studio_server_pro_app_settings controls per-user RStudio Server Pro access.
Requires RStudio to be enabled on the domain via r_studio_server_pro_domain_settings.

- rule: access_status must be 'ENABLED' or 'DISABLED'
- rule: user_group must be 'R_STUDIO_ADMIN' or 'R_STUDIO_USER'
- rule: user_group is only honored when access_status is 'ENABLED'

### spec.defaultUserSettings.rStudioServerProAppSettings.accessStatus

`string`

access_status grants or denies the user access to RStudio Server Pro.
"ENABLED": the user sees and can launch RStudio.
"DISABLED": RStudio is hidden for the user.

### spec.defaultUserSettings.rStudioServerProAppSettings.userGroup

`string`

user_group assigns the RStudio authorization level.
"R_STUDIO_ADMIN": administrative access to the RStudio Workbench admin dashboard.
"R_STUDIO_USER" (AWS default): regular RStudio user.
Only meaningful when access_status is "ENABLED" -- AWS ignores it otherwise, so the
spec rejects the dead combination up front.

### spec.defaultUserSettings.canvasAppSettings

`AwsSagemakerDomainCanvasAppSettings`

canvas_app_settings configures SageMaker Canvas, the no-code ML workspace.
Each sub-block is an independent Canvas capability (direct model deployment,
EMR Serverless big-data processing, Bedrock generative AI, SaaS data connectors,
Kendra document search, cross-account model registration, time-series forecasting,
and the shared artifact workspace).

- rule: direct_deploy_status must be 'ENABLED' or 'DISABLED'
- rule: kendra_settings_status must be 'ENABLED' or 'DISABLED'

### spec.defaultUserSettings.canvasAppSettings.directDeployStatus

`string` · optional (explicit presence)

direct_deploy_status controls whether Canvas users can deploy models they build
directly to SageMaker real-time endpoints ("ENABLED" or "DISABLED"). Direct
deployment creates billable endpoints, so governance-minded teams often disable it
and route models through the registry instead (model_register_settings).

### spec.defaultUserSettings.canvasAppSettings.emrServerlessSettings

`AwsSagemakerDomainCanvasEmrServerlessSettings`

emr_serverless_settings lets Canvas run large data preparation and processing jobs
on EMR Serverless.

- rule: status must be 'ENABLED' or 'DISABLED'

### spec.defaultUserSettings.canvasAppSettings.emrServerlessSettings.executionRoleArn

`string | valueFrom`

execution_role_arn is the IAM role Canvas uses to submit and manage EMR Serverless
jobs.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultUserSettings.canvasAppSettings.emrServerlessSettings.status

`string`

status enables or disables the capability ("ENABLED" or "DISABLED").

### spec.defaultUserSettings.canvasAppSettings.generativeAiBedrockRoleArn

`string | valueFrom`

generative_ai_bedrock_role_arn is the IAM role Canvas assumes to call Amazon Bedrock
for generative-AI features. Setting the role is what enables the capability.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings

`[]AwsSagemakerDomainCanvasIdentityProviderOauthSettings`

identity_provider_oauth_settings wire Canvas to external SaaS data sources
(Salesforce Data Cloud, Snowflake) through OAuth. Each entry names the data source
and the Secrets Manager secret holding its OAuth client credentials. Maximum 20.

- rule: {"repeated":{"maxItems":"20"}}
- rule: data_source_name must be 'SalesforceGenie' or 'Snowflake'
- rule: status must be 'ENABLED' or 'DISABLED'

### spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings[].dataSourceName

`string`

data_source_name identifies the SaaS data source.
"SalesforceGenie": Salesforce Data Cloud.
"Snowflake": Snowflake.

### spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings[].secretArn

`string` · required

secret_arn is the Secrets Manager secret holding the data source's OAuth client
credentials (client ID and secret). The ARN is a reference Canvas resolves at
connection time -- never secret material itself.

- rule: {"required":true}

### spec.defaultUserSettings.canvasAppSettings.identityProviderOauthSettings[].status

`string`

status enables or disables this connector ("ENABLED" or "DISABLED").

### spec.defaultUserSettings.canvasAppSettings.kendraSettingsStatus

`string` · optional (explicit presence)

kendra_settings_status controls whether Canvas can query Amazon Kendra indexes for
document search ("ENABLED" or "DISABLED").

### spec.defaultUserSettings.canvasAppSettings.modelRegisterSettings

`AwsSagemakerDomainCanvasModelRegisterSettings`

model_register_settings controls whether Canvas users can register their models into
a SageMaker Model Registry, optionally in another AWS account.

- rule: status must be 'ENABLED' or 'DISABLED'

### spec.defaultUserSettings.canvasAppSettings.modelRegisterSettings.crossAccountModelRegisterRoleArn

`string`

cross_account_model_register_role_arn is the IAM role Canvas assumes to register
models into a Model Registry that lives in ANOTHER AWS account. Leave unset to
register into this account's registry.

### spec.defaultUserSettings.canvasAppSettings.modelRegisterSettings.status

`string`

status enables or disables model registration ("ENABLED" or "DISABLED").

### spec.defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings

`AwsSagemakerDomainCanvasTimeSeriesForecastingSettings`

time_series_forecasting_settings enables Canvas time-series forecasting, which uses
Amazon Forecast under the hood via the given IAM role.

- rule: status must be 'ENABLED' or 'DISABLED'

### spec.defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings.amazonForecastRoleArn

`string | valueFrom`

amazon_forecast_role_arn is the IAM role Canvas assumes to call Amazon Forecast.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings.status

`string`

status enables or disables forecasting ("ENABLED" or "DISABLED").

### spec.defaultUserSettings.canvasAppSettings.workspaceSettings

`AwsSagemakerDomainCanvasWorkspaceSettings`

workspace_settings pins the S3 location (and optional KMS key) where Canvas stores
its working artifacts -- datasets, intermediate results, generated models.

- rule: s3_artifact_path must be an s3:// or https:// URI

### spec.defaultUserSettings.canvasAppSettings.workspaceSettings.s3ArtifactPath

`string`

s3_artifact_path is the S3 URI where Canvas stores datasets, intermediate results,
and generated models. Example: "s3://my-canvas-workspace/artifacts/".

- rule: {"string":{"maxLen":"1024"}}

### spec.defaultUserSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId

`string | valueFrom`

s3_kms_key_id is the KMS key used to encrypt Canvas artifacts in S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.defaultUserSettings.sharingSettings

`AwsSagemakerDomainSharingSettings`

sharing_settings controls notebook output sharing to S3. When enabled, notebook cell
outputs are persisted to an S3 location, allowing team members to view results
without running the notebook.

- rule: notebook_output_option must be 'Allowed' or 'Disabled'
- rule: s3_output_path is required when notebook_output_option is 'Allowed'

### spec.defaultUserSettings.sharingSettings.notebookOutputOption

`string` · optional (explicit presence)

notebook_output_option controls whether notebook cell outputs are persisted to S3.
"Allowed": outputs are copied to S3 at the location specified by s3_output_path.
"Disabled" (default): outputs are not shared externally.

- default: `Disabled`

### spec.defaultUserSettings.sharingSettings.s3KmsKeyId

`string | valueFrom`

s3_kms_key_id is the KMS key used to encrypt shared notebook outputs in S3.
If omitted, outputs are encrypted with the default S3 bucket encryption.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.defaultUserSettings.sharingSettings.s3OutputPath

`string`

s3_output_path is the S3 URI where shared notebook outputs are stored.
Required when notebook_output_option is "Allowed".
Example: "s3://my-team-bucket/notebook-outputs/"

### spec.defaultUserSettings.spaceStorageSettings

`AwsSagemakerDomainSpaceStorageSettings`

space_storage_settings configures default EBS volume sizes for user spaces.
Spaces use EBS volumes for working storage beyond the shared EFS home directory.

- rule: maximum_ebs_volume_size_in_gb must be >= default_ebs_volume_size_in_gb

### spec.defaultUserSettings.spaceStorageSettings.defaultEbsVolumeSizeInGb

`int32` · required

default_ebs_volume_size_in_gb is the default EBS volume size (in GB) assigned to new spaces.

- rule: {"required":true}

### spec.defaultUserSettings.spaceStorageSettings.maximumEbsVolumeSizeInGb

`int32` · required

maximum_ebs_volume_size_in_gb is the maximum EBS volume size (in GB) that users can request
for their spaces. Must be >= default_ebs_volume_size_in_gb.

- rule: {"required":true}

### spec.defaultUserSettings.customFileSystemConfigs

`[]AwsSagemakerDomainCustomFileSystemConfig`

custom_file_system_configs mount additional file systems (beyond the domain's own
EFS home directories) into every user's apps -- shared datasets, feature stores,
model artifact trees. Each entry names one file system and the path where it mounts.

### spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig

`AwsSagemakerDomainEfsFileSystemConfig` · required

efs_file_system_config mounts an Amazon EFS file system.

- rule: {"required":true}

### spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId

`string | valueFrom` · required

file_system_id is the EFS file system to mount. The file system must be reachable
from the domain's subnets (mount targets + security groups are the file system's
own configuration).

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemPath

`string` · required

file_system_path is the path within the EFS file system to mount into apps.
Example: "/shared/datasets"

- rule: {"required":true}

### spec.defaultUserSettings.customPosixUserConfig

`AwsSagemakerDomainCustomPosixUserConfig`

custom_posix_user_config sets the POSIX identity (UID/GID) that apps run as when
accessing the EFS home directory and custom file systems. Set this when file-system
permissions on shared storage must map to a specific owner instead of SageMaker's
default identity.

### spec.defaultUserSettings.customPosixUserConfig.uid

`int64` · required

uid is the POSIX user ID. Must be at least 10000.

- rule: {"required":true,"int64":{"gte":"10000"}}

### spec.defaultUserSettings.customPosixUserConfig.gid

`int64` · required

gid is the POSIX group ID. Must be at least 1001.

- rule: {"required":true,"int64":{"gte":"1001"}}

### spec.defaultUserSettings.studioWebPortalSettings

`AwsSagemakerDomainStudioWebPortalSettings`

studio_web_portal_settings hides parts of the Studio UI from users -- entire app
types, specific instance types, or ML tools. Use it to keep expensive GPU instance
types or unused tooling out of the picker instead of policing them after the fact.

### spec.defaultUserSettings.studioWebPortalSettings.hiddenAppTypes

`[]string`

hidden_app_types are Studio app types to hide (e.g. "JupyterServer", "KernelGateway",
"Canvas", "CodeEditor", "JupyterLab", "TensorBoard", "RStudioServerPro"). Values are
SageMaker AppType names; AWS validates them at deploy time as the set grows.

### spec.defaultUserSettings.studioWebPortalSettings.hiddenInstanceTypes

`[]string`

hidden_instance_types are instance types to hide from app-creation pickers
(e.g. "ml.p3.2xlarge"). Values are SageMaker app instance type names.

### spec.defaultUserSettings.studioWebPortalSettings.hiddenMlTools

`[]string`

hidden_ml_tools are Studio ML tools to hide (e.g. "DataWrangler", "FeatureStore",
"EmrClusters", "AutoMl", "Experiments", "Pipelines"). Values are SageMaker MlTools
names; AWS validates them at deploy time as the set grows.

### spec.defaultSpaceSettings

`AwsSagemakerDomainDefaultSpaceSettings`

default_space_settings defines the default configuration inherited by all shared
(collaborative) spaces in the domain. Spaces are workspaces multiple users can attach
to; they carry their own execution role and app baselines, separate from the per-user
plane above.

### spec.defaultSpaceSettings.executionRoleArn

`string | valueFrom` · required

execution_role_arn is the IAM role assumed by SageMaker for workloads running in
shared spaces. Like the user-level role, it must trust sagemaker.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultSpaceSettings.securityGroupIds

`[]string | valueFrom`

security_group_ids are security groups attached to the ENIs of apps running in
shared spaces. Maximum 5 security groups.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.defaultSpaceSettings.jupyterLabAppSettings

`AwsSagemakerDomainJupyterLabAppSettings`

jupyter_lab_app_settings is the JupyterLab baseline for shared spaces (same shape
as the user-level block).

### spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new JupyterLab apps. Users can override the instance type at app creation time.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultSpaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultSpaceSettings.jupyterLabAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts that run when
JupyterLab apps start. Use lifecycle configs to install Python packages, configure
JupyterLab extensions, set environment variables, or mount additional storage.

### spec.defaultSpaceSettings.jupyterLabAppSettings.builtInLifecycleConfigArn

`string`

built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
configuration to run at app start, as opposed to the customer-authored scripts in
lifecycle_config_arns.

### spec.defaultSpaceSettings.jupyterLabAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as JupyterLab kernels.
Each image must be registered in SageMaker via an AppImageConfig that defines
the kernel specification and file system layout. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.defaultSpaceSettings.jupyterLabAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.defaultSpaceSettings.jupyterLabAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.defaultSpaceSettings.jupyterLabAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.defaultSpaceSettings.jupyterLabAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

code_repositories are Git repositories automatically cloned into JupyterLab on startup.
Provides immediate access to team code, shared notebooks, and documentation.
Maximum 10 repositories.

- rule: {"repeated":{"maxItems":"10"}}

### spec.defaultSpaceSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings

`AwsSagemakerDomainIdleSettings`

idle_settings configures automatic shutdown of idle JupyterLab instances.
Critical for cost management: without idle timeout, instances run 24/7 at full
compute cost even when no user is interacting with them.

- rule: max_idle_timeout_in_minutes must be >= min_idle_timeout_in_minutes
- rule: lifecycle_management must be 'ENABLED' or 'DISABLED'

### spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.lifecycleManagement

`string` · optional (explicit presence)

lifecycle_management is the enforcement switch: "ENABLED" (the default when
this block is present) turns automatic idle shutdown on; "DISABLED" keeps
the block's timeout values as published guardrails users may adopt WITHOUT
forcing auto-shutdown on — the defined-but-disabled state. Both engines
send the explicit value, so flipping to "DISABLED" genuinely turns
enforcement off.

- default: `ENABLED`

### spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes

`int32` · required

idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
is automatically shut down. Range: 60-525600 (1 hour to 365 days).
A reasonable production default is 120 (2 hours).

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.minIdleTimeoutInMinutes

`int32` · required

min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
configure for their own apps. Prevents users from setting excessively short timeouts
that would cause disruptive shutdowns during brief pauses. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultSpaceSettings.jupyterLabAppSettings.idleSettings.maxIdleTimeoutInMinutes

`int32` · required

max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
configure. Prevents users from effectively disabling idle shutdown by setting
extremely long timeouts. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings

`AwsSagemakerDomainEmrSettings`

emr_settings pre-authorizes the IAM roles JupyterLab uses to discover and connect
to Amazon EMR clusters for large-scale data processing directly from notebooks.

### spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns

`[]string | valueFrom`

assumable_role_arns are IAM roles the JupyterLab app can assume to CONNECT to EMR
clusters -- including clusters in other AWS accounts (cross-account access).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings.executionRoleArns

`[]string | valueFrom`

execution_role_arns are IAM runtime roles available for EMR workloads submitted from
the notebook (the role the EMR job itself runs under).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultSpaceSettings.jupyterServerAppSettings

`AwsSagemakerDomainJupyterServerAppSettings`

jupyter_server_app_settings is the classic Jupyter Server baseline for shared spaces.

### spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default configuration for the Jupyter Server app.
Jupyter Server runs on a lightweight system-managed instance ("system" instance type).

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultSpaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultSpaceSettings.jupyterServerAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
Jupyter Server app.

### spec.defaultSpaceSettings.jupyterServerAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

code_repositories are Git repositories automatically cloned on startup.
Maximum 10 repositories.

- rule: {"repeated":{"maxItems":"10"}}

### spec.defaultSpaceSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultSpaceSettings.kernelGatewayAppSettings

`AwsSagemakerDomainKernelGatewayAppSettings`

kernel_gateway_app_settings is the KernelGateway baseline for shared spaces.

### spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new KernelGateway apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.defaultSpaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.defaultSpaceSettings.kernelGatewayAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.

### spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as KernelGateway kernels.
Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.defaultSpaceSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.defaultSpaceSettings.spaceStorageSettings

`AwsSagemakerDomainSpaceStorageSettings`

space_storage_settings configures the default and maximum EBS volume sizes for
shared spaces.

- rule: maximum_ebs_volume_size_in_gb must be >= default_ebs_volume_size_in_gb

### spec.defaultSpaceSettings.spaceStorageSettings.defaultEbsVolumeSizeInGb

`int32` · required

default_ebs_volume_size_in_gb is the default EBS volume size (in GB) assigned to new spaces.

- rule: {"required":true}

### spec.defaultSpaceSettings.spaceStorageSettings.maximumEbsVolumeSizeInGb

`int32` · required

maximum_ebs_volume_size_in_gb is the maximum EBS volume size (in GB) that users can request
for their spaces. Must be >= default_ebs_volume_size_in_gb.

- rule: {"required":true}

### spec.defaultSpaceSettings.customFileSystemConfigs

`[]AwsSagemakerDomainCustomFileSystemConfig`

custom_file_system_configs mount additional file systems into every shared space's
apps (same shape as the user-level block).

### spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig

`AwsSagemakerDomainEfsFileSystemConfig` · required

efs_file_system_config mounts an Amazon EFS file system.

- rule: {"required":true}

### spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId

`string | valueFrom` · required

file_system_id is the EFS file system to mount. The file system must be reachable
from the domain's subnets (mount targets + security groups are the file system's
own configuration).

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemPath

`string` · required

file_system_path is the path within the EFS file system to mount into apps.
Example: "/shared/datasets"

- rule: {"required":true}

### spec.defaultSpaceSettings.customPosixUserConfig

`AwsSagemakerDomainCustomPosixUserConfig`

custom_posix_user_config sets the POSIX identity that shared-space apps run as.

### spec.defaultSpaceSettings.customPosixUserConfig.uid

`int64` · required

uid is the POSIX user ID. Must be at least 10000.

- rule: {"required":true,"int64":{"gte":"10000"}}

### spec.defaultSpaceSettings.customPosixUserConfig.gid

`int64` · required

gid is the POSIX group ID. Must be at least 1001.

- rule: {"required":true,"int64":{"gte":"1001"}}

### spec.domainSecurityGroupIds

`[]string | valueFrom`

domain_security_group_ids are security groups applied at the domain level for
domain-scoped apps and shared resources. These are separate from user-level security
groups (default_user_settings.security_group_ids) and control domain-wide network
boundaries. Maximum 3 security groups.
ForceNew: changing these forces domain replacement.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"3"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.dockerSettings

`AwsSagemakerDomainDockerSettings`

docker_settings controls Docker access within the SageMaker Domain.
When enabled, users can build and run custom Docker containers directly in their
notebooks and terminals -- essential for custom training containers, inference
endpoints, and reproducible ML pipelines.

- rule: enable_docker_access must be 'ENABLED' or 'DISABLED'

### spec.dockerSettings.enableDockerAccess

`string`

enable_docker_access controls whether users can build and run Docker containers
in their notebooks and terminals.
"ENABLED": Docker commands (build, pull, run) are available.
"DISABLED": Docker access is blocked.

### spec.dockerSettings.vpcOnlyTrustedAccounts

`[]string`

vpc_only_trusted_accounts restricts Docker image pulling to images from specified
AWS account IDs when app_network_access_type is "VpcOnly". This prevents users from
pulling arbitrary images from public registries, enforcing approved image sources.
Maximum 20 account IDs, each exactly 12 digits — a malformed entry would make this
security control silently not match the intended account.

- rule: {"repeated":{"maxItems":"20","items":{"string":{"pattern":"^[0-9]{12}$"}}}}

### spec.executionRoleIdentityConfig

`string` · optional (explicit presence)

execution_role_identity_config controls how user identity appears in AWS CloudTrail
and in the credentials SageMaker vends to apps.
"USER_PROFILE_NAME": the sts:SourceIdentity of every session is set to the user
  profile name, so CloudTrail events and IAM policies can distinguish WHICH user
  acted through the shared execution role -- strongly recommended for auditability.
"DISABLED": sessions carry only the execution role identity.

### spec.rStudioServerProDomainSettings

`AwsSagemakerDomainRStudioServerProDomainSettings`

r_studio_server_pro_domain_settings enables RStudio (Posit) Workbench on the domain.
RStudio on SageMaker requires a Posit license purchased through AWS License Manager;
configuring this block is what activates the RStudio app plane for the domain.

### spec.rStudioServerProDomainSettings.domainExecutionRoleArn

`string | valueFrom` · required

domain_execution_role_arn is the IAM role SageMaker assumes to run the RStudio
server for the domain (license validation, server lifecycle). Distinct from the
per-user execution role.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.rStudioServerProDomainSettings.rStudioConnectUrl

`string`

r_studio_connect_url is the URL of an RStudio Connect server where users publish
Shiny apps and R Markdown documents from their sessions.

### spec.rStudioServerProDomainSettings.rStudioPackageManagerUrl

`string`

r_studio_package_manager_url is the URL of an RStudio Package Manager server that
sessions resolve R packages from (instead of public CRAN).

### spec.rStudioServerProDomainSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the compute instance type and image configuration for the
RStudio server app itself.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.rStudioServerProDomainSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.rStudioServerProDomainSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.rStudioServerProDomainSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.rStudioServerProDomainSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.rStudioServerProDomainSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.trustedIdentityPropagationStatus

`string` · optional (explicit presence)

trusted_identity_propagation_status enables trusted identity propagation, which
forwards the IAM Identity Center user identity through SageMaker to downstream
AWS analytics services (Athena, Redshift, Lake Formation), so data-level permissions
apply per human user instead of per execution role.
Requires auth_mode "SSO" with ANY value: AWS rejects the setting outright on
IAM-auth domains ("TrustedIdentityPropagationSettings is only supported for
Domains with AWS IAM Identity Center enabled"), even when set to "DISABLED".
Leave unset on IAM domains.

### spec.userProfiles

`[]AwsSagemakerDomainUserProfile`

user_profiles are the per-person workspaces inside the domain, keyed by
name. Each profile inherits `default_user_settings` and may override any
of it via its own `user_settings`. Profiles add/remove in place as the
list changes; removing an entry deletes that profile AND its apps and
data surfaces, so treat removals as destructive. Profiles created
outside this manifest (IAM Identity Center auto-provisioning, console)
are not managed or removed by it.

- rule: single_sign_on_user_identifier and single_sign_on_user_value must be set together

### spec.userProfiles[].userProfileName

`string` · required

The profile name — unique within the domain, and the key both IaC modules
use for the satellite resource. 1-63 characters: alphanumeric and hyphens,
starting and ending alphanumeric. ForceNew: renaming replaces the profile
(and its home directory association).

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$"}}

### spec.userProfiles[].singleSignOnUserIdentifier

`string`

For SSO-auth domains: the IAM Identity Center attribute that identifies the
user. AWS supports exactly one identifier scheme, "UserName". Set together
with `single_sign_on_user_value`. ForceNew.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"const":"UserName"}}

### spec.userProfiles[].singleSignOnUserValue

`string`

For SSO-auth domains: the Identity Center username this profile belongs to.
Set together with `single_sign_on_user_identifier`. ForceNew.

### spec.userProfiles[].userSettings

`AwsSagemakerDomainUserSettings`

Per-user overrides of the domain's `default_user_settings` — the same
settings tree, applied on top of the domain baseline. Unset means the
profile inherits the baseline unchanged. `execution_role_arn` is required
inside this block whenever it is set (the AWS contract for UserSettings).

- rule: studio_web_portal must be 'ENABLED' or 'DISABLED'
- rule: auto_mount_home_efs must be 'Enabled' or 'Disabled' at the domain level

### spec.userProfiles[].userSettings.executionRoleArn

`string | valueFrom` · required

execution_role_arn is the IAM role assumed by SageMaker when running notebooks,
training jobs, and other ML workloads on behalf of users. This role determines what
AWS resources (S3 buckets, ECR repos, Secrets Manager, etc.) users can access from
their Studio sessions. The role must have a trust policy allowing
sagemaker.amazonaws.com to assume it.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.securityGroupIds

`[]string | valueFrom`

security_group_ids are user-level security groups controlling network access for
notebook instances and apps. These are attached to the ENIs created for each user's
apps and control inbound/outbound traffic at the user level. Maximum 5 security groups.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.userProfiles[].userSettings.defaultLandingUri

`string`

default_landing_uri is the URI of the default app opened when a user accesses the domain.
Common values:
  "studio::relative/JupyterLab" - opens JupyterLab (recommended for most teams)
  "studio::relative/JupyterServer:" - opens classic Jupyter Server
  "studio::" - opens SageMaker Studio home
If omitted, AWS uses the platform default.

### spec.userProfiles[].userSettings.studioWebPortal

`string` · optional (explicit presence)

studio_web_portal controls whether the SageMaker Studio web portal is accessible.
"ENABLED" (default): users can access the full Studio web interface.
"DISABLED": restricts access to programmatic-only usage (API/CLI).

- default: `ENABLED`

### spec.userProfiles[].userSettings.autoMountHomeEfs

`string` · optional (explicit presence)

auto_mount_home_efs controls whether each user's home directory on the domain's EFS
file system is automatically mounted into their apps.
"Enabled": home directories mount automatically (the classic Studio experience).
"Disabled": apps start without the shared home directory -- pair with space storage
  or custom file systems when home directories are not wanted.
(AWS's third value, "DefaultAsDomain", is only valid on per-user profiles -- it means
"inherit this domain-level setting" and is rejected here at the domain level.)

### spec.userProfiles[].userSettings.jupyterLabAppSettings

`AwsSagemakerDomainJupyterLabAppSettings`

jupyter_lab_app_settings configures JupyterLab, the primary IDE for SageMaker Studio.
JupyterLab provides a modern notebook and code editing experience with built-in Git
integration, terminal access, extension support, and collaborative editing.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new JupyterLab apps. Users can override the instance type at app creation time.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts that run when
JupyterLab apps start. Use lifecycle configs to install Python packages, configure
JupyterLab extensions, set environment variables, or mount additional storage.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.builtInLifecycleConfigArn

`string`

built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
configuration to run at app start, as opposed to the customer-authored scripts in
lifecycle_config_arns.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as JupyterLab kernels.
Each image must be registered in SageMaker via an AppImageConfig that defines
the kernel specification and file system layout. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

code_repositories are Git repositories automatically cloned into JupyterLab on startup.
Provides immediate access to team code, shared notebooks, and documentation.
Maximum 10 repositories.

- rule: {"repeated":{"maxItems":"10"}}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings

`AwsSagemakerDomainIdleSettings`

idle_settings configures automatic shutdown of idle JupyterLab instances.
Critical for cost management: without idle timeout, instances run 24/7 at full
compute cost even when no user is interacting with them.

- rule: max_idle_timeout_in_minutes must be >= min_idle_timeout_in_minutes
- rule: lifecycle_management must be 'ENABLED' or 'DISABLED'

### spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.lifecycleManagement

`string` · optional (explicit presence)

lifecycle_management is the enforcement switch: "ENABLED" (the default when
this block is present) turns automatic idle shutdown on; "DISABLED" keeps
the block's timeout values as published guardrails users may adopt WITHOUT
forcing auto-shutdown on — the defined-but-disabled state. Both engines
send the explicit value, so flipping to "DISABLED" genuinely turns
enforcement off.

- default: `ENABLED`

### spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes

`int32` · required

idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
is automatically shut down. Range: 60-525600 (1 hour to 365 days).
A reasonable production default is 120 (2 hours).

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.minIdleTimeoutInMinutes

`int32` · required

min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
configure for their own apps. Prevents users from setting excessively short timeouts
that would cause disruptive shutdowns during brief pauses. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.idleSettings.maxIdleTimeoutInMinutes

`int32` · required

max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
configure. Prevents users from effectively disabling idle shutdown by setting
extremely long timeouts. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings

`AwsSagemakerDomainEmrSettings`

emr_settings pre-authorizes the IAM roles JupyterLab uses to discover and connect
to Amazon EMR clusters for large-scale data processing directly from notebooks.

### spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns

`[]string | valueFrom`

assumable_role_arns are IAM roles the JupyterLab app can assume to CONNECT to EMR
clusters -- including clusters in other AWS accounts (cross-account access).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings.executionRoleArns

`[]string | valueFrom`

execution_role_arns are IAM runtime roles available for EMR workloads submitted from
the notebook (the role the EMR job itself runs under).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.jupyterServerAppSettings

`AwsSagemakerDomainJupyterServerAppSettings`

jupyter_server_app_settings configures the classic Jupyter Server app (Studio
Classic). Teams still running the previous-generation Studio experience configure
its default resources and startup repositories here.

### spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default configuration for the Jupyter Server app.
Jupyter Server runs on a lightweight system-managed instance ("system" instance type).

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.userProfiles[].userSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.userProfiles[].userSettings.jupyterServerAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
Jupyter Server app.

### spec.userProfiles[].userSettings.jupyterServerAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

code_repositories are Git repositories automatically cloned on startup.
Maximum 10 repositories.

- rule: {"repeated":{"maxItems":"10"}}

### spec.userProfiles[].userSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.userProfiles[].userSettings.kernelGatewayAppSettings

`AwsSagemakerDomainKernelGatewayAppSettings`

kernel_gateway_app_settings configures KernelGateway apps that provide custom compute
kernels. Use KernelGateway to bring your own Docker images with specialized ML
frameworks, custom libraries, or GPU-optimized environments that go beyond the
standard SageMaker-provided kernels.

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new KernelGateway apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as KernelGateway kernels.
Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.userProfiles[].userSettings.codeEditorAppSettings

`AwsSagemakerDomainCodeEditorAppSettings`

code_editor_app_settings configures the Code Editor app -- SageMaker's VS Code
(Code-OSS) IDE. It carries the same resource-spec/lifecycle/custom-image/idle
controls as JupyterLab, for teams that prefer a full code editor over notebooks.

### spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new Code Editor apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.userProfiles[].userSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.userProfiles[].userSettings.codeEditorAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for Code Editor apps.

### spec.userProfiles[].userSettings.codeEditorAppSettings.builtInLifecycleConfigArn

`string`

built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
configuration to run at app start.

### spec.userProfiles[].userSettings.codeEditorAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available in Code Editor. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.userProfiles[].userSettings.codeEditorAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.codeEditorAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.codeEditorAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings

`AwsSagemakerDomainIdleSettings`

idle_settings configures automatic shutdown of idle Code Editor instances -- the
same cost-control dial JupyterLab carries.

- rule: max_idle_timeout_in_minutes must be >= min_idle_timeout_in_minutes
- rule: lifecycle_management must be 'ENABLED' or 'DISABLED'

### spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.lifecycleManagement

`string` · optional (explicit presence)

lifecycle_management is the enforcement switch: "ENABLED" (the default when
this block is present) turns automatic idle shutdown on; "DISABLED" keeps
the block's timeout values as published guardrails users may adopt WITHOUT
forcing auto-shutdown on — the defined-but-disabled state. Both engines
send the explicit value, so flipping to "DISABLED" genuinely turns
enforcement off.

- default: `ENABLED`

### spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.idleTimeoutInMinutes

`int32` · required

idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
is automatically shut down. Range: 60-525600 (1 hour to 365 days).
A reasonable production default is 120 (2 hours).

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.minIdleTimeoutInMinutes

`int32` · required

min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
configure for their own apps. Prevents users from setting excessively short timeouts
that would cause disruptive shutdowns during brief pauses. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.userProfiles[].userSettings.codeEditorAppSettings.idleSettings.maxIdleTimeoutInMinutes

`int32` · required

max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
configure. Prevents users from effectively disabling idle shutdown by setting
extremely long timeouts. Range: 60-525600.

- rule: {"required":true,"int32":{"lte":525600,"gte":60}}

### spec.userProfiles[].userSettings.tensorBoardAppSettings

`AwsSagemakerDomainTensorBoardAppSettings`

tensor_board_app_settings configures the TensorBoard app used to visualize
training runs. Only the default resource spec is configurable.

### spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for the TensorBoard app.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.userProfiles[].userSettings.tensorBoardAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.userProfiles[].userSettings.rSessionAppSettings

`AwsSagemakerDomainRSessionAppSettings`

r_session_app_settings configures RSession apps (the R kernels backing RStudio
sessions). Requires RStudio to be enabled on the domain via
r_studio_server_pro_domain_settings.

### spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for RSession apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.userProfiles[].userSettings.rSessionAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.userProfiles[].userSettings.rSessionAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as RSession kernels. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.userProfiles[].userSettings.rSessionAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.rSessionAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.userProfiles[].userSettings.rSessionAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.userProfiles[].userSettings.rStudioServerProAppSettings

`AwsSagemakerDomainRStudioServerProAppSettings`

r_studio_server_pro_app_settings controls per-user RStudio Server Pro access.
Requires RStudio to be enabled on the domain via r_studio_server_pro_domain_settings.

- rule: access_status must be 'ENABLED' or 'DISABLED'
- rule: user_group must be 'R_STUDIO_ADMIN' or 'R_STUDIO_USER'
- rule: user_group is only honored when access_status is 'ENABLED'

### spec.userProfiles[].userSettings.rStudioServerProAppSettings.accessStatus

`string`

access_status grants or denies the user access to RStudio Server Pro.
"ENABLED": the user sees and can launch RStudio.
"DISABLED": RStudio is hidden for the user.

### spec.userProfiles[].userSettings.rStudioServerProAppSettings.userGroup

`string`

user_group assigns the RStudio authorization level.
"R_STUDIO_ADMIN": administrative access to the RStudio Workbench admin dashboard.
"R_STUDIO_USER" (AWS default): regular RStudio user.
Only meaningful when access_status is "ENABLED" -- AWS ignores it otherwise, so the
spec rejects the dead combination up front.

### spec.userProfiles[].userSettings.canvasAppSettings

`AwsSagemakerDomainCanvasAppSettings`

canvas_app_settings configures SageMaker Canvas, the no-code ML workspace.
Each sub-block is an independent Canvas capability (direct model deployment,
EMR Serverless big-data processing, Bedrock generative AI, SaaS data connectors,
Kendra document search, cross-account model registration, time-series forecasting,
and the shared artifact workspace).

- rule: direct_deploy_status must be 'ENABLED' or 'DISABLED'
- rule: kendra_settings_status must be 'ENABLED' or 'DISABLED'

### spec.userProfiles[].userSettings.canvasAppSettings.directDeployStatus

`string` · optional (explicit presence)

direct_deploy_status controls whether Canvas users can deploy models they build
directly to SageMaker real-time endpoints ("ENABLED" or "DISABLED"). Direct
deployment creates billable endpoints, so governance-minded teams often disable it
and route models through the registry instead (model_register_settings).

### spec.userProfiles[].userSettings.canvasAppSettings.emrServerlessSettings

`AwsSagemakerDomainCanvasEmrServerlessSettings`

emr_serverless_settings lets Canvas run large data preparation and processing jobs
on EMR Serverless.

- rule: status must be 'ENABLED' or 'DISABLED'

### spec.userProfiles[].userSettings.canvasAppSettings.emrServerlessSettings.executionRoleArn

`string | valueFrom`

execution_role_arn is the IAM role Canvas uses to submit and manage EMR Serverless
jobs.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.canvasAppSettings.emrServerlessSettings.status

`string`

status enables or disables the capability ("ENABLED" or "DISABLED").

### spec.userProfiles[].userSettings.canvasAppSettings.generativeAiBedrockRoleArn

`string | valueFrom`

generative_ai_bedrock_role_arn is the IAM role Canvas assumes to call Amazon Bedrock
for generative-AI features. Setting the role is what enables the capability.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings

`[]AwsSagemakerDomainCanvasIdentityProviderOauthSettings`

identity_provider_oauth_settings wire Canvas to external SaaS data sources
(Salesforce Data Cloud, Snowflake) through OAuth. Each entry names the data source
and the Secrets Manager secret holding its OAuth client credentials. Maximum 20.

- rule: {"repeated":{"maxItems":"20"}}
- rule: data_source_name must be 'SalesforceGenie' or 'Snowflake'
- rule: status must be 'ENABLED' or 'DISABLED'

### spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings[].dataSourceName

`string`

data_source_name identifies the SaaS data source.
"SalesforceGenie": Salesforce Data Cloud.
"Snowflake": Snowflake.

### spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings[].secretArn

`string` · required

secret_arn is the Secrets Manager secret holding the data source's OAuth client
credentials (client ID and secret). The ARN is a reference Canvas resolves at
connection time -- never secret material itself.

- rule: {"required":true}

### spec.userProfiles[].userSettings.canvasAppSettings.identityProviderOauthSettings[].status

`string`

status enables or disables this connector ("ENABLED" or "DISABLED").

### spec.userProfiles[].userSettings.canvasAppSettings.kendraSettingsStatus

`string` · optional (explicit presence)

kendra_settings_status controls whether Canvas can query Amazon Kendra indexes for
document search ("ENABLED" or "DISABLED").

### spec.userProfiles[].userSettings.canvasAppSettings.modelRegisterSettings

`AwsSagemakerDomainCanvasModelRegisterSettings`

model_register_settings controls whether Canvas users can register their models into
a SageMaker Model Registry, optionally in another AWS account.

- rule: status must be 'ENABLED' or 'DISABLED'

### spec.userProfiles[].userSettings.canvasAppSettings.modelRegisterSettings.crossAccountModelRegisterRoleArn

`string`

cross_account_model_register_role_arn is the IAM role Canvas assumes to register
models into a Model Registry that lives in ANOTHER AWS account. Leave unset to
register into this account's registry.

### spec.userProfiles[].userSettings.canvasAppSettings.modelRegisterSettings.status

`string`

status enables or disables model registration ("ENABLED" or "DISABLED").

### spec.userProfiles[].userSettings.canvasAppSettings.timeSeriesForecastingSettings

`AwsSagemakerDomainCanvasTimeSeriesForecastingSettings`

time_series_forecasting_settings enables Canvas time-series forecasting, which uses
Amazon Forecast under the hood via the given IAM role.

- rule: status must be 'ENABLED' or 'DISABLED'

### spec.userProfiles[].userSettings.canvasAppSettings.timeSeriesForecastingSettings.amazonForecastRoleArn

`string | valueFrom`

amazon_forecast_role_arn is the IAM role Canvas assumes to call Amazon Forecast.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.canvasAppSettings.timeSeriesForecastingSettings.status

`string`

status enables or disables forecasting ("ENABLED" or "DISABLED").

### spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings

`AwsSagemakerDomainCanvasWorkspaceSettings`

workspace_settings pins the S3 location (and optional KMS key) where Canvas stores
its working artifacts -- datasets, intermediate results, generated models.

- rule: s3_artifact_path must be an s3:// or https:// URI

### spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings.s3ArtifactPath

`string`

s3_artifact_path is the S3 URI where Canvas stores datasets, intermediate results,
and generated models. Example: "s3://my-canvas-workspace/artifacts/".

- rule: {"string":{"maxLen":"1024"}}

### spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId

`string | valueFrom`

s3_kms_key_id is the KMS key used to encrypt Canvas artifacts in S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.sharingSettings

`AwsSagemakerDomainSharingSettings`

sharing_settings controls notebook output sharing to S3. When enabled, notebook cell
outputs are persisted to an S3 location, allowing team members to view results
without running the notebook.

- rule: notebook_output_option must be 'Allowed' or 'Disabled'
- rule: s3_output_path is required when notebook_output_option is 'Allowed'

### spec.userProfiles[].userSettings.sharingSettings.notebookOutputOption

`string` · optional (explicit presence)

notebook_output_option controls whether notebook cell outputs are persisted to S3.
"Allowed": outputs are copied to S3 at the location specified by s3_output_path.
"Disabled" (default): outputs are not shared externally.

- default: `Disabled`

### spec.userProfiles[].userSettings.sharingSettings.s3KmsKeyId

`string | valueFrom`

s3_kms_key_id is the KMS key used to encrypt shared notebook outputs in S3.
If omitted, outputs are encrypted with the default S3 bucket encryption.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.userProfiles[].userSettings.sharingSettings.s3OutputPath

`string`

s3_output_path is the S3 URI where shared notebook outputs are stored.
Required when notebook_output_option is "Allowed".
Example: "s3://my-team-bucket/notebook-outputs/"

### spec.userProfiles[].userSettings.spaceStorageSettings

`AwsSagemakerDomainSpaceStorageSettings`

space_storage_settings configures default EBS volume sizes for user spaces.
Spaces use EBS volumes for working storage beyond the shared EFS home directory.

- rule: maximum_ebs_volume_size_in_gb must be >= default_ebs_volume_size_in_gb

### spec.userProfiles[].userSettings.spaceStorageSettings.defaultEbsVolumeSizeInGb

`int32` · required

default_ebs_volume_size_in_gb is the default EBS volume size (in GB) assigned to new spaces.

- rule: {"required":true}

### spec.userProfiles[].userSettings.spaceStorageSettings.maximumEbsVolumeSizeInGb

`int32` · required

maximum_ebs_volume_size_in_gb is the maximum EBS volume size (in GB) that users can request
for their spaces. Must be >= default_ebs_volume_size_in_gb.

- rule: {"required":true}

### spec.userProfiles[].userSettings.customFileSystemConfigs

`[]AwsSagemakerDomainCustomFileSystemConfig`

custom_file_system_configs mount additional file systems (beyond the domain's own
EFS home directories) into every user's apps -- shared datasets, feature stores,
model artifact trees. Each entry names one file system and the path where it mounts.

### spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig

`AwsSagemakerDomainEfsFileSystemConfig` · required

efs_file_system_config mounts an Amazon EFS file system.

- rule: {"required":true}

### spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId

`string | valueFrom` · required

file_system_id is the EFS file system to mount. The file system must be reachable
from the domain's subnets (mount targets + security groups are the file system's
own configuration).

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemPath

`string` · required

file_system_path is the path within the EFS file system to mount into apps.
Example: "/shared/datasets"

- rule: {"required":true}

### spec.userProfiles[].userSettings.customPosixUserConfig

`AwsSagemakerDomainCustomPosixUserConfig`

custom_posix_user_config sets the POSIX identity (UID/GID) that apps run as when
accessing the EFS home directory and custom file systems. Set this when file-system
permissions on shared storage must map to a specific owner instead of SageMaker's
default identity.

### spec.userProfiles[].userSettings.customPosixUserConfig.uid

`int64` · required

uid is the POSIX user ID. Must be at least 10000.

- rule: {"required":true,"int64":{"gte":"10000"}}

### spec.userProfiles[].userSettings.customPosixUserConfig.gid

`int64` · required

gid is the POSIX group ID. Must be at least 1001.

- rule: {"required":true,"int64":{"gte":"1001"}}

### spec.userProfiles[].userSettings.studioWebPortalSettings

`AwsSagemakerDomainStudioWebPortalSettings`

studio_web_portal_settings hides parts of the Studio UI from users -- entire app
types, specific instance types, or ML tools. Use it to keep expensive GPU instance
types or unused tooling out of the picker instead of policing them after the fact.

### spec.userProfiles[].userSettings.studioWebPortalSettings.hiddenAppTypes

`[]string`

hidden_app_types are Studio app types to hide (e.g. "JupyterServer", "KernelGateway",
"Canvas", "CodeEditor", "JupyterLab", "TensorBoard", "RStudioServerPro"). Values are
SageMaker AppType names; AWS validates them at deploy time as the set grows.

### spec.userProfiles[].userSettings.studioWebPortalSettings.hiddenInstanceTypes

`[]string`

hidden_instance_types are instance types to hide from app-creation pickers
(e.g. "ml.p3.2xlarge"). Values are SageMaker app instance type names.

### spec.userProfiles[].userSettings.studioWebPortalSettings.hiddenMlTools

`[]string`

hidden_ml_tools are Studio ML tools to hide (e.g. "DataWrangler", "FeatureStore",
"EmrClusters", "AutoMl", "Experiments", "Pipelines"). Values are SageMaker MlTools
names; AWS validates them at deploy time as the set grows.

### spec.spaces

`[]AwsSagemakerDomainSpace`

spaces are named shared (or private) workspaces inside the domain, keyed
by name — the collaboration plane, where a JupyterLab or Code Editor
runtime is shared rather than per-user. Spaces inherit
`default_space_settings` and may override via their own
`space_settings`. Add/remove in place; removing an entry deletes the
space and its EBS volume.

- rule: ownership_settings and space_sharing_settings must be set together — AWS requires both (or neither, for a legacy unscoped space)

### spec.spaces[].spaceName

`string` · required

The space name — unique within the domain, and the key both IaC modules use
for the satellite resource. 1-63 characters: alphanumeric and hyphens,
starting alphanumeric. ForceNew.

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$"}}

### spec.spaces[].displayName

`string`

Display name shown in Studio (the space name itself is immutable; this is
not). Maximum 64 characters. Updates in place.

- rule: {"string":{"maxLen":"64"}}

### spec.spaces[].ownershipSettings

`AwsSagemakerDomainSpaceOwnership`

Ownership: names the user profile that owns this space. Required for
PRIVATE spaces and for shared spaces alike when sharing is configured —
AWS requires ownership and sharing to be declared together (or neither,
which creates a legacy unscoped space). The owner profile must exist in
the domain; it is usually one of `user_profiles`, but profiles provisioned
outside this manifest (SSO auto-provisioning) are equally valid owners.
Not updatable after create — the provider never sends changes to it.

### spec.spaces[].ownershipSettings.ownerUserProfileName

`string` · required

The owning user profile's name (not ARN).

- rule: {"required":true}

### spec.spaces[].spaceSharingSettings

`AwsSagemakerDomainSpaceSharing`

Sharing posture: "Private" (one owner, one runtime) or "Shared"
(collaborative). Declared together with `ownership_settings`. Not
updatable after create.

### spec.spaces[].spaceSharingSettings.sharingType

`string` · required

"Private" or "Shared".

- rule: {"required":true,"string":{"in":["Private","Shared"]}}

### spec.spaces[].spaceSettings

`AwsSagemakerDomainSpaceSettings`

The space's own settings — app type, per-app configuration, storage, and
mounted file systems.

- rule: app_type must be 'JupyterLab', 'CodeEditor', 'JupyterServer', or 'KernelGateway'
- rule: jupyter_server_app_settings.default_resource_spec is required on a space
- rule: kernel_gateway_app_settings.default_resource_spec is required on a space

### spec.spaces[].spaceSettings.appType

`string` · optional (explicit presence)

Which app this space runs: "JupyterLab", "CodeEditor", "JupyterServer",
or "KernelGateway". Spaces are single-app workspaces; the matching
app-settings block below configures it.

### spec.spaces[].spaceSettings.jupyterLabAppSettings

`AwsSagemakerDomainSpaceJupyterLabAppSettings`

JupyterLab configuration for the space. `default_resource_spec` is
required here (unlike the domain baseline) — AWS's space contract.

### spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec` · required

The compute instance type and image for the space's JupyterLab runtime.

- rule: {"required":true}
- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.spaces[].spaceSettings.jupyterLabAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.spaces[].spaceSettings.jupyterLabAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

Git repositories cloned into the space's JupyterLab on startup. Maximum 10.

- rule: {"repeated":{"maxItems":"10"}}

### spec.spaces[].spaceSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.spaces[].spaceSettings.jupyterLabAppSettings.idleSettings

`AwsSagemakerDomainSpaceIdleSettings`

Automatic shutdown of the space's JupyterLab when idle.

### spec.spaces[].spaceSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes

`int32` · optional (explicit presence)

Minutes of inactivity before the space's app shuts down. Range: 60-525600.

- rule: {"int32":{"lte":525600,"gte":60}}

### spec.spaces[].spaceSettings.codeEditorAppSettings

`AwsSagemakerDomainSpaceCodeEditorAppSettings`

Code Editor (VS Code / Code-OSS) configuration for the space.

### spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec` · required

The compute instance type and image for the space's Code Editor runtime.

- rule: {"required":true}
- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.spaces[].spaceSettings.codeEditorAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.spaces[].spaceSettings.codeEditorAppSettings.idleSettings

`AwsSagemakerDomainSpaceIdleSettings`

Automatic shutdown of the space's Code Editor when idle.

### spec.spaces[].spaceSettings.codeEditorAppSettings.idleSettings.idleTimeoutInMinutes

`int32` · optional (explicit presence)

Minutes of inactivity before the space's app shuts down. Range: 60-525600.

- rule: {"int32":{"lte":525600,"gte":60}}

### spec.spaces[].spaceSettings.jupyterServerAppSettings

`AwsSagemakerDomainJupyterServerAppSettings`

Classic Jupyter Server configuration for the space. Same shape as the
domain baseline, but `default_resource_spec` is required here.

### spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default configuration for the Jupyter Server app.
Jupyter Server runs on a lightweight system-managed instance ("system" instance type).

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.spaces[].spaceSettings.jupyterServerAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.spaces[].spaceSettings.jupyterServerAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
Jupyter Server app.

### spec.spaces[].spaceSettings.jupyterServerAppSettings.codeRepositories

`[]AwsSagemakerDomainCodeRepository`

code_repositories are Git repositories automatically cloned on startup.
Maximum 10 repositories.

- rule: {"repeated":{"maxItems":"10"}}

### spec.spaces[].spaceSettings.jupyterServerAppSettings.codeRepositories[].repositoryUrl

`string` · required

repository_url is the HTTPS URL of the Git repository to clone.
Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
Examples: "https://github.com/org/ml-notebooks.git"
Maximum length: 1024 characters.

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.spaces[].spaceSettings.kernelGatewayAppSettings

`AwsSagemakerDomainKernelGatewayAppSettings`

KernelGateway configuration for the space. Same shape as the domain
baseline, but `default_resource_spec` is required here.

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec

`AwsSagemakerDomainResourceSpec`

default_resource_spec sets the default compute instance type and image configuration
for new KernelGateway apps.

- rule: sagemaker_image_version_alias and sagemaker_image_version_arn are mutually exclusive

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType

`string`

instance_type is the EC2 instance type for the app's compute.
Common choices:
  "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
  "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
  "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
  "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
  "system" - lightweight system-managed instance for JupyterServer
The shape is checked here ("system" or an "ml."-prefixed type) so a typo
fails at manifest time; membership in AWS's instance-type list grows with
every hardware launch and is validated by AWS at deploy time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(system|ml\\.[a-z0-9]+(\\.[a-z0-9]+)+)$"}}

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.lifecycleConfigArn

`string`

lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
when this specific app type starts.

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageArn

`string`

sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
Use this to specify a custom base image instead of the SageMaker-provided default.

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionAlias

`string`

sagemaker_image_version_alias is a human-readable alias for a specific image version
(e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.defaultResourceSpec.sagemakerImageVersionArn

`string`

sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
Pins the app to an exact image version for reproducibility across team members.
Mutually exclusive with sagemaker_image_version_alias.

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.lifecycleConfigArns

`[]string`

lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages

`[]AwsSagemakerDomainCustomImage`

custom_images are custom Docker images available as KernelGateway kernels.
Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.

- rule: {"repeated":{"maxItems":"200"}}

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages[].appImageConfigName

`string` · required

app_image_config_name is the name of the SageMaker AppImageConfig that defines how
the image is presented to users (kernel specifications, file system configuration).
The AppImageConfig must exist before referencing it here.

- rule: {"required":true}

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages[].imageName

`string` · required

image_name is the name of the SageMaker Image resource that contains this container image.
The Image resource must exist before referencing it here.

- rule: {"required":true}

### spec.spaces[].spaceSettings.kernelGatewayAppSettings.customImages[].imageVersionNumber

`int32` · optional (explicit presence)

image_version_number pins to a specific version of the image.
If omitted, the latest available version is used.

### spec.spaces[].spaceSettings.customFileSystems

`[]AwsSagemakerDomainSpaceCustomFileSystem`

Existing EFS file systems mounted into the space's apps (by id — the
space form has no per-mount path, unlike the domain baseline's config).

### spec.spaces[].spaceSettings.customFileSystems[].fileSystemId

`string | valueFrom` · required

The EFS file system to mount. The file system must have mount targets in
the domain's VPC.

Containment-exempt: a space MOUNTS the file system; the domain is not
deployed inside it. On a diagram the domain stands in its VPC with a
line to the file system -- the verdict the domain-level
AwsSagemakerDomainEfsFileSystemConfig.file_system_id already carries.

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.spaces[].spaceSettings.spaceStorageSettings

`AwsSagemakerDomainSpaceStorage`

The space's EBS boot/work volume size, in GB (5-16384). Persists across
app restarts; deleted with the space.

### spec.spaces[].spaceSettings.spaceStorageSettings.ebsVolumeSizeInGb

`int32` · required

The EBS volume size in GB. Range: 5-16384.

- rule: {"required":true,"int32":{"lte":16384,"gte":5}}

## Validation Rules

- `auth_mode_valid`: auth_mode must be 'IAM' or 'SSO'
- `app_network_access_type_valid`: app_network_access_type must be 'PublicInternetOnly' or 'VpcOnly'
- `app_security_group_management_valid`: app_security_group_management must be 'Service' or 'Customer'
- `app_security_group_management_requires_rstudio`: app_security_group_management is only honored when r_studio_server_pro_domain_settings is configured
- `tag_propagation_valid`: tag_propagation must be 'ENABLED' or 'DISABLED'
- `home_efs_retention_policy_valid`: home_efs_retention_policy must be 'Retain' or 'Delete'
- `execution_role_identity_config_valid`: execution_role_identity_config must be 'USER_PROFILE_NAME' or 'DISABLED'
- `trusted_identity_propagation_status_valid`: trusted_identity_propagation_status must be 'ENABLED' or 'DISABLED'
- `trusted_identity_propagation_requires_sso`: trusted_identity_propagation_status requires auth_mode 'SSO' (AWS rejects the setting on IAM-auth domains, even 'DISABLED')
- `user_profile_names_unique`: user_profiles names must be unique — each name keys one user profile
- `space_names_unique`: spaces names must be unique — each name keys one space
- `space_idle_requires_owner_lifecycle_jupyterlab`: a space may not set jupyter_lab_app_settings.idle_settings when JupyterLab lifecycle_management resolves 'DISABLED' for its owner profile (the profile's explicit setting, else the domain's default_user_settings) — AWS rejects SpaceIdleSettings on a disabled plane
- `space_idle_requires_owner_lifecycle_codeeditor`: a space may not set code_editor_app_settings.idle_settings when Code Editor lifecycle_management resolves 'DISABLED' for its owner profile (the profile's explicit setting, else the domain's default_user_settings) — AWS rejects SpaceIdleSettings on a disabled plane

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerDomain, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.domain_id` | `string` | domain_id is the unique identifier of the SageMaker Domain. Used in API calls and as a reference when creating user profiles, spaces, and apps. |
| `status.outputs.domain_arn` | `string` | domain_arn is the Amazon Resource Name of the SageMaker Domain. Used in IAM policies, CloudWatch metrics filters, and cross-service references. |
| `status.outputs.domain_url` | `string` | domain_url is the HTTPS URL for accessing the SageMaker Studio web interface. Users navigate to this URL (with their user profile appended) to open Studio. |
| `status.outputs.home_efs_file_system_id` | `string` | home_efs_file_system_id is the ID of the EFS file system automatically created by AWS for user home directories. Each user gets a dedicated directory on this file system. Useful for monitoring, backup policies, and lifecycle management. |
| `status.outputs.security_group_id_for_domain_boundary` | `string` | security_group_id_for_domain_boundary is the ID of the security group that AWS automatically creates to establish the network boundary for the domain. This security group controls cross-app and cross-user traffic within the domain. |
| `status.outputs.single_sign_on_application_arn` | `string` | single_sign_on_application_arn is the ARN of the IAM Identity Center application created for this domain. Only populated when auth_mode is "SSO". Useful for SSO configuration and access management. |
| `status.outputs.single_sign_on_managed_application_instance_id` | `string` | single_sign_on_managed_application_instance_id is the IAM Identity Center managed application instance ID for this domain. Only populated when auth_mode is "SSO". Used when assigning Identity Center users and groups to the domain programmatically. |
| `status.outputs.user_profile_arns` | `map<string, string>` | user_profile_arns maps each folded user profile's name (the spec key) to its ARN. Feeds IAM policies scoped to a profile and the import tooling's per-satellite IDs. |
| `status.outputs.space_arns` | `map<string, string>` | space_arns maps each folded space's name (the spec key) to its ARN — the space's canonical identity (space imports are ARN-keyed). |
| `status.outputs.space_urls` | `map<string, string>` | space_urls maps each folded space's name to the URL where its runtime is reached. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.defaultUserSettings.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultUserSettings.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.defaultUserSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultUserSettings.jupyterLabAppSettings.emrSettings.executionRoleArns` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultUserSettings.canvasAppSettings.emrServerlessSettings.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultUserSettings.canvasAppSettings.generativeAiBedrockRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings.amazonForecastRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultUserSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.defaultUserSettings.sharingSettings.s3KmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |
| `spec.defaultSpaceSettings.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultSpaceSettings.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultSpaceSettings.jupyterLabAppSettings.emrSettings.executionRoleArns` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |
| `spec.domainSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.rStudioServerProDomainSettings.domainExecutionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.userProfiles[].userSettings.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.userProfiles[].userSettings.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns` | AwsIamRole | `status.outputs.role_arn` |
| `spec.userProfiles[].userSettings.jupyterLabAppSettings.emrSettings.executionRoleArns` | AwsIamRole | `status.outputs.role_arn` |
| `spec.userProfiles[].userSettings.canvasAppSettings.emrServerlessSettings.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.userProfiles[].userSettings.canvasAppSettings.generativeAiBedrockRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.userProfiles[].userSettings.canvasAppSettings.timeSeriesForecastingSettings.amazonForecastRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.userProfiles[].userSettings.sharingSettings.s3KmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |
| `spec.spaces[].spaceSettings.customFileSystems[].fileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsSagemakerMlflowApp | `spec.defaultDomainIds` | `status.outputs.domain_id` |

## See Also

- [Overview](../README.md)
