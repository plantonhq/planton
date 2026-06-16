package module

import (
	"fmt"

	"github.com/pkg/errors"
	awss3objectsetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awss3objectset/v1"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awss3objectsetv1.AwsS3ObjectSetStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AwsS3ObjectSet.Spec

	// Create explicit AWS provider from stack input credentials
	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsS3ObjectSet.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Resolve the target bucket name from the foreign key
	bucketName := ""
	if spec.Bucket != nil {
		switch v := spec.Bucket.LiteralOrRef.(type) {
		case *foreignkeyv1.StringValueOrRef_Value:
			bucketName = v.Value
		}
	}
	if bucketName == "" {
		return errors.New("bucket name must be resolved before invoking IaC module")
	}

	// Track outputs
	etagMap := pulumi.StringMap{}
	versionIdMap := pulumi.StringMap{}

	// Create S3 objects
	for i, obj := range spec.Objects {
		if obj.Key == "" {
			return fmt.Errorf("object at index %d has an empty key", i)
		}

		resourceName := fmt.Sprintf("object-%d", i)

		// Prepare merged tags: set-level tags + object-level tags (object takes precedence)
		objTags := pulumi.StringMap{}
		for k, v := range locals.Labels {
			objTags[k] = pulumi.String(v)
		}
		for k, v := range spec.Tags {
			objTags[k] = pulumi.String(v)
		}
		for k, v := range obj.Tags {
			objTags[k] = pulumi.String(v)
		}

		args := &s3.BucketObjectv2Args{
			Bucket: pulumi.String(bucketName),
			Key:    pulumi.String(obj.Key),
			Tags:   objTags,
		}

		// Set content source
		switch source := obj.Source.(type) {
		case *awss3objectsetv1.AwsS3Object_Content:
			args.Content = pulumi.StringPtr(source.Content)
		case *awss3objectsetv1.AwsS3Object_ContentBase64:
			args.ContentBase64 = pulumi.StringPtr(source.ContentBase64)
		default:
			return fmt.Errorf("object %q has no content source", obj.Key)
		}

		// content_type is always populated by OpenMCF middleware (default: application/octet-stream)
		args.ContentType = pulumi.StringPtr(obj.GetContentType())
		if obj.CacheControl != "" {
			args.CacheControl = pulumi.StringPtr(obj.CacheControl)
		}
		if obj.ContentEncoding != "" {
			args.ContentEncoding = pulumi.StringPtr(obj.ContentEncoding)
		}
		if obj.Acl != "" {
			args.Acl = pulumi.StringPtr(obj.Acl)
		}

		s3Object, err := s3.NewBucketObjectv2(ctx, resourceName, args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create S3 object %q", obj.Key)
		}

		etagMap[obj.Key] = s3Object.Etag
		versionIdMap[obj.Key] = s3Object.VersionId
	}

	// Export outputs
	ctx.Export(OpObjectEtags, etagMap)
	ctx.Export(OpObjectVersionIds, versionIdMap)

	return nil
}
