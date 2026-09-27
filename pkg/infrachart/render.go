package infrachart

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/builtins"
	"github.com/nikolalohinski/gonja/v2/config"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/nikolalohinski/gonja/v2/loaders"
	"github.com/nikolalohinski/gonja/v2/logging"
	"github.com/pkg/errors"
)

// The platform renders charts with a SANDBOXED Jinja engine: the tags and
// filters below are disabled server-side, so a template using them would
// render offline but fail (or worse, misbehave) on the platform. They are
// rejected here by source scan, keeping the offline gate at least as strict
// as the platform.
var (
	bannedTagRe    = regexp.MustCompile(`\{%-?\s*(set|include|import|from|macro|call|do)\b`)
	bannedFilterRe = regexp.MustCompile(`\|\s*(attr|map|sort|shuffle)\b`)
)

func init() {
	// gonja's lexer and parser log through logrus, and a validation library
	// must not write to the host process's output. gonja's own switch for
	// that is its logging package, which every one of its log calls checks
	// first. Its exported gonja.SetLoggerOutput(io.Discard) is NOT the way:
	// it also calls logrus.SetOutput on the PROCESS-WIDE standard logger,
	// which silences every logrus line in any binary that links this
	// package (the CLI's own fatal messages included).
	logging.SetEnabled(false)
}

// chartEnvironment carries the builtin Jinja surface plus the custom filters
// the platform registers (bool) or that charts conventionally rely on
// (b64decode), with the broken gonja string methods patched. Built once;
// never mutates gonja's shared globals.
var chartEnvironment = &exec.Environment{
	Context:           gonja.DefaultContext,
	Filters:           exec.NewFilterSet(map[string]exec.FilterFunction{"bool": filterBool, "b64decode": filterB64Decode}).Update(builtins.Filters),
	Tests:             builtins.Tests,
	ControlStructures: builtins.ControlStructures,
	Methods: exec.Methods{
		Bool:  builtins.Methods.Bool,
		Int:   builtins.Methods.Int,
		Float: builtins.Methods.Float,
		Str:   patchedStrMethods(),
		Dict:  builtins.Methods.Dict,
		List:  builtins.Methods.List,
	},
}

// gonjaStrMethodNames is gonja v2.8.0's full builtin string-method surface. The
// patched method set below copies each of these from the builtins so overriding
// the broken ones does not silently drop the rest (the builtin map itself is
// unexported and cannot be merged into).
var gonjaStrMethodNames = []string{
	"capitalize", "capwords", "casefold", "center", "count", "encode",
	"endswith", "errors", "expandtabs", "find", "format", "format_map",
	"isalnum", "isalpha", "isascii", "isdecimal", "isdigit", "islower",
	"isnumeric", "isprintable", "isspace", "istitle", "isupper", "join",
	"ljust", "lower", "lstrip", "partition", "removeprefix", "removesuffix",
	"replace", "rfind", "rjust", "rpartition", "rsplit", "rstrip", "split",
	"splitlines", "startswith", "strip", "swapcase", "title", "upper", "zfill",
}

// patchedStrMethods returns the builtin string methods with strip/lstrip/rstrip
// replaced: gonja v2.8.0's implementations initialize the right-side cut index
// to 0 instead of the string length, so any input WITHOUT trailing cut
// characters strips to the empty string ("a.com".strip() == "") -- silently
// corrupting every chart that trims a split CSV param.
func patchedStrMethods() *exec.MethodSet[string] {
	patched := make(map[string]exec.Method[string], len(gonjaStrMethodNames))
	for _, name := range gonjaStrMethodNames {
		if m, ok := builtins.Methods.Str.Get(name); ok {
			patched[name] = m
		}
	}
	patched["strip"] = stripMethod(strings.Trim)
	patched["lstrip"] = stripMethod(strings.TrimLeft)
	patched["rstrip"] = stripMethod(strings.TrimRight)
	return exec.NewMethodSet(patched)
}

// stripMethod adapts a strings.Trim* function to Python's str.strip contract:
// no argument (or an empty one) trims whitespace, otherwise the argument is the
// set of characters to remove.
func stripMethod(trim func(s, cutset string) string) exec.Method[string] {
	return func(self string, _ *exec.Value, arguments *exec.VarArgs) (any, error) {
		var cutset string
		if err := arguments.Take(
			exec.PositionalArgument("cutset", exec.AsValue(""), exec.StringArgument(&cutset)),
		); err != nil {
			return nil, exec.ErrInvalidCall(err)
		}
		if cutset == "" {
			cutset = " \t\n\r\v\f"
		}
		return trim(self, cutset), nil
	}
}

// filterBool coerces a value to boolean the way chart authors use it on
// toggle params. It is strict AND platform-faithful: booleans pass through,
// the strings "true"/"false" (case-insensitive, matching the platform's Java
// Boolean.parseBoolean semantics) and the empty string parse, and numbers are
// non-zero. Every other string is a render ERROR rather than a silent
// coercion: the platform's engine would evaluate e.g. "1" or "yes" as false,
// so accepting them offline (as true OR false) would let a template validate
// here and behave differently when published. Erroring keeps the offline
// gate strict-or-stricter than the engine of record.
func filterBool(_ *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.IsError() {
		return in
	}
	if err := params.Take(); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}
	switch {
	case in.IsNil():
		return exec.AsValue(false)
	case in.IsBool():
		return exec.AsValue(in.Bool())
	case in.IsNumber():
		return exec.AsValue(in.Float() != 0)
	case in.IsString():
		switch strings.ToLower(strings.TrimSpace(in.String())) {
		case "true":
			return exec.AsValue(true)
		case "false", "":
			return exec.AsValue(false)
		}
		return exec.AsValue(exec.ErrInvalidCall(fmt.Errorf("cannot coerce %q to bool (only \"true\"/\"false\" parse identically offline and on the platform)", in.String())))
	}
	return exec.AsValue(exec.ErrInvalidCall(fmt.Errorf("cannot coerce %s to bool", in.String())))
}

// filterB64Decode decodes a base64 string — the platform registers the same
// filter for charts that embed encoded content.
func filterB64Decode(_ *exec.Evaluator, in *exec.Value, _ *exec.VarArgs) *exec.Value {
	decoded, err := base64.StdEncoding.DecodeString(in.String())
	if err != nil {
		return exec.AsValue(errors.Wrap(err, "b64decode: input is not valid base64"))
	}
	return exec.AsValue(string(decoded))
}

// scanBannedConstructs reports the sandbox-disabled tags/filters a template
// uses, so authors learn about a server-side render failure offline.
func scanBannedConstructs(templateSource string) []string {
	var findings []string
	for _, m := range bannedTagRe.FindAllStringSubmatch(templateSource, -1) {
		findings = append(findings, fmt.Sprintf("'%s' tag is disabled by the platform's sandboxed renderer", m[1]))
	}
	for _, m := range bannedFilterRe.FindAllStringSubmatch(templateSource, -1) {
		findings = append(findings, fmt.Sprintf("'%s' filter is disabled by the platform's sandboxed renderer", m[1]))
	}
	return findings
}

// chartConfig is gonja's configuration with one change: the platform's engine
// (Jinjava) keeps a template's final newline, and gonja drops it by default
// (Jinja2's keep_trailing_newline=False). Dropped, a YAML block scalar at the
// end of a file ended differently offline than on the server, so a manifest
// could validate here and be refused by install.
func chartConfig() *config.Config {
	c := config.New()
	c.KeepTrailingNewline = true
	return c
}

// renderTemplate renders one template file with the given context.
func renderTemplate(name, source string, ctx map[string]any) (string, error) {
	baseLoader, err := loaders.NewFileSystemLoader("")
	if err != nil {
		return "", errors.Wrap(err, "failed to construct template loader")
	}
	shifted, err := loaders.NewShiftedLoader(name, bytes.NewReader([]byte(source)), baseLoader)
	if err != nil {
		return "", errors.Wrap(err, "failed to construct template loader")
	}
	tpl, err := exec.NewTemplate(name, chartConfig(), shifted, chartEnvironment)
	if err != nil {
		return "", err
	}
	return tpl.ExecuteToString(exec.NewContext(ctx))
}

// RenderTemplate renders one template source with the given param values,
// rejecting sandbox-disabled constructs first. The rendering context mirrors
// the control plane's: every param is addressable both bare and through
// `values.*`, and org/env are always present in both scopes, exactly like the
// server-side overwriter. This is the exported single-template entry point
// (the engine-conformance suite pins its semantics); the validation pipeline
// goes through Validate, which layers param typing and placeholder handling
// on top via buildContext.
func RenderTemplate(name, source string, values map[string]any, org, env string) (string, error) {
	if findings := scanBannedConstructs(source); len(findings) > 0 {
		return "", errors.Errorf("template %s uses the %s", name, findings[0])
	}

	scoped := make(map[string]any, len(values)+2)
	for k, v := range values {
		scoped[k] = v
	}
	scoped["org"] = org
	scoped["env"] = env

	ctx := make(map[string]any, len(scoped)+1)
	for k, v := range scoped {
		ctx[k] = v
	}
	ctx["values"] = scoped

	return renderTemplate(name, source, ctx)
}
