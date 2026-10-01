// Package envsecrets is the one place GCP modules keep an environment
// variable's secret value in Secret Manager, so the resource the variable
// belongs to holds a reference to a secret the resource owns and never the
// value itself.
//
// Every runtime that takes environment variables carries the same hazard: a
// platform that resolves secret references before the module runs hands it a
// plain value, and whatever field that value lands in is what every viewer of
// the resource reads. The kinds that use this package give the secret a home
// field beside the plain one, and their modules hand that field's entries to
// Store. Cloud Run services, jobs, and worker pools, Cloud Functions, and
// Vertex AI Agent Engine then reference the stored version natively; Cloud
// Workflows and Cloud Composer, which have no secret field, receive the
// version's resource name as the variable and read the value through the
// Secret Manager API at run time.
//
// The trade-off it embodies: one Secret Manager secret per variable, not one
// per resource. The runtimes read a secret's whole payload into a variable
// (none can select a key inside a JSON secret), so a variable-per-secret
// shape is the only one they can reference directly, and it lets the grant
// be scoped to exactly the secrets one runtime identity needs.
//
// This file is the naming law, pure and unit-tested; secrets.go creates the
// resources. Not covered here: volumes that mount a secret (they name a
// Secret Manager secret the author already owns) and build-time variables,
// which no GCP build API reads from Secret Manager.
package envsecrets

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// secretIDMaxLen is Secret Manager's limit on a secret id.
const secretIDMaxLen = 255

// Kinds name the resource a secret belongs to, and lead its id so two kinds'
// resources with the same name in the same location never collide. A kind
// is lowercase letters only, so it never carries the underscore the id is
// split by.
const (
	KindService     = "run"
	KindJob         = "runjob"
	KindWorkerPool  = "runpool"
	KindFunction    = "function"
	KindAgentEngine = "agentengine"
	KindWorkflow    = "workflow"
	KindComposer    = "composer"
)

// Fallback identities: the email patterns, filled with the project number,
// of the identity a runtime runs as when its resource names none.
const (
	// ComputeDefaultAccount is the project's Compute Engine default service
	// account -- the fallback of Cloud Run, Cloud Functions, Cloud
	// Workflows, and Composer 2 nodes.
	ComputeDefaultAccount = "%s-compute@developer.gserviceaccount.com"
	// ReasoningEngineServiceAgent is the project's Vertex AI Reasoning
	// Engine service agent -- the fallback of Vertex AI Agent Engine.
	ReasoningEngineServiceAgent = "service-%s@gcp-sa-aiplatform-re.iam.gserviceaccount.com"
)

// Variable is one environment variable whose value the module stores.
type Variable struct {
	// ContainerIndex is the container's position in the spec; it addresses
	// the variable when the template is built. Zero when the placement has
	// no containers.
	ContainerIndex int
	// Container is the container's name, empty for an unnamed container
	// (Cloud Run allows one unnamed container per resource). Ignored when
	// the placement has no containers.
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
	// Kind is one of the Kind constants.
	Kind string
	// Resource is the GCP name of the resource the variables belong to, or
	// its Planton name when Google assigns the id (an Agent Engine). GCP
	// holds either to characters Secret Manager accepts in an id (the
	// Planton name rides a GCP label), so it needs no sanitizing.
	Resource string
	// Location is the resource's region ("global" for a multi-region Cloud
	// Run service).
	Location string
	// NoContainers marks a resource whose variables are not grouped by
	// container (a function, an agent, a workflow, an environment): the id
	// omits the container segment and every Key's ContainerIndex is zero.
	NoContainers bool
	// ReplicaRegions are the regions each secret replicates to: the regions
	// the resource actually serves from, so the value never leaves them.
	ReplicaRegions []string
	// Project is the GCP project; empty uses the provider's default project,
	// the same fallback the resource itself uses.
	Project string
	// RuntimeServiceAccount is the email of the identity the resource runs
	// as; empty means the fallback identity DefaultAccount names.
	RuntimeServiceAccount string
	// DefaultAccount is the fallback identity's email pattern, filled with
	// the project number; empty means ComputeDefaultAccount.
	DefaultAccount string
	// Labels are applied to every secret so the store is attributable to the
	// resource that owns it.
	Labels map[string]string
}

// unsafeIDChars is everything Secret Manager refuses in a secret id.
var unsafeIDChars = regexp.MustCompile(`[^A-Za-z0-9_-]`)

// SecretID renders the id of the secret that holds one variable's value:
// "<kind>_<location>_<resource>_<container>_<variable>", or
// "<kind>_<location>_<resource>_<variable>" for a placement without
// containers. The kind, location, and container are lowercase letters,
// digits and hyphens by their own validation rules, and so is the name of
// a Cloud Run, Cloud Functions, or Composer resource, so the leading
// underscores split those ids unambiguously and only the variable name, the
// last segment, may carry underscores. A workflow name or an agent's
// Planton name may carry underscores too, which can only make two such
// resources' ids coincide in one project; the second deploy then fails
// creating its secret instead of sharing one. An unnamed container is
// "c<index>". A character Secret Manager refuses in a variable name ('.')
// becomes '-'.
func SecretID(placement Placement, variable Variable) string {
	segments := []string{placement.Kind, placement.Location, placement.Resource}
	if !placement.NoContainers {
		container := variable.Container
		if container == "" {
			container = fmt.Sprintf("c%d", variable.ContainerIndex)
		}
		segments = append(segments, container)
	}
	return strings.Join(append(segments, unsafeIDChars.ReplaceAllString(variable.Name, "-")), "_")
}

// KeyOf addresses a variable the way Store's refs are keyed.
func KeyOf(placement Placement, variable Variable) Key {
	if placement.NoContainers {
		return Key{Name: variable.Name}
	}
	return Key{ContainerIndex: variable.ContainerIndex, Name: variable.Name}
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
			problems = append(problems, overlongProblem(placement, variable, id))
			continue
		}
		if earlier, taken := owners[id]; taken {
			problems = append(problems, collisionProblem(placement, earlier, variable, id))
			continue
		}
		owners[id] = variable
		ids[KeyOf(placement, variable)] = id
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, fmt.Errorf("secret values cannot be stored:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return ids, nil
}

// RuntimeMember is the IAM member the secretAccessor grant names: the
// resource's own runtime identity, or the placement's fallback identity in
// the project numbered projectNumber when none is set.
func RuntimeMember(placement Placement, projectNumber string) string {
	if placement.RuntimeServiceAccount != "" {
		return "serviceAccount:" + placement.RuntimeServiceAccount
	}
	pattern := placement.DefaultAccount
	if pattern == "" {
		pattern = ComputeDefaultAccount
	}
	return "serviceAccount:" + fmt.Sprintf(pattern, projectNumber)
}

func overlongProblem(placement Placement, variable Variable, id string) string {
	if placement.NoContainers {
		return fmt.Sprintf(
			"variable %q: its Secret Manager id %q is %d characters, over the %d Secret Manager allows -- shorten the variable or resource name",
			variable.Name, id, len(id), secretIDMaxLen)
	}
	return fmt.Sprintf(
		"variable %q in container %s: its Secret Manager id %q is %d characters, over the %d Secret Manager allows -- shorten the variable or container name",
		variable.Name, containerLabel(variable), id, len(id), secretIDMaxLen)
}

func collisionProblem(placement Placement, earlier, variable Variable, id string) string {
	if placement.NoContainers {
		return fmt.Sprintf(
			"variables %q and %q would both be stored as Secret Manager secret %q (a '.' in a name becomes '-') -- rename one of them",
			earlier.Name, variable.Name, id)
	}
	return fmt.Sprintf(
		"variables %q and %q in container %s would both be stored as Secret Manager secret %q (a '.' in a name becomes '-') -- rename one of them",
		earlier.Name, variable.Name, containerLabel(variable), id)
}

func containerLabel(variable Variable) string {
	if variable.Container != "" {
		return fmt.Sprintf("%q", variable.Container)
	}
	return fmt.Sprintf("#%d", variable.ContainerIndex+1)
}
