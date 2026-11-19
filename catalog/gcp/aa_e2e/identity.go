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
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
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
// How they are granted, as the platform renders a target's grants (its
// GcpRoleDocuments): the groups that name no predefined role become one
// project custom role, and each named predefined role is bound on the
// project. A group that reaches above the project (an organization, folder
// or billing account) fails the lane with the group named: the test
// projects hold no grant above themselves.
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
	lane := laneLabel(tc)
	accountID := laneAccountID(lane)

	var roles []string
	if len(grant.permissions) > 0 {
		role, err := h.ensureDeclaredRole(ctx, project, grant)
		if err != nil {
			return "", nil, err
		}
		roles = append(roles, role)
	}
	roles = append(roles, grant.predefinedRoles...)

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
		// Whatever the verdict: the binding, then the account (its key dies
		// with it). A failure is reported, never fatal -- the lane's verdict
		// is already in, and the next TestMain removes what is left.
		background := context.Background()
		if err := h.updateProjectPolicy(background, project, func(policy *cloudresourcemanager.Policy) bool {
			return removeMember(policy, member)
		}); err != nil {
			lanef(tc, "  [identity] WARN: removing %s's grants failed: %v", account.Email, err)
		}
		if _, err := h.services.Iam.Projects.ServiceAccounts.Delete(account.Name).Context(background).Do(); err != nil && !isStatus(err, 404) {
			lanef(tc, "  [identity] WARN: deleting %s failed: %v", account.Email, err)
		}
	}

	// A new account can be unknown to the policy service for a few seconds.
	if err := retryUntil(ctx, time.Minute, func() (bool, error) {
		err := h.updateProjectPolicy(ctx, project, func(policy *cloudresourcemanager.Policy) bool {
			return addMember(policy, member, roles)
		})
		if err != nil && isStatus(err, 400) && strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}
		return err == nil, err
	}); err != nil {
		cleanup()
		return "", nil, errors.Wrapf(err, "granting %s to the lane identity %s", strings.Join(roles, ", "), account.Email)
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

	if err := h.awaitGrants(ctx, project, keyJSON, grant.permissions); err != nil {
		cleanup()
		return "", nil, errors.Wrapf(err, "the lane identity %s", account.Email)
	}

	config, removeConfig, err := iacinputproviderconfig.BuildFromProto(
		&gcpprovider.GcpProviderConfig{ServiceAccountKey: string(keyJSON)}, catalogkind.CatalogProvider_gcp)
	if err != nil {
		cleanup()
		return "", nil, errors.Wrap(err, "writing the lane identity's provider configuration")
	}
	lanef(tc, "  [identity] %s holds %s on %s (%d permission(s) from permissions.yaml%s)",
		account.Email, strings.Join(roles, ", "), project, len(grant.permissions), grant.conditionalNote())
	held := make(map[string]bool, len(grant.permissions))
	for _, permission := range grant.permissions {
		held[permission] = true
	}
	firstAct := tc.ManifestPath
	h.mu.Lock()
	h.laneGrants[tc] = held
	h.laneAudits[firstAct] = &laneAudit{tc: tc, project: project, principal: account.Email, start: start}
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
// refused for a permission the lane's custom role holds is a grant still
// reaching the service's own permission cache (Discovery Engine, Dialogflow
// and Cloud Deploy were measured lagging a minute past the key's settle).
// Permissions granted through a named predefined role are not in the set, so
// their refusals fail the lane at once.
func (h *Harness) HeldPermissionRefused(tc *provider.KindTestContext, deployErr error) (string, bool) {
	match := refusedPermission.FindStringSubmatch(deployErr.Error())
	if match == nil {
		return "", false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return match[1], h.laneGrants[tc][match[1]]
}

// declaredGrantSet is what a lane's identity is granted.
type declaredGrantSet struct {
	kind            string
	permissions     []string
	predefinedRoles []string
	// conditionalPurposes are the conditional groups this scenario turned on;
	// they name the custom role, since two scenarios of one kind may differ.
	conditionalPurposes []string
}

func (g declaredGrantSet) conditionalNote() string {
	if len(g.conditionalPurposes) == 0 {
		return ""
	}
	return "; conditional: " + strings.Join(g.conditionalPurposes, ", ")
}

// declaredGrant reads the kind's manifest and the specs of the scenario's
// acts (the first act, then the second when it has one) and decides the
// lane's grant. It refuses, before anything is minted, a group that reaches
// above the project and a permission Google keeps out of custom roles
// (which the manifest must cover with a predefined role instead).
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

	grant := declaredGrantSet{kind: kind}
	seenPermission := map[string]bool{}
	seenRole := map[string]bool{}
	manifestPath := permissions.Path(repoRoot, "gcp", kindDir)
	for _, group := range manifest.GetSpec().GetGcp().GetGroups() {
		condition := group.GetCondition().GetSpecFieldSet()
		if !permissions.ConditionNeeded(condition, specs) {
			continue
		}
		if group.GetReach() != permissionsv1.PermissionReach_target {
			return declaredGrantSet{}, errors.Errorf("%s group %q reaches %s; the declared identity is granted on a test project only, so this group's lane needs a grant above the project that the harness does not hold",
				manifestPath, group.GetPurpose(), group.GetReach())
		}
		if condition != "" {
			grant.conditionalPurposes = append(grant.conditionalPurposes, group.GetPurpose())
		}
		if role := group.GetPredefinedRole(); role != "" {
			if !seenRole[role] {
				seenRole[role] = true
				grant.predefinedRoles = append(grant.predefinedRoles, role)
			}
			continue
		}
		for _, permission := range group.GetPermissions() {
			if inventory.CustomRoleUnsupported(permission) {
				return declaredGrantSet{}, errors.Errorf("%s group %q holds %s, which Google keeps out of custom roles: the group must name the predefined role that carries it",
					manifestPath, group.GetPurpose(), permission)
			}
			if !seenPermission[permission] {
				seenPermission[permission] = true
				grant.permissions = append(grant.permissions, permission)
			}
		}
	}
	sort.Strings(grant.permissions)
	sort.Strings(grant.conditionalPurposes)
	if len(grant.permissions) == 0 && len(grant.predefinedRoles) == 0 {
		return declaredGrantSet{}, errors.Errorf("%s grants nothing for this scenario, so there is no identity to declare", manifestPath)
	}
	return grant, nil
}

// declaredRoleID is the standing custom role a kind's lanes share:
// plantonE2eDeclared<Kind>, plus C and a short hash of the conditional groups
// a scenario turns on, so scenarios with different conditional groups never
// rewrite each other's role. It is keyed by what the role is for, never by
// its permissions, so a manifest fix rewrites the role in place.
func declaredRoleID(kind string, conditionalPurposes []string) string {
	id := declaredRolePrefix + kind
	if len(conditionalPurposes) > 0 {
		sum := sha256.Sum256([]byte(strings.Join(conditionalPurposes, ",")))
		id += "C" + hex.EncodeToString(sum[:])[:4]
	}
	return id
}

// ensureDeclaredRole writes the kind's standing custom role to exactly the
// grant's permissions and returns its name. It is never deleted: Google
// counts roles against a per-project cap and holds a deleted role's id, and a
// standing role rewritten before every lane costs neither.
func (h *Harness) ensureDeclaredRole(ctx context.Context, project string, grant declaredGrantSet) (string, error) {
	roleID := declaredRoleID(grant.kind, grant.conditionalPurposes)
	name := "projects/" + project + "/roles/" + roleID
	desired := &iam.Role{
		Title:               truncate("Planton e2e declared "+grant.kind, 100),
		Description:         truncate("Exactly what the catalog's permissions manifest for "+grant.kind+" grants; rewritten before every declared lane", 256),
		IncludedPermissions: grant.permissions,
		Stage:               "GA",
	}
	roles := h.services.Iam.Projects.Roles
	existing, err := roles.Get(name).Context(ctx).Do()
	switch {
	case err != nil && isStatus(err, 404):
		_, err = roles.Create("projects/"+project, &iam.CreateRoleRequest{RoleId: roleID, Role: desired}).Context(ctx).Do()
		if err == nil {
			return name, nil
		}
		// The other engine's lane of the same kind created it a moment ago.
		if !isStatus(err, 409) {
			return "", errors.Wrapf(err, "creating the declared role %s", roleID)
		}
	case err != nil:
		return "", errors.Wrapf(err, "reading the declared role %s", roleID)
	case existing.Deleted:
		if _, err := roles.Undelete(name, &iam.UndeleteRoleRequest{Etag: existing.Etag}).Context(ctx).Do(); err != nil {
			return "", errors.Wrapf(err, "undeleting the declared role %s", roleID)
		}
	}
	if _, err := roles.Patch(name, desired).UpdateMask("title,description,includedPermissions,stage").Context(ctx).Do(); err != nil {
		return "", errors.Wrapf(err, "writing the declared role %s", roleID)
	}
	return name, nil
}

// awaitGrants waits until Google answers, for the new account itself, that it
// holds every permission of its custom role, and has kept accepting its key
// for keySettleWindow: each attempt builds a fresh client, so each fetches a
// fresh token, and any refusal starts the window again. Some services check
// permissions through their own caches and can lag even this; a 403 in the
// lane for a permission the role holds is therefore re-run after a wait before
// anything is concluded (the e2e README's triage section).
func (h *Harness) awaitGrants(ctx context.Context, project string, keyJSON []byte, wanted []string) error {
	if len(wanted) == 0 {
		return nil
	}
	var missing []string
	var acceptedSince time.Time
	err := retryUntil(ctx, readinessBudget, func() (bool, error) {
		asLane, err := cloudresourcemanager.NewService(ctx,
			option.WithCredentialsJSON(keyJSON), option.WithQuotaProject(project))
		if err != nil {
			return false, errors.Wrap(err, "building a client that authenticates as the lane identity")
		}
		missing = missing[:0]
		for start := 0; start < len(wanted); start += 100 {
			end := start + 100
			if end > len(wanted) {
				end = len(wanted)
			}
			response, err := asLane.Projects.TestIamPermissions(project, &cloudresourcemanager.TestIamPermissionsRequest{
				Permissions: wanted[start:end],
			}).Context(ctx).Do()
			if err != nil {
				// A key minted a moment ago can be refused until it propagates.
				if keyNotYetAccepted(err) {
					missing = append(missing, wanted[start:end]...)
					continue
				}
				return false, err
			}
			held := map[string]bool{}
			for _, permission := range response.Permissions {
				held[permission] = true
			}
			for _, permission := range wanted[start:end] {
				if !held[permission] {
					missing = append(missing, permission)
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
			len(missing), readinessBudget, missing[0])
	}
	return err
}

// updateProjectPolicy is a read-modify-write of the project's IAM policy that
// retries when another writer got there first: lanes running side by side
// edit the one project policy concurrently.
func (h *Harness) updateProjectPolicy(ctx context.Context, project string, change func(*cloudresourcemanager.Policy) bool) error {
	return retryUntil(ctx, 2*time.Minute, func() (bool, error) {
		policy, err := h.services.Crm.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{
			Options: &cloudresourcemanager.GetPolicyOptions{RequestedPolicyVersion: 3},
		}).Context(ctx).Do()
		if err != nil {
			return false, err
		}
		if !change(policy) {
			return true, nil
		}
		policy.Version = 3
		_, err = h.services.Crm.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
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
// account in the harness's project, and, in every test project a lane may be
// granted in, every policy binding that names one or a deleted member. A
// killed test process (exit 137, a go test timeout) skips the deferred
// cleanup a lane registers, and the rule of one test process per provider
// account means that when a process starts, every lane account already
// present is such a leftover. Returns what it removed.
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
	for _, project := range h.testProjects() {
		if err := h.updateProjectPolicy(ctx, project, func(policy *cloudresourcemanager.Policy) bool {
			return removeMembersWhere(policy, isLaneOrDeletedMember)
		}); err != nil {
			return removed, errors.Wrapf(err, "removing lane bindings from %s", project)
		}
	}
	return removed, nil
}

// secondProjectEnv names the second test project, for the kinds whose lanes
// target one apart from the harness's own (a project-singleton policy, a
// fleet). The sweep reads both projects too.
const secondProjectEnv = "PLANTON_E2E_GCP_SECOND_PROJECT_ID"

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
	return strings.HasPrefix(member, "serviceAccount:"+laneAccountPrefix) || strings.HasPrefix(member, "deleted:")
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
