package keycloak

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/plantonhq/planton/operator/internal/resources"
)

// This file is the federation half of the owned set: the LDAP user-federation
// component, the identity-broker instance, their mappers, and the
// arm-specific groups protocol mapper on the sign-in clients -- everything a
// bound PlantonIdentityProvider declares.
//
// Ownership here is BY CONSTANT NAME (the component name, the broker alias,
// the mapper names), not by an attribute mark: federation objects' names are
// operator vocabulary the same way the three client ids are, and a name-keyed
// enumeration keeps the sweep deterministic on realms regardless of what
// Keycloak tolerates in config maps. An admin-created federation component
// under a different name is structurally invisible; one created under OUR
// exact name is the same case as a hand-edited owned client -- a repair, not
// an admin choice. The managed config key is still stamped on created objects
// for admin-console visibility.
//
// Secret config values (the bind credential, the broker client secret) are
// MASKED by Keycloak on read (**********), so they can never be diffed here.
// The caller (the identity component) keeps a fingerprint of the last-written
// value in an operator-owned Kubernetes Secret and passes RotateCredential
// when the referenced Secret's content moved. The fingerprint deliberately
// never enters Keycloak config: any deterministic digest of a password,
// readable by realm admins, is offline-crack material.

// OwnedLDAPFederation is the desired LDAP user-federation state, translated
// from the manifest's activeDirectory arm by the identity component.
type OwnedLDAPFederation struct {
	Servers  []string
	StartTLS bool

	// UseTruststore is set when a CA bundle is mounted (private-CA LDAPS);
	// it pins useTruststoreSpi=always so the bundle is consulted.
	UseTruststore bool

	BindDN string

	// BindCredential is written at create and on RotateCredential only --
	// see the package comment on masked reads.
	BindCredential   string
	RotateCredential bool

	UsersDN            string
	GroupsDN           string
	UserObjectClasses  []string
	UsernameAttribute  string
	EmailAttribute     string
	FirstNameAttribute string
	LastNameAttribute  string
	GroupNameAttribute string
	GroupMemberAttr    string
	NestedGroups       bool
	SyncPeriodMinutes  int32
}

// OIDCEndpoints are the upstream provider's endpoints as the OPERATOR
// discovered them -- written explicitly into the broker config, mirroring the
// explicit-endpoint posture the console's own sign-in stack uses, so the
// broker never depends on a live re-discovery and the config stays diffable.
type OIDCEndpoints struct {
	AuthorizationURL string
	TokenURL         string
	JWKSURL          string
	UserInfoURL      string
	LogoutURL        string
}

// OwnedOIDCBroker is the desired identity-broker state, translated from the
// manifest's oidc arm.
type OwnedOIDCBroker struct {
	IssuerURL string
	ClientID  string

	// ClientSecret is written at create and on RotateCredential only.
	ClientSecret     string
	RotateCredential bool

	Scopes      []string
	GroupsClaim string
	// SubjectClaim is the upstream claim carrying the directory's stable id
	// for the user (defaulted to "sub" by the component translation --
	// present on every OIDC provider, so "brokered means directory-born"
	// holds structurally; Entra deployments set oid, which is tenant-stable
	// where Entra's sub is pairwise per app registration).
	SubjectClaim string
	DisplayName  string

	// Primary makes the broker the realm's default sign-in: the browser
	// flow's Identity Provider Redirector carries an operator-owned config
	// naming this broker, so the identity server's own form is skipped
	// unless a sign-in arrives with a kc_idp_hint the redirector cannot
	// route (the break-glass, DD-023). Off removes that config.
	Primary bool

	// Endpoints are non-nil only when the caller ran discovery this pass
	// (verification cadence); nil leaves the live endpoint fields untouched,
	// which is what keeps the steady-state reconcile from fetching the
	// upstream's discovery document every 30 seconds.
	Endpoints *OIDCEndpoints
}

// OwnedFederation is what the bound manifest declares -- exactly one arm.
// The nil-vs-empty distinction is load-bearing SAFETY semantics:
//
//   - nil: leave federation untouched this pass. The caller could not build
//     desired state (a referenced Secret unreadable, discovery unreachable)
//     -- and a transient read failure must NEVER tear down the live
//     federation a company signs in through.
//   - empty (both arms nil): none is desired; the sweep removes every
//     operator-named federation object. This is the manifest-deletion path,
//     converging without a finalizer.
type OwnedFederation struct {
	LDAP   *OwnedLDAPFederation
	Broker *OwnedOIDCBroker
}

const (
	userStorageProviderType = "org.keycloak.storage.UserStorageProvider"
	ldapStorageMapperType   = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"

	// ldapGroupMapperName names the operator's group-sync mapper under the
	// LDAP component. Package-local: unlike the client ids and the protocol
	// mapper name, nothing outside the reconciler references it.
	ldapGroupMapperName = "planton-directory-group-sync"

	// directoryGroupsParentName is IdentityDirectoryGroupsPath's group name.
	directoryGroupsParentName = "directory"
)

// componentConfig renders the owned LDAP component config keys. Every key
// listed here is converged; keys Keycloak or an admin adds beyond them ride
// through untouched (the same owned-KEYS discipline as client attributes).
func (f *OwnedLDAPFederation) componentConfig() map[string][]string {
	useTruststore := "ldapsOnly"
	if f.UseTruststore {
		useTruststore = "always"
	}
	return map[string][]string{
		"enabled": {"true"},
		// Active Directory vendor mode: sAMAccountName semantics, objectGUID
		// as the immutable id, userAccountControl awareness. Generic-LDAP
		// directories are still expressible through the attribute overrides.
		"vendor":                {"ad"},
		"connectionUrl":         {strings.Join(f.Servers, " ")},
		"startTls":              {strconv.FormatBool(f.StartTLS)},
		"useTruststoreSpi":      {useTruststore},
		"authType":              {"simple"},
		"bindDn":                {f.BindDN},
		"usersDn":               {f.UsersDN},
		"usernameLDAPAttribute": {f.UsernameAttribute},
		"rdnLDAPAttribute":      {"cn"},
		"uuidLDAPAttribute":     {"objectGUID"},
		"userObjectClasses":     {strings.Join(f.UserObjectClasses, ", ")},
		// The directory is the truth; Planton never writes it (the manifest
		// pins editMode READ_ONLY -- re-pinned here so a hand-flip converges
		// back).
		"editMode":          {"READ_ONLY"},
		"importEnabled":     {"true"},
		"syncRegistrations": {"false"},
		// Directory emails are IT-managed: trusting them skips a verify-email
		// interstitial that would ask a corporate user to prove an address
		// their own IT department assigned.
		"trustEmail": {"true"},
		"pagination": {"true"},
		// Bounded probes: a firewalled directory should fail a verification
		// check in seconds, not hang a reconcile pass.
		"connectionTimeout": {"10000"},
		"readTimeout":       {"10000"},
		// The manifest's one sync window drives Keycloak's periodic FULL
		// sync; changed-users sync stays off (AD's usnChanged semantics vary
		// across forests -- the full sync at the stated window is the
		// honest, predictable contract).
		"fullSyncPeriod":    {strconv.Itoa(int(f.SyncPeriodMinutes) * 60)},
		"changedSyncPeriod": {"-1"},
		// Deliberately NO managed-mark config key here: Keycloak's LDAP
		// provider STRIPS unknown config keys on write (suite-proven -- the
		// mark re-read as drift every pass), and ownership is by the
		// component's constant name anyway. The broker instance keeps its
		// mark; its config store tolerates arbitrary keys.
	}
}

// groupMapperConfig renders the owned group-sync mapper config.
func (f *OwnedLDAPFederation) groupMapperConfig() map[string][]string {
	retrieveStrategy := "LOAD_GROUPS_BY_MEMBER_ATTRIBUTE"
	if f.NestedGroups {
		retrieveStrategy = "LOAD_GROUPS_BY_MEMBER_ATTRIBUTE_RECURSIVELY"
	}
	return map[string][]string{
		"groups.dn":                 {f.GroupsDN},
		"group.name.ldap.attribute": {f.GroupNameAttribute},
		"group.object.classes":      {"group"},
		// Directory groups mirror under ONE dedicated realm-group subtree:
		// drop-non-existing sync semantics stay scoped to directory state
		// and structurally away from admin-created realm groups.
		"groups.path": {resources.IdentityDirectoryGroupsPath},
		// The mirror is FLAT, never a nested tree: a directory group with
		// several parents (utterly normal in real ADs -- the lab's
		// platform-eng sits in two) fails a nested sync outright with
		// GroupsMultipleParents (suite-proven). Nested MEMBERSHIP still
		// resolves through the recursive retrieve strategy below; the
		// platform maps groups by stable id, so the mirror's shape carries
		// no meaning anyway.
		"preserve.group.inheritance":           {"false"},
		"ignore.missing.groups":                {"false"},
		"membership.ldap.attribute":            {f.GroupMemberAttr},
		"membership.attribute.type":            {"DN"},
		"membership.user.ldap.attribute":       {f.UsernameAttribute},
		"memberof.ldap.attribute":              {"memberOf"},
		"user.roles.retrieve.strategy":         {retrieveStrategy},
		"drop.non.existing.groups.during.sync": {"true"},
		// The directory group's objectGUID rides onto each mirrored realm
		// group as an attribute -- the STABLE id the platform's group->role
		// mappings key on. Without this, group identity on the LDAP arm
		// would be display names, which rename and collide.
		"mapped.group.attributes": {"objectGUID"},
		"mode":                    {"READ_ONLY"},
	}
}

// ownedAttributeMappers is the operator-owned set of user-attribute mappers
// on the LDAP component, keyed by Keycloak's well-known mapper names:
// mapper name -> (user model attribute, the manifest's directory attribute).
// Keycloak auto-creates SOME of these when the component is created, but
// the set differs by vendor -- the Active Directory path creates "full name"
// (splitting `cn`) and "last name", and NO "first name" -- so a mapper that
// is missing is CREATED, never skipped: the manifest's firstNameAttribute
// was a silent no-op on every AD directory until the lab showed a person
// imported with no first name and her surname where her first name goes.
func (f *OwnedLDAPFederation) ownedAttributeMappers() map[string]ownedAttributeMapper {
	return map[string]ownedAttributeMapper{
		"username":   {userModelAttribute: "username", ldapAttribute: f.UsernameAttribute},
		"email":      {userModelAttribute: "email", ldapAttribute: f.EmailAttribute},
		"first name": {userModelAttribute: "firstName", ldapAttribute: f.FirstNameAttribute},
		"last name":  {userModelAttribute: "lastName", ldapAttribute: f.LastNameAttribute},
	}
}

type ownedAttributeMapper struct {
	userModelAttribute string
	ldapAttribute      string
}

// vendorFullNameMapperName is Keycloak's AD-vendor default that derives BOTH
// names from one attribute (`cn`). It fights the explicit first/last mappers
// -- on import the two write the same fields in unspecified order -- so the
// operator removes it once it owns the name mappers.
const vendorFullNameMapperName = "full name"

// ldapAttributeMapperConfig renders a user-attribute mapper the operator
// creates: read-only (the directory is the source of truth), always read
// from the directory, never mandatory (a service account may carry no given
// name; an import must not fail on it -- the email-less posture is decided
// by the product, not by the mapper).
func ldapAttributeMapperConfig(m ownedAttributeMapper) map[string][]string {
	return map[string][]string{
		"user.model.attribute":        {m.userModelAttribute},
		"ldap.attribute":              {m.ldapAttribute},
		"read.only":                   {"true"},
		"always.read.value.from.ldap": {"true"},
		"is.mandatory.in.ldap":        {"false"},
		"is.binary.attribute":         {"false"},
	}
}

// brokerConfig renders the owned broker config keys (identity-provider
// config values are plain strings, unlike component config's string lists).
func (b *OwnedOIDCBroker) brokerConfig() map[string]string {
	cfg := map[string]string{
		"issuer":   b.IssuerURL,
		"clientId": b.ClientID,
		// Standard confidential-client authentication at the upstream token
		// endpoint.
		"clientAuthMethod": "client_secret_post",
		"defaultScope":     strings.Join(b.Scopes, " "),
		// Upstream tokens are verified against the upstream's JWKS; never
		// unsigned trust.
		"validateSignature": "true",
		"useJwksUrl":        "true",
		// FORCE re-imports the mapped claims at EVERY sign-in. The default
		// import-once mode would freeze a user's groups at first login --
		// silently breaking offboarding on the brokered arm.
		"syncMode": "FORCE",
		// Visibility in the admin console; ownership itself is by alias.
		resources.IdentityManagedAttribute: "true",
	}
	if b.Endpoints != nil {
		cfg["authorizationUrl"] = b.Endpoints.AuthorizationURL
		cfg["tokenUrl"] = b.Endpoints.TokenURL
		cfg["jwksUrl"] = b.Endpoints.JWKSURL
		// userinfo and end-session are spec-optional; absent stays absent.
		if b.Endpoints.UserInfoURL != "" {
			cfg["userInfoUrl"] = b.Endpoints.UserInfoURL
		}
		if b.Endpoints.LogoutURL != "" {
			cfg["logoutUrl"] = b.Endpoints.LogoutURL
		}
	}
	return cfg
}

// groupsImporterConfig renders the broker mapper that lands the upstream
// groups claim in the directory-groups user attribute at every sign-in.
func (b *OwnedOIDCBroker) groupsImporterConfig() map[string]string {
	return map[string]string{
		"claim":          b.GroupsClaim,
		"user.attribute": resources.IdentityDirectoryGroupsAttribute,
		"syncMode":       "FORCE",
	}
}

// subjectImporterConfig renders the broker mapper that lands the upstream's
// stable-subject claim in the directory-subject user attribute at every
// sign-in (FORCE, like groups: an upstream re-keying must never leave a
// stale directory id behind).
func (b *OwnedOIDCBroker) subjectImporterConfig() map[string]string {
	return map[string]string{
		"claim":          b.SubjectClaim,
		"user.attribute": resources.IdentityDirectorySubjectAttribute,
		"syncMode":       "FORCE",
	}
}

// groupsProtocolMapper renders the arm-appropriate protocol mapper emitting
// the canonical groups claim on the sign-in clients. The LDAP arm's groups
// live as realm groups (group-membership mapper); the brokered arm's live in
// a user attribute (user-attribute mapper). One claim name, two sources.
func groupsProtocolMapper(fed *OwnedFederation) *OwnedMapper {
	switch {
	case fed.LDAP != nil:
		return &OwnedMapper{
			Name:           resources.IdentityGroupsMapperName,
			ProtocolMapper: "oidc-group-membership-mapper",
			Config: map[string]string{
				"claim.name": resources.IdentityGroupsClaim,
				// Full paths (/directory/...) so the claim self-identifies
				// as directory-mirrored and nested chains stay legible.
				"full.path":            "true",
				"access.token.claim":   "true",
				"id.token.claim":       "true",
				"userinfo.token.claim": "true",
			},
		}
	case fed.Broker != nil:
		return &OwnedMapper{
			Name:           resources.IdentityGroupsMapperName,
			ProtocolMapper: "oidc-usermodel-attribute-mapper",
			Config: map[string]string{
				"user.attribute":       resources.IdentityDirectoryGroupsAttribute,
				"claim.name":           resources.IdentityGroupsClaim,
				"jsonType.label":       "String",
				"multivalued":          "true",
				"access.token.claim":   "true",
				"id.token.claim":       "true",
				"userinfo.token.claim": "true",
			},
		}
	default:
		return nil
	}
}

// subjectProtocolMapper renders the arm-appropriate protocol mapper emitting
// the canonical directory-subject claim on the sign-in clients. Both arms
// read a user attribute; only WHICH attribute differs: Keycloak itself stamps
// LDAP_ID (the value of the component's uuidLDAPAttribute -- objectGUID here)
// on every LDAP-federated user, while brokered users carry the attribute the
// subject importer filled. One claim name, two sources -- the groups-mapper
// discipline exactly.
func subjectProtocolMapper(fed *OwnedFederation) *OwnedMapper {
	var userAttribute string
	switch {
	case fed.LDAP != nil:
		userAttribute = "LDAP_ID"
	case fed.Broker != nil:
		userAttribute = resources.IdentityDirectorySubjectAttribute
	default:
		return nil
	}
	return &OwnedMapper{
		Name:           resources.IdentityDirectorySubjectMapperName,
		ProtocolMapper: "oidc-usermodel-attribute-mapper",
		Config: map[string]string{
			"user.attribute":       userAttribute,
			"claim.name":           resources.IdentityDirectorySubjectClaim,
			"jsonType.label":       "String",
			"multivalued":          "false",
			"access.token.claim":   "true",
			"id.token.claim":       "true",
			"userinfo.token.claim": "true",
		},
	}
}

// convergeFederation drives the realm's federation state to what the bound
// manifest declares -- and to NOTHING when none is desired, which is the
// manifest-deletion path. A nil fed is the hands-off case (see
// OwnedFederation's nil-vs-empty contract). Order: LDAP component (+ mappers
// + the directory groups parent), broker instance (+ mappers), the primary
// sign-in redirector config, then the groups + directory-subject protocol
// mappers on the sign-in clients.
func convergeFederation(ctx context.Context, admin *AdminClient, realm string, realmID string, fed *OwnedFederation, report *Report) error {
	if fed == nil {
		return nil
	}
	if err := convergeLDAPComponent(ctx, admin, realm, realmID, fed.LDAP, report); err != nil {
		return err
	}
	if err := convergeBroker(ctx, admin, realm, fed.Broker, report); err != nil {
		return err
	}
	// After the broker: the redirector's default provider must name an
	// instance that exists (DD-023); with no broker desired, the operator's
	// config on the redirector goes with it.
	if err := convergePrimaryBroker(ctx, admin, realm, fed.Broker, report); err != nil {
		return err
	}
	if err := convergeSignInClientMapper(ctx, admin, realm, resources.IdentityGroupsMapperName, groupsProtocolMapper(fed), report); err != nil {
		return err
	}
	return convergeSignInClientMapper(ctx, admin, realm, resources.IdentityDirectorySubjectMapperName, subjectProtocolMapper(fed), report)
}

// convergeLDAPComponent ensures exactly the declared LDAP federation exists
// (or none). The enumeration iterates the operator-NAMED component, never
// the realm's component inventory -- an admin-created federation provider
// under another name is never even diffed.
func convergeLDAPComponent(ctx context.Context, admin *AdminClient, realm, realmID string, ldap *OwnedLDAPFederation, report *Report) error {
	components, err := admin.ListComponents(ctx, realm, realmID, userStorageProviderType)
	if err != nil {
		return err
	}
	var live Representation
	for _, c := range components {
		if name, _ := c["name"].(string); name == resources.IdentityLDAPComponentName {
			live = c
			break
		}
	}

	if ldap == nil {
		if live != nil {
			id, _ := live["id"].(string)
			if err := admin.DeleteComponent(ctx, realm, id); err != nil {
				return err
			}
			report.repaired("LDAP federation %s removed (no identity provider is bound)", resources.IdentityLDAPComponentName)
		}
		return convergeDirectoryGroupsParent(ctx, admin, realm, false, report)
	}

	if live == nil {
		rep := Representation{
			"name":         resources.IdentityLDAPComponentName,
			"providerId":   "ldap",
			"providerType": userStorageProviderType,
			"parentId":     realmID,
			"config":       ldap.componentConfig(),
		}
		// The credential rides only creation and rotation writes.
		setComponentConfigValue(rep, "bindCredential", ldap.BindCredential)
		if err := admin.CreateComponent(ctx, realm, rep); err != nil {
			return err
		}
		report.repaired("LDAP federation %s created", resources.IdentityLDAPComponentName)
		// Re-fetch for the id the mapper convergence below needs.
		components, err = admin.ListComponents(ctx, realm, realmID, userStorageProviderType)
		if err != nil {
			return err
		}
		for _, c := range components {
			if name, _ := c["name"].(string); name == resources.IdentityLDAPComponentName {
				live = c
				break
			}
		}
		if live == nil {
			return fmt.Errorf("LDAP federation component not found immediately after creation")
		}
	} else if err := convergeComponentConfig(ctx, admin, realm, live, ldap.componentConfig(), ldap.rotationWrite(), report); err != nil {
		return err
	}

	componentID, _ := live["id"].(string)
	if componentID == "" {
		return fmt.Errorf("LDAP federation component carries no id")
	}

	if err := convergeDirectoryGroupsParent(ctx, admin, realm, true, report); err != nil {
		return err
	}
	return convergeLDAPMappers(ctx, admin, realm, componentID, ldap, report)
}

// rotationWrite returns the secret key/value pair to force-write when the
// referenced Secret's content moved, nil otherwise.
func (f *OwnedLDAPFederation) rotationWrite() map[string]string {
	if !f.RotateCredential {
		return nil
	}
	return map[string]string{"bindCredential": f.BindCredential}
}

// convergeComponentConfig read-modify-writes owned config keys on a live
// component: only listed keys are corrected, everything else rides back
// untouched. rotation carries secret keys that must be written this pass
// even though their live values are masked and undiffable.
func convergeComponentConfig(ctx context.Context, admin *AdminClient, realm string, live Representation, owned map[string][]string, rotation map[string]string, report *Report) error {
	name, _ := live["name"].(string)
	liveConfig, _ := live["config"].(map[string]any)
	if liveConfig == nil {
		liveConfig = map[string]any{}
	}

	var drifted []string
	for key, want := range owned {
		if !jsonEqual(want, liveConfig[key]) {
			liveConfig[key] = want
			drifted = append(drifted, key)
		}
	}
	for key, value := range rotation {
		liveConfig[key] = []string{value}
		drifted = append(drifted, key+" (rotated)")
	}
	if len(drifted) == 0 {
		return nil
	}

	live["config"] = liveConfig
	id, _ := live["id"].(string)
	if err := admin.UpdateComponent(ctx, realm, id, live); err != nil {
		return err
	}
	report.repaired("component %s corrected: %v", name, drifted)
	return nil
}

// convergeDirectoryGroupsParent ensures the /directory realm group exists
// while LDAP federation is declared, and removes it (with the mirrored
// groups under it) when none is. Admin-created realm groups live outside
// this path by construction.
func convergeDirectoryGroupsParent(ctx context.Context, admin *AdminClient, realm string, desired bool, report *Report) error {
	live, found, err := admin.GetGroupByPath(ctx, realm, resources.IdentityDirectoryGroupsPath)
	if err != nil {
		return err
	}

	switch {
	case desired && !found:
		if err := admin.CreateGroup(ctx, realm, Representation{
			"name": directoryGroupsParentName,
			"attributes": map[string][]string{
				resources.IdentityManagedAttribute: {"true"},
			},
		}); err != nil {
			return err
		}
		report.repaired("realm group %s created (directory groups mirror under it)", resources.IdentityDirectoryGroupsPath)
	case !desired && found:
		id, _ := live["id"].(string)
		if err := admin.do(ctx, "DELETE", admin.adminPath(realm, "/groups/"+id), nil, 204, nil); err != nil {
			return fmt.Errorf("deleting directory groups parent: %w", err)
		}
		report.repaired("realm group %s removed (no identity provider is bound)", resources.IdentityDirectoryGroupsPath)
	}
	return nil
}

// convergeLDAPMappers ensures the group-sync mapper exists with owned config
// and the manifest's attribute choices override the Keycloak-created
// user-attribute mappers. Child mappers beyond the enumeration (including
// any an admin adds under the operator's component) are never touched.
func convergeLDAPMappers(ctx context.Context, admin *AdminClient, realm, componentID string, ldap *OwnedLDAPFederation, report *Report) error {
	mappers, err := admin.ListComponents(ctx, realm, componentID, ldapStorageMapperType)
	if err != nil {
		return err
	}
	byName := make(map[string]Representation, len(mappers))
	for _, m := range mappers {
		if name, _ := m["name"].(string); name != "" {
			byName[name] = m
		}
	}

	if live, ok := byName[ldapGroupMapperName]; ok {
		if err := convergeComponentConfig(ctx, admin, realm, live, ldap.groupMapperConfig(), nil, report); err != nil {
			return err
		}
	} else {
		if err := admin.CreateComponent(ctx, realm, Representation{
			"name":         ldapGroupMapperName,
			"providerId":   "group-ldap-mapper",
			"providerType": ldapStorageMapperType,
			"parentId":     componentID,
			"config":       ldap.groupMapperConfig(),
		}); err != nil {
			return err
		}
		report.repaired("LDAP group mapper %s created", ldapGroupMapperName)
	}

	// The manifest's attribute schema onto the owned user-attribute mappers:
	// a live one converges on its ldap.attribute key (Keycloak's own
	// defaults for the rest stay its business); a missing one is created
	// whole. Deterministic order so the report reads the same every pass.
	names := make([]string, 0, len(ldap.ownedAttributeMappers()))
	for name := range ldap.ownedAttributeMappers() {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, mapperName := range names {
		owned := ldap.ownedAttributeMappers()[mapperName]
		if live, ok := byName[mapperName]; ok {
			if err := convergeComponentConfig(ctx, admin, realm, live,
				map[string][]string{"ldap.attribute": {owned.ldapAttribute}}, nil, report); err != nil {
				return err
			}
			continue
		}
		if err := admin.CreateComponent(ctx, realm, Representation{
			"name":         mapperName,
			"providerId":   "user-attribute-ldap-mapper",
			"providerType": ldapStorageMapperType,
			"parentId":     componentID,
			"config":       ldapAttributeMapperConfig(owned),
		}); err != nil {
			return err
		}
		report.repaired("LDAP attribute mapper %q created (%s <- %s)", mapperName, owned.userModelAttribute, owned.ldapAttribute)
	}

	// The vendor's full-name mapper would overwrite the names the owned
	// mappers just imported; with first and last name owned explicitly it
	// has no job left.
	if live, ok := byName[vendorFullNameMapperName]; ok {
		id, _ := live["id"].(string)
		if err := admin.DeleteComponent(ctx, realm, id); err != nil {
			return err
		}
		report.repaired("LDAP mapper %q removed: first and last name are mapped from the manifest's attributes", vendorFullNameMapperName)
	}
	return nil
}

// convergeBroker ensures exactly the declared broker instance exists (or
// none), keyed by the constant alias.
func convergeBroker(ctx context.Context, admin *AdminClient, realm string, broker *OwnedOIDCBroker, report *Report) error {
	instances, err := admin.ListIdentityProviders(ctx, realm)
	if err != nil {
		return err
	}
	var live Representation
	for _, inst := range instances {
		if alias, _ := inst["alias"].(string); alias == resources.IdentityBrokerAlias {
			live = inst
			break
		}
	}

	if broker == nil {
		if live != nil {
			if err := admin.DeleteIdentityProvider(ctx, realm, resources.IdentityBrokerAlias); err != nil {
				return err
			}
			report.repaired("identity broker %s removed (no identity provider is bound)", resources.IdentityBrokerAlias)
		}
		return nil
	}

	if live == nil {
		cfg := broker.brokerConfig()
		cfg["clientSecret"] = broker.ClientSecret
		if err := admin.CreateIdentityProvider(ctx, realm, Representation{
			"alias":       resources.IdentityBrokerAlias,
			"displayName": broker.DisplayName,
			"providerId":  "oidc",
			"enabled":     true,
			// Upstream emails are the company's IT-managed identities --
			// the same trust stance as the LDAP arm's trustEmail.
			"trustEmail": true,
			"storeToken": false,
			"config":     cfg,
		}); err != nil {
			return err
		}
		report.repaired("identity broker %s created", resources.IdentityBrokerAlias)
	} else if err := convergeBrokerFields(ctx, admin, realm, live, broker, report); err != nil {
		return err
	}

	return convergeBrokerMappers(ctx, admin, realm, broker, report)
}

// convergeBrokerFields read-modify-writes the owned top-level fields and
// config keys on a live broker instance.
func convergeBrokerFields(ctx context.Context, admin *AdminClient, realm string, live Representation, broker *OwnedOIDCBroker, report *Report) error {
	var drifted []string

	ownedFields := map[string]any{
		"displayName": broker.DisplayName,
		"enabled":     true,
		"trustEmail":  true,
	}
	for key, want := range ownedFields {
		if !jsonEqual(want, live[key]) {
			live[key] = want
			drifted = append(drifted, key)
		}
	}

	liveConfig, _ := live["config"].(map[string]any)
	if liveConfig == nil {
		liveConfig = map[string]any{}
	}
	for key, want := range broker.brokerConfig() {
		if got, _ := liveConfig[key].(string); got != want {
			liveConfig[key] = want
			drifted = append(drifted, "config."+key)
		}
	}
	if broker.RotateCredential {
		liveConfig["clientSecret"] = broker.ClientSecret
		drifted = append(drifted, "config.clientSecret (rotated)")
	}
	live["config"] = liveConfig

	if len(drifted) == 0 {
		return nil
	}
	if err := admin.UpdateIdentityProvider(ctx, realm, resources.IdentityBrokerAlias, live); err != nil {
		return err
	}
	report.repaired("identity broker %s corrected: %v", resources.IdentityBrokerAlias, drifted)
	return nil
}

// brokerImporterNames are the operator-owned claim importers under the
// broker instance -- package-local vocabulary like ldapGroupMapperName.
const (
	brokerGroupsImporterName  = "planton-directory-groups-import"
	brokerSubjectImporterName = "planton-directory-subject-import"
)

// convergeBrokerMappers ensures the owned claim importers (groups + subject)
// exist with owned config. Importers an admin adds under other names ride
// through untouched.
func convergeBrokerMappers(ctx context.Context, admin *AdminClient, realm string, broker *OwnedOIDCBroker, report *Report) error {
	mappers, err := admin.ListIdentityProviderMappers(ctx, realm, resources.IdentityBrokerAlias)
	if err != nil {
		return err
	}
	byName := make(map[string]Representation, len(mappers))
	for _, m := range mappers {
		if name, _ := m["name"].(string); name != "" {
			byName[name] = m
		}
	}

	owned := map[string]map[string]string{
		brokerGroupsImporterName:  broker.groupsImporterConfig(),
		brokerSubjectImporterName: broker.subjectImporterConfig(),
	}
	for _, importerName := range []string{brokerGroupsImporterName, brokerSubjectImporterName} {
		if err := convergeBrokerImporter(ctx, admin, realm, importerName, owned[importerName], byName[importerName], report); err != nil {
			return err
		}
	}
	return nil
}

// convergeBrokerImporter ensures one named claim importer exists with its
// owned config keys.
func convergeBrokerImporter(ctx context.Context, admin *AdminClient, realm, importerName string, ownedConfig map[string]string, live Representation, report *Report) error {
	if live == nil {
		if err := admin.CreateIdentityProviderMapper(ctx, realm, resources.IdentityBrokerAlias, Representation{
			"name":                   importerName,
			"identityProviderAlias":  resources.IdentityBrokerAlias,
			"identityProviderMapper": "oidc-user-attribute-idp-mapper",
			"config":                 ownedConfig,
		}); err != nil {
			return err
		}
		report.repaired("broker importer %s created", importerName)
		return nil
	}

	liveConfig, _ := live["config"].(map[string]any)
	if liveConfig == nil {
		liveConfig = map[string]any{}
	}
	var drifted []string
	for key, want := range ownedConfig {
		if got, _ := liveConfig[key].(string); got != want {
			liveConfig[key] = want
			drifted = append(drifted, key)
		}
	}
	if len(drifted) == 0 {
		return nil
	}
	live["config"] = liveConfig
	mapperID, _ := live["id"].(string)
	if err := admin.UpdateIdentityProviderMapper(ctx, realm, resources.IdentityBrokerAlias, mapperID, live); err != nil {
		return err
	}
	report.repaired("broker importer %s corrected: %v", importerName, drifted)
	return nil
}

// convergeSignInClientMapper ensures the sign-in clients carry the
// arm-appropriate mapper owned by the given constant NAME -- and none when no
// federation is bound (desired nil sweeps it). A type change (arm switch) is
// delete-and-recreate because Keycloak cannot change a mapper's type in
// place. Serves both federation-state mappers (groups, directory-subject);
// admin-added mappers under other names are structurally invisible.
func convergeSignInClientMapper(ctx context.Context, admin *AdminClient, realm, mapperName string, desired *OwnedMapper, report *Report) error {
	for _, clientID := range []string{resources.IdentityConsoleClientID, resources.IdentityCLIClientID} {
		client, found, err := admin.FindClientByClientID(ctx, realm, clientID)
		if err != nil {
			return err
		}
		if !found {
			// The client phase runs before this one and creates it; a miss
			// here means that phase failed and already reported.
			continue
		}
		clientUUID, _ := client["id"].(string)

		mappers, err := admin.ListProtocolMappers(ctx, realm, clientUUID)
		if err != nil {
			return err
		}
		var live Representation
		for _, m := range mappers {
			if name, _ := m["name"].(string); name == mapperName {
				live = m
				break
			}
		}

		switch {
		case desired == nil && live != nil:
			mapperID, _ := live["id"].(string)
			if err := admin.DeleteProtocolMapper(ctx, realm, clientUUID, mapperID); err != nil {
				return err
			}
			report.repaired("client %s mapper %s removed (no identity provider is bound)", clientID, mapperName)

		case desired != nil && live == nil:
			if err := admin.CreateProtocolMapper(ctx, realm, clientUUID, Representation{
				"name":            desired.Name,
				"protocol":        "openid-connect",
				"protocolMapper":  desired.ProtocolMapper,
				"consentRequired": false,
				"config":          desired.Config,
			}); err != nil {
				return err
			}
			report.repaired("client %s gained the %s mapper %s", clientID, desired.ProtocolMapper, mapperName)

		case desired != nil && live != nil:
			if liveType, _ := live["protocolMapper"].(string); liveType != desired.ProtocolMapper {
				mapperID, _ := live["id"].(string)
				if err := admin.DeleteProtocolMapper(ctx, realm, clientUUID, mapperID); err != nil {
					return err
				}
				if err := admin.CreateProtocolMapper(ctx, realm, clientUUID, Representation{
					"name":            desired.Name,
					"protocol":        "openid-connect",
					"protocolMapper":  desired.ProtocolMapper,
					"consentRequired": false,
					"config":          desired.Config,
				}); err != nil {
					return err
				}
				report.repaired("client %s mapper %s recreated as %s (federation arm changed)", clientID, mapperName, desired.ProtocolMapper)
				continue
			}
			liveConfig, _ := live["config"].(map[string]any)
			if liveConfig == nil {
				liveConfig = map[string]any{}
			}
			var drifted []string
			for key, want := range desired.Config {
				if got, _ := liveConfig[key].(string); got != want {
					liveConfig[key] = want
					drifted = append(drifted, key)
				}
			}
			if len(drifted) == 0 {
				continue
			}
			live["config"] = liveConfig
			mapperID, _ := live["id"].(string)
			if err := admin.UpdateProtocolMapper(ctx, realm, clientUUID, mapperID, live); err != nil {
				return err
			}
			report.repaired("client %s mapper %s corrected: %v", clientID, mapperName, drifted)
		}
	}
	return nil
}

// setComponentConfigValue sets one config key on a component representation
// being built for creation.
func setComponentConfigValue(rep Representation, key, value string) {
	cfg, _ := rep["config"].(map[string][]string)
	if cfg == nil {
		cfg = map[string][]string{}
		rep["config"] = cfg
	}
	cfg[key] = []string{value}
}
