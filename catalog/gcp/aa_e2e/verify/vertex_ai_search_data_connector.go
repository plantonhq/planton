package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

// vertexAiSearchDataConnectorVerifier probes a Vertex AI Search data
// connector and the collection it created through the Discovery Engine
// REST API: the connector reads back under its name, reports a state that
// is not a failure, and every data store it created for an entity exists.
type vertexAiSearchDataConnectorVerifier struct{}

// IDOutputKey is the connector's full resource name
// (projects/{p}/locations/{l}/collections/{c}/dataConnector).
func (v *vertexAiSearchDataConnectorVerifier) IDOutputKey() string { return "name" }

// VerifyExists confirms the connector reads back with the exported
// collection id and a non-failed state, and that every entity data store
// the module exported exists.
func (v *vertexAiSearchDataConnectorVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	connector, _, err := discoveryEngineGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "vertex ai search data connector %s not found after deploy", name)
	}
	if got := outputs["collection_id"]; got != "" && discoveryEngineCollection(stringField(connector, "name")) != got {
		return errors.Errorf("vertex ai search data connector %s collection_id output %q does not match live name %q", name, got, stringField(connector, "name"))
	}
	switch state := stringField(connector, "state"); state {
	case "FAILED", "INITIALIZATION_FAILED":
		return errors.Errorf("vertex ai search data connector %s is in state %s after deploy", name, state)
	}
	if got := outputs["state"]; got != "" && stringField(connector, "state") == "" {
		return errors.Errorf("vertex ai search data connector %s exports state %q but reports none", name, got)
	}

	for _, storeName := range listOutput(outputs, "entity_data_stores") {
		if _, _, err := discoveryEngineGet(ctx, svc, storeName); err != nil {
			return errors.Wrapf(err, "vertex ai search data store %s (created by the connector) not found after deploy", storeName)
		}
	}
	return nil
}

// VerifyAbsent confirms the connector is gone.
func (v *vertexAiSearchDataConnectorVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := discoveryEngineGet(ctx, svc, name)
	return restAbsent("vertex ai search data connector", name, status, err)
}

// discoveryEngineCollection extracts the collection id from a Discovery
// Engine resource name (.../collections/{c}/...).
func discoveryEngineCollection(name string) string {
	parts := strings.Split(name, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "collections" {
			return parts[i+1]
		}
	}
	return ""
}
