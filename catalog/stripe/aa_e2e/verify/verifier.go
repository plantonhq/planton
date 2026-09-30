// Package verify holds one verifier per Stripe kind: how to read the object a module created,
// and what its destroy must leave behind.
package verify

import (
	"net/url"
	"sort"

	"github.com/pkg/errors"
)

// ResourceChecker reads Stripe objects by their API path, which names the API version first
// ("v1/prices/price_1", "v2/core/event_destinations/ed_1").
type ResourceChecker interface {
	// ReadResource returns the object's JSON body and whether it exists.
	ReadResource(path string) (map[string]interface{}, bool, error)
}

// Verifier checks one kind's object after deploy and after destroy.
type Verifier interface {
	VerifyExists(checker ResourceChecker, id string) error
	// VerifyDestroyed asserts what the kind's destroy leaves, which is not always absence: Stripe
	// never deletes a configuration, product, price, feature, promotion code, shipping rate, tax
	// rate, payment link or meter (destroy deactivates it), and the provider's destroy of a
	// payment-method domain makes no call at all.
	VerifyDestroyed(checker ResourceChecker, id string) error
}

// deletedVerifier is for a kind whose destroy deletes the object: it must exist after deploy
// and answer 404 after destroy.
type deletedVerifier struct {
	component string
	path      string // the API collection, e.g. "v1/webhook_endpoints"
}

func (v *deletedVerifier) VerifyExists(checker ResourceChecker, id string) error {
	_, exists, err := checker.ReadResource(v.path + "/" + url.PathEscape(id))
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s after deploy", v.component, id)
	}
	if !exists {
		return errors.Errorf("%s: %s not found after deploy", v.component, id)
	}
	return nil
}

func (v *deletedVerifier) VerifyDestroyed(checker ResourceChecker, id string) error {
	_, exists, err := checker.ReadResource(v.path + "/" + url.PathEscape(id))
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s after destroy", v.component, id)
	}
	if exists {
		return errors.Errorf("%s: %s still exists after destroy, which deletes it", v.component, id)
	}
	return nil
}

// deactivatedVerifier is for a kind whose destroy deactivates the object: it must be active
// after deploy, and still readable and inactive after destroy. Asserting absence would fail
// every honest run. Most objects report it as active true or false; a billing meter reports it
// as status "active" or "inactive" (byStatus).
type deactivatedVerifier struct {
	component string
	path      string // the API collection, e.g. "v1/billing_portal/configurations"
	byStatus  bool
}

func (v *deactivatedVerifier) VerifyExists(checker ResourceChecker, id string) error {
	return v.requireActive(checker, id, true, "after deploy")
}

func (v *deactivatedVerifier) VerifyDestroyed(checker ResourceChecker, id string) error {
	return v.requireActive(checker, id, false, "after destroy, which deactivates it and Stripe keeps it")
}

func (v *deactivatedVerifier) requireActive(checker ResourceChecker, id string, want bool, when string) error {
	object, exists, err := checker.ReadResource(v.path + "/" + url.PathEscape(id))
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s %s", v.component, id, when)
	}
	if !exists {
		return errors.Errorf("%s: %s not found %s", v.component, id, when)
	}
	if v.byStatus {
		wantStatus := "inactive"
		if want {
			wantStatus = "active"
		}
		if status, _ := object["status"].(string); status != wantStatus {
			return errors.Errorf("%s: %s reads status=%q %s, want %q", v.component, id, status, when, wantStatus)
		}
		return nil
	}
	if active, _ := object["active"].(bool); active != want {
		return errors.Errorf("%s: %s reads active=%t %s, want %t", v.component, id, active, when, want)
	}
	return nil
}

// forgottenVerifier is for a kind whose destroy only removes the object from state: it must
// exist after deploy and STILL exist after destroy, which proves the GUIDE's word that Planton
// forgets the object and Stripe keeps it. Asserting absence would fail every honest run, and
// accepting absence would hide a provider change that started deleting.
type forgottenVerifier struct {
	component string
	path      string // the API collection, e.g. "v1/payment_method_domains"
}

func (v *forgottenVerifier) VerifyExists(checker ResourceChecker, id string) error {
	return v.requirePresent(checker, id, "after deploy")
}

func (v *forgottenVerifier) VerifyDestroyed(checker ResourceChecker, id string) error {
	return v.requirePresent(checker, id, "after destroy, which only forgets it and Stripe keeps it")
}

func (v *forgottenVerifier) requirePresent(checker ResourceChecker, id, when string) error {
	_, exists, err := checker.ReadResource(v.path + "/" + url.PathEscape(id))
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s %s", v.component, id, when)
	}
	if !exists {
		return errors.Errorf("%s: %s not found %s", v.component, id, when)
	}
	return nil
}

// ChildVerifier is implemented by the verifier of a kind that folds child objects (a meter's
// alerts, a product's feature links, a Radar list's items). The module reports the children as a
// map output from each child's key to its Stripe id, and each child's destroy truth is proven
// like the parent's: a folded child is not proven just because its parent is.
type ChildVerifier interface {
	// ChildOutput is the name of the map output that holds the children's ids.
	ChildOutput() string
	VerifyChildrenExist(checker ResourceChecker, parentID string, childIDs map[string]string) error
	VerifyChildrenDestroyed(checker ResourceChecker, parentID string, childIDs map[string]string) error
}

// childFate is what a parent's destroy leaves of a folded child.
type childFate int

const (
	// childDeleted children answer 404 after destroy.
	childDeleted childFate = iota
	// childForgotten children are still present after destroy: the provider only forgets them.
	childForgotten
)

// withChildren adds a folded child's checks to a parent's verifier.
type withChildren struct {
	Verifier
	component string
	output    string
	childPath func(parentID, childID string) string
	fate      childFate
}

func (v *withChildren) ChildOutput() string { return v.output }

func (v *withChildren) VerifyChildrenExist(checker ResourceChecker, parentID string, childIDs map[string]string) error {
	for _, key := range sortedKeys(childIDs) {
		if err := v.requireChild(checker, parentID, key, childIDs[key], true, "after deploy"); err != nil {
			return err
		}
	}
	return nil
}

func (v *withChildren) VerifyChildrenDestroyed(checker ResourceChecker, parentID string, childIDs map[string]string) error {
	want, when := false, "after destroy, which deletes it"
	if v.fate == childForgotten {
		want, when = true, "after destroy, which only forgets it and Stripe keeps it"
	}
	for _, key := range sortedKeys(childIDs) {
		if err := v.requireChild(checker, parentID, key, childIDs[key], want, when); err != nil {
			return err
		}
	}
	return nil
}

func (v *withChildren) requireChild(checker ResourceChecker, parentID, key, childID string, wantPresent bool, when string) error {
	_, exists, err := checker.ReadResource(v.childPath(parentID, childID))
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s %q (%s) %s", v.component, v.output, key, childID, when)
	}
	if exists != wantPresent {
		state := "not found"
		if exists {
			state = "still exists"
		}
		return errors.Errorf("%s: %s %q (%s) %s %s", v.component, v.output, key, childID, state, when)
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// verifiers maps each Stripe component directory to its verifier.
var verifiers = map[string]Verifier{
	"stripewebhookendpoint":            &deletedVerifier{component: "stripewebhookendpoint", path: "v1/webhook_endpoints"},
	"stripeeventdestination":           &deletedVerifier{component: "stripeeventdestination", path: "v2/core/event_destinations"},
	"stripebillingportalconfiguration": &deactivatedVerifier{component: "stripebillingportalconfiguration", path: "v1/billing_portal/configurations"},
	"stripepaymentmethodconfiguration": &deactivatedVerifier{component: "stripepaymentmethodconfiguration", path: "v1/payment_method_configurations"},
	"stripepaymentmethoddomain":        &forgottenVerifier{component: "stripepaymentmethoddomain", path: "v1/payment_method_domains"},
	"striperadarvaluelist": &withChildren{
		Verifier:  &deletedVerifier{component: "striperadarvaluelist", path: "v1/radar/value_lists"},
		component: "striperadarvaluelist",
		output:    "item_ids",
		childPath: func(_, id string) string { return "v1/radar/value_list_items/" + url.PathEscape(id) },
		fate:      childDeleted,
	},
	"stripeproduct": &withChildren{
		Verifier:  &deactivatedVerifier{component: "stripeproduct", path: "v1/products"},
		component: "stripeproduct",
		output:    "product_feature_ids",
		childPath: func(product, id string) string {
			return "v1/products/" + url.PathEscape(product) + "/features/" + url.PathEscape(id)
		},
		fate: childDeleted,
	},
	"stripeprice":              &deactivatedVerifier{component: "stripeprice", path: "v1/prices"},
	"stripeentitlementfeature": &deactivatedVerifier{component: "stripeentitlementfeature", path: "v1/entitlements/features"},
	"stripecoupon":             &deletedVerifier{component: "stripecoupon", path: "v1/coupons"},
	"stripepromotioncode":      &deactivatedVerifier{component: "stripepromotioncode", path: "v1/promotion_codes"},
	"stripeshippingrate":       &deactivatedVerifier{component: "stripeshippingrate", path: "v1/shipping_rates"},
	"stripetaxrate":            &deactivatedVerifier{component: "stripetaxrate", path: "v1/tax_rates"},
	"stripetaxregistration":    &forgottenVerifier{component: "stripetaxregistration", path: "v1/tax/registrations"},
	"stripepaymentlink":        &deactivatedVerifier{component: "stripepaymentlink", path: "v1/payment_links"},
	"stripebillingmeter": &withChildren{
		Verifier:  &deactivatedVerifier{component: "stripebillingmeter", path: "v1/billing/meters", byStatus: true},
		component: "stripebillingmeter",
		output:    "alert_ids",
		childPath: func(_, id string) string { return "v1/billing/alerts/" + url.PathEscape(id) },
		fate:      childForgotten,
	},
}

// GetVerifier returns the verifier for a component, or an error naming the unknown component.
func GetVerifier(component string) (Verifier, error) {
	v, ok := verifiers[component]
	if !ok {
		return nil, errors.Errorf("no Stripe verifier registered for component %q", component)
	}
	return v, nil
}
