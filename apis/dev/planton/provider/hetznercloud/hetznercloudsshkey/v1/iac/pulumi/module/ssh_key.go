package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// sshKey provisions the Hetzner Cloud SSH key and exports its ID and fingerprint.
func sshKey(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	createdSshKey, err := hcloud.NewSshKey(
		ctx,
		"ssh-key",
		&hcloud.SshKeyArgs{
			Name:      pulumi.String(locals.HetznerCloudSshKey.Metadata.Name),
			PublicKey: pulumi.String(locals.HetznerCloudSshKey.Spec.PublicKey),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud ssh key")
	}

	ctx.Export(OpSshKeyId, createdSshKey.ID())
	ctx.Export(OpFingerprint, createdSshKey.Fingerprint)

	return nil
}
