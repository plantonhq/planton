// componentsizing publishes the operator's sizing registry
// (internal/resources.ComponentSizing) as data: api/v1/component_sizing.json.
//
// The registry is the one home of every workload's measured default and the
// floor a program in its path enforces. Readers outside this module -- the
// catalog kind that installs a platform, which states each default as its own
// field's option, and the test that holds the two equal -- cannot import the
// operator's internal packages, so they read this file instead. It is
// generated, never edited: `make -C operator manifests` writes it beside the
// definition it describes, and the freshness guard fails a stale copy, so the
// published numbers can only ever be the operator's.
//
// Quantities are written in their canonical string form (resource.Quantity's
// String), sorted by path, so a re-measured default is a one-line diff.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	corev1 "k8s.io/api/core/v1"

	"github.com/plantonhq/planton/operator/internal/resources"
)

type quantities map[string]string

type entry struct {
	// Path is the spec field that overrides this workload's size.
	Path string `json:"path"`
	// Component is the component whose status reports the size in effect.
	Component string `json:"component"`
	// Requests and Limits are the operator's default.
	Requests quantities `json:"requests"`
	Limits   quantities `json:"limits"`
	// Floor is the least the requests may be, where a program enforces one.
	Floor quantities `json:"floor,omitempty"`
}

func main() {
	out := flag.String("out", "api/v1/component_sizing.json", "the file to write")
	flag.Parse()

	var entries []entry
	for _, path := range resources.SizingPaths() {
		s := resources.ComponentSizing[path]
		entries = append(entries, entry{
			Path:      path,
			Component: s.Component,
			Requests:  toQuantities(s.Default.Requests),
			Limits:    toQuantities(s.Default.Limits),
			Floor:     toQuantities(s.Floor),
		})
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "componentsizing: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, append(data, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "componentsizing: writing %s: %v\n", *out, err)
		os.Exit(1)
	}
}

func toQuantities(list corev1.ResourceList) quantities {
	if len(list) == 0 {
		return nil
	}
	out := quantities{}
	names := make([]string, 0, len(list))
	for name := range list {
		names = append(names, string(name))
	}
	sort.Strings(names)
	for _, name := range names {
		q := list[corev1.ResourceName(name)]
		out[name] = q.String()
	}
	return out
}
