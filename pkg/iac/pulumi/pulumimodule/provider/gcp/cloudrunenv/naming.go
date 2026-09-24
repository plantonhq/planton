// Package cloudrunenv keeps the secret values a Cloud Run service's, job's, or
// worker pool's environment variables carry (their secret_value arm) in
// Secret Manager, so the revision or task template holds a reference to a
// secret the resource owns and never the value itself. It is shared by the
// GcpCloudRun, GcpCloudRunJob, and GcpCloudRunWorkerPool modules because the
// three kinds carry the same env shape and the same hazard: a platform that resolves secret references before the
// module runs hands it a plain value, and whatever field that value lands in
// is what every viewer of the resource reads.
//
// The trade-off it embodies: one Secret Manager secret per variable, not one
// per resource. Cloud Run reads a secret's whole payload into a variable (it
// cannot select a key inside a JSON secret), so a variable-per-secret shape
// is the only one the platform can reference directly, and it lets the grant
// be scoped to exactly the secrets one runtime identity needs.
//
// This file is the naming law, pure and unit-tested; secrets.go creates the
// resources. Not covered here: volumes that mount a secret (they name a
// Secret Manager secret the author already owns) and Cloud Functions, which
// model secrets differently.
package cloudrunenv

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// secretIDMaxLen is Secret Manager's limit on a secret id.
const secretIDMaxLen = 255

// Kinds name the Cloud Run resource a secret belongs to, and lead its id so a
// service, a job, and a worker pool with the same name in the same region
// never collide.
const (
	KindService    = "run"
	KindJob        = "runjob"
	KindWorkerPool = "runpool"
)

// Variable is one environment variable whose value the module stores.
type Variable struct {
	// ContainerIndex is the container's position in the spec; it addresses
	// the variable when the template is built.
	ContainerIndex int
	// Container is the container's name, empty for an unnamed container
	// (Cloud Run allows one unnamed container per resource).
	Container string
	// Name is the environment variable's name.
	Name string
	// Value is the resolved secret value; never logged, never part of a name.
	Value string
}

// Key addresses one variable inside the resource: the template builder looks
// up a stored secret by the container's position and the variable's name.
type Key struct {
	ContainerIndex int
	Name           string
}

// Placement is where the secrets live and whose they are.
type Placement struct {
	// Kind is KindService, KindJob, or KindWorkerPool.
	Kind string
	// Resource is the Cloud Run service, job, or worker pool name.
	Resource string
	// Region is the resource's region ("global" for a multi-region service).
	Region string
	// ReplicaRegions are the regions each secret replicates to: the regions
	// the resource actually serves from, so the value never leaves them.
	ReplicaRegions []string
	// Project is the GCP project; empty uses the provider's default project,
	// the same fallback the resource itself uses.
	Project string
	// RuntimeServiceAccount is the email of the identity the revisions or
	// tasks run as; empty means the project's Compute Engine default service
	// account, which Cloud Run falls back to.
	RuntimeServiceAccount string
	// Labels are applied to every secret so the store is attributable to the
	// resource that owns it.
	Labels map[string]string
}

// unsafeIDChars is everything Secret Manager refuses in a secret id.
var unsafeIDChars = regexp.MustCompile(`[^A-Za-z0-9_-]`)

// SecretID renders the id of the secret that holds one variable's value:
// "<kind>_<region>_<resource>_<container>_<variable>". The kind, region,
// resource and container are lowercase letters, digits and hyphens by their
// own validation rules, so the first four underscores split the id
// unambiguously; only the variable name, the last segment, may carry
// underscores. An unnamed container is "c<index>". A character Secret
// Manager refuses in a variable name ('.') becomes '-'.
func SecretID(placement Placement, variable Variable) string {
	container := variable.Container
	if container == "" {
		container = fmt.Sprintf("c%d", variable.ContainerIndex)
	}
	return strings.Join([]string{
		placement.Kind,
		placement.Region,
		placement.Resource,
		container,
		unsafeIDChars.ReplaceAllString(variable.Name, "-"),
	}, "_")
}

// SecretIDs names every variable's secret and refuses the two shapes the
// naming law cannot store: two variables whose ids coincide once a '.' has
// become '-', and an id longer than Secret Manager accepts. Each refusal says
// which variables collide and how to fix it.
func SecretIDs(placement Placement, variables []Variable) (map[Key]string, error) {
	ids := make(map[Key]string, len(variables))
	owners := map[string]Variable{}
	var problems []string
	for _, variable := range variables {
		id := SecretID(placement, variable)
		if len(id) > secretIDMaxLen {
			problems = append(problems, fmt.Sprintf(
				"variable %q in container %s: its Secret Manager id %q is %d characters, over the %d Secret Manager allows -- shorten the variable or container name",
				variable.Name, containerLabel(variable), id, len(id), secretIDMaxLen))
			continue
		}
		if earlier, taken := owners[id]; taken {
			problems = append(problems, fmt.Sprintf(
				"variables %q and %q in container %s would both be stored as Secret Manager secret %q (a '.' in a name becomes '-') -- rename one of them",
				earlier.Name, variable.Name, containerLabel(variable), id))
			continue
		}
		owners[id] = variable
		ids[Key{ContainerIndex: variable.ContainerIndex, Name: variable.Name}] = id
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, fmt.Errorf("secret values cannot be stored:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return ids, nil
}

// RuntimeMember is the IAM member the secretAccessor grant names: the
// resource's own runtime identity, or the Compute Engine default service
// account of the project numbered projectNumber when none is set.
func RuntimeMember(runtimeServiceAccount string, projectNumber string) string {
	if runtimeServiceAccount != "" {
		return "serviceAccount:" + runtimeServiceAccount
	}
	return fmt.Sprintf("serviceAccount:%s-compute@developer.gserviceaccount.com", projectNumber)
}

func containerLabel(variable Variable) string {
	if variable.Container != "" {
		return fmt.Sprintf("%q", variable.Container)
	}
	return fmt.Sprintf("#%d", variable.ContainerIndex+1)
}
