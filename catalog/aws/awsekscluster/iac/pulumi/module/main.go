package module

import (
	"github.com/pkg/errors"
	awseksclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsekscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the EKS cluster CONTROL PLANE only. Compute attaches
// as separate composable nodes (AwsEksNodeGroup, or EKS Auto Mode enabled
// here), and the cluster role is a referenced AwsIamRole that carries its
// own AmazonEKSClusterPolicy -- this module never modifies a role it merely
// references.
func Resources(ctx *pulumi.Context, iacInput *awseksclusterv1alpha1.AwsEksClusterIacInput) (err error) {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEksCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	target := locals.AwsEksCluster
	spec := target.Spec

	subnetIds := make(pulumi.StringArray, 0, len(spec.SubnetIds))
	for _, subnet := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(subnet.GetValue()))
	}

	vpcConfig := &eks.ClusterVpcConfigArgs{
		SubnetIds: subnetIds,
	}

	// Endpoint exposure is a pair of independent toggles at AWS (public
	// defaults true, private defaults false). endpoint_public_access is
	// proto-optional so an explicit false ("private-only cluster") is
	// distinguishable from unset (keep the AWS default of true).
	if spec.EndpointPublicAccess != nil {
		vpcConfig.EndpointPublicAccess = pulumi.BoolPtr(*spec.EndpointPublicAccess)
	}
	if spec.EndpointPrivateAccess {
		vpcConfig.EndpointPrivateAccess = pulumi.BoolPtr(true)
	}
	if len(spec.PublicAccessCidrs) > 0 {
		vpcConfig.PublicAccessCidrs = pulumi.ToStringArray(spec.PublicAccessCidrs)
	}
	if len(spec.SecurityGroupIds) > 0 {
		securityGroupIds := make(pulumi.StringArray, 0, len(spec.SecurityGroupIds))
		for _, securityGroup := range spec.SecurityGroupIds {
			securityGroupIds = append(securityGroupIds, pulumi.String(securityGroup.GetValue()))
		}
		vpcConfig.SecurityGroupIds = securityGroupIds
	}
	// Sent only when set so AWS's default (AWS_MANAGED) keeps applying.
	// Reverting CUSTOMER_ROUTED to AWS_MANAGED forces cluster replacement.
	if spec.ControlPlaneEgressMode != "" {
		vpcConfig.ControlPlaneEgressMode = pulumi.StringPtr(spec.ControlPlaneEgressMode)
	}

	clusterArgs := &eks.ClusterArgs{
		Name:      pulumi.String(target.Metadata.Name),
		RoleArn:   pulumi.String(spec.ClusterRoleArn.GetValue()),
		VpcConfig: vpcConfig,
		Tags:      pulumi.ToStringMap(locals.AwsTags),
	}

	// An empty version lets AWS pick its current default; EKS upgrades one
	// minor at a time and never downgrades.
	if spec.Version != "" {
		clusterArgs.Version = pulumi.StringPtr(spec.Version)
	}
	if spec.ForceUpdateVersion {
		clusterArgs.ForceUpdateVersion = pulumi.BoolPtr(true)
	}
	if len(spec.EnabledClusterLogTypes) > 0 {
		clusterArgs.EnabledClusterLogTypes = pulumi.ToStringArray(spec.EnabledClusterLogTypes)
	}

	// Envelope encryption is a one-way door: AWS only ever associates a key
	// (never dissociates or re-keys in place), so the block is sent only
	// when a key is configured.
	if spec.KmsKeyArn.GetValue() != "" {
		clusterArgs.EncryptionConfig = &eks.ClusterEncryptionConfigArgs{
			Provider: &eks.ClusterEncryptionConfigProviderArgs{
				KeyArn: pulumi.String(spec.KmsKeyArn.GetValue()),
			},
			// "secrets" is the only resource type the EKS API accepts here,
			// which is why the spec folds it away.
			Resources: pulumi.StringArray{pulumi.String("secrets")},
		}
	}

	if spec.AccessConfig != nil {
		accessConfig := &eks.ClusterAccessConfigArgs{}
		if spec.AccessConfig.AuthenticationMode != "" {
			accessConfig.AuthenticationMode = pulumi.StringPtr(spec.AccessConfig.AuthenticationMode)
		}
		// Proto-optional: explicit false ("no creator admin") must be
		// distinguishable from unset (AWS defaults to true). Create-only.
		if spec.AccessConfig.BootstrapClusterCreatorAdminPermissions != nil {
			accessConfig.BootstrapClusterCreatorAdminPermissions = pulumi.BoolPtr(*spec.AccessConfig.BootstrapClusterCreatorAdminPermissions)
		}
		clusterArgs.AccessConfig = accessConfig
	}

	// kubernetes_network_config carries both the immutable address-family
	// settings and the Auto Mode load-balancing capability, so it is built
	// from either source and sent once.
	networkConfig := &eks.ClusterKubernetesNetworkConfigArgs{}
	networkConfigNeeded := false
	if spec.IpFamily != "" {
		networkConfig.IpFamily = pulumi.StringPtr(spec.IpFamily)
		networkConfigNeeded = true
	}
	if spec.ServiceIpv4Cidr != "" {
		networkConfig.ServiceIpv4Cidr = pulumi.StringPtr(spec.ServiceIpv4Cidr)
		networkConfigNeeded = true
	}

	// EKS Auto Mode is all-or-nothing across three AWS blocks (compute,
	// block storage, elastic load balancing) -- the API requires them to be
	// enabled or disabled together, which is why the spec models one
	// toggle. Expanded here into the three blocks AWS expects.
	if spec.AutoMode != nil && spec.AutoMode.Enabled {
		computeConfig := &eks.ClusterComputeConfigArgs{
			Enabled: pulumi.BoolPtr(true),
		}
		if len(spec.AutoMode.NodePools) > 0 {
			computeConfig.NodePools = pulumi.ToStringArray(spec.AutoMode.NodePools)
		}
		if spec.AutoMode.NodeRoleArn.GetValue() != "" {
			computeConfig.NodeRoleArn = pulumi.StringPtr(spec.AutoMode.NodeRoleArn.GetValue())
		}
		clusterArgs.ComputeConfig = computeConfig
		clusterArgs.StorageConfig = &eks.ClusterStorageConfigArgs{
			BlockStorage: &eks.ClusterStorageConfigBlockStorageArgs{
				Enabled: pulumi.BoolPtr(true),
			},
		}
		networkConfig.ElasticLoadBalancing = &eks.ClusterKubernetesNetworkConfigElasticLoadBalancingArgs{
			Enabled: pulumi.BoolPtr(true),
		}
		networkConfigNeeded = true
	}
	if networkConfigNeeded {
		clusterArgs.KubernetesNetworkConfig = networkConfig
	}

	if spec.UpgradeSupportType != "" {
		clusterArgs.UpgradePolicy = &eks.ClusterUpgradePolicyArgs{
			SupportType: pulumi.StringPtr(spec.UpgradeSupportType),
		}
	}
	if spec.ZonalShiftEnabled {
		clusterArgs.ZonalShiftConfig = &eks.ClusterZonalShiftConfigArgs{
			Enabled: pulumi.BoolPtr(true),
		}
	}
	// Provisioned control-plane capacity (billed hourly on top of the
	// cluster fee). Sent only when set so AWS's default (standard) keeps
	// applying; tier changes update in place.
	if spec.ControlPlaneScalingTier != "" {
		clusterArgs.ControlPlaneScalingConfig = &eks.ClusterControlPlaneScalingConfigArgs{
			Tier: pulumi.StringPtr(spec.ControlPlaneScalingTier),
		}
	}
	// EKS Hybrid Nodes: the on-premises node/pod CIDR ranges allowed to
	// join. Updatable in place; each nested block is sent only when its
	// list is non-empty (spec validation guarantees at least one list
	// carries a range).
	if spec.RemoteNetworks != nil {
		remoteNetworkConfig := &eks.ClusterRemoteNetworkConfigArgs{}
		if len(spec.RemoteNetworks.NodeCidrs) > 0 {
			remoteNetworkConfig.RemoteNodeNetworks = &eks.ClusterRemoteNetworkConfigRemoteNodeNetworksArgs{
				Cidrs: pulumi.ToStringArray(spec.RemoteNetworks.NodeCidrs),
			}
		}
		if len(spec.RemoteNetworks.PodCidrs) > 0 {
			remoteNetworkConfig.RemotePodNetworks = &eks.ClusterRemoteNetworkConfigRemotePodNetworksArgs{
				Cidrs: pulumi.ToStringArray(spec.RemoteNetworks.PodCidrs),
			}
		}
		clusterArgs.RemoteNetworkConfig = remoteNetworkConfig
	}
	// Always send the explicit boolean: the provider attribute is
	// Optional+Computed, so leaving it unset means "keep the cluster's
	// current value" -- omitting false would make protection impossible
	// to turn off once enabled (and destroys would stay blocked forever).
	clusterArgs.DeletionProtection = pulumi.BoolPtr(spec.DeletionProtection)
	// Proto-optional: explicit false ("bring your own add-ons") must be
	// distinguishable from unset (AWS defaults to true). Create-only.
	if spec.BootstrapSelfManagedAddons != nil {
		clusterArgs.BootstrapSelfManagedAddons = pulumi.BoolPtr(*spec.BootstrapSelfManagedAddons)
	}

	createdCluster, err := eks.NewCluster(ctx, target.Metadata.Name, clusterArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create EKS cluster")
	}

	// The OIDC issuer is the trust anchor for IRSA: an AwsIamOidcProvider
	// pointed at this output turns Kubernetes service accounts into IAM
	// principals.
	oidcIssuerUrl := createdCluster.Identities.Index(pulumi.Int(0)).Oidcs().Index(pulumi.Int(0)).Issuer()

	ctx.Export(OpEndpoint, createdCluster.Endpoint)
	ctx.Export(OpClusterCaCertificate, createdCluster.CertificateAuthority.Data().Elem())
	ctx.Export(OpClusterSecurityGroupId, createdCluster.VpcConfig.ClusterSecurityGroupId().Elem())
	ctx.Export(OpOidcIssuerUrl, oidcIssuerUrl)
	ctx.Export(OpClusterArn, createdCluster.Arn)
	ctx.Export(OpName, createdCluster.Name)
	ctx.Export(OpPlatformVersion, createdCluster.PlatformVersion)

	return nil
}
