package module

import (
	"fmt"

	gcpnetworkfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpnetworkfirewallpolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share -- the policy name, the
// arm switch, the resolved project, and each association's rendered name.
type Locals struct {
	GcpNetworkFirewallPolicy *gcpnetworkfirewallpolicyv1alpha1.GcpNetworkFirewallPolicy

	// PolicyName is the spec's policy_name, or metadata.name when the spec
	// leaves it empty -- the same naming basis every kind uses.
	PolicyName string

	// IsRegional selects the regional resource family when spec.region is
	// set; the global family otherwise. Every resource keys off it.
	IsRegional bool

	// ProjectId is the resolved project, or empty for the provider's
	// default project -- the ambient-project contract every GCP kind
	// honors.
	ProjectId string

	// AssociationNames holds each association's name, in declaration
	// order: the spec's name, or <policy_name>-<n> when empty.
	AssociationNames []string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpnetworkfirewallpolicyv1alpha1.GcpNetworkFirewallPolicyStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	policyName := spec.PolicyName
	if policyName == "" {
		policyName = target.Metadata.Name
	}

	associationNames := make([]string, 0, len(spec.Associations))
	for i, association := range spec.Associations {
		name := association.Name
		if name == "" {
			name = fmt.Sprintf("%s-%d", policyName, i+1)
		}
		associationNames = append(associationNames, name)
	}

	return &Locals{
		GcpNetworkFirewallPolicy: target,
		PolicyName:               policyName,
		IsRegional:               spec.Region != "",
		ProjectId:                spec.ProjectId.GetValue(),
		AssociationNames:         associationNames,
	}
}
