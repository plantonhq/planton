package module

import (
	"github.com/pkg/errors"
	alicloudkubernetesclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudkubernetescluster/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudkubernetesclusterv1.AlicloudKubernetesClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudKubernetesCluster.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	clusterName := spec.Name
	if clusterName == "" {
		clusterName = locals.AlicloudKubernetesCluster.Metadata.Name
	}

	vswitchIds := pulumi.StringArray{}
	for _, ref := range spec.VswitchIds {
		vswitchIds = append(vswitchIds, pulumi.String(ref.GetValue()))
	}

	args := &cs.ManagedKubernetesArgs{
		Name:               pulumi.String(clusterName),
		ClusterSpec:        pulumi.String(clusterSpec(spec)),
		ProxyMode:          pulumi.String(proxyMode(spec)),
		NodeCidrMask:       pulumi.Int(nodeCidrMask(spec)),
		NewNatGateway:      pulumi.Bool(newNatGateway(spec)),
		SlbInternetEnabled: pulumi.Bool(slbInternetEnabled(spec)),
		EnableRrsa:         pulumi.Bool(enableRrsa(spec)),
		DeletionProtection: pulumi.Bool(deletionProtection(spec)),
		VswitchIds:         vswitchIds,
		ServiceCidr:        pulumi.String(spec.ServiceCidr),
		Tags:               pulumi.ToStringMap(locals.Tags),
	}

	if spec.Version != "" {
		args.Version = pulumi.String(spec.Version)
	}

	if spec.ClusterDomain != "" {
		args.ClusterDomain = pulumi.String(spec.ClusterDomain)
	}

	if spec.PodCidr != "" {
		args.PodCidr = pulumi.String(spec.PodCidr)
	}

	if len(spec.PodVswitchIds) > 0 {
		podVswitchIds := pulumi.StringArray{}
		for _, ref := range spec.PodVswitchIds {
			podVswitchIds = append(podVswitchIds, pulumi.String(ref.GetValue()))
		}
		args.PodVswitchIds = podVswitchIds
	}

	if spec.SecurityGroupId != nil {
		args.SecurityGroupId = pulumi.String(spec.SecurityGroupId.GetValue())
	}

	if isEnterpriseSg(spec) {
		args.IsEnterpriseSecurityGroup = pulumi.Bool(true)
	}

	if spec.EncryptionProviderKey != nil {
		args.EncryptionProviderKey = pulumi.String(spec.EncryptionProviderKey.GetValue())
	}

	if spec.CustomSan != "" {
		args.CustomSan = pulumi.String(spec.CustomSan)
	}

	if spec.ResourceGroupId != "" {
		args.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	if spec.Timezone != "" {
		args.Timezone = pulumi.String(spec.Timezone)
	}

	if len(spec.Addons) > 0 {
		addons := cs.ManagedKubernetesAddonArray{}
		for _, addon := range spec.Addons {
			addonArgs := cs.ManagedKubernetesAddonArgs{
				Name: pulumi.String(addon.Name),
			}
			if addon.Config != "" {
				addonArgs.Config = pulumi.String(addon.Config)
			}
			if addon.Version != "" {
				addonArgs.Version = pulumi.String(addon.Version)
			}
			if addon.Disabled {
				addonArgs.Disabled = pulumi.Bool(true)
			}
			addons = append(addons, addonArgs)
		}
		args.Addons = addons
	}

	if spec.Logging != nil {
		configureLogging(args, spec.Logging)
	}

	if spec.MaintenanceWindow != nil {
		args.MaintenanceWindow = cs.ManagedKubernetesMaintenanceWindowArgs{
			Enable:          pulumi.Bool(spec.MaintenanceWindow.Enable),
			MaintenanceTime: pulumi.String(spec.MaintenanceWindow.MaintenanceTime),
			Duration:        pulumi.String(spec.MaintenanceWindow.Duration),
			WeeklyPeriod:    pulumi.String(spec.MaintenanceWindow.WeeklyPeriod),
		}
	}

	if spec.AutoUpgrade != nil && spec.AutoUpgrade.Enabled {
		args.OperationPolicy = cs.ManagedKubernetesOperationPolicyArgs{
			ClusterAutoUpgrade: cs.ManagedKubernetesOperationPolicyClusterAutoUpgradeArgs{
				Enabled: pulumi.Bool(true),
				Channel: pulumi.String(autoUpgradeChannel(spec.AutoUpgrade)),
			},
		}
	}

	cluster, err := cs.NewManagedKubernetes(ctx, clusterName, args, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create ACK managed cluster %s", clusterName)
	}

	exportOutputs(ctx, cluster)

	return nil
}

func configureLogging(args *cs.ManagedKubernetesArgs, logging *alicloudkubernetesclusterv1.AlicloudKubernetesClusterLogging) {
	if logging.ControlPlaneLogProject != nil {
		args.ControlPlaneLogProject = pulumi.String(logging.ControlPlaneLogProject.GetValue())
	}

	args.ControlPlaneLogTtl = pulumi.String(controlPlaneLogTtl(logging))

	if len(logging.ControlPlaneLogComponents) > 0 {
		args.ControlPlaneLogComponents = pulumi.ToStringArray(logging.ControlPlaneLogComponents)
	}

	if logging.AuditLogEnabled {
		auditConfig := cs.ManagedKubernetesAuditLogConfigArgs{
			Enabled: pulumi.Bool(true),
		}
		if logging.AuditLogSlsProject != "" {
			auditConfig.SlsProjectName = pulumi.String(logging.AuditLogSlsProject)
		}
		args.AuditLogConfig = auditConfig
	}
}

func autoUpgradeChannel(au *alicloudkubernetesclusterv1.AlicloudKubernetesClusterAutoUpgrade) string {
	if au.Channel != nil {
		return *au.Channel
	}
	return "patch"
}

func exportOutputs(ctx *pulumi.Context, cluster *cs.ManagedKubernetes) {
	ctx.Export(OpClusterId, cluster.ID())
	ctx.Export(OpClusterName, cluster.Name)

	ctx.Export(OpApiServerInternet, cluster.Connections.ApplyT(func(c cs.ManagedKubernetesConnections) string {
		if c.ApiServerInternet != nil {
			return *c.ApiServerInternet
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpApiServerIntranet, cluster.Connections.ApplyT(func(c cs.ManagedKubernetesConnections) string {
		if c.ApiServerIntranet != nil {
			return *c.ApiServerIntranet
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpVpcId, cluster.VpcId)
	ctx.Export(OpSecurityGroupId, cluster.SecurityGroupId)
	ctx.Export(OpNatGatewayId, cluster.NatGatewayId)
	ctx.Export(OpWorkerRamRoleName, cluster.WorkerRamRoleName)

	ctx.Export(OpRrsaOidcIssuerUrl, cluster.RrsaMetadata.ApplyT(func(m cs.ManagedKubernetesRrsaMetadata) string {
		if m.RrsaOidcIssuerUrl != nil {
			return *m.RrsaOidcIssuerUrl
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpRamOidcProviderName, cluster.RrsaMetadata.ApplyT(func(m cs.ManagedKubernetesRrsaMetadata) string {
		if m.RamOidcProviderName != nil {
			return *m.RamOidcProviderName
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpRamOidcProviderArn, cluster.RrsaMetadata.ApplyT(func(m cs.ManagedKubernetesRrsaMetadata) string {
		if m.RamOidcProviderArn != nil {
			return *m.RamOidcProviderArn
		}
		return ""
	}).(pulumi.StringOutput))
}
