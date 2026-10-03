package module

import (
	"strings"

	gcpprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProject *gcpprojectv1alpha1.GcpProject
	GcpLabels  map[string]string

	// DisplayName falls back to metadata.name — explicit conditional, so
	// both engines derive the identical display name.
	DisplayName string

	// DeletionPolicy defaults to DELETE explicitly so destroy semantics are
	// identical on both engines (the bridged provider would otherwise apply
	// its own client-side default).
	DeletionPolicy string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpprojectv1alpha1.GcpProjectIacInput) *Locals {
	locals := &Locals{}
	locals.GcpProject = iacInput.Target

	target := iacInput.Target

	locals.DisplayName = target.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = target.Metadata.Name
	}

	locals.DeletionPolicy = target.Spec.DeletionPolicy
	if locals.DeletionPolicy == "" {
		locals.DeletionPolicy = "DELETE"
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = make(map[string]string)
	for key, value := range target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpProject.String())

	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}

	return locals
}
