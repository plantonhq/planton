package module

import (
	"github.com/pkg/errors"
	gcpartifactregistryrepov1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpartifactregistryrepo/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpartifactregistryrepov1.GcpArtifactRegistryRepoStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	//create google provider using the credentials from the input
	googleProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create google provider")
	}

	//create service accounts for reader and writer access
	serviceAccounts, err := createServiceAccounts(ctx, locals, googleProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create service accounts")
	}

	//create docker repository with public access configuration and IAM bindings
	if err := repo(ctx, locals, googleProvider, serviceAccounts); err != nil {
		return errors.Wrap(err, "failed to create docker repo")
	}

	//export service account outputs
	ctx.Export(OpReaderServiceAccountEmail, serviceAccounts.ReaderAccount.Email)
	ctx.Export(OpReaderServiceAccountKeyBase64, serviceAccounts.ReaderKey.PrivateKey)
	ctx.Export(OpWriterServiceAccountEmail, serviceAccounts.WriterAccount.Email)
	ctx.Export(OpWriterServiceAccountKeyBase64, serviceAccounts.WriterKey.PrivateKey)

	return nil
}
