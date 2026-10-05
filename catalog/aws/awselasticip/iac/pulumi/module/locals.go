package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awselasticipv1alpha1 "github.com/plantonhq/planton/catalog/aws/awselasticip/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsElasticIp *awselasticipv1alpha1.AwsElasticIp
	AwsTags      map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awselasticipv1alpha1.AwsElasticIpIacInput) *Locals {
	locals := &Locals{}
	locals.AwsElasticIp = iacInput.Target

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsElasticIp.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsElasticIp.Metadata.Org,
		awstagkeys.Environment:  locals.AwsElasticIp.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsElasticIp.String(),
		awstagkeys.ResourceId:   locals.AwsElasticIp.Metadata.Id,
	}

	return locals
}
