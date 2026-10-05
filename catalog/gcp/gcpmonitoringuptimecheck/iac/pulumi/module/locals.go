package module

import (
	"strconv"
	"strings"

	gcpmonitoringuptimecheckv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmonitoringuptimecheck/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpMonitoringUptimeCheck *gcpmonitoringuptimecheckv1alpha1.GcpMonitoringUptimeCheck

	// The console display name defaults to metadata.name when the spec
	// leaves display_name empty (the GCP API requires one).
	DisplayName string

	// Merged user_labels: spec labels first so platform attribution labels
	// win on key conflicts — identical merge order to the Terraform module.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpmonitoringuptimecheckv1alpha1.GcpMonitoringUptimeCheckIacInput) *Locals {
	target := iacInput.Target

	displayName := target.Spec.DisplayName
	if displayName == "" {
		displayName = target.Metadata.Name
	}

	gcpLabels := map[string]string{}
	for key, value := range target.Spec.Labels {
		gcpLabels[key] = value
	}
	gcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	gcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpMonitoringUptimeCheck.String())

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
		GcpMonitoringUptimeCheck: target,
		DisplayName:              displayName,
		GcpLabels:                gcpLabels,
	}
}
