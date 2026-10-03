package module

import (
	"strconv"
	"strings"

	gcpeventarctriggerv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpeventarctrigger/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpEventarcTrigger *gcpeventarctriggerv1alpha1.GcpEventarcTrigger

	// The trigger name defaults to metadata.name when the spec leaves
	// trigger_name empty — the same naming basis every kind uses.
	TriggerName string

	// The partner channel name defaults to "{trigger name}-channel" when
	// the partner_channel arm is present with an empty channel_name.
	ChannelName string

	// Merged labels: spec labels first so platform attribution labels win
	// on key conflicts — identical merge order to the Terraform module.
	// Applied to the trigger and to the partner channel.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpeventarctriggerv1alpha1.GcpEventarcTriggerIacInput) *Locals {
	target := iacInput.Target

	triggerName := target.Spec.TriggerName
	if triggerName == "" {
		triggerName = target.Metadata.Name
	}

	channelName := ""
	if target.Spec.PartnerChannel != nil {
		channelName = target.Spec.PartnerChannel.ChannelName
		if channelName == "" {
			channelName = triggerName + "-channel"
		}
	}

	gcpLabels := map[string]string{}
	for key, value := range target.Spec.Labels {
		gcpLabels[key] = value
	}
	gcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	gcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpEventarcTrigger.String())
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
		GcpEventarcTrigger: target,
		TriggerName:        triggerName,
		ChannelName:        channelName,
		GcpLabels:          gcpLabels,
	}
}
