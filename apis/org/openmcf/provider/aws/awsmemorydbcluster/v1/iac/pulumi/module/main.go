package module

import (
	"github.com/pkg/errors"
	awsmemorydbclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmemorydbcluster/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS MemoryDB cluster resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsmemorydbclusterv1.AwsMemorydbClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.Target.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.Target.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Subnet group (only when subnet_ids provided)
	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "subnet group")
	}

	// Parameter group (when parameters provided with family)
	createdParamGroup, err := parameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "parameter group")
	}

	// MemoryDB cluster (always created)
	if err := cluster(ctx, locals, provider, createdSubnetGroup, createdParamGroup); err != nil {
		return errors.Wrap(err, "memorydb cluster")
	}

	return nil
}
