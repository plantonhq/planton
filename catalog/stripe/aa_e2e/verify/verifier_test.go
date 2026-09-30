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
	for _, component := range []string{
		"stripewebhookendpoint", "stripeeventdestination", "stripebillingportalconfiguration", "stripepaymentmethodconfiguration",
		"stripepaymentmethoddomain", "striperadarvaluelist", "stripeproduct", "stripeprice", "stripeentitlementfeature",
	} {
		if _, err := GetVerifier(component); err != nil {
			t.Error(err)
		}
	}
	if _, err := GetVerifier("stripeunknown"); err == nil {
		t.Error("an unknown component must be refused")
	}
}

func TestDeletedKinds_ExistThenGone(t *testing.T) {
	for component, path := range map[string]string{
		"stripewebhookendpoint":  "v1/webhook_endpoints/obj_1",
		"stripeeventdestination": "v2/core/event_destinations/obj_1",
		"striperadarvaluelist":   "v1/radar/value_lists/obj_1",
	} {
		v, _ := GetVerifier(component)
		checker := &recordingChecker{objects: map[string]map[string]interface{}{path: {"id": "obj_1"}}}
		if err := v.VerifyExists(checker, "obj_1"); err != nil {
			t.Fatalf("%s: %v", component, err)
		}
		if checker.asked[0] != path {
			t.Errorf("%s: read %q, want %q", component, checker.asked[0], path)
		}
		if err := v.VerifyDestroyed(checker, "obj_1"); err == nil || !strings.Contains(err.Error(), "still exists after destroy") {
			t.Errorf("%s: an object that survives destroy must fail, got %v", component, err)
		}
		delete(checker.objects, path)
		if err := v.VerifyDestroyed(checker, "obj_1"); err != nil {
			t.Errorf("%s: a deleted object passes, got %v", component, err)
		}
	}
}

// A payment-method domain's destroy makes no API call, so the domain must still be there: a run
// where it vanished means the provider started deleting and the GUIDE's word is stale.
func TestPaymentMethodDomain_StillPresentAfterDestroy(t *testing.T) {
	v, _ := GetVerifier("stripepaymentmethoddomain")
	path := "v1/payment_method_domains/pmd_1"
	checker := &recordingChecker{objects: map[string]map[string]interface{}{path: {"id": "pmd_1", "enabled": true}}}
	if err := v.VerifyExists(checker, "pmd_1"); err != nil {
		t.Fatal(err)
	}
	if checker.asked[0] != path {
		t.Errorf("read %q, want %q", checker.asked[0], path)
	}
	if err := v.VerifyDestroyed(checker, "pmd_1"); err != nil {
		t.Errorf("a domain Stripe keeps after destroy passes, got %v", err)
	}
	delete(checker.objects, path)
	if err := v.VerifyDestroyed(checker, "pmd_1"); err == nil || !strings.Contains(err.Error(), "only forgets it and Stripe keeps it") {
		t.Errorf("a domain gone after destroy must fail, got %v", err)
	}
}

func TestConfigurations_ActiveThenDeactivatedNotGone(t *testing.T) {
	for component, path := range map[string]string{
		"stripebillingportalconfiguration": "v1/billing_portal/configurations/cfg_1",
		"stripepaymentmethodconfiguration": "v1/payment_method_configurations/cfg_1",
		"stripeproduct":                    "v1/products/cfg_1",
		"stripeprice":                      "v1/prices/cfg_1",
		"stripeentitlementfeature":         "v1/entitlements/features/cfg_1",
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
