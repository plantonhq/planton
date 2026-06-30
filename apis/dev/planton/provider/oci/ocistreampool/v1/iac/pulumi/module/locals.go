package module

import (
	ocistreampoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocistreampool/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciStreamPool *ocistreampoolv1.OciStreamPool
	PoolName      string
	FreeformTags  map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocistreampoolv1.OciStreamPoolStackInput) *Locals {
	locals := &Locals{}
	locals.OciStreamPool = stackInput.Target

	locals.PoolName = stackInput.Target.Metadata.Name

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciStreamPool.String(),
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
