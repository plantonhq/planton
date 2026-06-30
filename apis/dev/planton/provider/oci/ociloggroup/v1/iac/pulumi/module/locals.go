package module

import (
	ociloggroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociloggroup/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciLogGroup  *ociloggroupv1.OciLogGroup
	GroupName    string
	FreeformTags map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ociloggroupv1.OciLogGroupStackInput) *Locals {
	locals := &Locals{}
	locals.OciLogGroup = stackInput.Target
	locals.GroupName = stackInput.Target.Metadata.Name

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciLogGroup.String(),
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
