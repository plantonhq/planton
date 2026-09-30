package verify

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/cloudflare/r2"
)

// The behavioral-r2 scenarios' object-store proof, shared by Loki and Tempo.
// An r2 arm is only worth anything if data actually lands in the bucket,
// so after the workload flushes, this proof lists the real bucket over R2's
// S3 API and requires objects written by this run under the workload's
// tenant prefix. The bucket and key pair are the owner-arranged test
// fixture in the planton-e2e Cloudflare account, handed to the lane in the
// same environment variables the scenario manifests read through
// ${E2E_ENV:...} tokens, so the proof reads exactly what the workload wrote
// with.

const (
	envR2AccountID       = "PLANTON_E2E_CLOUDFLARE_ACCOUNT_ID"
	envR2Bucket          = "PLANTON_E2E_CLOUDFLARE_R2_BUCKET"
	envR2AccessKeyID     = "PLANTON_E2E_CLOUDFLARE_R2_ACCESS_KEY_ID"
	envR2SecretAccessKey = "PLANTON_E2E_CLOUDFLARE_R2_SECRET_ACCESS_KEY"
)

// r2TestClient builds an S3 client for the test fixture bucket from the
// lane's environment, the host composed by pkg/cloudflare/r2 exactly as
// the modules compose it.
func r2TestClient() (*s3.Client, string, error) {
	values := map[string]string{}
	for _, name := range []string{envR2AccountID, envR2Bucket, envR2AccessKeyID, envR2SecretAccessKey} {
		values[name] = os.Getenv(name)
		if values[name] == "" {
			return nil, "", errors.Errorf("R2: %s is not set; the behavioral-r2 lane needs the owner-arranged test bucket and its key pair", name)
		}
	}
	client := s3.New(s3.Options{
		Region:       r2.Region,
		BaseEndpoint: awssdk.String(r2.S3Endpoint(values[envR2AccountID], "")),
		Credentials:  credentials.NewStaticCredentialsProvider(values[envR2AccessKeyID], values[envR2SecretAccessKey], ""),
		UsePathStyle: true,
	})
	return client, values[envR2Bucket], nil
}

// awaitR2Objects polls the test bucket until objects under prefix carry a
// LastModified at or after since, and returns their keys. Objects from
// earlier runs share the bucket; the time filter is what makes the proof
// this run's.
func awaitR2Objects(ctx context.Context, prefix string, since time.Time, budget time.Duration) ([]string, error) {
	client, bucket, err := r2TestClient()
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(budget)
	var lastErr error
	for time.Now().Before(deadline) {
		var fresh []string
		paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
			Bucket: awssdk.String(bucket),
			Prefix: awssdk.String(prefix),
		})
		lastErr = nil
		for paginator.HasMorePages() {
			page, pageErr := paginator.NextPage(ctx)
			if pageErr != nil {
				lastErr = pageErr
				break
			}
			for _, object := range page.Contents {
				if object.LastModified != nil && !object.LastModified.Before(since.Add(-time.Minute)) {
					fresh = append(fresh, awssdk.ToString(object.Key))
				}
			}
		}
		if lastErr == nil && len(fresh) > 0 {
			return fresh, nil
		}
		time.Sleep(10 * time.Second)
	}
	if lastErr != nil {
		return nil, errors.Wrapf(lastErr, "R2: listing %s/%s", bucket, prefix)
	}
	return nil, errors.Errorf("R2: no object under %s/%s was written by this run within %s", bucket, prefix, budget)
}

// removeR2Objects deletes the keys a proof found, so the shared fixture
// bucket holds only what the next run writes (its lifecycle rule is the
// backstop for anything a crashed run leaves).
func removeR2Objects(ctx context.Context, keys []string) {
	client, bucket, err := r2TestClient()
	if err != nil {
		return
	}
	for _, key := range keys {
		_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: awssdk.String(bucket), Key: awssdk.String(key)})
	}
}

// summarizeKeys renders a found key set for the lane log.
func summarizeKeys(keys []string) string {
	if len(keys) <= 3 {
		return strings.Join(keys, ", ")
	}
	return fmt.Sprintf("%s, ... (%d objects)", strings.Join(keys[:3], ", "), len(keys))
}
