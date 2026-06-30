package module

import (
	"strconv"

	awselasticfilesystemv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awselasticfilesystem/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsElasticFileSystem *awselasticfilesystemv1.AwsElasticFileSystem
	AwsTags              map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awselasticfilesystemv1.AwsElasticFileSystemStackInput) *Locals {
	locals := &Locals{}
	locals.AwsElasticFileSystem = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsElasticFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsElasticFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsElasticFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsElasticFileSystem.Metadata.Id,
	}

	return locals
}
