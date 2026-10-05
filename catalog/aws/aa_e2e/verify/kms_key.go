package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	pkgerrors "github.com/pkg/errors"
)

// kmsKeyVerifier verifies an AwsKmsKey via DescribeKey, keyed on the key_id
// output. KMS keys are never deleted immediately: destroy schedules deletion
// and the key sits in PendingDeletion for the 7-30 day recovery window, still
// fully describable. So existence is "described AND not pending deletion",
// and absence accepts either the (eventual) NotFoundException or the
// PendingDeletion/PendingReplicaDeletion states -- otherwise verify-absent
// could never pass within a test run's lifetime.
//
// When the outputs report grants (the grant_ids map, keyed by spec
// position), existence is asserted per grant via ListGrants -- CreateGrant
// returning is not proof the grant landed on the key. Absence asserts the
// grants are gone FIRST (the module's default teardown REVOKES each grant --
// an immediate hard stop, unlike retirement), then applies the key's
// recovery-window logic: grants on a PendingDeletion key are still listable,
// so a revoke that silently failed would otherwise hide behind the key's
// "honestly gone" state.
type kmsKeyVerifier struct{}

func (*kmsKeyVerifier) IDOutputKey() string { return "key_id" }

func (*kmsKeyVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := kmsKeyExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awskmskey verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awskmskey %q not found after deploy", id)
	}
	return nil
}

func (*kmsKeyVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := kmsKeyExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awskmskey verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awskmskey %q still exists after destroy", id)
	}
	return nil
}

func (v *kmsKeyVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	keyId := stringOutputMap(outputs, "key_id")
	if keyId == "" {
		return pkgerrors.New("awskmskey verify-exists: no key_id in outputs")
	}
	if err := v.VerifyExists(ctx, cfg, keyId, region); err != nil {
		return err
	}
	wantGrantIds := kmsGrantIds(outputs)
	if len(wantGrantIds) == 0 {
		return nil
	}
	live, err := kmsListGrantIds(ctx, cfg, keyId, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awskmskey %q: listing grants failed", keyId)
	}
	for position, grantId := range wantGrantIds {
		if !live[grantId] {
			return pkgerrors.Errorf("awskmskey %q: grant %s (spec position %s) not among the key's grants",
				keyId, grantId, position)
		}
	}
	return nil
}

func (v *kmsKeyVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	keyId := stringOutputMap(outputs, "key_id")
	if keyId == "" {
		return pkgerrors.New("awskmskey verify-absent: no key_id in outputs")
	}
	// Grants first: they are listable on a PendingDeletion key, so a failed
	// revoke is only visible BEFORE the key's recovery-window state is
	// accepted as gone. ListGrants NotFound means the key (and with it every
	// grant) is already fully deleted -- absent by construction.
	if wantGrantIds := kmsGrantIds(outputs); len(wantGrantIds) > 0 {
		live, err := kmsListGrantIds(ctx, cfg, keyId, region)
		if err != nil {
			var notFound *types.NotFoundException
			if !errors.As(err, &notFound) {
				return pkgerrors.Wrapf(err, "awskmskey %q: listing grants failed", keyId)
			}
		} else {
			for position, grantId := range wantGrantIds {
				if live[grantId] {
					return pkgerrors.Errorf("awskmskey %q: grant %s (spec position %s) still present after destroy -- the revoke did not land",
						keyId, grantId, position)
				}
			}
		}
	}
	return v.VerifyAbsent(ctx, cfg, keyId, region)
}

// kmsGrantIds reads the grant_ids output map (spec position -> grant id). A
// missing or empty map means the spec declared no grants.
func kmsGrantIds(outputs map[string]interface{}) map[string]string {
	raw, ok := outputs["grant_ids"]
	if !ok || raw == nil {
		return nil
	}
	grantIds := map[string]string{}
	switch m := raw.(type) {
	case map[string]interface{}:
		for position, id := range m {
			if s, isStr := id.(string); isStr && s != "" {
				grantIds[position] = s
			}
		}
	case map[string]string:
		for position, id := range m {
			if id != "" {
				grantIds[position] = id
			}
		}
	}
	return grantIds
}

// kmsListGrantIds returns the set of grant ids currently attached to the key,
// following pagination (a key can carry many service-created grants beyond
// the module's own).
func kmsListGrantIds(ctx context.Context, cfg aws.Config, keyId, region string) (map[string]bool, error) {
	client := kmsClient(cfg, region)
	ids := map[string]bool{}
	var marker *string
	for {
		out, err := client.ListGrants(ctx, &kms.ListGrantsInput{KeyId: &keyId, Marker: marker})
		if err != nil {
			return nil, err
		}
		for _, grant := range out.Grants {
			if grant.GrantId != nil {
				ids[*grant.GrantId] = true
			}
		}
		if out.Truncated && out.NextMarker != nil {
			marker = out.NextMarker
			continue
		}
		return ids, nil
	}
}

func kmsClient(cfg aws.Config, region string) *kms.Client {
	return kms.NewFromConfig(cfg, func(o *kms.Options) {
		if region != "" {
			o.Region = region
		}
	})
}

func kmsKeyExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := kmsClient(cfg, region)
	out, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{KeyId: &id})
	if err != nil {
		var notFound *types.NotFoundException
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	if out.KeyMetadata == nil {
		return false, nil
	}
	// A key scheduled for deletion is destroyed from the harness's point of
	// view: the IaC destroy succeeded and only AWS's mandatory recovery window
	// keeps the key describable.
	switch out.KeyMetadata.KeyState {
	case types.KeyStatePendingDeletion, types.KeyStatePendingReplicaDeletion:
		return false, nil
	}
	return true, nil
}
