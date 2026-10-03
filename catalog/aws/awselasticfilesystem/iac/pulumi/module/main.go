package module

import (
	"github.com/pkg/errors"
	awselasticfilesystemv1alpha1 "github.com/plantonhq/planton/catalog/aws/awselasticfilesystem/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awselasticfilesystemv1alpha1.AwsElasticFileSystemIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsElasticFileSystem.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// --- Phase 1: File system (encryption, throughput, lifecycle, protection) ---
	fsResult, err := fileSystem(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create elastic file system")
	}

	// --- Phase 2: Mount targets (one per subnet / AZ) ---
	mtResults, err := mountTargets(ctx, locals, provider, fsResult.FileSystem)
	if err != nil {
		return errors.Wrap(err, "failed to create mount targets")
	}

	// --- Phase 3: Policies (backup + resource policy) ---
	if err := policies(ctx, locals, provider, fsResult.FileSystem); err != nil {
		return errors.Wrap(err, "failed to create policies")
	}

	// --- Phase 4: Replication (cross-region / cross-AZ DR) ---
	replResult, err := replication(ctx, locals, provider, fsResult.FileSystem)
	if err != nil {
		return errors.Wrap(err, "failed to create replication configuration")
	}

	// --- Exports ---
	ctx.Export(OpFileSystemId, fsResult.FileSystem.ID())
	ctx.Export(OpFileSystemArn, fsResult.FileSystem.Arn)
	ctx.Export(OpDnsName, fsResult.FileSystem.DnsName)
	ctx.Export(OpMountTargetIds, mtResults.MountTargetIds)
	ctx.Export(OpMountTargetIps, mtResults.MountTargetIps)
	ctx.Export(OpMountTargetIpv6Addresses, mtResults.MountTargetIpv6Addresses)
	ctx.Export(OpMountTargetDnsNames, mtResults.MountTargetDnsNames)

	// The replica's file system ID is only known when replication was
	// requested; export the empty string otherwise so the output shape stays
	// stable for the conformance guard. Destination().FileSystemId() is a
	// PtrOutput — Elem() dereferences to its zero value when nil, so this
	// chain never panics (no ApplyT needed).
	if replResult.ReplicationConfiguration != nil {
		ctx.Export(OpReplicationDestinationFileSystemId,
			replResult.ReplicationConfiguration.Destination.FileSystemId().Elem())
	} else {
		ctx.Export(OpReplicationDestinationFileSystemId, pulumi.String(""))
	}

	return nil
}
