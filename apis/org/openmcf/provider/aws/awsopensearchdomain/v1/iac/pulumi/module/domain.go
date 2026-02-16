package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/opensearch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// domain creates the OpenSearch Service domain and exports outputs.
func domain(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	// -------------------------------------------------------------------
	// Build domain args
	// -------------------------------------------------------------------

	args := &opensearch.DomainArgs{
		DomainName:    pulumi.String(locals.Target.Metadata.Id),
		EngineVersion: pulumi.String(spec.EngineVersion),
		Tags:          pulumi.ToStringMap(locals.AwsTags),
	}

	// -------------------------------------------------------------------
	// Cluster configuration
	// -------------------------------------------------------------------

	if cc := spec.ClusterConfig; cc != nil {
		clusterCfg := &opensearch.DomainClusterConfigArgs{
			InstanceType: pulumi.String(cc.InstanceType),
		}

		// Instance count (optional proto field — uses *int32)
		if cc.InstanceCount != nil {
			clusterCfg.InstanceCount = pulumi.Int(int(*cc.InstanceCount))
		}

		// Dedicated master nodes
		if cc.DedicatedMasterEnabled {
			clusterCfg.DedicatedMasterEnabled = pulumi.Bool(true)
			if cc.DedicatedMasterType != "" {
				clusterCfg.DedicatedMasterType = pulumi.String(cc.DedicatedMasterType)
			}
			if cc.DedicatedMasterCount > 0 {
				clusterCfg.DedicatedMasterCount = pulumi.Int(int(cc.DedicatedMasterCount))
			}
		}

		// Zone awareness
		if cc.ZoneAwarenessEnabled {
			clusterCfg.ZoneAwarenessEnabled = pulumi.Bool(true)
			if cc.AvailabilityZoneCount > 0 {
				clusterCfg.ZoneAwarenessConfig = &opensearch.DomainClusterConfigZoneAwarenessConfigArgs{
					AvailabilityZoneCount: pulumi.Int(int(cc.AvailabilityZoneCount)),
				}
			}
		}

		// UltraWarm storage tier
		if cc.WarmEnabled {
			clusterCfg.WarmEnabled = pulumi.Bool(true)
			if cc.WarmType != "" {
				clusterCfg.WarmType = pulumi.String(cc.WarmType)
			}
			if cc.WarmCount > 0 {
				clusterCfg.WarmCount = pulumi.Int(int(cc.WarmCount))
			}
		}

		// Cold storage (requires UltraWarm)
		if cc.ColdStorageEnabled {
			clusterCfg.ColdStorageOptions = &opensearch.DomainClusterConfigColdStorageOptionsArgs{
				Enabled: pulumi.Bool(true),
			}
		}

		// Multi-AZ with Standby
		if cc.MultiAzWithStandbyEnabled {
			clusterCfg.MultiAzWithStandbyEnabled = pulumi.Bool(true)
		}

		args.ClusterConfig = clusterCfg
	}

	// -------------------------------------------------------------------
	// EBS options
	// -------------------------------------------------------------------

	if ebs := spec.EbsOptions; ebs != nil {
		ebsOpts := &opensearch.DomainEbsOptionsArgs{
			EbsEnabled: pulumi.Bool(ebs.EbsEnabled),
		}
		if ebs.VolumeType != "" {
			ebsOpts.VolumeType = pulumi.String(ebs.VolumeType)
		}
		if ebs.VolumeSize > 0 {
			ebsOpts.VolumeSize = pulumi.Int(int(ebs.VolumeSize))
		}
		if ebs.Iops > 0 {
			ebsOpts.Iops = pulumi.Int(int(ebs.Iops))
		}
		if ebs.Throughput > 0 {
			ebsOpts.Throughput = pulumi.Int(int(ebs.Throughput))
		}
		args.EbsOptions = ebsOpts
	}

	// -------------------------------------------------------------------
	// Encryption at rest
	// -------------------------------------------------------------------

	if spec.EncryptAtRestEnabled {
		encryptCfg := &opensearch.DomainEncryptAtRestArgs{
			Enabled: pulumi.Bool(true),
		}
		if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
			encryptCfg.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
		}
		args.EncryptAtRest = encryptCfg
	}

	// -------------------------------------------------------------------
	// Node-to-node encryption
	// -------------------------------------------------------------------

	if spec.NodeToNodeEncryptionEnabled {
		args.NodeToNodeEncryption = &opensearch.DomainNodeToNodeEncryptionArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	// -------------------------------------------------------------------
	// VPC options (ForceNew — choose carefully)
	// -------------------------------------------------------------------

	if vpc := spec.VpcOptions; vpc != nil {
		vpcOpts := &opensearch.DomainVpcOptionsArgs{}

		var subnetIds pulumi.StringArray
		for _, s := range vpc.SubnetIds {
			if s.GetValue() != "" {
				subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
			}
		}
		if len(subnetIds) > 0 {
			vpcOpts.SubnetIds = subnetIds
		}

		var sgIds pulumi.StringArray
		for _, sg := range vpc.SecurityGroupIds {
			if sg.GetValue() != "" {
				sgIds = append(sgIds, pulumi.String(sg.GetValue()))
			}
		}
		if len(sgIds) > 0 {
			vpcOpts.SecurityGroupIds = sgIds
		}

		args.VpcOptions = vpcOpts
	}

	// -------------------------------------------------------------------
	// Domain endpoint options (HTTPS, TLS, custom endpoint)
	// -------------------------------------------------------------------

	if deo := spec.DomainEndpointOptions; deo != nil {
		endpointOpts := &opensearch.DomainDomainEndpointOptionsArgs{}

		if deo.EnforceHttps != nil {
			endpointOpts.EnforceHttps = pulumi.Bool(*deo.EnforceHttps)
		}
		if deo.TlsSecurityPolicy != "" {
			endpointOpts.TlsSecurityPolicy = pulumi.String(deo.TlsSecurityPolicy)
		}
		if deo.CustomEndpointEnabled {
			endpointOpts.CustomEndpointEnabled = pulumi.Bool(true)
			if deo.CustomEndpoint != "" {
				endpointOpts.CustomEndpoint = pulumi.String(deo.CustomEndpoint)
			}
			if deo.CustomEndpointCertificateArn != nil && deo.CustomEndpointCertificateArn.GetValue() != "" {
				endpointOpts.CustomEndpointCertificateArn = pulumi.String(deo.CustomEndpointCertificateArn.GetValue())
			}
		}

		args.DomainEndpointOptions = endpointOpts
	}

	// -------------------------------------------------------------------
	// Advanced security options (FGAC — ForceNew if disabling)
	// -------------------------------------------------------------------

	if aso := spec.AdvancedSecurityOptions; aso != nil && aso.Enabled {
		secOpts := &opensearch.DomainAdvancedSecurityOptionsArgs{
			Enabled:                     pulumi.Bool(true),
			InternalUserDatabaseEnabled: pulumi.Bool(aso.InternalUserDatabaseEnabled),
		}

		masterUserOpts := &opensearch.DomainAdvancedSecurityOptionsMasterUserOptionsArgs{}
		hasMasterUser := false

		if aso.MasterUserArn != nil && aso.MasterUserArn.GetValue() != "" {
			masterUserOpts.MasterUserArn = pulumi.String(aso.MasterUserArn.GetValue())
			hasMasterUser = true
		}
		if aso.MasterUserName != "" {
			masterUserOpts.MasterUserName = pulumi.String(aso.MasterUserName)
			hasMasterUser = true
		}
		if aso.MasterUserPassword != nil && aso.MasterUserPassword.GetValue() != "" {
			masterUserOpts.MasterUserPassword = pulumi.String(aso.MasterUserPassword.GetValue())
			hasMasterUser = true
		}

		if hasMasterUser {
			secOpts.MasterUserOptions = masterUserOpts
		}

		args.AdvancedSecurityOptions = secOpts
	}

	// -------------------------------------------------------------------
	// Log publishing options
	// -------------------------------------------------------------------

	if len(spec.LogPublishingOptions) > 0 {
		var logOpts opensearch.DomainLogPublishingOptionArray
		for _, lpo := range spec.LogPublishingOptions {
			entry := &opensearch.DomainLogPublishingOptionArgs{
				LogType: pulumi.String(lpo.LogType),
			}
			if lpo.CloudwatchLogGroupArn != nil && lpo.CloudwatchLogGroupArn.GetValue() != "" {
				entry.CloudwatchLogGroupArn = pulumi.String(lpo.CloudwatchLogGroupArn.GetValue())
			}
			if lpo.Enabled != nil {
				entry.Enabled = pulumi.Bool(*lpo.Enabled)
			}
			logOpts = append(logOpts, entry)
		}
		args.LogPublishingOptions = logOpts
	}

	// -------------------------------------------------------------------
	// Access policies (IAM policy document as JSON)
	// -------------------------------------------------------------------

	if spec.AccessPolicies != nil {
		policyJSON, err := json.Marshal(spec.AccessPolicies.AsMap())
		if err != nil {
			return errors.Wrap(err, "failed to serialize access_policies to JSON")
		}
		args.AccessPolicies = pulumi.String(string(policyJSON))
	}

	// -------------------------------------------------------------------
	// Auto-Tune
	// -------------------------------------------------------------------

	if spec.AutoTuneEnabled {
		args.AutoTuneOptions = &opensearch.DomainAutoTuneOptionsArgs{
			DesiredState: pulumi.String("ENABLED"),
		}
	}

	// -------------------------------------------------------------------
	// Software update options
	// -------------------------------------------------------------------

	if spec.AutoSoftwareUpdateEnabled {
		args.SoftwareUpdateOptions = &opensearch.DomainSoftwareUpdateOptionsArgs{
			AutoSoftwareUpdateEnabled: pulumi.Bool(true),
		}
	}

	// -------------------------------------------------------------------
	// IP address type
	// -------------------------------------------------------------------

	if spec.IpAddressType != "" {
		args.IpAddressType = pulumi.String(spec.IpAddressType)
	}

	// -------------------------------------------------------------------
	// Advanced options (low-level key-value configuration)
	// -------------------------------------------------------------------

	if len(spec.AdvancedOptions) > 0 {
		args.AdvancedOptions = pulumi.ToStringMap(spec.AdvancedOptions)
	}

	// -------------------------------------------------------------------
	// Create the OpenSearch domain
	// -------------------------------------------------------------------

	osDomain, err := opensearch.NewDomain(ctx, "opensearch-domain", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create opensearch domain")
	}

	// -------------------------------------------------------------------
	// Export outputs
	// -------------------------------------------------------------------

	ctx.Export(OpDomainId, osDomain.DomainId)
	ctx.Export(OpDomainName, osDomain.DomainName)
	ctx.Export(OpDomainArn, osDomain.Arn)
	ctx.Export(OpEndpoint, osDomain.Endpoint)
	ctx.Export(OpDashboardEndpoint, pulumi.Sprintf("%s/_dashboards", osDomain.Endpoint))

	return nil
}
