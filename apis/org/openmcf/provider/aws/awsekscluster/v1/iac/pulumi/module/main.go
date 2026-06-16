package module

import (
	"github.com/pkg/errors"
	awseksclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsekscluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awseksclusterv1.AwsEksClusterStackInput) (err error) {
	// Initialize locals with computed values
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsEksCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Inputs
	target := locals.AwsEksCluster
	spec := target.Spec

	// Build subnet IDs input
	var subnetIds pulumi.StringArray
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
	}

	// Control plane logs
	var logTypes pulumi.StringArray
	if spec.EnableControlPlaneLogs {
		logTypes = pulumi.StringArray{
			pulumi.String("api"), pulumi.String("audit"), pulumi.String("authenticator"), pulumi.String("controllerManager"), pulumi.String("scheduler"),
		}
	}

	clusterArgs := &eks.ClusterArgs{
		Name:    pulumi.String(target.Metadata.Name),
		RoleArn: pulumi.String(spec.ClusterRoleArn.GetValue()),
		Version: pulumi.String(spec.Version),
		VpcConfig: &eks.ClusterVpcConfigArgs{
			SubnetIds:            subnetIds,
			EndpointPublicAccess: pulumi.Bool(!spec.DisablePublicEndpoint),
			PublicAccessCidrs:    pulumi.ToStringArray(spec.PublicAccessCidrs),
		},
		EnabledClusterLogTypes: logTypes,
		Tags:                   pulumi.ToStringMap(locals.AwsTags),
	}

	// Add KMS encryption if specified
	if spec.KmsKeyArn != nil && spec.KmsKeyArn.GetValue() != "" {
		clusterArgs.EncryptionConfig = &eks.ClusterEncryptionConfigArgs{
			Provider: &eks.ClusterEncryptionConfigProviderArgs{
				KeyArn: pulumi.String(spec.KmsKeyArn.GetValue()),
			},
			Resources: pulumi.StringArray{
				pulumi.String("secrets"),
			},
		}
	}

	createdCluster, err := eks.NewCluster(ctx, target.Metadata.Name, clusterArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create EKS cluster")
	}

	// Extract OIDC issuer URL from the cluster identity
	oidcIssuerUrl := createdCluster.Identities.Index(pulumi.Int(0)).Oidcs().Index(pulumi.Int(0)).Issuer()

	// Export outputs aligned to AwsEksClusterStackOutputs
	ctx.Export(OpEndpoint, createdCluster.Endpoint)
	ctx.Export(OpClusterCaCertificate, createdCluster.CertificateAuthority.Data().Elem())
	ctx.Export(OpClusterSecurityGroupId, createdCluster.VpcConfig.ClusterSecurityGroupId().Elem())
	ctx.Export(OpOidcIssuerUrl, oidcIssuerUrl)
	ctx.Export(OpClusterArn, createdCluster.Arn)
	ctx.Export(OpName, createdCluster.Name)

	return nil
}
