package module

import (
	"github.com/pkg/errors"
	awsredshiftserverlessnamespacev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsredshiftserverlessnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Redshift Serverless namespace -- the data
// plane of the serverless warehouse. The namespace composes onto its
// neighbors instead of embedding them: KMS keys and IAM roles attach by
// reference, and the compute that serves this data lives on
// AwsRedshiftServerlessWorkgroup nodes that attach by name -- this
// module never creates or mutates a resource that deserves to be its
// own node.
func Resources(ctx *pulumi.Context, iacInput *awsredshiftserverlessnamespacev1alpha1.AwsRedshiftServerlessNamespaceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsRedshiftServerlessNamespace.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdNamespace, err := namespace(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Redshift Serverless namespace")
	}

	// The name is the join key workgroups attach with, so it must
	// surface as an output -- downstream references resolve against
	// outputs, never metadata.
	ctx.Export(OpNamespaceName, createdNamespace.NamespaceName)
	ctx.Export(OpNamespaceId, createdNamespace.NamespaceId)
	ctx.Export(OpArn, createdNamespace.Arn)
	ctx.Export(OpDbName, createdNamespace.DbName)

	// The AWS-managed admin-password secret exists only when
	// manage_admin_password is on; the attribute resolves to "" otherwise,
	// so the output shape is stable across both password strategies.
	ctx.Export(OpAdminPasswordSecretArn, createdNamespace.AdminPasswordSecretArn)

	return nil
}
