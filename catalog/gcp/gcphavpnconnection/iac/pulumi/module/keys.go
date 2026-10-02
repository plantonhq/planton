package module

import (
	"github.com/pkg/errors"
	gcphavpnconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphavpnconnection/v1alpha1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// generatedSharedSecretLength and generatedMd5KeyLength: letters and
	// digits only -- symbols are the one class a device's configuration
	// syntax can need quoted. 32 for the pre-shared key (Google accepts up
	// to 63); 24 for the MD5 key (Google accepts up to 80, and some devices
	// cap a BGP password at 25). Twin of the Terraform module's
	// random_password arguments.
	generatedSharedSecretLength = 32
	generatedMd5KeyLength       = 24
)

// connectionKeys holds the connection-level keys every tunnel and MD5
// session falls back to when it declares none of its own: the declared
// spec value, or the one key the module minted. Minted* is set exactly
// when the module generated that key, for exportOutputs.
type connectionKeys struct {
	SharedSecret               pulumi.StringOutput
	Md5AuthenticationKey       pulumi.StringOutput
	MintedSharedSecret         *pulumi.StringOutput
	MintedMd5AuthenticationKey *pulumi.StringOutput
}

// mintKeys resolves the connection-level keys, minting the pre-shared key
// and the MD5 key when the predicates in locals.go say so -- one key each,
// never one per tunnel, so a tunnel reorder never swaps keys.
func mintKeys(ctx *pulumi.Context, locals *Locals) (*connectionKeys, error) {
	spec := locals.GcpHaVpnConnection.Spec
	keys := &connectionKeys{
		// Marked secret so the value is encrypted in the Pulumi state --
		// twin of the Terraform module's sensitive handling.
		SharedSecret:         pulumi.ToSecret(pulumi.String(spec.SharedSecret.GetValue())).(pulumi.StringOutput),
		Md5AuthenticationKey: pulumi.ToSecret(pulumi.String(spec.Md5AuthenticationKey.GetValue())).(pulumi.StringOutput),
	}

	if locals.GenerateSharedSecret {
		minted, err := mintKey(ctx, "shared-secret", generatedSharedSecretLength)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate the pre-shared key")
		}
		keys.SharedSecret = minted
		keys.MintedSharedSecret = &minted
	}
	if locals.GenerateMd5AuthenticationKey {
		minted, err := mintKey(ctx, "md5-authentication-key", generatedMd5KeyLength)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate the BGP MD5 key")
		}
		keys.Md5AuthenticationKey = minted
		keys.MintedMd5AuthenticationKey = &minted
	}
	return keys, nil
}

// mintKey creates one letters-and-digits random password. The
// generation-shape arguments are ignored after creation so an IMPORTED key
// never silently regenerates (which would recreate every tunnel that uses
// it): rotation stays an explicit act, never plan fallout. Twin: the
// Terraform module's lifecycle.ignore_changes on the same argument set.
func mintKey(ctx *pulumi.Context, name string, length int) (pulumi.StringOutput, error) {
	generated, err := random.NewRandomPassword(ctx, name,
		&random.RandomPasswordArgs{
			Length:     pulumi.Int(length),
			Special:    pulumi.Bool(false),
			MinUpper:   pulumi.Int(2),
			MinLower:   pulumi.Int(2),
			MinNumeric: pulumi.Int(2),
		},
		pulumi.IgnoreChanges([]string{
			"length", "special", "upper", "lower", "numeric",
			"minLower", "minNumeric", "minSpecial", "minUpper", "overrideSpecial",
		}))
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	return generated.Result, nil
}

// tunnelSharedSecret is the key a tunnel authenticates with: its own, else
// the connection's (declared or minted).
func (k *connectionKeys) tunnelSharedSecret(tunnel *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionTunnel) pulumi.StringOutput {
	if tunnel.SharedSecret != "" {
		return pulumi.ToSecret(pulumi.String(tunnel.SharedSecret)).(pulumi.StringOutput)
	}
	return k.SharedSecret
}

// sessionMd5Key is the MD5 key of a session that declares the block: its
// own, else the connection's (declared or minted).
func (k *connectionKeys) sessionMd5Key(key *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionBgpMd5AuthenticationKey) pulumi.StringOutput {
	if key.Key != "" {
		return pulumi.ToSecret(pulumi.String(key.Key)).(pulumi.StringOutput)
	}
	return k.Md5AuthenticationKey
}
