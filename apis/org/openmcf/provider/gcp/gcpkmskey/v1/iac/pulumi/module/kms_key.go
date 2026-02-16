package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kmsKey(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpKmsKey.Spec

	args := &kms.CryptoKeyArgs{
		Name:    pulumi.String(spec.KeyName),
		KeyRing: pulumi.String(spec.KeyRingId.GetValue()),
		Labels:  pulumi.ToStringMap(locals.GcpLabels),
	}

	// Set purpose if specified (GCP defaults to ENCRYPT_DECRYPT).
	if spec.Purpose != "" {
		args.Purpose = pulumi.StringPtr(spec.Purpose)
	}

	// Set rotation period if specified.
	if spec.RotationPeriod != "" {
		args.RotationPeriod = pulumi.StringPtr(spec.RotationPeriod)
	}

	// Set destroy scheduled duration if specified (GCP defaults to 30 days).
	if spec.DestroyScheduledDuration != "" {
		args.DestroyScheduledDuration = pulumi.StringPtr(spec.DestroyScheduledDuration)
	}

	// Set skip initial version creation if explicitly requested.
	if spec.SkipInitialVersionCreation {
		args.SkipInitialVersionCreation = pulumi.BoolPtr(true)
	}

	// Set version template if specified.
	if spec.VersionTemplate != nil {
		vt := &kms.CryptoKeyVersionTemplateArgs{
			Algorithm: pulumi.String(spec.VersionTemplate.Algorithm),
		}
		if spec.VersionTemplate.ProtectionLevel != "" {
			vt.ProtectionLevel = pulumi.StringPtr(spec.VersionTemplate.ProtectionLevel)
		}
		args.VersionTemplate = vt
	}

	createdKey, err := kms.NewCryptoKey(ctx, "kms-key", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create kms crypto key")
	}

	// Export the fully qualified crypto key resource path.
	// Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{name}
	// This is the primary CMEK reference used by downstream resources.
	ctx.Export(OpKeyId, createdKey.ID())
	ctx.Export(OpKeyName, createdKey.Name)

	return nil
}
