package verify

import (
	"context"

	"github.com/pkg/errors"
)

// dialogflowCxAgentVerifier probes a Dialogflow CX agent and everything
// folded into it through the REST API: the agent reads back under its
// name with the exported id and start flow, and every exported webhook,
// tool, tool version, flow version, environment, and generative-settings
// entry exists.
type dialogflowCxAgentVerifier struct{}

// IDOutputKey is the agent's full resource name
// (projects/{p}/locations/{l}/agents/{id}).
func (v *dialogflowCxAgentVerifier) IDOutputKey() string { return "name" }

func (v *dialogflowCxAgentVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	agent, _, err := dialogflowGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "dialogflow cx agent %s not found after deploy", name)
	}
	if got := outputs["agent_id"]; got != lastPathSegment(stringField(agent, "name")) {
		return errors.Errorf("dialogflow cx agent %s agent_id output %q does not match the live id %q", name, got, lastPathSegment(stringField(agent, "name")))
	}
	if got := outputs["start_flow"]; got != "" && got != stringField(agent, "startFlow") {
		return errors.Errorf("dialogflow cx agent %s start_flow output %q does not match the live start flow %q", name, got, stringField(agent, "startFlow"))
	}

	children := []struct{ what, key string }{
		{"dialogflow cx webhook", "webhook_names"},
		{"dialogflow cx tool", "tool_names"},
		{"dialogflow cx tool version", "tool_version_names"},
		{"dialogflow cx flow version", "version_names"},
		{"dialogflow cx environment", "environment_names"},
		{"dialogflow cx generative settings", "generative_settings_names"},
	}
	for _, child := range children {
		for _, childName := range listOutput(outputs, child.key) {
			if _, _, err := dialogflowGet(ctx, svc, childName); err != nil {
				return errors.Wrapf(err, "%s %s not found after deploy", child.what, childName)
			}
		}
	}
	return nil
}

// VerifyAbsent confirms the agent is gone (its webhooks, tools, versions,
// and environments cannot outlive it).
func (v *dialogflowCxAgentVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := dialogflowGet(ctx, svc, name)
	return restAbsent("dialogflow cx agent", name, status, err)
}
