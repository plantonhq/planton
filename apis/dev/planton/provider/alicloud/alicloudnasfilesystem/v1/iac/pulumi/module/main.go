package module

import (
	"github.com/pkg/errors"
	alicloudnasfilesystemv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudnasfilesystem/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/nas"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudnasfilesystemv1.AliCloudNasFileSystemStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudNasFileSystem.Spec
	resourceName := locals.AliCloudNasFileSystem.Metadata.Name

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	fsType := fileSystemType(spec)

	fsArgs := &nas.FileSystemArgs{
		ProtocolType:   pulumi.String(spec.ProtocolType),
		StorageType:    pulumi.String(spec.StorageType),
		FileSystemType: pulumi.String(fsType),
		Description:    optionalString(spec.Description),
		Tags:           pulumi.ToStringMap(locals.Tags),
	}

	if spec.ResourceGroupId != "" {
		fsArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	if spec.Encryption != nil {
		fsArgs.EncryptType = pulumi.Int(int(spec.Encryption.EncryptType))
		if spec.Encryption.KmsKeyId != "" {
			fsArgs.KmsKeyId = pulumi.String(spec.Encryption.KmsKeyId)
		}
	}

	if spec.Capacity > 0 {
		fsArgs.Capacity = pulumi.Int(int(spec.Capacity))
	}

	if spec.ZoneId != "" {
		fsArgs.ZoneId = pulumi.String(spec.ZoneId)
	}

	// Extreme NAS requires VPC and VSwitch at the file system level.
	if fsType == "extreme" {
		fsArgs.VpcId = pulumi.String(spec.VpcId.GetValue())
		fsArgs.VswitchId = pulumi.String(spec.VswitchId.GetValue())
	}

	fs, err := nas.NewFileSystem(ctx, resourceName, fsArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create NAS file system %s", resourceName)
	}

	// Create a custom access group if access rules are specified,
	// otherwise the mount target uses the default VPC group.
	var accessGroupName pulumi.StringInput
	if len(spec.AccessRules) > 0 {
		agName, err := accessGroup(ctx, alicloudProvider, resourceName, fsType, spec.AccessRules)
		if err != nil {
			return err
		}
		accessGroupName = agName
	}

	mtArgs := &nas.MountTargetArgs{
		FileSystemId: fs.ID(),
		VpcId:        pulumi.String(spec.VpcId.GetValue()),
		VswitchId:    pulumi.String(spec.VswitchId.GetValue()),
	}

	if accessGroupName != nil {
		mtArgs.AccessGroupName = accessGroupName
	}

	mt, err := nas.NewMountTarget(ctx, resourceName+"-mt", mtArgs,
		pulumi.Provider(alicloudProvider),
		pulumi.Parent(fs),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create mount target for NAS file system %s", resourceName)
	}

	ctx.Export(OpFileSystemId, fs.ID())
	ctx.Export(OpMountTargetDomain, mt.MountTargetDomain)

	return nil
}
