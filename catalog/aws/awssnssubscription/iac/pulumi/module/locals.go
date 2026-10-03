package module

import (
	awssnssubscriptionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssnssubscription/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// A subscription has no cloud-side name and is untaggable in AWS — it is
// identified by a server-assigned ARN — so unlike sibling modules there is no
// derived cloud name and no identity-tag map here. metadata.name only drives
// the Pulumi resource name.
type Locals struct {
	Target *awssnssubscriptionv1alpha1.AwsSnsSubscription
	Spec   *awssnssubscriptionv1alpha1.AwsSnsSubscriptionSpec
}

func initializeLocals(_ *pulumi.Context, in *awssnssubscriptionv1alpha1.AwsSnsSubscriptionIacInput) *Locals {
	return &Locals{
		Target: in.Target,
		Spec:   in.Target.Spec,
	}
}
