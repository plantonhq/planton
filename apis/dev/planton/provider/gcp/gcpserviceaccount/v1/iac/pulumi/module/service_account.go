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

	// Create the service account.
	createdServiceAccount, err := serviceaccount.NewAccount(
		ctx,
		locals.GcpServiceAccount.Metadata.Name,
		&serviceaccount.AccountArgs{
			AccountId:   pulumi.String(locals.GcpServiceAccount.Spec.ServiceAccountId),
			DisplayName: pulumi.String(locals.GcpServiceAccount.Metadata.Name),
			Project:     pulumi.String(locals.GcpServiceAccount.Spec.ProjectId.GetValue()),
		},
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create service account")
	}

	// Optionally create a key.
	var createdKey *serviceaccount.Key
	if locals.GcpServiceAccount.Spec.GetCreateKey() {
		createdKey, err = serviceaccount.NewKey(
			ctx,
			fmt.Sprintf("%s-key", locals.GcpServiceAccount.Metadata.Name),
			&serviceaccount.KeyArgs{
				ServiceAccountId: createdServiceAccount.Name,
			},
			pulumi.Provider(gcpProvider),
			pulumi.DependsOn([]pulumi.Resource{createdServiceAccount}),
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create service account key")
		}
	}

	return createdServiceAccount, createdKey, nil
}
