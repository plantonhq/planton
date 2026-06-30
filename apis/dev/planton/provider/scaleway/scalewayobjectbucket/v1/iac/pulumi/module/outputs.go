package module

const (
	// OpBucketId is the exported stack output name for the bucket's
	// unique identifier (format: "region/bucket-name").
	OpBucketId = "bucket_id"

	// OpEndpoint is the exported stack output name for the bucket's
	// FQDN endpoint URL (e.g., "bucket-name.s3.fr-par.scw.cloud").
	OpEndpoint = "endpoint"

	// OpApiEndpoint is the exported stack output name for the S3 API
	// endpoint URL (e.g., "https://s3.fr-par.scw.cloud").
	OpApiEndpoint = "api_endpoint"

	// OpBucketName is the exported stack output name for the bucket's
	// name as it exists in Scaleway Object Storage.
	OpBucketName = "bucket_name"

	// OpRegion is the exported stack output name for the region where
	// the bucket is deployed.
	OpRegion = "region"
)
