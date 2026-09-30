package verify

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// The behavioral-sso scenario's sign-in proof. Typed Google sign-in is
// only worth anything if a person actually gets in (and the wrong person
// does not), so this proof signs in end to end rather than reading the
// rendered configuration:
//   - Grafana's /login/google redirects to Google's real authorization
//     endpoint with the declared client id, the root_url callback and the
//     declared hosted domain;
//   - the code exchange reaches a stand-in for Google's token endpoint,
//     which answers only to the declared client secret, so a successful
//     sign-in proves the secret travelled from the manifest through the
//     module-owned `<name>-sso` Secret into Grafana's environment;
//   - an address the role mapping names signs in as Admin, another
//     address in the allowed domain as Viewer, and an address outside it
//     is refused (Grafana's allowed_domains check on the email);
//   - the secret appears in neither the chart's ConfigMap nor the
//     Deployment, and the pod template carries the credentials checksum;
//   - the admin screen cannot edit any provider (the SSO settings API
//     refuses both the declared provider and an undeclared one).
// Every expected value below is a literal of the scenario file
// (e2e/scenarios/behavioral-sso.yaml); both engines must satisfy the same
// strings.

const (
	signInStandInName    = "google-stand-in"
	signInStandInPort    = 8080
	signInClientID       = "e2e-grafana-client.apps.googleusercontent.com"
	signInClientSecret   = "e2e-stand-in-client-secret"
	signInRootURL        = "http://127.0.0.1:13000"
	signInHostedDomain   = "example.org"
	signInGoogleAuthHost = "accounts.google.com"
	signInAdminEmail     = "lead@example.org"
	signInViewerEmail    = "member@example.org"
	signInOutsiderEmail  = "someone@example.net"
	signInSecretSuffix   = "-sso"
	signInChecksumKey    = "checksum/credentials"
)

// googleStandInScript answers Google's token endpoint for exactly one
// client: the client id and secret the scenario declares, sent either in
// the Basic header or the form (golang.org/x/oauth2 tries both). The
// authorization code is the base64url of the email to sign in as; the ID
// token carries Google's claims (sub, email, email_verified, hd, iat, exp),
// which Grafana reads without a signature check unless validate_id_token
// is on. It also answers the refresh grant, as Google does: Grafana checks
// the ID token's exp on every request and refreshes the moment it reads
// expired, so a token without exp logs the person straight back out.
const googleStandInScript = `import base64, http.server, json, time, urllib.parse
CLIENT_ID = "` + signInClientID + `"
CLIENT_SECRET = "` + signInClientSecret + `"
def seg(d):
    return base64.urlsafe_b64encode(json.dumps(d).encode()).rstrip(b"=").decode()
class StandIn(http.server.BaseHTTPRequestHandler):
    def reply(self, code, body):
        data = json.dumps(body).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)
    def do_POST(self):
        form = urllib.parse.parse_qs(self.rfile.read(int(self.headers.get("Content-Length", 0))).decode())
        cid = form.get("client_id", [""])[0]
        secret = form.get("client_secret", [""])[0]
        auth = self.headers.get("Authorization", "")
        if auth.startswith("Basic "):
            cid, _, secret = base64.b64decode(auth[6:]).decode().partition(":")
            cid, secret = urllib.parse.unquote_plus(cid), urllib.parse.unquote_plus(secret)
        if cid != CLIENT_ID or secret != CLIENT_SECRET:
            print(json.dumps({"exchange": "refused"}), flush=True)
            return self.reply(401, {"error": "invalid_client"})
        if form.get("grant_type", [""])[0] == "refresh_token":
            email = form.get("refresh_token", ["rt-"])[0][3:]
        else:
            code = form.get("code", [""])[0]
            email = base64.urlsafe_b64decode(code + "=" * (-len(code) % 4)).decode()
        now = int(time.time())
        claims = {"sub": "e2e-" + email, "email": email, "email_verified": True, "name": email,
                  "hd": email.split("@")[1], "iat": now, "exp": now + 3600}
        print(json.dumps({"exchange": "issued", "grant": form.get("grant_type", [""])[0], "email": email}), flush=True)
        self.reply(200, {"access_token": "at-" + email, "token_type": "Bearer", "expires_in": 3600,
                         "refresh_token": "rt-" + email, "id_token": seg({"alg": "RS256", "typ": "JWT"}) + "." + seg(claims) + ".c2ln"})
http.server.HTTPServer(("", 8080), StandIn).serve_forever()
`

func (v *GrafanaVerifier) proveSignIn(ctx context.Context, kubeconfig, base, adminUser, adminPassword string) error {
	if err := v.deployGoogleStandIn(ctx, kubeconfig); err != nil {
		return err
	}

	if err := v.proveRedirectToGoogle(base); err != nil {
		return err
	}

	// A fresh Service can take a few seconds to route even after its
	// endpoint exists (kube-proxy programs it asynchronously). Until the
	// stand-in has seen one exchange, a failed sign-in means Grafana could
	// not reach it yet, so it is retried; once an exchange arrived, every
	// failure is a real one.
	var adminRole string
	var err error
	reachDeadline := time.Now().Add(2 * time.Minute)
	for {
		adminRole, err = v.signInAs(base, signInAdminEmail)
		if err == nil || v.standInExchanges(ctx, kubeconfig) > 0 || time.Now().After(reachDeadline) {
			break
		}
		fmt.Printf("  [verify] SIGN-IN: Grafana has not reached the stand-in yet (the Service is still routing); retrying\n")
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return errors.Wrapf(err, "SIGN-IN: %s could not sign in; the code exchange needs the declared client secret, so check the %s%s Secret and GF_AUTH_GOOGLE_CLIENT_SECRET on the Deployment\n%s", signInAdminEmail, v.Name, signInSecretSuffix, v.signInDiagnostics(ctx, kubeconfig))
	}
	if adminRole != "Admin" {
		return errors.Errorf("SIGN-IN: %s signed in as %q, want Admin from role_attribute_path\n%s", signInAdminEmail, adminRole, v.signInDiagnostics(ctx, kubeconfig))
	}
	fmt.Printf("  [verify] SIGN-IN: %s signed in through the stand-in Google as Admin (the client secret reached the exchange)\n", signInAdminEmail)

	viewerRole, err := v.signInAs(base, signInViewerEmail)
	if err != nil {
		return errors.Wrapf(err, "SIGN-IN: %s, inside allowed_domains, could not sign in\n%s", signInViewerEmail, v.signInDiagnostics(ctx, kubeconfig))
	}
	if viewerRole != "Viewer" {
		return errors.Errorf("SIGN-IN: %s signed in as %q, want Viewer", signInViewerEmail, viewerRole)
	}
	fmt.Printf("  [verify] SIGN-IN: %s signed in as Viewer\n", signInViewerEmail)

	if role, err := v.signInAs(base, signInOutsiderEmail); err == nil {
		return errors.Errorf("SIGN-IN: %s, outside allowed_domains, signed in as %q; Grafana must refuse any email outside the declared domains", signInOutsiderEmail, role)
	}
	fmt.Printf("  [verify] SIGN-IN: %s (outside allowed_domains) was refused\n", signInOutsiderEmail)

	if err := v.proveSecretStaysOutOfConfig(ctx, kubeconfig); err != nil {
		return err
	}
	return v.proveAdminScreenLocked(base, adminUser, adminPassword)
}

// deployGoogleStandIn starts the stand-in Pod and Service in Grafana's
// namespace and waits for it to be Ready. Grafana reaches it only at code
// exchange, so it can start after Grafana.
func (v *GrafanaVerifier) deployGoogleStandIn(ctx context.Context, kubeconfig string) error {
	indented := "            " + strings.ReplaceAll(strings.TrimRight(googleStandInScript, "\n"), "\n", "\n            ")
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: %[1]s
  namespace: %[2]s
  labels:
    app: %[1]s
spec:
  containers:
    - name: stand-in
      image: python:3.12-alpine
      command:
        - python
        - -u
        - -c
        - |
%[4]s
      ports:
        - containerPort: %[3]d
      # Ready means listening: without the probe the pod reads Ready while
      # Python is still starting, and Grafana's code exchange is refused.
      readinessProbe:
        tcpSocket:
          port: %[3]d
        periodSeconds: 1
---
apiVersion: v1
kind: Service
metadata:
  name: %[1]s
  namespace: %[2]s
spec:
  selector:
    app: %[1]s
  ports:
    - port: %[3]d
      targetPort: %[3]d
`, signInStandInName, v.Namespace, signInStandInPort, indented)
	if err := kubectlApplyStdin(ctx, kubeconfig, manifest); err != nil {
		return errors.Wrap(err, "SIGN-IN: deploying the stand-in Google token endpoint failed")
	}
	if err := kubectlWait(ctx, kubeconfig, "pod", signInStandInName, v.Namespace, "condition=Ready", 3*time.Minute); err != nil {
		return errors.Wrap(err, "SIGN-IN: the stand-in Google token endpoint never became ready")
	}
	// A Ready pod reaches the Service's endpoints a moment later; wait for
	// the endpoint so the first exchange is not refused.
	deadline := time.Now().Add(time.Minute)
	for time.Now().Before(deadline) {
		ip, _ := kubectlGetJSONPath(ctx, kubeconfig, "endpoints", signInStandInName, v.Namespace, "{.subsets[0].addresses[0].ip}")
		if ip != "" {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return errors.New("SIGN-IN: the stand-in Google token endpoint is Ready but its Service has no endpoint")
}

// standInExchanges counts the code exchanges the stand-in has answered or
// refused; zero means Grafana has not reached it.
func (v *GrafanaVerifier) standInExchanges(ctx context.Context, kubeconfig string) int {
	out, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"logs", signInStandInName, "-n", v.Namespace).CombinedOutput()
	return strings.Count(string(out), `"exchange"`)
}

// signInDiagnostics gathers what a failed sign-in leaves behind: Grafana's
// own account of the login (its oauth and login log lines) and what the
// stand-in saw at the token endpoint (issued, refused, or nothing at all,
// which means Grafana never reached it).
func (v *GrafanaVerifier) signInDiagnostics(ctx context.Context, kubeconfig string) string {
	grafanaLog, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"logs", "deployment/"+v.Name, "-n", v.Namespace, "-c", "grafana", "--tail=400").CombinedOutput()
	var relevant []string
	for _, line := range strings.Split(string(grafanaLog), "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "oauth") || strings.Contains(lower, "login") || strings.Contains(lower, "google") ||
			strings.Contains(lower, "role") || strings.Contains(lower, "level=error") || strings.Contains(lower, "level=warn") {
			relevant = append(relevant, line)
		}
	}
	if len(relevant) > 15 {
		relevant = relevant[len(relevant)-15:]
	}
	standInLog, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"logs", signInStandInName, "-n", v.Namespace, "--tail=20").CombinedOutput()
	ini, _ := kubectlGetJSONPath(ctx, kubeconfig, "configmap", v.Name, v.Namespace, `{.data.grafana\.ini}`)
	var googleSection []string
	inSection := false
	for _, line := range strings.Split(ini, "\n") {
		if strings.HasPrefix(line, "[") {
			inSection = strings.HasPrefix(line, "[auth.google]")
		}
		if inSection {
			googleSection = append(googleSection, line)
		}
	}
	return fmt.Sprintf("    grafana.ini [auth.google] as rendered:\n      %s\n    grafana (oauth/login/role lines):\n      %s\n    stand-in token endpoint:\n      %s",
		strings.Join(googleSection, "\n      "), strings.Join(relevant, "\n      "),
		strings.ReplaceAll(strings.TrimSpace(string(standInLog)), "\n", "\n      "))
}

// signInClient keeps cookies (Grafana's oauth_state and PKCE cookies must
// come back on the callback) and never follows redirects, so each hop is
// asserted rather than trusted.
func signInClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// startSignIn opens /login/google and returns the authorization redirect.
func startSignIn(client *http.Client, base string) (*url.URL, error) {
	resp, err := client.Get(base + "/login/google")
	if err != nil {
		return nil, errors.Wrap(err, "opening /login/google")
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusTemporaryRedirect {
		return nil, errors.Errorf("/login/google answered HTTP %d, want a redirect to Google", resp.StatusCode)
	}
	return url.Parse(resp.Header.Get("Location"))
}

func (v *GrafanaVerifier) proveRedirectToGoogle(base string) error {
	location, err := startSignIn(signInClient(), base)
	if err != nil {
		return errors.Wrap(err, "SIGN-IN: Google sign-in is not offered")
	}
	query := location.Query()
	checks := map[string][2]string{
		"authorization host": {location.Host, signInGoogleAuthHost},
		"client_id":          {query.Get("client_id"), signInClientID},
		"redirect_uri":       {query.Get("redirect_uri"), signInRootURL + "/login/google"},
		"hd":                 {query.Get("hd"), signInHostedDomain},
	}
	for what, pair := range checks {
		if pair[0] != pair[1] {
			return errors.Errorf("SIGN-IN: the redirect to Google carries %s %q, want %q (redirect: %s)", what, pair[0], pair[1], location.String())
		}
	}
	fmt.Printf("  [verify] SIGN-IN: /login/google redirects to %s with the declared client, the root_url callback and hd=%s\n", signInGoogleAuthHost, signInHostedDomain)
	return nil
}

// signInAs runs one whole sign-in as email and returns the Grafana role
// the session holds; an error means Grafana issued no session.
func (v *GrafanaVerifier) signInAs(base, email string) (string, error) {
	client := signInClient()
	location, err := startSignIn(client, base)
	if err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString([]byte(email))
	callback := fmt.Sprintf("%s/login/google?code=%s&state=%s", base, url.QueryEscape(code), url.QueryEscape(location.Query().Get("state")))
	resp, err := client.Get(callback)
	if err != nil {
		return "", errors.Wrap(err, "calling back into Grafana")
	}
	resp.Body.Close()

	orgs, err := client.Get(base + "/api/user/orgs")
	if err != nil {
		return "", errors.Wrap(err, "reading the session's organizations")
	}
	body, _ := io.ReadAll(orgs.Body)
	orgs.Body.Close()
	if orgs.StatusCode != http.StatusOK {
		return "", errors.Errorf("no session after the callback (HTTP %d from /api/user/orgs; the callback answered HTTP %d, Location %q)", orgs.StatusCode, resp.StatusCode, resp.Header.Get("Location"))
	}
	for _, role := range []string{"Admin", "Editor", "Viewer"} {
		if strings.Contains(string(body), `"role":"`+role+`"`) {
			return role, nil
		}
	}
	return "", errors.Errorf("the session holds no recognizable role: %s", firstLines(string(body), 2))
}

func (v *GrafanaVerifier) proveSecretStaysOutOfConfig(ctx context.Context, kubeconfig string) error {
	for _, object := range []string{"configmap/" + v.Name, "deployment/" + v.Name} {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", object, "-n", v.Namespace, "-o", "yaml").CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "SIGN-IN: reading %s: %s", object, string(out))
		}
		if strings.Contains(string(out), signInClientSecret) {
			return errors.Errorf("SIGN-IN: the client secret appears in %s; it must reach Grafana only through the %s%s Secret", object, v.Name, signInSecretSuffix)
		}
	}
	ref, _ := kubectlGetJSONPath(ctx, kubeconfig, "deployment", v.Name, v.Namespace,
		`{.spec.template.spec.containers[?(@.name=="grafana")].env[?(@.name=="GF_AUTH_GOOGLE_CLIENT_SECRET")].valueFrom.secretKeyRef.name}`)
	if ref != v.Name+signInSecretSuffix {
		return errors.Errorf("SIGN-IN: GF_AUTH_GOOGLE_CLIENT_SECRET reads Secret %q, want %q", ref, v.Name+signInSecretSuffix)
	}
	checksum, _ := kubectlGetJSONPath(ctx, kubeconfig, "deployment", v.Name, v.Namespace,
		`{.spec.template.metadata.annotations.checksum/credentials}`)
	if len(checksum) != 64 {
		return errors.Errorf("SIGN-IN: the pod template carries %s %q, want the SHA-256 of the sign-in Secret; without it a rotated secret never reaches Grafana", signInChecksumKey, checksum)
	}
	fmt.Printf("  [verify] SIGN-IN: the secret is absent from the ConfigMap and the Deployment, read from %s%s, and fingerprinted on the pod template\n", v.Name, signInSecretSuffix)
	return nil
}

// proveAdminScreenLocked tries what an admin would do in Administration >
// Authentication: save provider settings through the SSO settings API.
// Both the declared provider and an undeclared one must be refused.
func (v *GrafanaVerifier) proveAdminScreenLocked(base, adminUser, adminPassword string) error {
	for _, provider := range []string{"google", "github"} {
		body := `{"settings":{"enabled":true,"clientId":"out-of-band","clientSecret":"out-of-band"}}`
		req, err := http.NewRequest(http.MethodPut, base+"/api/v1/sso-settings/"+provider, strings.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(adminUser, adminPassword)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrapf(err, "SIGN-IN: calling the SSO settings API for %s", provider)
		}
		answer, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return errors.Errorf("SIGN-IN: an admin saved %s sign-in settings through the SSO settings API (HTTP %d); the manifest must own sign-in", provider, resp.StatusCode)
		}
		fmt.Printf("  [verify] SIGN-IN: the admin screen cannot edit %s (HTTP %d: %s)\n", provider, resp.StatusCode, firstLines(strings.TrimSpace(string(answer)), 1))
	}
	return nil
}
