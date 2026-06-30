package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/redshift"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// redshiftLogging creates a Redshift Logging resource when spec.Logging is configured.
func redshiftLogging(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	cluster *redshift.Cluster,
) error {
	spec := locals.AwsRedshiftCluster.Spec
	if spec == nil || spec.Logging == nil {
		return nil
	}

	loggingSpec := spec.Logging

	args := &redshift.LoggingArgs{
		ClusterIdentifier: cluster.ClusterIdentifier,
	}

	if loggingSpec.LogDestinationType != "" {
		args.LogDestinationType = pulumi.String(loggingSpec.LogDestinationType)
	}

	if loggingSpec.S3BucketName != "" {
		args.BucketName = pulumi.String(loggingSpec.S3BucketName)
	}

	if loggingSpec.S3KeyPrefix != "" {
		args.S3KeyPrefix = pulumi.String(loggingSpec.S3KeyPrefix)
	}

	if len(loggingSpec.LogExports) > 0 {
		args.LogExports = pulumi.ToStringArray(loggingSpec.LogExports)
	}

	_, err := redshift.NewLogging(ctx, "cluster-logging", args,
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{cluster}),
	)
	if err != nil {
		return errors.Wrap(err, "create redshift logging")
	}

	return nil
}
