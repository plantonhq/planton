package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpprivatecacertificatetemplatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificatetemplate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig               *gcpprovider.GcpProviderConfig
	GcpPrivateCaCertificateTemplate *gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplate
	GcpLabels                       map[string]string

	// TemplateId is spec.template_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.template_id.
	TemplateId string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplateStackInput) *Locals {
	locals := &Locals{}
	locals.GcpPrivateCaCertificateTemplate = stackInput.Target
	metadata := locals.GcpPrivateCaCertificateTemplate.Metadata
	spec := locals.GcpPrivateCaCertificateTemplate.Spec

	locals.TemplateId = spec.TemplateId
	if locals.TemplateId == "" {
		locals.TemplateId = metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpPrivateCaCertificateTemplate.String())

	if metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}

// lastSegment returns what follows the last "/" -- the bare ID of a full
// resource path, or the value itself when it is already bare.
func lastSegment(value string) string {
	return value[strings.LastIndex(value, "/")+1:]
}
