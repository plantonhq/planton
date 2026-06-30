package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/efs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// MountTargetResults holds the per-subnet mount target outputs keyed by subnet ID.
type MountTargetResults struct {
	MountTargetIds      pulumi.StringMap
	MountTargetIps      pulumi.StringMap
	MountTargetDnsNames pulumi.StringMap
}

func mountTargets(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, fs *efs.FileSystem) (*MountTargetResults, error) {
	spec := locals.AwsElasticFileSystem.Spec

	// Build security group list from StringValueOrRef fields.
	var securityGroups pulumi.StringArray
	for _, sg := range spec.SecurityGroupIds {
		if sg.GetValue() != "" {
			securityGroups = append(securityGroups, pulumi.String(sg.GetValue()))
		}
	}

	results := &MountTargetResults{
		MountTargetIds:      pulumi.StringMap{},
		MountTargetIps:      pulumi.StringMap{},
		MountTargetDnsNames: pulumi.StringMap{},
	}

	for i, subnetRef := range spec.SubnetIds {
		subnetId := subnetRef.GetValue()
		if subnetId == "" {
			continue
		}

		resourceName := fmt.Sprintf("mt-%d", i)

		args := &efs.MountTargetArgs{
			FileSystemId: fs.ID(),
			SubnetId:     pulumi.String(subnetId),
		}

		if len(securityGroups) > 0 {
			args.SecurityGroups = securityGroups
		}

		mt, err := efs.NewMountTarget(ctx, resourceName, args, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create mount target for subnet %s", subnetId)
		}

		results.MountTargetIds[subnetId] = mt.ID().ToStringOutput()
		results.MountTargetIps[subnetId] = mt.IpAddress
		results.MountTargetDnsNames[subnetId] = mt.MountTargetDnsName
	}

	return results, nil
}
