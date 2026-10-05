package module

import (
	"strings"

	gcpcertmanagercertv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertmanagercert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpCertManagerCert *gcpcertmanagercertv1alpha1.GcpCertManagerCert
	GcpLabels          map[string]string

	// ProjectId is empty when the manifest omits it — the provider's default
	// project then applies (the same ambient contract the Terraform module
	// honors by passing null).
	ProjectId string

	// CertName falls back to metadata.name — explicit conditional, so both
	// engines derive the identical cloud-side name.
	CertName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcertmanagercertv1alpha1.GcpCertManagerCertIacInput) *Locals {
	locals := &Locals{}

	locals.GcpCertManagerCert = iacInput.Target

	target := iacInput.Target

	locals.ProjectId = target.Spec.ProjectId.GetValue()

	locals.CertName = target.Spec.CertName
	if locals.CertName == "" {
		locals.CertName = target.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.CertName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCertManagerCert.String())

	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}

	return locals
}
