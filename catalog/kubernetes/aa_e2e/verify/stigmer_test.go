package verify

import (
	"errors"
	"strings"
	"testing"
)

// Each refusal class is pinned on the words both engines write (the module's
// preflight and its OpenTofu twin), and any other failure is never taken for
// it.
func TestStigmerRefusalsArePinnedOnTheirWords(t *testing.T) {
	refusals := map[string]string{
		"stigmer-downgrade": "observed: Stigmer stigmer/stigmer records Stigmer v0.3.2 on this cluster, and this KubernetesStigmer installs 0.3.1\n" +
			"meaning: Stigmer does not move from v0.3.2 to 0.3.1 in place: its server refuses to start on a database a newer release migrated\n" +
			"next step: set spec.version back to 0.3.2 or later",
		"stigmer-keys-missing": "observed: database.bundled.recover_from restores archive stigmer-postgres, and the keys Secret stigmer-keys is not in namespace stigmer\n" +
			"meaning: the archive is encrypted with keys this namespace does not hold\n" +
			"next step: apply the keys Secret you saved",
		"stigmer-operator-owned-elsewhere": "observed: the Stigmer definition stigmers.stigmer.ai on the cluster belongs to the Helm release other in namespace elsewhere, and this resource installs the operator as release stigmer-operator in namespace stigmer-system\n" +
			"meaning: Helm refuses to take over a definition another release owns, and Stigmer runs one operator per cluster\n" +
			"next step: remove the other Stigmer operator install first",
		"stigmer-prerequisite-orphaned": "observed: CloudNativePG's definition clusters.postgresql.cnpg.io is on the cluster, installed by the Helm release cnpg in namespace cnpg-system, and no controller runs it: that release has no Deployment left\n" +
			"meaning: Stigmer's operator trusts a CloudNativePG someone else installed and installs nothing over it\n" +
			"next step: install CloudNativePG again",
	}
	for class, text := range refusals {
		if err := verifyStigmerRefusal(class, errors.New("engine output\n"+text)); err != nil {
			t.Errorf("%s: its own words were not recognized: %v", class, err)
		}
		for other, otherText := range refusals {
			if other != class && verifyStigmerRefusal(class, errors.New(otherText)) == nil {
				t.Errorf("%s was recognized in the words of %s", class, other)
			}
		}
	}
	wrapped := "Error: Resource precondition failed\n  observed: the Stigmer definition stigmers.stigmer.ai on the cluster belongs\n  to the Helm release stigmer-operator in namespace stigmer-system\n  meaning: Helm refuses\n  next step: set it back"
	if err := verifyStigmerRefusal("stigmer-operator-owned-elsewhere", errors.New(wrapped)); err != nil {
		t.Errorf("OpenTofu's 80-column wrapping must not hide the words: %v", err)
	}
	if err := verifyStigmerRefusal("stigmer-downgrade", errors.New("exit status 1")); err == nil || !strings.Contains(err.Error(), "not with the stigmer-downgrade refusal") {
		t.Errorf("a bare failure must never pass as a refusal, got %v", err)
	}
	if err := verifyStigmerRefusal("stigmer-downgrade", nil); err == nil {
		t.Error("a deploy that succeeded must never pass as a refusal")
	}
}
