package module

import (
	awsbedrockinvocationloggingv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockinvocationlogging/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// No identity-tag map here: the upstream logging-configuration
// resource carries no tags argument (a settings singleton, not a
// taggable object), and metadata.name never reaches AWS.
type Locals struct {
	Target *awsbedrockinvocationloggingv1alpha1.AwsBedrockInvocationLogging
	Spec   *awsbedrockinvocationloggingv1alpha1.AwsBedrockInvocationLoggingSpec
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockinvocationloggingv1alpha1.AwsBedrockInvocationLoggingIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	return locals
}
