package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/filestorage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func mountTarget(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*filestorage.MountTarget, error) {
	spec := locals.OciFileSystem.Spec
	mt := spec.MountTarget

	mtArgs := &filestorage.MountTargetArgs{
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		AvailabilityDomain: pulumi.String(spec.AvailabilityDomain),
		SubnetId:           pulumi.String(mt.SubnetId.GetValue()),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
	}

	if mt.DisplayName != "" {
		mtArgs.DisplayName = pulumi.StringPtr(mt.DisplayName)
	}

	if mt.HostnameLabel != "" {
		mtArgs.HostnameLabel = pulumi.StringPtr(mt.HostnameLabel)
	}

	if mt.IpAddress != "" {
		mtArgs.IpAddress = pulumi.StringPtr(mt.IpAddress)
	}

	if len(mt.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(mt.NsgIds))
		for i, nsg := range mt.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		mtArgs.NsgIds = nsgIds
	}

	if mt.RequestedThroughput > 0 {
		mtArgs.RequestedThroughput = pulumi.StringPtr(fmt.Sprintf("%d", mt.RequestedThroughput))
	}

	resourceName := fmt.Sprintf("%s-mount-target", locals.DisplayName)
	createdMt, err := filestorage.NewMountTarget(ctx, resourceName, mtArgs, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mount target")
	}

	ctx.Export(OpMountTargetId, createdMt.ID())
	ctx.Export(OpMountTargetIpAddress, createdMt.IpAddress)
	ctx.Export(OpExportSetId, createdMt.ExportSetId)

	if mt.MaxFsStatBytes > 0 || mt.MaxFsStatFiles > 0 {
		if err := exportSet(ctx, locals, provider, createdMt); err != nil {
			return nil, errors.Wrap(err, "failed to configure export set")
		}
	}

	return createdMt, nil
}

func exportSet(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, mt *filestorage.MountTarget) error {
	spec := locals.OciFileSystem.Spec.MountTarget

	esArgs := &filestorage.ExportSetArgs{
		MountTargetId: mt.ID(),
	}

	if spec.MaxFsStatBytes > 0 {
		esArgs.MaxFsStatBytes = pulumi.StringPtr(fmt.Sprintf("%d", spec.MaxFsStatBytes))
	}

	if spec.MaxFsStatFiles > 0 {
		esArgs.MaxFsStatFiles = pulumi.StringPtr(fmt.Sprintf("%d", spec.MaxFsStatFiles))
	}

	resourceName := fmt.Sprintf("%s-export-set", locals.DisplayName)
	_, err := filestorage.NewExportSet(ctx, resourceName, esArgs,
		pulumiOciOpt(provider),
		pulumi.DependsOn([]pulumi.Resource{mt}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to configure export set")
	}

	return nil
}
