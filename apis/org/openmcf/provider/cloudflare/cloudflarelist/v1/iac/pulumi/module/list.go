package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// list provisions an account-scoped Cloudflare List. Items are managed
// separately (CloudflareListItem), so no inline items are set here.
func list(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.List, error) {
	spec := locals.CloudflareList.Spec

	args := &cloudflare.ListArgs{
		AccountId: pulumi.String(spec.AccountId),
		Kind:      pulumi.String(spec.Kind.String()),
		Name:      pulumi.String(spec.Name),
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	created, err := cloudflare.NewList(
		ctx,
		"list",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare list")
	}

	ctx.Export(OpListId, created.ID())
	ctx.Export(OpName, created.Name)
	ctx.Export(OpKind, created.Kind)

	return created, nil
}
