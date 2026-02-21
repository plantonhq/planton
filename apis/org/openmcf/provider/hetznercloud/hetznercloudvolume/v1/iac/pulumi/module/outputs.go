package module

const (
	// OpVolumeId is the exported stack output name that contains the
	// Hetzner Cloud numeric ID of the created volume.
	OpVolumeId = "volume_id"

	// OpLinuxDevice is the exported stack output name that contains the
	// Linux device path for the volume on the attached server
	// (e.g., "/dev/disk/by-id/scsi-0HC_Volume_12345678").
	OpLinuxDevice = "linux_device"
)
