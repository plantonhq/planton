package verify

import (
	"context"
	"errors"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// spacesKeyVerifier verifies a DigitalOceanSpacesKey. Keys are control-plane
// objects reached with the account token via GET /v2/spaces/keys/{access_key}
// (the access key ID is the resource identity), so existence and absence are
// ordinary godo probes.
//
// The kind's headline promise is the PAIR, and the secret half is write-once:
// DigitalOcean returns secret_key only in the create response and never
// again, so an engine that failed to capture it at create has shipped a key
// nobody can use, and no later read could reveal that. The deploy-side check
// therefore signs one S3 request with the captured pair against the Spaces
// data plane. The S3 plane's answer classifies the secret, not the grant:
// InvalidAccessKeyId / SignatureDoesNotMatch mean the pair is wrong;
// AccessDenied means the signature was accepted and only the grant scope
// stopped the call (a bucket-scoped key may not list buckets), which still
// proves the secret is real. Measured live: a freshly minted pair is accepted
// by the S3 plane within two seconds of the create response, so the probe is
// a single call, not a poll.
type spacesKeyVerifier struct{}

func (*spacesKeyVerifier) IDOutputKey() string { return "access_key" }

func (v *spacesKeyVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanspaceskey requires the full outputs map (access_key + secret_key); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (v *spacesKeyVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanspaceskey requires the full outputs map (access_key + secret_key); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *spacesKeyVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	accessKey := StringOutput(outputs, "access_key")
	if accessKey == "" {
		return pkgerrors.New("access_key output missing after deploy")
	}
	key, _, err := client.SpacesKeys.Get(ctx, accessKey)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceanspaceskey %q not found after deploy", accessKey)
		}
		return pkgerrors.Wrapf(err, "digitaloceanspaceskey verify-exists failed for %q", accessKey)
	}
	if key.AccessKey != accessKey {
		return pkgerrors.Errorf("digitaloceanspaceskey %q: live access key %q differs from the output", accessKey, key.AccessKey)
	}

	// Both engines export the secret (Terraform `sensitive`, Pulumi
	// `ToSecret`); the runner reads Pulumi outputs with --show-secrets, so an
	// empty value here means the engine lost the create response's secret.
	secretKey := StringOutput(outputs, "secret_key")
	if secretKey == "" {
		return pkgerrors.Errorf("digitaloceanspaceskey %q: secret_key output is empty -- the write-once secret was not captured at create", accessKey)
	}
	return spacesPairSignatureAccepted(ctx, accessKey, secretKey)
}

func (v *spacesKeyVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	accessKey := StringOutput(outputs, "access_key")
	if accessKey == "" {
		return pkgerrors.New("access_key output missing for destroy verification")
	}
	_, _, err := client.SpacesKeys.Get(ctx, accessKey)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceanspaceskey verify-absent failed for %q", accessKey)
	}
	return &StillExistsError{Component: "digitaloceanspaceskey", ID: accessKey}
}

// spacesPairSignatureAccepted signs one ListBuckets against the Spaces S3
// plane with the captured pair. Any regional endpoint serves the call (the
// listing is account-wide; region only shapes the host), and the client is
// built exactly as the bucket verifier builds its own.
func spacesPairSignatureAccepted(ctx context.Context, accessKey, secretKey string) error {
	s3Client := s3.New(s3.Options{
		Region:       "us-east-1",
		BaseEndpoint: awssdk.String("https://nyc3.digitaloceanspaces.com"),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	})
	_, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err == nil {
		return nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "AccessDenied":
			// Signature accepted; the grant scope (not the secret) refused the
			// listing. The pair is real.
			return nil
		case "InvalidAccessKeyId", "SignatureDoesNotMatch":
			return pkgerrors.Errorf("digitaloceanspaceskey %q: the Spaces S3 plane rejected the captured pair (%s) -- the secret_key output does not match the live key", accessKey, apiErr.ErrorCode())
		}
	}
	return pkgerrors.Wrapf(err, "digitaloceanspaceskey %q: S3 signature probe failed", accessKey)
}
