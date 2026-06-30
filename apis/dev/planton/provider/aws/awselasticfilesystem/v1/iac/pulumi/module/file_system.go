package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/efs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type FileSystemResult struct {
	FileSystem *efs.FileSystem
}

func fileSystem(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*FileSystemResult, error) {
	spec := locals.AwsElasticFileSystem.Spec
	name := locals.AwsElasticFileSystem.Metadata.Name

	args := &efs.FileSystemArgs{
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Encryption at rest.
	if spec.Encrypted {
		args.Encrypted = pulumi.Bool(true)
	}

	// Customer-managed KMS key.
	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	// Performance mode (generalPurpose or maxIO). Default: generalPurpose.
	if spec.PerformanceMode != "" {
		args.PerformanceMode = pulumi.StringPtr(spec.PerformanceMode)
	}

	// Throughput mode (bursting, provisioned, or elastic). Default: bursting.
	if spec.ThroughputMode != "" {
		args.ThroughputMode = pulumi.StringPtr(spec.ThroughputMode)
	}

	// Provisioned throughput (only valid with throughput_mode = "provisioned").
	if spec.ProvisionedThroughputInMibps > 0 {
		args.ProvisionedThroughputInMibps = pulumi.Float64Ptr(spec.ProvisionedThroughputInMibps)
	}

	// One Zone storage (single AZ). ForceNew.
	if spec.AvailabilityZoneName != "" {
		args.AvailabilityZoneName = pulumi.StringPtr(spec.AvailabilityZoneName)
	}

	// Lifecycle policies — up to 3 policies (IA, archive, primary storage class).
	var lifecyclePolicies efs.FileSystemLifecyclePolicyArray
	if spec.TransitionToIa != "" {
		lifecyclePolicies = append(lifecyclePolicies, &efs.FileSystemLifecyclePolicyArgs{
			TransitionToIa: pulumi.StringPtr(spec.TransitionToIa),
		})
	}
	if spec.TransitionToArchive != "" {
		lifecyclePolicies = append(lifecyclePolicies, &efs.FileSystemLifecyclePolicyArgs{
			TransitionToArchive: pulumi.StringPtr(spec.TransitionToArchive),
		})
	}
	if spec.TransitionToPrimaryStorageClass != "" {
		lifecyclePolicies = append(lifecyclePolicies, &efs.FileSystemLifecyclePolicyArgs{
			TransitionToPrimaryStorageClass: pulumi.StringPtr(spec.TransitionToPrimaryStorageClass),
		})
	}
	if len(lifecyclePolicies) > 0 {
		args.LifecyclePolicies = lifecyclePolicies
	}

	createdFs, err := efs.NewFileSystem(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create efs file system")
	}

	return &FileSystemResult{FileSystem: createdFs}, nil
}
