package module

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
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

// fetchTimeout bounds one download; the release asset is under half a
// megabyte.
const fetchTimeout = 2 * time.Minute

// fetchText downloads one pinned upstream file. A failure is explained in
// three parts: these URLs are module constants, so nothing in the spec can
// fix them, and the reader needs to know whether the network or the pin is
// at fault.
func fetchText(url string) (string, error) {
	client := &http.Client{Timeout: fetchTimeout}
	response, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("observed: %s could not be fetched (%s)\nmeaning: the host does not resolve, or egress to it is blocked from where this deploy runs\nnext step: confirm DNS and egress to github.com and raw.githubusercontent.com (`curl -I %s`), then re-run", url, err.Error(), url)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("observed: %s answered HTTP %d\nmeaning: upstream no longer serves agent-sandbox %s at the address this module pins, or the host is failing\nnext step: confirm what upstream publishes (`curl -I %s`); if the release was moved or withdrawn, the module's pin needs a catalog fix", url, response.StatusCode, vars.Release, url)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrapf(err, "reading %s", url)
	}
	return string(raw), nil
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
			return nil, errors.Wrap(err, "parsing an upstream manifest document")
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

// containerNamed returns the named container of a pod template's spec.
func containerNamed(podSpec document, name string) (document, error) {
	containers, _ := podSpec["containers"].([]interface{})
	for _, c := range containers {
		if container, ok := c.(map[string]interface{}); ok && container["name"] == name {
			return container, nil
		}
	}
	return nil, errors.Errorf("upstream's manifest at %s has no container named %q; the module's pin needs a catalog fix", vars.Release, name)
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
