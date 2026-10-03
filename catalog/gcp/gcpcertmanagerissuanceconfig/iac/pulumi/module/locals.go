package module

import (
	"strings"

	gcpcertmanagerissuanceconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertmanagerissuanceconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpCertManagerIssuanceConfig *gcpcertmanagerissuanceconfigv1alpha1.GcpCertManagerIssuanceConfig
	GcpLabels                    map[string]string

	// ProjectId is empty when the manifest omits it -- the provider's default
	// project then applies (the Terraform module passes null).
	ProjectId string

	// IssuanceConfigName falls back to metadata.name, identically on both engines.
	IssuanceConfigName string

	// Location is empty when the manifest omits it -- the provider defaults
	// to "global" (the Terraform module passes null).
	Location string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcertmanagerissuanceconfigv1alpha1.GcpCertManagerIssuanceConfigIacInput) *Locals {
	target := iacInput.Target

	locals := &Locals{
		GcpCertManagerIssuanceConfig: target,
		ProjectId:                    target.Spec.ProjectId.GetValue(),
		IssuanceConfigName:           target.Spec.IssuanceConfigName,
		Location:                     target.Spec.Location,
	}
	if locals.IssuanceConfigName == "" {
		locals.IssuanceConfigName = target.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.IssuanceConfigName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCertManagerIssuanceConfig.String())
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
