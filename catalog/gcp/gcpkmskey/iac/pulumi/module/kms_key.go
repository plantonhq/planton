package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/kms"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kmsKey(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpKmsKey.Spec

	// The key inherits its project from the ring path
	// (projects/{project}/locations/...), but the API-enablement resource
	// needs the project ID itself — extract it from the resolved path's
	// second segment.
	keyRingId := spec.KeyRingId.GetValue()
	ringPathParts := strings.Split(keyRingId, "/")
	if len(ringPathParts) < 2 || ringPathParts[0] != "projects" {
		return errors.Errorf("key_ring_id %q is not a fully qualified key ring path (projects/{project}/locations/{location}/keyRings/{name})", keyRingId)
	}
	ringProject := ringPathParts[1]

	// Enable the Cloud KMS API — the control plane that owns the key.
	// DisableOnDestroy stays false: tearing down one key must never
	// disable the API for everything else in the project (other keys may
	// be actively encrypting production data).
	createdCloudkmsApi, err := projects.NewService(ctx,
		"gcpkms-cloudkms.googleapis.com", &projects.ServiceArgs{
			Project:                  pulumi.String(ringProject),
			Service:                  pulumi.String("cloudkms.googleapis.com"),
			DisableDependentServices: pulumi.BoolPtr(true),
			DisableOnDestroy:         pulumi.BoolPtr(false),
		}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable cloudkms.googleapis.com api")
	}

	// The KMS crypto key — the resource downstream services reference for
	// customer-managed encryption (CMEK). Lifecycle sharp edges, all
	// taught by the API rather than invented here:
	//
	//   - The key can never be deleted from GCP. On destroy the engine
	//     destroys all key versions (data encrypted under them becomes
	//     unrecoverable once the destroy-scheduled window elapses),
	//     disables automatic rotation, and removes the key from state —
	//     the key object itself remains, permanently and at no cost, in
	//     the ring.
	//
	//   - Only rotation_period, version_template.algorithm, and labels
	//     update in place; every other field is immutable, which for an
	//     undeletable resource means "abandon and create under a new name".
	args := &kms.CryptoKeyArgs{
		Name:    pulumi.String(spec.KeyName),
		KeyRing: pulumi.String(keyRingId),
		Labels:  pulumi.ToStringMap(locals.GcpLabels),
	}

	// DELETE (provider default) destroys every key version on destroy;
	// PREVENT fails the destroy; ABANDON leaves versions intact and the
	// key decryptable. Sent only when set — mirrors the Terraform module.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Purpose defaults to ENCRYPT_DECRYPT server-side; send only when set.
	if spec.Purpose != "" {
		args.Purpose = pulumi.StringPtr(spec.Purpose)
	}

	// Rotation mints a new primary version on the cadence; old versions
	// stay decryptable until destroyed. Only valid for ENCRYPT_DECRYPT
	// keys (enforced pre-deploy by the spec).
	if spec.RotationPeriod != "" {
		args.RotationPeriod = pulumi.StringPtr(spec.RotationPeriod)
	}

	// The recovery window for destroyed versions (default 30 days).
	if spec.DestroyScheduledDuration != "" {
		args.DestroyScheduledDuration = pulumi.StringPtr(spec.DestroyScheduledDuration)
	}

	// Create-time-only flag; required for import_only keys, where GCP
	// must never generate material.
	if spec.SkipInitialVersionCreation {
		args.SkipInitialVersionCreation = pulumi.BoolPtr(true)
	}

	// BYOK: the key may only ever hold imported versions.
	if spec.ImportOnly {
		args.ImportOnly = pulumi.BoolPtr(true)
	}

	// EKM connection for EXTERNAL_VPC keys (the spec enforces the pairing
	// pre-deploy). Absent for the SOFTWARE/HSM/EXTERNAL protection levels.
	if spec.CryptoKeyBackend != nil && spec.CryptoKeyBackend.GetValue() != "" {
		args.CryptoKeyBackend = pulumi.StringPtr(spec.CryptoKeyBackend.GetValue())
	}

	// Version template: algorithm affects only versions created after a
	// change; protection level is immutable.
	if spec.VersionTemplate != nil {
		vt := &kms.CryptoKeyVersionTemplateArgs{
			Algorithm: pulumi.String(spec.VersionTemplate.Algorithm),
		}
		if spec.VersionTemplate.ProtectionLevel != "" {
			vt.ProtectionLevel = pulumi.StringPtr(spec.VersionTemplate.ProtectionLevel)
		}
		args.VersionTemplate = vt
	}

	createdKey, err := kms.NewCryptoKey(ctx, "kms-key", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdCloudkmsApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create kms crypto key")
	}

	// The resource ID is the fully qualified crypto key path
	// (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{name}) — the
	// exact string every CMEK consumer passes to its kms_key_name-style
	// attribute.
	ctx.Export(OpKeyId, createdKey.ID())
	ctx.Export(OpKeyName, createdKey.Name)

	// GCP populates primary only for ENCRYPT_DECRYPT keys; export empty
	// strings for asymmetric/raw/MAC keys and for keys created without an
	// initial version — identical to the Terraform module's try() guards.
	// (Guard on array length: indexing the empty Primaries array would
	// panic inside the generated Index accessor.)
	ctx.Export(OpPrimaryVersionName, createdKey.Primaries.ApplyT(func(primaries []kms.CryptoKeyPrimary) string {
		if len(primaries) == 0 || primaries[0].Name == nil {
			return ""
		}
		return *primaries[0].Name
	}).(pulumi.StringOutput))
	ctx.Export(OpPrimaryState, createdKey.Primaries.ApplyT(func(primaries []kms.CryptoKeyPrimary) string {
		if len(primaries) == 0 || primaries[0].State == nil {
			return ""
		}
		return *primaries[0].State
	}).(pulumi.StringOutput))

	// Version 1 is the one Google creates with the key, whatever its
	// purpose; asymmetric-sign keys have no primary, so this is the version
	// their consumers name. Empty when no initial version was created --
	// identical to the Terraform output.
	if spec.SkipInitialVersionCreation {
		ctx.Export(OpInitialVersionName, pulumi.String(""))
	} else {
		ctx.Export(OpInitialVersionName, createdKey.ID().ApplyT(func(id pulumi.ID) string {
			return string(id) + "/cryptoKeyVersions/1"
		}).(pulumi.StringOutput))
	}

	return nil
}
