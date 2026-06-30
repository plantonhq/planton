package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster provisions the Kubernetes cluster itself and exports its outputs.
func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.KubernetesCluster, error) {

	// 1. Collect tags from the spec.
	var tags pulumi.StringArray
	for _, t := range locals.DigitalOceanKubernetesCluster.Spec.Tags {
		tags = append(tags, pulumi.String(t))
	}

	// 2. Build maintenance window configuration if provided
	var maintenancePolicy *digitalocean.KubernetesClusterMaintenancePolicyArgs
	if locals.DigitalOceanKubernetesCluster.Spec.MaintenanceWindow != "" {
		maintenancePolicy = &digitalocean.KubernetesClusterMaintenancePolicyArgs{
			StartTime: pulumi.String(locals.DigitalOceanKubernetesCluster.Spec.MaintenanceWindow),
		}
	}

	// 3. Build control plane firewall configuration if IPs provided
	var controlPlaneFirewall *digitalocean.KubernetesClusterControlPlaneFirewallArgs
	if len(locals.DigitalOceanKubernetesCluster.Spec.ControlPlaneFirewallAllowedIps) > 0 {
		var allowedIPs pulumi.StringArray
		for _, ip := range locals.DigitalOceanKubernetesCluster.Spec.ControlPlaneFirewallAllowedIps {
			allowedIPs = append(allowedIPs, pulumi.String(ip))
		}
		controlPlaneFirewall = &digitalocean.KubernetesClusterControlPlaneFirewallArgs{
			AllowedAddresses: allowedIPs,
		}
	}

	// 4. Build the cluster arguments straight from proto fields.
	clusterArgs := &digitalocean.KubernetesClusterArgs{
		Name:                 pulumi.String(locals.DigitalOceanKubernetesCluster.Spec.ClusterName),
		Region:               pulumi.String(locals.DigitalOceanKubernetesCluster.Spec.Region.String()),
		Version:              pulumi.String(locals.DigitalOceanKubernetesCluster.Spec.KubernetesVersion),
		Ha:                   pulumi.BoolPtr(locals.DigitalOceanKubernetesCluster.Spec.HighlyAvailable),
		AutoUpgrade:          pulumi.BoolPtr(locals.DigitalOceanKubernetesCluster.Spec.AutoUpgrade),
		SurgeUpgrade:         pulumi.BoolPtr(!locals.DigitalOceanKubernetesCluster.Spec.DisableSurgeUpgrade),
		VpcUuid:              pulumi.String(locals.DigitalOceanKubernetesCluster.Spec.Vpc.GetValue()),
		RegistryIntegration:  pulumi.BoolPtr(locals.DigitalOceanKubernetesCluster.Spec.RegistryIntegration),
		MaintenancePolicy:    maintenancePolicy,
		ControlPlaneFirewall: controlPlaneFirewall,
		Tags:                 tags,
		NodePool: &digitalocean.KubernetesClusterNodePoolArgs{
			AutoScale: pulumi.BoolPtr(locals.DigitalOceanKubernetesCluster.Spec.DefaultNodePool.AutoScale),
			MaxNodes:  pulumi.IntPtr(int(locals.DigitalOceanKubernetesCluster.Spec.DefaultNodePool.MaxNodes)),
			MinNodes:  pulumi.IntPtr(int(locals.DigitalOceanKubernetesCluster.Spec.DefaultNodePool.MinNodes)),
			Name:      pulumi.String("default"),
			NodeCount: pulumi.IntPtr(int(locals.DigitalOceanKubernetesCluster.Spec.DefaultNodePool.NodeCount)),
			Size:      pulumi.String(locals.DigitalOceanKubernetesCluster.Spec.DefaultNodePool.Size),
		},
	}

	// 5. Create the cluster.
	createdCluster, err := digitalocean.NewKubernetesCluster(
		ctx,
		"cluster",
		clusterArgs,
		pulumi.Provider(digitalOceanProvider),
		pulumi.IgnoreChanges([]string{"version"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean kubernetes cluster")
	}

	// 6. Export required stack outputs.
	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpApiServerEndpoint, createdCluster.Endpoint)

	// Use the first kubeconfig in the list.
	kubeconfig := createdCluster.KubeConfigs.Index(pulumi.Int(0)).RawConfig()
	ctx.Export(OpKubeconfig, kubeconfig)

	return createdCluster, nil
}
