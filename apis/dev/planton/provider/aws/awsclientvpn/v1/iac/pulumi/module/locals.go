package module

import (
	awsclientvpnv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsclientvpn/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the simple, terraform‑style locals pattern seen in aws_vpc.
type Locals struct {
	AwsClientVpn *awsclientvpnv1.AwsClientVpn
	AwsTags      map[string]string
}

// initializeLocals prepares the Locals struct and common AWS tags.
func initializeLocals(ctx *pulumi.Context, stackInput *awsclientvpnv1.AwsClientVpnStackInput) *Locals {
	locals := &Locals{
		AwsClientVpn: stackInput.Target,
		AwsTags: map[string]string{
			awstagkeys.Environment:  stackInput.Target.Metadata.Env,
			awstagkeys.Name:         stackInput.Target.Metadata.Name,
			awstagkeys.Organization: stackInput.Target.Metadata.Org,
			awstagkeys.Resource:     "true",
			awstagkeys.ResourceId:   stackInput.Target.Metadata.Id,
			awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsClientVpn.String(),
		},
	}
	return locals
}
