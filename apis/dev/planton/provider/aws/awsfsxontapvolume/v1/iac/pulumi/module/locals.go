package module

import (
	"strconv"

	awsfsxontapvolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsfsxontapvolume/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOntapVolume *awsfsxontapvolumev1.AwsFsxOntapVolume
	AwsTags           map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsfsxontapvolumev1.AwsFsxOntapVolumeStackInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOntapVolume = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOntapVolume.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOntapVolume.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsFsxOntapVolume.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOntapVolume.Metadata.Id,
	}

	return locals
}
