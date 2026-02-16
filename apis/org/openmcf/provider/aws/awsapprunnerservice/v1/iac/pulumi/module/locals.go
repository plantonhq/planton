package module

import (
	"strconv"

	awsapprunnerservicev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsapprunnerservice/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsAppRunnerService *awsapprunnerservicev1.AwsAppRunnerService
	AwsTags             map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsapprunnerservicev1.AwsAppRunnerServiceStackInput) *Locals {
	locals := &Locals{
		AwsAppRunnerService: in.Target,
	}

	if in.Target != nil {
		locals.AwsTags = map[string]string{
			awstagkeys.Resource:     strconv.FormatBool(true),
			awstagkeys.Organization: in.Target.Metadata.Org,
			awstagkeys.Environment:  in.Target.Metadata.Env,
			awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsAppRunnerService.String(),
			awstagkeys.ResourceId:   in.Target.Metadata.Id,
			awstagkeys.Name:         in.Target.Metadata.Name,
		}
	} else {
		locals.AwsTags = map[string]string{}
	}

	return locals
}
