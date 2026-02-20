package module

import (
	ocisecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocisecuritygroup/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciSecurityGroup *ocisecuritygroupv1.OciSecurityGroup
	DisplayName             string
	FreeformTags            map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocisecuritygroupv1.OciSecurityGroupStackInput) *Locals {
	locals := &Locals{}
	locals.OciSecurityGroup = stackInput.Target

	if stackInput.Target.Spec.DisplayName != "" {
		locals.DisplayName = stackInput.Target.Spec.DisplayName
	} else {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciSecurityGroup.String(),
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
