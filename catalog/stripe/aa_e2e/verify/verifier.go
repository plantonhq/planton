// Package verify holds one verifier per Stripe kind: how to read the object a module created,
// and what its destroy must leave behind.
package verify

import (
	"net/url"
	"sort"
	"strings"

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

// ResourceDeleter deletes a Stripe object by its API path. Only the out-of-band act uses it: it
// deletes an object the way a person in the Dashboard would, behind the engine's back.
type ResourceDeleter interface {
	DeleteResource(path string) error
}

// OutOfBandDeletable is implemented by the verifier of a kind Stripe can delete outright, so the
// out-of-band act (someone deletes the object in the Dashboard) applies to it. A kind whose
// object Stripe only deactivates or forgets has nothing to delete and does not implement it.
type OutOfBandDeletable interface {
	DeleteOutOfBand(client ResourceClient, id string) error
}

// ResourceClient reads and deletes Stripe objects.
type ResourceClient interface {
	ResourceChecker
	ResourceDeleter
}

// SecretVerifier is implemented by the verifier of a kind whose module reports a signing secret,
// which Stripe returns only when it creates the object. The harness checks the secret's shape
// after deploy and, without ever logging it, that a replaced or recreated object came with a new
// one and an object updated in place kept its own.
type SecretVerifier interface {
	// SecretOutput is the name of the output that holds the secret; empty when the kind has none.
	SecretOutput() string
	// SecretRequired reports whether every deploy of the kind must report one (an event
	// destination reports one only when it delivers to a webhook).
	SecretRequired() bool
}

// signingSecretPrefix is how every Stripe signing secret begins.
const signingSecretPrefix = "whsec_"

// CheckSecretShape refuses a reported secret that is not a Stripe signing secret, and a missing
// one the kind requires. It names the output, never the value.
func CheckSecretShape(kind string, sv SecretVerifier, secret string) error {
	switch {
	case secret == "" && sv.SecretRequired():
		return errors.Errorf("%s: output %s is empty after create, but Stripe returns the signing secret at creation", kind, sv.SecretOutput())
	case secret != "" && !strings.HasPrefix(secret, signingSecretPrefix):
		return errors.Errorf("%s: output %s is not a signing secret (it does not begin %s)", kind, sv.SecretOutput(), signingSecretPrefix)
	}
	return nil
}

// deletedVerifier is for a kind whose destroy deletes the object: it must exist after deploy
// and answer 404 after destroy. Stripe can delete such an object outright, so the out-of-band act
// applies to it.
type deletedVerifier struct {
	kind string
	path string // the API collection, e.g. "v1/webhook_endpoints"
	// secretOutput names the module's signing-secret output, when the kind has one.
	secretOutput   string
	secretRequired bool
}

func (v *deletedVerifier) SecretOutput() string { return v.secretOutput }
func (v *deletedVerifier) SecretRequired() bool { return v.secretRequired }

// DeleteOutOfBand deletes the object through Stripe's API and confirms it now reads back absent,
// so the act's later phases prove what the engine does with a missing object.
func (v *deletedVerifier) DeleteOutOfBand(client ResourceClient, id string) error {
	path := v.path + "/" + url.PathEscape(id)
	if err := client.DeleteResource(path); err != nil {
		return errors.Wrapf(err, "%s: deleting %s outside Planton", v.kind, id)
	}
	_, exists, err := client.ReadResource(path)
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s after deleting it outside Planton", v.kind, id)
	}
	if exists {
		return errors.Errorf("%s: %s still exists after Stripe answered its delete", v.kind, id)
	}
	return nil
}

func (v *deletedVerifier) VerifyExists(checker ResourceChecker, id string) error {
	_, exists, err := checker.ReadResource(v.path + "/" + url.PathEscape(id))
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s after deploy", v.kind, id)
	}
	if !exists {
		return errors.Errorf("%s: %s not found after deploy", v.kind, id)
	}
	return nil
}

func (v *deletedVerifier) VerifyDestroyed(checker ResourceChecker, id string) error {
	_, exists, err := checker.ReadResource(v.path + "/" + url.PathEscape(id))
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s after destroy", v.kind, id)
	}
	if exists {
		return errors.Errorf("%s: %s still exists after destroy, which deletes it", v.kind, id)
	}
	return nil
}

// deactivatedVerifier is for a kind whose destroy deactivates the object: it must be active
// after deploy, and still readable and inactive after destroy. Asserting absence would fail
// every honest run. Most objects report it as active true or false; a billing meter reports it
// as status "active" or "inactive" (byStatus).
type deactivatedVerifier struct {
	kind     string
	path     string // the API collection, e.g. "v1/billing_portal/configurations"
	byStatus bool
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
		return errors.Wrapf(err, "%s: reading %s %s", v.kind, id, when)
	}
	if !exists {
		return errors.Errorf("%s: %s not found %s", v.kind, id, when)
	}
	if v.byStatus {
		wantStatus := "inactive"
		if want {
			wantStatus = "active"
		}
		if status, _ := object["status"].(string); status != wantStatus {
			return errors.Errorf("%s: %s reads status=%q %s, want %q", v.kind, id, status, when, wantStatus)
		}
		return nil
	}
	if active, _ := object["active"].(bool); active != want {
		return errors.Errorf("%s: %s reads active=%t %s, want %t", v.kind, id, active, when, want)
	}
	return nil
}

// forgottenVerifier is for a kind whose destroy only removes the object from state: it must
// exist after deploy and STILL exist after destroy, which proves the GUIDE's word that Planton
// forgets the object and Stripe keeps it. Asserting absence would fail every honest run, and
// accepting absence would hide a provider change that started deleting.
type forgottenVerifier struct {
	kind string
	path string // the API collection, e.g. "v1/payment_method_domains"
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
		return errors.Wrapf(err, "%s: reading %s %s", v.kind, id, when)
	}
	if !exists {
		return errors.Errorf("%s: %s not found %s", v.kind, id, when)
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
	kind      string
	output    string
	childPath func(parentID, childID string) string
	fate      childFate
}

func (v *withChildren) ChildOutput() string { return v.output }

// DeleteOutOfBand deletes the parent the way its own verifier does; Stripe deletes a deleted
// parent's children with it (a Radar list's items), and the recovery recreates both.
func (v *withChildren) DeleteOutOfBand(client ResourceClient, id string) error {
	d, ok := v.Verifier.(OutOfBandDeletable)
	if !ok {
		return errors.Errorf("%s: Stripe keeps this object after its destroy, so it cannot be deleted outside Planton", v.kind)
	}
	return d.DeleteOutOfBand(client, id)
}

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
		return errors.Wrapf(err, "%s: reading %s %q (%s) %s", v.kind, v.output, key, childID, when)
	}
	if exists != wantPresent {
		state := "not found"
		if exists {
			state = "still exists"
		}
		return errors.Errorf("%s: %s %q (%s) %s %s", v.kind, v.output, key, childID, state, when)
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
	"stripewebhookendpoint":            &deletedVerifier{kind: "stripewebhookendpoint", path: "v1/webhook_endpoints", secretOutput: "secret", secretRequired: true},
	"stripeeventdestination":           &deletedVerifier{kind: "stripeeventdestination", path: "v2/core/event_destinations", secretOutput: "signing_secret"},
	"stripebillingportalconfiguration": &deactivatedVerifier{kind: "stripebillingportalconfiguration", path: "v1/billing_portal/configurations"},
	"stripepaymentmethodconfiguration": &deactivatedVerifier{kind: "stripepaymentmethodconfiguration", path: "v1/payment_method_configurations"},
	"stripepaymentmethoddomain":        &forgottenVerifier{kind: "stripepaymentmethoddomain", path: "v1/payment_method_domains"},
	"striperadarvaluelist": &withChildren{
		Verifier:  &deletedVerifier{kind: "striperadarvaluelist", path: "v1/radar/value_lists"},
		kind:      "striperadarvaluelist",
		output:    "item_ids",
		childPath: func(_, id string) string { return "v1/radar/value_list_items/" + url.PathEscape(id) },
		fate:      childDeleted,
	},
	"stripeproduct": &withChildren{
		Verifier: &deactivatedVerifier{kind: "stripeproduct", path: "v1/products"},
		kind:     "stripeproduct",
		output:   "product_feature_ids",
		childPath: func(product, id string) string {
			return "v1/products/" + url.PathEscape(product) + "/features/" + url.PathEscape(id)
		},
		fate: childDeleted,
	},
	"stripeprice":              &deactivatedVerifier{kind: "stripeprice", path: "v1/prices"},
	"stripeentitlementfeature": &deactivatedVerifier{kind: "stripeentitlementfeature", path: "v1/entitlements/features"},
	"stripecoupon":             &deletedVerifier{kind: "stripecoupon", path: "v1/coupons"},
	"stripepromotioncode":      &deactivatedVerifier{kind: "stripepromotioncode", path: "v1/promotion_codes"},
	"stripeshippingrate":       &deactivatedVerifier{kind: "stripeshippingrate", path: "v1/shipping_rates"},
	"stripetaxrate":            &deactivatedVerifier{kind: "stripetaxrate", path: "v1/tax_rates"},
	"stripetaxregistration":    &forgottenVerifier{kind: "stripetaxregistration", path: "v1/tax/registrations"},
	"stripepaymentlink":        &deactivatedVerifier{kind: "stripepaymentlink", path: "v1/payment_links"},
	"stripebillingmeter": &withChildren{
		Verifier:  &deactivatedVerifier{kind: "stripebillingmeter", path: "v1/billing/meters", byStatus: true},
		kind:      "stripebillingmeter",
		output:    "alert_ids",
		childPath: func(_, id string) string { return "v1/billing/alerts/" + url.PathEscape(id) },
		fate:      childForgotten,
	},
}

// GetVerifier returns the verifier for a component, or an error naming the unknown component.
func GetVerifier(kind string) (Verifier, error) {
	v, ok := verifiers[kind]
	if !ok {
		return nil, errors.Errorf("no Stripe verifier registered for kind %q", kind)
	}
	return v, nil
}
