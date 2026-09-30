// Package verify holds one verifier per Stripe kind: how to read the object a module created,
// and what its destroy must leave behind.
package verify

import (
	"net/url"

	"github.com/pkg/errors"
)

// ResourceChecker reads Stripe objects by their API path.
type ResourceChecker interface {
	// ReadResource returns the object's JSON body and whether it exists.
	ReadResource(path string) (map[string]interface{}, bool, error)
}

// Verifier checks one kind's object after deploy and after destroy.
type Verifier interface {
	VerifyExists(checker ResourceChecker, id string) error
	// VerifyDestroyed asserts what the kind's destroy leaves, which is not always absence:
	// Stripe never deletes a portal or payment-method configuration, and destroy deactivates it.
	VerifyDestroyed(checker ResourceChecker, id string) error
}

// deletedVerifier is for a kind whose destroy deletes the object: it must exist after deploy
// and answer 404 after destroy.
type deletedVerifier struct {
	component string
	path      string // the API collection, e.g. "webhook_endpoints"
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
	path      string // the API collection, e.g. "billing_portal/configurations"
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

// verifiers maps each Stripe component directory to its verifier.
var verifiers = map[string]Verifier{
	"stripewebhookendpoint":            &deletedVerifier{component: "stripewebhookendpoint", path: "webhook_endpoints"},
	"stripebillingportalconfiguration": &deactivatedVerifier{component: "stripebillingportalconfiguration", path: "billing_portal/configurations"},
	"stripepaymentmethodconfiguration": &deactivatedVerifier{component: "stripepaymentmethodconfiguration", path: "payment_method_configurations"},
}

// GetVerifier returns the verifier for a component, or an error naming the unknown component.
func GetVerifier(component string) (Verifier, error) {
	v, ok := verifiers[component]
	if !ok {
		return nil, errors.Errorf("no Stripe verifier registered for component %q", component)
	}
	return v, nil
}
