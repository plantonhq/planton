package pulumiakskubernetesprovider

import (
	"github.com/pkg/errors"
	azure "github.com/plantonhq/planton/apis/dev/planton/provider/azure"
	"github.com/pulumi/pulumi-azure/sdk/v5/go/azure/containerservice"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// GetWithAddedClusterWithAzureProviderConfigs returns kubernetes provider for the added AKS cluster based on the azure provider
func GetWithAddedClusterWithAzureProviderConfigs(ctx *pulumi.Context,
	addedAksCluster *containerservice.KubernetesCluster,
	azureProviderConfig *azure.AzureProviderConfig,
	dependencies []pulumi.Resource,
	providerName string) (*kubernetes.Provider, error) {

	clusterCaCert := addedAksCluster.KubeConfigs.ApplyT(
		func(kubeConfigs []containerservice.KubernetesClusterKubeConfig) string {
			return *kubeConfigs[0].ClusterCaCertificate
		})

	provider, err := kubernetes.NewProvider(ctx,
		providerName,
		&kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
			Kubeconfig: pulumi.Sprintf(AzureExecPluginKubeConfigTemplate,
				addedAksCluster.Fqdn,
				clusterCaCert,
				azureProviderConfig.ClientId,
				azureProviderConfig.ClientSecret,
				azureProviderConfig.TenantId,
			),
		}, pulumi.DependsOn(dependencies))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get new provider")
	}
	return provider, nil
}
