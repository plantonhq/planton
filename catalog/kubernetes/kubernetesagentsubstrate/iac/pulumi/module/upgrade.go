package module

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/plantonhq/planton/pkg/failure"
)

// Agent Substrate's upgrade policy (docs/upgrade.md at the pinned tag):
// before 1.0 every minor release is a reinstall, because the API server
// migrates its database schema irreversibly on its first start and the new
// minor's components are not built to run against the older minor's data;
// from 1.0, upgrades go one minor at a time and a rollback goes back at most
// one. Patch releases are always free. Upstream writes no version marker on
// the cluster and checks nothing, so the module stamps the installed release
// on the API server Deployment, and the minor that wrote the bundled
// database on its volume (which outlives a destroy while another resource
// owns the namespace), and refuses at plan time a move the policy forbids.
// The Terraform twin computes the same rule in preflight.tf.

// version is a parsed release, "v0.6.0", or "v0.6" for a minor stamp.
type version struct {
	major, minor int
}

func parseVersion(text string) (version, bool) {
	parts := strings.Split(strings.TrimPrefix(text, "v"), ".")
	if len(parts) < 2 {
		return version{}, false
	}
	major, errMajor := strconv.Atoi(parts[0])
	minor, errMinor := strconv.Atoi(parts[1])
	if errMajor != nil || errMinor != nil {
		return version{}, false
	}
	return version{major: major, minor: minor}, true
}

// The rules a forbidden move breaks, in the words the refusal uses.
const (
	ruleBeforeOne = "before 1.0 every minor release is a reinstall, because the API server migrates its database schema irreversibly on its first start, and a new minor's components are not built to run on an older minor's data"
	ruleMajor     = "a major release is a reinstall"
	ruleSkip      = "upgrades go one minor release at a time"
	ruleRollback  = "a rollback goes back at most one minor release, because the newer API server has already migrated the database schema forward"
)

// upgradeRule is the rule a move from installed to requested breaks, or ""
// when upstream supports it. An unreadable stamp names no rule: the module
// writes the stamps, so one it cannot read was not written by it.
func upgradeRule(installed, requested string) string {
	from, okFrom := parseVersion(installed)
	to, okTo := parseVersion(requested)
	switch {
	case !okFrom || !okTo || from == to:
		return ""
	case from.major != to.major:
		if from.major == 0 || to.major == 0 {
			return ruleBeforeOne
		}
		return ruleMajor
	case from.major == 0:
		return ruleBeforeOne
	case to.minor > from.minor+1:
		return ruleSkip
	case to.minor < from.minor-1:
		return ruleRollback
	}
	return ""
}

// deploymentUpgradeRefused refuses an apply over an installation whose
// stamp names a release the requested one cannot move from.
func deploymentUpgradeRefused(installed, requested, rule string) error {
	return failure.KubernetesUpgradeRefused(kindName, "Agent Substrate",
		fmt.Sprintf("Deployment %s/%s", vars.Namespace, vars.APIServerName), installed, requested, rule,
		fmt.Sprintf("to install %s, back up what you need from the database, destroy this resource, then clear the database it used (`kubectl -n %s delete pvc %s` for the bundled one, or drop the %s schema of an external or Cloud SQL one), replace every node that ran %s (its node agent's state outlives the destroy), and apply again; to keep %s, set spec.version back to it, or, for a release of another install shape, deploy this resource from the Planton release that installed it", requested, vars.Namespace, vars.PostgresVolumeClaim, "substrate", installed, installed))
}

// volumeUpgradeRefused refuses a bundled database whose volume an earlier
// installation's minor wrote: the API server would migrate or misread it.
func volumeUpgradeRefused(installed, requested, rule string) error {
	return failure.KubernetesUpgradeRefused(kindName, "Agent Substrate",
		fmt.Sprintf("PersistentVolumeClaim %s/%s, the bundled database's volume left by an earlier installation,", vars.Namespace, vars.PostgresVolumeClaim), installed, requested, rule,
		fmt.Sprintf("to install %s, back up what you need from that database, delete the volume (`kubectl -n %s delete pvc %s`), and apply again; to keep the data, set spec.version to a release of the minor that wrote it, or, for one of another install shape, deploy this resource from the Planton release that wrote it", requested, vars.Namespace, vars.PostgresVolumeClaim))
}
