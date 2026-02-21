package module

import (
	ociblockvolumev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociblockvolume/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciBlockVolume *ociblockvolumev1.OciBlockVolume
	DisplayName    string
	FreeformTags   map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ociblockvolumev1.OciBlockVolumeStackInput) *Locals {
	locals := &Locals{}
	locals.OciBlockVolume = stackInput.Target

	locals.DisplayName = stackInput.Target.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciBlockVolume.String(),
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
