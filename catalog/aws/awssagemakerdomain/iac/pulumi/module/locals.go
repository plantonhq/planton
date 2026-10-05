package module

import (
	"strconv"

	awssagemakerdomainv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakerdomain/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target     *awssagemakerdomainv1alpha1.AwsSagemakerDomain
	Spec       *awssagemakerdomainv1alpha1.AwsSagemakerDomainSpec
	AwsTags    map[string]string
	DomainName string
}

func initializeLocals(ctx *pulumi.Context, in *awssagemakerdomainv1alpha1.AwsSagemakerDomainIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	// The domain's cloud name is metadata.name. AWS constrains it to 1-63
	// characters of [0-9A-Za-z-] and makes it create-time-immutable (changing
	// the name replaces the domain), which is why it is not spec surface.
	// Same basis as the Terraform module.
	locals.DomainName = in.Target.Metadata.Name

	// Resource-identity tags match the Terraform module key-for-key. Identity
	// tagging is the only tagging surface this module manages; with
	// spec.tag_propagation = "ENABLED", AWS copies these onto the apps,
	// spaces, and user profiles created inside the domain.
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerDomain.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
