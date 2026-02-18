package module

import (
	"github.com/pkg/errors"
	awselasticfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awselasticfilesystem/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awselasticfilesystemv1.AwsElasticFileSystemStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsElasticFileSystem.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsElasticFileSystem.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// --- Phase 1: File system ---
	fsResult, err := fileSystem(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create elastic file system")
	}

	// --- Phase 2: Mount targets (one per subnet) ---
	mtResults, err := mountTargets(ctx, locals, provider, fsResult.FileSystem)
	if err != nil {
		return errors.Wrap(err, "failed to create mount targets")
	}

	// --- Phase 3: Access points ---
	apResults, err := accessPoints(ctx, locals, provider, fsResult.FileSystem)
	if err != nil {
		return errors.Wrap(err, "failed to create access points")
	}

	// --- Phase 4: Policies (backup + resource policy) ---
	if err := policies(ctx, locals, provider, fsResult.FileSystem); err != nil {
		return errors.Wrap(err, "failed to create policies")
	}

	// --- Exports ---
	ctx.Export(OpFileSystemId, fsResult.FileSystem.ID())
	ctx.Export(OpFileSystemArn, fsResult.FileSystem.Arn)
	ctx.Export(OpDnsName, fsResult.FileSystem.DnsName)
	ctx.Export(OpMountTargetIds, mtResults.MountTargetIds)
	ctx.Export(OpMountTargetIps, mtResults.MountTargetIps)
	ctx.Export(OpMountTargetDnsNames, mtResults.MountTargetDnsNames)
	ctx.Export(OpAccessPointIds, apResults.AccessPointIds)
	ctx.Export(OpAccessPointArns, apResults.AccessPointArns)

	return nil
}
