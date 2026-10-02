package module

import (
	"strings"

	gcpcertmanagertrustconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertmanagertrustconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpCertManagerTrustConfig *gcpcertmanagertrustconfigv1alpha1.GcpCertManagerTrustConfig
	GcpLabels                 map[string]string

	// ProjectId is empty when the manifest omits it -- the provider's default
	// project then applies (the Terraform module passes null).
	ProjectId string

	// TrustConfigName falls back to metadata.name, identically on both engines.
	TrustConfigName string

	// Location falls back to "global": the provider requires one.
	Location string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcertmanagertrustconfigv1alpha1.GcpCertManagerTrustConfigStackInput) *Locals {
	target := stackInput.Target

	locals := &Locals{
		GcpCertManagerTrustConfig: target,
		ProjectId:                 target.Spec.ProjectId.GetValue(),
		TrustConfigName:           target.Spec.TrustConfigName,
		Location:                  target.Spec.Location,
	}
	if locals.TrustConfigName == "" {
		locals.TrustConfigName = target.Metadata.Name
	}
	if locals.Location == "" {
		locals.Location = "global"
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.TrustConfigName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCertManagerTrustConfig.String())
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
