package runner

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// RunIDToken is the placeholder scenario and prerequisite manifests use for
// values that must be unique per test run. Some provider resources reserve their
// user-chosen identifier long after deletion (soft-delete retention windows —
// e.g. workload identity pools and KMS key rings), so a manifest that hardcodes
// such an identifier can never be deployed twice: the second run (even the
// second engine of the same run) collides with the soft-deleted ghost of the
// first. Embedding this token in the identifier field keeps the manifest a
// plain, reviewable YAML file while giving every run a fresh identifier.
const RunIDToken = "${E2E_RUN_ID}"

// UnderscoreRunIDToken is RunIDToken for identifier classes that forbid
// hyphens (letters/numbers/underscores only — e.g. a Vertex AI
// deployed_index_id or a BigQuery dataset id). The engine-scoped run id
// carries a hyphenated engine suffix ("-p"/"-t"), so the plain token would
// break such fields; this variant expands with every hyphen replaced by an
// underscore. Use it whenever a run-scoped identifier lives in an
// underscore-only field.
const UnderscoreRunIDToken = "${E2E_RUN_ID_UNDERSCORE}"

// ScenarioToken is the placeholder for the running scenario's slug (the
// scenario file's base name without extension). It is the reservation-window
// sibling of RunIDToken: prerequisite install manifests redeploy per
// scenario, so a run-scoped cloud-side name alone is still DESTROYED AND
// RECREATED at every scenario boundary within one engine run — and resource
// classes that reserve a deleted name for minutes (Firestore database IDs)
// collide the next scenario's chain with the previous scenario's ghost.
// Embedding the scenario slug alongside the run id gives every scenario its
// own cloud-side identifier, so no name is ever recreated within a run
// (first hit: a Firestore-index kind growing a second scenario whose chain
// reinstalls the same database override).
const ScenarioToken = "${E2E_SCENARIO}"

// TimeTokenPrefix opens a run-clock token: ${E2E_UNIX_TIME_PLUS:<offset>}
// expands to a Unix timestamp in seconds, the run's clock plus the offset.
// Some providers bound how far ahead a date may sit (Stripe refuses a
// promotion code's expiry or a tax registration's start more than five years
// out), so no literal date stays valid: a near one passes and the lane fails
// the day it does, a far one is refused. The offset is days and then a Go
// duration, either part optional ("30d", "365d", "365d1m", "90m"). The clock
// is the lane's: read once when the lane starts and passed to every expansion
// in it, so every manifest one lane expands -- the scenario, its
// prerequisites, its second act -- names the same instant, a second act that
// repeats the token changes nothing, and a lane that runs late in a long
// process still gets dates measured from its own start.
const TimeTokenPrefix = "${E2E_UNIX_TIME_PLUS:"

// timeTokenPattern matches run-clock tokens; the offset is captured and
// parsed by parseTimeOffset, so a malformed one fails loudly.
var timeTokenPattern = regexp.MustCompile(`\$\{E2E_UNIX_TIME_PLUS:([^}]*)\}`)

// parseTimeOffset reads a run-clock offset: an optional whole number of days
// ("<N>d") followed by an optional Go duration ("1m", "2h30m").
func parseTimeOffset(offset string) (time.Duration, error) {
	rest := strings.TrimSpace(offset)
	var total time.Duration
	if i := strings.Index(rest, "d"); i > 0 {
		if days, err := strconv.Atoi(rest[:i]); err == nil {
			total = time.Duration(days) * 24 * time.Hour
			rest = rest[i+1:]
		}
	}
	if rest != "" {
		d, err := time.ParseDuration(rest)
		if err != nil {
			return 0, errors.Errorf("run-clock offset %q is not days followed by a duration (like 30d or 365d1m)", offset)
		}
		total += d
	}
	if total <= 0 {
		return 0, errors.Errorf("run-clock offset %q must point into the future", offset)
	}
	return total, nil
}

// LaneClock is the instant a lane's run-clock tokens count from: now, to the
// second, in UTC.
func LaneClock() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

// expandRunClock replaces every run-clock token in text with its timestamp.
func expandRunClock(text string, clock time.Time) (string, error) {
	var offsetErr error
	expanded := timeTokenPattern.ReplaceAllStringFunc(text, func(token string) string {
		offset, err := parseTimeOffset(timeTokenPattern.FindStringSubmatch(token)[1])
		if err != nil {
			offsetErr = err
			return token
		}
		return strconv.FormatInt(clock.Add(offset).Unix(), 10)
	})
	return expanded, offsetErr
}

// EnvTokenPrefix restricts which environment variables scenario manifests may
// reference through ${E2E_ENV:...} tokens. Batched real-cluster lanes need
// batch-specific values a committed manifest cannot carry honestly — IRSA
// role ARNs, bucket names, queue URLs differ per test account, and hardcoding
// one account's identifiers would make the scenario deploy a lie everywhere
// else. The batch bootstrap exports them as PLANTON_E2E_-prefixed variables;
// the prefix keeps expansion opt-in and scoped (a manifest can never read
// arbitrary process environment like AWS_SECRET_ACCESS_KEY).
const EnvTokenPrefix = "PLANTON_E2E_"

// envTokenPattern matches ${E2E_ENV:PLANTON_E2E_*} occurrences. The variable
// name is captured; names outside EnvTokenPrefix never match and therefore
// fail expansion loudly via the residual-token check below.
var envTokenPattern = regexp.MustCompile(`\$\{E2E_ENV:(` + EnvTokenPrefix + `[A-Z0-9_]+)\}`)

// ExpandManifestTokens substitutes RunIDToken and ScenarioToken occurrences
// in the manifest, expands run-clock tokens from clock (the lane's, see
// LaneClock), expands ${E2E_ENV:PLANTON_E2E_*} environment tokens, and
// returns the path to the expanded copy (a temp file, so the source manifest
// is never modified). Manifests without tokens pass through untouched — the
// original path is returned and no file is written. scenario is the running
// scenario's slug ("" when there is none — a manifest carrying ScenarioToken
// then fails loudly rather than deploying the literal token).
//
// Expansion is a plain text substitution performed before the manifest is
// parsed, so the token can appear in any field. Only cloud-side identifier
// fields should carry it: metadata names must stay stable because prerequisite
// FK resolution and human debugging both key off them.
func ExpandManifestTokens(manifestPath, runID, scenario string, clock time.Time) (string, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read manifest %s for token expansion", manifestPath)
	}
	hasEnvTokens := strings.Contains(string(raw), "${E2E_ENV:")
	hasScenarioToken := strings.Contains(string(raw), ScenarioToken)
	hasTimeTokens := strings.Contains(string(raw), TimeTokenPrefix)
	if !strings.Contains(string(raw), RunIDToken) && !strings.Contains(string(raw), UnderscoreRunIDToken) && !hasEnvTokens && !hasScenarioToken && !hasTimeTokens {
		return manifestPath, nil
	}
	if (strings.Contains(string(raw), RunIDToken) || strings.Contains(string(raw), UnderscoreRunIDToken)) && runID == "" {
		return "", errors.Errorf("manifest %s uses %s but no run id was provided", manifestPath, RunIDToken)
	}
	if hasScenarioToken && scenario == "" {
		return "", errors.Errorf("manifest %s uses %s but no scenario is in scope", manifestPath, ScenarioToken)
	}

	// The underscore variant substitutes first: RunIDToken is not a textual
	// prefix of it (the closing brace differs), but explicit ordering keeps
	// the intent obvious.
	expanded := strings.ReplaceAll(string(raw), UnderscoreRunIDToken, strings.ReplaceAll(runID, "-", "_"))
	expanded = strings.ReplaceAll(expanded, RunIDToken, runID)
	expanded = strings.ReplaceAll(expanded, ScenarioToken, scenario)

	if hasTimeTokens {
		if expanded, err = expandRunClock(expanded, clock); err != nil {
			return "", errors.Wrapf(err, "manifest %s", manifestPath)
		}
	}

	if hasEnvTokens {
		var missing []string
		multiline := map[string]string{}
		expanded = envTokenPattern.ReplaceAllStringFunc(expanded, func(token string) string {
			name := envTokenPattern.FindStringSubmatch(token)[1]
			value := os.Getenv(name)
			if value == "" {
				missing = append(missing, name)
			}
			// A multi-line value (certificate/key PEM material) cannot be
			// spliced as plain text -- its second line would land outside
			// the field's YAML scalar and corrupt the document. Leave the
			// token for the position-aware pass below.
			if strings.Contains(value, "\n") {
				multiline[name] = value
				return token
			}
			return value
		})
		if len(missing) > 0 {
			return "", errors.Errorf(
				"manifest %s uses environment tokens whose variables are unset: %s (exported by the real-cluster batch bootstrap)",
				manifestPath, strings.Join(missing, ", "))
		}
		for name, value := range multiline {
			expanded, err = expandMultilineEnvToken(expanded, name, value)
			if err != nil {
				return "", errors.Wrapf(err, "manifest %s", manifestPath)
			}
		}
		// A residual ${E2E_ENV: token means the name fell outside the
		// allowed prefix or its syntax is malformed — never deploy it as
		// literal text.
		if idx := strings.Index(expanded, "${E2E_ENV:"); idx >= 0 {
			end := strings.IndexByte(expanded[idx:], '}')
			residual := expanded[idx:]
			if end >= 0 {
				residual = expanded[idx : idx+end+1]
			}
			return "", errors.Errorf(
				"manifest %s carries an invalid environment token %q: only %s-prefixed variables may be referenced",
				manifestPath, residual, EnvTokenPrefix)
		}
	}

	// A temp file (not next to the scenario, so discovery never picks it up)
	// that KEEPS the scenario's basename: verifier dispatch keys behavioral
	// variants off the scenario name in the manifest path, and a
	// random-only temp name would silently demote every token-carrying
	// behavioral scenario to its plain verifier.
	tmpDir, err := os.MkdirTemp("", "planton-e2e-expanded-*")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp dir for expanded manifest")
	}
	tmpPath := filepath.Join(tmpDir, filepath.Base(manifestPath))
	if err := os.WriteFile(tmpPath, []byte(expanded), 0o600); err != nil {
		return "", errors.Wrap(err, "failed to write expanded manifest")
	}
	return tmpPath, nil
}

// expandMultilineEnvToken substitutes one env token whose value spans multiple
// lines (PEM certificates and keys are the motivating class). Plain text
// splicing would corrupt the YAML document -- every line after the first
// would sit outside the field's scalar -- so a multi-line value is rendered
// as a double-quoted single-line YAML scalar, and ONLY when the token
// occupies a whole mapping value ("field: ${E2E_ENV:...}"). Any other
// placement (mid-string, block scalar, list item) has no safe mechanical
// rendering and fails loudly instead of writing an unparseable manifest.
func expandMultilineEnvToken(expanded, name, value string) (string, error) {
	token := "${E2E_ENV:" + name + "}"
	// strconv.Quote emits \n, \", \\ and \t escapes -- the exact subset YAML
	// double-quoted scalars share with Go string literals. Values here are
	// printable ASCII plus newlines (PEM), so no Go-only escape form (\xNN)
	// is ever produced for them.
	quoted := strconv.Quote(value)
	wholeValue := regexp.MustCompile(`^(\s*[^:#\s][^:]*:\s*)` + regexp.QuoteMeta(token) + `\s*$`)
	lines := strings.Split(expanded, "\n")
	for i, line := range lines {
		if m := wholeValue.FindStringSubmatch(line); m != nil {
			lines[i] = m[1] + quoted
		}
	}
	rejoined := strings.Join(lines, "\n")
	if strings.Contains(rejoined, token) {
		return "", errors.Errorf(
			"environment token %s carries a multi-line value, which is only supported when the token is a whole mapping value (\"field: %s\")",
			token, token)
	}
	return rejoined, nil
}

// ScenarioSlug derives the ScenarioToken expansion from a scenario manifest
// path: the base filename without extension. Expanded/resolved manifest
// copies keep their basename (the temp-file contract above), so the slug is
// stable across every copy of the scenario.
func ScenarioSlug(scenarioManifestPath string) string {
	if scenarioManifestPath == "" {
		return ""
	}
	base := filepath.Base(scenarioManifestPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// EngineScopedRunID derives the value substituted for RunIDToken from the test
// run's id and the engine executing the scenario. Both engines run the same
// scenario within one test invocation, so the run id alone is not unique enough:
// a soft-delete-reserving identifier created by the Pulumi run would collide
// with the Terraform run minutes later. The engine's first letter keeps the
// suffix short (identifier fields like workload identity pool ids cap at 32
// characters) while making each engine's expansion distinct.
func EngineScopedRunID(runID, engine string) string {
	if engine == "" {
		return runID
	}
	return runID + "-" + engine[:1]
}
