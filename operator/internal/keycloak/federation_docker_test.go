//go:build requires_docker

package keycloak

// The federation half of the convergence suite: LDAP federation and OIDC
// brokering provisioned from desired state against a REAL Keycloak and the
// REAL lab directory (a Samba AD over private-CA LDAPS -- see
// hack/lab-directory). Every test here is a rehearsal of an adopter
// journey: connect, verify with honest verdicts, rotate a credential, delete
// the manifest, switch arms.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/plantonhq/planton/operator/internal/resources"
)

func federationInput(realm string, fed *OwnedFederation) ConvergeInput {
	in := convergeInput(realm)
	in.Federation = fed
	return in
}

func verifyInput(realm string, fed *OwnedFederation) VerifyInput {
	return VerifyInput{
		Realm:         realm,
		ServerRoot:    testServerRoot,
		AdminUsername: testBootstrapAdminUser,
		AdminPassword: testBootstrapAdminPass,
		HTTPClient:    &http.Client{Timeout: 30 * time.Second},
		Federation:    fed,
	}
}

func checkByName(t *testing.T, checks []Check, name string) Check {
	t.Helper()
	for _, check := range checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("no %q check among %+v", name, checks)
	return Check{}
}

// The whole LDAP journey: a manifest's federation provisions onto a fresh
// realm (component, group mapper, the /directory parent, the groups protocol
// mapper on both sign-in clients), verification passes with concrete counts
// from the REAL directory, the realm is idempotent (zero writes), and a
// credential rotation is exactly one deliberate write.
func TestFederation_LDAPProvisionVerifyAndIdempotency(t *testing.T) {
	admin := authedAdmin(t)
	createRealm(t, admin, "ldapfed", nil)
	fed := &OwnedFederation{LDAP: testLab.ldapFederation()}
	fed.LDAP.RotateCredential = true // first write carries the credential
	in := federationInput("ldapfed", fed)
	mustConverge(t, in)

	ctx := context.Background()

	// The provisioned shape.
	realmRep, err := admin.GetRealm(ctx, "ldapfed")
	if err != nil {
		t.Fatal(err)
	}
	realmID, _ := realmRep["id"].(string)
	components, err := admin.ListComponents(ctx, "ldapfed", realmID, userStorageProviderType)
	if err != nil {
		t.Fatal(err)
	}
	var ldapComponent Representation
	for _, c := range components {
		if c["name"] == resources.IdentityLDAPComponentName {
			ldapComponent = c
		}
	}
	if ldapComponent == nil {
		t.Fatal("LDAP federation component not provisioned")
	}
	if _, found, err := admin.GetGroupByPath(ctx, "ldapfed", resources.IdentityDirectoryGroupsPath); err != nil || !found {
		t.Fatalf("directory groups parent not provisioned (err=%v)", err)
	}
	for _, clientID := range []string{resources.IdentityConsoleClientID, resources.IdentityCLIClientID} {
		client, _, err := admin.FindClientByClientID(ctx, "ldapfed", clientID)
		if err != nil {
			t.Fatal(err)
		}
		clientUUID, _ := client["id"].(string)
		mappers, err := admin.ListProtocolMappers(ctx, "ldapfed", clientUUID)
		if err != nil {
			t.Fatal(err)
		}
		var groupsMapper, subjectMapper Representation
		for _, m := range mappers {
			switch m["name"] {
			case resources.IdentityGroupsMapperName:
				groupsMapper = m
			case resources.IdentityDirectorySubjectMapperName:
				subjectMapper = m
			}
		}
		if groupsMapper == nil || groupsMapper["protocolMapper"] != "oidc-group-membership-mapper" {
			t.Errorf("client %s groups mapper = %v, want oidc-group-membership-mapper", clientID, groupsMapper)
		}
		if subjectMapper == nil {
			t.Errorf("client %s must carry the directory-subject mapper", clientID)
		} else if cfg, _ := subjectMapper["config"].(map[string]any); cfg == nil || cfg["user.attribute"] != "LDAP_ID" {
			t.Errorf("client %s subject mapper must read LDAP_ID on the LDAP arm, got %v", clientID, subjectMapper["config"])
		}
	}

	// Verification against the real directory, with concrete counts. The
	// seed has 16 people under the users OU and 5 groups under the groups
	// OU; every check must pass -- including the disabled account and the
	// email-less fixtures importing cleanly.
	verifyIn := verifyInput("ldapfed", fed)
	verifyIn.SeededAdminEmail = "it.admin@lab.example.internal"
	checks, err := Verify(ctx, verifyIn)
	if err != nil {
		t.Fatalf("verification could not run: %v", err)
	}
	for _, name := range []string{"connection", "bind", "usersSearch", "groupsSearch"} {
		if check := checkByName(t, checks, name); check.Verdict != VerdictPassed {
			t.Errorf("%s = %s (%s), want Passed", name, check.Verdict, check.Message)
		}
	}
	if check := checkByName(t, checks, "usersSearch"); !strings.Contains(check.Message, "16") {
		t.Errorf("usersSearch must carry the real count (16 seeded staff): %q", check.Message)
	}
	// The directory's it.admin imported as a FEDERATED user; a federated
	// holder of the seeded email is exactly the healthy state.
	if check := checkByName(t, checks, "seededAdminCollision"); check.Verdict != VerdictPassed {
		t.Errorf("seededAdminCollision = %s (%s), want Passed for a federated holder", check.Verdict, check.Message)
	}

	// The premise the directory-subject mapper stands on, proven against the
	// REAL directory: Keycloak stamps LDAP_ID (the objectGUID) on every
	// synced federated user. Without this attribute the claim would be
	// silently empty and every directory account would lose its origin_ref.
	var syncedUsers []Representation
	if err := admin.do(ctx, http.MethodGet,
		admin.adminPath("ldapfed", "/users?exact=true&briefRepresentation=false&username=ada.lovelace"),
		nil, http.StatusOK, &syncedUsers); err != nil {
		t.Fatal(err)
	}
	if len(syncedUsers) != 1 {
		t.Fatalf("expected the synced ada.lovelace, got %d users", len(syncedUsers))
	}
	attrs, _ := syncedUsers[0]["attributes"].(map[string]any)
	if ldapID, _ := attrs["LDAP_ID"].([]any); len(ldapID) == 0 || ldapID[0] == "" {
		t.Errorf("a synced federated user must carry the LDAP_ID attribute (the objectGUID), got attributes %v", attrs)
	}

	// The person's NAMES arrive as the manifest's attributes say (givenName
	// -> firstName, sn -> lastName), on a directory whose cn is the username.
	// Keycloak's AD-vendor default set has no first-name mapper and a
	// full-name mapper splitting cn; the e2e lab caught a person imported
	// with no first name and her surname in its place. The operator now
	// owns the name mappers -- this is the outcome that must hold.
	if first, last := syncedUsers[0]["firstName"], syncedUsers[0]["lastName"]; first != "Ada" || last != "Lovelace" {
		t.Errorf("synced ada.lovelace names = %v / %v, want Ada / Lovelace (the manifest's givenName/sn mapping)", first, last)
	}
	ldapComponentID, _ := ldapComponent["id"].(string)
	childMappers, err := admin.ListComponents(ctx, "ldapfed", ldapComponentID, ldapStorageMapperType)
	if err != nil {
		t.Fatal(err)
	}
	mapperNames := map[string]Representation{}
	for _, m := range childMappers {
		name, _ := m["name"].(string)
		mapperNames[name] = m
	}
	for name, owned := range fed.LDAP.ownedAttributeMappers() {
		live, ok := mapperNames[name]
		if !ok {
			t.Errorf("owned attribute mapper %q must exist on the component", name)
			continue
		}
		cfg, _ := live["config"].(map[string]any)
		if got, _ := cfg["ldap.attribute"].([]any); len(got) != 1 || got[0] != owned.ldapAttribute {
			t.Errorf("mapper %q ldap.attribute = %v, want %s", name, got, owned.ldapAttribute)
		}
	}
	if _, stillThere := mapperNames[vendorFullNameMapperName]; stillThere {
		t.Errorf("the vendor %q mapper must be removed once first and last name are owned explicitly", vendorFullNameMapperName)
	}

	// Idempotency: a converged realm produces ZERO writes -- including the
	// masked bind credential, which is never diffed (write-on-rotation
	// only).
	fed.LDAP.RotateCredential = false
	second := mustConverge(t, in)
	if !second.Clean() {
		t.Fatalf("second pass must write nothing, wrote %d: %v", second.Writes, second.Repairs)
	}

	// Rotation: exactly one deliberate write carrying the credential.
	fed.LDAP.RotateCredential = true
	rotation := mustConverge(t, in)
	if rotation.Writes != 1 {
		t.Fatalf("rotation must be exactly one write, got %d: %v", rotation.Writes, rotation.Repairs)
	}
	if !strings.Contains(strings.Join(rotation.Repairs, " "), "rotated") {
		t.Errorf("rotation repair must say so: %v", rotation.Repairs)
	}
}

// A wrong bind password is a VERDICT naming the account and the remedy --
// never a stack trace. Connection still passes (the server answered), which
// is what makes the verdict list diagnostic instead of binary.
func TestFederation_WrongBindPasswordVerdict(t *testing.T) {
	admin := authedAdmin(t)
	createRealm(t, admin, "wrongbind", nil)

	fed := &OwnedFederation{LDAP: testLab.ldapFederation()}
	fed.LDAP.BindCredential = "not-the-password"

	checks, err := Verify(context.Background(), verifyInput("wrongbind", fed))
	if err != nil {
		t.Fatalf("verification could not run: %v", err)
	}
	if check := checkByName(t, checks, "connection"); check.Verdict != VerdictPassed {
		t.Errorf("connection = %s, want Passed (the server IS reachable)", check.Verdict)
	}
	bind := checkByName(t, checks, "bind")
	if bind.Verdict != VerdictFailed {
		t.Fatalf("bind = %s, want Failed", bind.Verdict)
	}
	if !strings.Contains(bind.Message, fed.LDAP.BindDN) || !strings.Contains(bind.Message, "Secret") {
		t.Errorf("the bind verdict must name the account and the remedy: %q", bind.Message)
	}
	if strings.Contains(bind.Message, "Exception") || strings.Contains(bind.Message, "\tat ") {
		t.Errorf("a verdict must never carry a stack trace: %q", bind.Message)
	}
}

// Manifest deletion: the empty federation sweeps every operator-named object
// -- and ONLY those. An admin-created federation provider under another name
// survives byte-identical (the never-clobber contract extends to federation).
func TestFederation_SweepOnManifestDeletion(t *testing.T) {
	admin := authedAdmin(t)
	createRealm(t, admin, "sweepfed", nil)
	fed := &OwnedFederation{LDAP: testLab.ldapFederation()}
	fed.LDAP.RotateCredential = true
	mustConverge(t, federationInput("sweepfed", fed))

	ctx := context.Background()
	realmRep, err := admin.GetRealm(ctx, "sweepfed")
	if err != nil {
		t.Fatal(err)
	}
	realmID, _ := realmRep["id"].(string)

	// An ADMIN-created federation provider, not operator-named, no mark.
	if err := admin.CreateComponent(ctx, "sweepfed", Representation{
		"name":         "corp-own-ldap",
		"providerId":   "ldap",
		"providerType": userStorageProviderType,
		"parentId":     realmID,
		"config": map[string][]string{
			"enabled":               {"false"},
			"vendor":                {"ad"},
			"connectionUrl":         {"ldaps://corp-own.example.com:636"},
			"bindDn":                {"CN=corp,DC=example,DC=com"},
			"bindCredential":        {"corp-secret"},
			"usersDn":               {"DC=example,DC=com"},
			"usernameLDAPAttribute": {"sAMAccountName"},
			"rdnLDAPAttribute":      {"cn"},
			"uuidLDAPAttribute":     {"objectGUID"},
			"userObjectClasses":     {"person"},
			"editMode":              {"READ_ONLY"},
			"authType":              {"simple"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// The manifest-deletion pass: none desired.
	report := mustConverge(t, federationInput("sweepfed", &OwnedFederation{}))
	if report.Clean() {
		t.Fatal("the sweep must report what it removed")
	}

	components, err := admin.ListComponents(ctx, "sweepfed", realmID, userStorageProviderType)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range components {
		if c["name"] == resources.IdentityLDAPComponentName {
			t.Error("the operator's federation component must be swept")
		}
	}
	survived := false
	for _, c := range components {
		if c["name"] == "corp-own-ldap" {
			survived = true
		}
	}
	if !survived {
		t.Error("an admin-created federation provider must SURVIVE the sweep")
	}

	if _, found, err := admin.GetGroupByPath(ctx, "sweepfed", resources.IdentityDirectoryGroupsPath); err != nil || found {
		t.Errorf("the directory groups parent must be swept (found=%v err=%v)", found, err)
	}
	for _, clientID := range []string{resources.IdentityConsoleClientID, resources.IdentityCLIClientID} {
		client, _, err := admin.FindClientByClientID(ctx, "sweepfed", clientID)
		if err != nil {
			t.Fatal(err)
		}
		clientUUID, _ := client["id"].(string)
		mappers, err := admin.ListProtocolMappers(ctx, "sweepfed", clientUUID)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range mappers {
			if m["name"] == resources.IdentityGroupsMapperName {
				t.Errorf("client %s groups mapper must be swept", clientID)
			}
			if m["name"] == resources.IdentityDirectorySubjectMapperName {
				t.Errorf("client %s directory-subject mapper must be swept", clientID)
			}
		}
	}

	// Deletion is idempotent too.
	if !mustConverge(t, federationInput("sweepfed", &OwnedFederation{})).Clean() {
		t.Error("a swept realm must be idempotent")
	}
}

// The brokered arm: operator-side discovery against a REAL issuer (the
// entra-sim realm on the suite's own Keycloak), explicit endpoints written
// into the broker, the groups importer with FORCE sync, idempotency -- and
// the arm SWITCH, where the groups protocol mapper is recreated with the
// other arm's type. Also pins the premise the masked-secret design stands
// on: Keycloak returns the client secret masked.
func TestFederation_BrokerProvisionAndArmSwitch(t *testing.T) {
	admin := authedAdmin(t)
	createRealm(t, admin, "entra-sim", nil)
	createRealm(t, admin, "brokered", nil)

	// ONE issuer string valid on both sides of the network boundary: the
	// test process resolves the suite alias through a dialer override; the
	// identity server resolves it through container DNS.
	issuer := "http://" + testNetworkAlias + ":8080" + resources.IdentityPathPrefix + "/realms/entra-sim"
	client := aliasResolvingClient()

	ctx := context.Background()
	endpoints, advertised, err := DiscoverOIDC(ctx, client, issuer)
	if err != nil {
		t.Fatalf("discovery against the entra-sim issuer: %v", err)
	}
	if advertised != issuer {
		t.Fatalf("advertised issuer %q != configured %q", advertised, issuer)
	}

	fed := &OwnedFederation{Broker: &OwnedOIDCBroker{
		IssuerURL:        issuer,
		ClientID:         "planton-app-registration",
		ClientSecret:     "upstream-client-secret",
		RotateCredential: true,
		Scopes:           []string{"openid", "profile", "email"},
		GroupsClaim:      "groups",
		SubjectClaim:     "sub",
		DisplayName:      "Sign in with your organization",
		Endpoints:        endpoints,
	}}
	in := federationInput("brokered", fed)
	mustConverge(t, in)

	instances, err := admin.ListIdentityProviders(ctx, "brokered")
	if err != nil {
		t.Fatal(err)
	}
	var broker Representation
	for _, inst := range instances {
		if inst["alias"] == resources.IdentityBrokerAlias {
			broker = inst
		}
	}
	if broker == nil {
		t.Fatal("broker instance not provisioned")
	}
	config, _ := broker["config"].(map[string]any)
	if config["authorizationUrl"] != endpoints.AuthorizationURL {
		t.Errorf("authorizationUrl = %v, want the discovered endpoint", config["authorizationUrl"])
	}
	// The masked-read premise: the secret can never be diffed.
	if secret, _ := config["clientSecret"].(string); !strings.Contains(secret, "**") {
		t.Errorf("expected the client secret masked on read, got %q -- the rotation design's premise", secret)
	}
	mappers, err := admin.ListIdentityProviderMappers(ctx, "brokered", resources.IdentityBrokerAlias)
	if err != nil {
		t.Fatal(err)
	}
	// Both owned claim importers, each FORCE: groups (import-once freezes
	// memberships at first login) and the directory subject (a stale id
	// breaks the account correlation offboarding runs on).
	importers := map[string]string{
		brokerGroupsImporterName:  resources.IdentityDirectoryGroupsAttribute,
		brokerSubjectImporterName: resources.IdentityDirectorySubjectAttribute,
	}
	for _, m := range mappers {
		name, _ := m["name"].(string)
		wantAttr, owned := importers[name]
		if !owned {
			continue
		}
		cfg, _ := m["config"].(map[string]any)
		if cfg == nil || cfg["syncMode"] != "FORCE" || cfg["user.attribute"] != wantAttr {
			t.Errorf("importer %s config = %v, want FORCE sync into %s", name, cfg, wantAttr)
		}
		delete(importers, name)
	}
	for name := range importers {
		t.Errorf("owned broker importer %s must exist", name)
	}

	// Idempotency at the new arm (endpoints still supplied, as a
	// verification pass would).
	fed.Broker.RotateCredential = false
	if second := mustConverge(t, in); !second.Clean() {
		t.Fatalf("second broker pass must write nothing, wrote %d: %v", second.Writes, second.Repairs)
	}
	// Steady state passes carry NO endpoints (owned-when-discovered); the
	// live endpoint config must not read as drift.
	fed.Broker.Endpoints = nil
	if steady := mustConverge(t, in); !steady.Clean() {
		t.Fatalf("endpoint-less steady pass must write nothing, wrote %d: %v", steady.Writes, steady.Repairs)
	}

	// The arm switch: the manifest now declares LDAP. The broker dies with
	// its mappers, the LDAP component appears, and the groups protocol
	// mapper is RECREATED with the LDAP arm's type (a mapper's type cannot
	// change in place).
	ldapFed := &OwnedFederation{LDAP: testLab.ldapFederation()}
	ldapFed.LDAP.RotateCredential = true
	mustConverge(t, federationInput("brokered", ldapFed))

	instances, err = admin.ListIdentityProviders(ctx, "brokered")
	if err != nil {
		t.Fatal(err)
	}
	for _, inst := range instances {
		if inst["alias"] == resources.IdentityBrokerAlias {
			t.Error("the broker must be swept on arm switch")
		}
	}
	console, _, err := admin.FindClientByClientID(ctx, "brokered", resources.IdentityConsoleClientID)
	if err != nil {
		t.Fatal(err)
	}
	consoleUUID, _ := console["id"].(string)
	protocolMappers, err := admin.ListProtocolMappers(ctx, "brokered", consoleUUID)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range protocolMappers {
		if m["name"] == resources.IdentityGroupsMapperName && m["protocolMapper"] != "oidc-group-membership-mapper" {
			t.Errorf("groups mapper after arm switch = %v, want oidc-group-membership-mapper", m["protocolMapper"])
		}
		if m["name"] == resources.IdentityDirectorySubjectMapperName {
			if cfg, _ := m["config"].(map[string]any); cfg == nil || cfg["user.attribute"] != "LDAP_ID" {
				t.Errorf("subject mapper after arm switch must read LDAP_ID, got %v", m["config"])
			}
		}
	}
}

// The seeded-admin collision: a LOCAL user holding the declared admin email
// blocks the directory twin from ever materializing (realm emails are
// unique). The verdict names the collision AND the remedy; the user sync
// honestly reports its import failure the same pass.
func TestFederation_SeededAdminCollisionVerdict(t *testing.T) {
	admin := authedAdmin(t)
	createRealm(t, admin, "collision", nil)

	// The seeded local admin, exactly as the declared-adminEmail import
	// bakes it: local (no federation link), holding the email.
	if err := admin.do(context.Background(), http.MethodPost,
		admin.serverRoot+"/admin/realms/collision/users",
		Representation{"username": "admin", "email": "it.admin@lab.example.internal", "enabled": true},
		http.StatusCreated, nil); err != nil {
		t.Fatal(err)
	}

	fed := &OwnedFederation{LDAP: testLab.ldapFederation()}
	fed.LDAP.RotateCredential = true
	mustConverge(t, federationInput("collision", fed))

	verifyIn := verifyInput("collision", fed)
	verifyIn.SeededAdminEmail = "it.admin@lab.example.internal"
	checks, err := Verify(context.Background(), verifyIn)
	if err != nil {
		t.Fatalf("verification could not run: %v", err)
	}

	collision := checkByName(t, checks, "seededAdminCollision")
	if collision.Verdict != VerdictFailed {
		t.Fatalf("seededAdminCollision = %s (%s), want Failed with a local holder", collision.Verdict, collision.Message)
	}
	if !strings.Contains(collision.Message, "bootstrap.admins") {
		t.Errorf("the collision verdict must name the remedy: %q", collision.Message)
	}
	// The sync sees the same reality: the directory's it.admin cannot
	// import past the local twin.
	if usersCheck := checkByName(t, checks, "usersSearch"); usersCheck.Verdict != VerdictFailed {
		t.Errorf("usersSearch = %s (%s), want Failed while the collision blocks an import", usersCheck.Verdict, usersCheck.Message)
	}
}

// The other half of the collision law, against the real directory: a LOCAL
// user holding an email NO directory person shares -- the shape of nearly
// every install, whose seeded admin is rarely a lab fixture -- must NOT fail
// the manifest. The user sync imports cleanly, so the verdict is the
// advisory (Unknown) that still names the remedy. The e2e lab caught the old
// Failed-on-existence shape live on the published line.
func TestFederation_SeededAdminWithoutDirectoryTwinIsAdvisory(t *testing.T) {
	admin := authedAdmin(t)
	createRealm(t, admin, "no-twin", nil)

	if err := admin.do(context.Background(), http.MethodPost,
		admin.serverRoot+"/admin/realms/no-twin/users",
		Representation{"username": "admin", "email": "nobody-in-the-directory@planton.local", "enabled": true},
		http.StatusCreated, nil); err != nil {
		t.Fatal(err)
	}

	fed := &OwnedFederation{LDAP: testLab.ldapFederation()}
	fed.LDAP.RotateCredential = true
	mustConverge(t, federationInput("no-twin", fed))

	verifyIn := verifyInput("no-twin", fed)
	verifyIn.SeededAdminEmail = "nobody-in-the-directory@planton.local"
	checks, err := Verify(context.Background(), verifyIn)
	if err != nil {
		t.Fatalf("verification could not run: %v", err)
	}

	if usersCheck := checkByName(t, checks, "usersSearch"); usersCheck.Verdict != VerdictPassed {
		t.Fatalf("usersSearch = %s (%s), want Passed -- nothing in the directory collides", usersCheck.Verdict, usersCheck.Message)
	}
	collision := checkByName(t, checks, "seededAdminCollision")
	if collision.Verdict != VerdictUnknown {
		t.Fatalf("seededAdminCollision = %s (%s), want the advisory Unknown for a local holder with a clean import", collision.Verdict, collision.Message)
	}
	if !strings.Contains(collision.Message, "bootstrap.admins") {
		t.Errorf("the advisory must still name the remedy: %q", collision.Message)
	}
}

// The group-read seam's premise, proven with the EXACT credential the
// control plane holds (the planton-users service account -- never the
// bootstrap admin, which the product must never touch): the mirrored realm
// groups under /directory are readable through the admin API, and each
// carries the directory group's objectGUID attribute as a canonical dashed
// UUID -- the stable id the group->role mappings key on, and the value the
// sign-in join resolves a token's group paths against. Cross-checked against
// the directory's OWN objectGUID via ldbsearch, so the attribute's encoding
// can never drift silently: if Keycloak ever stored the raw binary GUID
// here, this test lands red before any mapping could silently never match.
func TestFederation_PlantonUsersCredentialReadsMirroredGroupGUIDs(t *testing.T) {
	admin := authedAdmin(t)
	createRealm(t, admin, "groupread", nil)
	fed := &OwnedFederation{LDAP: testLab.ldapFederation()}
	fed.LDAP.RotateCredential = true
	mustConverge(t, federationInput("groupread", fed))

	// Verification's group sync populates the /directory mirror -- the same
	// pass that populates it on a live install.
	ctx := context.Background()
	if _, err := Verify(ctx, verifyInput("groupread", fed)); err != nil {
		t.Fatalf("verification could not run: %v", err)
	}

	// The control plane's own credential and grant: client_credentials as
	// planton-users on the platform realm's token endpoint.
	users := NewAdminClient(&http.Client{Timeout: 10 * time.Second}, testServerRoot)
	users.token = clientCredentialsToken(t, "groupread", resources.IdentityUsersClientID, "users-secret-value")

	parent, found, err := users.GetGroupByPath(ctx, "groupread", resources.IdentityDirectoryGroupsPath)
	if err != nil || !found {
		t.Fatalf("planton-users must be able to read the %s parent group (found=%v err=%v)",
			resources.IdentityDirectoryGroupsPath, found, err)
	}
	parentID, _ := parent["id"].(string)

	// The exact read the control plane's group index performs: the parent's
	// children with full representations (briefRepresentation=false is what
	// carries the attributes).
	var children []Representation
	if err := users.do(ctx, http.MethodGet,
		users.adminPath("groupread", "/groups/"+parentID+"/children?briefRepresentation=false&max=100"),
		nil, http.StatusOK, &children); err != nil {
		t.Fatalf("planton-users must be able to list the mirror's children with attributes: %v", err)
	}
	if len(children) == 0 {
		t.Fatal("no mirrored groups under the directory parent after the group sync")
	}

	// Keycloak stores the mirrored attribute as it read it over LDAP:
	// base64 of AD's 16 raw GUID bytes (suite-discovered; the LDAP_ID
	// decode path is separate code). The control plane's group index
	// normalizes to the canonical dashed display form -- this test performs
	// the SAME normalization and cross-checks it against the directory's
	// own answer, so the encoding contract can never drift silently.
	guidsByName := map[string]string{}
	for _, child := range children {
		name, _ := child["name"].(string)
		attrs, _ := child["attributes"].(map[string]any)
		values, _ := attrs["objectGUID"].([]any)
		if len(values) == 0 {
			t.Fatalf("mirrored group %q carries no objectGUID attribute (attributes = %v) -- the stable-id premise of the mapping join", name, attrs)
		}
		raw, _ := values[0].(string)
		normalized := normalizedObjectGUID(t, name, raw)
		guidsByName[name] = normalized
	}

	// Two groups cross-checked against the directory's own records: the
	// value an admin sees in their AD tooling IS the value the index
	// serves, byte for byte (case-insensitively -- hex case is display).
	for _, groupName := range []string{"platform-eng", "all-engineering"} {
		directoryGUID := testLab.groupObjectGUID(t, groupName)
		if mirrored := guidsByName[groupName]; !strings.EqualFold(mirrored, directoryGUID) {
			t.Errorf("normalized %s objectGUID = %q, the directory says %q", groupName, mirrored, directoryGUID)
		}
	}
}

// normalizedObjectGUID mirrors the control plane's normalization (the
// RealmGroupsClient seam): an already-dashed value passes through; a base64
// value decodes to AD's 16 raw bytes and renders in the Microsoft display
// form -- the GUID structure's first three fields are little-endian on the
// wire -- which is the form AD tooling shows, Entra uses, and Keycloak's own
// LDAP_ID decode produces.
func normalizedObjectGUID(t *testing.T, name, value string) string {
	t.Helper()
	dashed := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if dashed.MatchString(value) {
		return value
	}
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(raw) != 16 {
		t.Fatalf("mirrored group %q objectGUID = %q is neither a dashed UUID nor base64 of 16 bytes (err=%v) -- the mapping join cannot key on this", name, value, err)
	}
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		raw[3], raw[2], raw[1], raw[0], raw[5], raw[4], raw[7], raw[6],
		raw[8], raw[9], raw[10], raw[11], raw[12], raw[13], raw[14], raw[15])
}

// clientCredentialsToken mints a service-account token on the PLATFORM
// realm's own token endpoint -- the grant the control plane's realm clients
// use, as opposed to the bootstrap admin's master-realm password grant.
func clientCredentialsToken(t *testing.T, realm, clientID, secret string) string {
	t.Helper()
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", secret)
	resp, err := http.PostForm(testServerRoot+"/realms/"+realm+"/protocol/openid-connect/token", form)
	if err != nil {
		t.Fatalf("requesting %s token: %v", clientID, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s token grant returned %d", clientID, resp.StatusCode)
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil || tokenResp.AccessToken == "" {
		t.Fatalf("decoding %s token response: %v", clientID, err)
	}
	return tokenResp.AccessToken
}

// aliasResolvingClient resolves the suite's network alias to the published
// host port -- the seam that lets the test process fetch the same issuer URL
// the identity server resolves through container DNS.
func aliasResolvingClient() *http.Client {
	hostPort := strings.TrimSuffix(strings.TrimPrefix(testServerRoot, "http://"), resources.IdentityPathPrefix)
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if addr == testNetworkAlias+":8080" {
					addr = hostPort
				}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}
}
