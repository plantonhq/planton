package module

import (
	gcpcloudidentitygroupv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudidentitygroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share.
type Locals struct {
	GcpCloudIdentityGroup *gcpcloudidentitygroupv1alpha1.GcpCloudIdentityGroup

	// DisplayName is the spec's display_name, or metadata.name when the
	// spec leaves it empty — the same naming basis every kind uses.
	DisplayName string

	// GroupLabels is the label map Google requires: every Google Group
	// carries the discussion-forum label (empty value), and a security
	// group adds the security label on top. Derived from spec.security --
	// never a user-facing map, because Google accepts nothing else here.
	GroupLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcloudidentitygroupv1alpha1.GcpCloudIdentityGroupStackInput) *Locals {
	target := stackInput.Target

	displayName := target.Spec.DisplayName
	if displayName == "" {
		displayName = target.Metadata.Name
	}

	labels := map[string]string{"cloudidentity.googleapis.com/groups.discussion_forum": ""}
	if target.Spec.Security {
		labels["cloudidentity.googleapis.com/groups.security"] = ""
	}

	return &Locals{
		GcpCloudIdentityGroup: target,
		DisplayName:           displayName,
		GroupLabels:           labels,
	}
}
