package module

import (
	gcpglobalforwardingrulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpglobalforwardingrule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpGlobalForwardingRule *gcpglobalforwardingrulev1alpha1.GcpGlobalForwardingRule

	// The cloud-side name defaults to metadata.name when the spec leaves
	// forwarding_rule_name empty — the same naming basis every kind uses.
	ForwardingRuleName string

	// The scope selector: a set spec.region builds the regional forwarding
	// rule, an empty one the global rule — the same switch the Terraform
	// module's count guards make.
	IsRegional bool

	// The scheme sent to GCP. The spec's NONE sentinel (Private Service
	// Connect) maps to the API's empty scheme; an unset spec value stays
	// empty here and the resource builder sends EXTERNAL explicitly (see the
	// scheme comment in the builders), on both scopes.
	LoadBalancingScheme string

	// Whether the spec explicitly chose the PSC form (scheme NONE) — the
	// one case where an empty scheme string must be SENT, not omitted.
	IsPrivateServiceConnect bool
}

func initializeLocals(ctx *pulumi.Context, stackInput *gcpglobalforwardingrulev1alpha1.GcpGlobalForwardingRuleStackInput) *Locals {
	target := stackInput.Target

	ruleName := target.Spec.ForwardingRuleName
	if ruleName == "" {
		ruleName = target.Metadata.Name
	}

	scheme := target.Spec.GetLoadBalancingScheme()
	isPsc := scheme == "NONE"
	if isPsc {
		scheme = ""
	}

	return &Locals{
		GcpGlobalForwardingRule: target,
		ForwardingRuleName:      ruleName,
		IsRegional:              target.Spec.Region != "",
		LoadBalancingScheme:     scheme,
		IsPrivateServiceConnect: isPsc,
	}
}
