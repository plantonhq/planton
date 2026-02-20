package module

import (
	ocikmsvaultv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocikmsvault/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciKmsVault  *ocikmsvaultv1.OciKmsVault
	DisplayName  string
	FreeformTags map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocikmsvaultv1.OciKmsVaultStackInput) *Locals {
	locals := &Locals{}
	locals.OciKmsVault = stackInput.Target

	locals.DisplayName = stackInput.Target.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciKmsVault.String(),
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
