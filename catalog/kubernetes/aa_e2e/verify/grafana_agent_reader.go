package verify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// grafanaAgentReader is what a manifest's spec.agent_reader promises, with
// the proto's defaults applied: the account, the token's generation, and
// whether every token of the account must be refused.
type grafanaAgentReader struct {
	ServiceAccountName string
	TokenGeneration    int
	Disabled           bool
}

// agentReaderTokensSeen keeps, per Grafana (namespace/name), the token the
// first act's verification read, so the second act can prove Grafana now
// refuses it -- a replaced or disabled token answering 401 is the promise,
// and only the old token's own value can prove it. Both acts run in this
// test process; a lane resumed at its second act has no first token and
// says so instead of passing on the token list alone.
var agentReaderTokensSeen sync.Map

// proveAgentReader holds spec.agent_reader to what it promises, through
// the same port-forward and admin credentials as the rest of the proof:
//   - the Job the module rendered completed;
//   - its Role grants exactly the token Secret and the ServiceAccount the
//     owner reference names, nothing else;
//   - on: the Secret `<name>-agent-reader` holds the declared generation,
//     is owned by that ServiceAccount, and its token signs in as the
//     account, a Viewer that may query datasources, holds no write action
//     and is refused a dashboard write; the account holds exactly one
//     token;
//   - disabled: the account is disabled and the Secret is gone;
//   - second act: the first act's token is refused (401).
func (v *GrafanaVerifier) proveAgentReader(ctx context.Context, kubeconfig, base, user, password string) error {
	r := v.AgentReader
	readerName := v.Name + "-agent-reader"
	fmt.Printf("  [verify] AGENT READER: account %q, generation %d, disabled %v\n",
		r.ServiceAccountName, r.TokenGeneration, r.Disabled)

	if err := v.agentReaderJobCompleted(ctx, kubeconfig, readerName); err != nil {
		return err
	}
	if err := v.agentReaderRoleIsExact(ctx, kubeconfig, readerName); err != nil {
		return err
	}

	account, err := grafanaServiceAccount(ctx, base, user, password, r.ServiceAccountName)
	if err != nil {
		return err
	}
	key := v.Namespace + "/" + v.Name
	firstToken, haveFirst := agentReaderTokensSeen.Load(key)

	if r.Disabled {
		if !account.IsDisabled {
			return errors.Errorf("service account %q is not disabled, yet agent_reader.disabled is true", r.ServiceAccountName)
		}
		fmt.Printf("  [verify] AGENT READER: service account %q is disabled\n", r.ServiceAccountName)
		if err := KubectlResourceAbsent(ctx, kubeconfig, "secret", readerName, v.Namespace); err != nil {
			return errors.Wrapf(err, "the token Secret %q outlived disabled", readerName)
		}
		fmt.Printf("  [verify] AGENT READER: Secret %q is gone\n", readerName)
		return v.proveFirstTokenRefused(ctx, base, firstToken, haveFirst)
	}

	if account.IsDisabled || account.Role != "Viewer" {
		return errors.Errorf("service account %q reads role %q, disabled %v; want a Viewer, enabled",
			r.ServiceAccountName, account.Role, account.IsDisabled)
	}
	token, err := v.agentReaderSecret(ctx, kubeconfig, readerName, r.TokenGeneration)
	if err != nil {
		return err
	}

	status, body, err := grafanaBearer(ctx, http.MethodGet, base+"/api/user", token, "")
	if err != nil || status != http.StatusOK {
		return errors.Errorf("the Secret's token was not accepted by /api/user (HTTP %d, %v): %s", status, err, firstLines(body, 3))
	}
	var me struct {
		Login string `json:"login"`
	}
	_ = json.Unmarshal([]byte(body), &me)
	if !strings.HasSuffix(me.Login, "-"+r.ServiceAccountName) {
		return errors.Errorf("the token signs in as %q, not the service account %q", me.Login, r.ServiceAccountName)
	}
	fmt.Printf("  [verify] AGENT READER: the token signs in as %s\n", me.Login)

	status, body, err = grafanaBearer(ctx, http.MethodGet, base+"/api/access-control/user/permissions", token, "")
	if err != nil || status != http.StatusOK {
		return errors.Errorf("reading the token's permissions failed (HTTP %d, %v): %s", status, err, firstLines(body, 3))
	}
	var permissions map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &permissions); err != nil {
		return errors.Wrapf(err, "parsing the token's permissions: %s", firstLines(body, 3))
	}
	if _, ok := permissions["datasources:query"]; !ok {
		return errors.New("the token may not query datasources (datasources:query is missing)")
	}
	var writes []string
	for action := range permissions {
		verb := action[strings.LastIndex(action, ":")+1:]
		switch verb {
		case "create", "write", "update", "delete", "push":
			writes = append(writes, action)
		}
	}
	if len(writes) > 0 {
		sort.Strings(writes)
		return errors.Errorf("the token holds write actions, so it is not read-only: %v", writes)
	}
	fmt.Printf("  [verify] AGENT READER: the token may query datasources and holds no write action\n")

	status, body, _ = grafanaBearer(ctx, http.MethodPost, base+"/api/dashboards/db", token,
		`{"dashboard": {"title": "agent-reader-write-attempt", "panels": []}, "overwrite": false}`)
	if status != http.StatusForbidden {
		return errors.Errorf("a dashboard write with the token answered HTTP %d, want 403: %s", status, firstLines(body, 3))
	}
	fmt.Printf("  [verify] AGENT READER: Grafana refuses the token a dashboard write (403)\n")

	tokens, err := grafanaServiceAccountTokens(ctx, base, user, password, account.ID)
	if err != nil {
		return err
	}
	if len(tokens) != 1 {
		return errors.Errorf("service account %q holds %d tokens, want exactly one: %v", r.ServiceAccountName, len(tokens), tokens)
	}
	if want := fmt.Sprintf("generation-%d-", r.TokenGeneration); !strings.HasPrefix(tokens[0], want) {
		return errors.Errorf("the account's one token is %q, not a generation %d token", tokens[0], r.TokenGeneration)
	}
	fmt.Printf("  [verify] AGENT READER: the account holds exactly one token (%s)\n", tokens[0])

	if haveFirst && firstToken.(string) != token {
		if err := v.proveFirstTokenRefused(ctx, base, firstToken, true); err != nil {
			return err
		}
	}
	agentReaderTokensSeen.Store(key, token)
	return nil
}

// proveFirstTokenRefused proves, in a second act, that the token the first
// act read no longer signs in.
func (v *GrafanaVerifier) proveFirstTokenRefused(ctx context.Context, base string, firstToken interface{}, haveFirst bool) error {
	if !haveFirst {
		fmt.Printf("  [verify] AGENT READER: no first-act token was read in this process; the refusal of the old token is not proven here\n")
		return nil
	}
	status, body, _ := grafanaBearer(ctx, http.MethodGet, base+"/api/user", firstToken.(string), "")
	if status != http.StatusUnauthorized {
		return errors.Errorf("the first act's token still answers HTTP %d, want 401: %s", status, firstLines(body, 3))
	}
	fmt.Printf("  [verify] AGENT READER: the first act's token is refused (401)\n")
	return nil
}

// agentReaderJobCompleted finds the module's Job (`<name>-agent-reader-<8
// hex>`) and requires it to have succeeded; both engines wait for it, so a
// verify that finds it running or failed has caught a broken await.
func (v *GrafanaVerifier) agentReaderJobCompleted(ctx context.Context, kubeconfig, readerName string) error {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "jobs", "-n", v.Namespace,
		"-o", "jsonpath={range .items[*]}{.metadata.name}={.status.succeeded}{\"\\n\"}{end}").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "listing jobs: %s", string(out))
	}
	prefix := readerName + "-"
	var found []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name, succeeded, _ := strings.Cut(line, "=")
		if strings.HasPrefix(name, prefix) && len(name) == len(prefix)+8 {
			if succeeded != "1" {
				return errors.Errorf("the agent reader's Job %s has not succeeded (succeeded=%q); kubectl logs job/%s -n %s",
					name, succeeded, name, v.Namespace)
			}
			found = append(found, name)
		}
	}
	if len(found) != 1 {
		return errors.Errorf("want exactly one agent reader Job %s<8 hex>, found %v", prefix, found)
	}
	fmt.Printf("  [verify] AGENT READER: Job %s completed\n", found[0])
	return nil
}

// agentReaderRoleIsExact requires the Role to grant exactly what the Job
// needs: get, update and delete on the one token Secret, create on
// Secrets (Kubernetes cannot scope create to a name), and get on the one
// ServiceAccount the owner reference names.
func (v *GrafanaVerifier) agentReaderRoleIsExact(ctx context.Context, kubeconfig, readerName string) error {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "role", readerName,
		"-n", v.Namespace, "-o", "json").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "reading Role %s: %s", readerName, string(out))
	}
	var role struct {
		Rules []struct {
			APIGroups     []string `json:"apiGroups"`
			Resources     []string `json:"resources"`
			ResourceNames []string `json:"resourceNames"`
			Verbs         []string `json:"verbs"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(out, &role); err != nil {
		return errors.Wrap(err, "parsing the Role")
	}
	// A rule's verbs and names are sets (Terraform's provider keeps verbs
	// as a set and writes them in its own order), so compare them sorted.
	var got []string
	for _, rule := range role.Rules {
		sort.Strings(rule.Verbs)
		sort.Strings(rule.ResourceNames)
		got = append(got, fmt.Sprintf("%s|%s|%s|%s", strings.Join(rule.APIGroups, ","), strings.Join(rule.Resources, ","),
			strings.Join(rule.ResourceNames, ","), strings.Join(rule.Verbs, ",")))
	}
	sort.Strings(got)
	want := []string{
		"|secrets||create",
		"|secrets|" + readerName + "|delete,get,update",
		"|serviceaccounts|" + readerName + "|get",
	}
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		return errors.Errorf("Role %s grants %v, want exactly %v", readerName, got, want)
	}
	fmt.Printf("  [verify] AGENT READER: Role %s grants exactly the token Secret and its owner\n", readerName)
	return nil
}

// agentReaderSecret reads the token Secret, requires the declared
// generation and the owner reference to the module's ServiceAccount, and
// returns the token.
func (v *GrafanaVerifier) agentReaderSecret(ctx context.Context, kubeconfig, readerName string, generation int) (string, error) {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "secret", readerName,
		"-n", v.Namespace, "-o", "json").CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "reading the token Secret %s: %s", readerName, string(out))
	}
	var secret struct {
		Metadata struct {
			OwnerReferences []struct {
				Kind string `json:"kind"`
				Name string `json:"name"`
			} `json:"ownerReferences"`
		} `json:"metadata"`
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(out, &secret); err != nil {
		return "", errors.Wrap(err, "parsing the token Secret")
	}
	owners := secret.Metadata.OwnerReferences
	if len(owners) != 1 || owners[0].Kind != "ServiceAccount" || owners[0].Name != readerName {
		return "", errors.Errorf("the token Secret's owners are %+v, want the ServiceAccount %s alone", owners, readerName)
	}
	decode := func(key string) string {
		raw, _ := base64.StdEncoding.DecodeString(secret.Data[key])
		return string(raw)
	}
	if got := decode("generation"); got != fmt.Sprint(generation) {
		return "", errors.Errorf("the token Secret holds generation %q, want %d", got, generation)
	}
	token := decode("token")
	if token == "" {
		return "", errors.New("the token Secret holds no token")
	}
	if strings.TrimSpace(token) != token {
		return "", errors.New("the token Secret's token carries surrounding whitespace, which no Authorization header accepts")
	}
	fmt.Printf("  [verify] AGENT READER: Secret %s holds a generation %d token, owned by ServiceAccount %s\n",
		readerName, generation, readerName)
	return token, nil
}

type grafanaServiceAccountView struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	IsDisabled bool   `json:"isDisabled"`
}

// grafanaServiceAccount finds one service account by its exact name, as the
// admin.
func grafanaServiceAccount(ctx context.Context, base, user, password, name string) (grafanaServiceAccountView, error) {
	status, body, err := grafanaBasic(ctx, http.MethodGet,
		base+"/api/serviceaccounts/search?perpage=1000&query="+url.QueryEscape(name), user, password)
	if err != nil || status != http.StatusOK {
		return grafanaServiceAccountView{}, errors.Errorf("searching service accounts (HTTP %d, %v): %s", status, err, firstLines(body, 3))
	}
	var found struct {
		ServiceAccounts []grafanaServiceAccountView `json:"serviceAccounts"`
	}
	if err := json.Unmarshal([]byte(body), &found); err != nil {
		return grafanaServiceAccountView{}, errors.Wrap(err, "parsing the service account search")
	}
	for _, sa := range found.ServiceAccounts {
		if sa.Name == name {
			return sa, nil
		}
	}
	return grafanaServiceAccountView{}, errors.Errorf("no service account named %q exists", name)
}

// grafanaServiceAccountTokens lists the names of an account's tokens, as
// the admin.
func grafanaServiceAccountTokens(ctx context.Context, base, user, password string, id int) ([]string, error) {
	status, body, err := grafanaBasic(ctx, http.MethodGet, fmt.Sprintf("%s/api/serviceaccounts/%d/tokens", base, id), user, password)
	if err != nil || status != http.StatusOK {
		return nil, errors.Errorf("listing the account's tokens (HTTP %d, %v): %s", status, err, firstLines(body, 3))
	}
	var tokens []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(body), &tokens); err != nil {
		return nil, errors.Wrap(err, "parsing the account's tokens")
	}
	names := make([]string, 0, len(tokens))
	for _, t := range tokens {
		names = append(names, t.Name)
	}
	return names, nil
}

// grafanaBearer and grafanaBasic make one request and return its status:
// unlike GrafanaVerifier.request, a refusal is an answer here, not a
// failure to retry.
func grafanaBearer(ctx context.Context, method, target, token, body string) (int, string, error) {
	return grafanaOnce(ctx, method, target, body, func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
	})
}

func grafanaBasic(ctx context.Context, method, target, user, password string) (int, string, error) {
	return grafanaOnce(ctx, method, target, "", func(req *http.Request) { req.SetBasicAuth(user, password) })
}

func grafanaOnce(ctx context.Context, method, target, body string, authenticate func(*http.Request)) (int, string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, method, target, bytes.NewReader([]byte(body)))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	authenticate(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), err
}
