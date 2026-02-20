package module

import (
	ocikmskeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocikmskey/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciKmsKey    *ocikmskeyv1.OciKmsKey
	DisplayName  string
	FreeformTags map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocikmskeyv1.OciKmsKeyStackInput) *Locals {
	locals := &Locals{}
	locals.OciKmsKey = stackInput.Target

	locals.DisplayName = stackInput.Target.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciKmsKey.String(),
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
