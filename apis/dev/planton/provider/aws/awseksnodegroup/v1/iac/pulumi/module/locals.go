package module

import (
	awseksnodegroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awseksnodegroup/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildNodeGroupArgs constructs the EKS NodeGroup arguments from the spec.
// This function centralizes the argument mapping logic, making the main
// resource creation cleaner and more maintainable.
func buildNodeGroupArgs(spec *awseksnodegroupv1.AwsEksNodeGroupSpec) *eks.NodeGroupArgs {
	// Convert subnet IDs from StringValueOrRef to pulumi.StringArray
	var subnetIds pulumi.StringArray
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
	}

	// Build scaling configuration
	scaling := &eks.NodeGroupScalingConfigArgs{
		MinSize:     pulumi.Int(int(spec.Scaling.MinSize)),
		MaxSize:     pulumi.Int(int(spec.Scaling.MaxSize)),
		DesiredSize: pulumi.Int(int(spec.Scaling.DesiredSize)),
	}

	// Map capacity type enum to AWS API string format
	capacityType := getCapacityType(spec.CapacityType)

	// Build base node group arguments
	args := &eks.NodeGroupArgs{
		ClusterName:   pulumi.String(spec.ClusterName.GetValue()),
		NodeRoleArn:   pulumi.String(spec.NodeRoleArn.GetValue()),
		SubnetIds:     subnetIds,
		InstanceTypes: pulumi.ToStringArray([]string{spec.InstanceType}),
		ScalingConfig: scaling,
		CapacityType:  capacityType,
		DiskSize:      pulumi.Int(int(spec.DiskSizeGb)),
		Labels:        pulumi.ToStringMap(spec.Labels),
	}

	// Add optional remote access configuration if SSH key is provided
	if spec.SshKeyName != "" {
		args.RemoteAccess = &eks.NodeGroupRemoteAccessArgs{
			Ec2SshKey: pulumi.String(spec.SshKeyName),
		}
	}

	return args
}

// getCapacityType converts the protobuf enum to AWS API string format.
// AWS expects "ON_DEMAND" or "SPOT" in uppercase with underscore.
func getCapacityType(capacityType awseksnodegroupv1.AwsEksNodeGroupCapacityType) pulumi.StringInput {
	if capacityType == awseksnodegroupv1.AwsEksNodeGroupCapacityType_spot {
		return pulumi.String("SPOT")
	}
	return pulumi.String("ON_DEMAND")
}

// getDefaultDiskSize returns the recommended default disk size in GiB.
// This follows AWS best practices: default 20GB is often insufficient,
// so we recommend 100GB for container images and application data.
func getDefaultDiskSize() int {
	return 100
}
