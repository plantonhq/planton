package resources

import (
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	v1 "github.com/plantonhq/planton/operator/api/v1"
)

// Every workload the platform runs is sized by the operator, and every size
// is one entry in ComponentSizing, keyed by the spec field that overrides it.
// The builders read the effective value through Effective; the sizing check
// (platformsizing), the status readback (status.components.<c>.sizing), the
// house-pattern test, and the generated api/v1/component_sizing.json all read
// the same entries -- so a default measured again is changed in one place and
// cannot drift from what the catalog and the console say it is.
//
// The house pattern for every default: CPU and memory requests so the pod
// schedules honestly and is never starved into failing its own probes, a
// memory limit so a leak or a runaway dataset is a restart instead of a node
// taken down, and never a CPU limit, so a cold start or a burst is not
// throttled into the slowness the probes then punish. An adopter may set a
// CPU limit (some clusters' policies require one); the operator never does.
//
// Deliberately NOT sized here, each for its reason:
//   - The cluster-shared operators -- CloudNativePG, the Barman Cloud plugin,
//     Tekton Pipelines. They are vendored manifests installed once per
//     cluster and shared by every platform on it (and by databases deployed
//     through Planton), so no one platform's spec may size them.
//   - One-shot init containers and Jobs: identity's ensure-database init
//     container, OpenFGA's migrate Job, and Temporal's schema Jobs, together
//     with Temporal's web UI and admin tools, which keep one small auxiliary
//     size (temporal.go). They run for seconds or carry no platform load,
//     and the OpenFGA chart has no key that would size its migrate Job.
//   - The operator's own pod: its chart sizes it (the KubernetesPlantonOperator
//     catalog kind's resources), because it runs before any platform exists.

// The spec field paths that size each workload -- the registry's keys, the
// status readback's paths, and the words every refusal names.
const (
	SizingControlPlane     = "spec.controlPlane.resources"
	SizingConsole          = "spec.console.resources"
	SizingRunner           = "spec.runner.resources"
	SizingGateway          = "spec.gateway.resources"
	SizingIdentity         = "spec.identity.resources"
	SizingPostgreSQL       = "spec.database.postgresql.resources"
	SizingRedis            = "spec.database.redis.resources"
	SizingOpenBAO          = "spec.vault.resources"
	SizingOpenFGA          = "spec.openfga.resources"
	SizingNeo4j            = "spec.components.graph.resources"
	SizingTemporalFrontend = "spec.temporal.frontend.resources"
	SizingTemporalHistory  = "spec.temporal.history.resources"
	SizingTemporalMatching = "spec.temporal.matching.resources"
	SizingTemporalWorker   = "spec.temporal.worker.resources"
)

// Sizing is one sized workload: the size every install gets, the least a
// program in its path will run with, and where the spec overrides it.
type Sizing struct {
	// Component is the component whose status reports this sizing (the
	// component's Name()).
	Component string

	// Default is the operator's measured size, in the house pattern.
	Default corev1.ResourceRequirements

	// Floor is the least the workload's REQUESTS may be, set only where a
	// program in its path enforces one (a chart that fails its own render
	// below it). No floor is guessed: a real minimum shows itself as an
	// out-of-memory kill, which the component's status names with this
	// entry's path.
	Floor corev1.ResourceList

	// Declared reads the spec's override, nil when none is declared.
	Declared func(spec *v1.PlantonPlatformSpec) *v1.ComponentResources
}

// ComponentSizing is the registry of every sized workload.
var ComponentSizing = map[string]Sizing{
	// The control plane. The JVM's heap is sized by the image's own
	// -XX:MaxRAMPercentage=60 from the container limit, so the limit IS the
	// heap rule (a limit alone never changes the heap silently): at 6Gi the
	// heap may take 3.6Gi and 2.4Gi is left for everything outside it.
	//
	// The limit is the heaviest measured parallel-deploy peak plus 25%,
	// rounded up. Each peak is the container's cgroup memory.peak across a
	// whole wave of the product's own promises on four workers, read at an 8Gi
	// limit -- so each bounds the need from above, since a smaller limit
	// shrinks the heap and collects sooner:
	//   - 3.04Gi: five stored-key connect-and-deploy cells
	//     (run 20260928-185127-self-hosted);
	//   - 4.52Gi: ten build, registry and deploy checks of the service path
	//     (run 20260928-074059-self-hosted), the heaviest.
	// A 4Gi limit was killed by three parallel deploys: with the heap capped
	// at 2.4Gi, more than 1.6Gi lived outside it. A change that grows the
	// peak is re-measured the same way before this number moves.
	//
	// The request stays at 1Gi so a node is not reserved for a burst. The
	// hosted product sizes its own service separately (with a CPU limit of
	// its own); nothing couples the two numbers.
	SizingControlPlane: {
		Component: "controlplane",
		Default:   houseSizing("250m", "1Gi", "6Gi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.ControlPlane == nil {
				return nil
			}
			return s.ControlPlane.Resources
		},
	},

	// The console (the same lesson as the database and the identity server):
	// without a request it can be starved on a busy node into failing its
	// own probes; without a memory limit it is OOM-killed confusingly.
	SizingConsole: {
		Component: "console",
		Default:   houseSizing("250m", "512Mi", "2Gi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Console == nil {
				return nil
			}
			return s.Console.Resources
		},
	},

	// The runner: ~780Mi resident live working its queues on a one-node
	// install -- a Go binary that forks the OpenTofu and Pulumi engines,
	// whose own memory counts against the pod. The limit is the headroom for
	// an engine run.
	SizingRunner: {
		Component: "runner",
		Default:   houseSizing("100m", "512Mi", "2Gi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Runner == nil {
				return nil
			}
			return s.Runner.Resources
		},
	},

	// The front door: nginx proxying one platform's traffic is a small,
	// steady workload; the limit keeps a buffer leak from taking the node.
	SizingGateway: {
		Component: "gateway",
		Default:   houseSizing("50m", "64Mi", "256Mi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Gateway == nil {
				return nil
			}
			return s.Gateway.Resources
		},
	},

	// The identity server (the data-layer OOM lesson): a JVM plus a
	// first-boot realm import dies confusingly under default limits. The
	// recovery Job runs the same image and takes the same size.
	SizingIdentity: {
		Component: "identity",
		Default:   houseSizing("250m", "512Mi", "1536Mi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Identity == nil {
				return nil
			}
			return s.Identity.Resources
		},
	},

	// The database, chosen rather than left to CloudNativePG's default
	// (none): one instance serving every platform database, ~520Mi resident
	// live with the control plane's pools open. The control plane's first
	// boot migrates every database at once, enough concurrent connections
	// and shared buffers to OOM-kill a tiny default.
	SizingPostgreSQL: {
		Component: "postgresql",
		Default:   houseSizing("250m", "512Mi", "2Gi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Database == nil || s.Database.PostgreSQL == nil {
				return nil
			}
			return s.Database.PostgreSQL.Resources
		},
	},

	// The redis-protocol store, chosen rather than left to the chart. The
	// chart's own default is its "nano" preset (a 192Mi limit its header
	// calls "for basic testing"), no maxmemory, and noeviction -- so a
	// dataset that outgrew the preset was OOM-killed and then reloaded more
	// persisted data than it may hold on every restart, a crash loop behind
	// a front door still answering 200. The numbers are the hosted
	// product's for the same store role, with the ceiling
	// (ValkeyDefaultMaxMemory) inside the limit so what is persisted always
	// reloads.
	SizingRedis: {
		Component: "redis",
		Default:   houseSizing("100m", "256Mi", "1Gi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Database == nil || s.Database.Redis == nil {
				return nil
			}
			return s.Database.Redis.Resources
		},
	},

	// The vault, chosen rather than left to the chart (which ships none): a
	// single-tenant vault serving one control plane is small and steady,
	// ~35Mi resident live.
	SizingOpenBAO: {
		Component: "openbao",
		Default:   houseSizing("50m", "128Mi", "512Mi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Vault == nil {
				return nil
			}
			return s.Vault.Resources
		},
	},

	// The authorization engine, chosen rather than left to the chart (which
	// ships none): a Go server answering one control plane's checks, ~30Mi
	// resident live.
	SizingOpenFGA: {
		Component: "openfga",
		Default:   houseSizing("50m", "64Mi", "256Mi"),
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.OpenFGA == nil {
				return nil
			}
			return s.OpenFGA.Resources
		},
	},

	// The graph database: the chart's own size (its values.yaml ships 1000m
	// and 2Gi), which the chart applied to both requests and limits -- the
	// house pattern keeps the memory limit and drops the CPU limit. The
	// chart refuses to render with requests under 500m CPU or 2Gi memory
	// (_helpers.tpl, neo4j.resources.evaluateCPU/evaluateMemory), so that is
	// the floor, judged in the operator's own words before the chart would
	// fail. The chart's arithmetic reads "2Gi" and "2G" alike; the floor is
	// the binary 2Gi, which never admits a value the chart refuses.
	SizingNeo4j: {
		Component: "neo4j",
		Default:   houseSizing("1000m", "2Gi", "2Gi"),
		Floor: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Declared: func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
			if s.Components == nil || s.Components.Graph == nil {
				return nil
			}
			return s.Components.Graph.Resources
		},
	},

	// Temporal's four server services, chosen rather than left to the chart
	// (which ships none). Read live on a one-node install: history ~375Mi (it
	// holds the mutable state and its caches), matching ~145Mi, frontend
	// ~90Mi, worker ~60Mi. One size for three of them with history above it.
	SizingTemporalFrontend: {
		Component: "temporal",
		Default:   houseSizing("50m", "128Mi", "512Mi"),
		Declared:  temporalService(func(t *v1.TemporalSpec) *v1.TemporalServiceSpec { return t.Frontend }),
	},
	SizingTemporalHistory: {
		Component: "temporal",
		Default:   houseSizing("100m", "256Mi", "1Gi"),
		Declared:  temporalService(func(t *v1.TemporalSpec) *v1.TemporalServiceSpec { return t.History }),
	},
	SizingTemporalMatching: {
		Component: "temporal",
		Default:   houseSizing("50m", "128Mi", "512Mi"),
		Declared:  temporalService(func(t *v1.TemporalSpec) *v1.TemporalServiceSpec { return t.Matching }),
	},
	SizingTemporalWorker: {
		Component: "temporal",
		Default:   houseSizing("50m", "128Mi", "512Mi"),
		Declared:  temporalService(func(t *v1.TemporalSpec) *v1.TemporalServiceSpec { return t.Worker }),
	},
}

// SizingPaths returns every registered path, sorted -- the one iteration
// order for everything that walks the registry, so a refusal lists its
// problems and the status its entries the same way every time.
func SizingPaths() []string {
	paths := make([]string, 0, len(ComponentSizing))
	for path := range ComponentSizing {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// Effective is the size a workload runs with: the registered default with
// every quantity the override declares laid over it, one quantity at a time.
// A quantity the override sets wins -- including one the default lacks, such
// as a CPU limit or ephemeral storage; a quantity it leaves unset keeps the
// default. A nil override is the default. The result is a copy the caller
// may keep. An unregistered path is a defect in the operator, never input.
func Effective(path string, override *v1.ComponentResources) corev1.ResourceRequirements {
	sizing, ok := ComponentSizing[path]
	if !ok {
		panic(fmt.Sprintf("resources: no sizing is registered for %s", path))
	}
	declared := override.Requirements()
	return corev1.ResourceRequirements{
		Requests: mergeQuantities(sizing.Default.Requests, declared.Requests),
		Limits:   mergeQuantities(sizing.Default.Limits, declared.Limits),
	}
}

// EffectiveFor is Effective with the override read from the spec.
func EffectiveFor(path string, spec *v1.PlantonPlatformSpec) corev1.ResourceRequirements {
	sizing, ok := ComponentSizing[path]
	if !ok {
		panic(fmt.Sprintf("resources: no sizing is registered for %s", path))
	}
	return Effective(path, sizing.Declared(spec))
}

// houseSizing is a default in the house pattern: CPU and memory requests, a
// memory limit, no CPU limit.
func houseSizing(cpuRequest, memoryRequest, memoryLimit string) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuRequest),
			corev1.ResourceMemory: resource.MustParse(memoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(memoryLimit),
		},
	}
}

// temporalService reads one Temporal service's override from the spec.
func temporalService(service func(*v1.TemporalSpec) *v1.TemporalServiceSpec) func(*v1.PlantonPlatformSpec) *v1.ComponentResources {
	return func(s *v1.PlantonPlatformSpec) *v1.ComponentResources {
		if s.Temporal == nil {
			return nil
		}
		if svc := service(s.Temporal); svc != nil {
			return svc.Resources
		}
		return nil
	}
}

// mergeQuantities lays declared over defaults, quantity by quantity, into a
// fresh list. Nil when both are empty, so an unsized half stays absent.
func mergeQuantities(defaults, declared corev1.ResourceList) corev1.ResourceList {
	if len(defaults) == 0 && len(declared) == 0 {
		return nil
	}
	out := make(corev1.ResourceList, len(defaults)+len(declared))
	for name, q := range defaults {
		out[name] = q.DeepCopy()
	}
	for name, q := range declared {
		out[name] = q.DeepCopy()
	}
	return out
}

// mustBeSized is how every builder takes the sizing it is handed. Every
// registered workload has a default, so a builder handed an empty size means
// its component never resolved one -- a defect in the operator that would
// otherwise ship a pod with no requests and no limit, exactly the unsized
// workload the registry exists to rule out. It fails the way an unregistered
// path does: loudly, in the first test that renders the workload.
func mustBeSized(path string, r corev1.ResourceRequirements) corev1.ResourceRequirements {
	if len(r.Requests) == 0 && len(r.Limits) == 0 {
		panic(fmt.Sprintf("resources: the %s workload was handed no sizing; its component must pass resources.EffectiveFor(%q, spec)", path, path))
	}
	return r
}

// SizingStatusFor is what status.components.<component>.sizing reports: the
// effective size of each workload the component runs, by its spec path. It is
// the readback that shows what is really running -- including when a field
// was declared against an older operator's definition, which the API server
// drops without a word and this list then shows as the default.
func SizingStatusFor(component string, spec *v1.PlantonPlatformSpec) []v1.ComponentSizingStatus {
	var out []v1.ComponentSizingStatus
	for _, path := range SizingPaths() {
		if ComponentSizing[path].Component != component {
			continue
		}
		effective := EffectiveFor(path, spec)
		out = append(out, v1.ComponentSizingStatus{Path: path, Requests: effective.Requests, Limits: effective.Limits})
	}
	return out
}
