package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/batch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func computeEnvironment(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
) (*batch.ComputeEnvironment, error) {
	spec := locals.AwsBatchComputeEnvironment.Spec
	cr := spec.ComputeResources

	// Build compute_resources block
	computeResources := &batch.ComputeEnvironmentComputeResourcesArgs{
		Type:     pulumi.String(cr.Type),
		MaxVcpus: pulumi.Int(cr.MaxVcpus),
	}

	// Subnets (required)
	var subnetIds pulumi.StringArray
	for _, s := range cr.SubnetIds {
		if s.GetValue() != "" {
			subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
		}
	}
	computeResources.Subnets = subnetIds

	// Security groups
	var sgIds pulumi.StringArray
	for _, sg := range cr.SecurityGroupIds {
		if sg.GetValue() != "" {
			sgIds = append(sgIds, pulumi.String(sg.GetValue()))
		}
	}
	if len(sgIds) > 0 {
		computeResources.SecurityGroupIds = sgIds
	}

	// EC2/SPOT-specific fields
	if cr.Type == "EC2" || cr.Type == "SPOT" {
		if cr.GetMinVcpus() > 0 || cr.MinVcpus != nil {
			computeResources.MinVcpus = pulumi.IntPtr(int(cr.GetMinVcpus()))
		}
		if cr.DesiredVcpus > 0 {
			computeResources.DesiredVcpus = pulumi.IntPtr(int(cr.DesiredVcpus))
		}
		if len(cr.InstanceTypes) > 0 {
			computeResources.InstanceTypes = pulumi.ToStringArray(cr.InstanceTypes)
		}
		if cr.AllocationStrategy != "" {
			computeResources.AllocationStrategy = pulumi.StringPtr(cr.AllocationStrategy)
		}
		if cr.InstanceRole != nil && cr.InstanceRole.GetValue() != "" {
			computeResources.InstanceRole = pulumi.StringPtr(cr.InstanceRole.GetValue())
		}
		if cr.Ec2KeyPair != "" {
			computeResources.Ec2KeyPair = pulumi.StringPtr(cr.Ec2KeyPair)
		}
		if cr.ResourceTags != nil && len(cr.ResourceTags) > 0 {
			computeResources.Tags = pulumi.ToStringMap(cr.ResourceTags)
		}

		// Launch template
		if cr.LaunchTemplate != nil {
			lt := &batch.ComputeEnvironmentComputeResourcesLaunchTemplateArgs{}
			if cr.LaunchTemplate.LaunchTemplateId != "" {
				lt.LaunchTemplateId = pulumi.StringPtr(cr.LaunchTemplate.LaunchTemplateId)
			}
			if cr.LaunchTemplate.LaunchTemplateName != "" {
				lt.LaunchTemplateName = pulumi.StringPtr(cr.LaunchTemplate.LaunchTemplateName)
			}
			if cr.LaunchTemplate.Version != "" {
				lt.Version = pulumi.StringPtr(cr.LaunchTemplate.Version)
			}
			computeResources.LaunchTemplate = lt
		}

		// EC2 configurations
		if len(cr.Ec2Configurations) > 0 {
			var ec2Configs batch.ComputeEnvironmentComputeResourcesEc2ConfigurationArray
			for _, ec2Cfg := range cr.Ec2Configurations {
				cfg := &batch.ComputeEnvironmentComputeResourcesEc2ConfigurationArgs{}
				if ec2Cfg.ImageType != "" {
					cfg.ImageType = pulumi.StringPtr(ec2Cfg.ImageType)
				}
				if ec2Cfg.ImageIdOverride != "" {
					cfg.ImageIdOverride = pulumi.StringPtr(ec2Cfg.ImageIdOverride)
				}
				ec2Configs = append(ec2Configs, cfg)
			}
			computeResources.Ec2Configurations = ec2Configs
		}
	}

	// SPOT-specific fields
	if cr.Type == "SPOT" {
		if cr.BidPercentage != nil {
			computeResources.BidPercentage = pulumi.IntPtr(int(cr.GetBidPercentage()))
		}
		if cr.SpotIamFleetRole != nil && cr.SpotIamFleetRole.GetValue() != "" {
			computeResources.SpotIamFleetRole = pulumi.StringPtr(cr.SpotIamFleetRole.GetValue())
		}
	}

	// Build compute environment args
	args := &batch.ComputeEnvironmentArgs{
		Name:             pulumi.StringPtr(locals.AwsBatchComputeEnvironment.Metadata.Id),
		Type:             pulumi.String("MANAGED"),
		State:            pulumi.StringPtr(spec.GetState()),
		ComputeResources: computeResources,
		Tags:             pulumi.ToStringMap(locals.Labels),
	}

	// Service role (optional -- omit to use service-linked role)
	if spec.ServiceRole != nil && spec.ServiceRole.GetValue() != "" {
		args.ServiceRole = pulumi.StringPtr(spec.ServiceRole.GetValue())
	}

	// Update policy
	if spec.UpdatePolicy != nil {
		up := &batch.ComputeEnvironmentUpdatePolicyArgs{
			TerminateJobsOnUpdate: pulumi.Bool(spec.UpdatePolicy.TerminateJobsOnUpdate),
		}
		if spec.UpdatePolicy.JobExecutionTimeoutMinutes != nil {
			up.JobExecutionTimeoutMinutes = pulumi.Int(int(spec.UpdatePolicy.GetJobExecutionTimeoutMinutes()))
		}
		args.UpdatePolicy = up
	}

	ce, err := batch.NewComputeEnvironment(ctx, "batch-compute-environment", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create batch compute environment")
	}

	return ce, nil
}
