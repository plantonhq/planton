package resources_test

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/operator/internal/resources"
)

func TestOpenFGAHelmValues_DatastoreURI(t *testing.T) {
	values := resources.OpenFGAHelmValues(defaultOpenFGA("myplanton", "ns"))

	ds, ok := values["datastore"].(map[string]any)
	if !ok {
		t.Fatal("Expected datastore key in values")
	}
	if ds["engine"] != resources.OpenFGADatastoreEngine {
		t.Errorf("Expected engine %q, got %q", resources.OpenFGADatastoreEngine, ds["engine"])
	}

	uri, ok := ds["uri"].(string)
	if !ok {
		t.Fatal("Expected datastore.uri to be a string")
	}
	if !strings.HasPrefix(uri, "postgres://postgres:") {
		t.Errorf("URI should use the platform superuser contract, got %q", uri)
	}
	if !strings.Contains(uri, "myplanton-postgres-rw.ns.svc.cluster.local") {
		t.Errorf("URI should reference the cluster's read-write Service, got %q", uri)
	}
	if !strings.Contains(uri, "/openfga?") {
		t.Errorf("URI should reference openfga database, got %q", uri)
	}
	if !strings.Contains(uri, "$(OPENFGA_DATASTORE_PASSWORD)") {
		t.Errorf("URI should use env var expansion for password, got %q", uri)
	}
}

func TestOpenFGAHelmValues_ExtraEnvVars(t *testing.T) {
	values := resources.OpenFGAHelmValues(defaultOpenFGA("myplanton", "ns"))

	envVars, ok := values["extraEnvVars"].([]any)
	if !ok {
		t.Fatal("Expected extraEnvVars key in values")
	}
	if len(envVars) != 1 {
		t.Fatalf("Expected 1 extra env var, got %d", len(envVars))
	}

	envVar := envVars[0].(map[string]any)
	if envVar["name"] != "OPENFGA_DATASTORE_PASSWORD" {
		t.Errorf("Expected env var name OPENFGA_DATASTORE_PASSWORD, got %q", envVar["name"])
	}

	valueFrom := envVar["valueFrom"].(map[string]any)
	secretRef := valueFrom["secretKeyRef"].(map[string]any)
	if secretRef["name"] != "myplanton-postgres-superuser" {
		t.Errorf("password must come from the CloudNativePG superuser Secret, got %q", secretRef["name"])
	}
	if secretRef["key"] != "password" {
		t.Errorf("expected the basic-auth password key, got %q", secretRef["key"])
	}
}

func TestOpenFGAServiceNames(t *testing.T) {
	if got := resources.OpenFGAServiceName("myp"); got != "myp-openfga" {
		t.Errorf("OpenFGAServiceName = %q, want %q", got, "myp-openfga")
	}
	if got := resources.OpenFGAServiceFQDN("myp", "ns"); got != "myp-openfga.ns.svc.cluster.local" {
		t.Errorf("OpenFGAServiceFQDN = %q", got)
	}
	if got := resources.OpenFGAHTTPURL("myp", "ns"); got != "http://myp-openfga.ns.svc.cluster.local:8080" {
		t.Errorf("OpenFGAHTTPURL = %q", got)
	}
}
