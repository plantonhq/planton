package module

import (
	awssagemakerdomainv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awssagemakerdomain/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsSagemakerDomain *awssagemakerdomainv1.AwsSagemakerDomain
	Labels             map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awssagemakerdomainv1.AwsSagemakerDomainStackInput) *Locals {
	locals := &Locals{}

	locals.AwsSagemakerDomain = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsSagemakerDomain.Metadata.Org,
		"planton.org/environment":   locals.AwsSagemakerDomain.Metadata.Env,
		"planton.org/resource-kind": "AwsSagemakerDomain",
		"planton.org/resource-id":   locals.AwsSagemakerDomain.Metadata.Id,
	}

	return locals
}
