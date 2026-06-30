package module

import (
	"strconv"

	awsec2instancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsec2instance/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals acts like Terraform “locals”, grouping derived values
// that the rest of the module re‑uses.
type Locals struct {
	AwsEc2Instance *awsec2instancev1.AwsEc2Instance
	AwsTags        map[string]string
}

// initializeLocals converts the stack‑input into Locals.
func initializeLocals(ctx *pulumi.Context, stackInput *awsec2instancev1.AwsEc2InstanceStackInput) *Locals {
	locals := &Locals{
		AwsEc2Instance: stackInput.Target,
	}

	// Base tags (always present)
	locals.AwsTags = map[string]string{
		awstagkeys.Environment:  locals.AwsEc2Instance.Metadata.Env,
		awstagkeys.Name:         locals.AwsEc2Instance.Metadata.Name,
		awstagkeys.Organization: locals.AwsEc2Instance.Metadata.Org,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.ResourceId:   locals.AwsEc2Instance.Metadata.Id,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsEc2Instance.String(),
	}

	// Merge user‑supplied tags (override on collision)
	for k, v := range locals.AwsEc2Instance.Spec.Tags {
		locals.AwsTags[k] = v
	}

	return locals
}
