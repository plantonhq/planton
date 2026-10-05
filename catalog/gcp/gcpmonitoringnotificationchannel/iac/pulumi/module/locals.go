package module

import (
	"strconv"
	"strings"

	gcpmonitoringnotificationchannelv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmonitoringnotificationchannel/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpMonitoringNotificationChannel *gcpmonitoringnotificationchannelv1alpha1.GcpMonitoringNotificationChannel

	// The console display name defaults to metadata.name when the spec
	// leaves display_name empty — the same naming basis every kind uses.
	DisplayName string

	// Merged user_labels: spec labels first so platform attribution labels
	// win on key conflicts — identical merge order to the Terraform module.
	// These land on the provider's user_labels argument; the per-type
	// channel configuration is a DIFFERENT argument (labels) fed from
	// spec.channel_labels.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpmonitoringnotificationchannelv1alpha1.GcpMonitoringNotificationChannelIacInput) *Locals {
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
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpMonitoringNotificationChannel.String())

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
		GcpMonitoringNotificationChannel: target,
		DisplayName:                      displayName,
		GcpLabels:                        gcpLabels,
	}
}
