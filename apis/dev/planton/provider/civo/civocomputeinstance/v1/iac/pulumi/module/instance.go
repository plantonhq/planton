package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// instance provisions the Civo Compute Instance and exports stack outputs.
func instance(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.Instance, error) {

	// 1. Build Instance arguments directly from the proto spec.
	instanceArgs := &civo.InstanceArgs{
		DiskImage: pulumi.String(locals.CivoComputeInstance.Spec.Image),
		Hostname:  pulumi.String(locals.CivoComputeInstance.Metadata.Name),
		Region:    pulumi.String(locals.CivoComputeInstance.Spec.Region.String()),
		Size:      pulumi.String(locals.CivoComputeInstance.Spec.Size),
	}

	// Optional: network.
	if locals.CivoComputeInstance.Spec.Network != nil &&
		locals.CivoComputeInstance.Spec.Network.GetValue() != "" {
		instanceArgs.NetworkId = pulumi.String(
			locals.CivoComputeInstance.Spec.Network.GetValue())
	}

	// Optional: SSH keys (first one only—Civo currently allows one).
	if len(locals.CivoComputeInstance.Spec.SshKeyIds) > 0 &&
		locals.CivoComputeInstance.Spec.SshKeyIds[0] != "" {
		instanceArgs.SshkeyId = pulumi.String(locals.CivoComputeInstance.Spec.SshKeyIds[0])
	}

	// Optional: firewalls (first one only—Civo currently allows one).
	if len(locals.CivoComputeInstance.Spec.FirewallIds) > 0 &&
		locals.CivoComputeInstance.Spec.FirewallIds[0].GetValue() != "" {
		instanceArgs.FirewallId = pulumi.String(
			locals.CivoComputeInstance.Spec.FirewallIds[0].GetValue())
	}

	// Optional: static public IP.
	if locals.CivoComputeInstance.Spec.ReservedIpId != nil &&
		locals.CivoComputeInstance.Spec.ReservedIpId.GetValue() != "" {
		instanceArgs.ReservedIpv4 = pulumi.String(
			locals.CivoComputeInstance.Spec.ReservedIpId.GetValue())
	}

	// Optional: tags.
	if len(locals.CivoComputeInstance.Spec.Tags) > 0 {
		var tags pulumi.StringArray
		for _, t := range locals.CivoComputeInstance.Spec.Tags {
			tags = append(tags, pulumi.String(t))
		}
		instanceArgs.Tags = tags
	}

	// Optional: cloud‑init user data.
	if locals.CivoComputeInstance.Spec.UserData != "" {
		instanceArgs.Script = pulumi.String(
			locals.CivoComputeInstance.Spec.UserData)
	}

	// 2. Create the Instance.
	createdInstance, err := civo.NewInstance(
		ctx,
		"instance",
		instanceArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create civo compute instance")
	}

	// 3. Export stack outputs.
	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpPublicIpv4, createdInstance.PublicIp)
	ctx.Export(OpPrivateIpv4, createdInstance.PrivateIp)
	ctx.Export(OpStatus, createdInstance.Status)
	ctx.Export(OpCreatedAt, createdInstance.CreatedAt)

	return createdInstance, nil
}
