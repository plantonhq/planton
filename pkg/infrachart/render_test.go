package infrachart

import (
	"strings"
	"testing"
)

func render(t *testing.T, source string, ctx map[string]any) string {
	t.Helper()
	out, err := renderTemplate("test.yaml", source, ctx)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	return out
}

func TestRenderVariableBindings(t *testing.T) {
	ctx, _, err := buildContext([]Param{
		{Name: "region", Value: "us-central1"},
	}, "acme", "dev", nil)
	if err != nil {
		t.Fatal(err)
	}

	out := render(t, "{{ values.region }}/{{ region }}/{{ org }}/{{ env }}", ctx)
	if out != "us-central1/us-central1/acme/dev" {
		t.Fatalf("unexpected render: %q", out)
	}
}

// These cases pin the render semantics for every Jinja construct the chart fleet uses.
// The control plane renders with a different engine implementation, so any drift between
// engines must surface here as a red test, never as a chart that validates offline but
// renders differently when published.
func TestRenderTemplateConformance(t *testing.T) {
	values := map[string]any{
		"name":         "demo",
		"enabled":      true,
		"disabled":     false,
		"enabled_str":  "true",
		"disabled_str": "false",
		"count":        "3",
		"csv":          "a,b,c",
		"items":        []any{"x", "y"},
		"padded":       "  padded  ",
		"long":         "abcdefghijklmnopqrstuvwxy",
	}

	cases := []struct {
		name     string
		template string
		want     string
	}{
		{"interpolation", `v: {{ values.name }}`, `v: demo`},
		{"bare org and env", `v: {{ org }}/{{ env }}`, `v: planton/dev`},
		{"values org and env", `v: {{ values.org }}/{{ values.env }}`, `v: planton/dev`},
		{"if bool true", `{% if values.enabled | bool %}yes{% else %}no{% endif %}`, `yes`},
		{"if bool false", `{% if values.disabled | bool %}yes{% else %}no{% endif %}`, `no`},
		{"if bool true string", `{% if values.enabled_str | bool %}yes{% else %}no{% endif %}`, `yes`},
		{"if bool false string", `{% if values.disabled_str | bool %}yes{% else %}no{% endif %}`, `no`},
		{"elif", `{% if values.disabled | bool %}a{% elif values.enabled | bool %}b{% else %}c{% endif %}`, `b`},
		{"for range int filter", `{% for i in range(values.count | int) %}{{ i }}{% endfor %}`, `012`},
		{"for split", `{% for s in values.csv.split(',') %}[{{ s }}]{% endfor %}`, `[a][b][c]`},
		{"for list param", `{% for s in values.items %}[{{ s }}]{% endfor %}`, `[x][y]`},
		{"replace", `{{ values.name | replace('e', 'a') }}`, `damo`},
		{"trim", `[{{ values.padded | trim }}]`, `[padded]`},
		{"upper", `{{ values.name | upper }}`, `DEMO`},
		{"lower", `{{ "DEMO" | lower }}`, `demo`},
		// Mirrors the fleet's only truncate call shape: truncate(n, true, ''). Jinja
		// leaves strings within the default 5-char leeway untouched, so only inputs
		// longer than n+5 are actually cut.
		{"truncate", `{{ values.long | truncate(15, true, '') }}`, `abcdefghijklmno`},
		{"truncate within leeway", `{{ "abcdefghij" | truncate(8, true, '') }}`, `abcdefghij`},
		{"length", `{{ values.csv.split(',') | length }}`, `3`},
		// strip/lstrip/rstrip are patched over gonja's broken builtins, which
		// stripped any input without trailing cut characters to "".
		{"strip no whitespace", `[{{ "a.com".strip() }}]`, `[a.com]`},
		{"strip whitespace", `[{{ "  a.com ".strip() }}]`, `[a.com]`},
		{"strip cutset", `[{{ "xxa.comxx".strip('x') }}]`, `[a.com]`},
		{"lstrip", `[{{ "  a.com ".lstrip() }}]`, `[a.com ]`},
		{"rstrip", `[{{ "  a.com ".rstrip() }}]`, `[  a.com]`},
		{"split strip loop", `{% for s in "*.ubuntu.com, *.debian.org".split(',') %}[{{ s.strip() }}]{% endfor %}`, `[*.ubuntu.com][*.debian.org]`},
		// The platform's engine keeps a template's final newline; dropping it
		// here changed what a YAML block scalar at the end of a file holds, so
		// a manifest validated offline that the server then refused.
		{"final newline kept", "v: {{ values.name }}\n", "v: demo\n"},
		{"final newline after a block kept", "{% if values.enabled | bool %}v: yes{% endif %}\n", "v: yes\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RenderTemplate(tc.name, tc.template, values, "planton", "dev")
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}
			if got != tc.want {
				t.Fatalf("rendered %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderBoolFilter(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"typed bool true", true, "yes"},
		{"typed bool false", false, "no"},
		{"string true", "true", "yes"},
		{"string True mixed case", "True", "yes"},
		{"string false", "false", "no"},
		{"number nonzero", float64(2), "yes"},
		{"number zero", float64(0), "no"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, err := buildContext([]Param{{Name: "flag", Value: tc.value}}, "acme", "dev", nil)
			if err != nil {
				t.Fatal(err)
			}
			out := render(t, "{% if values.flag | bool %}yes{% else %}no{% endif %}", ctx)
			if out != tc.want {
				t.Fatalf("flag=%v: got %q, want %q", tc.value, out, tc.want)
			}
		})
	}
}

// Strings other than "true"/"false" are render errors, never silent booleans:
// the platform's engine evaluates "1", "yes", or "maybe" as false (Java
// Boolean.parseBoolean semantics), so silently accepting them offline — as
// either value — would let a template validate here and behave differently
// when published.
func TestBoolFilterRejectsAmbiguousStrings(t *testing.T) {
	for _, bad := range []string{"maybe", "1", "0", "yes"} {
		t.Run(bad, func(t *testing.T) {
			_, err := RenderTemplate("bad-bool", `{% if values.v | bool %}x{% endif %}`,
				map[string]any{"v": bad}, "planton", "dev")
			if err == nil {
				t.Fatalf("expected an error coercing %q to bool", bad)
			}
		})
	}
}

func TestRenderB64DecodeFilter(t *testing.T) {
	ctx, _, err := buildContext([]Param{{Name: "blob", Value: "aGVsbG8="}}, "acme", "dev", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out := render(t, "{{ values.blob | b64decode }}", ctx); out != "hello" {
		t.Fatalf("b64decode: got %q", out)
	}
}

func TestRenderForLoopOverListParam(t *testing.T) {
	ctx, _, err := buildContext([]Param{
		{Name: "zones", Type: "list", Value: []any{"us-central1-a", "us-central1-b"}},
	}, "acme", "dev", nil)
	if err != nil {
		t.Fatal(err)
	}
	out := render(t, "{% for z in values.zones %}[{{ z }}]{% endfor %}", ctx)
	if out != "[us-central1-a][us-central1-b]" {
		t.Fatalf("for loop: got %q", out)
	}
}

func TestRenderPlaceholderForValuelessParam(t *testing.T) {
	ctx, placeholders, err := buildContext([]Param{{Name: "bucket_name"}}, "acme", "dev", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(placeholders) != 1 || placeholders[0] != "<bucket_name>" {
		t.Fatalf("placeholders: got %v", placeholders)
	}
	if out := render(t, "{{ values.bucket_name }}", ctx); out != "<bucket_name>" {
		t.Fatalf("placeholder render: got %q", out)
	}
}

func TestRenderTemplateRejectsSandboxedConstructs(t *testing.T) {
	cases := []struct {
		name     string
		template string
		wantIn   string
	}{
		{"set tag", `{% set x = 1 %}{{ x }}`, "'set' tag"},
		{"include tag", `{% include "other.yaml" %}`, "'include' tag"},
		{"macro tag", `{% macro m() %}{% endmacro %}`, "'macro' tag"},
		{"map filter", `{{ values.items | map('upper') }}`, "'map' filter"},
		{"sort filter", `{{ values.items | sort }}`, "'sort' filter"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RenderTemplate(tc.name, tc.template, map[string]any{"items": []any{"a"}}, "planton", "dev")
			if err == nil {
				t.Fatalf("expected rejection, got none")
			}
			if !strings.Contains(err.Error(), tc.wantIn) {
				t.Fatalf("error %q does not mention %q", err.Error(), tc.wantIn)
			}
		})
	}
}

func TestScanBannedConstructs(t *testing.T) {
	findings := scanBannedConstructs(`
{% set x = 1 %}
{% include "other.yaml" %}
{{ items | map('upper') }}
`)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d: %v", len(findings), findings)
	}
	joined := strings.Join(findings, "\n")
	for _, want := range []string{"set", "include", "map"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("findings missing %q: %v", want, findings)
		}
	}
}

func TestScanAllowsPermittedConstructs(t *testing.T) {
	findings := scanBannedConstructs(`
{% if values.flag | bool %}
value: {{ values.name | default("x") | upper }}
{% endif %}
{% for z in values.zones %}{{ z | trim }}{% endfor %}
{# a comment #}
`)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %v", findings)
	}
}
