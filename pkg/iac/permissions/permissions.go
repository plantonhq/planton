// Package permissions loads and validates per-kind runner permission
// manifests -- the catalog/<provider>/<kind>/iac/permissions.yaml sidecars
// declaring the least-privilege permissions the IaC runner's cloud
// principal needs for the PLATFORM modules. Every entry carries provenance
// (derived from module static analysis, or proven by live-run capture);
// the distinction is a trust feature consumers display. Enrollment is the
// file's presence: every discovered permissions.yaml is held to this
// package's conformance gate, with no allowlist to keep in sync.
package permissions

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	permissionsv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// FileName is the sidecar's name inside a kind's iac/ directory.
const FileName = "permissions.yaml"

// Path is a kind's permissions manifest location.
func Path(repoRoot, provider, kindDir string) string {
	return filepath.Join(repoRoot, "catalog", provider, kindDir, "iac", FileName)
}

// Discover returns provider -> kinds that ship a permissions manifest,
// sorted.
func Discover(repoRoot string) (map[string][]string, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, "catalog", "*", "*", "iac", FileName))
	if err != nil {
		return nil, err
	}
	discovered := map[string][]string{}
	for _, m := range matches {
		kindDir := filepath.Base(filepath.Dir(filepath.Dir(m)))
		provider := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(m))))
		discovered[provider] = append(discovered[provider], kindDir)
	}
	for _, kindDirs := range discovered {
		sort.Strings(kindDirs)
	}
	return discovered, nil
}

// Load reads and strictly parses a kind's permissions manifest.
func Load(repoRoot, provider, kindDir string) (*permissionsv1.CatalogKindPermissions, error) {
	manifest := &permissionsv1.CatalogKindPermissions{}
	if err := protobufyaml.Load(Path(repoRoot, provider, kindDir), manifest); err != nil {
		return nil, errors.Wrapf(err, "loading permissions for %s/%s", provider, kindDir)
	}
	return manifest, nil
}

// GcpRolesGrantedAboveAProject are predefined roles Google grants above a
// project only: each is absent from a project's grantable roles
// (gcloud iam list-grantable-roles //cloudresourcemanager.googleapis.com/projects/<id>,
// checked 2026-10-21). A group whose permissions only such a role can carry is
// granted where the role is -- reach organization (or billing_account) --
// even though Google answers the permissions themselves on a project, because
// a connection's setup binds a target-reach role on the project, where Google
// refuses it.
var GcpRolesGrantedAboveAProject = map[string]bool{
	"roles/iam.denyAdmin":                    true,
	"roles/iam.denyReviewer":                 true,
	"roles/orgpolicy.policyAdmin":            true,
	"roles/accesscontextmanager.policyAdmin": true,
	"roles/resourcemanager.folderAdmin":      true,
	"roles/billing.admin":                    true,
}

// GcpPermissionsGrantedAboveAProject are permissions Google authorizes above
// a project only, although its testable-permissions inventory also lists them
// on a project. A group holding one is granted where Google checks it --
// reach organization -- because a connection's project role can hold the
// permission and still never pass Google's check. Each entry says why; the
// list holds the same review as GcpRolesGrantedAboveAProject (reviewed
// 2026-11-22 against Google's hierarchical firewall policy and Shared VPC
// documentation).
var GcpPermissionsGrantedAboveAProject = map[string]bool{
	// A hierarchical firewall policy is a resource of the organization or of
	// a folder, never of a project: Google checks every compute.firewallPolicies
	// permission on the policy's parent (roles/compute.orgFirewallPolicyAdmin
	// and roles/compute.orgFirewallPolicyUser are granted on the organization
	// or a folder), so a project grant never reaches the policy.
	"compute.firewallPolicies.create": true,
	"compute.firewallPolicies.get":    true,
	"compute.firewallPolicies.update": true,
	"compute.firewallPolicies.delete": true,
	"compute.firewallPolicies.use":    true,
	// Associating a hierarchical policy with the organization or a folder is
	// checked on that node (roles/compute.orgSecurityResourceAdmin, granted on
	// the organization or a folder).
	"compute.organizations.setFirewallPolicy": true,
	// Shared VPC administration: Google checks the Xpn permissions on the
	// organization or on a folder above both the host and the service project
	// (roles/compute.xpnAdmin is granted there, never on a project), so a
	// project grant cannot enable a host or attach a service project.
	"compute.organizations.enableXpnHost":      true,
	"compute.organizations.disableXpnHost":     true,
	"compute.organizations.enableXpnResource":  true,
	"compute.organizations.disableXpnResource": true,
}

// GcpKindsAboveAProject are the GCP kinds whose resource never lives in a
// project: a folder is a child of the organization or of another folder, a
// hierarchical firewall policy is owned by the organization or a folder, and
// a custom organization policy constraint is defined on the organization.
// Every group of such a kind may be granted above the project (reach
// organization), whatever the inventory's narrowest scope for its
// permissions. A list rather than a property of the manifest: GcpProject also
// carries an unconditional organization group, yet its resource is a project.
var GcpKindsAboveAProject = map[string]bool{
	"GcpFolder":                     true,
	"GcpHierarchicalFirewallPolicy": true,
	"GcpOrgPolicyCustomConstraint":  true,
}

// GcpKindLivesAboveAProject reports whether the kind whose manifest lives in
// kindDir (its lowercase catalog directory, e.g. "gcpfolder") is on
// GcpKindsAboveAProject.
func GcpKindLivesAboveAProject(kindDir string) bool {
	for kind := range GcpKindsAboveAProject {
		if strings.ToLower(kind) == kindDir {
			return true
		}
	}
	return false
}

// GcpKindsJoiningTwoProjects are the GCP kinds whose one resource always
// joins two projects, so every call it makes reaches a project other than
// the one a connection points at. A connection's project role is bound on its
// own project only; the one grant that reaches both projects is the
// organization's. Any group of such a kind may therefore be granted at reach
// organization, whatever the inventory's narrowest scope for its permissions.
// A list rather than a property of the manifest, holding only kinds that
// join two projects by their nature: a kind that merely may point at another
// project (a VPC peering, a DNS zone's peering network, an HA VPN peer) is a
// one-project kind like any other. Each entry says why (reviewed 2026-11-26
// against the modules and Google's Shared VPC documentation).
var GcpKindsJoiningTwoProjects = map[string]bool{
	// Attaching a service project to a Shared VPC host: the provider reads
	// both projects through Resource Manager before it attaches, reads the
	// host's attachments with compute.projects.get, and polls the host's
	// operation; the host and the service project are always two projects.
	"GcpSharedVpcServiceProject": true,
}

// GcpKindJoinsTwoProjects reports whether the kind whose manifest lives in
// kindDir (its lowercase catalog directory) is on GcpKindsJoiningTwoProjects.
func GcpKindJoinsTwoProjects(kindDir string) bool {
	for kind := range GcpKindsJoiningTwoProjects {
		if strings.ToLower(kind) == kindDir {
			return true
		}
	}
	return false
}

// GcpScopeConditionReach is the reach a group's condition implies when it
// names the scope the shape lives at: a spec path ending in organization_id
// or folder_id is a shape Google authorizes on the organization or folder, a
// path ending in billing_account one it authorizes on the billing account.
// ok is false for any other path (or none).
func GcpScopeConditionReach(specFieldSet string) (reach permissionsv1.PermissionReach, ok bool) {
	last := specFieldSet
	if i := strings.LastIndex(specFieldSet, "."); i >= 0 {
		last = specFieldSet[i+1:]
	}
	switch last {
	case "organization_id", "folder_id":
		return permissionsv1.PermissionReach_organization, true
	case "billing_account":
		return permissionsv1.PermissionReach_billing_account, true
	}
	return permissionsv1.PermissionReach_permission_reach_unspecified, false
}
