package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceAccount provisions the account and (optionally) a key.
// Returns the created account and key pointers (key may be nil).
func serviceAccount(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*serviceaccount.Account, *serviceaccount.Key, error) {
	spec := locals.GcpServiceAccount.Spec

	// Display name falls back to the resource's metadata name so every account
	// is human-identifiable in the console even when the field is omitted.
	displayName := spec.DisplayName
	if displayName == "" {
		displayName = locals.GcpServiceAccount.Metadata.Name
	}

	accountArgs := &serviceaccount.AccountArgs{
		// account_id and project are immutable (ForceNew in the provider):
		// changing either destroys and recreates the account, invalidating every
		// IAM binding and workload identity referencing the old email.
		AccountId:   pulumi.String(spec.ServiceAccountId),
		DisplayName: pulumi.String(displayName),
		// GCP flips disabled state via separate Enable/Disable API calls (not the
		// regular update mask); the provider handles that internally, so toggling
		// this field never recreates the account.
		Disabled: pulumi.Bool(spec.GetDisabled()),
	}

	// Omitted description stays unset (matching the Terraform module's null)
	// rather than being sent as an empty string.
	if spec.Description != "" {
		accountArgs.Description = pulumi.String(spec.Description)
	}

	// Honor the spec contract: an empty project_id falls back to the provider's
	// default project. Leaving Project unset lets the gcp provider resolve its
	// own project (configuration or the GOOGLE_PROJECT / GOOGLE_CLOUD_PROJECT
	// environment chain); an empty string would be sent verbatim and rejected.
	if spec.ProjectId.GetValue() != "" {
		accountArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// Adopt an existing account with the same email instead of failing —
	// idempotent bootstrap flows that may race other provisioning paths.
	if spec.CreateIgnoreAlreadyExists {
		accountArgs.CreateIgnoreAlreadyExists = pulumi.Bool(true)
	}

	// Destroy-time guard: PREVENT fails any destroy while set. Unset falls
	// back to the provider default (DELETE).
	if spec.DeletionPolicy != "" {
		accountArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdServiceAccount, err := serviceaccount.NewAccount(
		ctx,
		locals.GcpServiceAccount.Metadata.Name,
		accountArgs,
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create service account")
	}

	// Optional user-managed key. Created only when spec.user_managed_key is
	// present — keyless patterns (Workload Identity, impersonation,
	// federation) are the recommended default. In the generate flow the
	// private key is returned once at creation and marked secret in state by
	// the engine; in the upload flow (public_key_data set) GCP never sees a
	// private key at all.
	var createdKey *serviceaccount.Key
	// The block's switch (on by default once declared, filled by the platform
	// before this runs) decides whether a key exists.
	if spec.UserManagedKey != nil && spec.UserManagedKey.GetEnabled() {
		keyArgs := &serviceaccount.KeyArgs{
			ServiceAccountId: createdServiceAccount.Name,
		}

		// Generate-flow shape. Unset fields fall back to GCP defaults
		// (2048-bit RSA, JSON credentials file, X.509 PEM public key).
		if spec.UserManagedKey.Algorithm != "" {
			keyArgs.KeyAlgorithm = pulumi.String(spec.UserManagedKey.Algorithm)
		}
		if spec.UserManagedKey.PrivateKeyType != "" {
			keyArgs.PrivateKeyType = pulumi.String(spec.UserManagedKey.PrivateKeyType)
		}
		if spec.UserManagedKey.PublicKeyType != "" {
			keyArgs.PublicKeyType = pulumi.String(spec.UserManagedKey.PublicKeyType)
		}

		// Upload-flow shape: the caller's own public key (base64 X.509 PEM);
		// mutually exclusive with the *_type args above (spec CEL enforces it).
		if spec.UserManagedKey.PublicKeyData != "" {
			keyArgs.PublicKeyData = pulumi.String(spec.UserManagedKey.PublicKeyData)
		}

		// Rotation trigger: any change to this map replaces the key.
		if len(spec.UserManagedKey.Keepers) > 0 {
			keyArgs.Keepers = pulumi.ToStringMap(spec.UserManagedKey.Keepers)
		}

		if spec.UserManagedKey.DeletionPolicy != "" {
			keyArgs.DeletionPolicy = pulumi.String(spec.UserManagedKey.DeletionPolicy)
		}

		createdKey, err = serviceaccount.NewKey(
			ctx,
			fmt.Sprintf("%s-key", locals.GcpServiceAccount.Metadata.Name),
			keyArgs,
			pulumi.Provider(gcpProvider),
			pulumi.DependsOn([]pulumi.Resource{createdServiceAccount}),
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create service account key")
		}
	}

	return createdServiceAccount, createdKey, nil
}
