// Tests for the secret naming law: the id each stored variable gets, and the
// two shapes it refuses before any resource exists. What they pin: the
// Cloud Run kinds' ids exactly as their existing stacks hold them, every
// kind's resource never sharing a secret with another's, an unnamed
// container still getting a distinct id, a placement without containers
// dropping that segment, a name Secret Manager would refuse made legal, a
// collision or an overlong id failing with a sentence naming the variables
// and the fix, and the grant naming the right fallback identity.
package envsecrets

import (
	"strings"
	"testing"
)

func servicePlacement() Placement {
	return Placement{Kind: KindService, Location: "us-central1", Resource: "payments-api"}
}

func functionPlacement() Placement {
	return Placement{Kind: KindFunction, Location: "us-central1", Resource: "payments-hook", NoContainers: true}
}

func TestSecretIDNamesKindRegionResourceContainerVariable(t *testing.T) {
	got := SecretID(servicePlacement(), Variable{ContainerIndex: 0, Container: "app", Name: "STRIPE_KEY"})
	if want := "run_us-central1_payments-api_app_STRIPE_KEY"; got != want {
		t.Fatalf("SecretID = %q, want %q", got, want)
	}
}

// The ids existing Cloud Run stacks hold, one per kind: a changed byte would
// make the module look for a secret that does not exist and create a second.
func TestTheCloudRunKindsIDsAreTheOnesTheirStacksHold(t *testing.T) {
	for _, tc := range []struct {
		kind, location, resource string
		variable                 Variable
		want                     string
	}{
		{KindService, "global", "edge-api", Variable{ContainerIndex: 0, Container: "app", Name: "db.password"}, "run_global_edge-api_app_db-password"},
		{KindJob, "europe-west1", "nightly-etl", Variable{ContainerIndex: 1, Name: "TOKEN"}, "runjob_europe-west1_nightly-etl_c1_TOKEN"},
		{KindWorkerPool, "us-east4", "queue-consumer", Variable{ContainerIndex: 0, Container: "worker", Name: "API_KEY"}, "runpool_us-east4_queue-consumer_worker_API_KEY"},
	} {
		got := SecretID(Placement{Kind: tc.kind, Location: tc.location, Resource: tc.resource}, tc.variable)
		if got != tc.want {
			t.Errorf("%s: SecretID = %q, want %q", tc.kind, got, tc.want)
		}
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

func TestSecretIDWithoutContainersNamesKindLocationResourceVariable(t *testing.T) {
	got := SecretID(functionPlacement(), Variable{ContainerIndex: 3, Container: "ignored", Name: "db.password"})
	if want := "function_us-central1_payments-hook_db-password"; got != want {
		t.Fatalf("SecretID = %q, want %q", got, want)
	}
	if key := KeyOf(functionPlacement(), Variable{ContainerIndex: 3, Name: "X"}); key != (Key{Name: "X"}) {
		t.Fatalf("a placement without containers keys by name alone; got %+v", key)
	}
}

func TestNoTwoKindsWithOneNameEverShareASecret(t *testing.T) {
	variable := Variable{Container: "app", Name: "TOKEN"}
	seen := map[string]string{}
	for _, kind := range []string{KindService, KindJob, KindWorkerPool, KindFunction, KindAgentEngine, KindWorkflow, KindComposer} {
		placement := servicePlacement()
		placement.Kind = kind
		id := SecretID(placement, variable)
		if other, taken := seen[id]; taken {
			t.Fatalf("kinds %q and %q with the same name and region got the same secret id %q", other, kind, id)
		}
		seen[id] = kind
		if strings.Contains(kind, "_") {
			t.Fatalf("kind %q carries the underscore the id is split by", kind)
		}
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

func TestSecretIDsWithoutContainersRefusesACollisionWithoutNamingAContainer(t *testing.T) {
	_, err := SecretIDs(functionPlacement(), []Variable{{Name: "db.password"}, {Name: "db-password"}})
	if err == nil || !strings.Contains(err.Error(), "rename one of them") || strings.Contains(err.Error(), "container") {
		t.Fatalf("SecretIDs did not refuse the collision in the placement's own terms: %v", err)
	}
}

func TestSecretIDsRefusesAnIDOverSecretManagersLimit(t *testing.T) {
	_, err := SecretIDs(servicePlacement(), []Variable{{Container: "app", Name: strings.Repeat("A", 250)}})
	if err == nil || !strings.Contains(err.Error(), "shorten the variable or container name") {
		t.Fatalf("SecretIDs did not refuse an overlong id with its fix: %v", err)
	}
	_, err = SecretIDs(functionPlacement(), []Variable{{Name: strings.Repeat("A", 250)}})
	if err == nil || !strings.Contains(err.Error(), "shorten the variable or resource name") {
		t.Fatalf("SecretIDs did not refuse an overlong containerless id with its fix: %v", err)
	}
}

func TestSecretIDsAcceptsAnIDAtExactlySecretManagersLimit(t *testing.T) {
	prefix := len("function_us-central1_payments-hook_")
	if _, err := SecretIDs(functionPlacement(), []Variable{{Name: strings.Repeat("A", secretIDMaxLen-prefix)}}); err != nil {
		t.Fatalf("SecretIDs refused a %d-character id: %v", secretIDMaxLen, err)
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

func TestRuntimeMemberNamesTheRuntimeIdentityOrTheFallback(t *testing.T) {
	placement := servicePlacement()
	placement.RuntimeServiceAccount = "svc@p.iam.gserviceaccount.com"
	if got := RuntimeMember(placement, "123"); got != "serviceAccount:svc@p.iam.gserviceaccount.com" {
		t.Fatalf("RuntimeMember with an identity = %q", got)
	}
	if got := RuntimeMember(servicePlacement(), "123"); got != "serviceAccount:123-compute@developer.gserviceaccount.com" {
		t.Fatalf("RuntimeMember without an identity = %q", got)
	}
	agent := Placement{Kind: KindAgentEngine, DefaultAccount: ReasoningEngineServiceAgent}
	if got := RuntimeMember(agent, "123"); got != "serviceAccount:service-123@gcp-sa-aiplatform-re.iam.gserviceaccount.com" {
		t.Fatalf("RuntimeMember for an agent without an identity = %q", got)
	}
}
