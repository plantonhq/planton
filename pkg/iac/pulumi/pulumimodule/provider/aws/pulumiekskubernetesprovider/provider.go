package pulumiekskubernetesprovider

import (
	"github.com/pkg/errors"
	aws "github.com/plantonhq/planton/apis/dev/planton/provider/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// GetWithCreatedEksClusterWithAwsProviderConfigs returns kubernetes provider for the added eks cluster based on the aws provider
func GetWithCreatedEksClusterWithAwsProviderConfigs(ctx *pulumi.Context, createdEksCluster *eks.Cluster,
	awsProviderConfig *aws.AwsProviderConfig,
	dependencies []pulumi.Resource, providerName string) (*kubernetes.Provider, error) {
	provider, err := kubernetes.NewProvider(ctx,
		providerName,
		&kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
			Kubeconfig: pulumi.Sprintf(AwsExecPluginKubeConfigTemplate,
				createdEksCluster.Endpoint,
				createdEksCluster.CertificateAuthority.Data().Elem(),
				awsProviderConfig.AccessKeyId,
				awsProviderConfig.SecretAccessKey,
			),
		}, pulumi.DependsOn(dependencies))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get new provider")
	}
	return provider, nil
}
