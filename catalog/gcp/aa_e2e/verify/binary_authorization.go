package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// binaryAuthorizationGet reads a Binary Authorization policy or attestor by
// its full name from https://binaryauthorization.googleapis.com/v1/.
func binaryAuthorizationGet(ctx context.Context, svc *Services, what, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, what, fmt.Sprintf("https://binaryauthorization.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// containerAnalysisNoteGet reads an Artifact Analysis note by its full
// name from Google's global https://containeranalysis.googleapis.com/v1/
// endpoint (Binary Authorization supports no regional one).
func containerAnalysisNoteGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "artifact analysis note", fmt.Sprintf("https://containeranalysis.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// binaryAuthorizationPolicyVerifier probes a project's policy. Google
// keeps one policy per project, so destroy (deletion_policy DELETE) writes
// Google's default back rather than deleting it: the verdict is that the
// policy reads back allowing every image with no per-cluster rules.
type binaryAuthorizationPolicyVerifier struct{}

// IDOutputKey is the policy's name (projects/{project}/policy).
func (v *binaryAuthorizationPolicyVerifier) IDOutputKey() string { return "name" }

func (v *binaryAuthorizationPolicyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	policy, _, err := binaryAuthorizationGet(ctx, svc, "binary authorization policy", name)
	if err != nil {
		return errors.Wrapf(err, "binary authorization policy %s not readable after deploy", name)
	}
	rule, _ := policy["defaultAdmissionRule"].(map[string]interface{})
	if stringField(rule, "evaluationMode") == "" || stringField(rule, "enforcementMode") == "" {
		return errors.Errorf("binary authorization policy %s has no default admission rule after deploy", name)
	}
	if want := "projects/" + outputs["project_id"] + "/policy"; name != want {
		return errors.Errorf("binary authorization policy name %s does not match the project_id output (%s)", name, want)
	}
	return nil
}

func (v *binaryAuthorizationPolicyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	policy, _, err := binaryAuthorizationGet(ctx, svc, "binary authorization policy", name)
	if err != nil {
		return errors.Wrapf(err, "binary authorization policy %s not readable after destroy", name)
	}
	rule, _ := policy["defaultAdmissionRule"].(map[string]interface{})
	if mode := stringField(rule, "evaluationMode"); mode != "ALWAYS_ALLOW" {
		return errors.Errorf("binary authorization policy %s evaluates %q after destroy; Google's default is ALWAYS_ALLOW", name, mode)
	}
	if rules, _ := policy["clusterAdmissionRules"].(map[string]interface{}); len(rules) > 0 {
		return errors.Errorf("binary authorization policy %s keeps %d per-cluster rules after destroy", name, len(rules))
	}
	return nil
}

// binaryAuthorizationAttestorVerifier probes an attestor: it reads back
// under its full name naming the note_reference output, carrying the
// delegation service account the output reports, and the note it names
// exists.
type binaryAuthorizationAttestorVerifier struct{}

// IDOutputKey is the attestor's full name
// (projects/{project}/attestors/{name}).
func (v *binaryAuthorizationAttestorVerifier) IDOutputKey() string { return "attestor_id" }

func (v *binaryAuthorizationAttestorVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["attestor_id"]
	if name == "" {
		return errors.New("attestor_id output missing after deploy")
	}
	attestor, _, err := binaryAuthorizationGet(ctx, svc, "binary authorization attestor", name)
	if err != nil {
		return errors.Wrapf(err, "binary authorization attestor %s not found after deploy", name)
	}
	note, _ := attestor["userOwnedGrafeasNote"].(map[string]interface{})
	if ref := stringField(note, "noteReference"); ref == "" || ref != outputs["note_reference"] {
		return errors.Errorf("binary authorization attestor %s names note %q, but the note_reference output is %q", name, ref, outputs["note_reference"])
	}
	if email := stringField(note, "delegationServiceAccountEmail"); email == "" || email != outputs["delegation_service_account_email"] {
		return errors.Errorf("binary authorization attestor %s delegates to %q, but the output is %q", name, email, outputs["delegation_service_account_email"])
	}
	if _, _, err := containerAnalysisNoteGet(ctx, svc, outputs["note_reference"]); err != nil {
		return errors.Wrapf(err, "note %s of attestor %s not found", outputs["note_reference"], name)
	}
	return nil
}

func (v *binaryAuthorizationAttestorVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["attestor_id"]
	if name == "" {
		return nil
	}
	_, status, err := binaryAuthorizationGet(ctx, svc, "binary authorization attestor", name)
	return restAbsent("binary authorization attestor", name, status, err)
}
