package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ipAddress provisions a Civo reserved (static) IP and exports its identifiers.
func ipAddress(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.ReservedIp, error) {

	// NOTE: The Civo Terraform provider exposes resource `civo_reserved_ip`.
	// In Pulumi, this maps to `civo.ReservedIp`.  Only the minimal required
	// arguments (region and name/label) are used, mirroring the proto spec.

	// 1. Build args.
	args := &civo.ReservedIpArgs{
		// Region is mandatory; proto enum → string.
		Region: pulumi.String(locals.CivoIpAddress.Spec.Region.String()),
	}

	// Optional description maps to Name in the provider (acts like a label).
	if locals.CivoIpAddress.Spec.Description != "" {
		args.Name = pulumi.StringPtr(locals.CivoIpAddress.Spec.Description)
	}

	// 2. Create resource with provider.
	createdReservedIp, err := civo.NewReservedIp(
		ctx,
		"reservedIp",
		args,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reserved IP")
	}

	// 3. Export outputs expected by stack‑output proto.
	ctx.Export(OpReservedIpId, createdReservedIp.ID())
	ctx.Export(OpReservedIpAddress, createdReservedIp.Ip)

	return createdReservedIp, nil
}
