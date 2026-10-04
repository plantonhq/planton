package verify

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	fsxtypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	pkgerrors "github.com/pkg/errors"
)

// The FSx family verifies through the fsx control plane. All kinds are
// lifecycle-state-aware (the NAT-gateway class): a deleted resource can stay
// describable in the DELETING state for a while before the typed NotFound
// appears, so that state counts as absent rather than racing the API's
// cleanup. One file-system verifier serves the Lustre, OpenZFS, and Windows
// kinds — DescribeFileSystems is type-agnostic and every kind exports the
// same file_system_id output.

// fsxFileSystemVerifier verifies an FSx file system (any type) via
// DescribeFileSystems, keyed on the file_system_id output.
type fsxFileSystemVerifier struct {
	// component names the kind in error messages so a chain failure reads
	// unambiguously.
	kind string
}

func (v *fsxFileSystemVerifier) IDOutputKey() string { return "file_system_id" }

func (v *fsxFileSystemVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxFileSystemExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-exists failed for %q", v.kind, id)
	}
	if !exists {
		return pkgerrors.Errorf("%s %q not found after deploy", v.kind, id)
	}
	return nil
}

func (v *fsxFileSystemVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxFileSystemExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-absent failed for %q", v.kind, id)
	}
	if exists {
		return pkgerrors.Errorf("%s %q still exists after destroy", v.kind, id)
	}
	return nil
}

func fsxFileSystemExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := fsxClient(cfg, region)
	out, err := client.DescribeFileSystems(ctx, &fsx.DescribeFileSystemsInput{
		FileSystemIds: []string{id},
	})
	if err != nil {
		var notFound *fsxtypes.FileSystemNotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	for _, fileSystem := range out.FileSystems {
		if !fsxLifecycleGone(string(fileSystem.Lifecycle)) {
			return true, nil
		}
	}
	return false, nil
}

// fsxStorageVirtualMachineVerifier verifies an
// AwsFsxOntapStorageVirtualMachine via DescribeStorageVirtualMachines, keyed
// on the svm_id output.
type fsxStorageVirtualMachineVerifier struct{}

func (*fsxStorageVirtualMachineVerifier) IDOutputKey() string { return "svm_id" }

func (*fsxStorageVirtualMachineVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxStorageVirtualMachineExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsfsxontapstoragevirtualmachine verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsfsxontapstoragevirtualmachine %q not found after deploy", id)
	}
	return nil
}

func (*fsxStorageVirtualMachineVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxStorageVirtualMachineExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsfsxontapstoragevirtualmachine verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsfsxontapstoragevirtualmachine %q still exists after destroy", id)
	}
	return nil
}

func fsxStorageVirtualMachineExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := fsxClient(cfg, region)
	out, err := client.DescribeStorageVirtualMachines(ctx, &fsx.DescribeStorageVirtualMachinesInput{
		StorageVirtualMachineIds: []string{id},
	})
	if err != nil {
		var notFound *fsxtypes.StorageVirtualMachineNotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	for _, svm := range out.StorageVirtualMachines {
		if !fsxLifecycleGone(string(svm.Lifecycle)) {
			return true, nil
		}
	}
	return false, nil
}

// fsxVolumeVerifier verifies an AwsFsxOntapVolume via DescribeVolumes, keyed
// on the volume_id output.
type fsxVolumeVerifier struct{}

func (*fsxVolumeVerifier) IDOutputKey() string { return "volume_id" }

func (*fsxVolumeVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxVolumeExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsfsxontapvolume verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsfsxontapvolume %q not found after deploy", id)
	}
	return nil
}

func (*fsxVolumeVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxVolumeExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsfsxontapvolume verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsfsxontapvolume %q still exists after destroy", id)
	}
	return nil
}

func fsxVolumeExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := fsxClient(cfg, region)
	out, err := client.DescribeVolumes(ctx, &fsx.DescribeVolumesInput{
		VolumeIds: []string{id},
	})
	if err != nil {
		var notFound *fsxtypes.VolumeNotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	for _, vol := range out.Volumes {
		if !fsxLifecycleGone(string(vol.Lifecycle)) {
			return true, nil
		}
	}
	return false, nil
}

// fsxDataRepositoryAssociationVerifier verifies an
// AwsFsxDataRepositoryAssociation via DescribeDataRepositoryAssociations,
// keyed on the association_id output.
type fsxDataRepositoryAssociationVerifier struct{}

func (*fsxDataRepositoryAssociationVerifier) IDOutputKey() string { return "association_id" }

func (*fsxDataRepositoryAssociationVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxDataRepositoryAssociationExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsfsxdatarepositoryassociation verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsfsxdatarepositoryassociation %q not found after deploy", id)
	}
	return nil
}

func (*fsxDataRepositoryAssociationVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := fsxDataRepositoryAssociationExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsfsxdatarepositoryassociation verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsfsxdatarepositoryassociation %q still exists after destroy", id)
	}
	return nil
}

func fsxDataRepositoryAssociationExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := fsxClient(cfg, region)
	out, err := client.DescribeDataRepositoryAssociations(ctx, &fsx.DescribeDataRepositoryAssociationsInput{
		AssociationIds: []string{id},
	})
	if err != nil {
		var notFound *fsxtypes.DataRepositoryAssociationNotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	for _, association := range out.Associations {
		if !fsxLifecycleGone(string(association.Lifecycle)) {
			return true, nil
		}
	}
	return false, nil
}

func fsxClient(cfg aws.Config, region string) *fsx.Client {
	return fsx.NewFromConfig(cfg, func(o *fsx.Options) {
		if region != "" {
			o.Region = region
		}
	})
}

// fsxLifecycleGone treats the teardown-side lifecycle state as absent
// (case-insensitive by convention, though the FSx API reports uppercase).
func fsxLifecycleGone(state string) bool {
	return strings.EqualFold(state, "DELETING")
}
