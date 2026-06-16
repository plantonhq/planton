package module

import (
	"github.com/pkg/errors"
	awsec2instancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsec2instance/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry‑point invoked by OpenMCF’s CLI.
// It wires provider credentials, initialises locals, and delegates
// to ec2Instance(...) to create the EC2 VM.
func Resources(ctx *pulumi.Context, stackInput *awsec2instancev1.AwsEc2InstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsEc2Instance.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := ec2Instance(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "create aws ec2 instance resource")
	}

	return nil
}
