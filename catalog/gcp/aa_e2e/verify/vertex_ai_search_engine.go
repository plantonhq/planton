package verify

import (
	"context"

	"github.com/pkg/errors"
)

// vertexAiSearchEngineVerifier probes a Vertex AI Search engine -- whichever
// of the three arms built it -- and its folded controls, serving config,
// widget config, and assistants through the Discovery Engine REST API. The
// engine's solutionType must agree with the exported engine_type; the
// serving config's control lists must name every control the module
// exported for that action.
type vertexAiSearchEngineVerifier struct{}

// IDOutputKey is the engine's full resource name
// (projects/{p}/locations/{l}/collections/{c}/engines/{id}).
func (v *vertexAiSearchEngineVerifier) IDOutputKey() string { return "name" }

// solutionTypeForArm mirrors the modules' derivation: the Discovery Engine
// solution each arm hard-codes.
func solutionTypeForArm(engineType string) string {
	switch engineType {
	case "CHAT":
		return "SOLUTION_TYPE_CHAT"
	case "RECOMMENDATION":
		return "SOLUTION_TYPE_RECOMMENDATION"
	default:
		return "SOLUTION_TYPE_SEARCH"
	}
}

// VerifyExists confirms the engine reads back with the arm's solution type,
// that a chat engine reports the Dialogflow agent the module exported, and
// that every exported control, assistant, serving config, and widget config
// exists.
func (v *vertexAiSearchEngineVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	engine, _, err := discoveryEngineGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "vertex ai search engine %s not found after deploy", name)
	}
	if got := outputs["engine_id"]; got != "" && lastPathSegment(stringField(engine, "name")) != got {
		return errors.Errorf("vertex ai search engine %s engine_id output %q does not match live name %q", name, got, lastPathSegment(stringField(engine, "name")))
	}
	if want := solutionTypeForArm(outputs["engine_type"]); stringField(engine, "solutionType") != want {
		return errors.Errorf("vertex ai search engine %s has solutionType %q, want %q for engine_type %q", name, stringField(engine, "solutionType"), want, outputs["engine_type"])
	}
	if outputs["engine_type"] == "CHAT" {
		metadata, _ := engine["chatEngineMetadata"].(map[string]interface{})
		liveAgent := stringField(metadata, "dialogflowAgent")
		if liveAgent == "" {
			return errors.Errorf("vertex ai search chat engine %s reports no Dialogflow agent after deploy", name)
		}
		if got := outputs["dialogflow_agent"]; got != "" && got != liveAgent {
			return errors.Errorf("vertex ai search chat engine %s dialogflow_agent output %q does not match live %q", name, got, liveAgent)
		}
	}

	for _, controlName := range listOutput(outputs, "control_names") {
		if _, _, err := discoveryEngineGet(ctx, svc, controlName); err != nil {
			return errors.Wrapf(err, "vertex ai search control %s not found after deploy", controlName)
		}
	}
	for _, assistantName := range listOutput(outputs, "assistant_names") {
		if _, _, err := discoveryEngineGet(ctx, svc, assistantName); err != nil {
			return errors.Wrapf(err, "vertex ai search assistant %s not found after deploy", assistantName)
		}
	}
	if servingConfigName := outputs["serving_config_name"]; servingConfigName != "" {
		servingConfig, _, err := discoveryEngineGet(ctx, svc, servingConfigName)
		if err != nil {
			return errors.Wrapf(err, "vertex ai search serving config %s not found after deploy", servingConfigName)
		}
		// Every exported control the serving config could list must appear
		// in one of its id lists -- the PATCH landed after the controls.
		listed := map[string]bool{}
		for _, key := range []string{"boostControlIds", "filterControlIds", "promoteControlIds", "redirectControlIds", "synonymsControlIds"} {
			ids, _ := servingConfig[key].([]interface{})
			for _, id := range ids {
				if s, ok := id.(string); ok {
					listed[s] = true
				}
			}
		}
		if len(listed) == 0 && len(listOutput(outputs, "control_names")) > 0 {
			return errors.Errorf("vertex ai search serving config %s lists no controls though the engine exported %d", servingConfigName, len(listOutput(outputs, "control_names")))
		}
	}
	if widgetConfigName := outputs["widget_config_name"]; widgetConfigName != "" {
		if _, _, err := discoveryEngineGet(ctx, svc, widgetConfigName); err != nil {
			return errors.Wrapf(err, "vertex ai search widget config %s not found after deploy", widgetConfigName)
		}
	}
	return nil
}

// VerifyAbsent confirms the engine is gone (its controls and assistants
// cannot outlive it; the serving and widget configs are Google's own and
// vanish with the engine).
func (v *vertexAiSearchEngineVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := discoveryEngineGet(ctx, svc, name)
	return discoveryEngineAbsent("vertex ai search engine", name, status, err)
}
