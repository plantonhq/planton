package module

import (
	"github.com/pkg/errors"
	gcpcloudbuildtriggerv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudbuildtrigger/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudbuild"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// trigger enables the Cloud Build API and creates the trigger: at most one
// event source and exactly one build configuration, every optional
// argument sent only when set.
func trigger(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudBuildTrigger.Spec
	resourceName := locals.GcpCloudBuildTrigger.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Cloud Build API. DisableOnDestroy stays false: tearing down one
	// trigger must never disable the API for every other build in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("cloudbuild.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdApi, err := projects.NewService(ctx, "gcpcbtrg-cloudbuild.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable cloudbuild.googleapis.com")
	}

	// The trigger name defaults to metadata.name -- identical to the
	// Terraform module. An empty location leaves the provider's "global".
	triggerName := spec.TriggerName
	if triggerName == "" {
		triggerName = resourceName
	}

	args := &cloudbuild.TriggerArgs{
		Project:                      optionalString(project),
		Location:                     optionalString(spec.Location),
		Name:                         pulumi.String(triggerName),
		Description:                  optionalString(spec.Description),
		Disabled:                     optionalTrue(spec.Disabled),
		Tags:                         optionalStrings(spec.Tags),
		Substitutions:                optionalStringMap(spec.Substitutions),
		ServiceAccount:               optionalString(spec.GetServiceAccount().GetValue()),
		Filter:                       optionalString(spec.Filter),
		IgnoredFiles:                 optionalStrings(spec.IgnoredFiles),
		IncludedFiles:                optionalStrings(spec.IncludedFiles),
		IncludeBuildLogs:             optionalString(spec.IncludeBuildLogs),
		Filename:                     optionalString(spec.Filename),
		DeletionPolicy:               optionalString(spec.DeletionPolicy),
		ApprovalConfig:               approvalConfig(spec.ApprovalConfig),
		RepositoryEventConfig:        repositoryEventConfig(spec.RepositoryEventConfig),
		Github:                       github(spec.Github),
		BitbucketServerTriggerConfig: bitbucketServerTriggerConfig(spec.BitbucketServerTriggerConfig),
		DeveloperConnectEventConfig:  developerConnectEventConfig(spec.DeveloperConnectEventConfig),
		TriggerTemplate:              triggerTemplate(spec.TriggerTemplate),
		PubsubConfig:                 pubsubConfig(spec.PubsubConfig),
		WebhookConfig:                webhookConfig(spec.WebhookConfig),
		SourceToBuild:                sourceToBuild(spec.SourceToBuild),
		GitFileSource:                gitFileSource(spec.GitFileSource),
		Build:                        build(spec.Build),
	}

	created, err := cloudbuild.NewTrigger(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create trigger")
	}

	ctx.Export(OpId, created.ID())
	ctx.Export(OpTriggerId, created.TriggerId)
	ctx.Export(OpName, created.Name)
	return nil
}

// optionalString returns nil for an empty value so the provider default
// applies -- the Terraform module's `!= "" ? value : null`.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.StringPtr(value)
}

// optionalInt returns nil for zero -- the Terraform module's
// `!= 0 ? value : null`.
func optionalInt(value int64) pulumi.IntPtrInput {
	if value == 0 {
		return nil
	}
	return pulumi.IntPtr(int(value))
}

// optionalTrue returns nil for false -- the Terraform module's
// `value ? true : null`.
func optionalTrue(value bool) pulumi.BoolPtrInput {
	if !value {
		return nil
	}
	return pulumi.BoolPtr(true)
}

// optionalStrings returns nil for an empty list -- the Terraform module's
// `length(value) > 0 ? value : null`.
func optionalStrings(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// optionalStringMap returns nil for an empty map.
func optionalStringMap(values map[string]string) pulumi.StringMapInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringMap(values)
}

func approvalConfig(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerApprovalConfig) cloudbuild.TriggerApprovalConfigPtrInput {
	if c == nil {
		return nil
	}
	return &cloudbuild.TriggerApprovalConfigArgs{ApprovalRequired: optionalTrue(c.ApprovalRequired)}
}

// --- Event sources ---

func repositoryEventConfig(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerRepositoryEventConfig) cloudbuild.TriggerRepositoryEventConfigPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuild.TriggerRepositoryEventConfigArgs{
		Repository: optionalString(c.GetRepository().GetValue()),
	}
	if pr := c.PullRequest; pr != nil {
		args.PullRequest = &cloudbuild.TriggerRepositoryEventConfigPullRequestArgs{
			Branch:         optionalString(pr.Branch),
			CommentControl: optionalString(pr.CommentControl),
			InvertRegex:    optionalTrue(pr.InvertRegex),
		}
	}
	if push := c.Push; push != nil {
		args.Push = &cloudbuild.TriggerRepositoryEventConfigPushArgs{
			Branch:      optionalString(push.Branch),
			Tag:         optionalString(push.Tag),
			InvertRegex: optionalTrue(push.InvertRegex),
		}
	}
	return args
}

func github(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerGithub) cloudbuild.TriggerGithubPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuild.TriggerGithubArgs{
		Owner:                        optionalString(c.Owner),
		Name:                         optionalString(c.Name),
		EnterpriseConfigResourceName: optionalString(c.EnterpriseConfigResourceName),
	}
	if pr := c.PullRequest; pr != nil {
		args.PullRequest = &cloudbuild.TriggerGithubPullRequestArgs{
			Branch:         pulumi.String(pr.Branch),
			CommentControl: optionalString(pr.CommentControl),
			InvertRegex:    optionalTrue(pr.InvertRegex),
		}
	}
	if push := c.Push; push != nil {
		args.Push = &cloudbuild.TriggerGithubPushArgs{
			Branch:      optionalString(push.Branch),
			Tag:         optionalString(push.Tag),
			InvertRegex: optionalTrue(push.InvertRegex),
		}
	}
	return args
}

func bitbucketServerTriggerConfig(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerBitbucketServerTriggerConfig) cloudbuild.TriggerBitbucketServerTriggerConfigPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuild.TriggerBitbucketServerTriggerConfigArgs{
		BitbucketServerConfigResource: pulumi.String(c.BitbucketServerConfigResource),
		ProjectKey:                    pulumi.String(c.ProjectKey),
		RepoSlug:                      pulumi.String(c.RepoSlug),
	}
	if pr := c.PullRequest; pr != nil {
		args.PullRequest = &cloudbuild.TriggerBitbucketServerTriggerConfigPullRequestArgs{
			Branch:         pulumi.String(pr.Branch),
			CommentControl: optionalString(pr.CommentControl),
			InvertRegex:    optionalTrue(pr.InvertRegex),
		}
	}
	if push := c.Push; push != nil {
		args.Push = &cloudbuild.TriggerBitbucketServerTriggerConfigPushArgs{
			Branch:      optionalString(push.Branch),
			Tag:         optionalString(push.Tag),
			InvertRegex: optionalTrue(push.InvertRegex),
		}
	}
	return args
}

func developerConnectEventConfig(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerDeveloperConnectEventConfig) cloudbuild.TriggerDeveloperConnectEventConfigPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuild.TriggerDeveloperConnectEventConfigArgs{
		GitRepositoryLink: pulumi.String(c.GitRepositoryLink),
	}
	if pr := c.PullRequest; pr != nil {
		args.PullRequest = &cloudbuild.TriggerDeveloperConnectEventConfigPullRequestArgs{
			Branch:         optionalString(pr.Branch),
			CommentControl: optionalString(pr.CommentControl),
			InvertRegex:    optionalTrue(pr.InvertRegex),
		}
	}
	if push := c.Push; push != nil {
		args.Push = &cloudbuild.TriggerDeveloperConnectEventConfigPushArgs{
			Branch:      optionalString(push.Branch),
			Tag:         optionalString(push.Tag),
			InvertRegex: optionalTrue(push.InvertRegex),
		}
	}
	return args
}

func triggerTemplate(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerTriggerTemplate) cloudbuild.TriggerTriggerTemplatePtrInput {
	if c == nil {
		return nil
	}
	return &cloudbuild.TriggerTriggerTemplateArgs{
		ProjectId:   optionalString(c.ProjectId),
		RepoName:    optionalString(c.RepoName),
		Dir:         optionalString(c.Dir),
		BranchName:  optionalString(c.BranchName),
		TagName:     optionalString(c.TagName),
		CommitSha:   optionalString(c.CommitSha),
		InvertRegex: optionalTrue(c.InvertRegex),
	}
}

func pubsubConfig(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerPubsubConfig) cloudbuild.TriggerPubsubConfigPtrInput {
	if c == nil {
		return nil
	}
	return &cloudbuild.TriggerPubsubConfigArgs{
		Topic:               pulumi.String(c.GetTopic().GetValue()),
		ServiceAccountEmail: optionalString(c.GetServiceAccountEmail().GetValue()),
	}
}

func webhookConfig(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerWebhookConfig) cloudbuild.TriggerWebhookConfigPtrInput {
	if c == nil {
		return nil
	}
	return &cloudbuild.TriggerWebhookConfigArgs{Secret: pulumi.String(c.GetSecret().GetValue())}
}

func sourceToBuild(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerSourceToBuild) cloudbuild.TriggerSourceToBuildPtrInput {
	if c == nil {
		return nil
	}
	return &cloudbuild.TriggerSourceToBuildArgs{
		Repository:             optionalString(c.GetRepository().GetValue()),
		Uri:                    optionalString(c.Uri),
		Ref:                    pulumi.String(c.Ref),
		RepoType:               pulumi.String(c.RepoType),
		GithubEnterpriseConfig: optionalString(c.GithubEnterpriseConfig),
		BitbucketServerConfig:  optionalString(c.BitbucketServerConfig),
	}
}

// --- Build configuration ---

func gitFileSource(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerGitFileSource) cloudbuild.TriggerGitFileSourcePtrInput {
	if c == nil {
		return nil
	}
	return &cloudbuild.TriggerGitFileSourceArgs{
		Path:                   pulumi.String(c.Path),
		RepoType:               pulumi.String(c.RepoType),
		Repository:             optionalString(c.GetRepository().GetValue()),
		Uri:                    optionalString(c.Uri),
		Revision:               optionalString(c.Revision),
		GithubEnterpriseConfig: optionalString(c.GithubEnterpriseConfig),
		BitbucketServerConfig:  optionalString(c.BitbucketServerConfig),
	}
}

func build(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerBuild) cloudbuild.TriggerBuildPtrInput {
	if c == nil {
		return nil
	}
	steps := cloudbuild.TriggerBuildStepArray{}
	for _, s := range c.Steps {
		steps = append(steps, buildStep(s))
	}
	secrets := cloudbuild.TriggerBuildSecretArray{}
	for _, s := range c.Secrets {
		secrets = append(secrets, &cloudbuild.TriggerBuildSecretArgs{
			KmsKeyName: pulumi.String(s.GetKmsKeyName().GetValue()),
			SecretEnv:  optionalStringMap(s.SecretEnv),
		})
	}
	args := &cloudbuild.TriggerBuildArgs{
		Steps:            steps,
		Timeout:          optionalString(c.Timeout),
		QueueTtl:         optionalString(c.QueueTtl),
		Images:           optionalStrings(c.Images),
		LogsBucket:       optionalString(c.GetLogsBucket().GetValue()),
		Substitutions:    optionalStringMap(c.Substitutions),
		Tags:             optionalStrings(c.Tags),
		Artifacts:        buildArtifacts(c.Artifacts),
		Options:          buildOptions(c.Options),
		Source:           buildSource(c.Source),
		AvailableSecrets: buildAvailableSecrets(c.AvailableSecrets),
	}
	if len(secrets) > 0 {
		args.Secrets = secrets
	}
	return args
}

func buildStep(s *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerBuildStep) *cloudbuild.TriggerBuildStepArgs {
	args := &cloudbuild.TriggerBuildStepArgs{
		Name:         pulumi.String(s.Name),
		Id:           optionalString(s.Id),
		Args:         optionalStrings(s.Args),
		Entrypoint:   optionalString(s.Entrypoint),
		Script:       optionalString(s.Script),
		Dir:          optionalString(s.Dir),
		Envs:         optionalStrings(s.Env),
		SecretEnvs:   optionalStrings(s.SecretEnv),
		WaitFors:     optionalStrings(s.WaitFor),
		Timeout:      optionalString(s.Timeout),
		AllowFailure: optionalTrue(s.AllowFailure),
	}
	if len(s.AllowExitCodes) > 0 {
		codes := pulumi.IntArray{}
		for _, code := range s.AllowExitCodes {
			codes = append(codes, pulumi.Int(int(code)))
		}
		args.AllowExitCodes = codes
	}
	if len(s.Volumes) > 0 {
		volumes := cloudbuild.TriggerBuildStepVolumeArray{}
		for _, v := range s.Volumes {
			volumes = append(volumes, &cloudbuild.TriggerBuildStepVolumeArgs{
				Name: pulumi.String(v.Name),
				Path: pulumi.String(v.Path),
			})
		}
		args.Volumes = volumes
	}
	return args
}

func buildArtifacts(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerBuildArtifacts) cloudbuild.TriggerBuildArtifactsPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuild.TriggerBuildArtifactsArgs{
		Images: optionalStrings(c.Images),
	}
	if o := c.Objects; o != nil {
		args.Objects = &cloudbuild.TriggerBuildArtifactsObjectsArgs{
			Location: optionalString(o.Location),
			Paths:    optionalStrings(o.Paths),
		}
	}
	if len(c.MavenArtifacts) > 0 {
		maven := cloudbuild.TriggerBuildArtifactsMavenArtifactArray{}
		for _, m := range c.MavenArtifacts {
			maven = append(maven, &cloudbuild.TriggerBuildArtifactsMavenArtifactArgs{
				Repository: optionalString(m.Repository),
				Path:       optionalString(m.Path),
				ArtifactId: optionalString(m.ArtifactId),
				GroupId:    optionalString(m.GroupId),
				Version:    optionalString(m.Version),
			})
		}
		args.MavenArtifacts = maven
	}
	if len(c.NpmPackages) > 0 {
		npm := cloudbuild.TriggerBuildArtifactsNpmPackageArray{}
		for _, n := range c.NpmPackages {
			npm = append(npm, &cloudbuild.TriggerBuildArtifactsNpmPackageArgs{
				Repository:  optionalString(n.Repository),
				PackagePath: optionalString(n.PackagePath),
			})
		}
		args.NpmPackages = npm
	}
	if len(c.PythonPackages) > 0 {
		python := cloudbuild.TriggerBuildArtifactsPythonPackageArray{}
		for _, p := range c.PythonPackages {
			python = append(python, &cloudbuild.TriggerBuildArtifactsPythonPackageArgs{
				Repository: optionalString(p.Repository),
				Paths:      optionalStrings(p.Paths),
			})
		}
		args.PythonPackages = python
	}
	return args
}

// buildOptions never sends DynamicSubstitutions or SubstitutionOption:
// Google fixes both for triggered builds.
func buildOptions(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerBuildOptions) cloudbuild.TriggerBuildOptionsPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuild.TriggerBuildOptionsArgs{
		MachineType:            optionalString(c.MachineType),
		DiskSizeGb:             optionalInt(c.DiskSizeGb),
		WorkerPool:             optionalString(c.GetWorkerPool().GetValue()),
		Logging:                optionalString(c.Logging),
		LogStreamingOption:     optionalString(c.LogStreamingOption),
		RequestedVerifyOption:  optionalString(c.RequestedVerifyOption),
		SourceProvenanceHashes: optionalStrings(c.SourceProvenanceHash),
		Envs:                   optionalStrings(c.Env),
		SecretEnvs:             optionalStrings(c.SecretEnv),
	}
	if len(c.Volumes) > 0 {
		volumes := cloudbuild.TriggerBuildOptionsVolumeArray{}
		for _, v := range c.Volumes {
			volumes = append(volumes, &cloudbuild.TriggerBuildOptionsVolumeArgs{
				Name: optionalString(v.Name),
				Path: optionalString(v.Path),
			})
		}
		args.Volumes = volumes
	}
	return args
}

func buildSource(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerBuildSource) cloudbuild.TriggerBuildSourcePtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuild.TriggerBuildSourceArgs{}
	if r := c.RepoSource; r != nil {
		args.RepoSource = &cloudbuild.TriggerBuildSourceRepoSourceArgs{
			ProjectId:     optionalString(r.ProjectId),
			RepoName:      pulumi.String(r.RepoName),
			Dir:           optionalString(r.Dir),
			BranchName:    optionalString(r.BranchName),
			TagName:       optionalString(r.TagName),
			CommitSha:     optionalString(r.CommitSha),
			InvertRegex:   optionalTrue(r.InvertRegex),
			Substitutions: optionalStringMap(r.Substitutions),
		}
	}
	if s := c.StorageSource; s != nil {
		args.StorageSource = &cloudbuild.TriggerBuildSourceStorageSourceArgs{
			Bucket:     pulumi.String(s.Bucket),
			Object:     pulumi.String(s.Object),
			Generation: optionalString(s.Generation),
		}
	}
	return args
}

func buildAvailableSecrets(c *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerBuildAvailableSecrets) cloudbuild.TriggerBuildAvailableSecretsPtrInput {
	if c == nil {
		return nil
	}
	secrets := cloudbuild.TriggerBuildAvailableSecretsSecretManagerArray{}
	for _, s := range c.SecretManager {
		secrets = append(secrets, &cloudbuild.TriggerBuildAvailableSecretsSecretManagerArgs{
			Env:         pulumi.String(s.Env),
			VersionName: pulumi.String(s.GetVersionName().GetValue()),
		})
	}
	return &cloudbuild.TriggerBuildAvailableSecretsArgs{SecretManagers: secrets}
}
