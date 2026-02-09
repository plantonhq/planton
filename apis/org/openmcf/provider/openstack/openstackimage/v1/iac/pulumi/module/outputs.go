package module

const (
	// OpImageId is the exported stack output containing the Glance image UUID.
	OpImageId = "image_id"
	// OpName is the exported stack output containing the image name.
	OpName = "name"
	// OpChecksum is the exported stack output containing the MD5 checksum.
	OpChecksum = "checksum"
	// OpSizeBytes is the exported stack output containing the image size in bytes.
	OpSizeBytes = "size_bytes"
	// OpStatus is the exported stack output containing the image lifecycle status.
	OpStatus = "status"
	// OpFile is the exported stack output containing the URL path to image data.
	OpFile = "file"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
