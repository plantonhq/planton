package module

import (
	ocialarmv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocialarm/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciAlarm     *ocialarmv1.OciAlarm
	AlarmName    string
	FreeformTags map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocialarmv1.OciAlarmStackInput) *Locals {
	locals := &Locals{}
	locals.OciAlarm = stackInput.Target
	locals.AlarmName = stackInput.Target.Metadata.Name

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciAlarm.String(),
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
