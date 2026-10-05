package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// iamDenyPolicyVerifier probes an IAM deny policy through the IAM v2 API.
// The output policy_name carries {url-encoded-parent}/{policy_name};
// the v2 API addresses the policy as
// policies/{url-encoded-parent}/denypolicies/{policy_name}. The encoded
// parent contains no "/" (that is the point of the encoding), so the first
// path segment is always the parent.
type iamDenyPolicyVerifier struct{}

func (v *iamDenyPolicyVerifier) IDOutputKey() string { return "policy_name" }

// denyPolicyApiName renders the v2 API resource name from the output.
func denyPolicyApiName(policyName string) (string, error) {
	parent, name, found := strings.Cut(policyName, "/")
	if !found || parent == "" || name == "" {
		return "", errors.Errorf("deny policy output %q is not {encoded-parent}/{name}", policyName)
	}
	return fmt.Sprintf("policies/%s/denypolicies/%s", parent, name), nil
}

func (v *iamDenyPolicyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	apiName, err := denyPolicyApiName(outputs["policy_name"])
	if err != nil {
		return err
	}
	if _, err := svc.IamV2.Policies.Get(apiName).Context(ctx).Do(); err != nil {
		return errors.Wrapf(err, "deny policy %s not found after deploy", apiName)
	}
	return nil
}

func (v *iamDenyPolicyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	apiName, err := denyPolicyApiName(outputs["policy_name"])
	if err != nil {
		return err
	}
	_, err = svc.IamV2.Policies.Get(apiName).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing deny policy %s after destroy", apiName)
	}
	return errors.Errorf("deny policy %s still exists after destroy", apiName)
}
