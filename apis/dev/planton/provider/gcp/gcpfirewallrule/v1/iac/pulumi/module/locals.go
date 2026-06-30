package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpfirewallrulev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpfirewallrule/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds frequently-used values derived from the stack input.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpFirewallRule   *gcpfirewallrulev1.GcpFirewallRule
	GcpLabels         map[string]string
}

// initializeLocals populates the Locals struct from the stack input.
// GCP compute firewall rules do not support labels directly, but we compute
// them here for consistency with the label strategy used across GCP components.
func initializeLocals(_ *pulumi.Context, stackInput *gcpfirewallrulev1.GcpFirewallRuleStackInput) *Locals {
	locals := &Locals{}

	locals.GcpFirewallRule = stackInput.Target

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: locals.GcpFirewallRule.Spec.RuleName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpFirewallRule.String()),
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

	locals.GcpProviderConfig = stackInput.ProviderConfig

	return locals
}
