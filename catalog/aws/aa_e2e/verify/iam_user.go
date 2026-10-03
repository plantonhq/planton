package verify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	pkgerrors "github.com/pkg/errors"
)

// iamUserVerifier verifies an AwsIamUser via GetUser, keyed on the user name.
// IAM is a global service, so the region parameter is ignored. A deleted user
// returns the typed NoSuchEntity error, which is the "absent" signal; any
// other error is a genuine failure and must surface.
//
// When the outputs report an access key (access_key_id -- present
// exactly when the spec asked for one), existence asserts the key exists on
// the user via ListAccessKeys: CreateUser succeeding says nothing about the
// key satellite. The key's Active/Inactive STATUS is a spec-authored
// expectation not derivable from outputs, so status evidence stays with the
// lane's CLI snapshot, never a verifier guess. Absence needs only the user
// probe -- access keys are children of the user (and force_destroy removes
// them with it).
type iamUserVerifier struct{}

func (*iamUserVerifier) IDOutputKey() string { return "user_name" }

func (*iamUserVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamUserExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamuser verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsiamuser %q not found after deploy", id)
	}
	return nil
}

func (*iamUserVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamUserExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamuser verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsiamuser %q still exists after destroy", id)
	}
	return nil
}

func (v *iamUserVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	userName := stringOutputMap(outputs, "user_name")
	if userName == "" {
		return pkgerrors.New("awsiamuser verify-exists: no user_name in outputs")
	}
	if err := v.VerifyExists(ctx, cfg, userName, region); err != nil {
		return err
	}
	wantAccessKeyId := stringOutputMap(outputs, "access_key_id")
	if wantAccessKeyId == "" {
		return nil
	}
	listed, err := iam.NewFromConfig(cfg).ListAccessKeys(ctx, &iam.ListAccessKeysInput{UserName: aws.String(userName)})
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamuser %q: listing access keys failed", userName)
	}
	// The status output carries the provider's post-apply read of the
	// rotation lever -- asserting it against IAM's own listing proves an
	// Inactive key really landed Inactive, not merely that the apply
	// returned.
	wantStatus := stringOutputMap(outputs, "access_key_status")
	for _, key := range listed.AccessKeyMetadata {
		if aws.ToString(key.AccessKeyId) != wantAccessKeyId {
			continue
		}
		if wantStatus != "" && string(key.Status) != wantStatus {
			return pkgerrors.Errorf("awsiamuser %q: access key %s status is %q, outputs claim %q",
				userName, wantAccessKeyId, key.Status, wantStatus)
		}
		return nil
	}
	return pkgerrors.Errorf("awsiamuser %q: access key %s not among the user's keys", userName, wantAccessKeyId)
}

func (v *iamUserVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	userName := stringOutputMap(outputs, "user_name")
	if userName == "" {
		return pkgerrors.New("awsiamuser verify-absent: no user_name in outputs")
	}
	return v.VerifyAbsent(ctx, cfg, userName, region)
}

func iamUserExists(ctx context.Context, cfg aws.Config, userName string) (bool, error) {
	client := iam.NewFromConfig(cfg)
	_, err := client.GetUser(ctx, &iam.GetUserInput{UserName: aws.String(userName)})
	if err != nil {
		if isIamNoSuchEntity(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
