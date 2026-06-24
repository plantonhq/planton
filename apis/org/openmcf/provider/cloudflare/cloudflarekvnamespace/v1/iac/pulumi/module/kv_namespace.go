package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// kvNamespace provisions the Workers KV namespace and exports its outputs.
func kvNamespace(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.WorkersKvNamespace, error) {

	// Build the namespace arguments directly from proto fields.
	kvArgs := &cloudflare.WorkersKvNamespaceArgs{
		AccountId: pulumi.String(locals.CloudflareKvNamespace.Spec.AccountId),
		Title:     pulumi.String(locals.CloudflareKvNamespace.Spec.NamespaceName),
		// NOTE:
		// The Cloudflare KV namespace resource does not expose "ttl_seconds" or
		// "description" fields, so those spec fields are not set here.
	}

	// Create the namespace.
	createdKvNamespace, err := cloudflare.NewWorkersKvNamespace(
		ctx,
		"kv_namespace",
		kvArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create workers kv namespace")
	}

	// Export the namespace ID as a stack output.
	ctx.Export(OpNamespaceId, createdKvNamespace.ID())

	return createdKvNamespace, nil
}
