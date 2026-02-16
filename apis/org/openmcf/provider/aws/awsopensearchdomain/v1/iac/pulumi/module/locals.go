package module

import (
	"strconv"

	awsopensearchdomainv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsopensearchdomain/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target  *awsopensearchdomainv1.AwsOpenSearchDomain
	Spec    *awsopensearchdomainv1.AwsOpenSearchDomainSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsopensearchdomainv1.AwsOpenSearchDomainStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsOpenSearchDomain.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
