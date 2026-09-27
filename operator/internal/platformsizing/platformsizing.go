// Package platformsizing judges a platform's declared sizing before any
// component runs -- the twin of platformversion, for the other thing a spec
// can declare that no workload can run with.
//
// Every size a component runs with is the registry's measured default with the
// spec's quantities laid over it (resources.Effective), so a problem often
// lives between the two halves: a memory limit declared below the default
// request, or a request raised above the default limit. Kubernetes would
// refuse such a pod only at admission, one workload at a time, in words about
// a Pod the person never wrote. Judged here, the whole platform is refused
// before anything is created or changed -- a running platform keeps running
// as it is -- and the reason names the field the person wrote and the fix.
//
// What it judges, and nothing else: each effective request against its limit,
// each floor a program in the workload's path enforces (the registry names
// the only ones), and the store's own ceiling (spec.database.redis.maxMemory)
// against its effective memory limit. It does not judge whether a size is
// enough: a real minimum shows itself as an out-of-memory kill, which the
// component's status names with the field to raise.
package platformsizing

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// Condition reasons for ResourcesValid.
const (
	// ReasonRunnable: every declared size is one its workload can run with.
	ReasonRunnable = "Runnable"
	// ReasonNotRunnable: at least one declared size is not; the message
	// names each field and its fix.
	ReasonNotRunnable = "NotRunnable"
)

// Verdict is the outcome of judging a spec's sizing.
type Verdict struct {
	// Runnable is true when every workload's effective size can run.
	Runnable bool
	// Reason is the stable condition reason for the outcome.
	Reason string
	// Message names every problem, one sentence each, in registry order;
	// on success, the one sentence ResourcesValid carries.
	Message string
}

// Check judges every registered workload's effective size.
func Check(spec *v1.PlantonPlatformSpec) Verdict {
	var problems []string
	for _, path := range resources.SizingPaths() {
		sizing := resources.ComponentSizing[path]
		declared := sizing.Declared(spec)
		effective := resources.Effective(path, declared)
		problems = append(problems, pairProblems(path, declared, effective)...)
		problems = append(problems, floorProblems(path, sizing.Floor, effective)...)
	}
	problems = append(problems, cacheProblem(spec)...)
	if len(problems) == 0 {
		return Verdict{Runnable: true, Reason: ReasonRunnable, Message: "every component's size is one its workload can run with"}
	}
	return Verdict{Reason: ReasonNotRunnable, Message: strings.Join(problems, "; ")}
}

// pairProblems finds each quantity whose effective request is above its
// effective limit, saying which half the person declared and which is the
// operator's default, so the fix is the half they did not write.
func pairProblems(path string, declared *v1.ComponentResources, effective corev1.ResourceRequirements) []string {
	var out []string
	for _, name := range sortedNames(effective.Requests) {
		request := effective.Requests[name]
		limit, limited := effective.Limits[name]
		if !limited || request.Cmp(limit) <= 0 {
			continue
		}
		wrote := declared.Requirements()
		_, requestDeclared := wrote.Requests[name]
		_, limitDeclared := wrote.Limits[name]
		switch {
		case requestDeclared && !limitDeclared:
			out = append(out, fmt.Sprintf("%s.requests.%s %s is above %s, the limit the operator uses by default; set limits.%s too, at least %s",
				path, name, request.String(), limit.String(), name, request.String()))
		case limitDeclared && !requestDeclared:
			out = append(out, fmt.Sprintf("%s.limits.%s %s is below %s, the request the operator uses by default; set requests.%s too, or raise the limit",
				path, name, limit.String(), request.String(), name))
		default:
			out = append(out, fmt.Sprintf("%s.requests.%s %s is above its limit %s; a container can never be granted more than its limit, so lower the request or raise the limit",
				path, name, request.String(), limit.String()))
		}
	}
	return out
}

// floorProblems finds each request below a floor the workload's own program
// enforces.
func floorProblems(path string, floor corev1.ResourceList, effective corev1.ResourceRequirements) []string {
	var out []string
	for _, name := range sortedNames(floor) {
		least := floor[name]
		if request, ok := effective.Requests[name]; ok && request.Cmp(least) < 0 {
			out = append(out, fmt.Sprintf("%s.requests.%s %s is below %s, the least this workload's chart will run with; set it to %s or more",
				path, name, request.String(), least.String(), least.String()))
		}
	}
	return out
}

// cacheProblem judges the store's own memory ceiling against its container:
// Valkey needs headroom above maxmemory for its append-only-file rewrite and
// client buffers, and a ceiling at or above the limit is an out-of-memory kill
// that reloads more data than it may hold on every restart.
func cacheProblem(spec *v1.PlantonPlatformSpec) []string {
	ceiling := resources.ValkeyDefaultMaxMemory
	if spec.Database != nil && spec.Database.Redis != nil && spec.Database.Redis.MaxMemory != "" {
		ceiling = spec.Database.Redis.MaxMemory
	}
	bytes, ok := valkeyBytes(ceiling)
	if !ok {
		return nil // the schema's pattern refuses what this cannot read
	}
	limits := resources.EffectiveFor(resources.SizingRedis, spec).Limits
	limit, limited := limits[corev1.ResourceMemory]
	if !limited || bytes < limit.Value() {
		return nil
	}
	return []string{fmt.Sprintf("spec.database.redis.maxMemory %s is not below the store's memory limit %s; Valkey needs headroom above its ceiling, so raise %s.limits.memory above it or lower maxMemory",
		ceiling, limit.String(), resources.SizingRedis)}
}

// valkeyBytes reads a size the way Valkey's configuration does: k, m and g are
// powers of 1000; kb, mb and gb are powers of 1024; a bare number or b is
// bytes.
func valkeyBytes(size string) (int64, bool) {
	units := []struct {
		suffix string
		factor int64
	}{{"kb", 1 << 10}, {"mb", 1 << 20}, {"gb", 1 << 30}, {"k", 1000}, {"m", 1000 * 1000}, {"g", 1000 * 1000 * 1000}, {"b", 1}}
	factor := int64(1)
	for _, u := range units {
		if strings.HasSuffix(size, u.suffix) {
			factor, size = u.factor, strings.TrimSuffix(size, u.suffix)
			break
		}
	}
	n, err := strconv.ParseInt(size, 10, 64)
	if err != nil || n < 0 {
		return 0, false
	}
	return n * factor, true
}

func sortedNames(list corev1.ResourceList) []corev1.ResourceName {
	names := make([]corev1.ResourceName, 0, len(list))
	for name := range list {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names
}
