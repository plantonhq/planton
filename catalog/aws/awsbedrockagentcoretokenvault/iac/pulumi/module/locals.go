package module

import (
	awsbedrockagentcoretokenvaultv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagentcoretokenvault/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// No identity-tag map here: the upstream token-vault-CMK resource
// carries no tags argument (a settings singleton, not a taggable
// object), and metadata.name never reaches AWS.
type Locals struct {
	Target *awsbedrockagentcoretokenvaultv1alpha1.AwsBedrockAgentCoreTokenVault
	Spec   *awsbedrockagentcoretokenvaultv1alpha1.AwsBedrockAgentCoreTokenVaultSpec

	// TokenVaultId is the vault targeted: the spec value, or AWS's one
	// default vault when unset.
	TokenVaultId string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockagentcoretokenvaultv1alpha1.AwsBedrockAgentCoreTokenVaultIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.TokenVaultId = locals.Spec.TokenVaultId
	if locals.TokenVaultId == "" {
		locals.TokenVaultId = "default"
	}

	return locals
}
