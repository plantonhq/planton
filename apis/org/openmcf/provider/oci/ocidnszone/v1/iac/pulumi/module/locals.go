package module

import (
	ocidnszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocidnszone/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciDnsZone   *ocidnszonev1.OciDnsZone
	ZoneName     string
	FreeformTags map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocidnszonev1.OciDnsZoneStackInput) *Locals {
	locals := &Locals{}
	locals.OciDnsZone = stackInput.Target
	locals.ZoneName = stackInput.Target.Metadata.Name

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciDnsZone.String(),
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
