// Package verify holds one verifier per Stripe kind: how to read the object a module created,
// and what its destroy must leave behind.
package verify

import (
	"net/url"

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
	// never deletes a configuration, product, price or feature (destroy deactivates it), and the
	// provider's destroy of a payment-method domain makes no call at all.
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
// after deploy, and still readable with active false after destroy. Asserting absence would fail
// every honest run.
type deactivatedVerifier struct {
	component string
	path      string // the API collection, e.g. "v1/billing_portal/configurations"
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

// verifiers maps each Stripe component directory to its verifier.
var verifiers = map[string]Verifier{
	"stripewebhookendpoint":            &deletedVerifier{component: "stripewebhookendpoint", path: "v1/webhook_endpoints"},
	"stripeeventdestination":           &deletedVerifier{component: "stripeeventdestination", path: "v2/core/event_destinations"},
	"stripebillingportalconfiguration": &deactivatedVerifier{component: "stripebillingportalconfiguration", path: "v1/billing_portal/configurations"},
	"stripepaymentmethodconfiguration": &deactivatedVerifier{component: "stripepaymentmethodconfiguration", path: "v1/payment_method_configurations"},
	"stripepaymentmethoddomain":        &forgottenVerifier{component: "stripepaymentmethoddomain", path: "v1/payment_method_domains"},
	"striperadarvaluelist":             &deletedVerifier{component: "striperadarvaluelist", path: "v1/radar/value_lists"},
	"stripeproduct":                    &deactivatedVerifier{component: "stripeproduct", path: "v1/products"},
	"stripeprice":                      &deactivatedVerifier{component: "stripeprice", path: "v1/prices"},
	"stripeentitlementfeature":         &deactivatedVerifier{component: "stripeentitlementfeature", path: "v1/entitlements/features"},
}

// GetVerifier returns the verifier for a component, or an error naming the unknown component.
func GetVerifier(component string) (Verifier, error) {
	v, ok := verifiers[component]
	if !ok {
		return nil, errors.Errorf("no Stripe verifier registered for component %q", component)
	}
	return v, nil
}
