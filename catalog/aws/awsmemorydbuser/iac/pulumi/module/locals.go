package module

import (
	"strconv"

	awsmemorydbuserv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmemorydbuser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsMemorydbUser *awsmemorydbuserv1alpha1.AwsMemorydbUser

	// UserName is metadata.name -- in MemoryDB the user name IS the user's
	// single identity (there is no separate user id), it is create-time
	// immutable, and metadata.name is the naming basis both engines share
	// so a manifest deploys identically on either. AWS caps it at 40
	// characters.
	UserName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsmemorydbuserv1alpha1.AwsMemorydbUserIacInput) *Locals {
	locals := &Locals{}
	locals.AwsMemorydbUser = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.UserName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsMemorydbUser.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
