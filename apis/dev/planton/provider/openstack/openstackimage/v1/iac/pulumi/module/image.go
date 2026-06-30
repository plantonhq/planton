package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/images"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// image provisions the OpenStack Glance image and exports outputs.
func image(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackImage.Spec
	resourceName := locals.OpenStackImage.Metadata.Name

	imageArgs := &images.ImageArgs{
		Name:            pulumi.String(resourceName),
		ContainerFormat: pulumi.String(spec.ContainerFormat),
		DiskFormat:      pulumi.String(spec.DiskFormat),
	}

	// Set image source URL if provided.
	if spec.ImageSourceUrl != "" {
		imageArgs.ImageSourceUrl = pulumi.StringPtr(spec.ImageSourceUrl)
	}

	// Set minimum disk and RAM requirements.
	if spec.MinDiskGb > 0 {
		imageArgs.MinDiskGb = pulumi.IntPtr(int(spec.MinDiskGb))
	}
	if spec.MinRamMb > 0 {
		imageArgs.MinRamMb = pulumi.IntPtr(int(spec.MinRamMb))
	}

	// Set protected flag. Middleware guarantees the default (false) is applied.
	imageArgs.Protected = pulumi.BoolPtr(spec.GetProtected())

	// Set hidden flag. Middleware guarantees the default (false) is applied.
	imageArgs.Hidden = pulumi.BoolPtr(spec.GetHidden())

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := pulumi.StringArray{}
		for _, tag := range spec.Tags {
			tags = append(tags, pulumi.String(tag))
		}
		imageArgs.Tags = tags
	}

	// Set visibility. Middleware guarantees the default ("private") is applied.
	imageArgs.Visibility = pulumi.StringPtr(spec.GetVisibility())

	// Set region override if provided.
	if spec.Region != "" {
		imageArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdImage, err := images.NewImage(
		ctx,
		strings.ToLower(resourceName),
		imageArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack glance image")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpImageId, createdImage.ID())
	ctx.Export(OpName, createdImage.Name)
	ctx.Export(OpChecksum, createdImage.Checksum)
	ctx.Export(OpSizeBytes, createdImage.SizeBytes)
	ctx.Export(OpStatus, createdImage.Status)
	ctx.Export(OpFile, createdImage.File)
	ctx.Export(OpRegion, createdImage.Region)

	return nil
}
