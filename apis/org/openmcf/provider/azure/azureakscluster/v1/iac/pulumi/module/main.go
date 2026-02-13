package module

import (
	"github.com/pkg/errors"
	azureaksclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureakscluster/v1"
	"github.com/pulumi/pulumi-azure-native-sdk/containerservice/v3"
	azurenative "github.com/pulumi/pulumi-azure-native-sdk/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureaksclusterv1.AzureAksClusterStackInput) error {
	azureProviderConfig := stackInput.ProviderConfig

	// Create Azure provider using the credentials from the input
	provider, err := azurenative.NewProvider(ctx,
		"azure",
		&azurenative.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// Get inputs
	target := stackInput.Target
	spec := target.Spec

	// Resource group is provided externally
	resourceGroupName := spec.ResourceGroup.GetValue()

	// Determine control plane SKU tier
	skuTier := "Standard" // Default to Standard for production
	if spec.ControlPlaneSku == azureaksclusterv1.AzureAksClusterControlPlaneSku_FREE {
		skuTier = "Free"
	}

	// Determine network plugin
	networkPlugin := "azure"
	if spec.NetworkPlugin == azureaksclusterv1.AzureAksClusterNetworkPlugin_KUBENET {
		networkPlugin = "kubenet"
	}

	// Determine network plugin mode (only for Azure CNI)
	var networkPluginMode pulumi.StringPtrInput
	if spec.NetworkPlugin == azureaksclusterv1.AzureAksClusterNetworkPlugin_AZURE_CNI {
		if spec.NetworkPluginMode == azureaksclusterv1.AzureAksClusterNetworkPluginMode_OVERLAY {
			networkPluginMode = pulumi.String("overlay")
		} else if spec.NetworkPluginMode == azureaksclusterv1.AzureAksClusterNetworkPluginMode_DYNAMIC {
			networkPluginMode = pulumi.String("dynamic")
		} else {
			// Default to overlay for Azure CNI
			networkPluginMode = pulumi.String("overlay")
		}
	}

	// Build authorized IP ranges
	var authorizedIpRanges pulumi.StringArray
	for _, cidr := range spec.AuthorizedIpRanges {
		authorizedIpRanges = append(authorizedIpRanges, pulumi.String(cidr))
	}

	// Determine AAD RBAC configuration
	aadEnabled := !spec.DisableAzureAdRbac

	// Build addon profiles
	addonProfiles := containerservice.ManagedClusterAddonProfileMap{}

	// Process add-ons configuration
	if spec.Addons != nil {
		// Container Insights (monitoring)
		if spec.Addons.EnableContainerInsights && spec.Addons.LogAnalyticsWorkspaceId != "" {
			addonProfiles["omsagent"] = &containerservice.ManagedClusterAddonProfileArgs{
				Enabled: pulumi.Bool(true),
				Config: pulumi.StringMap{
					"logAnalyticsWorkspaceResourceID": pulumi.String(spec.Addons.LogAnalyticsWorkspaceId),
				},
			}
		}

		// Key Vault CSI driver
		if spec.Addons.EnableKeyVaultCsiDriver {
			addonProfiles["azureKeyvaultSecretsProvider"] = &containerservice.ManagedClusterAddonProfileArgs{
				Enabled: pulumi.Bool(true),
			}
		}

		// Azure Policy
		if spec.Addons.EnableAzurePolicy {
			addonProfiles["azurepolicy"] = &containerservice.ManagedClusterAddonProfileArgs{
				Enabled: pulumi.Bool(true),
			}
		}
	}

	// Build network profile
	networkProfile := &containerservice.ContainerServiceNetworkProfileArgs{
		NetworkPlugin:   pulumi.String(networkPlugin),
		LoadBalancerSku: pulumi.String("standard"),
	}

	// Set network plugin mode if applicable
	if networkPluginMode != nil {
		networkProfile.NetworkPluginMode = networkPluginMode
	}

	// Apply advanced networking configuration if provided
	if spec.AdvancedNetworking != nil {
		if spec.AdvancedNetworking.PodCidr != "" {
			networkProfile.PodCidr = pulumi.String(spec.AdvancedNetworking.PodCidr)
		}
		if spec.AdvancedNetworking.ServiceCidr != "" {
			networkProfile.ServiceCidr = pulumi.String(spec.AdvancedNetworking.ServiceCidr)
		}
		if spec.AdvancedNetworking.DnsServiceIp != "" {
			networkProfile.DnsServiceIP = pulumi.String(spec.AdvancedNetworking.DnsServiceIp)
		}
	} else {
		// Use sensible defaults for service CIDR and DNS
		networkProfile.ServiceCidr = pulumi.String("10.0.0.0/16")
		networkProfile.DnsServiceIP = pulumi.String("10.0.0.10")
	}

	// Build system node pool from spec
	systemNodePool := spec.SystemNodePool
	if systemNodePool == nil {
		return errors.New("system_node_pool is required in spec")
	}

	var systemAvailabilityZones pulumi.StringArray
	for _, az := range systemNodePool.AvailabilityZones {
		systemAvailabilityZones = append(systemAvailabilityZones, pulumi.String(az))
	}

	agentPoolProfiles := containerservice.ManagedClusterAgentPoolProfileArray{
		&containerservice.ManagedClusterAgentPoolProfileArgs{
			Name:              pulumi.String("system"),
			Count:             pulumi.Int(int(systemNodePool.Autoscaling.MinCount)),
			VmSize:            pulumi.String(systemNodePool.VmSize),
			Mode:              pulumi.String("System"),
			OsType:            pulumi.String("Linux"),
			VnetSubnetID:      pulumi.String(spec.VnetSubnetId.GetValue()),
			EnableAutoScaling: pulumi.Bool(true),
			MinCount:          pulumi.Int(int(systemNodePool.Autoscaling.MinCount)),
			MaxCount:          pulumi.Int(int(systemNodePool.Autoscaling.MaxCount)),
			AvailabilityZones: systemAvailabilityZones,
		},
	}

	// Build AKS cluster arguments
	aksClusterArgs := &containerservice.ManagedClusterArgs{
		ResourceGroupName: pulumi.String(resourceGroupName),
		ResourceName:      pulumi.String(target.Metadata.Name),
		Location:          pulumi.String(spec.Region),

		// Control plane SKU
		Sku: &containerservice.ManagedClusterSKUArgs{
			Name: pulumi.String("Base"),
			Tier: pulumi.String(skuTier),
		},

		// Kubernetes version
		KubernetesVersion: pulumi.String(spec.KubernetesVersion),

		// DNS prefix
		DnsPrefix: pulumi.Sprintf("%s-dns", target.Metadata.Name),

		// Identity - Use System-Assigned Managed Identity
		Identity: &containerservice.ManagedClusterIdentityArgs{
			Type: containerservice.ResourceIdentityTypeSystemAssigned,
		},

		// Network Profile
		NetworkProfile: networkProfile,

		// System node pool
		AgentPoolProfiles: agentPoolProfiles,

		// API Server Access Profile
		ApiServerAccessProfile: &containerservice.ManagedClusterAPIServerAccessProfileArgs{
			EnablePrivateCluster:           pulumi.Bool(spec.PrivateClusterEnabled),
			AuthorizedIPRanges:             authorizedIpRanges,
			EnablePrivateClusterPublicFQDN: pulumi.Bool(!spec.PrivateClusterEnabled),
		},

		// Azure AD Integration
		AadProfile: &containerservice.ManagedClusterAADProfileArgs{
			Managed:         pulumi.Bool(aadEnabled),
			EnableAzureRBAC: pulumi.Bool(aadEnabled),
		},

		// Add-ons
		AddonProfiles: addonProfiles,
	}

	// Enable Workload Identity if specified in add-ons
	if spec.Addons != nil && spec.Addons.EnableWorkloadIdentity {
		aksClusterArgs.OidcIssuerProfile = &containerservice.ManagedClusterOIDCIssuerProfileArgs{
			Enabled: pulumi.Bool(true),
		}
		aksClusterArgs.SecurityProfile = &containerservice.ManagedClusterSecurityProfileArgs{
			WorkloadIdentity: &containerservice.ManagedClusterSecurityProfileWorkloadIdentityArgs{
				Enabled: pulumi.Bool(true),
			},
		}
	}

	// Create AKS cluster
	aksCluster, err := containerservice.NewManagedCluster(ctx, target.Metadata.Name, aksClusterArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create AKS cluster")
	}

	// Create user node pools if specified
	for _, userPool := range spec.UserNodePools {
		var userAvailabilityZones pulumi.StringArray
		for _, az := range userPool.AvailabilityZones {
			userAvailabilityZones = append(userAvailabilityZones, pulumi.String(az))
		}

		userPoolArgs := &containerservice.AgentPoolArgs{
			ResourceGroupName: pulumi.String(resourceGroupName),
			ResourceName:      aksCluster.Name,
			AgentPoolName:     pulumi.String(userPool.Name),
			Count:             pulumi.Int(int(userPool.Autoscaling.MinCount)),
			VmSize:            pulumi.String(userPool.VmSize),
			Mode:              pulumi.String("User"),
			OsType:            pulumi.String("Linux"),
			VnetSubnetID:      pulumi.String(spec.VnetSubnetId.GetValue()),
			EnableAutoScaling: pulumi.Bool(true),
			MinCount:          pulumi.Int(int(userPool.Autoscaling.MinCount)),
			MaxCount:          pulumi.Int(int(userPool.Autoscaling.MaxCount)),
			AvailabilityZones: userAvailabilityZones,
		}

		// Enable spot instances if specified
		if userPool.SpotEnabled {
			userPoolArgs.ScaleSetPriority = pulumi.String("Spot")
			userPoolArgs.ScaleSetEvictionPolicy = pulumi.String("Delete")
			userPoolArgs.SpotMaxPrice = pulumi.Float64(-1) // Pay up to regular price
		}

		_, err := containerservice.NewAgentPool(ctx, userPool.Name, userPoolArgs, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{aksCluster}))
		if err != nil {
			return errors.Wrapf(err, "failed to create user node pool: %s", userPool.Name)
		}
	}

	// Get kubeconfig
	listCredentials := containerservice.ListManagedClusterUserCredentialsOutput(ctx, containerservice.ListManagedClusterUserCredentialsOutputArgs{
		ResourceGroupName: pulumi.String(resourceGroupName),
		ResourceName:      aksCluster.Name,
	})

	kubeconfig := listCredentials.Kubeconfigs().Index(pulumi.Int(0)).Value()

	// Export outputs aligned to AzureAksClusterStackOutputs
	ctx.Export(OpApiServerEndpoint, aksCluster.Fqdn)
	ctx.Export(OpClusterResourceId, aksCluster.ID())
	ctx.Export(OpClusterKubeconfig, kubeconfig)
	ctx.Export(OpManagedIdentityPrincipalId, aksCluster.IdentityProfile.MapIndex(pulumi.String("kubeletidentity")).ObjectId())

	return nil
}
