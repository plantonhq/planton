package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/codebuild"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// webhook creates a CodeBuild webhook for source-triggered builds.
// Only called when spec.webhook is configured.
func webhook(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	proj *codebuild.Project,
) error {
	spec := locals.AwsCodeBuildProject.Spec.Webhook

	args := &codebuild.WebhookArgs{
		ProjectName: proj.Name,
	}

	if spec.BuildType != "" {
		args.BuildType = pulumi.StringPtr(spec.BuildType)
	}

	if len(spec.FilterGroups) > 0 {
		var filterGroups codebuild.WebhookFilterGroupArray
		for _, fg := range spec.FilterGroups {
			var filters codebuild.WebhookFilterGroupFilterArray
			for _, f := range fg.Filters {
				filterArgs := &codebuild.WebhookFilterGroupFilterArgs{
					Type:    pulumi.String(f.Type),
					Pattern: pulumi.String(f.Pattern),
				}
				if f.ExcludeMatchedPattern {
					filterArgs.ExcludeMatchedPattern = pulumi.BoolPtr(true)
				}
				filters = append(filters, filterArgs)
			}
			filterGroups = append(filterGroups, &codebuild.WebhookFilterGroupArgs{
				Filters: filters,
			})
		}
		args.FilterGroups = filterGroups
	}

	created, err := codebuild.NewWebhook(ctx, "codebuild-webhook", args,
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{proj}),
	)
	if err != nil {
		return errors.Wrap(err, "create codebuild webhook")
	}

	ctx.Export(OpWebhookUrl, created.Url)
	ctx.Export(OpWebhookPayload, created.PayloadUrl)

	return nil
}
