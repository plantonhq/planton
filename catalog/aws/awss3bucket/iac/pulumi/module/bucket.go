package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// bucket creates the root bucket resource and its security-posture
// satellites: versioning, default encryption, public access block, ownership
// controls, canned ACL, and the bucket policy. It returns the bucket and the
// versioning satellite (nil when unmanaged) so lifecycle and replication can
// take an explicit ordering edge on versioning.
//
// AWS models the bucket as a small root resource plus one satellite resource
// per behavioral setting; this module mirrors that grain. Two satellites are
// ALWAYS created — public access block and ownership controls — because their
// absence in the spec means the secure default, and stating that default
// explicitly in state is what keeps a bucket provably private.
func bucket(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*s3.BucketV2, *s3.BucketVersioningV2, error) {
	spec := locals.Spec

	bucketArgs := &s3.BucketV2Args{
		Bucket: pulumi.String(locals.BucketName),
		// force_destroy empties the bucket (all versions and delete markers)
		// before deletion — without it a non-empty bucket fails the destroy.
		ForceDestroy: pulumi.Bool(spec.ForceDestroy),
		// Object Lock can only be turned on at creation (changing it replaces
		// the bucket), which is why it lives on the root resource and not in
		// the object-lock satellite.
		ObjectLockEnabled: pulumi.Bool(spec.ObjectLockEnabled),
		Tags:              pulumi.ToStringMap(locals.AwsTags),
	}
	// Namespace is create-time identity like the name itself: unset keeps
	// the classic global namespace; "account-regional" scopes the name to
	// this account+region. Changing it replaces the bucket.
	if spec.BucketNamespace != "" {
		bucketArgs.BucketNamespace = pulumi.StringPtr(spec.BucketNamespace)
	}
	createdBucket, err := s3.NewBucketV2(ctx, "bucket", bucketArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create S3 bucket")
	}

	// --- Versioning -----------------------------------------------------
	// Only managed when the spec sets a state: an unset spec leaves the
	// bucket unversioned WITHOUT creating the satellite. AWS cannot return a
	// bucket to the never-versioned state once enabled — use Suspended.
	var versioning *s3.BucketVersioningV2
	if spec.VersioningStatus != "" {
		versioning, err = s3.NewBucketVersioningV2(ctx, "versioning", &s3.BucketVersioningV2Args{
			Bucket: createdBucket.ID(),
			VersioningConfiguration: &s3.BucketVersioningV2VersioningConfigurationArgs{
				Status: pulumi.String(spec.VersioningStatus),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to configure versioning")
		}
	}
	// --- Default encryption ----------------------------------------------
	// Created only when the spec opts in: since January 2023 AWS itself
	// encrypts every object with SSE-S3 by default, so an absent block
	// already means "encrypted with AES256" without configuration to manage.
	if spec.Encryption != nil {
		sseAlgorithm := spec.Encryption.SseAlgorithm
		if sseAlgorithm == "" {
			sseAlgorithm = "AES256"
		}
		rule := &s3.BucketServerSideEncryptionConfigurationV2RuleArgs{
			ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationV2RuleApplyServerSideEncryptionByDefaultArgs{
				SseAlgorithm: pulumi.String(sseAlgorithm),
			},
			// Bucket keys cut SSE-KMS request costs by up to 99%; harmless
			// under AES256 where AWS ignores the flag.
			BucketKeyEnabled: pulumi.Bool(spec.Encryption.BucketKeyEnabled),
		}
		if spec.Encryption.KmsKeyId.GetValue() != "" {
			rule.ApplyServerSideEncryptionByDefault = &s3.BucketServerSideEncryptionConfigurationV2RuleApplyServerSideEncryptionByDefaultArgs{
				SseAlgorithm:   pulumi.String(sseAlgorithm),
				KmsMasterKeyId: pulumi.String(spec.Encryption.KmsKeyId.GetValue()),
			}
		}
		// Encryption types PUTs may no longer use (e.g. "SSE-C"); sent only
		// when the manifest states a posture so unset keeps AWS's default.
		if len(spec.Encryption.BlockedEncryptionTypes) > 0 {
			rule.BlockedEncryptionTypes = pulumi.ToStringArray(spec.Encryption.BlockedEncryptionTypes)
		}
		_, err = s3.NewBucketServerSideEncryptionConfigurationV2(ctx, "encryption", &s3.BucketServerSideEncryptionConfigurationV2Args{
			Bucket: createdBucket.ID(),
			Rules:  s3.BucketServerSideEncryptionConfigurationV2RuleArray{rule},
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to configure encryption")
		}
	}

	// --- Public access block (always managed) ------------------------------
	// ABSENCE of the spec block means fully private (all four guards on); the
	// spec block only exists to relax specific guards. Inside the block each
	// guard is presence-tracked with a declared default of ON, so a guard the
	// manifest left unset arrives here already true and only an explicit
	// false relaxes it.
	blockPublicAcls, blockPublicPolicy, ignorePublicAcls, restrictPublicBuckets := true, true, true, true
	if spec.PublicAccessBlock != nil {
		blockPublicAcls = spec.PublicAccessBlock.GetBlockPublicAcls()
		blockPublicPolicy = spec.PublicAccessBlock.GetBlockPublicPolicy()
		ignorePublicAcls = spec.PublicAccessBlock.GetIgnorePublicAcls()
		restrictPublicBuckets = spec.PublicAccessBlock.GetRestrictPublicBuckets()
	}
	publicAccessBlock, err := s3.NewBucketPublicAccessBlock(ctx, "public-access-block", &s3.BucketPublicAccessBlockArgs{
		Bucket:                createdBucket.ID(),
		BlockPublicAcls:       pulumi.Bool(blockPublicAcls),
		BlockPublicPolicy:     pulumi.Bool(blockPublicPolicy),
		IgnorePublicAcls:      pulumi.Bool(ignorePublicAcls),
		RestrictPublicBuckets: pulumi.Bool(restrictPublicBuckets),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to configure public access block")
	}

	// --- Ownership controls (always managed) -------------------------------
	// Empty means the modern BucketOwnerEnforced (ACLs disabled). Spelled out
	// so the posture is explicit in state.
	objectOwnership := spec.ObjectOwnership
	if objectOwnership == "" {
		objectOwnership = "BucketOwnerEnforced"
	}
	ownershipControls, err := s3.NewBucketOwnershipControls(ctx, "ownership-controls", &s3.BucketOwnershipControlsArgs{
		Bucket: createdBucket.ID(),
		Rule: &s3.BucketOwnershipControlsRuleArgs{
			ObjectOwnership: pulumi.String(objectOwnership),
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to configure ownership controls")
	}

	// --- Canned ACL --------------------------------------------------------
	// Only valid when ownership controls re-enable ACLs (CEL guarantees the
	// coupling). Ordering matters: the ACL PUT fails while BucketOwnerEnforced
	// is in effect, so the ownership setting must land first.
	if spec.Acl != "" {
		_, err = s3.NewBucketAclV2(ctx, "acl", &s3.BucketAclV2Args{
			Bucket: createdBucket.ID(),
			Acl:    pulumi.String(spec.Acl),
		}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{ownershipControls}))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to configure acl")
		}
	}

	// --- Bucket policy -------------------------------------------------------
	// Applied after the public access block: a policy granting public access
	// is rejected while block_public_policy is on, so when a manifest relaxes
	// the guard and adds a public-read statement in the same apply, the guard
	// change must land first.
	if spec.Policy != nil {
		policyJSON, err := json.Marshal(spec.Policy.AsMap())
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to serialize bucket policy")
		}
		_, err = s3.NewBucketPolicy(ctx, "policy", &s3.BucketPolicyArgs{
			Bucket: createdBucket.ID(),
			Policy: pulumi.String(string(policyJSON)),
		}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{publicAccessBlock}))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to configure bucket policy")
		}
	}

	return createdBucket, versioning, nil
}
