package module

import (
	"strconv"

	awssagemakermlflowappv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakermlflowapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakermlflowappv1alpha1.AwsSagemakerMlflowApp
	Spec   *awssagemakermlflowappv1alpha1.AwsSagemakerMlflowAppSpec

	AppName string
	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakermlflowappv1alpha1.AwsSagemakerMlflowAppIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The component's name IS the app name (updateable in place - the
	// ARN is the app's identity).
	locals.AppName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerMlflowApp.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
