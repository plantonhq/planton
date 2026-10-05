package module

import (
	"strconv"

	awsredshiftserverlessworkgroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsredshiftserverlessworkgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRedshiftServerlessWorkgroup *awsredshiftserverlessworkgroupv1alpha1.AwsRedshiftServerlessWorkgroup

	// WorkgroupName is metadata.name -- create-only in AWS, and the
	// basis both engines share so a manifest deploys identically on
	// either.
	WorkgroupName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsredshiftserverlessworkgroupv1alpha1.AwsRedshiftServerlessWorkgroupIacInput) *Locals {
	locals := &Locals{}
	locals.AwsRedshiftServerlessWorkgroup = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.WorkgroupName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRedshiftServerlessWorkgroup.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
