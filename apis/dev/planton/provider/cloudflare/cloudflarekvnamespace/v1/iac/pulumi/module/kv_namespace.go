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
	// A KV namespace carries only an account and a title; entries are seeded as
	// CloudflareWorkersKvPair resources or written by the Worker at runtime.
	kvArgs := &cloudflare.WorkersKvNamespaceArgs{
		AccountId: pulumi.String(locals.CloudflareKvNamespace.Spec.AccountId),
		Title:     pulumi.String(locals.CloudflareKvNamespace.Spec.NamespaceName),
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

	// Export the namespace ID and URL-encoding support as stack outputs.
	ctx.Export(OpNamespaceId, createdKvNamespace.ID())
	ctx.Export(OpSupportsUrlEncoding, createdKvNamespace.SupportsUrlEncoding)

	return createdKvNamespace, nil
}
