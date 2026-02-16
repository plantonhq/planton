package module

import (
	"strconv"

	awsfsxwindowsfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxwindowsfilesystem/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxWindowsFileSystem *awsfsxwindowsfilesystemv1.AwsFsxWindowsFileSystem
	AwsTags                 map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsfsxwindowsfilesystemv1.AwsFsxWindowsFileSystemStackInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxWindowsFileSystem = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxWindowsFileSystem.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxWindowsFileSystem.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsFsxWindowsFileSystem.String(),
		awstagkeys.ResourceId:   locals.AwsFsxWindowsFileSystem.Metadata.Id,
	}

	return locals
}
