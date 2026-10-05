package module

import (
	"strconv"

	awsfsxlustrefilesystemv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxlustrefilesystem/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxLustreFileSystem *awsfsxlustrefilesystemv1alpha1.AwsFsxLustreFileSystem
	AwsTags                map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsfsxlustrefilesystemv1alpha1.AwsFsxLustreFileSystemIacInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxLustreFileSystem = iacInput.Target

	// Resource-identity tags follow the catalog convention. The Name tag is
	// the resource's metadata.name — FSx has no name argument, so the console
	// name is this tag; the Terraform module pins the same basis, keeping the
	// two engines' physical identity converged.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsFsxLustreFileSystem.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxLustreFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxLustreFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsFsxLustreFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsFsxLustreFileSystem.Metadata.Id,
	}

	return locals
}
