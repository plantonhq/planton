package module

import (
	"strconv"
	"strings"

	gcpeventarcmessagebusv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpeventarcmessagebus/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpEventarcMessageBus *gcpeventarcmessagebusv1alpha1.GcpEventarcMessageBus

	// The bus ID defaults to metadata.name when the spec leaves
	// message_bus_id empty — the same naming basis every kind uses.
	MessageBusId string

	// Merged labels: spec labels first so platform attribution labels win
	// on key conflicts — identical merge order to the Terraform module.
	// Applied to the bus and every satellite (per-satellite spec labels
	// are merged on top per resource).
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpeventarcmessagebusv1alpha1.GcpEventarcMessageBusIacInput) *Locals {
	target := iacInput.Target

	messageBusId := target.Spec.MessageBusId
	if messageBusId == "" {
		messageBusId = target.Metadata.Name
	}

	gcpLabels := map[string]string{}
	for key, value := range target.Spec.Labels {
		gcpLabels[key] = value
	}
	gcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	gcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpEventarcMessageBus.String())
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
		GcpEventarcMessageBus: target,
		MessageBusId:          messageBusId,
		GcpLabels:             gcpLabels,
	}
}

// satelliteLabels layers the shared bus-level set (user labels + platform
// attribution keys) over a satellite's own labels, so platform attribution
// can never be shadowed by a satellite label — identical merge order to
// the Terraform module.
func satelliteLabels(shared map[string]string, own map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range own {
		merged[key] = value
	}
	for key, value := range shared {
		merged[key] = value
	}
	return merged
}
