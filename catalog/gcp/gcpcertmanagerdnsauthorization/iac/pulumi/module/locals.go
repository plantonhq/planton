package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcertmanagerdnsauthorizationv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertmanagerdnsauthorization/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig              *gcpprovider.GcpProviderConfig
	GcpCertManagerDnsAuthorization *gcpcertmanagerdnsauthorizationv1alpha1.GcpCertManagerDnsAuthorization
	GcpLabels                      map[string]string

	// ProjectId is empty when the manifest omits it — the provider's default
	// project then applies (the same ambient contract the Terraform module
	// honors by passing null).
	ProjectId string

	// AuthorizationName falls back to metadata.name — explicit conditional,
	// so both engines derive the identical cloud-side name.
	AuthorizationName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcertmanagerdnsauthorizationv1alpha1.GcpCertManagerDnsAuthorizationIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCertManagerDnsAuthorization = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	locals.ProjectId = iacInput.Target.Spec.ProjectId.GetValue()

	locals.AuthorizationName = iacInput.Target.Spec.AuthorizationName
	if locals.AuthorizationName == "" {
		locals.AuthorizationName = iacInput.Target.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range iacInput.Target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.AuthorizationName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCertManagerDnsAuthorization.String())

	if iacInput.Target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = iacInput.Target.Metadata.Org
	}
	if iacInput.Target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = iacInput.Target.Metadata.Env
	}
	if iacInput.Target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = iacInput.Target.Metadata.Id
	}

	return locals
}
