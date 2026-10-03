package module

import (
	awsconfigconformancepackv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsconfigconformancepack/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// No AwsTags here: neither conformance-pack provider resource carries
// a tags argument (the one untaggable surface in the Config family).
type Locals struct {
	Target *awsconfigconformancepackv1alpha1.AwsConfigConformancePack
	Spec   *awsconfigconformancepackv1alpha1.AwsConfigConformancePackSpec
}

func initializeLocals(_ *pulumi.Context, in *awsconfigconformancepackv1alpha1.AwsConfigConformancePackIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	return locals
}
