package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// kvPair writes a single key-value entry into the target KV namespace.
func kvPair(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.WorkersKv, error) {
	spec := locals.CloudflareWorkersKvPair.Spec

	namespaceId := ""
	if spec.NamespaceId != nil {
		namespaceId = spec.NamespaceId.GetValue()
	}

	args := &cloudflare.WorkersKvArgs{
		AccountId:   pulumi.String(spec.AccountId),
		NamespaceId: pulumi.String(namespaceId),
		KeyName:     pulumi.String(spec.KeyName),
		Value:       pulumi.String(spec.Value),
	}
	if spec.Metadata != "" {
		args.Metadata = pulumi.StringPtr(spec.Metadata)
	}

	created, err := cloudflare.NewWorkersKv(
		ctx,
		"kv-pair",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare workers kv pair")
	}

	ctx.Export(OpKeyName, pulumi.String(spec.KeyName))
	ctx.Export(OpNamespaceId, pulumi.String(namespaceId))

	return created, nil
}
