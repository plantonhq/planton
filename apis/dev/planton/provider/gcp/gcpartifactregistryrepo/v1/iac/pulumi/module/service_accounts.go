package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ServiceAccounts holds the created reader and writer service accounts and their keys
type ServiceAccounts struct {
	ReaderAccount *serviceaccount.Account
	ReaderKey     *serviceaccount.Key
	WriterAccount *serviceaccount.Account
	WriterKey     *serviceaccount.Key
}

// createServiceAccounts creates reader and writer service accounts with keys
// for artifact registry access
func createServiceAccounts(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*ServiceAccounts, error) {
	gcpArtifactRegistryRepo := locals.GcpArtifactRegistryRepo

	// Get project ID from StringValueOrRef (currently only supports literal value)
	// TODO: Implement reference resolution in a shared library
	projectId := gcpArtifactRegistryRepo.Spec.ProjectId.GetValue()

	// Generate a random 6-character suffix for service account uniqueness
	suffix, err := random.NewRandomString(
		ctx,
		fmt.Sprintf("%s-sa-suffix", gcpArtifactRegistryRepo.Metadata.Name),
		&random.RandomStringArgs{
			Length:  pulumi.Int(6),
			Special: pulumi.Bool(false),
			Numeric: pulumi.Bool(true),
			Upper:   pulumi.Bool(false),
			Lower:   pulumi.Bool(true),
		},
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate random suffix for service accounts")
	}

	// Create reader service account
	readerAccountId := pulumi.Sprintf("%s-%s-ro", gcpArtifactRegistryRepo.Metadata.Name, suffix.Result)
	readerAccount, err := serviceaccount.NewAccount(
		ctx,
		fmt.Sprintf("%s-reader", gcpArtifactRegistryRepo.Metadata.Name),
		&serviceaccount.AccountArgs{
			Project:     pulumi.String(projectId),
			AccountId:   readerAccountId,
			DisplayName: readerAccountId,
		},
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reader service account")
	}

	// Create key for reader service account
	readerKey, err := serviceaccount.NewKey(
		ctx,
		fmt.Sprintf("%s-reader-key", gcpArtifactRegistryRepo.Metadata.Name),
		&serviceaccount.KeyArgs{
			ServiceAccountId: readerAccount.Name,
		},
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{readerAccount}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reader service account key")
	}

	// Create writer service account
	writerAccountId := pulumi.Sprintf("%s-%s-rw", gcpArtifactRegistryRepo.Metadata.Name, suffix.Result)
	writerAccount, err := serviceaccount.NewAccount(
		ctx,
		fmt.Sprintf("%s-writer", gcpArtifactRegistryRepo.Metadata.Name),
		&serviceaccount.AccountArgs{
			Project:     pulumi.String(projectId),
			AccountId:   writerAccountId,
			DisplayName: writerAccountId,
		},
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create writer service account")
	}

	// Create key for writer service account
	writerKey, err := serviceaccount.NewKey(
		ctx,
		fmt.Sprintf("%s-writer-key", gcpArtifactRegistryRepo.Metadata.Name),
		&serviceaccount.KeyArgs{
			ServiceAccountId: writerAccount.Name,
		},
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{writerAccount}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create writer service account key")
	}

	return &ServiceAccounts{
		ReaderAccount: readerAccount,
		ReaderKey:     readerKey,
		WriterAccount: writerAccount,
		WriterKey:     writerKey,
	}, nil
}
