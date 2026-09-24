package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdatastreamprivateconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdatastreamprivateconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig              *gcpprovider.GcpProviderConfig
	GcpDatastreamPrivateConnection *gcpdatastreamprivateconnectionv1alpha1.GcpDatastreamPrivateConnection
	GcpLabels                      map[string]string

	// PrivateConnectionId is spec.private_connection_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.private_connection_id.
	PrivateConnectionId string

	// DisplayName is spec.display_name when set, otherwise metadata.name --
	// identical to the Terraform module's locals.display_name.
	DisplayName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpdatastreamprivateconnectionv1alpha1.GcpDatastreamPrivateConnectionStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDatastreamPrivateConnection = stackInput.Target
	metadata := locals.GcpDatastreamPrivateConnection.Metadata
	spec := locals.GcpDatastreamPrivateConnection.Spec

	locals.PrivateConnectionId = spec.PrivateConnectionId
	if locals.PrivateConnectionId == "" {
		locals.PrivateConnectionId = metadata.Name
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
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpDatastreamPrivateConnection.String())

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
