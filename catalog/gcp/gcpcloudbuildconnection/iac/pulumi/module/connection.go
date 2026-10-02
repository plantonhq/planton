package module

import (
	"github.com/pkg/errors"
	gcpcloudbuildconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudbuildconnection/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudbuildv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// connection enables the Cloud Build API and creates the connection to one
// code host. Every credential is a Secret Manager secret version name.
func connection(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudBuildConnection.Spec
	resourceName := locals.GcpCloudBuildConnection.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Cloud Build API. DisableOnDestroy stays false: tearing down one
	// connection must never disable the API for every other build in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("cloudbuild.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdApi, err := projects.NewService(ctx, "gcpcbcon-cloudbuild.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable cloudbuild.googleapis.com")
	}

	// The connection ID defaults to metadata.name -- identical to the
	// Terraform module.
	connectionId := spec.ConnectionId
	if connectionId == "" {
		connectionId = resourceName
	}

	args := &cloudbuildv2.ConnectionArgs{
		Location:                  pulumi.String(spec.Location),
		Name:                      pulumi.String(connectionId),
		GithubConfig:              githubConfig(spec.GithubConfig),
		GithubEnterpriseConfig:    githubEnterpriseConfig(spec.GithubEnterpriseConfig),
		GitlabConfig:              gitlabConfig(spec.GitlabConfig),
		BitbucketCloudConfig:      bitbucketCloudConfig(spec.BitbucketCloudConfig),
		BitbucketDataCenterConfig: bitbucketDataCenterConfig(spec.BitbucketDataCenterConfig),
	}
	if project != "" {
		args.Project = pulumi.StringPtr(project)
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}
	if spec.Disabled {
		args.Disabled = pulumi.BoolPtr(true)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := cloudbuildv2.NewConnection(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create connection")
	}

	ctx.Export(OpName, created.ID())
	ctx.Export(OpConnectionId, created.Name)
	ctx.Export(OpInstallationStage, created.InstallationStates.ApplyT(func(states []cloudbuildv2.ConnectionInstallationState) string {
		if len(states) == 0 || states[0].Stage == nil {
			return ""
		}
		return *states[0].Stage
	}).(pulumi.StringOutput))
	ctx.Export(OpInstallationActionUri, created.InstallationStates.ApplyT(func(states []cloudbuildv2.ConnectionInstallationState) string {
		if len(states) == 0 || states[0].ActionUri == nil {
			return ""
		}
		return *states[0].ActionUri
	}).(pulumi.StringOutput))
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

func githubConfig(c *gcpcloudbuildconnectionv1alpha1.GcpCloudBuildConnectionGithubConfig) cloudbuildv2.ConnectionGithubConfigPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuildv2.ConnectionGithubConfigArgs{
		AppInstallationId: optionalInt(c.AppInstallationId),
	}
	if a := c.AuthorizerCredential; a != nil {
		args.AuthorizerCredential = &cloudbuildv2.ConnectionGithubConfigAuthorizerCredentialArgs{
			OauthTokenSecretVersion: optionalString(a.GetOauthTokenSecretVersion().GetValue()),
		}
	}
	return args
}

func githubEnterpriseConfig(c *gcpcloudbuildconnectionv1alpha1.GcpCloudBuildConnectionGithubEnterpriseConfig) cloudbuildv2.ConnectionGithubEnterpriseConfigPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuildv2.ConnectionGithubEnterpriseConfigArgs{
		HostUri:                    pulumi.String(c.HostUri),
		AppId:                      optionalInt(c.AppId),
		AppInstallationId:          optionalInt(c.AppInstallationId),
		AppSlug:                    optionalString(c.AppSlug),
		PrivateKeySecretVersion:    optionalString(c.GetPrivateKeySecretVersion().GetValue()),
		WebhookSecretSecretVersion: optionalString(c.GetWebhookSecretSecretVersion().GetValue()),
		SslCa:                      optionalString(c.SslCa),
	}
	if s := c.ServiceDirectoryConfig; s != nil {
		args.ServiceDirectoryConfig = &cloudbuildv2.ConnectionGithubEnterpriseConfigServiceDirectoryConfigArgs{
			Service: pulumi.String(s.Service),
		}
	}
	return args
}

func gitlabConfig(c *gcpcloudbuildconnectionv1alpha1.GcpCloudBuildConnectionGitlabConfig) cloudbuildv2.ConnectionGitlabConfigPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuildv2.ConnectionGitlabConfigArgs{
		HostUri:                    optionalString(c.HostUri),
		WebhookSecretSecretVersion: pulumi.String(c.GetWebhookSecretSecretVersion().GetValue()),
		SslCa:                      optionalString(c.SslCa),
		AuthorizerCredential: &cloudbuildv2.ConnectionGitlabConfigAuthorizerCredentialArgs{
			UserTokenSecretVersion: pulumi.String(c.GetAuthorizerCredential().GetUserTokenSecretVersion().GetValue()),
		},
		ReadAuthorizerCredential: &cloudbuildv2.ConnectionGitlabConfigReadAuthorizerCredentialArgs{
			UserTokenSecretVersion: pulumi.String(c.GetReadAuthorizerCredential().GetUserTokenSecretVersion().GetValue()),
		},
	}
	if s := c.ServiceDirectoryConfig; s != nil {
		args.ServiceDirectoryConfig = &cloudbuildv2.ConnectionGitlabConfigServiceDirectoryConfigArgs{
			Service: pulumi.String(s.Service),
		}
	}
	return args
}

func bitbucketCloudConfig(c *gcpcloudbuildconnectionv1alpha1.GcpCloudBuildConnectionBitbucketCloudConfig) cloudbuildv2.ConnectionBitbucketCloudConfigPtrInput {
	if c == nil {
		return nil
	}
	return &cloudbuildv2.ConnectionBitbucketCloudConfigArgs{
		Workspace:                  pulumi.String(c.Workspace),
		WebhookSecretSecretVersion: pulumi.String(c.GetWebhookSecretSecretVersion().GetValue()),
		AuthorizerCredential: &cloudbuildv2.ConnectionBitbucketCloudConfigAuthorizerCredentialArgs{
			UserTokenSecretVersion: pulumi.String(c.GetAuthorizerCredential().GetUserTokenSecretVersion().GetValue()),
		},
		ReadAuthorizerCredential: &cloudbuildv2.ConnectionBitbucketCloudConfigReadAuthorizerCredentialArgs{
			UserTokenSecretVersion: pulumi.String(c.GetReadAuthorizerCredential().GetUserTokenSecretVersion().GetValue()),
		},
	}
}

func bitbucketDataCenterConfig(c *gcpcloudbuildconnectionv1alpha1.GcpCloudBuildConnectionBitbucketDataCenterConfig) cloudbuildv2.ConnectionBitbucketDataCenterConfigPtrInput {
	if c == nil {
		return nil
	}
	args := &cloudbuildv2.ConnectionBitbucketDataCenterConfigArgs{
		HostUri:                    pulumi.String(c.HostUri),
		WebhookSecretSecretVersion: pulumi.String(c.GetWebhookSecretSecretVersion().GetValue()),
		SslCa:                      optionalString(c.SslCa),
		AuthorizerCredential: &cloudbuildv2.ConnectionBitbucketDataCenterConfigAuthorizerCredentialArgs{
			UserTokenSecretVersion: pulumi.String(c.GetAuthorizerCredential().GetUserTokenSecretVersion().GetValue()),
		},
		ReadAuthorizerCredential: &cloudbuildv2.ConnectionBitbucketDataCenterConfigReadAuthorizerCredentialArgs{
			UserTokenSecretVersion: pulumi.String(c.GetReadAuthorizerCredential().GetUserTokenSecretVersion().GetValue()),
		},
	}
	if s := c.ServiceDirectoryConfig; s != nil {
		args.ServiceDirectoryConfig = &cloudbuildv2.ConnectionBitbucketDataCenterConfigServiceDirectoryConfigArgs{
			Service: pulumi.String(s.Service),
		}
	}
	return args
}
