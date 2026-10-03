package module

import (
	"strconv"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcomputemigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputemig/v1alpha1"
)

// Locals holds handy references and derived values used across this module.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpComputeMig     *gcpcomputemigv1alpha1.GcpComputeMig
	GcpLabels         map[string]string

	// MigName is the cloud-side group name: spec.mig_name when set,
	// metadata.name otherwise — the same explicit conditional as the
	// Terraform module, so both engines derive the identical name.
	MigName string

	// BaseInstanceName defaults to the group name when the spec leaves it
	// empty — instances are then named "<mig-name>-<suffix>".
	BaseInstanceName string

	// AutoscalerName defaults to the group name when the spec leaves it
	// empty.
	AutoscalerName string

	// TemplateNamePrefix drives the template's native rotation naming:
	// templates are immutable, so every template change creates a fresh
	// "<prefix><timestamp+counter>" name. Kept at or under 37 characters
	// so the provider uses its readable timestamp form (beyond 37 it
	// falls back to a collision-prone shortened UUID) — identical
	// truncation in the Terraform module.
	TemplateNamePrefix string

	// IsRegional selects the regional resource family (region set) over
	// the zonal one (zone set); the spec's CEL guarantees exactly one.
	IsRegional bool

	// Location is the group's zone or region — exported as the location
	// output for scope-compatibility checks downstream.
	Location string
}

// initializeLocals fills the Locals struct from the incoming IaC input.
func initializeLocals(iacInput *gcpcomputemigv1alpha1.GcpComputeMigIacInput) *Locals {
	locals := &Locals{}

	locals.GcpComputeMig = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	target := iacInput.Target

	locals.MigName = target.Spec.MigName
	if locals.MigName == "" {
		locals.MigName = target.Metadata.Name
	}

	locals.BaseInstanceName = target.Spec.BaseInstanceName
	if locals.BaseInstanceName == "" {
		locals.BaseInstanceName = locals.MigName
	}

	locals.AutoscalerName = locals.MigName
	if target.Spec.Autoscaler != nil && target.Spec.Autoscaler.AutoscalerName != "" {
		locals.AutoscalerName = target.Spec.Autoscaler.AutoscalerName
	}

	// "<mig-name>-" capped at 37 characters (the provider's readable
	// timestamp-naming threshold). The trailing hyphen separates the
	// rotation suffix from the group name.
	prefix := locals.MigName + "-"
	if len(prefix) > 37 {
		prefix = prefix[:36] + "-"
	}
	locals.TemplateNamePrefix = prefix

	locals.IsRegional = target.Spec.Region != ""
	if locals.IsRegional {
		locals.Location = target.Spec.Region
	} else {
		locals.Location = target.Spec.Zone
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module. These
	// are the VM labels the template stamps on every instance.
	locals.GcpLabels = map[string]string{}
	if target.Spec.Template != nil {
		for key, value := range target.Spec.Template.Labels {
			locals.GcpLabels[key] = value
		}
	}
	locals.GcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.MigName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpComputeMig.String())

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
