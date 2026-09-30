package verify

import (
	"strings"
	"testing"
)

// recordingChecker answers from a fixed table and records the paths it was asked for, so the
// tests pin the exact API path each id becomes.
type recordingChecker struct {
	objects map[string]map[string]interface{}
	asked   []string
}

func (c *recordingChecker) ReadResource(path string) (map[string]interface{}, bool, error) {
	c.asked = append(c.asked, path)
	object, ok := c.objects[path]
	return object, ok, nil
}

func TestEveryStripeKindHasAVerifier(t *testing.T) {
	for _, component := range []string{"stripewebhookendpoint", "stripebillingportalconfiguration", "stripepaymentmethodconfiguration"} {
		if _, err := GetVerifier(component); err != nil {
			t.Error(err)
		}
	}
	if _, err := GetVerifier("stripeunknown"); err == nil {
		t.Error("an unknown component must be refused")
	}
}

func TestWebhookEndpoint_ExistsThenGone(t *testing.T) {
	v, _ := GetVerifier("stripewebhookendpoint")
	checker := &recordingChecker{objects: map[string]map[string]interface{}{"webhook_endpoints/we_1": {"id": "we_1"}}}
	if err := v.VerifyExists(checker, "we_1"); err != nil {
		t.Fatal(err)
	}
	if checker.asked[0] != "webhook_endpoints/we_1" {
		t.Errorf("read %q, want webhook_endpoints/we_1", checker.asked[0])
	}
	if err := v.VerifyDestroyed(checker, "we_1"); err == nil || !strings.Contains(err.Error(), "still exists after destroy") {
		t.Errorf("an endpoint that survives destroy must fail, got %v", err)
	}
	delete(checker.objects, "webhook_endpoints/we_1")
	if err := v.VerifyDestroyed(checker, "we_1"); err != nil {
		t.Errorf("a deleted endpoint passes, got %v", err)
	}
}

func TestConfigurations_ActiveThenDeactivatedNotGone(t *testing.T) {
	for component, path := range map[string]string{
		"stripebillingportalconfiguration": "billing_portal/configurations/cfg_1",
		"stripepaymentmethodconfiguration": "payment_method_configurations/cfg_1",
	} {
		v, _ := GetVerifier(component)
		checker := &recordingChecker{objects: map[string]map[string]interface{}{path: {"active": true}}}
		if err := v.VerifyExists(checker, "cfg_1"); err != nil {
			t.Fatalf("%s: %v", component, err)
		}
		if checker.asked[0] != path {
			t.Errorf("%s: read %q, want %q", component, checker.asked[0], path)
		}
		if err := v.VerifyDestroyed(checker, "cfg_1"); err == nil || !strings.Contains(err.Error(), "active=true") {
			t.Errorf("%s: a configuration still active after destroy must fail, got %v", component, err)
		}
		checker.objects[path] = map[string]interface{}{"active": false}
		if err := v.VerifyDestroyed(checker, "cfg_1"); err != nil {
			t.Errorf("%s: a deactivated configuration passes, got %v", component, err)
		}
		delete(checker.objects, path)
		if err := v.VerifyDestroyed(checker, "cfg_1"); err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("%s: a configuration Stripe no longer has must fail -- Stripe keeps deactivated ones, got %v", component, err)
		}
	}
}
