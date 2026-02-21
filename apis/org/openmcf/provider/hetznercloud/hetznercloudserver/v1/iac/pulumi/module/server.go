package module

import (
	"strconv"

	"github.com/pkg/errors"
	hetznercloudserverv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudserver/v1"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func server(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudServer.Spec

	serverArgs := &hcloud.ServerArgs{
		Name:                   pulumi.String(locals.HetznerCloudServer.Metadata.Name),
		ServerType:             pulumi.String(spec.ServerType),
		Image:                  pulumi.StringPtr(spec.Image),
		Location:               pulumi.StringPtr(spec.Location),
		Labels:                 pulumi.ToStringMap(locals.Labels),
		Backups:                pulumi.BoolPtr(spec.Backups),
		KeepDisk:               pulumi.BoolPtr(spec.KeepDisk),
		DeleteProtection:       pulumi.BoolPtr(spec.DeleteProtection),
		RebuildProtection:      pulumi.BoolPtr(spec.RebuildProtection),
		ShutdownBeforeDeletion: pulumi.BoolPtr(spec.ShutdownBeforeDeletion),
	}

	if spec.UserData != "" {
		serverArgs.UserData = pulumi.StringPtr(spec.UserData)
	}

	// SSH keys: StringValueOrRef[] -> string[] (provider accepts names or IDs as strings)
	if len(spec.SshKeys) > 0 {
		sshKeys := make([]string, 0, len(spec.SshKeys))
		for _, ref := range spec.SshKeys {
			sshKeys = append(sshKeys, ref.GetValue())
		}
		serverArgs.SshKeys = pulumi.ToStringArray(sshKeys)
	}

	// Placement group: StringValueOrRef -> int
	if spec.PlacementGroupId != nil && spec.PlacementGroupId.GetValue() != "" {
		pgId, err := strconv.Atoi(spec.PlacementGroupId.GetValue())
		if err != nil {
			return errors.Wrapf(err, "failed to parse placement_group_id %q as integer",
				spec.PlacementGroupId.GetValue())
		}
		serverArgs.PlacementGroupId = pulumi.IntPtr(pgId)
	}

	// Firewall IDs: StringValueOrRef[] -> int[]
	if len(spec.FirewallIds) > 0 {
		fwIds := make([]int, 0, len(spec.FirewallIds))
		for _, ref := range spec.FirewallIds {
			id, err := strconv.Atoi(ref.GetValue())
			if err != nil {
				return errors.Wrapf(err, "failed to parse firewall_id %q as integer",
					ref.GetValue())
			}
			fwIds = append(fwIds, id)
		}
		serverArgs.FirewallIds = pulumi.IntArray(toIntInputArray(fwIds))
	}

	// Public network configuration: only set when explicitly specified to
	// preserve the provider default (auto-assigned IPv4 + IPv6).
	if spec.PublicNet != nil {
		publicNet, err := buildPublicNet(spec.PublicNet)
		if err != nil {
			return errors.Wrap(err, "failed to build public_net configuration")
		}
		serverArgs.PublicNets = hcloud.ServerPublicNetArray{publicNet}
	}

	// Private network attachments
	if len(spec.Networks) > 0 {
		networkArgs, err := buildNetworkAttachments(spec.Networks)
		if err != nil {
			return errors.Wrap(err, "failed to build network attachments")
		}
		serverArgs.Networks = networkArgs
	}

	createdServer, err := hcloud.NewServer(
		ctx,
		"server",
		serverArgs,
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud server")
	}

	// Optional reverse DNS for the server's auto-assigned IPv4
	if spec.DnsPtr != "" {
		serverIdInt := createdServer.ID().ApplyT(func(id pulumi.ID) (int, error) {
			return strconv.Atoi(string(id))
		}).(pulumi.IntOutput)

		if _, err := hcloud.NewRdns(
			ctx,
			"rdns",
			&hcloud.RdnsArgs{
				ServerId:  serverIdInt,
				IpAddress: createdServer.Ipv4Address,
				DnsPtr:    pulumi.String(spec.DnsPtr),
			},
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create reverse dns record")
		}
	}

	ctx.Export(OpServerId, createdServer.ID())
	ctx.Export(OpIpv4Address, createdServer.Ipv4Address)
	ctx.Export(OpIpv6Address, createdServer.Ipv6Address)
	ctx.Export(OpStatus, createdServer.Status)

	return nil
}

func buildPublicNet(
	pn *hetznercloudserverv1.HetznerCloudServerSpec_PublicNet,
) (*hcloud.ServerPublicNetArgs, error) {
	args := &hcloud.ServerPublicNetArgs{}

	// optional bool with default true: nil -> true, non-nil -> use value
	if pn.Ipv4Enabled != nil {
		args.Ipv4Enabled = pulumi.BoolPtr(*pn.Ipv4Enabled)
	} else {
		args.Ipv4Enabled = pulumi.BoolPtr(true)
	}

	if pn.Ipv6Enabled != nil {
		args.Ipv6Enabled = pulumi.BoolPtr(*pn.Ipv6Enabled)
	} else {
		args.Ipv6Enabled = pulumi.BoolPtr(true)
	}

	// Primary IP references: StringValueOrRef -> int
	if pn.Ipv4 != nil && pn.Ipv4.GetValue() != "" {
		id, err := strconv.Atoi(pn.Ipv4.GetValue())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse public_net.ipv4 %q as integer",
				pn.Ipv4.GetValue())
		}
		args.Ipv4 = pulumi.IntPtr(id)
	}

	if pn.Ipv6 != nil && pn.Ipv6.GetValue() != "" {
		id, err := strconv.Atoi(pn.Ipv6.GetValue())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse public_net.ipv6 %q as integer",
				pn.Ipv6.GetValue())
		}
		args.Ipv6 = pulumi.IntPtr(id)
	}

	return args, nil
}

func buildNetworkAttachments(
	networks []*hetznercloudserverv1.HetznerCloudServerSpec_NetworkAttachment,
) (hcloud.ServerNetworkTypeArray, error) {
	result := make(hcloud.ServerNetworkTypeArray, 0, len(networks))

	for _, net := range networks {
		networkId, err := strconv.Atoi(net.NetworkId.GetValue())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse network_id %q as integer",
				net.NetworkId.GetValue())
		}

		args := hcloud.ServerNetworkTypeArgs{
			NetworkId: pulumi.Int(networkId),
			// Always pass AliasIps to avoid the Terraform bridge bug (#650)
			// that causes network detach/reattach on every apply.
			AliasIps: pulumi.ToStringArray(net.AliasIps),
		}

		if net.Ip != "" {
			args.Ip = pulumi.StringPtr(net.Ip)
		}

		result = append(result, args)
	}

	return result, nil
}

func toIntInputArray(ints []int) []pulumi.IntInput {
	result := make([]pulumi.IntInput, len(ints))
	for i, v := range ints {
		result[i] = pulumi.Int(v)
	}
	return result
}
