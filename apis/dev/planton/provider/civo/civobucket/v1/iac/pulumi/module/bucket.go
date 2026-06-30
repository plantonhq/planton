package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// bucket provisions the Civo Object Store bucket and exports its ID, endpoint & keys.
func bucket(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.ObjectStore, error) {

	// 1. Create an Object‑Store credential (no explicit keys → Civo generates them).
	createdCredential, err := civo.NewObjectStoreCredential(
		ctx,
		"bucket-creds",
		&civo.ObjectStoreCredentialArgs{},
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create object‑store credential")
	}

	// 2. Build resource arguments directly from proto fields.
	bucketArgs := &civo.ObjectStoreArgs{
		Name:        pulumi.String(locals.CivoBucket.Spec.BucketName),
		Region:      pulumi.String(locals.CivoBucket.Spec.Region.String()),
		AccessKeyId: createdCredential.AccessKeyId,
	}

	// 3. Create the bucket.
	createdBucket, err := civo.NewObjectStore(
		ctx,
		"bucket",
		bucketArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create civo object‑store bucket")
	}

	// 4. Handle versioning configuration.
	// Note: Civo Object Storage versioning is configured via the S3 API, not the control plane.
	// The Civo Pulumi provider doesn't expose versioning as a direct attribute.
	// Users need to configure versioning post-deployment using AWS S3 SDK/CLI pointed at the Civo endpoint.
	if locals.CivoBucket.Spec.VersioningEnabled {
		ctx.Log.Info(fmt.Sprintf(
			"Versioning requested for bucket '%s'. "+
				"Note: Civo Object Storage versioning must be configured via S3 API after bucket creation. "+
				"Use AWS CLI or SDK with the Civo endpoint to enable versioning.",
			locals.CivoBucket.Spec.BucketName,
		), nil)
	}

	// 5. Handle tags.
	// Note: The Civo provider's ObjectStore resource doesn't currently support tags.
	// Tags in the spec are available for logical organization but aren't applied to the Civo resource.
	if len(locals.CivoBucket.Spec.Tags) > 0 {
		ctx.Log.Info(fmt.Sprintf(
			"Tags specified for bucket '%s': %v. "+
				"Note: Civo ObjectStore provider doesn't currently support tags. "+
				"Tags are recorded in metadata but not applied to the Civo resource.",
			locals.CivoBucket.Spec.BucketName,
			locals.CivoBucket.Spec.Tags,
		), nil)
	}

	// 6. Export stack outputs.
	ctx.Export(OpBucketId, createdBucket.ID())
	ctx.Export(OpEndpointUrl, createdBucket.BucketUrl)
	ctx.Export(OpAccessKeySecretRef, createdCredential.AccessKeyId)
	ctx.Export(OpSecretKeySecretRef, createdCredential.SecretAccessKey)

	return createdBucket, nil
}
