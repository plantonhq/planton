// activityfixture exposes only recorded responses. It has no live endpoint.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/plantonhq/planton/pkg/skills/hostprobe"
)

func main() {
	scenario := flag.String("scenario", "", "scenario JSON")
	skill := flag.String("skill", "", "planton skill directory")
	log := flag.String("calls", "", "append-only call evidence")
	flag.Parse()
	if err := run(*scenario, *skill, *log); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(scenario, skill, log string) error {
	f, err := os.Open(scenario)
	if err != nil {
		return err
	}
	defer f.Close()
	s, err := hostprobe.LoadActivityScenario(f)
	if err != nil {
		return err
	}
	calls, err := os.OpenFile(log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer calls.Close()
	root, err := os.OpenRoot(skill)
	if err != nil {
		return err
	}
	defer root.Close()
	return hostprobe.ServeActivity(os.Stdin, os.Stdout, calls, &hostprobe.ActivityTransport{Scenario: s, Skill: root.FS()})
}
