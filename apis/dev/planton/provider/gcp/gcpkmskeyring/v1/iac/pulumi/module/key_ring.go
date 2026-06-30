package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func keyRing(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpKmsKeyRing.Spec

	createdKeyRing, err := kms.NewKeyRing(ctx, "key-ring", &kms.KeyRingArgs{
		Name:     pulumi.String(spec.KeyRingName),
		Location: pulumi.String(spec.Location),
		Project:  pulumi.String(spec.ProjectId.GetValue()),
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create kms key ring")
	}

	// Export the fully qualified key ring resource path.
	// Format: projects/{project}/locations/{location}/keyRings/{name}
	// This is the primary reference used by GcpKmsCryptoKey.
	ctx.Export(OpKeyRingId, createdKeyRing.ID())
	ctx.Export(OpKeyRingName, createdKeyRing.Name)

	return nil
}
