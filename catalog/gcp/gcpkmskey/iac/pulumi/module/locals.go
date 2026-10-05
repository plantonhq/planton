package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpkmskeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpkmskey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpKmsKey         *gcpkmskeyv1alpha1.GcpKmsKey
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpkmskeyv1alpha1.GcpKmsKeyIacInput) *Locals {
	locals := &Locals{}
	locals.GcpKmsKey = iacInput.Target

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpKmsKey.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpKmsKey.Spec.KeyName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpKmsKey.String())

	if locals.GcpKmsKey.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpKmsKey.Metadata.Org
	}
	if locals.GcpKmsKey.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpKmsKey.Metadata.Env
	}
	if locals.GcpKmsKey.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpKmsKey.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
