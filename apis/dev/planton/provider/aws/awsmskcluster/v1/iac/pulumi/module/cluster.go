package module

import (
	"github.com/pkg/errors"
	awsmskclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsmskcluster/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/msk"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSg *ec2.SecurityGroup,
	createdConfig *msk.Configuration,
) (*msk.Cluster, error) {
	spec := locals.AwsMskCluster.Spec

	// Resolve subnets
	var subnets pulumi.StringArray
	for _, s := range spec.SubnetIds {
		subnets = append(subnets, pulumi.String(s.GetValue()))
	}

	// Resolve security groups: combine associate SGs + managed SG
	var securityGroups pulumi.StringArray
	for _, sg := range spec.AssociateSecurityGroupIds {
		securityGroups = append(securityGroups, pulumi.String(sg.GetValue()))
	}
	if createdSg != nil {
		securityGroups = append(securityGroups, createdSg.ID())
	}

	// Storage info
	var storageInfo msk.ClusterBrokerNodeGroupInfoStorageInfoPtrInput
	if spec.EbsVolumeSizeGib != nil || spec.ProvisionedThroughputEnabled {
		ebsArgs := &msk.ClusterBrokerNodeGroupInfoStorageInfoEbsStorageInfoArgs{}
		if spec.EbsVolumeSizeGib != nil {
			ebsArgs.VolumeSize = pulumi.Int(int(spec.GetEbsVolumeSizeGib()))
		}
		if spec.ProvisionedThroughputEnabled {
			ebsArgs.ProvisionedThroughput = &msk.ClusterBrokerNodeGroupInfoStorageInfoEbsStorageInfoProvisionedThroughputArgs{
				Enabled:          pulumi.Bool(true),
				VolumeThroughput: pulumi.Int(int(spec.ProvisionedThroughputMbs)),
			}
		}
		storageInfo = &msk.ClusterBrokerNodeGroupInfoStorageInfoArgs{
			EbsStorageInfo: ebsArgs,
		}
	}

	// Connectivity info (public access)
	var connectivityInfo msk.ClusterBrokerNodeGroupInfoConnectivityInfoPtrInput
	if spec.PublicAccessType != "" {
		connectivityInfo = &msk.ClusterBrokerNodeGroupInfoConnectivityInfoArgs{
			PublicAccess: &msk.ClusterBrokerNodeGroupInfoConnectivityInfoPublicAccessArgs{
				Type: pulumi.String(spec.PublicAccessType),
			},
		}
	}

	args := &msk.ClusterArgs{
		ClusterName:         pulumi.String(locals.AwsMskCluster.Metadata.Id),
		KafkaVersion:        pulumi.String(spec.KafkaVersion),
		NumberOfBrokerNodes: pulumi.Int(int(spec.NumberOfBrokerNodes)),
		BrokerNodeGroupInfo: &msk.ClusterBrokerNodeGroupInfoArgs{
			InstanceType:     pulumi.String(spec.InstanceType),
			ClientSubnets:    subnets,
			SecurityGroups:   securityGroups,
			StorageInfo:      storageInfo,
			ConnectivityInfo: connectivityInfo,
		},
		Tags: pulumi.ToStringMap(locals.Labels),
	}

	// Storage mode
	if spec.StorageMode != "" {
		args.StorageMode = pulumi.String(spec.StorageMode)
	}

	// Encryption
	encryptionInTransit := &msk.ClusterEncryptionInfoEncryptionInTransitArgs{}
	hasEncryptionInTransit := false

	if spec.ClientBrokerEncryption != nil && spec.GetClientBrokerEncryption() != "" {
		encryptionInTransit.ClientBroker = pulumi.String(spec.GetClientBrokerEncryption())
		hasEncryptionInTransit = true
	}
	if spec.InClusterEncryption != nil {
		encryptionInTransit.InCluster = pulumi.Bool(spec.GetInClusterEncryption())
		hasEncryptionInTransit = true
	}

	encryptionArgs := &msk.ClusterEncryptionInfoArgs{}
	hasEncryption := false

	if spec.KmsKeyArn != nil && spec.KmsKeyArn.GetValue() != "" {
		encryptionArgs.EncryptionAtRestKmsKeyArn = pulumi.String(spec.KmsKeyArn.GetValue())
		hasEncryption = true
	}
	if hasEncryptionInTransit {
		encryptionArgs.EncryptionInTransit = encryptionInTransit
		hasEncryption = true
	}
	if hasEncryption {
		args.EncryptionInfo = encryptionArgs
	}

	// Authentication
	if spec.Authentication != nil {
		auth := spec.Authentication
		authArgs := &msk.ClusterClientAuthenticationArgs{}
		hasAuth := false

		if auth.SaslIamEnabled || auth.SaslScramEnabled {
			authArgs.Sasl = &msk.ClusterClientAuthenticationSaslArgs{
				Iam:   pulumi.Bool(auth.SaslIamEnabled),
				Scram: pulumi.Bool(auth.SaslScramEnabled),
			}
			hasAuth = true
		}

		if auth.TlsEnabled {
			tlsArgs := &msk.ClusterClientAuthenticationTlsArgs{}
			if len(auth.TlsCertificateAuthorityArns) > 0 {
				var caArns pulumi.StringArray
				for _, ca := range auth.TlsCertificateAuthorityArns {
					caArns = append(caArns, pulumi.String(ca.GetValue()))
				}
				tlsArgs.CertificateAuthorityArns = caArns
			}
			authArgs.Tls = tlsArgs
			hasAuth = true
		}

		if auth.Unauthenticated {
			authArgs.Unauthenticated = pulumi.Bool(true)
			hasAuth = true
		}

		if hasAuth {
			args.ClientAuthentication = authArgs
		}
	}

	// Configuration
	if createdConfig != nil {
		args.ConfigurationInfo = &msk.ClusterConfigurationInfoArgs{
			Arn:      createdConfig.Arn,
			Revision: createdConfig.LatestRevision,
		}
	} else if spec.ConfigurationArn != "" {
		args.ConfigurationInfo = &msk.ClusterConfigurationInfoArgs{
			Arn:      pulumi.String(spec.ConfigurationArn),
			Revision: pulumi.Int(int(spec.ConfigurationRevision)),
		}
	}

	// Logging
	if spec.Logging != nil {
		loggingArgs := buildLogging(spec)
		if loggingArgs != nil {
			args.LoggingInfo = loggingArgs
		}
	}

	// Monitoring
	if spec.EnhancedMonitoring != "" {
		args.EnhancedMonitoring = pulumi.String(spec.EnhancedMonitoring)
	}

	if spec.JmxExporterEnabled || spec.NodeExporterEnabled {
		args.OpenMonitoring = &msk.ClusterOpenMonitoringArgs{
			Prometheus: &msk.ClusterOpenMonitoringPrometheusArgs{
				JmxExporter: &msk.ClusterOpenMonitoringPrometheusJmxExporterArgs{
					EnabledInBroker: pulumi.Bool(spec.JmxExporterEnabled),
				},
				NodeExporter: &msk.ClusterOpenMonitoringPrometheusNodeExporterArgs{
					EnabledInBroker: pulumi.Bool(spec.NodeExporterEnabled),
				},
			},
		}
	}

	mskCluster, err := msk.NewCluster(ctx, "msk-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create msk cluster")
	}

	return mskCluster, nil
}

// buildLogging constructs the logging configuration from the spec.
func buildLogging(spec *awsmskclusterv1.AwsMskClusterSpec) *msk.ClusterLoggingInfoArgs {
	logging := spec.Logging
	if logging == nil {
		return nil
	}

	brokerLogsArgs := &msk.ClusterLoggingInfoBrokerLogsArgs{}
	hasBrokerLogs := false

	if logging.CloudwatchLogs != nil {
		cwArgs := &msk.ClusterLoggingInfoBrokerLogsCloudwatchLogsArgs{
			Enabled: pulumi.Bool(logging.CloudwatchLogs.Enabled),
		}
		if logging.CloudwatchLogs.LogGroup != nil && logging.CloudwatchLogs.LogGroup.GetValue() != "" {
			cwArgs.LogGroup = pulumi.String(logging.CloudwatchLogs.LogGroup.GetValue())
		}
		brokerLogsArgs.CloudwatchLogs = cwArgs
		hasBrokerLogs = true
	}

	if logging.Firehose != nil {
		fhArgs := &msk.ClusterLoggingInfoBrokerLogsFirehoseArgs{
			Enabled: pulumi.Bool(logging.Firehose.Enabled),
		}
		if logging.Firehose.DeliveryStream != nil && logging.Firehose.DeliveryStream.GetValue() != "" {
			fhArgs.DeliveryStream = pulumi.String(logging.Firehose.DeliveryStream.GetValue())
		}
		brokerLogsArgs.Firehose = fhArgs
		hasBrokerLogs = true
	}

	if logging.S3 != nil {
		s3Args := &msk.ClusterLoggingInfoBrokerLogsS3Args{
			Enabled: pulumi.Bool(logging.S3.Enabled),
		}
		if logging.S3.Bucket != nil && logging.S3.Bucket.GetValue() != "" {
			s3Args.Bucket = pulumi.String(logging.S3.Bucket.GetValue())
		}
		if logging.S3.Prefix != "" {
			s3Args.Prefix = pulumi.String(logging.S3.Prefix)
		}
		brokerLogsArgs.S3 = s3Args
		hasBrokerLogs = true
	}

	if !hasBrokerLogs {
		return nil
	}

	return &msk.ClusterLoggingInfoArgs{
		BrokerLogs: brokerLogsArgs,
	}
}
