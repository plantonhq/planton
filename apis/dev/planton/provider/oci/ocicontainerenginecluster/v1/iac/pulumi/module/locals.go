package module

import (
	ocicontainerengineclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocicontainerenginecluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciContainerEngineCluster *ocicontainerengineclusterv1.OciContainerEngineCluster
	DisplayName               string
	FreeformTags              map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocicontainerengineclusterv1.OciContainerEngineClusterStackInput) *Locals {
	locals := &Locals{}
	locals.OciContainerEngineCluster = stackInput.Target

	if stackInput.Target.Spec.Name != "" {
		locals.DisplayName = stackInput.Target.Spec.Name
	} else {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciContainerEngineCluster.String(),
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
