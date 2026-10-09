package hostprobe

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ActivityHostResult keeps the complete evidence even on timeout. Auth or
// host startup failures are not successful behavioral checks.
type ActivityHostResult struct {
	Directory  string
	Transcript []byte
	Answer     string
	Calls      []ActivityCall
	Violations []Violation
}

// RunActivityHost uses tool-restricted host modes rather than the older
// authoring probe's unrestricted shell. The only executable tool available
// to the model is the closed MCP fixture. The Claude read tool is confined
// by restricted mode; Codex has no shell and runs in a read-only sandbox.
func RunActivityHost(ctx context.Context, host, binary, skill, scenarioPath string, timeout time.Duration) (ActivityHostResult, error) {
	var result ActivityHostResult
	f, err := os.Open(scenarioPath)
	if err != nil {
		return result, err
	}
	scenario, err := LoadActivityScenario(f)
	f.Close()
	if err != nil {
		return result, err
	}
	root, err := os.MkdirTemp("", "planton-activity-probe-")
	if err != nil {
		return result, err
	}
	result.Directory = root
	// Freeze the content under proof. Another author can edit the canonical
	// skill while a model is running; that must not change this run's input.
	snapshot := filepath.Join(root, "skill")
	if err := copyTree(skill, snapshot); err != nil {
		return result, err
	}
	manifest := map[string]string{}
	if err := filepath.WalkDir(snapshot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(snapshot, path)
		manifest[filepath.ToSlash(rel)] = fmt.Sprintf("%x", sha256.Sum256(data))
		return nil
	}); err != nil {
		return result, err
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(root, "skill-manifest.json"), manifestBytes, 0600); err != nil {
		return result, err
	}
	scenarioBytes, _ := json.MarshalIndent(scenario, "", "  ")
	scenarioPath = filepath.Join(root, "scenario.json")
	if err := os.WriteFile(scenarioPath, scenarioBytes, 0600); err != nil {
		return result, err
	}
	skill = snapshot
	calls := filepath.Join(root, "calls.jsonl")
	fixtureArgs := []string{"--scenario", scenarioPath, "--skill", skill, "--calls", calls}
	config := map[string]any{"mcpServers": map[string]any{"activity_fixture": map[string]any{"command": binary, "args": fixtureArgs}}}
	configBytes, _ := json.Marshal(config)
	configPath := filepath.Join(root, "mcp.json")
	if err := os.WriteFile(configPath, configBytes, 0600); err != nil {
		return result, err
	}
	prompt := "Use the available planton skill. Read its SKILL.md with read_skill_file, then the references needed for this request. The connected tools operate on the instance and account named below.\n\n" + scenario.Prompt
	var args []string
	switch host {
	case "claude":
		args = []string{"--restricted", "--tools", "Read", "--allowedTools", "Read,mcp__activity_fixture__*", "--permission-mode", "dontAsk", "--setting-sources", "", "--strict-mcp-config", "--mcp-config", configPath, "--no-session-persistence", "--no-chrome", "-p", "--output-format", "stream-json", "--verbose", prompt}
	case "codex":
		commandJSON, _ := json.Marshal(binary)
		argsJSON, _ := json.Marshal(fixtureArgs)
		// Approval applies only to this closed fixture server. It does not
		// weaken the host sandbox or grant a live tool any permission.
		args = []string{"--no-daemon", "exec", "--ignore-user-config", "--ignore-rules", "--disable", "shell_tool", "--disable", "multi_agent", "--enable", "skip_host_skill_discovery", "--sandbox", "read-only", "--skip-git-repo-check", "--ephemeral", "--json", "-c", `web_search="disabled"`, "-c", "mcp_servers.activity_fixture.command=" + string(commandJSON), "-c", "mcp_servers.activity_fixture.args=" + string(argsJSON), "-c", `mcp_servers.activity_fixture.default_tools_approval_mode="approve"`, prompt}
	default:
		return result, fmt.Errorf("%s has no verified restricted activity adapter", host)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, host, args...)
	cmd.Dir = root
	cmd.Env = activityHostEnv()
	cmd.WaitDelay = 10 * time.Second
	var transcript bytes.Buffer
	cmd.Stdout = &transcript
	cmd.Stderr = &transcript
	runErr := cmd.Run()
	result.Transcript = transcript.Bytes()
	result.Answer = activityAnswer(result.Transcript)
	if err := os.WriteFile(filepath.Join(root, "transcript.jsonl"), result.Transcript, 0600); err != nil {
		return result, err
	}
	if err := os.WriteFile(filepath.Join(root, "answer.md"), []byte(result.Answer), 0600); err != nil {
		return result, err
	}
	log, err := os.ReadFile(calls)
	if err != nil {
		return result, fmt.Errorf("%s made no recorded calls (host error %v): %w", host, runErr, err)
	}
	for _, line := range bytes.Split(log, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var call ActivityCall
		if err := json.Unmarshal(line, &call); err != nil {
			return result, err
		}
		result.Calls = append(result.Calls, call)
	}
	result.Violations = judgeActivityJourney(scenario, os.DirFS(skill), result.Calls)
	if strings.TrimSpace(result.Answer) == "" {
		result.Violations = append(result.Violations, Violation{"a completed user answer", "host returned no final answer"})
	}
	return result, runErr
}

func judgeActivityJourney(scenario ActivityScenario, skill fs.FS, calls []ActivityCall) []Violation {
	judge := &ActivityTransport{Scenario: scenario, Skill: skill}
	loaded := false
	for _, call := range calls {
		_, failed := judge.Call(call.Tool, call.Input)
		if !failed && call.Tool == "read_skill_file" && call.Input["path"] == "SKILL.md" {
			loaded = true
		}
	}
	violations := judge.Verdict()
	if !loaded {
		violations = append(violations, Violation{"skill actually loaded", "no successful SKILL.md read"})
	}
	return violations
}

func activityAnswer(transcript []byte) string {
	var answer string
	for _, line := range bytes.Split(transcript, []byte("\n")) {
		var event struct {
			Type   string `json:"type"`
			Result string `json:"result"`
			Item   struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
		}
		if json.Unmarshal(line, &event) != nil {
			continue
		}
		if event.Type == "result" {
			answer = event.Result
		}
		if event.Type == "item.completed" && event.Item.Type == "agent_message" {
			answer = event.Item.Text
		}
	}
	return answer
}

// Keep the host's model login location unchanged, but pass no Planton,
// cloud, kube, endpoint, agent-hook, or unrelated MCP environment through.
// The fixture itself never consumes credentials, even when the model host
// uses its own sign-in. Never log environment values in probe evidence.
func activityHostEnv() []string {
	allowed := map[string]bool{}
	for _, key := range []string{"PATH", "HOME", "USER", "LOGNAME", "TMPDIR", "LANG", "LC_ALL", "TERM", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_TOKEN", "CURSOR_API_KEY"} {
		allowed[key] = true
	}
	var out []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if allowed[key] {
			out = append(out, entry)
		}
	}
	return out
}
