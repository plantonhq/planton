package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// gcsBucketIamMemberVerifier confirms an additive (role, member) pair is
// present in (or absent from) the BUCKET's IAM policy. The grant has no
// standalone server-side object -- the policy itself is the source of truth,
// so the verifier reads the whole policy and looks for the exact pair. The
// bucket output is the bucket name, exactly what the IAM methods take.
type gcsBucketIamMemberVerifier struct{}

func (v *gcsBucketIamMemberVerifier) IDOutputKey() string { return "member" }

func (v *gcsBucketIamMemberVerifier) grantInPolicy(ctx context.Context, svc *Services, outputs map[string]string) (bool, error) {
	bucket := outputs["bucket"]
	if bucket == "" {
		return false, errors.New("bucket output missing")
	}

	// Version 3 so bindings with IAM conditions are returned as distinct
	// entries rather than collapsed.
	policy, err := svc.Storage.Buckets.GetIamPolicy(bucket).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return false, err
	}

	for _, binding := range policy.Bindings {
		if binding.Role != outputs["role"] {
			continue
		}
		for _, member := range binding.Members {
			// IAM treats member emails case-insensitively; compare the same
			// way so server-side casing normalization never fails a probe.
			if strings.EqualFold(member, outputs["member"]) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (v *gcsBucketIamMemberVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	found, err := v.grantInPolicy(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "failed to read IAM policy of bucket %s", outputs["bucket"])
	}
	if !found {
		return errors.Errorf("grant of %s to %s not present in bucket IAM policy after deploy",
			outputs["role"], outputs["member"])
	}
	return nil
}

func (v *gcsBucketIamMemberVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	found, err := v.grantInPolicy(ctx, svc, outputs)
	if err != nil {
		// A deleted bucket takes every grant on it along.
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "failed to read IAM policy of bucket %s", outputs["bucket"])
	}
	if found {
		return errors.Errorf("grant of %s to %s still present in bucket IAM policy after destroy",
			outputs["role"], outputs["member"])
	}
	return nil
}
