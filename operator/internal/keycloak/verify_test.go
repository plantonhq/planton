package keycloak

import (
	"strings"
	"testing"
)

// The seeded-admin collision verdict's law: a local holder is a precondition,
// never the collision. Failed is reserved for the pass that PROVED a
// collision (a local holder plus import failures from the user sync); a
// local holder with a clean import, or on the arm that cannot read the
// directory, is an advisory that never fails the manifest. The lab caught the
// old shape live -- a clean 222-user import marked "verification failed"
// because a local admin merely existed -- and this pins the class shut.
func TestSeededAdminCollisionVerdict(t *testing.T) {
	const email = "admin@example.internal"

	cases := []struct {
		name         string
		localHolder  bool
		evidence     collisionEvidence
		wantVerdict  Verdict
		wantInMsg    []string
		wantNotInMsg []string
	}{
		{
			name:        "no local holder passes",
			localHolder: false,
			wantVerdict: VerdictPassed,
			wantInMsg:   []string{"no local user holds " + email},
		},
		{
			name:        "local holder with a clean import is advisory, never Failed",
			localHolder: true,
			evidence:    collisionEvidence{userImportFailures: 0},
			wantVerdict: VerdictUnknown,
			wantInMsg:   []string{"imported every directory user cleanly", "bootstrap.admins"},
		},
		{
			name:        "local holder with import failures is the proven collision",
			localHolder: true,
			evidence:    collisionEvidence{userImportFailures: 1},
			wantVerdict: VerdictFailed,
			wantInMsg:   []string{"1 failed import", "bootstrap.admins"},
		},
		{
			name:        "local holder on the brokered arm is advisory: nothing is readable before a sign-in",
			localHolder: true,
			evidence:    collisionEvidence{directoryUnbrowsable: true},
			wantVerdict: VerdictUnknown,
			wantInMsg:   []string{"first brokered sign-in", "bootstrap.admins"},
		},
		{
			name:         "import failures without a local holder are not this check's finding",
			localHolder:  false,
			evidence:     collisionEvidence{userImportFailures: 3},
			wantVerdict:  VerdictPassed,
			wantNotInMsg: []string{"failed import"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := seededAdminCollisionVerdict(email, tc.localHolder, tc.evidence)
			if got.Name != "seededAdminCollision" {
				t.Fatalf("check name = %q", got.Name)
			}
			if got.Verdict != tc.wantVerdict {
				t.Fatalf("verdict = %s, want %s (%s)", got.Verdict, tc.wantVerdict, got.Message)
			}
			for _, want := range tc.wantInMsg {
				if !strings.Contains(got.Message, want) {
					t.Errorf("message lacks %q: %s", want, got.Message)
				}
			}
			for _, notWant := range tc.wantNotInMsg {
				if strings.Contains(got.Message, notWant) {
					t.Errorf("message must not carry %q: %s", notWant, got.Message)
				}
			}
		})
	}
}

// Every branch that advises names the same remedy sentence: an adopter who
// reads the verdict in the console and again in kubectl sees one instruction.
func TestSeededAdminCollisionRemedyIsOneSentence(t *testing.T) {
	branches := []Check{
		seededAdminCollisionVerdict("a@b", true, collisionEvidence{}),
		seededAdminCollisionVerdict("a@b", true, collisionEvidence{userImportFailures: 2}),
		seededAdminCollisionVerdict("a@b", true, collisionEvidence{directoryUnbrowsable: true}),
	}
	for _, check := range branches {
		if !strings.HasSuffix(check.Message, seededAdminCollisionRemedy) {
			t.Errorf("%s verdict does not end with the shared remedy: %s", check.Verdict, check.Message)
		}
	}
}
