package module

import (
	"github.com/pkg/errors"
	awssagemakerdomainv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awssagemakerdomain/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sagemaker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func domain(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*sagemaker.Domain, error) {
	spec := locals.AwsSagemakerDomain.Spec

	// Resolve subnet IDs
	var subnetIds pulumi.StringArray
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
	}

	args := &sagemaker.DomainArgs{
		DomainName:          pulumi.String(locals.AwsSagemakerDomain.Metadata.Id),
		AuthMode:            pulumi.String(spec.AuthMode),
		VpcId:               pulumi.String(spec.VpcId.GetValue()),
		SubnetIds:           subnetIds,
		DefaultUserSettings: buildDefaultUserSettings(spec.DefaultUserSettings),
		Tags:                pulumi.ToStringMap(locals.Labels),
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	if spec.AppNetworkAccessType != nil && spec.GetAppNetworkAccessType() != "" {
		args.AppNetworkAccessType = pulumi.StringPtr(spec.GetAppNetworkAccessType())
	}

	domainSettings := buildDomainSettings(spec)
	if domainSettings != nil {
		args.DomainSettings = domainSettings
	}

	createdDomain, err := sagemaker.NewDomain(ctx, "sagemaker-domain", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create sagemaker domain")
	}

	return createdDomain, nil
}

func buildDefaultUserSettings(dus *awssagemakerdomainv1.AwsSagemakerDomainDefaultUserSettings) *sagemaker.DomainDefaultUserSettingsArgs {
	settings := &sagemaker.DomainDefaultUserSettingsArgs{
		ExecutionRole: pulumi.String(dus.ExecutionRoleArn.GetValue()),
	}

	if len(dus.SecurityGroupIds) > 0 {
		var sgs pulumi.StringArray
		for _, sg := range dus.SecurityGroupIds {
			sgs = append(sgs, pulumi.String(sg.GetValue()))
		}
		settings.SecurityGroups = sgs
	}

	if dus.DefaultLandingUri != "" {
		settings.DefaultLandingUri = pulumi.StringPtr(dus.DefaultLandingUri)
	}

	if dus.StudioWebPortal != nil && dus.GetStudioWebPortal() != "" {
		settings.StudioWebPortal = pulumi.StringPtr(dus.GetStudioWebPortal())
	}

	if dus.JupyterLabAppSettings != nil {
		settings.JupyterLabAppSettings = buildJupyterLabAppSettings(dus.JupyterLabAppSettings)
	}

	if dus.KernelGatewayAppSettings != nil {
		settings.KernelGatewayAppSettings = buildKernelGatewayAppSettings(dus.KernelGatewayAppSettings)
	}

	if dus.SharingSettings != nil {
		settings.SharingSettings = buildSharingSettings(dus.SharingSettings)
	}

	if dus.SpaceStorageSettings != nil {
		settings.SpaceStorageSettings = buildSpaceStorageSettings(dus.SpaceStorageSettings)
	}

	return settings
}

func buildJupyterLabAppSettings(jl *awssagemakerdomainv1.AwsSagemakerDomainJupyterLabAppSettings) *sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsArgs {
	jlArgs := &sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsArgs{}

	if jl.DefaultResourceSpec != nil {
		rs := jl.DefaultResourceSpec
		rsArgs := &sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsDefaultResourceSpecArgs{}
		if rs.InstanceType != "" {
			rsArgs.InstanceType = pulumi.StringPtr(rs.InstanceType)
		}
		if rs.LifecycleConfigArn != "" {
			rsArgs.LifecycleConfigArn = pulumi.StringPtr(rs.LifecycleConfigArn)
		}
		if rs.SagemakerImageArn != "" {
			rsArgs.SagemakerImageArn = pulumi.StringPtr(rs.SagemakerImageArn)
		}
		if rs.SagemakerImageVersionAlias != "" {
			rsArgs.SagemakerImageVersionAlias = pulumi.StringPtr(rs.SagemakerImageVersionAlias)
		}
		if rs.SagemakerImageVersionArn != "" {
			rsArgs.SagemakerImageVersionArn = pulumi.StringPtr(rs.SagemakerImageVersionArn)
		}
		jlArgs.DefaultResourceSpec = rsArgs
	}

	if len(jl.LifecycleConfigArns) > 0 {
		var arns pulumi.StringArray
		for _, arn := range jl.LifecycleConfigArns {
			arns = append(arns, pulumi.String(arn))
		}
		jlArgs.LifecycleConfigArns = arns
	}

	if len(jl.CustomImages) > 0 {
		var images sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsCustomImageArray
		for _, img := range jl.CustomImages {
			imgArgs := &sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsCustomImageArgs{
				AppImageConfigName: pulumi.String(img.AppImageConfigName),
				ImageName:          pulumi.String(img.ImageName),
			}
			if img.ImageVersionNumber != nil {
				imgArgs.ImageVersionNumber = pulumi.IntPtr(int(img.GetImageVersionNumber()))
			}
			images = append(images, imgArgs)
		}
		jlArgs.CustomImages = images
	}

	if len(jl.CodeRepositories) > 0 {
		var repos sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsCodeRepositoryArray
		for _, repo := range jl.CodeRepositories {
			repos = append(repos, &sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsCodeRepositoryArgs{
				RepositoryUrl: pulumi.String(repo.RepositoryUrl),
			})
		}
		jlArgs.CodeRepositories = repos
	}

	// Idle settings map to AppLifecycleManagement.IdleSettings in the Pulumi SDK
	if jl.IdleSettings != nil {
		idle := jl.IdleSettings
		idleArgs := &sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsAppLifecycleManagementIdleSettingsArgs{}
		if idle.LifecycleManagement != "" {
			idleArgs.LifecycleManagement = pulumi.StringPtr(idle.LifecycleManagement)
		}
		if idle.IdleTimeoutInMinutes != 0 {
			idleArgs.IdleTimeoutInMinutes = pulumi.IntPtr(int(idle.IdleTimeoutInMinutes))
		}
		if idle.MinIdleTimeoutInMinutes != 0 {
			idleArgs.MinIdleTimeoutInMinutes = pulumi.IntPtr(int(idle.MinIdleTimeoutInMinutes))
		}
		if idle.MaxIdleTimeoutInMinutes != 0 {
			idleArgs.MaxIdleTimeoutInMinutes = pulumi.IntPtr(int(idle.MaxIdleTimeoutInMinutes))
		}
		jlArgs.AppLifecycleManagement = &sagemaker.DomainDefaultUserSettingsJupyterLabAppSettingsAppLifecycleManagementArgs{
			IdleSettings: idleArgs,
		}
	}

	return jlArgs
}

func buildKernelGatewayAppSettings(kg *awssagemakerdomainv1.AwsSagemakerDomainKernelGatewayAppSettings) *sagemaker.DomainDefaultUserSettingsKernelGatewayAppSettingsArgs {
	kgArgs := &sagemaker.DomainDefaultUserSettingsKernelGatewayAppSettingsArgs{}

	if kg.DefaultResourceSpec != nil {
		rs := kg.DefaultResourceSpec
		rsArgs := &sagemaker.DomainDefaultUserSettingsKernelGatewayAppSettingsDefaultResourceSpecArgs{}
		if rs.InstanceType != "" {
			rsArgs.InstanceType = pulumi.StringPtr(rs.InstanceType)
		}
		if rs.LifecycleConfigArn != "" {
			rsArgs.LifecycleConfigArn = pulumi.StringPtr(rs.LifecycleConfigArn)
		}
		if rs.SagemakerImageArn != "" {
			rsArgs.SagemakerImageArn = pulumi.StringPtr(rs.SagemakerImageArn)
		}
		if rs.SagemakerImageVersionAlias != "" {
			rsArgs.SagemakerImageVersionAlias = pulumi.StringPtr(rs.SagemakerImageVersionAlias)
		}
		if rs.SagemakerImageVersionArn != "" {
			rsArgs.SagemakerImageVersionArn = pulumi.StringPtr(rs.SagemakerImageVersionArn)
		}
		kgArgs.DefaultResourceSpec = rsArgs
	}

	if len(kg.LifecycleConfigArns) > 0 {
		var arns pulumi.StringArray
		for _, arn := range kg.LifecycleConfigArns {
			arns = append(arns, pulumi.String(arn))
		}
		kgArgs.LifecycleConfigArns = arns
	}

	if len(kg.CustomImages) > 0 {
		var images sagemaker.DomainDefaultUserSettingsKernelGatewayAppSettingsCustomImageArray
		for _, img := range kg.CustomImages {
			imgArgs := &sagemaker.DomainDefaultUserSettingsKernelGatewayAppSettingsCustomImageArgs{
				AppImageConfigName: pulumi.String(img.AppImageConfigName),
				ImageName:          pulumi.String(img.ImageName),
			}
			if img.ImageVersionNumber != nil {
				imgArgs.ImageVersionNumber = pulumi.IntPtr(int(img.GetImageVersionNumber()))
			}
			images = append(images, imgArgs)
		}
		kgArgs.CustomImages = images
	}

	return kgArgs
}

func buildSharingSettings(ss *awssagemakerdomainv1.AwsSagemakerDomainSharingSettings) *sagemaker.DomainDefaultUserSettingsSharingSettingsArgs {
	ssArgs := &sagemaker.DomainDefaultUserSettingsSharingSettingsArgs{}

	if ss.NotebookOutputOption != nil && ss.GetNotebookOutputOption() != "" {
		ssArgs.NotebookOutputOption = pulumi.StringPtr(ss.GetNotebookOutputOption())
	}

	if ss.S3KmsKeyId != nil && ss.S3KmsKeyId.GetValue() != "" {
		ssArgs.S3KmsKeyId = pulumi.StringPtr(ss.S3KmsKeyId.GetValue())
	}

	if ss.S3OutputPath != "" {
		ssArgs.S3OutputPath = pulumi.StringPtr(ss.S3OutputPath)
	}

	return ssArgs
}

func buildSpaceStorageSettings(sss *awssagemakerdomainv1.AwsSagemakerDomainSpaceStorageSettings) *sagemaker.DomainDefaultUserSettingsSpaceStorageSettingsArgs {
	return &sagemaker.DomainDefaultUserSettingsSpaceStorageSettingsArgs{
		DefaultEbsStorageSettings: &sagemaker.DomainDefaultUserSettingsSpaceStorageSettingsDefaultEbsStorageSettingsArgs{
			DefaultEbsVolumeSizeInGb: pulumi.Int(int(sss.DefaultEbsVolumeSizeInGb)),
			MaximumEbsVolumeSizeInGb: pulumi.Int(int(sss.MaximumEbsVolumeSizeInGb)),
		},
	}
}

func buildDomainSettings(spec *awssagemakerdomainv1.AwsSagemakerDomainSpec) *sagemaker.DomainDomainSettingsArgs {
	hasDomainSettings := false
	dsArgs := &sagemaker.DomainDomainSettingsArgs{}

	// Domain-level security groups
	if len(spec.DomainSecurityGroupIds) > 0 {
		var sgIds pulumi.StringArray
		for _, sg := range spec.DomainSecurityGroupIds {
			sgIds = append(sgIds, pulumi.String(sg.GetValue()))
		}
		dsArgs.SecurityGroupIds = sgIds
		hasDomainSettings = true
	}

	// Docker settings
	if spec.DockerSettings != nil {
		docker := spec.DockerSettings
		dockerArgs := &sagemaker.DomainDomainSettingsDockerSettingsArgs{}
		if docker.EnableDockerAccess != "" {
			dockerArgs.EnableDockerAccess = pulumi.StringPtr(docker.EnableDockerAccess)
		}
		if len(docker.VpcOnlyTrustedAccounts) > 0 {
			var accounts pulumi.StringArray
			for _, acct := range docker.VpcOnlyTrustedAccounts {
				accounts = append(accounts, pulumi.String(acct))
			}
			dockerArgs.VpcOnlyTrustedAccounts = accounts
		}
		dsArgs.DockerSettings = dockerArgs
		hasDomainSettings = true
	}

	if !hasDomainSettings {
		return nil
	}

	return dsArgs
}
