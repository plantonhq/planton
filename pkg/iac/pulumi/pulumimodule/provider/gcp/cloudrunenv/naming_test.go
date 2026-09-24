// Tests for the secret naming law: the id each stored variable gets, and the
// two shapes it refuses before any resource exists. What they pin: a service,
// a job, and a worker pool never share a secret, an unnamed container still gets a distinct
// id, a name Secret Manager would refuse is made legal, and a collision or an
// overlong id fails with a sentence naming the variables and the fix.
package cloudrunenv

import (
	"strings"
	"testing"
)

func servicePlacement() Placement {
	return Placement{Kind: KindService, Region: "us-central1", Resource: "payments-api"}
}

func TestSecretIDNamesKindRegionResourceContainerVariable(t *testing.T) {
	got := SecretID(servicePlacement(), Variable{ContainerIndex: 0, Container: "app", Name: "STRIPE_KEY"})
	if want := "run_us-central1_payments-api_app_STRIPE_KEY"; got != want {
		t.Fatalf("SecretID = %q, want %q", got, want)
	}
}

func TestSecretIDNamesAnUnnamedContainerByItsPosition(t *testing.T) {
	got := SecretID(servicePlacement(), Variable{ContainerIndex: 1, Name: "TOKEN"})
	if want := "run_us-central1_payments-api_c1_TOKEN"; got != want {
		t.Fatalf("SecretID = %q, want %q", got, want)
	}
}

func TestSecretIDMakesADottedVariableNameLegal(t *testing.T) {
	got := SecretID(servicePlacement(), Variable{Container: "app", Name: "db.password"})
	if want := "run_us-central1_payments-api_app_db-password"; got != want {
		t.Fatalf("SecretID = %q, want %q", got, want)
	}
}

func TestAServiceAJobAndAWorkerPoolWithOneNameNeverShareASecret(t *testing.T) {
	variable := Variable{Container: "app", Name: "TOKEN"}
	seen := map[string]string{}
	for _, kind := range []string{KindService, KindJob, KindWorkerPool} {
		placement := servicePlacement()
		placement.Kind = kind
		id := SecretID(placement, variable)
		if other, taken := seen[id]; taken {
			t.Fatalf("kinds %q and %q with the same name and region got the same secret id %q", other, kind, id)
		}
		seen[id] = kind
	}
}

func TestSecretIDsRefusesTwoVariablesThatCollideAfterSanitizing(t *testing.T) {
	_, err := SecretIDs(servicePlacement(), []Variable{
		{Container: "app", Name: "db.password", Value: "a"},
		{Container: "app", Name: "db-password", Value: "b"},
	})
	if err == nil {
		t.Fatal("SecretIDs accepted two variables that map to one secret id")
	}
	for _, want := range []string{`"db.password"`, `"db-password"`, "rename one of them"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal %q does not name %s", err.Error(), want)
		}
	}
}

func TestSecretIDsRefusesAnIDOverSecretManagersLimit(t *testing.T) {
	_, err := SecretIDs(servicePlacement(), []Variable{{Container: "app", Name: strings.Repeat("A", 250)}})
	if err == nil || !strings.Contains(err.Error(), "shorten the variable or container name") {
		t.Fatalf("SecretIDs did not refuse an overlong id with its fix: %v", err)
	}
}

func TestSecretIDsAddressesEachVariableByContainerAndName(t *testing.T) {
	ids, err := SecretIDs(servicePlacement(), []Variable{
		{ContainerIndex: 0, Container: "app", Name: "TOKEN"},
		{ContainerIndex: 1, Container: "proxy", Name: "TOKEN"},
	})
	if err != nil {
		t.Fatalf("SecretIDs refused distinct containers: %v", err)
	}
	if ids[Key{ContainerIndex: 0, Name: "TOKEN"}] == ids[Key{ContainerIndex: 1, Name: "TOKEN"}] {
		t.Fatal("two containers' variables with one name share a secret id")
	}
}

func TestRuntimeMemberNamesTheRuntimeIdentityOrTheComputeDefault(t *testing.T) {
	if got := RuntimeMember("svc@p.iam.gserviceaccount.com", "123"); got != "serviceAccount:svc@p.iam.gserviceaccount.com" {
		t.Fatalf("RuntimeMember with an identity = %q", got)
	}
	if got := RuntimeMember("", "123"); got != "serviceAccount:123-compute@developer.gserviceaccount.com" {
		t.Fatalf("RuntimeMember without an identity = %q", got)
	}
}
