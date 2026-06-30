package module

import (
	ocidevopsprojectv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocidevopsproject/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciDevopsProject *ocidevopsprojectv1.OciDevopsProject
	ProjectName      string
	FreeformTags     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocidevopsprojectv1.OciDevopsProjectStackInput) *Locals {
	locals := &Locals{}
	locals.OciDevopsProject = stackInput.Target
	locals.ProjectName = stackInput.Target.Metadata.Name

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciDevopsProject.String(),
		"resource_id":   stackInput.Target.Metadata.Id,
	}
	if stackInput.Target.Metadata.Org != "" {
		locals.FreeformTags["organization"] = stackInput.Target.Metadata.Org
	}
	if stackInput.Target.Metadata.Env != "" {
		locals.FreeformTags["environment"] = stackInput.Target.Metadata.Env
	}
	for k, v := range stackInput.Target.Metadata.Labels {
		locals.FreeformTags[k] = v
	}

	return locals
}
