package module

import (
	"strconv"

	awsfsxlustrefilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxlustrefilesystem/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxLustreFileSystem *awsfsxlustrefilesystemv1.AwsFsxLustreFileSystem
	AwsTags                map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsfsxlustrefilesystemv1.AwsFsxLustreFileSystemStackInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxLustreFileSystem = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxLustreFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxLustreFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsFsxLustreFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsFsxLustreFileSystem.Metadata.Id,
	}

	return locals
}
