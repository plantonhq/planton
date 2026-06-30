package module

import (
	ocivaultsecretv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocivaultsecret/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciVaultSecret *ocivaultsecretv1.OciVaultSecret
	DisplayName    string
	FreeformTags   map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocivaultsecretv1.OciVaultSecretStackInput) *Locals {
	locals := &Locals{}
	locals.OciVaultSecret = stackInput.Target

	locals.DisplayName = stackInput.Target.Spec.SecretName

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciVaultSecret.String(),
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
