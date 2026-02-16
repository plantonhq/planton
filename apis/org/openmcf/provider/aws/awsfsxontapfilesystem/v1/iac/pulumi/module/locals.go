package module

import (
	"strconv"

	awsfsxontapfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxontapfilesystem/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOntapFileSystem *awsfsxontapfilesystemv1.AwsFsxOntapFileSystem
	AwsTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsfsxontapfilesystemv1.AwsFsxOntapFileSystemStackInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOntapFileSystem = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOntapFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOntapFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsFsxOntapFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOntapFileSystem.Metadata.Id,
	}

	return locals
}
