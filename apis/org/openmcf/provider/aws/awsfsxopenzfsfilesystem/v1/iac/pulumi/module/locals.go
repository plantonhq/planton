package module

import (
	"strconv"

	awsfsxopenzfsfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxopenzfsfilesystem/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOpenzfsFileSystem *awsfsxopenzfsfilesystemv1.AwsFsxOpenzfsFileSystem
	AwsTags                 map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsfsxopenzfsfilesystemv1.AwsFsxOpenzfsFileSystemStackInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOpenzfsFileSystem = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOpenzfsFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOpenzfsFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsFsxOpenzfsFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOpenzfsFileSystem.Metadata.Id,
	}

	return locals
}
