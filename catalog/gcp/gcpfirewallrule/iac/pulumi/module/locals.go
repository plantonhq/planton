package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpfirewallrulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirewallrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds frequently-used values derived from the IaC input.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpFirewallRule   *gcpfirewallrulev1alpha1.GcpFirewallRule
	GcpLabels         map[string]string
}

// initializeLocals populates the Locals struct from the IaC input.
// GCP compute firewall rules do not support labels directly, but we compute
// them here for consistency with the label strategy used across GCP components.
func initializeLocals(_ *pulumi.Context, iacInput *gcpfirewallrulev1alpha1.GcpFirewallRuleIacInput) *Locals {
	locals := &Locals{}

	locals.GcpFirewallRule = iacInput.Target

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: locals.GcpFirewallRule.Spec.RuleName,
		gcplabelkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_GcpFirewallRule.String()),
	}

	if locals.GcpFirewallRule.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpFirewallRule.Metadata.Org
	}

	if locals.GcpFirewallRule.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpFirewallRule.Metadata.Env
	}

	if locals.GcpFirewallRule.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpFirewallRule.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig

	return locals
}
