package verify

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
)

// firebaseProjectVerifier probes a project's Firebase enablement through
// the Firebase Management API (projects/{project}).
//
// Firebase enablement is PERMANENT by Google's design: there is no API to
// remove Firebase from a project, and both engines' destroy detaches the
// resource from state. VERIFY-CLN therefore asserts the detach contract --
// the project must still read as a Firebase project after destroy -- rather
// than absence (the Identity Platform grain, not the KMS key-ring grain:
// the enablement genuinely cannot go away, so unreadable is a regression).
type firebaseProjectVerifier struct{}

func (v *firebaseProjectVerifier) IDOutputKey() string { return "project_id" }

func (v *firebaseProjectVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	projectID := outputs["project_id"]
	if projectID == "" {
		return errors.New("project_id output missing after deploy")
	}
	fp, err := svc.Firebase.Projects.Get("projects/" + projectID).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "firebase project %s not readable after deploy", projectID)
	}
	if fp.State != "ACTIVE" {
		return errors.Errorf("firebase project %s is in state %q, want ACTIVE", projectID, fp.State)
	}
	// The project number is the FCM sender id every client registers
	// with; the output must be the API's own value, not a guess.
	if want := outputs["project_number"]; want != "" && strconv.FormatInt(fp.ProjectNumber, 10) != want {
		return errors.Errorf("firebase project %s reports project number %d, want the project_number output %q", projectID, fp.ProjectNumber, want)
	}
	return nil
}

func (v *firebaseProjectVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	projectID := outputs["project_id"]
	if projectID == "" {
		return nil
	}
	// The detach contract: the project keeps Firebase enabled after
	// destroy -- a readable, ACTIVE Firebase project IS the expected clean
	// state for this kind.
	fp, err := svc.Firebase.Projects.Get("projects/" + projectID).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "firebase project %s unreadable after destroy -- the detach contract expects it to remain", projectID)
	}
	if fp.State != "ACTIVE" {
		return errors.Errorf("firebase project %s is in state %q after destroy, want ACTIVE (enablement is permanent)", projectID, fp.State)
	}
	return nil
}
