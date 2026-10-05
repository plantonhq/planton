package module

import (
	"strconv"

	awsfsxopenzfsfilesystemv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxopenzfsfilesystem/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOpenzfsFileSystem *awsfsxopenzfsfilesystemv1alpha1.AwsFsxOpenzfsFileSystem
	AwsTags                 map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsfsxopenzfsfilesystemv1alpha1.AwsFsxOpenzfsFileSystemIacInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOpenzfsFileSystem = iacInput.Target

	// Resource-identity tags follow the catalog convention. The Name tag is
	// the resource's metadata.name — FSx has no name argument, so the console
	// name is this tag; the Terraform module pins the same basis, keeping the
	// two engines' physical identity converged.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsFsxOpenzfsFileSystem.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOpenzfsFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOpenzfsFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsFsxOpenzfsFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOpenzfsFileSystem.Metadata.Id,
	}

	return locals
}
