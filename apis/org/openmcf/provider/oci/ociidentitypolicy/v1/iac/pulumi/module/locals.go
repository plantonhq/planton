package module

import (
	ociidentitypolicyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociidentitypolicy/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciIdentityPolicy *ociidentitypolicyv1.OciIdentityPolicy
	Name              string
	FreeformTags      map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ociidentitypolicyv1.OciIdentityPolicyStackInput) *Locals {
	locals := &Locals{}
	locals.OciIdentityPolicy = stackInput.Target

	if stackInput.Target.Spec.Name != "" {
		locals.Name = stackInput.Target.Spec.Name
	} else {
		locals.Name = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciIdentityPolicy.String(),
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
