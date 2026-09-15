package verify

import (
	"context"
	"errors"
	"fmt"
	"os"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// bucketVerifier verifies a DigitalOceanBucket (a Spaces bucket). Spaces is
// an S3-compatible object store the DigitalOcean REST API cannot read, so
// this verifier speaks the S3 API against the bucket's regional endpoint
// (https://{region}.digitaloceanspaces.com) with the provider-canonical
// SPACES_ACCESS_KEY_ID / SPACES_SECRET_ACCESS_KEY credentials -- the same
// variables both IaC engines' provider blocks resolve. HeadBucket is the
// canonical existence probe. Addressing needs region + name, so the verifier
// uses the outputs form.
type bucketVerifier struct{}

func (*bucketVerifier) IDOutputKey() string { return "bucket_id" }

func (v *bucketVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanbucket requires the full outputs map (bucket_id + region); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (v *bucketVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanbucket requires the full outputs map (bucket_id + region); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *bucketVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := spacesBucketExistsFromOutputs(ctx, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanbucket verify-exists failed")
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceanbucket %q not found after deploy", StringOutput(outputs, "bucket_id"))
	}
	return nil
}

func (v *bucketVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := spacesBucketExistsFromOutputs(ctx, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanbucket verify-absent failed")
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanbucket", ID: StringOutput(outputs, "bucket_id")}
	}
	return nil
}

func spacesBucketExistsFromOutputs(ctx context.Context, outputs map[string]interface{}) (bool, error) {
	// The bucket_id output IS the bucket name (a Spaces bucket's provider
	// resource id); region completes the S3 endpoint address.
	name := StringOutput(outputs, "bucket_id")
	region := StringOutput(outputs, "region")
	if name == "" || region == "" {
		return false, pkgerrors.Errorf("outputs must carry bucket_id and region (got bucket_id=%q, region=%q)", name, region)
	}

	accessKey := os.Getenv("SPACES_ACCESS_KEY_ID")
	secretKey := os.Getenv("SPACES_SECRET_ACCESS_KEY")
	if accessKey == "" || secretKey == "" {
		return false, pkgerrors.New("Spaces credentials not in the environment: bucket lanes need " +
			"SPACES_ACCESS_KEY_ID and SPACES_SECRET_ACCESS_KEY (the provider-canonical names)")
	}

	// The region shapes only the endpoint; the SDK's Region field is
	// satisfied with a fixed value, mirroring the upstream provider's own
	// Spaces client construction.
	s3Client := s3.New(s3.Options{
		Region:       "us-east-1",
		BaseEndpoint: awssdk.String(fmt.Sprintf("https://%s.digitaloceanspaces.com", region)),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	})

	_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: awssdk.String(name)})
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
