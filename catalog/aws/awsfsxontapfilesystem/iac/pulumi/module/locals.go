package module

import (
	"strconv"

	awsfsxontapfilesystemv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxontapfilesystem/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOntapFileSystem *awsfsxontapfilesystemv1alpha1.AwsFsxOntapFileSystem
	AwsTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsfsxontapfilesystemv1alpha1.AwsFsxOntapFileSystemIacInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOntapFileSystem = iacInput.Target

	// Resource-identity tags follow the catalog convention. The Name tag is
	// the resource's metadata.name — FSx has no name argument, so the console
	// name is this tag; the Terraform module pins the same basis, keeping the
	// two engines' physical identity converged.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsFsxOntapFileSystem.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOntapFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOntapFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsFsxOntapFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOntapFileSystem.Metadata.Id,
	}

	return locals
}
