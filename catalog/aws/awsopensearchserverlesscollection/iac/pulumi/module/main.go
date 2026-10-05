package module

import (
	"github.com/pkg/errors"
	awsopensearchserverlesscollectionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsopensearchserverlesscollection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the OpenSearch Serverless collection
// and its collection-scoped policies, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsopensearchserverlesscollectionv1alpha1.AwsOpenSearchServerlessCollectionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// The encryption policy MUST exist before the collection (AWS rejects
	// CreateCollection without a matching encryption policy), and the
	// dependency also serializes destroy the right way around (collection
	// first, then its policy).
	createdEncryptionPolicy, err := encryptionPolicy(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "encryption policy")
	}

	if err := collection(ctx, locals, provider, createdEncryptionPolicy); err != nil {
		return errors.Wrap(err, "collection")
	}

	// Network, data-access, and retention policies attach by name pattern;
	// they have no create-order requirement against the collection.
	if err := networkPolicy(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "network policy")
	}
	if err := dataAccessPolicy(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "data access policy")
	}
	if err := lifecyclePolicy(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "lifecycle policy")
	}

	return nil
}
