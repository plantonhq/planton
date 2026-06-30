package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/codebuild"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func project(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
) (*codebuild.Project, error) {
	spec := locals.AwsCodeBuildProject.Spec

	// --- Source ---
	src := spec.Source
	sourceArgs := &codebuild.ProjectSourceArgs{
		Type: pulumi.String(src.Type),
	}
	if src.Location != "" {
		sourceArgs.Location = pulumi.StringPtr(src.Location)
	}
	if src.Buildspec != "" {
		sourceArgs.Buildspec = pulumi.StringPtr(src.Buildspec)
	}
	if src.GitCloneDepth > 0 {
		sourceArgs.GitCloneDepth = pulumi.IntPtr(int(src.GitCloneDepth))
	}
	if src.ReportBuildStatus {
		sourceArgs.ReportBuildStatus = pulumi.BoolPtr(true)
	}
	if src.FetchSubmodules {
		sourceArgs.GitSubmodulesConfig = &codebuild.ProjectSourceGitSubmodulesConfigArgs{
			FetchSubmodules: pulumi.Bool(true),
		}
	}

	// --- Environment ---
	env := spec.Environment
	envArgs := &codebuild.ProjectEnvironmentArgs{
		Type:        pulumi.String(env.Type),
		ComputeType: pulumi.String(env.ComputeType),
		Image:       pulumi.String(env.Image),
	}
	if env.PrivilegedMode {
		envArgs.PrivilegedMode = pulumi.BoolPtr(true)
	}
	if env.GetImagePullCredentialsType() != "" && env.GetImagePullCredentialsType() != "CODEBUILD" {
		envArgs.ImagePullCredentialsType = pulumi.StringPtr(env.GetImagePullCredentialsType())
	}
	if len(env.EnvironmentVariables) > 0 {
		var envVars codebuild.ProjectEnvironmentEnvironmentVariableArray
		for _, ev := range env.EnvironmentVariables {
			evArgs := &codebuild.ProjectEnvironmentEnvironmentVariableArgs{
				Name:  pulumi.String(ev.Name),
				Value: pulumi.String(ev.Value),
			}
			if ev.GetType() != "" && ev.GetType() != "PLAINTEXT" {
				evArgs.Type = pulumi.StringPtr(ev.GetType())
			}
			envVars = append(envVars, evArgs)
		}
		envArgs.EnvironmentVariables = envVars
	}
	if env.RegistryCredential != nil {
		envArgs.RegistryCredential = &codebuild.ProjectEnvironmentRegistryCredentialArgs{
			Credential:         pulumi.String(env.RegistryCredential.Credential),
			CredentialProvider: pulumi.String(env.RegistryCredential.CredentialProvider),
		}
	}

	// --- Artifacts ---
	art := spec.Artifacts
	artifactsArgs := &codebuild.ProjectArtifactsArgs{
		Type: pulumi.String(art.Type),
	}
	if art.Location != nil && art.Location.GetValue() != "" {
		artifactsArgs.Location = pulumi.StringPtr(art.Location.GetValue())
	}
	if art.Name != "" {
		artifactsArgs.Name = pulumi.StringPtr(art.Name)
	}
	if art.Path != "" {
		artifactsArgs.Path = pulumi.StringPtr(art.Path)
	}
	if art.Packaging != "" {
		artifactsArgs.Packaging = pulumi.StringPtr(art.Packaging)
	}
	if art.NamespaceType != "" {
		artifactsArgs.NamespaceType = pulumi.StringPtr(art.NamespaceType)
	}
	if art.EncryptionDisabled {
		artifactsArgs.EncryptionDisabled = pulumi.BoolPtr(true)
	}

	// --- Project args ---
	args := &codebuild.ProjectArgs{
		Name:        pulumi.StringPtr(locals.AwsCodeBuildProject.Metadata.Id),
		ServiceRole: pulumi.String(spec.ServiceRole.GetValue()),
		Source:      sourceArgs,
		Environment: envArgs,
		Artifacts:   artifactsArgs,
		Tags:        pulumi.ToStringMap(locals.Labels),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.EncryptionKey != nil && spec.EncryptionKey.GetValue() != "" {
		args.EncryptionKey = pulumi.StringPtr(spec.EncryptionKey.GetValue())
	}
	if spec.GetBuildTimeout() != 0 {
		args.BuildTimeout = pulumi.IntPtr(int(spec.GetBuildTimeout()))
	}
	if spec.GetQueuedTimeout() != 0 {
		args.QueuedTimeout = pulumi.IntPtr(int(spec.GetQueuedTimeout()))
	}
	if spec.ConcurrentBuildLimit > 0 {
		args.ConcurrentBuildLimit = pulumi.IntPtr(int(spec.ConcurrentBuildLimit))
	}
	if spec.SourceVersion != "" {
		args.SourceVersion = pulumi.StringPtr(spec.SourceVersion)
	}

	// --- Cache ---
	if spec.Cache != nil && spec.Cache.GetType() != "" && spec.Cache.GetType() != "NO_CACHE" {
		cacheArgs := &codebuild.ProjectCacheArgs{
			Type: pulumi.StringPtr(spec.Cache.GetType()),
		}
		if spec.Cache.Location != nil && spec.Cache.Location.GetValue() != "" {
			cacheArgs.Location = pulumi.StringPtr(spec.Cache.Location.GetValue())
		}
		if len(spec.Cache.Modes) > 0 {
			cacheArgs.Modes = pulumi.ToStringArray(spec.Cache.Modes)
		}
		args.Cache = cacheArgs
	}

	// --- Logs config ---
	if spec.LogsConfig != nil {
		logsArgs := &codebuild.ProjectLogsConfigArgs{}
		if spec.LogsConfig.CloudwatchLogs != nil {
			cwArgs := &codebuild.ProjectLogsConfigCloudwatchLogsArgs{}
			if spec.LogsConfig.CloudwatchLogs.GetStatus() != "" {
				cwArgs.Status = pulumi.StringPtr(spec.LogsConfig.CloudwatchLogs.GetStatus())
			}
			if spec.LogsConfig.CloudwatchLogs.GroupName != nil && spec.LogsConfig.CloudwatchLogs.GroupName.GetValue() != "" {
				cwArgs.GroupName = pulumi.StringPtr(spec.LogsConfig.CloudwatchLogs.GroupName.GetValue())
			}
			if spec.LogsConfig.CloudwatchLogs.StreamName != "" {
				cwArgs.StreamName = pulumi.StringPtr(spec.LogsConfig.CloudwatchLogs.StreamName)
			}
			logsArgs.CloudwatchLogs = cwArgs
		}
		if spec.LogsConfig.S3Logs != nil {
			s3Args := &codebuild.ProjectLogsConfigS3LogsArgs{}
			if spec.LogsConfig.S3Logs.GetStatus() != "" {
				s3Args.Status = pulumi.StringPtr(spec.LogsConfig.S3Logs.GetStatus())
			}
			if spec.LogsConfig.S3Logs.Location != nil && spec.LogsConfig.S3Logs.Location.GetValue() != "" {
				s3Args.Location = pulumi.StringPtr(spec.LogsConfig.S3Logs.Location.GetValue())
			}
			if spec.LogsConfig.S3Logs.EncryptionDisabled {
				s3Args.EncryptionDisabled = pulumi.BoolPtr(true)
			}
			logsArgs.S3Logs = s3Args
		}
		args.LogsConfig = logsArgs
	}

	// --- VPC config ---
	if spec.VpcConfig != nil {
		var subnetIds pulumi.StringArray
		for _, s := range spec.VpcConfig.SubnetIds {
			if s.GetValue() != "" {
				subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
			}
		}
		var sgIds pulumi.StringArray
		for _, sg := range spec.VpcConfig.SecurityGroupIds {
			if sg.GetValue() != "" {
				sgIds = append(sgIds, pulumi.String(sg.GetValue()))
			}
		}
		args.VpcConfig = &codebuild.ProjectVpcConfigArgs{
			VpcId:            pulumi.String(spec.VpcConfig.VpcId.GetValue()),
			Subnets:          subnetIds,
			SecurityGroupIds: sgIds,
		}
	}

	created, err := codebuild.NewProject(ctx, "codebuild-project", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create codebuild project")
	}

	return created, nil
}
