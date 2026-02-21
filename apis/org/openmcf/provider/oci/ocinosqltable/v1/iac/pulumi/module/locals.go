package module

import (
	ocinosqltablev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocinosqltable/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciNosqlTable *ocinosqltablev1.OciNosqlTable
	TableName     string
	FreeformTags  map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocinosqltablev1.OciNosqlTableStackInput) *Locals {
	locals := &Locals{}
	locals.OciNosqlTable = stackInput.Target

	locals.TableName = stackInput.Target.Spec.Name

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciNosqlTable.String(),
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
