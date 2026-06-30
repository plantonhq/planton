package module

import (
	ociapigatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociapigateway/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciApiGateway *ociapigatewayv1.OciApiGateway
	DisplayName   string
	FreeformTags  map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ociapigatewayv1.OciApiGatewayStackInput) *Locals {
	locals := &Locals{}
	locals.OciApiGateway = stackInput.Target

	locals.DisplayName = stackInput.Target.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciApiGateway.String(),
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
