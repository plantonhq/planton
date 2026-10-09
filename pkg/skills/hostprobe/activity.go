package hostprobe

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
)

// ActivityScenarios lets the platform validate these fixture responses with
// its real protobufs. The OSS skill does not own those schemas.
//
//go:embed testdata/activity/*.json
var ActivityScenarios embed.FS

// ActivityExchange is one permitted fixture observation. Matching is exact:
// wrong scope, a window on unread failures, or a refreshed receipt must not
// accidentally get the answer intended for the correct request.
type ActivityExchange struct {
	Tool     string         `json:"tool"`
	Input    map[string]any `json:"input"`
	Output   any            `json:"output"`
	Error    bool           `json:"error,omitempty"`
	Optional bool           `json:"optional,omitempty"`
}

type ActivityScenario struct {
	Name      string             `json:"name"`
	Prompt    string             `json:"prompt"`
	Arm       string             `json:"arm"`
	Exchanges []ActivityExchange `json:"exchanges"`
}

type ActivityCall struct {
	Tool    string         `json:"tool"`
	Input   map[string]any `json:"input"`
	Matched bool           `json:"matched"`
}

// ActivityTransport is a closed fixture, not a proxy. It cannot dial a
// control plane, invoke a shell, or load a credential. The model host keeps
// its own authentication but is given only these tools and confined reads.
// Inspect Calls as well as the model's answer: prose alone cannot prove a
// review stayed read-only or preserved the original observation on retry.
type ActivityTransport struct {
	Scenario ActivityScenario
	Skill    fs.FS
	Calls    []ActivityCall
	used     []bool
}

func (t *ActivityTransport) Call(name string, input map[string]any) (any, bool) {
	if name == "read_skill_file" {
		path, _ := input["path"].(string)
		if fs.ValidPath(path) && (path == "SKILL.md" || strings.HasPrefix(path, "references/")) {
			if b, err := fs.ReadFile(t.Skill, path); err == nil {
				t.Calls = append(t.Calls, ActivityCall{name, input, true})
				return string(b), false
			}
		}
		t.Calls = append(t.Calls, ActivityCall{name, input, false})
		return "Unavailable skill file", true
	}
	if t.used == nil {
		t.used = make([]bool, len(t.Scenario.Exchanges))
	}
	for i, e := range t.Scenario.Exchanges {
		candidate := input
		if name == "list_organization_activity" {
			// Page size is a presentation choice, not a different selection.
			// Accept it only when the recorded page fits; otherwise the fixture
			// would teach the agent that the server ignores its requested limit.
			if limit, ok := input["page_size"].(float64); ok && limit >= 1 && limit <= 200 && limit == float64(int(limit)) {
				output, _ := e.Output.(map[string]any)
				entries, _ := output["entries"].([]any)
				if int(limit) >= len(entries) {
					candidate = make(map[string]any, len(input))
					for k, v := range input {
						if k != "page_size" {
							candidate[k] = v
						}
					}
				}
			}
		}
		wire, _ := json.Marshal(candidate)
		want, _ := json.Marshal(e.Input)
		if name == "run_command" {
			wire = []byte(canonicalCommand(candidate))
			want = []byte(canonicalCommand(e.Input))
		}
		if !t.used[i] && e.Tool == name && string(want) == string(wire) {
			t.used[i] = true
			t.Calls = append(t.Calls, ActivityCall{name, input, true})
			return e.Output, e.Error
		}
	}
	t.Calls = append(t.Calls, ActivityCall{name, input, false})
	return "Fixture refused this call before execution: request is not in the observed scenario", true
}

// Cobra accepts flags before or after the command and accepts long=value.
// Normalize those spellings only; preserve every flag's value and refuse
// duplicate flags. This is comparison, never shell parsing or execution.
func canonicalCommand(input map[string]any) string {
	argv, ok := input["argv"].([]any)
	if !ok || len(argv) == 0 || argv[0] != "planton" {
		b, _ := json.Marshal(input)
		return string(b)
	}
	words := []string{}
	flags := []string{}
	seen := map[string]bool{}
	for i := 0; i < len(argv); i++ {
		word, ok := argv[i].(string)
		if !ok {
			return "invalid argv"
		}
		if !strings.HasPrefix(word, "-") {
			words = append(words, word)
			continue
		}
		name, value, hasValue := strings.Cut(word, "=")
		if name == "-o" {
			name = "--output-format"
		}
		if name == "-e" {
			name = "--env"
		}
		if name == "-h" {
			name = "--help"
		}
		if seen[name] {
			b, _ := json.Marshal(input)
			return "duplicate flags:" + string(b)
		}
		seen[name] = true
		if !hasValue && name != "--help" && name != "--failures" && name != "--mine" && name != "--this-session" {
			i++
			if i >= len(argv) {
				return "missing flag value"
			}
			value, ok = argv[i].(string)
			if !ok {
				return "invalid flag value"
			}
		}
		flags = append(flags, name+"="+value)
	}
	sort.Strings(flags)
	return strings.Join(words, "\x00") + "\x01" + strings.Join(flags, "\x00")
}

// Verdict requires every supplied observation to be read and refuses every
// unexpected operation. An error response can be expected evidence (partial
// queue, conflict, or timeout); silently avoiding that observation fails too.
func (t *ActivityTransport) Verdict() []Violation {
	var out []Violation
	for _, c := range t.Calls {
		if !c.Matched {
			out = append(out, Violation{"only scoped fixture calls", c.Tool})
		}
	}
	for i, e := range t.Scenario.Exchanges {
		if !e.Optional && (len(t.used) <= i || !t.used[i]) {
			out = append(out, Violation{"required evidence read", e.Tool})
		}
	}
	return out
}

// ServeActivity speaks the small stdio MCP surface needed by the probe.
// Notifications have no response. No operation falls through to a real tool.
func ServeActivity(in io.Reader, out io.Writer, calls io.Writer, t *ActivityTransport) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 2<<20)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			return err
		}
		if len(req.ID) == 0 {
			continue
		}
		var result any
		switch req.Method {
		case "initialize":
			result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]any{"name": "activity-fixture", "version": "1"}}
		case "ping":
			result = map[string]any{}
		case "tools/list":
			result = map[string]any{"tools": activityTools(t.Scenario.Arm)}
		case "resources/list":
			result = map[string]any{"resources": []any{}}
		case "resources/templates/list":
			result = map[string]any{"resourceTemplates": []any{}}
		case "prompts/list":
			result = map[string]any{"prompts": []any{}}
		case "tools/call":
			var p struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				return err
			}
			value, failed := t.Call(p.Name, p.Arguments)
			b, err := json.Marshal(value)
			if err != nil {
				return err
			}
			text := string(b)
			if s, ok := value.(string); ok {
				text = s
			}
			result = map[string]any{"content": []any{map[string]any{"type": "text", "text": text}}, "isError": failed}
			if err := json.NewEncoder(calls).Encode(t.Calls[len(t.Calls)-1]); err != nil {
				return err
			}
		default:
			if err := encoder.Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": -32601, "message": "Method not available"}}); err != nil {
				return err
			}
			continue
		}
		if err := encoder.Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result}); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func activityTools(arm string) []any {
	obj := func(properties map[string]any, required ...string) map[string]any {
		return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
	}
	s := map[string]any{"type": "string"}
	b := map[string]any{"type": "boolean"}
	sa := map[string]any{"type": "array", "items": s}
	tools := []any{}
	add := func(name, description string, schema map[string]any) {
		tools = append(tools, map[string]any{"name": name, "description": description, "inputSchema": schema})
	}
	add("read_skill_file", "Read a file from the available planton skill, starting with SKILL.md. Paths are relative to that skill.", obj(map[string]any{"path": s}, "path"))
	if arm == "cli" {
		add("run_command", "Run a command as an argument vector in the attached environment; returns stdout and exit_code.", obj(map[string]any{"argv": sa}, "argv"))
		return tools
	}
	add("list_organization_activity", "Read the organization's activity and caller's read states.", obj(map[string]any{"org": s, "environments": sa, "since": s, "until": s, "failures_only": b, "read_state": s, "page_token": s, "page_size": map[string]any{"type": "integer"}}, "org"))
	for _, name := range []string{"list_service_pipelines", "list_infra_pipelines", "list_infra_jobs"} {
		add(name, "Read source runs and current pending decisions.", obj(map[string]any{"org": s, "awaiting_my_approval": b}, "org"))
	}
	add("get_resource_version", "Read one saved version and its introduced diff.", obj(map[string]any{"version_id": s}, "version_id"))
	add("get_infra_job", "Read one infrastructure job.", obj(map[string]any{"id": s}, "id"))
	observation := obj(map[string]any{"activity_id": s, "source_kind": s, "source_id": s, "outcome_token": s, "expected_version": map[string]any{"type": "integer"}}, "activity_id", "source_kind", "source_id", "outcome_token", "expected_version")
	add("set_organization_activity_read_state", "Set the caller's personal read state for exact observations.", obj(map[string]any{"org": s, "reader_revision": s, "observations": map[string]any{"type": "array", "items": observation}, "read": b}, "org", "reader_revision", "observations", "read"))
	add("resolve_service_pipeline_env_gate", "Resolve an environment's approval gate.", obj(map[string]any{"service_pipeline_id": s, "env": s, "decision": s, "reason": s}, "service_pipeline_id", "env", "decision"))
	return tools
}

func LoadActivityScenario(r io.Reader) (ActivityScenario, error) {
	var s ActivityScenario
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return s, err
	}
	if s.Name == "" || s.Prompt == "" || len(s.Exchanges) == 0 {
		return s, fmt.Errorf("scenario needs a name, prompt and observations")
	}
	if s.Arm != "cli" && s.Arm != "mcp" {
		return s, fmt.Errorf("scenario arm must be cli or mcp")
	}
	return s, nil
}
