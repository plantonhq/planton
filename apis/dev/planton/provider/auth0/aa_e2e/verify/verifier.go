package verify

import (
	"fmt"

	"github.com/pkg/errors"
)

// ResourceChecker can test whether a Management API resource exists.
type ResourceChecker interface {
	ResourceExists(path string) (bool, error)
}

// Verifier checks Auth0 resources for a single component type.
type Verifier interface {
	VerifyExists(checker ResourceChecker, id string) error
	VerifyAbsent(checker ResourceChecker, id string) error
}

// apiPathVerifier is the common implementation: one Management API path template.
type apiPathVerifier struct {
	component  string
	pathFormat string // e.g. "clients/%s"
}

func (v *apiPathVerifier) VerifyExists(checker ResourceChecker, id string) error {
	path := v.formatPath(id)
	exists, err := checker.ResourceExists(path)
	if err != nil {
		return errors.Wrapf(err, "%s verify-exists failed", v.component)
	}
	if !exists {
		return errors.Errorf("%s %s not found after deploy", v.component, id)
	}
	return nil
}

func (v *apiPathVerifier) VerifyAbsent(checker ResourceChecker, id string) error {
	path := v.formatPath(id)
	exists, err := checker.ResourceExists(path)
	if err != nil {
		return errors.Wrapf(err, "%s verify-absent failed", v.component)
	}
	if exists {
		return errors.Errorf("%s %s still exists after destroy", v.component, id)
	}
	return nil
}

func (v *apiPathVerifier) formatPath(id string) string {
	return fmt.Sprintf(v.pathFormat, id)
}

// verifiers maps component name to the Management API path used for verification.
var verifiers = map[string]Verifier{
	"auth0client":         &apiPathVerifier{component: "auth0client", pathFormat: "clients/%s"},
	"auth0connection":     &apiPathVerifier{component: "auth0connection", pathFormat: "connections/%s"},
	"auth0resourceserver": &apiPathVerifier{component: "auth0resourceserver", pathFormat: "resource-servers/%s"},
	"auth0action":         &apiPathVerifier{component: "auth0action", pathFormat: "actions/actions/%s"},
	"auth0eventstream":    &apiPathVerifier{component: "auth0eventstream", pathFormat: "event-streams/%s"},
	"auth0role":           &apiPathVerifier{component: "auth0role", pathFormat: "roles/%s"},
}

// GetVerifier returns the verifier for a component, or an error if unknown.
func GetVerifier(component string) (Verifier, error) {
	v, ok := verifiers[component]
	if !ok {
		return nil, errors.Errorf("no Auth0 verifier registered for component %q", component)
	}
	return v, nil
}
