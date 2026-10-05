package module

import (
	"github.com/pkg/errors"
	awsec2instancev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsec2instance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions a single EC2 instance. The instance composes onto
// its neighbors instead of embedding them: the subnet it lives in, the
// security groups that guard it, the IAM instance profile that gives it an
// identity, the launch template it may inherit its shape from, and the KMS
// keys that encrypt its volumes all attach by reference -- this module
// creates exactly one AWS resource, the instance itself.
func Resources(ctx *pulumi.Context, iacInput *awsec2instancev1alpha1.AwsEc2InstanceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEc2Instance.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdInstance, err := instance(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create EC2 instance")
	}

	// Address outputs resolve to "" when absent (a private-only instance
	// has no public address) -- SDK string outputs carry the zero value
	// without ApplyT.
	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpArn, createdInstance.Arn)
	ctx.Export(OpInstanceState, createdInstance.InstanceState)
	ctx.Export(OpAvailabilityZone, createdInstance.AvailabilityZone)
	ctx.Export(OpPrivateIp, createdInstance.PrivateIp)
	ctx.Export(OpPrivateDns, createdInstance.PrivateDns)
	ctx.Export(OpPublicIp, createdInstance.PublicIp)
	ctx.Export(OpPublicDns, createdInstance.PublicDns)
	ctx.Export(OpPrimaryNetworkInterfaceId, createdInstance.PrimaryNetworkInterfaceId)

	return nil
}
