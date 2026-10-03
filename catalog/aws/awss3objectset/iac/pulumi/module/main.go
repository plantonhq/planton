package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	awss3objectsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3objectset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awss3objectsetv1alpha1.AwsS3ObjectSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AwsS3ObjectSet.Spec

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// The bucket foreign key is resolved to a literal name before the module
	// runs (a referenced AwsS3Bucket resolves to status.outputs.bucket_id).
	bucketName := spec.Bucket.GetValue()
	if bucketName == "" {
		return errors.New("bucket name must be resolved before invoking IaC module")
	}

	arnMap := pulumi.StringMap{}
	etagMap := pulumi.StringMap{}
	versionIdMap := pulumi.StringMap{}

	// Content-sourced objects render first: a copy may name another object of
	// this same set as its source, and the copy source string carries no
	// dependency the graph can see, so every copy waits on the whole content
	// pass (mirroring the Terraform module's depends_on).
	var contentObjects []pulumi.Resource

	for _, obj := range spec.Objects {
		if obj.Key == "" {
			return errors.New("every object must have a non-empty key")
		}
		if obj.GetCopyFrom() != nil {
			continue
		}

		args := &s3.BucketObjectv2Args{
			Bucket: pulumi.String(bucketName),
			Key:    pulumi.String(obj.Key),
			Tags:   objectTags(locals, spec, obj),
		}

		// Content source: the spec guarantees exactly one arm of the oneof is
		// set (proto oneof + CEL).
		switch source := obj.Source.(type) {
		case *awss3objectsetv1alpha1.AwsS3Object_Content:
			args.Content = pulumi.StringPtr(source.Content)
		case *awss3objectsetv1alpha1.AwsS3Object_ContentBase64:
			args.ContentBase64 = pulumi.StringPtr(source.ContentBase64)
		default:
			return fmt.Errorf("object %q has no content source", obj.Key)
		}

		// Sent explicitly, resolving the spec default (application/octet-stream)
		// here rather than leaving it to the provider: an omitted Content-Type
		// would make S3 store its own default (binary/octet-stream), diverging
		// from what the manifest declares and from the Terraform engine.
		args.ContentType = pulumi.StringPtr(obj.GetContentType())

		// HTTP presentation headers — sent only when set so unset stays
		// indistinguishable from the AWS defaults (no phantom headers stored).
		if obj.CacheControl != "" {
			args.CacheControl = pulumi.StringPtr(obj.CacheControl)
		}
		if obj.ContentEncoding != "" {
			args.ContentEncoding = pulumi.StringPtr(obj.ContentEncoding)
		}
		if obj.ContentDisposition != "" {
			args.ContentDisposition = pulumi.StringPtr(obj.ContentDisposition)
		}
		if obj.ContentLanguage != "" {
			args.ContentLanguage = pulumi.StringPtr(obj.ContentLanguage)
		}

		// User metadata (x-amz-meta-*). Keys are spec-validated lowercase,
		// matching what S3 stores, so reads never drift from the manifest.
		if len(obj.Metadata) > 0 {
			metadataMap := pulumi.StringMap{}
			for k, v := range obj.Metadata {
				metadataMap[k] = pulumi.String(v)
			}
			args.Metadata = metadataMap
		}

		// Website redirect: inert unless the bucket has static website hosting.
		if obj.WebsiteRedirect != "" {
			args.WebsiteRedirect = pulumi.StringPtr(obj.WebsiteRedirect)
		}

		applyPlacement(obj,
			&args.StorageClass, &args.ServerSideEncryption, &args.KmsKeyId,
			&args.BucketKeyEnabled, &args.ChecksumAlgorithm,
			&args.ObjectLockMode, &args.ObjectLockRetainUntilDate, &args.ObjectLockLegalHoldStatus,
			&args.Acl, &args.ForceDestroy)

		// The resource is named by the S3 object key, so adding, removing, or
		// REORDERING entries in the manifest never churns unrelated objects —
		// each object's Pulumi identity is its key, mirroring both its S3
		// identity and the Terraform module's for_each keying.
		s3Object, err := s3.NewBucketObjectv2(ctx, obj.Key, args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create S3 object %q", obj.Key)
		}
		contentObjects = append(contentObjects, s3Object)

		arnMap[obj.Key] = s3Object.Arn
		etagMap[obj.Key] = s3Object.Etag
		versionIdMap[obj.Key] = s3Object.VersionId
	}

	for _, obj := range spec.Objects {
		copyFrom := obj.GetCopyFrom()
		if copyFrom == nil {
			continue
		}

		sourceBucket := copyFrom.SourceBucket.GetValue()
		if sourceBucket == "" {
			return fmt.Errorf("object %q: copy_from.source_bucket must be resolved before invoking IaC module", obj.Key)
		}

		// "bucket/key", or "arn.../object/key" for access-point sources (the
		// provider's two documented source formats). The provider URL-escapes
		// the whole string, which is also why a ?versionId= selector cannot
		// work at this pin — copies take the source's current version.
		source := sourceBucket + "/" + copyFrom.SourceKey
		if strings.HasPrefix(sourceBucket, "arn:") {
			source = sourceBucket + "/object/" + copyFrom.SourceKey
		}

		args := &s3.ObjectCopyArgs{
			Bucket: pulumi.String(bucketName),
			Key:    pulumi.String(obj.Key),
			Source: pulumi.String(source),
			// Identity tags always reach the copy: the tag-set is sent with
			// the REPLACE tagging directive because the provider never derives
			// the directive itself — tags sent under the AWS-default COPY
			// directive are silently ignored.
			TaggingDirective: pulumi.String("REPLACE"),
			Tags:             objectTags(locals, spec, obj),
		}

		// Metadata directive: REPLACE writes this entry's metadata/headers to
		// the copy; the omitted default (COPY) preserves everything the source
		// object carried. AWS silently ignores header values sent under COPY,
		// so headers are gated here exactly as the spec CEL gates them —
		// content_type included, because the manifest loader materializes its
		// default on every object and an ungated send would echo-drift against
		// the source's real content type.
		if copyFrom.ReplaceMetadata {
			args.MetadataDirective = pulumi.String("REPLACE")
			args.ContentType = pulumi.StringPtr(obj.GetContentType())
			if obj.CacheControl != "" {
				args.CacheControl = pulumi.StringPtr(obj.CacheControl)
			}
			if obj.ContentEncoding != "" {
				args.ContentEncoding = pulumi.StringPtr(obj.ContentEncoding)
			}
			if obj.ContentDisposition != "" {
				args.ContentDisposition = pulumi.StringPtr(obj.ContentDisposition)
			}
			if obj.ContentLanguage != "" {
				args.ContentLanguage = pulumi.StringPtr(obj.ContentLanguage)
			}
			if len(obj.Metadata) > 0 {
				metadataMap := pulumi.StringMap{}
				for k, v := range obj.Metadata {
					metadataMap[k] = pulumi.String(v)
				}
				args.Metadata = metadataMap
			}
			if obj.WebsiteRedirect != "" {
				args.WebsiteRedirect = pulumi.StringPtr(obj.WebsiteRedirect)
			}
		}

		// Copy-time preconditions, evaluated by S3 against the SOURCE object;
		// a failed precondition fails the deploy (the promotion guard).
		if copyFrom.CopyIfMatch != "" {
			args.CopyIfMatch = pulumi.StringPtr(copyFrom.CopyIfMatch)
		}
		if copyFrom.CopyIfNoneMatch != "" {
			args.CopyIfNoneMatch = pulumi.StringPtr(copyFrom.CopyIfNoneMatch)
		}
		if copyFrom.CopyIfModifiedSince != "" {
			args.CopyIfModifiedSince = pulumi.StringPtr(copyFrom.CopyIfModifiedSince)
		}
		if copyFrom.CopyIfUnmodifiedSince != "" {
			args.CopyIfUnmodifiedSince = pulumi.StringPtr(copyFrom.CopyIfUnmodifiedSince)
		}

		// The Expires header on the copied object (only the copy resource
		// offers it).
		if copyFrom.Expires != "" {
			args.Expires = pulumi.StringPtr(copyFrom.Expires)
		}

		// Requester Pays acknowledgment for the SOURCE bucket.
		if copyFrom.RequestPayer != "" {
			args.RequestPayer = pulumi.StringPtr(copyFrom.RequestPayer)
		}

		applyPlacement(obj,
			&args.StorageClass, &args.ServerSideEncryption, &args.KmsKeyId,
			&args.BucketKeyEnabled, &args.ChecksumAlgorithm,
			&args.ObjectLockMode, &args.ObjectLockRetainUntilDate, &args.ObjectLockLegalHoldStatus,
			&args.Acl, &args.ForceDestroy)

		objectCopy, err := s3.NewObjectCopy(ctx, obj.Key, args,
			pulumi.Provider(provider), pulumi.DependsOn(contentObjects))
		if err != nil {
			return errors.Wrapf(err, "failed to create S3 object copy %q", obj.Key)
		}

		arnMap[obj.Key] = objectCopy.Arn
		etagMap[obj.Key] = objectCopy.Etag
		versionIdMap[obj.Key] = objectCopy.VersionId
	}

	ctx.Export(OpBucketID, pulumi.String(bucketName))
	ctx.Export(OpObjectArns, arnMap)
	ctx.Export(OpObjectEtags, etagMap)
	ctx.Export(OpObjectVersionIds, versionIdMap)

	return nil
}

// objectTags builds one object's tag map. Merge order: identity tags <
// set-level tags < object-level tags, so an object can specialize but never
// lose its resource-identity attribution — the same precedence the Terraform
// module applies.
func objectTags(locals *Locals, spec *awss3objectsetv1alpha1.AwsS3ObjectSetSpec, obj *awss3objectsetv1alpha1.AwsS3Object) pulumi.StringMap {
	tags := pulumi.StringMap{}
	for k, v := range locals.AwsTags {
		tags[k] = pulumi.String(v)
	}
	for k, v := range spec.Tags {
		tags[k] = pulumi.String(v)
	}
	for k, v := range obj.Tags {
		tags[k] = pulumi.String(v)
	}
	return tags
}

// applyPlacement wires the destination placement surface shared verbatim by
// content-sourced and copy-sourced objects: storage class, encryption
// override, upload checksum, Object Lock, canned ACL, and the
// GOVERNANCE-retention delete bypass. Field semantics are documented on the
// content resource above; both resource types get identical sends.
func applyPlacement(obj *awss3objectsetv1alpha1.AwsS3Object,
	storageClass, serverSideEncryption, kmsKeyId *pulumi.StringPtrInput,
	bucketKeyEnabled *pulumi.BoolPtrInput, checksumAlgorithm *pulumi.StringPtrInput,
	objectLockMode, objectLockRetainUntilDate, objectLockLegalHoldStatus *pulumi.StringPtrInput,
	acl *pulumi.StringPtrInput, forceDestroy *pulumi.BoolPtrInput) {

	// Storage placement. Unset means STANDARD (AWS's default) — omit the
	// argument so the provider computes rather than pins a value.
	if obj.StorageClass != "" {
		*storageClass = pulumi.StringPtr(obj.StorageClass)
	}

	// Per-object encryption OVERRIDE. Unset inherits the bucket's default
	// encryption, which is where uniform posture belongs. A kms_key alone
	// implies aws:kms (the provider sets ServerSideEncryption when a KMS key
	// is sent); the spec CEL rejects the contradictory kms_key + AES256 pair.
	if obj.ServerSideEncryption != "" {
		*serverSideEncryption = pulumi.StringPtr(obj.ServerSideEncryption)
	}
	if kmsKeyArn := obj.KmsKey.GetValue(); kmsKeyArn != "" {
		*kmsKeyId = pulumi.StringPtr(kmsKeyArn)
	}
	if obj.BucketKeyEnabled != nil {
		*bucketKeyEnabled = pulumi.BoolPtr(obj.GetBucketKeyEnabled())
	}

	// Upload-integrity checksum, stored alongside the object.
	if obj.ChecksumAlgorithm != "" {
		*checksumAlgorithm = pulumi.StringPtr(obj.ChecksumAlgorithm)
	}

	// Object Lock (requires an Object Lock-enabled bucket; the spec CEL
	// guarantees mode and retain-until arrive as a pair).
	if obj.ObjectLockMode != "" {
		*objectLockMode = pulumi.StringPtr(obj.ObjectLockMode)
	}
	if obj.ObjectLockRetainUntilDate != "" {
		*objectLockRetainUntilDate = pulumi.StringPtr(obj.ObjectLockRetainUntilDate)
	}
	if obj.ObjectLockLegalHoldStatus != "" {
		*objectLockLegalHoldStatus = pulumi.StringPtr(obj.ObjectLockLegalHoldStatus)
	}

	// Canned ACL — only valid on buckets whose ownership controls permit
	// ACLs; modern (BucketOwnerEnforced) buckets reject it at apply time.
	if obj.Acl != "" {
		*acl = pulumi.StringPtr(obj.Acl)
	}

	// The GOVERNANCE-retention delete bypass
	// (x-amz-bypass-governance-retention). Only valid on Object Lock-enabled
	// buckets — S3 rejects the flag on regular buckets, failing the destroy.
	// Versioned-bucket cleanup needs no flag: deleting an object always
	// removes all of its versions.
	if obj.ForceDestroy {
		*forceDestroy = pulumi.BoolPtr(true)
	}
}
