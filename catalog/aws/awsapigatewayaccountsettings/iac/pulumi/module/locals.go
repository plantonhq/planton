package module

import (
	awsapigatewayaccountsettingsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapigatewayaccountsettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// No identity-tag map here: the upstream account resource carries no
// tags argument (a settings singleton, not a taggable object), and
// metadata.name never reaches AWS.
type Locals struct {
	Target *awsapigatewayaccountsettingsv1alpha1.AwsApiGatewayAccountSettings
	Spec   *awsapigatewayaccountsettingsv1alpha1.AwsApiGatewayAccountSettingsSpec
}

func initializeLocals(_ *pulumi.Context, in *awsapigatewayaccountsettingsv1alpha1.AwsApiGatewayAccountSettingsIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	return locals
}
