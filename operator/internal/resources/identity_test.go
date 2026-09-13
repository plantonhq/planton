package resources

import (
	"encoding/json"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
)

const (
	testIdentityRealm = "planton"
	// kcDBEngine / portNameHTTP mirror the builder literals under test.
	kcDBEngine   = "postgres"
	portNameHTTP = "http"
	// stringTrue mirrors Keycloak's stringly-typed protocol-mapper config.
	stringTrue = "true"
)

func testIdentityConfig() IdentityConfig {
	return IdentityConfig{
		CRName:          testIdentityRealm,
		Namespace:       "default",
		Realm:           testIdentityRealm,
		PublicURL:       "http://planton.example.com",
		RealmImportHash: "abc123",
		PostgreSQL:      PostgreSQLConnection(testIdentityRealm, "default"),
	}
}

func TestIdentityIssuerURL(t *testing.T) {
	got := IdentityIssuerURL("https://planton.example.com", testIdentityRealm)
	want := "https://planton.example.com/idp/realms/planton"
	if got != want {
		t.Errorf("issuer = %s, want %s", got, want)
	}
}

func TestIdentityDeployment_Image(t *testing.T) {
	deploy := IdentityDeployment(testIdentityConfig())
	want := IdentityDefaultImageRepo + ":" + IdentityDefaultImageTag
	if got := deploy.Spec.Template.Spec.Containers[0].Image; got != want {
		t.Errorf("image = %s, want %s", got, want)
	}

	cfg := testIdentityConfig()
	cfg.ImageRepository = "example.com/keycloak"
	cfg.ImageTag = "custom"
	deploy = IdentityDeployment(cfg)
	if got := deploy.Spec.Template.Spec.Containers[0].Image; got != "example.com/keycloak:custom" {
		t.Errorf("image override = %s, want example.com/keycloak:custom", got)
	}
}

func TestIdentityDeployment_ServingContract(t *testing.T) {
	deploy := IdentityDeployment(testIdentityConfig())
	envMap := envVarMap(deploy.Spec.Template.Spec.Containers[0].Env)

	// The path prefix and the full public URL are what make same-hostname
	// serving behind the ingress work; the issuer in tokens derives from them.
	if envMap["KC_HTTP_RELATIVE_PATH"] != "/idp" {
		t.Errorf("KC_HTTP_RELATIVE_PATH = %q, want /idp", envMap["KC_HTTP_RELATIVE_PATH"])
	}
	if envMap["KC_HOSTNAME"] != "http://planton.example.com/idp" {
		t.Errorf("KC_HOSTNAME = %q, want http://planton.example.com/idp", envMap["KC_HOSTNAME"])
	}
	if envMap["KC_PROXY_HEADERS"] != "xforwarded" {
		t.Errorf("KC_PROXY_HEADERS = %q, want xforwarded", envMap["KC_PROXY_HEADERS"])
	}
	if envMap["KC_HTTP_ENABLED"] != stringTrue {
		t.Errorf("KC_HTTP_ENABLED = %q, want true (TLS terminates at the ingress)", envMap["KC_HTTP_ENABLED"])
	}
	// Health probes must not move with the serving path.
	if envMap["KC_HTTP_MANAGEMENT_RELATIVE_PATH"] != "/" {
		t.Errorf("KC_HTTP_MANAGEMENT_RELATIVE_PATH = %q, want /", envMap["KC_HTTP_MANAGEMENT_RELATIVE_PATH"])
	}

	args := deploy.Spec.Template.Spec.Containers[0].Args
	if len(args) != 3 || args[0] != "start" || args[1] != "--import-realm" ||
		args[2] != "--spi-theme--default=planton" {
		t.Errorf("args = %v, want [start --import-realm --spi-theme--default=planton]", args)
	}
}

func TestIdentityDeployment_Database(t *testing.T) {
	deploy := IdentityDeployment(testIdentityConfig())
	envMap := envVarMap(deploy.Spec.Template.Spec.Containers[0].Env)

	if envMap["KC_DB"] != kcDBEngine {
		t.Errorf("KC_DB = %q, want %s", envMap["KC_DB"], kcDBEngine)
	}
	if envMap["KC_DB_URL_DATABASE"] != "keycloak" {
		t.Errorf("KC_DB_URL_DATABASE = %q, want keycloak", envMap["KC_DB_URL_DATABASE"])
	}
	if envMap["KC_DB_PASSWORD"] != fromSecretRef {
		t.Error("KC_DB_PASSWORD must come from the credential Secret, never a literal")
	}

	// The control plane self-provisions only its own databases; identity
	// carries its own idempotent CREATE DATABASE init step.
	inits := deploy.Spec.Template.Spec.InitContainers
	if len(inits) != 1 || inits[0].Name != "ensure-database" {
		t.Fatalf("expected one ensure-database init container, got %v", inits)
	}
	script := strings.Join(inits[0].Command, " ")
	if !strings.Contains(script, "CREATE DATABASE") || !strings.Contains(script, "pg_database") {
		t.Errorf("init script must create the database idempotently, got %q", script)
	}
}

func TestIdentityRecoveryAdminJob(t *testing.T) {
	cfg := testIdentityConfig()
	cfg.ImageRepository = "example.com/keycloak"
	cfg.ImageTag = "custom"
	job := IdentityRecoveryAdminJob(cfg)
	deploy := IdentityDeployment(cfg)

	if job.Name != "planton-identity-recovery-admin" || job.Namespace != "default" {
		t.Errorf("job named after the platform: %s/%s", job.Namespace, job.Name)
	}
	pod := job.Spec.Template.Spec
	if len(pod.Containers) != 1 || len(pod.InitContainers) != 0 {
		t.Fatalf("one container and no ensure-database step (the restored database exists): %+v", pod)
	}
	c := pod.Containers[0]

	// The same image as the server, so the command speaks the same
	// schema version the restored database carries.
	if c.Image != deploy.Spec.Template.Spec.Containers[0].Image {
		t.Errorf("job image %q must be the identity server's %q", c.Image, deploy.Spec.Template.Spec.Containers[0].Image)
	}

	// Keycloak's recovery command, the recovery username, the password
	// read from the environment (the command accepts no literal).
	wantArgs := []string{"bootstrap-admin", "user", "--username", "planton-recovery", "--password:env", "KC_BOOTSTRAP_ADMIN_PASSWORD"}
	if strings.Join(c.Args, " ") != strings.Join(wantArgs, " ") {
		t.Errorf("args = %v, want %v", c.Args, wantArgs)
	}

	// The database env is byte-identical to the Deployment's: the Job and
	// the server can never disagree about which database the realm is in.
	jobEnv := envVarMap(c.Env)
	deployEnv := envVarMap(deploy.Spec.Template.Spec.Containers[0].Env)
	for _, key := range []string{"KC_DB", "KC_DB_URL_HOST", "KC_DB_URL_PORT", "KC_DB_URL_DATABASE", "KC_DB_USERNAME", "KC_DB_PASSWORD"} {
		if jobEnv[key] != deployEnv[key] {
			t.Errorf("%s: job %q, deployment %q", key, jobEnv[key], deployEnv[key])
		}
	}
	if jobEnv["KC_BOOTSTRAP_ADMIN_PASSWORD"] != fromSecretRef {
		t.Error("the recovery password comes from the recovery Secret, never a literal")
	}
	for _, env := range c.Env {
		if env.Name == "KC_BOOTSTRAP_ADMIN_PASSWORD" && env.ValueFrom.SecretKeyRef.Name != "planton-identity-recovery-admin" {
			t.Errorf("the recovery password Secret is the job's own: %q", env.ValueFrom.SecretKeyRef.Name)
		}
	}
	if _, present := jobEnv["KC_BOOTSTRAP_ADMIN_USERNAME"]; present {
		t.Error("the username is the command's --username flag, not the startup env the server reads")
	}

	if pod.RestartPolicy != "Never" || job.Spec.BackoffLimit == nil || *job.Spec.BackoffLimit != 2 ||
		job.Spec.ActiveDeadlineSeconds == nil || *job.Spec.ActiveDeadlineSeconds != 600 {
		t.Errorf("one-shot shape: restartPolicy %q, backoffLimit %v, activeDeadline %v",
			pod.RestartPolicy, job.Spec.BackoffLimit, job.Spec.ActiveDeadlineSeconds)
	}
}

func TestIdentityDeployment_RealmImportMountAndRestartHash(t *testing.T) {
	deploy := IdentityDeployment(testIdentityConfig())

	vols := deploy.Spec.Template.Spec.Volumes
	if len(vols) != 2 || vols[0].Secret == nil || vols[0].Secret.SecretName != "planton-identity-realm-import" {
		t.Fatalf("expected the realm-import Secret volume + the theme volume, got %v", vols)
	}

	// A changed realm import must roll the pod (Secret content changes alone
	// do not restart consumers).
	ann := deploy.Spec.Template.Annotations[IdentityRealmImportHashAnnotation]
	if ann != "abc123" {
		t.Errorf("realm import hash annotation = %q, want abc123", ann)
	}

	// Two Keycloaks importing one realm into one database mid-rollout is a
	// conflict, not HA.
	if deploy.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType {
		t.Errorf("strategy = %s, want Recreate", deploy.Spec.Strategy.Type)
	}
}

func TestIdentityDeployment_ProbesAndResources(t *testing.T) {
	deploy := IdentityDeployment(testIdentityConfig())
	container := deploy.Spec.Template.Spec.Containers[0]

	if container.StartupProbe == nil || container.StartupProbe.HTTPGet == nil ||
		container.StartupProbe.HTTPGet.Path != "/health/started" {
		t.Fatal("expected HTTP startup probe on /health/started")
	}
	// First boot runs schema migrations + realm import; the kubelet must not
	// kill the JVM mid-import.
	if container.StartupProbe.FailureThreshold < 30 {
		t.Errorf("startup failureThreshold = %d, want a generous first-boot window", container.StartupProbe.FailureThreshold)
	}
	if container.ReadinessProbe.HTTPGet.Path != "/health/ready" {
		t.Errorf("readiness path = %s, want /health/ready", container.ReadinessProbe.HTTPGet.Path)
	}
	if container.LivenessProbe.HTTPGet.Path != "/health/live" {
		t.Errorf("liveness path = %s, want /health/live", container.LivenessProbe.HTTPGet.Path)
	}

	// Explicit memory floor (the data-layer OOM lesson) and no CPU
	// limit so the first-boot import is never throttled.
	if container.Resources.Requests.Memory().IsZero() {
		t.Error("expected an explicit memory request floor")
	}
	if !container.Resources.Limits.Cpu().IsZero() {
		t.Error("expected no CPU limit")
	}
}

func TestIdentityService(t *testing.T) {
	svc := IdentityService("planton", "default", nil)
	if svc.Name != "planton-identity" {
		t.Errorf("service name = %s, want planton-identity", svc.Name)
	}
	ports := svc.Spec.Ports
	if len(ports) != 1 || ports[0].Name != portNameHTTP || ports[0].Port != 80 {
		t.Errorf("expected one http port 80, got %v", ports)
	}
	if ports[0].TargetPort.IntValue() != 8080 {
		t.Errorf("target port = %d, want 8080", ports[0].TargetPort.IntValue())
	}
}

// parseTestRealmImport renders and parses the realm import used by the
// realm-content tests below.
func parseTestRealmImport(t *testing.T) map[string]any {
	t.Helper()
	data, err := IdentityRealmImport(IdentityRealmImportConfig{
		Realm:         testIdentityRealm,
		PublicURL:     "https://planton.example.com",
		ClientSecret:  "client-secret-value",
		UsersSecret:   "users-secret-value",
		AdminEmail:    "admin@example.com",
		AdminPassword: "one-time-password",
	})
	if err != nil {
		t.Fatal(err)
	}

	var realm map[string]any
	if err := json.Unmarshal(data, &realm); err != nil {
		t.Fatalf("realm import is not valid JSON: %v", err)
	}
	return realm
}

func TestIdentityRealmImport(t *testing.T) {
	realm := parseTestRealmImport(t)

	if realm["realm"] != testIdentityRealm {
		t.Errorf("realm = %v, want planton", realm["realm"])
	}
	// The friction ladder's plain-HTTP step must be able to log in.
	if realm["sslRequired"] != "none" {
		t.Errorf("sslRequired = %v, want none (the ingress owns transport security)", realm["sslRequired"])
	}
	// The realm is the access boundary: no self-registration.
	if realm["registrationAllowed"] != false {
		t.Error("registrationAllowed must be false")
	}

	clients := realm["clients"].([]any)
	if len(clients) != 3 {
		t.Fatalf("expected the console + user-directory + CLI clients, got %d", len(clients))
	}
	client := clients[0].(map[string]any)
	if client["clientId"] != IdentityConsoleClientID {
		t.Errorf("clientId = %v, want %s", client["clientId"], IdentityConsoleClientID)
	}
	if client["publicClient"] != false || client["secret"] != "client-secret-value" {
		t.Error("expected a confidential client carrying the generated secret")
	}
	// Exact redirect URIs -- no wildcards on an auth surface. Two callbacks:
	// the sign-in stack's own, and the device-auth broker's (desktop + CLI
	// sign-ins ride this same confidential client through the broker).
	redirects := client["redirectUris"].([]any)
	wantRedirects := []string{
		"https://planton.example.com/api/auth/callback/iam",
		"https://planton.example.com/api/device/auth/callback",
	}
	if len(redirects) != len(wantRedirects) {
		t.Fatalf("redirectUris = %v, want %v", redirects, wantRedirects)
	}
	for i, uri := range wantRedirects {
		if redirects[i] != uri {
			t.Errorf("redirectUris[%d] = %v, want %s", i, redirects[i], uri)
		}
	}
	attrs := client["attributes"].(map[string]any)
	if attrs["post.logout.redirect.uris"] != "https://planton.example.com/login" {
		t.Errorf("post-logout URI = %v, want the login page", attrs["post.logout.redirect.uris"])
	}

	// Keycloak does not stamp a custom API audience by default; without this
	// mapper the control plane rejects every browser token.
	mappers := client["protocolMappers"].([]any)
	if len(mappers) != 1 {
		t.Fatalf("expected the audience protocol mapper, got %d mappers", len(mappers))
	}
	mapper := mappers[0].(map[string]any)
	if mapper["protocolMapper"] != "oidc-audience-mapper" {
		t.Errorf("mapper = %v, want oidc-audience-mapper", mapper["protocolMapper"])
	}
	mapperCfg := mapper["config"].(map[string]any)
	if mapperCfg["included.custom.audience"] != IDPAPIAudience {
		t.Errorf("audience = %v, want %s", mapperCfg["included.custom.audience"], IDPAPIAudience)
	}
	if mapperCfg["access.token.claim"] != stringTrue {
		t.Error("the audience must land in the ACCESS token (what the API validates)")
	}

	// offline_access must be grantable or the console's token exchange fails
	// outright (the sign-in stack always requests it for session refresh).
	optionalScopes := client["optionalClientScopes"].([]any)
	foundOffline := false
	for _, s := range optionalScopes {
		if s == "offline_access" {
			foundOffline = true
		}
	}
	if !foundOffline {
		t.Error("offline_access missing from the client's optional scopes")
	}
}

func TestIdentityRealmImport_UserDirectoryClient(t *testing.T) {
	realm := parseTestRealmImport(t)

	// The user-directory client: client-credentials only, no Planton audience
	// mapper (its token talks to the identity server's admin API, never to
	// Planton), and its entire privilege is the service-account role mapping
	// asserted below.
	clients := realm["clients"].([]any)
	users := clients[1].(map[string]any)
	if users["clientId"] != IdentityUsersClientID {
		t.Fatalf("clients[1] = %v, want %s", users["clientId"], IdentityUsersClientID)
	}
	if users["serviceAccountsEnabled"] != true || users["standardFlowEnabled"] != false {
		t.Error("user-directory client must be client-credentials only")
	}
	if users["secret"] != "users-secret-value" {
		t.Errorf("user-directory secret = %v, want the generated secret", users["secret"])
	}
	if mappers := users["protocolMappers"].([]any); len(mappers) != 0 {
		t.Error("user-directory client must carry no Planton audience mapper")
	}

	// The role grant rides the import as the service account's clientRoles
	// (proven live: without this entry the service account exists but holds
	// no roles and every admin API call is 403).
	sa := serviceAccountUser(t, realm)
	if sa["serviceAccountClientId"] != IdentityUsersClientID {
		t.Errorf("service account clientId = %v, want %s", sa["serviceAccountClientId"], IdentityUsersClientID)
	}
	roles := sa["clientRoles"].(map[string]any)["realm-management"].([]any)
	if len(roles) != 2 || roles[0] != "manage-users" || roles[1] != "view-users" {
		t.Errorf("realm-management roles = %v, want exactly [manage-users view-users] -- least privilege", roles)
	}
	if _, hasCreds := sa["credentials"]; hasCreds {
		t.Error("the service-account user must carry no password credential")
	}
}

// The CLI client is what makes `planton login` against a self-hosted install
// possible at all: public because a CLI binary cannot hold a secret, PKCE
// S256 enforced because that is what makes a public client safe, exact
// loopback redirect URIs because Keycloak has no port wildcards and the CLI's
// callback port is fixed.
func TestIdentityRealmImport_CLIClient(t *testing.T) {
	realm := parseTestRealmImport(t)

	clients := realm["clients"].([]any)
	cli := clients[2].(map[string]any)
	if cli["clientId"] != IdentityCLIClientID {
		t.Fatalf("clients[2] = %v, want %s", cli["clientId"], IdentityCLIClientID)
	}
	if cli["publicClient"] != true {
		t.Error("CLI client must be public (a CLI binary cannot hold a secret)")
	}
	if cli["standardFlowEnabled"] != true || cli["directAccessGrantsEnabled"] != false {
		t.Error("CLI client must use the standard (authorization-code) flow only")
	}
	if cli["secret"] != "" {
		t.Errorf("CLI client must carry no secret, got %v", cli["secret"])
	}

	attrs := cli["attributes"].(map[string]any)
	if attrs[IdentityPKCEMethodAttribute] != IdentityPKCEMethodS256 {
		t.Error("CLI client must ENFORCE PKCE S256 -- without the attribute Keycloak accepts plain non-PKCE requests on a public client")
	}

	redirects := cli["redirectUris"].([]any)
	want := IdentityCLIRedirectURIs()
	if len(redirects) != len(want) {
		t.Fatalf("redirectUris = %v, want %v", redirects, want)
	}
	for i, uri := range want {
		if redirects[i] != uri {
			t.Errorf("redirectUris[%d] = %v, want %s", i, redirects[i], uri)
		}
	}

	// The CLI's tokens talk to the same control plane the console's do, so
	// the same API audience must be stamped.
	mappers := cli["protocolMappers"].([]any)
	if len(mappers) != 1 {
		t.Fatalf("expected the audience protocol mapper on the CLI client, got %d", len(mappers))
	}
	mapper := mappers[0].(map[string]any)
	if mapper["name"] != IdentityAudienceMapperName || mapper["protocolMapper"] != "oidc-audience-mapper" {
		t.Errorf("CLI client mapper = %v, want the shared audience mapper", mapper)
	}

	// offline_access is a DEFAULT scope here: the CLI always requests it so
	// sessions survive access-token expiry without re-prompting.
	defaultScopes := cli["defaultClientScopes"].([]any)
	foundOffline := false
	for _, s := range defaultScopes {
		if s == "offline_access" {
			foundOffline = true
		}
	}
	if !foundOffline {
		t.Error("offline_access missing from the CLI client's default scopes")
	}
}

func TestIdentityInternalServerRootURL(t *testing.T) {
	got := IdentityInternalServerRootURL("planton", "default")
	want := "http://planton-identity.default.svc.cluster.local/idp"
	if got != want {
		t.Errorf("server root = %s, want %s", got, want)
	}
}

// serviceAccountUser finds the user-directory client's service-account entry.
func serviceAccountUser(t *testing.T, realm map[string]any) map[string]any {
	t.Helper()
	for _, u := range realm["users"].([]any) {
		user := u.(map[string]any)
		if user["serviceAccountClientId"] == IdentityUsersClientID {
			return user
		}
	}
	t.Fatal("user-directory service-account user missing from the realm import")
	return nil
}

// humanUsers filters the realm's user entries down to people (non-service-account).
func humanUsers(realm map[string]any) []map[string]any {
	var humans []map[string]any
	for _, u := range realm["users"].([]any) {
		user := u.(map[string]any)
		if user["serviceAccountClientId"] == nil {
			humans = append(humans, user)
		}
	}
	return humans
}

func TestIdentityRealmImport_SeededAdminUser(t *testing.T) {
	realm := parseTestRealmImport(t)

	users := humanUsers(realm)
	if len(users) != 1 {
		t.Fatalf("expected the seeded admin user, got %d human users", len(users))
	}
	user := users[0]
	// The LOGIN is the fixed short "admin" (what gets typed at the sign-in
	// form); the EMAIL is the declared real person -- the key every
	// downstream grant (account creation, org ownership, platform operator)
	// matches on. Swapping either side breaks a different half of the flow.
	if user["username"] != IdentityAdminUsername {
		t.Errorf("admin username = %v, want %q", user["username"], IdentityAdminUsername)
	}
	if user["email"] != "admin@example.com" {
		t.Errorf("admin email = %v, want the declared adminEmail", user["email"])
	}
	// Imported users get exactly the listed roles (no default-role composite
	// application); without offline_access the token exchange fails outright.
	roles := user["realmRoles"].([]any)
	foundOfflineRole := false
	for _, r := range roles {
		if r == "offline_access" {
			foundOfflineRole = true
		}
	}
	if !foundOfflineRole {
		t.Error("seeded admin must carry the offline_access realm role explicitly")
	}

	// No name is baked into the seeded user: UPDATE_PROFILE asks the real
	// person for theirs at first sign-in, on the same screen as the forced
	// password change (the credential's temporary flag alone does not survive
	// realm import; the required actions are what actually run).
	actions := user["requiredActions"].([]any)
	if len(actions) != 2 || actions[0] != "UPDATE_PASSWORD" || actions[1] != "UPDATE_PROFILE" {
		t.Errorf("requiredActions = %v, want [UPDATE_PASSWORD UPDATE_PROFILE]", actions)
	}
	if _, hasName := user["firstName"]; hasName {
		t.Error("seeded admin must not carry a baked-in name; UPDATE_PROFILE collects the real one")
	}

	creds := user["credentials"].([]any)
	cred := creds[0].(map[string]any)
	// The generated password must never remain a long-lived credential.
	if cred["temporary"] != true {
		t.Error("the seeded admin password must be temporary (forced change at first sign-in)")
	}
	if cred["value"] != "one-time-password" {
		t.Errorf("credential value = %v, want the generated password", cred["value"])
	}
}

// Without a declared admin email NO PERSON is seeded: there is deliberately no
// generic pre-baked identity -- the first-run setup flow creates the first
// admin. The user-directory service account (an OAuth client identity, not a
// person) is present in both shapes; setup itself depends on it.
func TestIdentityRealmImport_NoAdminMeansNoHumanUsers(t *testing.T) {
	data, err := IdentityRealmImport(IdentityRealmImportConfig{
		Realm:        testIdentityRealm,
		PublicURL:    "https://planton.example.com",
		ClientSecret: "client-secret-value",
		UsersSecret:  "users-secret-value",
	})
	if err != nil {
		t.Fatal(err)
	}
	var realm map[string]any
	if err := json.Unmarshal(data, &realm); err != nil {
		t.Fatal(err)
	}
	if humans := humanUsers(realm); len(humans) != 0 {
		t.Errorf("expected no human users without adminEmail, got %d", len(humans))
	}
	// The setup flow's credential must exist even (especially) with no admin.
	serviceAccountUser(t, realm)
}

// The federation facts file is a CROSS-LANGUAGE contract: the control plane's
// Java reader parses exactly this serialization (its own test pins the same
// fixture). A field rename here without the reader's is a contract break --
// this test makes it loud on the Go side.
func TestIdentityFederationFactsConfigMap_ContractShape(t *testing.T) {
	facts := IdentityFederationFacts{
		Configured:    true,
		Arm:           "ldap",
		ProviderLabel: "Active Directory",
		Provisioned:   true,
		ObservedAt:    "2026-08-24T10:00:00Z",
		Checks: []IdentityFederationFactsCheck{
			{Name: "connection", Verdict: "Passed", Message: "the directory answered"},
			{Name: "usersSearch", Verdict: "Failed", Message: "the users base is not readable"},
		},
	}
	configMap, err := IdentityFederationFactsConfigMap("planton", "default", facts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if configMap.Name != "planton-identity-federation-facts" {
		t.Errorf("name = %q, want planton-identity-federation-facts", configMap.Name)
	}
	want := `{"configured":true,"arm":"ldap","providerLabel":"Active Directory",` +
		`"provisioned":true,"observedAt":"2026-08-24T10:00:00Z","checks":[` +
		`{"name":"connection","verdict":"Passed","message":"the directory answered"},` +
		`{"name":"usersSearch","verdict":"Failed","message":"the users base is not readable"}]}`
	if got := configMap.Data[IdentityFederationFactsKey]; got != want {
		t.Errorf("facts payload drifted from the pinned contract:\n got: %s\nwant: %s", got, want)
	}
}

// The unconfigured projection serializes to the minimal honest document --
// no stale arm, label, or verdicts ride along.
func TestIdentityFederationFactsConfigMap_UnconfiguredShape(t *testing.T) {
	configMap, err := IdentityFederationFactsConfigMap("planton", "default", IdentityFederationFacts{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := configMap.Data[IdentityFederationFactsKey]; got != `{"configured":false}` {
		t.Errorf("unconfigured payload = %s, want {\"configured\":false}", got)
	}
}
