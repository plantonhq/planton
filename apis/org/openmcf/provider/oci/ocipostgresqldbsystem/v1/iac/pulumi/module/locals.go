package module

import (
	ocipostgresqldbsystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocipostgresqldbsystem/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciPostgresqlDbSystem *ocipostgresqldbsystemv1.OciPostgresqlDbSystem
	DisplayName           string
	FreeformTags          map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocipostgresqldbsystemv1.OciPostgresqlDbSystemStackInput) *Locals {
	locals := &Locals{}
	locals.OciPostgresqlDbSystem = stackInput.Target

	if stackInput.Target.Spec.DisplayName != "" {
		locals.DisplayName = stackInput.Target.Spec.DisplayName
	} else {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciPostgresqlDbSystem.String(),
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
