package module

import (
	ocinetworkfirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkfirewall/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciNetworkFirewall *ocinetworkfirewallv1.OciNetworkFirewall
	DisplayName        string
	PolicyDisplayName  string
	FreeformTags       map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocinetworkfirewallv1.OciNetworkFirewallStackInput) *Locals {
	locals := &Locals{}
	locals.OciNetworkFirewall = stackInput.Target

	locals.DisplayName = stackInput.Target.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.PolicyDisplayName = stackInput.Target.Spec.Policy.DisplayName
	if locals.PolicyDisplayName == "" {
		locals.PolicyDisplayName = locals.DisplayName + "-policy"
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciNetworkFirewall.String(),
		"resource_id":   stackInput.Target.Metadata.Id,
	}
	if stackInput.Target.Metadata.Org != "" {
		locals.FreeformTags["organization"] = stackInput.Target.Metadata.Org
	}
	if stackInput.Target.Metadata.Env != "" {
		locals.FreeformTags["environment"] = stackInput.Target.Metadata.Env
	}
	for k, v := range stackInput.Target.Metadata.Labels {
		locals.FreeformTags[k] = v
	}

	return locals
}
