package module

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// document is one Kubernetes object as a JSON-compatible map, the form the
// Terraform twin's yamldecode produces and the classic ConfigGroup applies.
type document = map[string]interface{}

// documentSeparator splits a multi-document stream on `---` lines only: a
// definition's descriptions may contain "---" mid-line. The Terraform twin
// splits on the same boundary.
var documentSeparator = regexp.MustCompile(`(?m)^---[ \t]*$`)

// fetchTimeout bounds one download; the largest file is under 100 KB.
const fetchTimeout = 2 * time.Minute

// fetchAttempts is how many times a transient answer (a dropped connection,
// HTTP 429, a 5xx) is tried before the deploy is refused. The Terraform
// twin's http data sources retry the same answers (their retry block).
const fetchAttempts = 4

// absent explains a release file the host does not serve (an HTTP 4xx): what
// its absence means, and what to do about it. The caller knows which: a
// release Planton never published, or a published one missing a file.
type absent struct {
	meaning string
	next    string
}

// fetchText downloads one release file, retrying the answers that pass: a
// plan fetches every release file, so a host limiting a burst of plans must
// slow the deploy down, not fail it. A failure is explained in three parts,
// because the reader needs to know whether the network, the host or the
// release is at fault.
func fetchText(url string, missing absent) (string, error) {
	client := &http.Client{Timeout: fetchTimeout}
	var response *http.Response
	var err error
	for attempt := 1; ; attempt++ {
		response, err = client.Get(url)
		retryable := err != nil || response.StatusCode == http.StatusTooManyRequests ||
			(response.StatusCode >= 500 && response.StatusCode != http.StatusNotImplemented)
		if !retryable || attempt == fetchAttempts {
			break
		}
		wait := time.Duration(1<<(attempt-1)) * time.Second
		if response != nil {
			if seconds, parseErr := strconv.Atoi(response.Header.Get("Retry-After")); parseErr == nil && seconds > 0 && seconds <= 30 {
				wait = time.Duration(seconds) * time.Second
			}
			_ = response.Body.Close()
		}
		time.Sleep(wait)
	}
	if err != nil {
		return "", fmt.Errorf("observed: %s could not be fetched after %d attempts (%s)\nmeaning: the host does not resolve, or egress to it is blocked from where this deploy runs\nnext step: confirm DNS and egress to downloads.planton.ai (`curl -I %s`), then re-run", url, fetchAttempts, err.Error(), url)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", fetchRefusal(url, response.StatusCode, missing)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrapf(err, "reading %s", url)
	}
	return string(raw), nil
}

// fetchRefusal explains a release file the host would not serve, by what
// its answer means: a limit to wait out, a failing host, or a file that is
// not there.
func fetchRefusal(url string, status int, missing absent) error {
	switch {
	case status == http.StatusTooManyRequests:
		return fmt.Errorf("observed: %s answered HTTP 429 on all %d attempts\nmeaning: the host is limiting how often this address may download; a burst of plans from one address can reach the limit\nnext step: wait a minute and re-run", url, fetchAttempts)
	case status >= 500:
		return fmt.Errorf("observed: %s answered HTTP %d on all %d attempts\nmeaning: the host is failing, not the release\nnext step: re-run once it answers 200 (`curl -I %s`)", url, status, fetchAttempts, url)
	default:
		return fmt.Errorf("observed: %s answered HTTP %d\nmeaning: %s\nnext step: %s", url, status, missing.meaning, missing.next)
	}
}

// splitDocuments parses a stream into its objects, dropping fragments that
// carry no kind (licence headers, comment-only pieces).
func splitDocuments(stream string) ([]document, error) {
	var documents []document
	for _, piece := range documentSeparator.Split(stream, -1) {
		if strings.TrimSpace(piece) == "" {
			continue
		}
		var object document
		if err := yaml.Unmarshal([]byte(piece), &object); err != nil {
			return nil, errors.Wrap(err, "parsing a release manifest document")
		}
		if kind, _ := object["kind"].(string); kind == "" {
			continue
		}
		documents = append(documents, object)
	}
	return documents, nil
}

// documentID is an object's composed identity,
// apiVersion//kind//name[//namespace]: the Terraform twin's for_each key and
// the kubectl importer's ID form, so an import map derives every ID from the
// state address. Cluster-scoped objects have three parts.
func documentID(object document) string {
	parts := []string{str(object, "apiVersion"), str(object, "kind"), str(object, "metadata", "name")}
	if namespace := str(object, "metadata", "namespace"); namespace != "" {
		parts = append(parts, namespace)
	}
	return strings.Join(parts, "//")
}

// str reads a string at a path, "" when any step is missing.
func str(object document, path ...string) string {
	value, _ := lookup(object, path...).(string)
	return value
}

// lookup walks nested maps, nil when any step is missing.
func lookup(object document, path ...string) interface{} {
	var current interface{} = object
	for _, key := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = m[key]
	}
	return current
}

// child returns the map at key, creating it when absent.
func child(object document, key string) document {
	if m, ok := object[key].(map[string]interface{}); ok {
		return m
	}
	m := document{}
	object[key] = m
	return m
}

// podSpec returns a workload's pod template spec.
func podSpec(workload document) document {
	return child(child(child(workload, "spec"), "template"), "spec")
}

// namedItem returns the item of a list of maps whose key field equals value.
func namedItem(list interface{}, key, value string) document {
	items, _ := list.([]interface{})
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok && m[key] == value {
			return m
		}
	}
	return nil
}

// containerNamed returns the named container (or init container, with init
// true) of a pod spec, or refuses when the pinned release no longer has it.
func containerNamed(spec document, name string, init bool) (document, error) {
	field := "containers"
	if init {
		field = "initContainers"
	}
	if c := namedItem(spec[field], "name", name); c != nil {
		return c, nil
	}
	return nil, errors.Errorf("Agent Substrate %s's manifests have no %s named %q; the module's pin needs a catalog fix", vars.Release, strings.TrimSuffix(field, "s"), name)
}

// stringList converts a []interface{} of strings, as yaml parses arguments.
func stringList(value interface{}) []string {
	items, _ := value.([]interface{})
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func toInterfaces[T any](items []T) []interface{} {
	out := make([]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
