package module

import (
	"fmt"

	"github.com/pkg/errors"
	cloudfl "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// pagesProject provisions the Cloudflare Pages project, its optional git source
// and build configuration, its per-environment deployment configs, and any
// attached custom domains.
func pagesProject(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudfl.Provider,
) error {
	spec := locals.CloudflarePagesProject.Spec

	projectArgs := &cloudfl.PagesProjectArgs{
		AccountId:        pulumi.String(spec.AccountId),
		Name:             pulumi.String(spec.Name),
		ProductionBranch: pulumi.String(spec.ProductionBranch),
	}

	if bc := spec.BuildConfig; bc != nil {
		bcArgs := &cloudfl.PagesProjectBuildConfigArgs{}
		if bc.BuildCommand != "" {
			bcArgs.BuildCommand = pulumi.String(bc.BuildCommand)
		}
		if bc.DestinationDir != "" {
			bcArgs.DestinationDir = pulumi.String(bc.DestinationDir)
		}
		if bc.RootDir != "" {
			bcArgs.RootDir = pulumi.String(bc.RootDir)
		}
		if bc.BuildCaching {
			bcArgs.BuildCaching = pulumi.Bool(true)
		}
		if bc.WebAnalyticsTag != "" {
			bcArgs.WebAnalyticsTag = pulumi.String(bc.WebAnalyticsTag)
		}
		if bc.WebAnalyticsToken != "" {
			bcArgs.WebAnalyticsToken = pulumi.String(bc.WebAnalyticsToken)
		}
		projectArgs.BuildConfig = bcArgs
	}

	if s := spec.Source; s != nil {
		cfgArgs := cloudfl.PagesProjectSourceConfigArgs{}
		if cfg := s.Config; cfg != nil {
			if cfg.Owner != "" {
				cfgArgs.Owner = pulumi.String(cfg.Owner)
			}
			if cfg.RepoName != "" {
				cfgArgs.RepoName = pulumi.String(cfg.RepoName)
			}
			if cfg.ProductionBranch != "" {
				cfgArgs.ProductionBranch = pulumi.String(cfg.ProductionBranch)
			}
			if cfg.PrCommentsEnabled {
				cfgArgs.PrCommentsEnabled = pulumi.Bool(true)
			}
			if cfg.DeploymentsEnabled {
				cfgArgs.DeploymentsEnabled = pulumi.Bool(true)
			}
			if cfg.ProductionDeploymentsEnabled {
				cfgArgs.ProductionDeploymentsEnabled = pulumi.Bool(true)
			}
			if cfg.PreviewDeploymentSetting != "" {
				cfgArgs.PreviewDeploymentSetting = pulumi.String(cfg.PreviewDeploymentSetting)
			}
			if len(cfg.PreviewBranchIncludes) > 0 {
				cfgArgs.PreviewBranchIncludes = pulumi.ToStringArray(cfg.PreviewBranchIncludes)
			}
			if len(cfg.PreviewBranchExcludes) > 0 {
				cfgArgs.PreviewBranchExcludes = pulumi.ToStringArray(cfg.PreviewBranchExcludes)
			}
			if len(cfg.PathIncludes) > 0 {
				cfgArgs.PathIncludes = pulumi.ToStringArray(cfg.PathIncludes)
			}
			if len(cfg.PathExcludes) > 0 {
				cfgArgs.PathExcludes = pulumi.ToStringArray(cfg.PathExcludes)
			}
		}
		projectArgs.Source = &cloudfl.PagesProjectSourceArgs{
			Type:   pulumi.String(s.Type),
			Config: cfgArgs,
		}
	}

	projectArgs.DeploymentConfigs = deploymentConfigs(spec.DeploymentConfigs)

	createdProject, err := cloudfl.NewPagesProject(
		ctx,
		"pages-project",
		projectArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare pages project")
	}

	// Attach custom domains; each must be a hostname in a zone on this account.
	for i, domain := range spec.Domains {
		if _, err := cloudfl.NewPagesDomain(
			ctx,
			fmt.Sprintf("pages-domain-%d", i),
			&cloudfl.PagesDomainArgs{
				AccountId:   pulumi.String(spec.AccountId),
				ProjectName: createdProject.Name,
				Name:        pulumi.String(domain),
			},
			pulumi.Provider(cloudflareProvider),
			pulumi.Parent(createdProject),
		); err != nil {
			return errors.Wrap(err, "failed to attach cloudflare pages domain")
		}
	}

	ctx.Export(OpProjectName, createdProject.Name)
	ctx.Export(OpSubdomain, createdProject.Subdomain)
	ctx.Export(OpDomains, createdProject.Domains)
	ctx.Export(OpCreatedOn, createdProject.CreatedOn)

	return nil
}
