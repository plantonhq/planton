package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// pubSubTopicIamMemberVerifier confirms an additive (role, member) pair is
// present in (or absent from) the TOPIC's IAM policy. The grant has no
// standalone server-side object -- the policy itself is the source of truth,
// so the verifier reads the whole policy and looks for the exact pair. The
// topic output is the full name (projects/<project>/topics/<topic>), exactly
// the resource the IAM methods take.
type pubSubTopicIamMemberVerifier struct{}

func (v *pubSubTopicIamMemberVerifier) IDOutputKey() string { return "member" }

func (v *pubSubTopicIamMemberVerifier) grantInPolicy(ctx context.Context, svc *Services, outputs map[string]string) (bool, error) {
	topic := outputs["topic"]
	if topic == "" {
		return false, errors.New("topic output missing")
	}

	// Version 3 so any conditional binding comes back as its own entry
	// rather than collapsed.
	policy, err := svc.PubSub.Projects.Topics.GetIamPolicy(topic).
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

func (v *pubSubTopicIamMemberVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	found, err := v.grantInPolicy(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "failed to read IAM policy of topic %s", outputs["topic"])
	}
	if !found {
		return errors.Errorf("grant of %s to %s not present in topic IAM policy after deploy",
			outputs["role"], outputs["member"])
	}
	return nil
}

func (v *pubSubTopicIamMemberVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	found, err := v.grantInPolicy(ctx, svc, outputs)
	if err != nil {
		// A deleted topic takes every grant on it along.
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "failed to read IAM policy of topic %s", outputs["topic"])
	}
	if found {
		return errors.Errorf("grant of %s to %s still present in topic IAM policy after destroy",
			outputs["role"], outputs["member"])
	}
	return nil
}
