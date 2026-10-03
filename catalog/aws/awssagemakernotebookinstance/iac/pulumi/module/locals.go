package module

import (
	"strconv"

	awssagemakernotebookinstancev1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakernotebookinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakernotebookinstancev1alpha1.AwsSagemakerNotebookInstance
	Spec   *awssagemakernotebookinstancev1alpha1.AwsSagemakerNotebookInstanceSpec

	NotebookName        string
	LifecycleConfigName string
	AwsTags             map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakernotebookinstancev1alpha1.AwsSagemakerNotebookInstanceIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The instance's AWS name derives from metadata.name; the folded
	// lifecycle configuration rides a stable derived name.
	locals.NotebookName = metadata.Name
	locals.LifecycleConfigName = metadata.Name + "-lifecycle"

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerNotebookInstance.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
