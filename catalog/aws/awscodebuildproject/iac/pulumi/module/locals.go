package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awscodebuildprojectv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscodebuildproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsCodeBuildProject *awscodebuildprojectv1alpha1.AwsCodeBuildProject

	// ProjectName is the project's cloud name, derived from metadata.name.
	// CodeBuild project names are create-time-immutable (a name change
	// replaces the project), which is why the name is not spec surface.
	// Same basis as the Terraform module.
	ProjectName string

	// AwsTags are the resource-identity tags, matching the Terraform module
	// key-for-key. DELIBERATELY five keys, no Name: the project carries its
	// own name argument (the ECR/GA recorded convention). Identity tagging is
	// the only tagging surface this module manages; user-defined custom tags
	// are a platform-wide concern, not per-kind spec surface.
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awscodebuildprojectv1alpha1.AwsCodeBuildProjectIacInput) *Locals {
	locals := &Locals{}
	locals.AwsCodeBuildProject = in.Target

	locals.ProjectName = locals.AwsCodeBuildProject.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsCodeBuildProject.Metadata.Org,
		awstagkeys.Environment:  locals.AwsCodeBuildProject.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsCodeBuildProject.String(),
		awstagkeys.ResourceId:   locals.AwsCodeBuildProject.Metadata.Id,
	}

	return locals
}
