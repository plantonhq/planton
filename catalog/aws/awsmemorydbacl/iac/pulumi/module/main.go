package module

import (
	"github.com/pkg/errors"
	awsmemorydbaclv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmemorydbacl/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the MemoryDB ACL -- the attachment unit between users
// and clusters. Membership is modeled here as references: the ACL is the
// single place an application's cluster access is granted or revoked, and
// this module never mutates the users it references.
func Resources(ctx *pulumi.Context, iacInput *awsmemorydbaclv1alpha1.AwsMemorydbAclIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsMemorydbAcl.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := acl(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create MemoryDB ACL")
	}

	return nil
}
