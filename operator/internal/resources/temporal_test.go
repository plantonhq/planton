package resources_test

import (
	"testing"

	"github.com/plantonhq/planton/operator/internal/resources"
)

func TestTemporalHelmValues_DatabaseWiring(t *testing.T) {
	values := resources.TemporalHelmValues(defaultTemporal("myplanton", "planton-ns"))

	sql := extractTemporalDefaultSQL(t, values)

	expectedHost := "myplanton-postgres-rw.planton-ns.svc.cluster.local"
	if sql["host"] != expectedHost {
		t.Errorf("Expected host %q, got %q", expectedHost, sql["host"])
	}
	if sql["driver"] != resources.TemporalPostgresDriver {
		t.Errorf("Expected driver %q, got %q", resources.TemporalPostgresDriver, sql["driver"])
	}
	if sql["database"] != resources.TemporalDefaultDB {
		t.Errorf("Expected database %q, got %q", resources.TemporalDefaultDB, sql["database"])
	}
	// Temporal's schema job creates its two databases itself, which is why
	// the connection rides the superuser contract.
	if sql["user"] != resources.PostgreSQLSuperuser {
		t.Errorf("Expected user %q, got %q", resources.PostgreSQLSuperuser, sql["user"])
	}
	if sql["existingSecret"] != "myplanton-postgres-superuser" {
		t.Errorf("password must come from the CloudNativePG superuser Secret, got %q", sql["existingSecret"])
	}
}

func extractTemporalDefaultSQL(t *testing.T, values map[string]any) map[string]any {
	t.Helper()
	server, ok := values["server"].(map[string]any)
	if !ok {
		t.Fatal("Expected server key in values")
	}
	config, ok := server["config"].(map[string]any)
	if !ok {
		t.Fatal("Expected server.config key")
	}
	persistence, ok := config["persistence"].(map[string]any)
	if !ok {
		t.Fatal("Expected server.config.persistence key")
	}
	defaultStore, ok := persistence["default"].(map[string]any)
	if !ok {
		t.Fatal("Expected persistence.default key")
	}
	sql, ok := defaultStore["sql"].(map[string]any)
	if !ok {
		t.Fatal("Expected persistence.default.sql key")
	}
	return sql
}

func TestTemporalHelmValues_DisablesEmbeddedDatastores(t *testing.T) {
	values := resources.TemporalHelmValues(defaultTemporal("myplanton", "ns"))

	for _, store := range []string{"cassandra", "mysql", "postgresql"} {
		m, ok := values[store].(map[string]any)
		if !ok {
			t.Errorf("Expected %s key in values", store)
			continue
		}
		if m["enabled"] != false {
			t.Errorf("Expected %s.enabled=false, got %v", store, m["enabled"])
		}
	}
}

func TestTemporalServiceNames(t *testing.T) {
	if got := resources.TemporalFrontendServiceName("myp"); got != "myp-temporal-frontend" {
		t.Errorf("TemporalFrontendServiceName = %q, want %q", got, "myp-temporal-frontend")
	}
	if got := resources.TemporalWebUIServiceName("myp"); got != "myp-temporal-web" {
		t.Errorf("TemporalWebUIServiceName = %q, want %q", got, "myp-temporal-web")
	}
	if got := resources.TemporalFrontendEndpoint("myp", "ns"); got != "myp-temporal-frontend.ns.svc.cluster.local:7233" {
		t.Errorf("TemporalFrontendEndpoint = %q", got)
	}
}
