package module

import (
	"strconv"
	"strings"

	gcpmonitoringslov1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmonitoringslo/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpMonitoringSlo *gcpmonitoringslov1alpha1.GcpMonitoringSlo

	// The SLO display name defaults to metadata.name when the spec leaves
	// display_name empty — the same naming basis every kind uses.
	DisplayName string

	// The service ID for a service this kind CREATES (custom_service or
	// basic_service arm) — the arm's service_id, defaulting to
	// metadata.name.
	CreatedServiceId string

	// Merged user_labels: spec labels first so platform attribution labels
	// win on key conflicts — identical merge order to the Terraform module.
	// Applied to the SLO and to any service this kind creates.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpmonitoringslov1alpha1.GcpMonitoringSloIacInput) *Locals {
	target := iacInput.Target

	displayName := target.Spec.DisplayName
	if displayName == "" {
		displayName = target.Metadata.Name
	}

	createdServiceId := target.Metadata.Name
	if custom := target.Spec.Service.GetCustomService(); custom != nil && custom.ServiceId != "" {
		createdServiceId = custom.ServiceId
	}
	if basic := target.Spec.Service.GetBasicService(); basic != nil && basic.ServiceId != "" {
		createdServiceId = basic.ServiceId
	}

	gcpLabels := map[string]string{}
	for key, value := range target.Spec.Labels {
		gcpLabels[key] = value
	}
	gcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	gcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpMonitoringSlo.String())

	if target.Metadata.Org != "" {
		gcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		gcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}
	if target.Metadata.Id != "" {
		gcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	return &Locals{
		GcpMonitoringSlo: target,
		DisplayName:      displayName,
		CreatedServiceId: createdServiceId,
		GcpLabels:        gcpLabels,
	}
}
