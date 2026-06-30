package module

import (
	awsdocumentdbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsdocumentdb/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsDocumentDb *awsdocumentdbv1.AwsDocumentDb
	Labels        map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsdocumentdbv1.AwsDocumentDbStackInput) *Locals {
	locals := &Locals{}

	locals.AwsDocumentDb = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsDocumentDb.Metadata.Org,
		"planton.org/environment":   locals.AwsDocumentDb.Metadata.Env,
		"planton.org/resource-kind": "AwsDocumentDb",
		"planton.org/resource-id":   locals.AwsDocumentDb.Metadata.Id,
	}

	return locals
}
