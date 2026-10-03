package module

import (
	awscloudwatchlogaccountpolicyv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchlogaccountpolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awscloudwatchlogaccountpolicyv1alpha1.AwsCloudwatchLogAccountPolicy
	Spec   *awscloudwatchlogaccountpolicyv1alpha1.AwsCloudwatchLogAccountPolicySpec
}

// Account policies are untaggable at AWS (the resource has no tags
// argument), so this module carries no tag map - the one deliberate
// absence against the catalog's tag convention (mirrored in the
// Terraform module).
func initializeLocals(_ *pulumi.Context, in *awscloudwatchlogaccountpolicyv1alpha1.AwsCloudwatchLogAccountPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	return locals
}
