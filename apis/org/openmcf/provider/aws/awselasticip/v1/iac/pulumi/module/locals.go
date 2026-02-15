package module

import (
	"strconv"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"

	awselasticipv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awselasticip/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsElasticIp *awselasticipv1.AwsElasticIp
	AwsTags      map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awselasticipv1.AwsElasticIpStackInput) *Locals {
	locals := &Locals{}
	locals.AwsElasticIp = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsElasticIp.Metadata.Org,
		awstagkeys.Environment:  locals.AwsElasticIp.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsElasticIp.String(),
		awstagkeys.ResourceId:   locals.AwsElasticIp.Metadata.Id,
	}

	return locals
}
