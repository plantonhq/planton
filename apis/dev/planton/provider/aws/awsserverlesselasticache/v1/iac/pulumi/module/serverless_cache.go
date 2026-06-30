package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/elasticache"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serverlessCache creates the ElastiCache Serverless cache and exports outputs.
func serverlessCache(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &elasticache.ServerlessCacheArgs{
		Engine: pulumi.String(spec.Engine),
		Name:   pulumi.String(locals.Target.Metadata.Id),
		Tags:   pulumi.ToStringMap(locals.AwsTags),
	}

	// Description
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	// Major engine version
	if spec.MajorEngineVersion != "" {
		args.MajorEngineVersion = pulumi.String(spec.MajorEngineVersion)
	}

	// -------------------------------------------------------------------
	// Cache usage limits (flattened in spec, nested in AWS)
	// -------------------------------------------------------------------

	hasDataStorage := spec.DataStorageMinGb > 0 || spec.DataStorageMaxGb > 0
	hasEcpu := spec.EcpuMin > 0 || spec.EcpuMax > 0

	if hasDataStorage || hasEcpu {
		limits := &elasticache.ServerlessCacheCacheUsageLimitsArgs{}

		if hasDataStorage {
			ds := &elasticache.ServerlessCacheCacheUsageLimitsDataStorageArgs{
				Unit: pulumi.String("GB"),
			}
			if spec.DataStorageMinGb > 0 {
				ds.Minimum = pulumi.Int(int(spec.DataStorageMinGb))
			}
			if spec.DataStorageMaxGb > 0 {
				ds.Maximum = pulumi.Int(int(spec.DataStorageMaxGb))
			}
			limits.DataStorage = ds
		}

		if hasEcpu {
			ecpu := &elasticache.ServerlessCacheCacheUsageLimitsEcpuPerSecondArgs{}
			if spec.EcpuMin > 0 {
				ecpu.Minimum = pulumi.Int(int(spec.EcpuMin))
			}
			if spec.EcpuMax > 0 {
				ecpu.Maximum = pulumi.Int(int(spec.EcpuMax))
			}
			limits.EcpuPerSeconds = elasticache.ServerlessCacheCacheUsageLimitsEcpuPerSecondArray{ecpu}
		}

		args.CacheUsageLimits = limits
	}

	// -------------------------------------------------------------------
	// Networking
	// -------------------------------------------------------------------

	var subnetIds pulumi.StringArray
	for _, s := range spec.SubnetIds {
		if s.GetValue() != "" {
			subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
		}
	}
	if len(subnetIds) > 0 {
		args.SubnetIds = subnetIds
	}

	var sgIds pulumi.StringArray
	for _, sg := range spec.SecurityGroupIds {
		if sg.GetValue() != "" {
			sgIds = append(sgIds, pulumi.String(sg.GetValue()))
		}
	}
	if len(sgIds) > 0 {
		args.SecurityGroupIds = sgIds
	}

	// -------------------------------------------------------------------
	// Encryption
	// -------------------------------------------------------------------

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// -------------------------------------------------------------------
	// Snapshots (Redis/Valkey only — CEL guards prevent Memcached usage)
	// -------------------------------------------------------------------

	if spec.DailySnapshotTime != "" {
		args.DailySnapshotTime = pulumi.String(spec.DailySnapshotTime)
	}

	if spec.SnapshotRetentionLimit > 0 {
		args.SnapshotRetentionLimit = pulumi.Int(int(spec.SnapshotRetentionLimit))
	}

	// -------------------------------------------------------------------
	// Authentication (Redis/Valkey only — CEL guards prevent Memcached usage)
	// -------------------------------------------------------------------

	if spec.UserGroupId != "" {
		args.UserGroupId = pulumi.String(spec.UserGroupId)
	}

	// -------------------------------------------------------------------
	// Create serverless cache
	// -------------------------------------------------------------------

	cache, err := elasticache.NewServerlessCache(ctx, "serverless-cache", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create serverless cache")
	}

	// -------------------------------------------------------------------
	// Export outputs
	// -------------------------------------------------------------------

	ctx.Export(OpArn, cache.Arn)
	ctx.Export(OpFullEngineVersion, cache.FullEngineVersion)
	ctx.Export(OpName, cache.Name)

	// Endpoint and reader endpoint are arrays of objects; extract the first element.
	ctx.Export(OpEndpointAddress, cache.Endpoints.Index(pulumi.Int(0)).Address())
	ctx.Export(OpEndpointPort, cache.Endpoints.Index(pulumi.Int(0)).Port())
	ctx.Export(OpReaderEndpointAddress, cache.ReaderEndpoints.Index(pulumi.Int(0)).Address())
	ctx.Export(OpReaderEndpointPort, cache.ReaderEndpoints.Index(pulumi.Int(0)).Port())

	return nil
}
