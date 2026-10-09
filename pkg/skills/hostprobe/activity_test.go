package hostprobe

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

func observationInput() map[string]any {
	return map[string]any{"org": "acme", "reader_revision": "reader-one", "read": true, "observations": []any{map[string]any{"activity_id": "oact_failure", "source_kind": "infra_job", "source_id": "ij_cleanup", "outcome_token": "outcome-one", "expected_version": 7}}}
}

func TestActivityTransportRefusesEveryUnrecordedOperation(t *testing.T) {
	for _, name := range []string{"resolve_service_pipeline_env_gate", "set_organization_activity_read_state", "run_command", "unknown"} {
		t.Run(name, func(t *testing.T) {
			fixture := &ActivityTransport{}
			_, failed := fixture.Call(name, map[string]any{"command": "planton apply -f live.yaml"})
			if !failed || len(fixture.Verdict()) != 1 {
				t.Fatal("unexpected operation escaped the fixture")
			}
		})
	}
}

func TestActivityTransportPreservesOriginalMutationIntent(t *testing.T) {
	input := observationInput()
	fixture := &ActivityTransport{Scenario: ActivityScenario{Exchanges: []ActivityExchange{
		{Tool: "set_organization_activity_read_state", Input: input, Output: "timeout", Error: true},
		{Tool: "set_organization_activity_read_state", Input: input, Output: map[string]any{"results": []any{map[string]any{"outcome": "applied"}}}},
	}}}
	if _, failed := fixture.Call("set_organization_activity_read_state", observationInput()); !failed {
		t.Fatal("lost timeout")
	}
	if _, failed := fixture.Call("set_organization_activity_read_state", observationInput()); failed {
		t.Fatal("original retry refused")
	}
	if vs := fixture.Verdict(); len(vs) != 0 {
		t.Fatal(vs)
	}
	if _, failed := fixture.Call("set_organization_activity_read_state", observationInput()); !failed {
		t.Fatal("unbounded retry accepted")
	}
}

func TestActivityTransportRejectsChangedReaderOutcomeDirectionAndScope(t *testing.T) {
	for _, field := range []string{"org", "reader_revision", "read", "source_id", "outcome_token", "expected_version"} {
		t.Run(field, func(t *testing.T) {
			fixture := &ActivityTransport{Scenario: ActivityScenario{Exchanges: []ActivityExchange{{Tool: "set_organization_activity_read_state", Input: observationInput()}}}}
			input := observationInput()
			switch field {
			case "org", "reader_revision":
				input[field] = "changed"
			case "read":
				input[field] = false
			default:
				observation := input["observations"].([]any)[0].(map[string]any)
				if field == "expected_version" {
					observation[field] = 8
				} else {
					observation[field] = "changed"
				}
			}
			if _, failed := fixture.Call("set_organization_activity_read_state", input); !failed {
				t.Fatal("changed intent accepted")
			}
		})
	}
}

func TestActivityTransportSeparatesUnreadAndWindowedReads(t *testing.T) {
	input := map[string]any{"org": "acme", "read_state": "unread", "failures_only": true}
	fixture := &ActivityTransport{Scenario: ActivityScenario{Exchanges: []ActivityExchange{{Tool: "list_organization_activity", Input: input, Output: map[string]any{"unread_failure_count": "42", "entries": []any{}}}}}}
	wrong := map[string]any{"org": "acme", "read_state": "unread", "failures_only": true, "since": "24h"}
	if _, failed := fixture.Call("list_organization_activity", wrong); !failed {
		t.Fatal("window narrowed old unread failures")
	}
	value, failed := fixture.Call("list_organization_activity", input)
	if failed || value.(map[string]any)["unread_failure_count"] != "42" {
		t.Fatal("count replaced by page length")
	}
}

func TestActivityTransportConfinesSkillReads(t *testing.T) {
	fixture := &ActivityTransport{Skill: fstest.MapFS{"SKILL.md": {Data: []byte("skill")}, "private.txt": {Data: []byte("do not read")}}}
	for _, path := range []string{"../SKILL.md", "/etc/passwd", "private.txt", "references/../../private.txt", "references/missing.md"} {
		if _, failed := fixture.Call("read_skill_file", map[string]any{"path": path}); !failed {
			t.Fatalf("read %s", path)
		}
	}
	if value, failed := fixture.Call("read_skill_file", map[string]any{"path": "SKILL.md"}); failed || value != "skill" {
		t.Fatal("valid read failed")
	}
}

func TestActivityMCPWireRetainsPartialErrorsAndCallEvidence(t *testing.T) {
	fixture := &ActivityTransport{Scenario: ActivityScenario{Exchanges: []ActivityExchange{{Tool: "list_infra_jobs", Input: map[string]any{"org": "acme", "awaiting_my_approval": true}, Output: "Unavailable", Error: true}}}}
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_infra_jobs","arguments":{"org":"acme","awaiting_my_approval":true}}}
`)
	var out, calls bytes.Buffer
	if err := ServeActivity(in, &out, &calls, fixture); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("notification received a response: %s", out.String())
	}
	var response struct {
		Result struct {
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[2]), &response); err != nil || !response.Result.IsError {
		t.Fatal("source failure was erased")
	}
	var call ActivityCall
	if err := json.Unmarshal(bytes.TrimSpace(calls.Bytes()), &call); err != nil || !call.Matched || call.Tool != "list_infra_jobs" {
		t.Fatal("lost call evidence")
	}
	if vs := fixture.Verdict(); len(vs) != 0 {
		t.Fatal(vs)
	}
}

func TestActivityVerdictRequiresEvidenceEvenWhenNothingFailed(t *testing.T) {
	fixture := &ActivityTransport{Scenario: ActivityScenario{Exchanges: []ActivityExchange{{Tool: "list_infra_jobs", Input: map[string]any{"org": "acme"}}}}}
	if len(fixture.Verdict()) != 1 {
		t.Fatal("no calls mistaken for success")
	}
}

func TestActivityCommandMatchingPreservesScopeAndAcceptsFlagOrder(t *testing.T) {
	input := func(args ...string) map[string]any {
		v := make([]any, len(args))
		for i, s := range args {
			v[i] = s
		}
		return map[string]any{"argv": v}
	}
	a := input("planton", "activity", "--org", "acme", "--env", "prod", "-o", "json")
	b := input("planton", "--org=acme", "activity", "--output-format=json", "-e", "prod")
	if canonicalCommand(a) != canonicalCommand(b) {
		t.Fatal("equivalent Cobra flag spellings differ")
	}
	for _, wrong := range []map[string]any{
		input("planton", "activity", "--org", "other", "--env", "prod", "-o", "json"),
		input("planton", "activity", "--org", "acme", "--env", "prod", "-o", "json", "--attention"),
		input("planton", "activity", "--org", "acme", "--org", "other", "--env", "prod", "-o", "json"),
	} {
		if canonicalCommand(a) == canonicalCommand(wrong) {
			t.Fatal("different scope or invalid flags accepted")
		}
	}
}

func TestActivityPageSizeCannotNarrowAwayRecordedEvidence(t *testing.T) {
	fixture := &ActivityTransport{Scenario: ActivityScenario{Exchanges: []ActivityExchange{{Tool: "list_organization_activity", Input: map[string]any{"org": "acme"}, Output: map[string]any{"entries": []any{1, 2}}}}}}
	if _, failed := fixture.Call("list_organization_activity", map[string]any{"org": "acme", "page_size": float64(1)}); !failed {
		t.Fatal("fixture ignored page size")
	}
	if _, failed := fixture.Call("list_organization_activity", map[string]any{"org": "acme", "page_size": float64(100)}); failed {
		t.Fatal("valid page size refused")
	}
}

func TestActivityAnswerExcludesToolOutput(t *testing.T) {
	transcript := []byte("{\"type\":\"tool_result\",\"result\":\"not the answer\"}\n{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"actual answer\"}}\n")
	if activityAnswer(transcript) != "actual answer" {
		t.Fatal("tool evidence mistaken for the model's answer")
	}
}
