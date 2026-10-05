package module

import (
	"strconv"

	awshttpapidomainv1alpha1 "github.com/plantonhq/planton/catalog/aws/awshttpapidomain/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the domain name.
type Locals struct {
	AwsHttpApiDomain *awshttpapidomainv1alpha1.AwsHttpApiDomain
	AwsTags          map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awshttpapidomainv1alpha1.AwsHttpApiDomainIacInput) *Locals {
	locals := &Locals{}
	locals.AwsHttpApiDomain = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsHttpApiDomain.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsHttpApiDomain.Metadata.Org,
		awstagkeys.Environment:  locals.AwsHttpApiDomain.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsHttpApiDomain.String(),
		awstagkeys.ResourceId:   locals.AwsHttpApiDomain.Metadata.Id,
	}

	return locals
}
