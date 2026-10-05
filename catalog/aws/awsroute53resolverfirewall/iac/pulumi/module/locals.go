package module

import (
	"strconv"

	awsroute53resolverfirewallv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsroute53resolverfirewall/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsroute53resolverfirewallv1alpha1.AwsRoute53ResolverFirewall
	Spec   *awsroute53resolverfirewallv1alpha1.AwsRoute53ResolverFirewallSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsroute53resolverfirewallv1alpha1.AwsRoute53ResolverFirewallIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRoute53ResolverFirewall.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
