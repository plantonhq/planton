package module

import (
	"strconv"

	awsfsxontapvolumev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxontapvolume/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOntapVolume *awsfsxontapvolumev1alpha1.AwsFsxOntapVolume
	AwsTags           map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsfsxontapvolumev1alpha1.AwsFsxOntapVolumeIacInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOntapVolume = iacInput.Target

	// Resource-identity tags follow the catalog convention. The Name tag is
	// the resource's metadata.name — distinct from spec.name, the
	// ONTAP-internal volume identity; the Terraform module pins the same
	// basis, keeping the two engines' physical identity converged.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsFsxOntapVolume.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOntapVolume.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOntapVolume.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsFsxOntapVolume.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOntapVolume.Metadata.Id,
	}

	return locals
}
