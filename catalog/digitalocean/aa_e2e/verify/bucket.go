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
//
// Beyond existence, the deploy side asserts every wiring value the module
// CLAIMS in its stack outputs: `region` against the bucket's live location
// (Spaces answers GetBucketLocation with the region slug), and `endpoint`,
// `bucket_domain_name`, and `urn` against the shapes DigitalOcean defines
// for them. The region assertion is the one that matters most: a Spaces
// bucket addressed through the WRONG regional endpoint answers 404, not a
// redirect, so a module exporting a wrong region would make the
// destroy-side absence check pass for a bucket that still exists. Proving
// the region at deploy is what makes the 404 at destroy trustworthy.
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
	name, region, s3Client, err := spacesClientFromOutputs(outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanbucket verify-exists failed")
	}

	exists, err := spacesBucketExists(ctx, s3Client, name)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanbucket verify-exists failed")
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceanbucket %q not found after deploy", name)
	}

	// The claimed region must be where the bucket really lives. Spaces
	// reports the region slug as the LocationConstraint (measured: "nyc3").
	location, err := s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: awssdk.String(name)})
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanbucket %q GetBucketLocation failed", name)
	}
	if live := string(location.LocationConstraint); live != region {
		return pkgerrors.Errorf("digitaloceanbucket %q region mismatch: output %q, live location %q", name, region, live)
	}

	// The remaining wiring outputs follow DigitalOcean's fixed shapes for a
	// Spaces bucket; a mismatch means a module exported the wrong attribute
	// (for example the engine's own resource URN instead of the DigitalOcean
	// URN). `endpoint` is the region-level HOST with no scheme and no bucket
	// name -- the kind's contract and the provider's `endpoint` attribute
	// agree on that (measured live: "nyc3.digitaloceanspaces.com").
	claims := []struct {
		output string
		want   string
	}{
		{"endpoint", fmt.Sprintf("%s.digitaloceanspaces.com", region)},
		{"bucket_domain_name", fmt.Sprintf("%s.%s.digitaloceanspaces.com", name, region)},
		{"urn", fmt.Sprintf("do:space:%s", name)},
	}
	for _, c := range claims {
		if claimed := StringOutput(outputs, c.output); claimed != "" && claimed != c.want {
			return pkgerrors.Errorf("digitaloceanbucket %q %s mismatch: output %q, want %q", name, c.output, claimed, c.want)
		}
	}

	return nil
}

func (v *bucketVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	name, _, s3Client, err := spacesClientFromOutputs(outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanbucket verify-absent failed")
	}
	exists, err := spacesBucketExists(ctx, s3Client, name)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanbucket verify-absent failed")
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanbucket", ID: name}
	}
	return nil
}

// spacesClientFromOutputs reads the bucket's address from the outputs and
// builds the S3 client for its regional endpoint.
func spacesClientFromOutputs(outputs map[string]interface{}) (name, region string, client *s3.Client, err error) {
	// The bucket_id output IS the bucket name (a Spaces bucket's provider
	// resource id); region completes the S3 endpoint address.
	name = StringOutput(outputs, "bucket_id")
	region = StringOutput(outputs, "region")
	if name == "" || region == "" {
		return "", "", nil, pkgerrors.Errorf("outputs must carry bucket_id and region (got bucket_id=%q, region=%q)", name, region)
	}

	accessKey := os.Getenv("SPACES_ACCESS_KEY_ID")
	secretKey := os.Getenv("SPACES_SECRET_ACCESS_KEY")
	if accessKey == "" || secretKey == "" {
		return "", "", nil, pkgerrors.New("Spaces credentials not in the environment: bucket lanes need " +
			"SPACES_ACCESS_KEY_ID and SPACES_SECRET_ACCESS_KEY (the provider-canonical names)")
	}

	// The region shapes only the endpoint; the SDK's Region field is
	// satisfied with a fixed value, mirroring the upstream provider's own
	// Spaces client construction.
	client = s3.New(s3.Options{
		Region:       "us-east-1",
		BaseEndpoint: awssdk.String(fmt.Sprintf("https://%s.digitaloceanspaces.com", region)),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	})
	return name, region, client, nil
}

func spacesBucketExists(ctx context.Context, s3Client *s3.Client, name string) (bool, error) {
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
