package module

import (
	"fmt"
	"strings"

	gcphierarchicalfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphierarchicalfirewallpolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share -- the policy's short
// name, the rendered parent, and each association's rendered name and
// target.
type Locals struct {
	GcpHierarchicalFirewallPolicy *gcphierarchicalfirewallpolicyv1alpha1.GcpHierarchicalFirewallPolicy

	// ShortName is the spec's short_name, or metadata.name when the spec
	// leaves it empty -- the same naming basis every kind uses.
	ShortName string

	// Parent is Google's one-string parent, organizations/{id} or
	// folders/{id}, rendered from whichever arm of spec.parent is set.
	Parent string

	// AssociationNames holds each association's name, in declaration
	// order: the spec's name, or <short_name>-<n> when empty.
	AssociationNames []string

	// AssociationTargets holds each association's attachment target, in
	// declaration order, rendered like Parent.
	AssociationTargets []string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcphierarchicalfirewallpolicyv1alpha1.GcpHierarchicalFirewallPolicyStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	shortName := spec.ShortName
	if shortName == "" {
		shortName = target.Metadata.Name
	}

	associationNames := make([]string, 0, len(spec.Associations))
	associationTargets := make([]string, 0, len(spec.Associations))
	for i, association := range spec.Associations {
		name := association.Name
		if name == "" {
			name = fmt.Sprintf("%s-%d", shortName, i+1)
		}
		associationNames = append(associationNames, name)
		associationTargets = append(associationTargets,
			renderHierarchyNode(association.Target.OrganizationId, association.Target.FolderId.GetValue()))
	}

	return &Locals{
		GcpHierarchicalFirewallPolicy: target,
		ShortName:                     shortName,
		Parent:                        renderHierarchyNode(spec.Parent.OrganizationId, spec.Parent.FolderId.GetValue()),
		AssociationNames:              associationNames,
		AssociationTargets:            associationTargets,
	}
}

// renderHierarchyNode turns the exactly-one-arm organization/folder pair
// (proto-CEL-enforced) into Google's single resource-name string. A folder
// id that already carries its folders/ prefix passes through so a
// hand-written full name still works -- the same rule as GcpFolder.
func renderHierarchyNode(organizationId, folderId string) string {
	if organizationId != "" {
		return "organizations/" + organizationId
	}
	if strings.HasPrefix(folderId, "folders/") {
		return folderId
	}
	return "folders/" + folderId
}
