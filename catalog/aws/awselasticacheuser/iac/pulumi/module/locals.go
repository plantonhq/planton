package module

import (
	"strconv"

	awselasticacheuserv1alpha1 "github.com/plantonhq/planton/catalog/aws/awselasticacheuser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsElasticacheUser *awselasticacheuserv1alpha1.AwsElasticacheUser

	// UserId is metadata.name -- the AWS user id is create-time immutable,
	// and metadata.name is the naming basis both engines share so a manifest
	// deploys identically on either.
	UserId string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awselasticacheuserv1alpha1.AwsElasticacheUserIacInput) *Locals {
	locals := &Locals{}
	locals.AwsElasticacheUser = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.UserId = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsElasticacheUser.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
