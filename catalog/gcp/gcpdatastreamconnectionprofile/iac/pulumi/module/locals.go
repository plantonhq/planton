package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdatastreamconnectionprofilev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdatastreamconnectionprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig              *gcpprovider.GcpProviderConfig
	GcpDatastreamConnectionProfile *gcpdatastreamconnectionprofilev1alpha1.GcpDatastreamConnectionProfile
	GcpLabels                      map[string]string

	// ConnectionProfileId is spec.connection_profile_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.connection_profile_id.
	ConnectionProfileId string

	// DisplayName is spec.display_name when set, otherwise metadata.name --
	// identical to the Terraform module's locals.display_name.
	DisplayName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpdatastreamconnectionprofilev1alpha1.GcpDatastreamConnectionProfileStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDatastreamConnectionProfile = stackInput.Target
	metadata := locals.GcpDatastreamConnectionProfile.Metadata
	spec := locals.GcpDatastreamConnectionProfile.Spec

	locals.ConnectionProfileId = spec.ConnectionProfileId
	if locals.ConnectionProfileId == "" {
		locals.ConnectionProfileId = metadata.Name
	}
	locals.DisplayName = spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpDatastreamConnectionProfile.String())

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
