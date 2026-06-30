package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	pkgerrors "github.com/pkg/errors"
)

// s3Verifier verifies an AwsS3Bucket via HeadBucket -- the canonical existence
// probe, which needs no s3:ListAllMyBuckets permission.
type s3Verifier struct{}

func (*s3Verifier) IDOutputKey() string { return "bucket_id" }

func (*s3Verifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := headBucket(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awss3bucket verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awss3bucket %q not found after deploy", id)
	}
	return nil
}

func (*s3Verifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := headBucket(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awss3bucket verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awss3bucket %q still exists after destroy", id)
	}
	return nil
}

// headBucket returns whether the bucket exists. A typed not-found error is the
// "absent" signal; any other error is a genuine failure (e.g. no credentials,
// access denied) and must surface rather than masquerade as absence.
func headBucket(ctx context.Context, cfg aws.Config, bucket, region string) (bool, error) {
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if region != "" {
			o.Region = region
		}
	})
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return true, nil
	}
	var notFound *s3types.NotFound
	if errors.As(err, &notFound) {
		return false, nil
	}
	var noSuchBucket *s3types.NoSuchBucket
	if errors.As(err, &noSuchBucket) {
		return false, nil
	}
	return false, err
}
