package module

const (
	// OpVolumeId is the exported stack output name for the volume's
	// zoned identifier (format: "{zone}/{uuid}").
	OpVolumeId = "volume_id"

	// OpVolumeName is the exported stack output name for the volume's
	// name as it exists in Scaleway Block Storage.
	OpVolumeName = "volume_name"

	// OpZone is the exported stack output name for the Availability Zone
	// where the volume is deployed. Used by downstream resources to
	// verify zone co-location.
	OpZone = "zone"
)
