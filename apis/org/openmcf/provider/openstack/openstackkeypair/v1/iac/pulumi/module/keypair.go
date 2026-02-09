package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// keypair provisions the OpenStack compute keypair and exports outputs.
func keypair(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackKeypair.Spec
	keypairName := locals.OpenStackKeypair.Metadata.Name

	keypairArgs := &compute.KeypairArgs{
		Name: pulumi.String(keypairName),
	}

	// Set public_key if provided (import mode).
	// If not set, OpenStack generates a new keypair.
	if spec.PublicKey != "" {
		keypairArgs.PublicKey = pulumi.StringPtr(spec.PublicKey)
	}

	// Set region override if provided.
	if spec.Region != "" {
		keypairArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdKeypair, err := compute.NewKeypair(
		ctx,
		strings.ToLower(keypairName),
		keypairArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack compute keypair")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpName, createdKeypair.Name)
	ctx.Export(OpFingerprint, createdKeypair.Fingerprint)
	ctx.Export(OpPublicKey, createdKeypair.PublicKey)
	ctx.Export(OpRegion, createdKeypair.Region)

	// Export private_key as a Pulumi secret output.
	// Only populated when OpenStack generates the keypair (no public_key in spec).
	ctx.Export(OpPrivateKey, pulumi.ToSecret(createdKeypair.PrivateKey))

	return nil
}
