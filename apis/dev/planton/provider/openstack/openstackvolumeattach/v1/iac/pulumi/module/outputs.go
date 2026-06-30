package module

const (
	// OpId is the exported stack output containing the Terraform resource ID.
	OpId = "id"
	// OpInstanceId is the exported stack output containing the instance UUID.
	OpInstanceId = "instance_id"
	// OpVolumeId is the exported stack output containing the volume UUID.
	OpVolumeId = "volume_id"
	// OpDevice is the exported stack output containing the device path.
	// Computed by OpenStack if not explicitly specified.
	OpDevice = "device"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
