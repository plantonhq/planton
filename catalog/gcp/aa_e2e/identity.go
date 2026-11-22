package aa_e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/auth"
	"github.com/pkg/errors"
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	permissionsv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
	"github.com/plantonhq/planton/pkg/iac/actioninventory"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/permissions"
	"github.com/plantonhq/planton/shared/catalogkind"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	serviceusage "google.golang.org/api/serviceusage/v1"
	"sigs.k8s.io/yaml"
)

// The identity a GCP scenario may declare with planton.dev/e2e-identity.
//
//	declared   a service account holding exactly what the kind's
//	           iac/permissions.yaml grants for this scenario's manifest
//
// The lane deploys, re-plans, updates and destroys the kind as that account,
// which is what the catalog means by `proven`: an identity holding exactly
// these permissions, nothing more, did all of it. The fixtures, the minting
// and the verifiers keep the harness's own credential (the fixtures are the
// lane's stage, not its subject).
//
// Which permissions: every group with no condition, plus each conditional
// group whose spec field the scenario's manifest, or its second act, sets
// (permissions.ConditionNeeded, the rule the platform's runner policy uses).
// One identity deploys both acts, so it holds what the upgrade turns on too. A customer's
// connection is granted the conditional groups too, so a lane proves the
// tighter claim: the unconditional groups suffice, and each condition is
// drawn where the module needs it.
//
// How they are granted, as the platform renders a connection's grants: each
// group where it reaches (permissionsv1.PermissionReach), since a role on a
// project cannot hold a permission Google authorizes above it.
//
//	target           on the project the kind targets: the groups that name
//	                 no predefined role become one project custom role, and
//	                 each named predefined role is bound beside it
//	organization     on the test organization (organizationEnv), split the
//	                 same way, with the custom role written on the
//	                 organization: Google authorizes these permissions only
//	                 there or on a folder, and a folder cannot hold a custom
//	                 role
//	billing_account  on the test billing account (billingAccountEnv): the
//	                 group's predefined role alone, since a billing account
//	                 takes custom roles only from the organization that owns
//	                 it, which the harness's own may not be
//
// A scope's variable is named in the scenario's planton.dev/e2e-required-env,
// so a lane without it skips before it gets here.
//
// Which project: the harness's own, unless the scenario names the project
// its kind targets with provider.IdentityProjectAnnotation (a policy that
// lives in the second test project). The custom role is written and bound
// there, since Google binds a project's custom role only inside that
// project; the account itself always lives in the harness's project, where
// CleanLaneIdentities finds a crashed lane's leftovers.
const identityDeclared = "declared"

const (
	// laneAccountPrefix starts every lane's service account id, so a crashed
	// process's leftovers are recognizable (see CleanLaneIdentities).
	laneAccountPrefix = "pe2e-"
	// declaredRolePrefix starts every standing per-kind custom role id.
	declaredRolePrefix = "plantonE2eDeclared"
	// readinessBudget bounds the wait for Google to answer a new account's
	// grants. IAM propagation is usually seconds and occasionally minutes.
	readinessBudget = 5 * time.Minute
	// aboveProjectReadinessBudget bounds the same wait when the lane is also
	// granted on an organization or a billing account. Google gives a role
	// grant "typically 2 minutes, potentially 7 minutes or longer" to
	// propagate (IAM's access change propagation page); project grants were
	// measured in seconds, which the shorter budget keeps honest, and the
	// scopes above the project get the documented seven minutes with room
	// for the key's own settle.
	aboveProjectReadinessBudget = 10 * time.Minute
	// keySettleWindow is how long a new key must be accepted without a single
	// refusal before the lane uses it. Google says a key may need 60 seconds or
	// more before every operation accepts it, and acceptance is not monotonic:
	// measured, a key that had just answered testIamPermissions was refused by
	// the engine's next token fetch with invalid_grant.
	keySettleWindow = 60 * time.Second
)

// ProvisionIdentity implements provider.IdentityProvisioner (see the
// identityDeclared comment for what the identity holds).
func (h *Harness) ProvisionIdentity(ctx context.Context, tc *provider.KindTestContext, spec string) (string, func(), error) {
	if spec != identityDeclared {
		return "", nil, errors.Errorf("unknown identity %q: the GCP harness accepts %q", spec, identityDeclared)
	}
	// An access token in the process environment outranks a key in both the
	// google provider and pulumi-gcp, so the lane would deploy as whoever the
	// token belongs to rather than as its declared identity.
	if os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN") != "" {
		return "", nil, errors.New("GOOGLE_OAUTH_ACCESS_TOKEN is set in the test process; it outranks the lane's key in both engines' google providers, so the lane could not run as its declared identity -- unset it")
	}
	start := time.Now()
	acts := []string{tc.ManifestPath}
	if secondAct, _ := ctx.Value(provider.SecondActManifestPathKey{}).(string); secondAct != "" {
		if err := requireUpgradeExpectation(secondAct); err != nil {
			return "", nil, err
		}
		acts = append(acts, secondAct)
	}
	grant, err := declaredGrant(tc.RepoRoot, tc.Kind, acts)
	if err != nil {
		return "", nil, err
	}
	project, err := h.identityProject(tc)
	if err != nil {
		return "", nil, err
	}
	scopes, err := h.grantScopes(project, grant)
	if err != nil {
		return "", nil, err
	}
	if err := h.enableCallerProjectServices(ctx, tc, grant.callerProjectServices); err != nil {
		return "", nil, err
	}
	lane := laneLabel(tc)
	accountID := laneAccountID(lane)

	for _, scope := range scopes {
		if len(scope.grant.permissions) == 0 {
			continue
		}
		role, err := h.ensureDeclaredRole(ctx, scope.policy.name, grant.kind, grant.conditionalPurposes, scope.grant.permissions)
		if err != nil {
			return "", nil, err
		}
		scope.roles = append(scope.roles, role)
	}
	for _, scope := range scopes {
		scope.roles = append(scope.roles, scope.grant.predefinedRoles...)
	}

	account, err := h.services.Iam.Projects.ServiceAccounts.Create("projects/"+h.services.Project, &iam.CreateServiceAccountRequest{
		AccountId: accountID,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: truncate("e2e "+lane, 100),
			Description: truncate("A lane identity the catalog e2e harness created; it holds exactly what "+permissions.Path("", "gcp", tc.Kind)+" grants, and is deleted after the lane", 256),
		},
	}).Context(ctx).Do()
	if err != nil {
		return "", nil, errors.Wrapf(err, "creating the lane identity %s", accountID)
	}
	member := "serviceAccount:" + account.Email
	cleanup := func() {
		// Whatever the verdict: the bindings on every scope, then the account
		// (its key dies with it). A failure is reported, never fatal -- the
		// lane's verdict is already in, and the next TestMain removes what is
		// left.
		background := context.Background()
		for _, scope := range scopes {
			if err := h.updatePolicy(background, scope.policy, func(policy *cloudresourcemanager.Policy) bool {
				return removeMember(policy, member)
			}); err != nil {
				lanef(tc, "  [identity] WARN: removing %s's grants on %s failed: %v", account.Email, scope.policy.name, err)
			}
		}
		if _, err := h.services.Iam.Projects.ServiceAccounts.Delete(account.Name).Context(background).Do(); err != nil && !isStatus(err, 404) {
			lanef(tc, "  [identity] WARN: deleting %s failed: %v", account.Email, err)
		}
	}

	for _, scope := range scopes {
		// A new account can be unknown to the policy service for a few seconds.
		if err := retryUntil(ctx, time.Minute, func() (bool, error) {
			err := h.updatePolicy(ctx, scope.policy, func(policy *cloudresourcemanager.Policy) bool {
				return addMember(policy, member, scope.roles)
			})
			if err != nil && isStatus(err, 400) && strings.Contains(err.Error(), "does not exist") {
				return false, nil
			}
			return err == nil, err
		}); err != nil {
			cleanup()
			return "", nil, errors.Wrapf(err, "granting %s on %s to the lane identity %s", strings.Join(scope.roles, ", "), scope.policy.name, account.Email)
		}
	}

	var key *iam.ServiceAccountKey
	if err := retryUntil(ctx, time.Minute, func() (bool, error) {
		created, err := h.services.Iam.Projects.ServiceAccounts.Keys.Create(account.Name, &iam.CreateServiceAccountKeyRequest{}).Context(ctx).Do()
		if err != nil && isStatus(err, 404) {
			return false, nil
		}
		key = created
		return err == nil, err
	}); err != nil {
		cleanup()
		return "", nil, errors.Wrapf(err, "minting a key for the lane identity %s", account.Email)
	}
	keyJSON, err := base64.StdEncoding.DecodeString(key.PrivateKeyData)
	if err != nil {
		cleanup()
		return "", nil, errors.Wrap(err, "decoding the lane identity's key")
	}

	if err := h.awaitGrants(ctx, keyJSON, scopes); err != nil {
		cleanup()
		return "", nil, errors.Wrapf(err, "the lane identity %s", account.Email)
	}

	config, removeConfig, err := iacinputproviderconfig.BuildFromProto(
		&gcpprovider.GcpProviderConfig{ServiceAccountKey: string(keyJSON)}, catalogkind.CatalogProvider_gcp)
	if err != nil {
		cleanup()
		return "", nil, errors.Wrap(err, "writing the lane identity's provider configuration")
	}
	held := map[string]bool{}
	var where, auditScopes []string
	for _, scope := range scopes {
		where = append(where, strings.Join(scope.roles, ", ")+" on "+scope.policy.name)
		auditScopes = append(auditScopes, scope.policy.name)
		for _, permission := range scope.grant.awaited() {
			held[permission] = true
		}
	}
	lanef(tc, "  [identity] %s holds %s (%d permission(s) from permissions.yaml%s)",
		account.Email, strings.Join(where, "; "), len(held), grant.conditionalNote())
	firstAct := tc.ManifestPath
	h.mu.Lock()
	h.laneGrants[tc] = held
	h.laneAudits[firstAct] = &laneAudit{tc: tc, project: project, scopes: auditScopes, principal: account.Email, start: start}
	h.mu.Unlock()
	return config.Path, func() {
		h.mu.Lock()
		delete(h.laneGrants, tc)
		delete(h.laneAudits, firstAct)
		h.mu.Unlock()
		cleanup()
		removeConfig()
	}, nil
}

// refusedPermission matches Google's two wordings of a refusal that names a
// permission: "Permission 'p' denied on ..." and "IAM permission 'p' on '...'
// denied".
var refusedPermission = regexp.MustCompile(`[Pp]ermission '([A-Za-z0-9_.]+)'(?: on '[^']*')? denied`)

// HeldPermissionRefused implements provider.GrantPropagationJudge: a deploy
// refused for a permission the lane's identity was granted, on any scope, is
// a grant still reaching the service's own permission cache (Discovery
// Engine, Dialogflow and Cloud Deploy were measured lagging a minute past the
// key's settle). The held set is every permission readiness waited on: the
// custom roles' and those the manifest lists for each predefined role it
// names, so a refusal of anything else fails the lane at once.
func (h *Harness) HeldPermissionRefused(tc *provider.KindTestContext, deployErr error) (string, bool) {
	match := refusedPermission.FindStringSubmatch(deployErr.Error())
	if match == nil {
		return "", false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return match[1], h.laneGrants[tc][match[1]]
}

// declaredGrantSet is what a lane's identity is granted, by where each part
// reaches.
type declaredGrantSet struct {
	kind           string
	target         scopeGrant
	organization   scopeGrant
	billingAccount scopeGrant
	// conditionalPurposes are the conditional groups this scenario turned on;
	// they name the custom roles, since two scenarios of one kind may differ.
	conditionalPurposes []string
	// callerProjectServices are the APIs the kind's calls need on in the
	// project Google bills them to (the manifest's caller_project_services).
	callerProjectServices []string
}

// scopeGrant is what the identity is granted on one scope: the permissions
// its standing custom role holds there, the predefined roles bound beside
// it, and the permissions the manifest lists for those roles. Readiness
// waits on both permission lists, since a predefined role propagates no
// faster than a custom one.
type scopeGrant struct {
	permissions     []string
	predefinedRoles []string
	rolePermissions []string
}

func (g scopeGrant) empty() bool {
	return len(g.permissions) == 0 && len(g.predefinedRoles) == 0
}

// awaited is every permission readiness waits on for the scope, sorted.
func (g scopeGrant) awaited() []string {
	seen := map[string]bool{}
	var all []string
	for _, permission := range append(append([]string{}, g.permissions...), g.rolePermissions...) {
		if !seen[permission] {
			seen[permission] = true
			all = append(all, permission)
		}
	}
	sort.Strings(all)
	return all
}

func (g declaredGrantSet) conditionalNote() string {
	if len(g.conditionalPurposes) == 0 {
		return ""
	}
	return "; conditional: " + strings.Join(g.conditionalPurposes, ", ")
}

// declaredGrant reads the kind's manifest and the specs of the scenario's
// acts (the first act, then the second when it has one) and decides the
// lane's grant, each group on the scope it reaches. It refuses, before
// anything is minted, a permission Google keeps out of custom roles (which
// the manifest must cover with a predefined role instead) and a billing
// account group that names no predefined role.
func declaredGrant(repoRoot, kindDir string, actManifestPaths []string) (declaredGrantSet, error) {
	manifest, err := permissions.Load(repoRoot, "gcp", kindDir)
	if err != nil {
		return declaredGrantSet{}, errors.Wrap(err, "the declared identity is built from the kind's permissions manifest")
	}
	var kind string
	specs := make([]map[string]interface{}, 0, len(actManifestPaths))
	for _, path := range actManifestPaths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return declaredGrantSet{}, errors.Wrap(err, "reading the scenario manifest")
		}
		var act struct {
			Kind string                 `json:"kind"`
			Spec map[string]interface{} `json:"spec"`
		}
		if err := yaml.Unmarshal(raw, &act); err != nil {
			return declaredGrantSet{}, errors.Wrapf(err, "parsing the scenario manifest %s", path)
		}
		if kind == "" {
			kind = act.Kind
		}
		specs = append(specs, act.Spec)
	}
	inventory, err := actioninventory.LoadGcp(filepath.Join(repoRoot, "pkg", "iac", "actioninventory"))
	if err != nil {
		return declaredGrantSet{}, errors.Wrap(err, "loading Google's permission inventory")
	}

	grant := declaredGrantSet{kind: kind, callerProjectServices: manifest.GetSpec().GetGcp().GetCallerProjectServices()}
	manifestPath := permissions.Path(repoRoot, "gcp", kindDir)
	for _, group := range manifest.GetSpec().GetGcp().GetGroups() {
		condition := group.GetCondition().GetSpecFieldSet()
		if !permissions.ConditionNeeded(condition, specs) {
			continue
		}
		var scope *scopeGrant
		switch group.GetReach() {
		case permissionsv1.PermissionReach_target:
			scope = &grant.target
		case permissionsv1.PermissionReach_organization:
			scope = &grant.organization
		case permissionsv1.PermissionReach_billing_account:
			if group.GetPredefinedRole() == "" {
				return declaredGrantSet{}, errors.Errorf("%s group %q reaches the billing account with permissions and no predefined role; a billing account takes custom roles only from the organization that owns it, so the harness binds a billing account group's predefined role alone -- the group must name the role that carries its permissions",
					manifestPath, group.GetPurpose())
			}
			scope = &grant.billingAccount
		default:
			return declaredGrantSet{}, errors.Errorf("%s group %q has no reach", manifestPath, group.GetPurpose())
		}
		if condition != "" {
			grant.conditionalPurposes = append(grant.conditionalPurposes, group.GetPurpose())
		}
		if role := group.GetPredefinedRole(); role != "" {
			scope.predefinedRoles = appendNew(scope.predefinedRoles, role)
			scope.rolePermissions = appendNew(scope.rolePermissions, group.GetPermissions()...)
			continue
		}
		for _, permission := range group.GetPermissions() {
			if inventory.CustomRoleUnsupported(permission) {
				return declaredGrantSet{}, errors.Errorf("%s group %q holds %s, which Google keeps out of custom roles: the group must name the predefined role that carries it",
					manifestPath, group.GetPurpose(), permission)
			}
			scope.permissions = appendNew(scope.permissions, permission)
		}
	}
	for _, scope := range []*scopeGrant{&grant.target, &grant.organization, &grant.billingAccount} {
		sort.Strings(scope.permissions)
		sort.Strings(scope.rolePermissions)
	}
	sort.Strings(grant.conditionalPurposes)
	if grant.target.empty() && grant.organization.empty() && grant.billingAccount.empty() {
		return declaredGrantSet{}, errors.Errorf("%s grants nothing for this scenario, so there is no identity to declare", manifestPath)
	}
	return grant, nil
}

// appendNew appends each value the list does not hold yet.
func appendNew(list []string, values ...string) []string {
	for _, value := range values {
		if !containsString(list, value) {
			list = append(list, value)
		}
	}
	return list
}

// grantScope is one resource the lane's identity is bound on: its policy,
// what the grant holds there, the roles bound (the custom role's name is
// known once it is written), and how the identity asks Google, as itself,
// which permissions it holds there.
type grantScope struct {
	policy iamPolicy
	grant  scopeGrant
	roles  []string
	test   permissionTest
}

// permissionTest answers, for the identity whose key it is given, which of
// the permissions it holds on one scope (that scope's testIamPermissions).
type permissionTest func(ctx context.Context, keyJSON []byte, permissions []string) ([]string, error)

// grantScopes resolves where each part of the grant is bound: the project,
// then the organization and the billing account their variables name. A
// part whose variable is unset fails the lane with the variable named; the
// scenario declares it in planton.dev/e2e-required-env, so normally the lane
// skipped before this.
func (h *Harness) grantScopes(project string, grant declaredGrantSet) ([]*grantScope, error) {
	var scopes []*grantScope
	if !grant.target.empty() {
		scopes = append(scopes, &grantScope{policy: h.projectPolicy(project), grant: grant.target, test: h.projectPermissionTest(project)})
	}
	if !grant.organization.empty() {
		organization, err := scopeFromEnv(organizationEnv, "organizations/", grant.kind, "the organization")
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, &grantScope{policy: h.organizationPolicy(organization), grant: grant.organization, test: h.organizationPermissionTest(organization)})
	}
	if !grant.billingAccount.empty() {
		account, err := scopeFromEnv(billingAccountEnv, "billingAccounts/", grant.kind, "the billing account")
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, &grantScope{policy: h.billingAccountPolicy(account), grant: grant.billingAccount, test: h.billingAccountPermissionTest(account)})
	}
	return scopes, nil
}

// scopeFromEnv is the resource name a scope's variable holds, with or
// without its collection prefix.
func scopeFromEnv(variable, prefix, kind, what string) (string, error) {
	value := strings.TrimPrefix(os.Getenv(variable), prefix)
	if value == "" {
		return "", errors.Errorf("%s's permissions manifest grants on %s, and %s is unset: set it to the test %s's id, and list it in the scenario's planton.dev/e2e-required-env so a lane without it skips",
			kind, what, variable, strings.TrimPrefix(what, "the "))
	}
	return prefix + value, nil
}

// enableCallerProjectServices turns on, in the harness's project, the APIs
// the kind's calls need where Google bills them: a resource in no project
// (an organization policy, a billing budget) is managed through calls
// charged to the caller's own project, and the lane's account lives in the
// harness's. A customer's connection setup turns the same list on in the
// connection's project when it is granted the reach these kinds need.
// Turning on a service already on is a no-op, so only the missing ones are
// sent.
func (h *Harness) enableCallerProjectServices(ctx context.Context, tc *provider.KindTestContext, services []string) error {
	if len(services) == 0 {
		return nil
	}
	parent := "projects/" + h.services.Project
	names := make([]string, 0, len(services))
	for _, service := range services {
		names = append(names, parent+"/services/"+service)
	}
	current, err := h.serviceUsage.Services.BatchGet(parent).Names(names...).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "reading whether %s are on in %s", strings.Join(services, ", "), h.services.Project)
	}
	var missing []string
	for _, service := range current.Services {
		if service.State != "ENABLED" {
			missing = append(missing, service.Name[strings.LastIndex(service.Name, "/")+1:])
		}
	}
	if len(missing) == 0 {
		lanef(tc, "  [identity] %s already on in %s, where Google bills the lane identity's calls", strings.Join(services, ", "), h.services.Project)
		return nil
	}
	operation, err := h.serviceUsage.Services.BatchEnable(parent, &serviceusage.BatchEnableServicesRequest{ServiceIds: missing}).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "turning on %s in %s", strings.Join(missing, ", "), h.services.Project)
	}
	if err := retryUntil(ctx, 5*time.Minute, func() (bool, error) {
		if operation.Done {
			if operation.Error != nil {
				return false, errors.Errorf("%s", operation.Error.Message)
			}
			return true, nil
		}
		operation, err = h.serviceUsage.Operations.Get(operation.Name).Context(ctx).Do()
		return false, err
	}); err != nil {
		return errors.Wrapf(err, "turning on %s in %s", strings.Join(missing, ", "), h.services.Project)
	}
	lanef(tc, "  [identity] turned on %s in %s, where Google bills the lane identity's calls", strings.Join(missing, ", "), h.services.Project)
	return nil
}

// declaredRoleID is the standing custom role a kind's lanes share:
// plantonE2eDeclared<Kind>, plus C and a short hash of the conditional groups
// a scenario turns on, so scenarios with different conditional groups never
// rewrite each other's role. It is keyed by what the role is for, never by
// its permissions, so a manifest fix rewrites the role in place. A project
// and the organization each hold their own role under the same id.
func declaredRoleID(kind string, conditionalPurposes []string) string {
	id := declaredRolePrefix + kind
	if len(conditionalPurposes) > 0 {
		sum := sha256.Sum256([]byte(strings.Join(conditionalPurposes, ",")))
		id += "C" + hex.EncodeToString(sum[:])[:4]
	}
	return id
}

// ensureDeclaredRole writes the kind's standing custom role on parent (a
// project or the organization) to exactly the given permissions and returns
// its name. It is never deleted: Google counts roles against a per-project
// and per-organization cap and holds a deleted role's id, and a standing
// role rewritten before every lane costs neither.
func (h *Harness) ensureDeclaredRole(ctx context.Context, parent, kind string, conditionalPurposes, rolePermissions []string) (string, error) {
	roleID := declaredRoleID(kind, conditionalPurposes)
	name := parent + "/roles/" + roleID
	desired := &iam.Role{
		Title:               truncate("Planton e2e declared "+kind, 100),
		Description:         truncate("Exactly what the catalog's permissions manifest for "+kind+" grants here; rewritten before every declared lane", 256),
		IncludedPermissions: rolePermissions,
		Stage:               "GA",
	}
	roles := h.roleStore(parent)
	existing, err := roles.get(ctx, name)
	switch {
	case err != nil && isStatus(err, 404):
		err = roles.create(ctx, parent, &iam.CreateRoleRequest{RoleId: roleID, Role: desired})
		if err == nil {
			return name, nil
		}
		// The other engine's lane of the same kind created it a moment ago.
		if !isStatus(err, 409) {
			return "", errors.Wrapf(err, "creating the declared role %s", name)
		}
	case err != nil:
		return "", errors.Wrapf(err, "reading the declared role %s", name)
	case existing.Deleted:
		if err := roles.undelete(ctx, name, existing.Etag); err != nil {
			return "", errors.Wrapf(err, "undeleting the declared role %s", name)
		}
	}
	if err := roles.patch(ctx, name, desired); err != nil {
		return "", errors.Wrapf(err, "writing the declared role %s", name)
	}
	return name, nil
}

// roleStore is IAM's custom-role calls for one parent: a project's and an
// organization's roles are the same resource behind two services.
type roleStore struct {
	get      func(ctx context.Context, name string) (*iam.Role, error)
	create   func(ctx context.Context, parent string, request *iam.CreateRoleRequest) error
	undelete func(ctx context.Context, name, etag string) error
	patch    func(ctx context.Context, name string, role *iam.Role) error
}

const roleFields = "title,description,includedPermissions,stage"

func (h *Harness) roleStore(parent string) roleStore {
	if strings.HasPrefix(parent, "organizations/") {
		roles := h.services.Iam.Organizations.Roles
		return roleStore{
			get: func(ctx context.Context, name string) (*iam.Role, error) { return roles.Get(name).Context(ctx).Do() },
			create: func(ctx context.Context, parent string, request *iam.CreateRoleRequest) error {
				_, err := roles.Create(parent, request).Context(ctx).Do()
				return err
			},
			undelete: func(ctx context.Context, name, etag string) error {
				_, err := roles.Undelete(name, &iam.UndeleteRoleRequest{Etag: etag}).Context(ctx).Do()
				return err
			},
			patch: func(ctx context.Context, name string, role *iam.Role) error {
				_, err := roles.Patch(name, role).UpdateMask(roleFields).Context(ctx).Do()
				return err
			},
		}
	}
	roles := h.services.Iam.Projects.Roles
	return roleStore{
		get: func(ctx context.Context, name string) (*iam.Role, error) { return roles.Get(name).Context(ctx).Do() },
		create: func(ctx context.Context, parent string, request *iam.CreateRoleRequest) error {
			_, err := roles.Create(parent, request).Context(ctx).Do()
			return err
		},
		undelete: func(ctx context.Context, name, etag string) error {
			_, err := roles.Undelete(name, &iam.UndeleteRoleRequest{Etag: etag}).Context(ctx).Do()
			return err
		},
		patch: func(ctx context.Context, name string, role *iam.Role) error {
			_, err := roles.Patch(name, role).UpdateMask(roleFields).Context(ctx).Do()
			return err
		},
	}
}

// awaitGrants waits until Google answers, for the new account itself, that it
// holds every awaited permission on every scope it was granted on, and has
// kept accepting its key for keySettleWindow: each attempt builds fresh
// clients, so each fetches a fresh token, and any refusal on any scope starts
// the window again. Some services check permissions through their own caches
// and can lag even this; a 403 in the lane for a permission the identity
// holds is therefore re-run after a wait before anything is concluded (the
// e2e README's triage section).
func (h *Harness) awaitGrants(ctx context.Context, keyJSON []byte, scopes []*grantScope) error {
	budget := readinessBudget
	waited := 0
	for _, scope := range scopes {
		waited += len(scope.grant.awaited())
		if !strings.HasPrefix(scope.policy.name, "projects/") {
			budget = aboveProjectReadinessBudget
		}
	}
	if waited == 0 {
		return nil
	}
	var missing []string
	var acceptedSince time.Time
	err := retryUntil(ctx, budget, func() (bool, error) {
		missing = missing[:0]
		for _, scope := range scopes {
			wanted := scope.grant.awaited()
			for start := 0; start < len(wanted); start += 100 {
				end := start + 100
				if end > len(wanted) {
					end = len(wanted)
				}
				answered, err := scope.test(ctx, keyJSON, wanted[start:end])
				if err != nil {
					// A key minted a moment ago can be refused until it propagates.
					if keyNotYetAccepted(err) {
						missing = append(missing, wanted[start:end]...)
						continue
					}
					return false, errors.Wrapf(err, "asking %s", scope.policy.name)
				}
				for _, permission := range wanted[start:end] {
					if !containsString(answered, permission) {
						missing = append(missing, permission+" on "+scope.policy.name)
					}
				}
			}
		}
		if len(missing) > 0 {
			acceptedSince = time.Time{}
			return false, nil
		}
		if acceptedSince.IsZero() {
			acceptedSince = time.Now()
		}
		return time.Since(acceptedSince) >= keySettleWindow, nil
	})
	if err != nil && len(missing) > 0 {
		return errors.Wrapf(err, "Google still does not answer %d granted permission(s) for it after %s (first: %s)",
			len(missing), budget, missing[0])
	}
	return err
}

// The identity's own view of each scope. Quota is billed to the account's own
// project, the harness's: a lane granted in the second project holds no
// serviceusage.services.use there, and Google answers a call billed to it
// with 403, which would read as a key not yet accepted until the budget ran
// out.

func (h *Harness) projectPermissionTest(project string) permissionTest {
	return func(ctx context.Context, keyJSON []byte, wanted []string) ([]string, error) {
		asLane, err := cloudresourcemanager.NewService(ctx, option.WithCredentialsJSON(keyJSON), option.WithQuotaProject(h.services.Project))
		if err != nil {
			return nil, errors.Wrap(err, "building a client that authenticates as the lane identity")
		}
		response, err := asLane.Projects.TestIamPermissions(project, &cloudresourcemanager.TestIamPermissionsRequest{Permissions: wanted}).Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		return response.Permissions, nil
	}
}

func (h *Harness) organizationPermissionTest(organization string) permissionTest {
	return func(ctx context.Context, keyJSON []byte, wanted []string) ([]string, error) {
		asLane, err := cloudresourcemanager.NewService(ctx, option.WithCredentialsJSON(keyJSON), option.WithQuotaProject(h.services.Project))
		if err != nil {
			return nil, errors.Wrap(err, "building a client that authenticates as the lane identity")
		}
		response, err := asLane.Organizations.TestIamPermissions(organization, &cloudresourcemanager.TestIamPermissionsRequest{Permissions: wanted}).Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		return response.Permissions, nil
	}
}

func (h *Harness) billingAccountPermissionTest(account string) permissionTest {
	return func(ctx context.Context, keyJSON []byte, wanted []string) ([]string, error) {
		asLane, err := cloudbilling.NewService(ctx, option.WithCredentialsJSON(keyJSON), option.WithQuotaProject(h.services.Project))
		if err != nil {
			return nil, errors.Wrap(err, "building a client that authenticates as the lane identity")
		}
		response, err := asLane.BillingAccounts.TestIamPermissions(account, &cloudbilling.TestIamPermissionsRequest{Permissions: wanted}).Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		return response.Permissions, nil
	}
}

// iamPolicy is one resource's IAM policy, read and written through its own
// API as the harness. A billing account's policy has the shape of Resource
// Manager's, so every edit goes through the one type.
type iamPolicy struct {
	name string
	get  func(ctx context.Context) (*cloudresourcemanager.Policy, error)
	set  func(ctx context.Context, policy *cloudresourcemanager.Policy) error
}

func (h *Harness) projectPolicy(project string) iamPolicy {
	return iamPolicy{
		name: "projects/" + project,
		get: func(ctx context.Context) (*cloudresourcemanager.Policy, error) {
			return h.services.Crm.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{
				Options: &cloudresourcemanager.GetPolicyOptions{RequestedPolicyVersion: 3},
			}).Context(ctx).Do()
		},
		set: func(ctx context.Context, policy *cloudresourcemanager.Policy) error {
			policy.Version = 3
			_, err := h.services.Crm.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
			return err
		},
	}
}

func (h *Harness) organizationPolicy(organization string) iamPolicy {
	return iamPolicy{
		name: organization,
		get: func(ctx context.Context) (*cloudresourcemanager.Policy, error) {
			return h.services.Crm.Organizations.GetIamPolicy(organization, &cloudresourcemanager.GetIamPolicyRequest{
				Options: &cloudresourcemanager.GetPolicyOptions{RequestedPolicyVersion: 3},
			}).Context(ctx).Do()
		},
		set: func(ctx context.Context, policy *cloudresourcemanager.Policy) error {
			policy.Version = 3
			_, err := h.services.Crm.Organizations.SetIamPolicy(organization, &cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
			return err
		},
	}
}

func (h *Harness) billingAccountPolicy(account string) iamPolicy {
	return iamPolicy{
		name: account,
		get: func(ctx context.Context) (*cloudresourcemanager.Policy, error) {
			policy, err := h.billing.BillingAccounts.GetIamPolicy(account).Context(ctx).Do()
			if err != nil {
				return nil, err
			}
			var converted cloudresourcemanager.Policy
			return &converted, convertPolicy(policy, &converted)
		},
		set: func(ctx context.Context, policy *cloudresourcemanager.Policy) error {
			var converted cloudbilling.Policy
			if err := convertPolicy(policy, &converted); err != nil {
				return err
			}
			_, err := h.billing.BillingAccounts.SetIamPolicy(account, &cloudbilling.SetIamPolicyRequest{Policy: &converted}).Context(ctx).Do()
			return err
		},
	}
}

// convertPolicy carries an IAM policy between two APIs' types through its
// JSON, which Google defines once for every resource.
func convertPolicy(from, to interface{}) error {
	raw, err := json.Marshal(from)
	if err != nil {
		return errors.Wrap(err, "converting an IAM policy")
	}
	return errors.Wrap(json.Unmarshal(raw, to), "converting an IAM policy")
}

// updatePolicy is a read-modify-write of a resource's IAM policy that
// retries when another writer got there first: lanes running side by side
// edit the one policy concurrently.
func (h *Harness) updatePolicy(ctx context.Context, resource iamPolicy, change func(*cloudresourcemanager.Policy) bool) error {
	return retryUntil(ctx, 2*time.Minute, func() (bool, error) {
		policy, err := resource.get(ctx)
		if err != nil {
			return false, err
		}
		if !change(policy) {
			return true, nil
		}
		err = resource.set(ctx, policy)
		if err != nil && (isStatus(err, 409) || strings.Contains(err.Error(), "concurrent policy changes")) {
			return false, nil
		}
		return err == nil, err
	})
}

func addMember(policy *cloudresourcemanager.Policy, member string, roles []string) bool {
	changed := false
	for _, role := range roles {
		var binding *cloudresourcemanager.Binding
		for _, b := range policy.Bindings {
			if b.Role == role && b.Condition == nil {
				binding = b
				break
			}
		}
		if binding == nil {
			binding = &cloudresourcemanager.Binding{Role: role}
			policy.Bindings = append(policy.Bindings, binding)
		}
		if !containsString(binding.Members, member) {
			binding.Members = append(binding.Members, member)
			changed = true
		}
	}
	return changed
}

func removeMember(policy *cloudresourcemanager.Policy, member string) bool {
	return removeMembersWhere(policy, func(m string) bool { return m == member })
}

func removeMembersWhere(policy *cloudresourcemanager.Policy, drop func(string) bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, binding := range policy.Bindings {
		members := binding.Members[:0]
		for _, m := range binding.Members {
			if drop(m) {
				changed = true
				continue
			}
			members = append(members, m)
		}
		binding.Members = members
		if len(binding.Members) > 0 {
			kept = append(kept, binding)
		}
	}
	policy.Bindings = kept
	return changed
}

// CleanLaneIdentities removes what crashed lanes left: every lane service
// account in the harness's project; in every test project a lane may be
// granted in, and in the test organization when organizationEnv is set,
// every policy binding that names one or a deleted member; and on the test
// billing account, when billingAccountEnv is set, every binding that names a
// lane account. The harness owns the test projects' and organization's
// policies, so a deleted member there is a lane's; a billing account is
// shared with whoever else it pays for, and their history stays. A killed
// test process (exit 137, a go test timeout) skips the deferred cleanup a
// lane registers, and the rule of one test process per provider account
// means that when a process starts, every lane account already present is
// such a leftover. Returns what it removed.
func (h *Harness) CleanLaneIdentities(ctx context.Context) ([]string, error) {
	var removed []string
	err := h.services.Iam.Projects.ServiceAccounts.List("projects/"+h.services.Project).Pages(ctx, func(page *iam.ListServiceAccountsResponse) error {
		for _, account := range page.Accounts {
			if !strings.HasPrefix(account.Email, laneAccountPrefix) {
				continue
			}
			if _, err := h.services.Iam.Projects.ServiceAccounts.Delete(account.Name).Context(ctx).Do(); err != nil && !isStatus(err, 404) {
				return errors.Wrapf(err, "deleting the leftover lane identity %s", account.Email)
			}
			removed = append(removed, account.Email)
		}
		return nil
	})
	if err != nil {
		return removed, err
	}
	type leftoverBindings struct {
		policy iamPolicy
		drop   func(string) bool
	}
	var policies []leftoverBindings
	for _, project := range h.testProjects() {
		policies = append(policies, leftoverBindings{h.projectPolicy(project), isLaneOrDeletedMember})
	}
	if organization := strings.TrimPrefix(os.Getenv(organizationEnv), "organizations/"); organization != "" {
		policies = append(policies, leftoverBindings{h.organizationPolicy("organizations/" + organization), isLaneOrDeletedMember})
	}
	if account := strings.TrimPrefix(os.Getenv(billingAccountEnv), "billingAccounts/"); account != "" {
		policies = append(policies, leftoverBindings{h.billingAccountPolicy("billingAccounts/" + account), isLaneMember})
	}
	for _, leftover := range policies {
		if err := h.updatePolicy(ctx, leftover.policy, func(policy *cloudresourcemanager.Policy) bool {
			return removeMembersWhere(policy, leftover.drop)
		}); err != nil {
			return removed, errors.Wrapf(err, "removing lane bindings from %s", leftover.policy.name)
		}
	}
	return removed, nil
}

// secondProjectEnv names the second test project, for the kinds whose lanes
// target one apart from the harness's own (a project-singleton policy, a
// fleet). The sweep reads both projects too.
const secondProjectEnv = "PLANTON_E2E_GCP_SECOND_PROJECT_ID"

// organizationEnv and billingAccountEnv name the test organization and the
// test billing account, where a lane's identity is granted the groups that
// reach them. The sweep reads both too.
const (
	organizationEnv   = "PLANTON_E2E_GCP_ORGANIZATION_ID"
	billingAccountEnv = "PLANTON_E2E_GCP_BILLING_ACCOUNT_ID"
)

// testProjects is every project a lane's identity may be granted in: the
// harness's own, and the second test project when one is set.
func (h *Harness) testProjects() []string {
	projects := []string{h.services.Project}
	if second := os.Getenv(secondProjectEnv); second != "" && second != h.services.Project {
		projects = append(projects, second)
	}
	return projects
}

// identityProject is the project a lane's identity is granted in: the one
// the scenario names with provider.IdentityProjectAnnotation (tokens already
// expanded), else the harness's own. It must be a test project, so a typo
// never grants a lane identity somewhere CleanLaneIdentities would not look.
func (h *Harness) identityProject(tc *provider.KindTestContext) (string, error) {
	named, err := runner.ManifestAnnotation(tc.ManifestPath, provider.IdentityProjectAnnotation)
	if err != nil {
		return "", errors.Wrapf(err, "reading %s", provider.IdentityProjectAnnotation)
	}
	if named == "" {
		return h.services.Project, nil
	}
	for _, project := range h.testProjects() {
		if named == project {
			return project, nil
		}
	}
	return "", errors.Errorf("%s names %q, which is not a test project (%s)",
		provider.IdentityProjectAnnotation, named, strings.Join(h.testProjects(), ", "))
}

// isLaneOrDeletedMember names a policy member a lane left behind: a lane
// account, or any deleted principal Google keeps listed until its binding is
// removed.
func isLaneOrDeletedMember(member string) bool {
	return isLaneMember(member) || strings.HasPrefix(member, "deleted:")
}

// isLaneMember names a lane account, for a policy the harness shares with
// others (a billing account), where a deleted principal may be anyone's. A
// lane account deleted before its binding (a crashed lane's, which
// CleanLaneIdentities deletes first) is listed as deleted:serviceAccount:
// with the same lane prefix, which is the lane's alone.
func isLaneMember(member string) bool {
	return strings.HasPrefix(strings.TrimPrefix(member, "deleted:"), "serviceAccount:"+laneAccountPrefix)
}

// laneLabel names a lane in full: kind, engine, scenario and the lane's run
// id, the parts that make it unique among lanes running side by side.
func laneLabel(tc *provider.KindTestContext) string {
	return fmt.Sprintf("%s/%s/%s/%s", tc.Kind, tc.Engine, runner.ScenarioSlug(tc.ManifestPath), tc.RunID)
}

// laneAccountID is the lane's service account id: Google caps an id at 30
// characters, so it is a prefix and a hash of the lane's full label, which
// the account's display name carries readably.
func laneAccountID(lane string) string {
	sum := sha256.Sum256([]byte(lane))
	return laneAccountPrefix + hex.EncodeToString(sum[:])[:10]
}

// retryUntil calls attempt until it reports done, returns an error, or the
// budget runs out, waiting a few jittered seconds between calls.
func retryUntil(ctx context.Context, budget time.Duration, attempt func() (bool, error)) error {
	deadline := time.Now().Add(budget)
	for {
		done, err := attempt()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.Errorf("not done after %s", budget)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3*time.Second + time.Duration(rand.Intn(3000))*time.Millisecond):
		}
	}
}

// keyNotYetAccepted reports the refusals Google gives a service account key in
// the first seconds after it is minted, before the key has propagated: the
// token endpoint's OAuth error invalid_grant ("Invalid JWT Signature"), and a
// 401 or 403 from the API once a token is issued. Only the readiness wait reads
// these as "not yet"; anywhere else they are real refusals.
func keyNotYetAccepted(err error) bool {
	var refused *auth.Error
	if errors.As(err, &refused) {
		var body struct {
			Error string `json:"error"`
		}
		return json.Unmarshal(refused.Body, &body) == nil && body.Error == "invalid_grant"
	}
	return isStatus(err, 401) || isStatus(err, 403)
}

func isStatus(err error, code int) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == code
}

func containsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// lanef writes through the lane's own test log, so lanes running side by side
// keep their lines apart.
func lanef(tc *provider.KindTestContext, format string, args ...interface{}) {
	if tc.T == nil {
		fmt.Printf(format+"\n", args...)
		return
	}
	tc.T.Logf(format, args...)
}
