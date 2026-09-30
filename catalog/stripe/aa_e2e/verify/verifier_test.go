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
		"stripecoupon", "stripepromotioncode", "stripeshippingrate", "stripetaxrate", "stripebillingmeter", "stripepaymentlink",
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
		"stripecoupon":           "v1/coupons/obj_1",
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
		"stripepromotioncode":              "v1/promotion_codes/cfg_1",
		"stripeshippingrate":               "v1/shipping_rates/cfg_1",
		"stripetaxrate":                    "v1/tax_rates/cfg_1",
		"stripepaymentlink":                "v1/payment_links/cfg_1",
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

// A billing meter reports deactivation as status, not active: reading active would pass a live
// meter after destroy (active is absent, so false) and fail a live meter after deploy.
func TestBillingMeter_StatusActiveThenInactive(t *testing.T) {
	v, _ := GetVerifier("stripebillingmeter")
	path := "v1/billing/meters/mtr_1"
	checker := &recordingChecker{objects: map[string]map[string]interface{}{path: {"status": "active"}}}
	if err := v.VerifyExists(checker, "mtr_1"); err != nil {
		t.Fatal(err)
	}
	if checker.asked[0] != path {
		t.Errorf("read %q, want %q", checker.asked[0], path)
	}
	if err := v.VerifyDestroyed(checker, "mtr_1"); err == nil || !strings.Contains(err.Error(), `status="active"`) {
		t.Errorf("a meter still active after destroy must fail, got %v", err)
	}
	checker.objects[path] = map[string]interface{}{"status": "inactive"}
	if err := v.VerifyDestroyed(checker, "mtr_1"); err != nil {
		t.Errorf("a deactivated meter passes, got %v", err)
	}
	if err := v.VerifyExists(checker, "mtr_1"); err == nil {
		t.Error("an inactive meter after deploy must fail")
	}
	delete(checker.objects, path)
	if err := v.VerifyDestroyed(checker, "mtr_1"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("a meter Stripe no longer has must fail -- Stripe keeps deactivated ones, got %v", err)
	}
}

func childVerifier(t *testing.T, component string) ChildVerifier {
	t.Helper()
	v, _ := GetVerifier(component)
	cv, ok := v.(ChildVerifier)
	if !ok {
		t.Fatalf("%s folds children and must verify them", component)
	}
	return cv
}

// Folded children are proven like their parent: a Radar list's items and a product's feature
// links are deleted with it, and a meter's alerts are only forgotten.
func TestFoldedChildren_DeletedKinds(t *testing.T) {
	for component, tc := range map[string]struct {
		output string
		path   string
	}{
		"striperadarvaluelist": {"item_ids", "v1/radar/value_list_items/rsli_1"},
		"stripeproduct":        {"product_feature_ids", "v1/products/parent_1/features/prodft_1"},
	} {
		cv := childVerifier(t, component)
		if cv.ChildOutput() != tc.output {
			t.Errorf("%s: child output %q, want %q", component, cv.ChildOutput(), tc.output)
		}
		id := tc.path[strings.LastIndex(tc.path, "/")+1:]
		children := map[string]string{"key": id}
		checker := &recordingChecker{objects: map[string]map[string]interface{}{tc.path: {"id": id}}}
		if err := cv.VerifyChildrenExist(checker, "parent_1", children); err != nil {
			t.Fatalf("%s: %v", component, err)
		}
		if checker.asked[0] != tc.path {
			t.Errorf("%s: read %q, want %q", component, checker.asked[0], tc.path)
		}
		if err := cv.VerifyChildrenDestroyed(checker, "parent_1", children); err == nil || !strings.Contains(err.Error(), "still exists after destroy, which deletes it") {
			t.Errorf("%s: a child that survives destroy must fail, got %v", component, err)
		}
		delete(checker.objects, tc.path)
		if err := cv.VerifyChildrenDestroyed(checker, "parent_1", children); err != nil {
			t.Errorf("%s: a deleted child passes, got %v", component, err)
		}
		if err := cv.VerifyChildrenExist(checker, "parent_1", children); err == nil || !strings.Contains(err.Error(), "not found after deploy") {
			t.Errorf("%s: a child missing after deploy must fail, got %v", component, err)
		}
	}
}

func TestFoldedChildren_MeterAlertsAreForgotten(t *testing.T) {
	cv := childVerifier(t, "stripebillingmeter")
	if cv.ChildOutput() != "alert_ids" {
		t.Errorf("child output %q, want alert_ids", cv.ChildOutput())
	}
	path := "v1/billing/alerts/alrt_1"
	children := map[string]string{"10k requests": "alrt_1"}
	checker := &recordingChecker{objects: map[string]map[string]interface{}{path: {"id": "alrt_1", "status": "active"}}}
	if err := cv.VerifyChildrenExist(checker, "mtr_1", children); err != nil {
		t.Fatal(err)
	}
	if checker.asked[0] != path {
		t.Errorf("read %q, want %q", checker.asked[0], path)
	}
	if err := cv.VerifyChildrenDestroyed(checker, "mtr_1", children); err != nil {
		t.Errorf("an alert Stripe keeps after destroy passes, got %v", err)
	}
	delete(checker.objects, path)
	if err := cv.VerifyChildrenDestroyed(checker, "mtr_1", children); err == nil || !strings.Contains(err.Error(), "only forgets it and Stripe keeps it") {
		t.Errorf("an alert gone after destroy must fail -- the GUIDE says Stripe keeps it, got %v", err)
	}
}

// A kind with no folded children must not claim any: the harness would otherwise look for an
// output the module never writes.
func TestOnlyFoldingKindsVerifyChildren(t *testing.T) {
	folding := map[string]bool{"striperadarvaluelist": true, "stripeproduct": true, "stripebillingmeter": true}
	for component, v := range verifiers {
		if _, ok := v.(ChildVerifier); ok != folding[component] {
			t.Errorf("%s: child verifier %t, want %t", component, ok, folding[component])
		}
	}
}
