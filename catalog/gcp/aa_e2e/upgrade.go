package aa_e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
)

// The judgment of a declared lane's second act: did the upgrade change the
// kind's objects in place, or replace them? The catalog's `proven` means an
// identity holding exactly the kind's permissions deployed, UPDATED and
// destroyed it; a replacement proves create and delete again and never the
// update permissions, so a second act must say which it is
// (provider.ExpectUpgradeAnnotation) and the harness checks it against
// Google's own record.
//
// The record is the Admin Activity audit log of the project the lane's
// identity is granted in, read for the lane's account alone. Fixtures deploy
// as the harness and both engines deploy as the lane's key, so every entry
// there is the kind under test. The rule needs no per-kind knowledge:
//
//   - the kind's objects are the resource names the lane's account CREATED
//     during the first act (from the identity to VERIFY-RES);
//   - during the upgrade (from there to VERIFY-UPGRADED), any of those names
//     deleted or created again means "replaced";
//   - any other write (a patch, a set, addRule, attachNetworkEndpoints,
//     updatePeering) means "in place".
//
// Silence is never a verdict: an upgrade Google recorded no change for (a
// second act whose field the module never sends) fails as not judged.
// Kinds whose create, update and destroy are one call (a project-singleton
// policy) have no created name to lose, so any write reads in place, which is
// true of a singleton; kinds whose every field replaces (an IAM member) carry
// no second act and say so on their profiles.

// upgradeUnauditedAnnotation, on a GCP second act, names why Google keeps no
// Admin Activity entry for the kind -- measured, never assumed: Security
// Command Center writes none for notification configs, not even their
// create, while it logs its BigQuery exports. The upgrade is then not judged
// from the log; the engine's own plan (an update in place) and the
// verifier's read of the upgraded object stand, and the lane says so.
const upgradeUnauditedAnnotation = "planton.dev/e2e-upgrade-unaudited"

const (
	// upgradeJudgmentBudget bounds the wait for the upgrade's first entry:
	// Admin Activity entries usually land within a minute.
	upgradeJudgmentBudget = 3 * time.Minute
	// upgradeJudgmentPoll is the read interval: three lanes side by side stay
	// well inside Logging's per-project read quota.
	upgradeJudgmentPoll = 15 * time.Second
	// upgradeSettle is how long the harness keeps reading after the first
	// entry, so an entry that lands late (a delete behind a create) counts.
	upgradeSettle = 60 * time.Second
)

// laneAudit is what the judgment needs about one declared lane, recorded
// when its identity is provisioned and keyed by the first act's manifest
// (the path the runner later passes as provider.FirstActManifestPathKey).
type laneAudit struct {
	tc        *provider.KindTestContext
	project   string
	principal string
	start     time.Time
	// firstActDone is when the first act's last verification ran (VERIFY-RES,
	// or VERIFY-RECOVERED after an out-of-band act); zero until then.
	firstActDone time.Time
}

// requireUpgradeExpectation refuses, before anything is minted, a declared
// lane's second act that does not say what the upgrade does.
func requireUpgradeExpectation(secondAct string) error {
	expectation, err := runner.ManifestAnnotation(secondAct, provider.ExpectUpgradeAnnotation)
	if err != nil {
		return errors.Wrapf(err, "reading %s on the second act", provider.ExpectUpgradeAnnotation)
	}
	switch expectation {
	case provider.UpgradeInPlace, provider.UpgradeReplaced:
		return nil
	case "":
		return errors.Errorf("the second act carries no %s; a declared lane's upgrade must say whether it changes the kind in place (%q) or replaces it (%q), and the harness judges it from the audit log",
			provider.ExpectUpgradeAnnotation, provider.UpgradeInPlace, provider.UpgradeReplaced)
	default:
		return errors.Errorf("%s is %q; it takes %q or %q", provider.ExpectUpgradeAnnotation, expectation, provider.UpgradeInPlace, provider.UpgradeReplaced)
	}
}

// noteVerified is called from VerifyDeployed: on a first act it marks the end
// of the first act's window; on a second act it judges the upgrade.
func (h *Harness) noteVerified(ctx context.Context) error {
	firstAct, _ := ctx.Value(provider.FirstActManifestPathKey{}).(string)
	if firstAct == "" {
		manifest, _ := ctx.Value(provider.ManifestPathKey{}).(string)
		h.mu.Lock()
		// The latest first-act verification closes the window: an
		// out-of-band act's recovery, verified after VERIFY-RES, re-creates
		// the object before the upgrade starts.
		if audit := h.laneAudits[manifest]; audit != nil {
			audit.firstActDone = time.Now()
		}
		h.mu.Unlock()
		return nil
	}
	h.mu.Lock()
	audit := h.laneAudits[firstAct]
	h.mu.Unlock()
	if audit == nil || audit.firstActDone.IsZero() {
		// A lane without a declared identity has no account to read the log
		// for; its upgrade is not judged.
		return nil
	}
	secondAct, _ := ctx.Value(provider.ManifestPathKey{}).(string)
	expectation, _ := runner.ManifestAnnotation(secondAct, provider.ExpectUpgradeAnnotation)
	if unaudited, _ := runner.ManifestAnnotation(secondAct, upgradeUnauditedAnnotation); unaudited != "" {
		lanef(audit.tc, "  [upgrade] not judged from the audit log (%s): the engine planned an update in place and the verifier read the upgraded object", unaudited)
		return nil
	}
	verdict, evidence, err := h.judgeUpgrade(ctx, *audit, time.Now())
	if err != nil {
		return err
	}
	if verdict != expectation {
		return errors.Errorf("the second act declares %s %q, but Google's audit log shows the upgrade %s: %s",
			provider.ExpectUpgradeAnnotation, expectation, verdict, evidence)
	}
	lanef(audit.tc, "  [upgrade] %s, as declared -- Google's audit log in %s: %s", verdict, audit.project, evidence)
	return nil
}

// auditEntry is the part of an Admin Activity entry the judgment reads.
type auditEntry struct {
	at       time.Time
	method   string
	resource string
}

// judgeUpgrade reads the lane's entries and returns the verdict
// (provider.UpgradeInPlace or provider.UpgradeReplaced) and the entries
// behind it.
func (h *Harness) judgeUpgrade(ctx context.Context, audit laneAudit, upgradeDone time.Time) (string, string, error) {
	deadline := time.Now().Add(upgradeJudgmentBudget)
	var settleUntil time.Time
	for {
		entries, err := h.laneAuditEntries(ctx, audit)
		if err != nil {
			return "", "", err
		}
		created, upgrade := splitAuditWindows(entries, audit.firstActDone, upgradeDone)
		now := time.Now()
		if len(upgrade) > 0 && settleUntil.IsZero() {
			settleUntil = now.Add(upgradeSettle)
		}
		if !settleUntil.IsZero() && now.After(settleUntil) {
			verdict, evidence := upgradeVerdict(created, upgrade)
			return verdict, evidence, nil
		}
		if settleUntil.IsZero() && now.After(deadline) {
			return "", "", errors.Errorf("update not judged: Google's audit log in %s holds no change by %s during the upgrade after %s -- the second act's change may never reach Google (a field the module does not send)",
				audit.project, audit.principal, upgradeJudgmentBudget)
		}
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case <-time.After(upgradeJudgmentPoll):
		}
	}
}

// laneAuditEntries lists the lane account's Admin Activity entries since the
// lane began, oldest first.
func (h *Harness) laneAuditEntries(ctx context.Context, audit laneAudit) ([]auditEntry, error) {
	service, err := logging.NewService(ctx, option.WithQuotaProject(audit.project))
	if err != nil {
		return nil, errors.Wrap(err, "building the Logging client")
	}
	filter := fmt.Sprintf(`logName="projects/%s/logs/cloudaudit.googleapis.com%%2Factivity" AND protoPayload.authenticationInfo.principalEmail="%s" AND timestamp>="%s"`,
		audit.project, audit.principal, audit.start.UTC().Format(time.RFC3339))
	var entries []auditEntry
	err = service.Entries.List(&logging.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + audit.project},
		Filter:        filter,
		OrderBy:       "timestamp asc",
		PageSize:      1000,
	}).Pages(ctx, func(page *logging.ListLogEntriesResponse) error {
		for _, entry := range page.Entries {
			var payload struct {
				MethodName   string `json:"methodName"`
				ResourceName string `json:"resourceName"`
			}
			if err := json.Unmarshal(entry.ProtoPayload, &payload); err != nil {
				return errors.Wrap(err, "reading an audit entry")
			}
			at, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
			if err != nil {
				return errors.Wrapf(err, "reading an audit entry's time %q", entry.Timestamp)
			}
			entries = append(entries, auditEntry{at: at, method: payload.MethodName, resource: payload.ResourceName})
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "reading the Admin Activity audit log of %s", audit.project)
	}
	return entries, nil
}

// splitAuditWindows returns the names the first act created and the
// upgrade's entries.
func splitAuditWindows(entries []auditEntry, firstActDone, upgradeDone time.Time) (map[string]bool, []auditEntry) {
	created := map[string]bool{}
	var upgrade []auditEntry
	for _, entry := range entries {
		switch {
		case entry.at.Before(firstActDone):
			if auditMethodCreates(entry.method) && entry.resource != "" {
				created[entry.resource] = true
			}
		case !entry.at.After(upgradeDone):
			upgrade = append(upgrade, entry)
		}
	}
	return created, upgrade
}

// upgradeVerdict applies the rule to the upgrade's entries.
func upgradeVerdict(created map[string]bool, upgrade []auditEntry) (string, string) {
	var replaced, changed []string
	for _, entry := range upgrade {
		line := entry.method + " " + entry.resource
		if created[entry.resource] && (auditMethodCreates(entry.method) || auditMethodDeletes(entry.method)) {
			replaced = append(replaced, line)
			continue
		}
		changed = append(changed, line)
	}
	if len(replaced) > 0 {
		return provider.UpgradeReplaced, strings.Join(dedupe(replaced), "; ")
	}
	return provider.UpgradeInPlace, strings.Join(dedupe(changed), "; ")
}

// auditMethodCreates and auditMethodDeletes read a method's verb from its
// last segment, which every Google API names the same way: Compute's
// v1.compute.networks.insert, Discovery Engine's
// google.cloud.discoveryengine.v1.EngineService.CreateEngine, IAM's
// google.iam.admin.v1.DeleteRole.
func auditMethodCreates(method string) bool {
	verb := methodVerb(method)
	return strings.HasPrefix(verb, "create") || strings.HasPrefix(verb, "insert")
}

func auditMethodDeletes(method string) bool {
	return strings.HasPrefix(methodVerb(method), "delete")
}

func methodVerb(method string) string {
	return strings.ToLower(method[strings.LastIndex(method, ".")+1:])
}

func dedupe(lines []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			out = append(out, line)
		}
	}
	sort.Strings(out)
	return out
}
